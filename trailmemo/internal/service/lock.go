package service

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/trailmemo/internal/config"
	"github.com/trailmemo/internal/middleware"
	"go.uber.org/zap"
)

// 初始化随机数生成器
func init() {
	rand.Seed(time.Now().UnixNano())
}

// RedisLock 分布式锁实现
type RedisLock struct {
	client          *redis.Client
	key             string
	value           string
	expiration      time.Duration
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	isAcquired      bool
	refreshTicker   *time.Ticker
	maxLockDuration time.Duration
	logger          *zap.Logger
	acquireTime     time.Time // 锁获取时间，用于计算持有时间
}

// NewRedisLock 创建分布式锁实例
// expiration: 锁的过期时间，建议设置为事务最大执行时间的3倍
func NewRedisLock(key string, expiration time.Duration) *RedisLock {
	ctx, cancel := context.WithCancel(context.Background())
	return &RedisLock{
		client:          config.GetRedis(),
		key:             key,
		value:           uuid.New().String(), // 使用UUID防止误删其他客户端的锁
		expiration:      expiration,
		ctx:             ctx,
		cancel:          cancel,
		maxLockDuration: 5 * time.Minute, // 最大持有锁时间，防止死锁
		logger:          middleware.GetLogger(),
	}
}

// Acquire 获取锁（单次尝试）
func (lock *RedisLock) Acquire() (bool, error) {
	// 使用 SET NX 命令：只有当 key 不存在时才设置成功
	result, err := lock.client.SetNX(lock.ctx, lock.key, lock.value, lock.expiration).Result()
	if err != nil {
		lock.logger.Error("Failed to acquire lock",
			zap.String("key", lock.key),
			zap.Error(err))
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if result {
		lock.isAcquired = true
		lock.acquireTime = time.Now()
		lock.logger.Info("Lock acquired",
			zap.String("key", lock.key),
			zap.Duration("expiration", lock.expiration))
		// 启动锁续命机制：每 expiration/3 时间刷新一次
		lock.startRefresh()
	} else {
		lock.logger.Debug("Lock not acquired (already held)",
			zap.String("key", lock.key))
	}
	return result, nil
}

// startRefresh 启动锁续命机制
// 每隔 expiration/3 的时间刷新锁的过期时间，确保锁在事务执行期间不会过期
func (lock *RedisLock) startRefresh() {
	refreshInterval := lock.expiration / 3
	lock.refreshTicker = time.NewTicker(refreshInterval)
	lock.wg.Add(1)

	go func() {
		defer lock.wg.Done()
		startTime := time.Now()
		refreshCount := 0

		lock.logger.Debug("Lock refresh goroutine started",
			zap.String("key", lock.key),
			zap.Duration("refresh_interval", refreshInterval))

		for {
			select {
			case <-lock.refreshTicker.C:
				// 检查是否超过最大持有时间
				if time.Since(startTime) > lock.maxLockDuration {
					lock.logger.Warn("Lock exceeded max duration, stopping refresh",
						zap.String("key", lock.key),
						zap.Duration("max_duration", lock.maxLockDuration),
						zap.Duration("actual_duration", time.Since(startTime)))
					return // 超过最大时间，自动停止续命
				}

				// 刷新锁的过期时间
				_, err := lock.client.Expire(lock.ctx, lock.key, lock.expiration).Result()
				if err != nil {
					lock.logger.Error("Failed to refresh lock",
						zap.String("key", lock.key),
						zap.Error(err))
					return // 刷新失败，退出
				}

				refreshCount++
				lock.logger.Debug("Lock refreshed",
					zap.String("key", lock.key),
					zap.Int("refresh_count", refreshCount),
					zap.Duration("elapsed", time.Since(startTime)))

			case <-lock.ctx.Done():
				lock.logger.Debug("Lock refresh goroutine stopped (context cancelled)",
					zap.String("key", lock.key),
					zap.Int("refresh_count", refreshCount),
					zap.Duration("elapsed", time.Since(startTime)))
				return // 上下文取消，退出
			}
		}
	}()
}

// Release 释放锁
func (lock *RedisLock) Release() error {
	// 如果没有获取到锁，无需释放
	if !lock.isAcquired {
		lock.logger.Debug("Skipping lock release (not acquired)",
			zap.String("key", lock.key))
		return nil
	}

	// 计算锁持有时间
	holdDuration := time.Since(lock.acquireTime)

	// 停止续命 goroutine
	if lock.refreshTicker != nil {
		lock.refreshTicker.Stop()
	}

	// 取消上下文并等待 goroutine 退出
	lock.cancel()
	lock.wg.Wait()

	// 使用 Lua 脚本保证原子性：只有当 value 匹配时才删除
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result, err := lock.client.Eval(lock.ctx, script, []string{lock.key}, lock.value).Result()
	if err != nil {
		lock.logger.Error("Failed to release lock",
			zap.String("key", lock.key),
			zap.Duration("hold_duration", holdDuration),
			zap.Error(err))
		return fmt.Errorf("failed to release lock: %w", err)
	}

	lock.isAcquired = false

	if result.(int64) > 0 {
		lock.logger.Info("Lock released",
			zap.String("key", lock.key),
			zap.Duration("hold_duration", holdDuration))
	} else {
		lock.logger.Warn("Lock not released (value mismatch or key not found)",
			zap.String("key", lock.key),
			zap.Duration("hold_duration", holdDuration))
	}

	return nil
}

// TryAcquireResult 表示尝试获取锁的结果
type TryAcquireResult struct {
	Acquired  bool
	Cancelled bool // 是否因为上下文取消而失败
	Err       error
}

// TryAcquire 尝试获取锁，带指数退避重试
// maxRetries: 最大重试次数
// baseInterval: 基础重试间隔
func (lock *RedisLock) TryAcquire(maxRetries int, baseInterval time.Duration) (TryAcquireResult, error) {
	for i := 0; i < maxRetries; i++ {
		acquired, err := lock.Acquire()
		if err != nil {
			return TryAcquireResult{Acquired: false}, err
		}
		if acquired {
			return TryAcquireResult{Acquired: true}, nil
		}

		// 指数退避：第 i 次重试等待 baseInterval * 2^i 时间
		// 添加随机抖动（0-100ms），避免多个请求同时重试导致惊群效应
		waitTime := baseInterval * time.Duration(1<<i)
		jitter := time.Duration(rand.Intn(100)) * time.Millisecond
		waitTime += jitter

		lock.logger.Debug("Retrying lock acquisition",
			zap.String("key", lock.key),
			zap.Int("retry", i+1),
			zap.Duration("wait_time", waitTime),
			zap.Duration("jitter", jitter))

		// 等待期间检查上下文是否取消
		select {
		case <-time.After(waitTime):
			// 继续重试
		case <-lock.ctx.Done():
			lock.logger.Info("Lock acquisition cancelled",
				zap.String("key", lock.key),
				zap.Int("retries_attempted", i+1))
			return TryAcquireResult{Acquired: false, Cancelled: true}, nil
		}
	}

	lock.logger.Warn("Lock acquisition failed after retries",
		zap.String("key", lock.key),
		zap.Int("max_retries", maxRetries))
	return TryAcquireResult{Acquired: false}, nil
}

// 获取点赞操作的锁 key
// 格式: trailmemo:like:lock:{userID}:{targetID}:{targetType}
func getLikeLockKey(userID, targetID uint64, targetType string) string {
	return fmt.Sprintf("trailmemo:like:lock:%d:%d:%s", userID, targetID, targetType)
}

// 获取收藏操作的锁 key
// 格式: trailmemo:favorite:lock:{userID}:{routeID}
func getFavoriteLockKey(userID, routeID uint64) string {
	return fmt.Sprintf("trailmemo:favorite:lock:%d:%d", userID, routeID)
}
