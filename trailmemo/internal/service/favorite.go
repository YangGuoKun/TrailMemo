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

// FavoriteService 路由收藏服务接口
type FavoriteService interface {
	ToggleFavorite(ctx context.Context, userID, routeID uint64) (bool, error)
	CheckFavoriteStatus(ctx context.Context, userID, routeID uint64) (bool, error)
	GetFavoriteCount(ctx context.Context, routeID uint64) (int64, error)
	GetUserFavorites(ctx context.Context, userID uint64, page, size int) ([]*model.Favorite, int64, error)
}

type favoriteService struct {
	favoriteRepo repository.FavoriteRepository
	routeRepo    repository.RouteRepository
}

func NewFavoriteService() FavoriteService {
	return &favoriteService{
		favoriteRepo: repository.NewFavoriteRepository(),
		routeRepo:    repository.NewRouteRepository(),
	}
}

// ToggleFavorite 路由收藏/取消收藏
func (s *favoriteService) ToggleFavorite(ctx context.Context, userID, routeID uint64) (bool, error) {
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "favorite"),
		zap.String("operation", "toggle_favorite"),
		zap.Uint64("user_id", userID),
		zap.Uint64("route_id", routeID),
	)

	// 创建分布式锁，锁的过期时间设置为事务最大可能执行时间的3倍
	lockKey := getFavoriteLockKey(userID, routeID)
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
		return false, apperror.New(apperror.CodeFavoriteConflict, apperror.KindValidation, "操作过于频繁，请稍后重试", http.StatusTooManyRequests)
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

	// 在事务中再次检查状态（双重检查）
	exists, err := s.favoriteRepo.CheckExistsWithTx(tx, userID, routeID)
	if err != nil {
		tx.Rollback()
		log.Error("db_query_failed",
			zap.String("event", "db_query_failed"),
			zap.String("transaction_node", "check_exists"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return false, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询收藏状态失败", http.StatusInternalServerError)
	}

	if exists {
		// 取消收藏逻辑
		err = s.favoriteRepo.DeleteWithTx(tx, userID, routeID)
		if err != nil {
			tx.Rollback()
			log.Error("favorite_delete_failed",
				zap.String("event", "favorite_delete_failed"),
				zap.String("transaction_node", "delete_favorite"),
				zap.String("error_kind", apperror.KindDB),
				zap.Error(err),
			)
			return false, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "删除收藏失败", http.StatusInternalServerError)
		}

		err = s.routeRepo.DecreaseFavoriteCountWithTx(tx, routeID)
		if err != nil {
			tx.Rollback()
			log.Error("decrease_count_failed",
				zap.String("event", "decrease_count_failed"),
				zap.String("transaction_node", "decrease_count"),
				zap.String("error_kind", apperror.KindDB),
				zap.Error(err),
			)
			return false, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "更新收藏数量失败", http.StatusInternalServerError)
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

		log.Info("favorite_removed",
			zap.String("event", "favorite_removed"),
			zap.String("entity_type", "favorite"),
			zap.Uint64("user_id", userID),
			zap.Uint64("route_id", routeID),
			zap.Bool("is_favorited", false),
		)
		return false, nil
	}

	// 添加收藏逻辑
	favorite := &model.Favorite{
		UserID:  userID,
		RouteID: routeID,
	}

	err = s.favoriteRepo.CreateWithTx(tx, favorite)
	if err != nil {
		tx.Rollback()
		log.Error("favorite_create_failed",
			zap.String("event", "favorite_create_failed"),
			zap.String("transaction_node", "create_favorite"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return false, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "创建收藏失败", http.StatusInternalServerError)
	}

	err = s.routeRepo.IncreaseFavoriteCountWithTx(tx, routeID)
	if err != nil {
		tx.Rollback()
		log.Error("increase_count_failed",
			zap.String("event", "increase_count_failed"),
			zap.String("transaction_node", "increase_count"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return false, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "更新收藏数量失败", http.StatusInternalServerError)
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

	log.Info("favorite_created",
		zap.String("event", "favorite_created"),
		zap.String("entity_type", "favorite"),
		zap.Uint64("entity_id", favorite.ID),
		zap.Uint64("user_id", userID),
		zap.Uint64("route_id", routeID),
		zap.Bool("is_favorited", true),
	)
	return true, nil
}

// CheckFavoriteStatus 检查路由收藏状态
func (s *favoriteService) CheckFavoriteStatus(ctx context.Context, userID, routeID uint64) (bool, error) {
	exists, err := s.favoriteRepo.CheckExists(userID, routeID)
	if err != nil {
		platformlogger.FromContext(ctx).Error("db_query_failed",
			zap.String("module", "favorite"),
			zap.String("operation", "check_favorite_status"),
			zap.Uint64("user_id", userID),
			zap.Uint64("route_id", routeID),
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return false, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询收藏状态失败", http.StatusInternalServerError)
	}
	return exists, nil
}

// GetFavoriteCount 获取路由收藏数量
func (s *favoriteService) GetFavoriteCount(ctx context.Context, routeID uint64) (int64, error) {
	count, err := s.favoriteRepo.GetFavoriteCount(routeID)
	if err != nil {
		platformlogger.FromContext(ctx).Error("db_query_failed",
			zap.String("module", "favorite"),
			zap.String("operation", "get_favorite_count"),
			zap.Uint64("route_id", routeID),
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return 0, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询收藏数量失败", http.StatusInternalServerError)
	}
	return count, nil
}

// GetUserFavorites 获取用户收藏列表
func (s *favoriteService) GetUserFavorites(ctx context.Context, userID uint64, page, size int) ([]*model.Favorite, int64, error) {
	favorites, total, err := s.favoriteRepo.GetUserFavorites(userID, page, size)
	if err != nil {
		platformlogger.FromContext(ctx).Error("db_query_failed",
			zap.String("module", "favorite"),
			zap.String("operation", "get_user_favorites"),
			zap.Uint64("user_id", userID),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return nil, 0, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询用户收藏列表失败", http.StatusInternalServerError)
	}
	return favorites, total, nil
}
