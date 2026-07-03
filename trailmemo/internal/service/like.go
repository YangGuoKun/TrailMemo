package service

import (
	"context"
	"net/http"
	"time"

	"github.com/trailmemo/internal/config"
	"github.com/trailmemo/internal/model"
	platformlogger "github.com/trailmemo/internal/platform/logger"
	"github.com/trailmemo/internal/repository"
	"github.com/trailmemo/pkg/apperror"
	"go.uber.org/zap"
)

type LikeService interface {
	ToggleLike(ctx context.Context, userID, targetID uint64, targetType string) (bool, error)
	CheckLikeStatus(ctx context.Context, userID, targetID uint64, targetType string) (bool, error)
	GetLikeCount(ctx context.Context, targetID uint64, targetType string) (int64, error)
}

type likeService struct {
	likeRepo    repository.LikeRepository
	postRepo    repository.PostRepository
	commentRepo repository.CommentRepository
}

func NewLikeService() LikeService {
	return &likeService{
		likeRepo:    repository.NewLikeRepository(),
		postRepo:    repository.NewPostRepository(),
		commentRepo: repository.NewCommentRepository(),
	}
}

// 切换点赞状态
func (s *likeService) ToggleLike(ctx context.Context, userID, targetID uint64, targetType string) (bool, error) {
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "like"),
		zap.String("operation", "toggle_like"),
		zap.Uint64("user_id", userID),
		zap.Uint64("target_id", targetID),
		zap.String("target_type", targetType),
	)

	// 创建分布式锁，锁的过期时间设置为事务最大可能执行时间的3倍
	lockKey := getLikeLockKey(userID, targetID, targetType)
	lock := NewRedisLock(lockKey, 30*time.Second)

	// 尝试获取锁，最多重试3次
	log.Info("lock_acquire_start",
		zap.String("event", "lock_acquire_start"),
		zap.String("lock_key", lockKey),
	)
	result, err := lock.TryAcquire(3, 100*time.Millisecond)
	if err != nil {
		log.Error("lock_acquire_failed",
			zap.String("event", "lock_acquire_failed"),
			zap.String("error_kind", apperror.KindExternal),
			zap.Error(err),
		)
		return false, apperror.Wrap(err, apperror.CodeExternalError, apperror.KindExternal, "获取分布式锁失败", http.StatusInternalServerError)
	}
	if !result.Acquired {
		if result.Cancelled {
			log.Warn("lock_acquire_cancelled",
				zap.String("event", "lock_acquire_cancelled"),
			)
			return false, apperror.New(apperror.CodeConflict, apperror.KindValidation, "操作已取消", http.StatusBadRequest)
		}
		log.Warn("lock_acquire_timeout",
			zap.String("event", "lock_acquire_timeout"),
		)
		return false, apperror.New(apperror.CodeLikeConflict, apperror.KindValidation, "操作过于频繁，请稍后重试", http.StatusTooManyRequests)
	}

	log.Info("lock_acquire_success",
		zap.String("event", "lock_acquire_success"),
	)

	// 确保释放锁
	defer func() {
		lock.Release()
		log.Info("lock_release",
			zap.String("event", "lock_release"),
			zap.String("lock_key", lockKey),
		)
	}()

	// 获取数据库连接并开始事务
	db := config.GetDB()
	tx := db.Begin()
	log.Info("transaction_start",
		zap.String("event", "transaction_start"),
	)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error("transaction_panic",
				zap.String("event", "transaction_panic"),
				zap.Any("panic", r),
			)
		}
	}()

	// 在事务中再次检查状态（双重检查），检查是否已点赞或取消点赞过
	exists, err := s.likeRepo.CheckExistsWithTx(tx, userID, targetID, targetType)
	if err != nil {
		tx.Rollback()
		log.Error("db_query_failed",
			zap.String("event", "db_query_failed"),
			zap.String("transaction_node", "check_exists"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return false, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询点赞状态失败", http.StatusInternalServerError)
	}

	if exists { // 取消点赞
		like, err := s.likeRepo.FindByUserAndTargetWithTx(tx, userID, targetID, targetType)
		if err != nil {
			tx.Rollback()
			log.Error("db_query_failed",
				zap.String("event", "db_query_failed"),
				zap.String("transaction_node", "find_like"),
				zap.String("error_kind", apperror.KindDB),
				zap.Error(err),
			)
			return false, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询点赞记录失败", http.StatusInternalServerError)
		}
		if like != nil {
			err = s.likeRepo.DeleteWithTx(tx, like.ID)
			if err != nil {
				tx.Rollback()
				log.Error("like_delete_failed",
					zap.String("event", "like_delete_failed"),
					zap.String("transaction_node", "delete_like"),
					zap.String("error_kind", apperror.KindDB),
					zap.Error(err),
				)
				return false, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "删除点赞失败", http.StatusInternalServerError)
			}
		}

		if targetType == "post" {
			err = s.postRepo.DecreaseLikeCountWithTx(tx, targetID)
		} else if targetType == "comment" {
			err = s.commentRepo.DecreaseLikeCountWithTx(tx, targetID)
		}
		if err != nil {
			tx.Rollback()
			log.Error("decrease_count_failed",
				zap.String("event", "decrease_count_failed"),
				zap.String("transaction_node", "decrease_count"),
				zap.String("error_kind", apperror.KindDB),
				zap.Error(err),
			)
			return false, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "更新点赞数量失败", http.StatusInternalServerError)
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			log.Error("transaction_commit_failed",
				zap.String("event", "transaction_commit_failed"),
				zap.String("error_kind", apperror.KindDB),
				zap.Error(err),
			)
			return false, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "提交事务失败", http.StatusInternalServerError)
		}

		log.Info("like_removed",
			zap.String("event", "like_removed"),
			zap.String("entity_type", "like"),
			zap.Uint64("user_id", userID),
			zap.Uint64("target_id", targetID),
			zap.String("target_type", targetType),
			zap.Bool("is_liked", false),
		)
		return false, nil
	}

	// 点赞帖子或评论
	like := &model.Like{
		UserID:     userID,
		TargetID:   targetID,
		TargetType: targetType,
	}

	err = s.likeRepo.CreateWithTx(tx, like)
	if err != nil {
		tx.Rollback()
		log.Error("like_create_failed",
			zap.String("event", "like_create_failed"),
			zap.String("transaction_node", "create_like"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return false, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "创建点赞失败", http.StatusInternalServerError)
	}

	if targetType == "post" {
		err = s.postRepo.IncreaseLikeCountWithTx(tx, targetID)
	} else if targetType == "comment" {
		err = s.commentRepo.IncreaseLikeCountWithTx(tx, targetID)
	}
	if err != nil {
		tx.Rollback()
		log.Error("increase_count_failed",
			zap.String("event", "increase_count_failed"),
			zap.String("transaction_node", "increase_count"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return false, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "更新点赞数量失败", http.StatusInternalServerError)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Error("transaction_commit_failed",
			zap.String("event", "transaction_commit_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return false, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "提交事务失败", http.StatusInternalServerError)
	}

	log.Info("like_created",
		zap.String("event", "like_created"),
		zap.String("entity_type", "like"),
		zap.Uint64("entity_id", like.ID),
		zap.Uint64("user_id", userID),
		zap.Uint64("target_id", targetID),
		zap.String("target_type", targetType),
		zap.Bool("is_liked", true),
	)
	return true, nil
}

// 检查点赞状态
func (s *likeService) CheckLikeStatus(ctx context.Context, userID, targetID uint64, targetType string) (bool, error) {
	exists, err := s.likeRepo.CheckExists(userID, targetID, targetType)
	if err != nil {
		platformlogger.FromContext(ctx).Error("db_query_failed",
			zap.String("module", "like"),
			zap.String("operation", "check_like_status"),
			zap.Uint64("user_id", userID),
			zap.Uint64("target_id", targetID),
			zap.String("target_type", targetType),
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return false, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询点赞状态失败", http.StatusInternalServerError)
	}
	return exists, nil
}

// 获取点赞数量
func (s *likeService) GetLikeCount(ctx context.Context, targetID uint64, targetType string) (int64, error) {
	count, err := s.likeRepo.GetLikeCount(targetID, targetType)
	if err != nil {
		platformlogger.FromContext(ctx).Error("db_query_failed",
			zap.String("module", "like"),
			zap.String("operation", "get_like_count"),
			zap.Uint64("target_id", targetID),
			zap.String("target_type", targetType),
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return 0, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询点赞数量失败", http.StatusInternalServerError)
	}
	return count, nil
}
