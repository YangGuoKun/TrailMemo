package repository

import (
	"context"
	"errors"

	"github.com/trailmemo/internal/config"
	"github.com/trailmemo/internal/model"
	"gorm.io/gorm"
)

// CheckinRepository 打卡仓储接口
// 同时保留旧方法和 Context 方法，方便业务模块逐步迁移到 request_id 链路。
type CheckinRepository interface {
	Create(checkin *model.Checkin) error
	CreateContext(ctx context.Context, checkin *model.Checkin) error
	FindByID(id uint64) (*model.Checkin, error)
	FindByIDContext(ctx context.Context, id uint64) (*model.Checkin, error)
	FindByUserID(userID uint64, page, size int) ([]*model.Checkin, int64, error)
	FindByUserIDContext(ctx context.Context, userID uint64, page, size int) ([]*model.Checkin, int64, error)
	FindByRouteID(routeID uint64, page, size int) ([]*model.Checkin, int64, error)
	FindByRouteIDContext(ctx context.Context, routeID uint64, page, size int) ([]*model.Checkin, int64, error)
	FindByCheckpointID(checkpointID uint64) ([]*model.Checkin, error)
	FindByCheckpointIDContext(ctx context.Context, checkpointID uint64) ([]*model.Checkin, error)
	Update(checkin *model.Checkin) error
	UpdateContext(ctx context.Context, checkin *model.Checkin) error
	Delete(id uint64) error
	DeleteContext(ctx context.Context, id uint64) error
	CheckCheckinExists(userID, checkpointID uint64) (bool, error)
	CheckCheckinExistsContext(ctx context.Context, userID, checkpointID uint64) (bool, error)
	GetCheckinCountByRoute(userID, routeID uint64) (int64, error)
	GetCheckinCountByRouteContext(ctx context.Context, userID, routeID uint64) (int64, error)
}

type checkinRepository struct {
	db *gorm.DB
}

func NewCheckinRepository() CheckinRepository {
	return &checkinRepository{
		db: config.GetDB(),
	}
}

func (r *checkinRepository) Create(checkin *model.Checkin) error {
	return r.CreateContext(context.Background(), checkin)
}

func (r *checkinRepository) CreateContext(ctx context.Context, checkin *model.Checkin) error {
	return r.db.WithContext(ctx).Create(checkin).Error
}

func (r *checkinRepository) FindByID(id uint64) (*model.Checkin, error) {
	return r.FindByIDContext(context.Background(), id)
}

// FindByIDContext 使用请求上下文查询打卡记录，未找到时返回 nil，交给 service 转成业务错误。
func (r *checkinRepository) FindByIDContext(ctx context.Context, id uint64) (*model.Checkin, error) {
	var checkin model.Checkin
	err := r.db.WithContext(ctx).First(&checkin, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &checkin, nil
}

func (r *checkinRepository) FindByUserID(userID uint64, page, size int) ([]*model.Checkin, int64, error) {
	return r.FindByUserIDContext(context.Background(), userID, page, size)
}

func (r *checkinRepository) FindByUserIDContext(ctx context.Context, userID uint64, page, size int) ([]*model.Checkin, int64, error) {
	var checkins []*model.Checkin
	var total int64

	db := r.db.WithContext(ctx)
	err := db.Model(&model.Checkin{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Preload("Checkpoint").Preload("Route").Where("user_id = ?", userID).
		Order("checkin_time DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&checkins).Error
	if err != nil {
		return nil, 0, err
	}

	return checkins, total, nil
}

func (r *checkinRepository) FindByRouteID(routeID uint64, page, size int) ([]*model.Checkin, int64, error) {
	return r.FindByRouteIDContext(context.Background(), routeID, page, size)
}

func (r *checkinRepository) FindByRouteIDContext(ctx context.Context, routeID uint64, page, size int) ([]*model.Checkin, int64, error) {
	var checkins []*model.Checkin
	var total int64

	db := r.db.WithContext(ctx)
	err := db.Model(&model.Checkin{}).Where("route_id = ?", routeID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Preload("Checkpoint").Preload("User").Where("route_id = ?", routeID).
		Order("checkin_time DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&checkins).Error
	if err != nil {
		return nil, 0, err
	}

	return checkins, total, nil
}

func (r *checkinRepository) FindByCheckpointID(checkpointID uint64) ([]*model.Checkin, error) {
	return r.FindByCheckpointIDContext(context.Background(), checkpointID)
}

func (r *checkinRepository) FindByCheckpointIDContext(ctx context.Context, checkpointID uint64) ([]*model.Checkin, error) {
	var checkins []*model.Checkin
	err := r.db.WithContext(ctx).Preload("User").Where("checkpoint_id = ?", checkpointID).Order("checkin_time DESC").Find(&checkins).Error
	if err != nil {
		return nil, err
	}
	return checkins, nil
}

func (r *checkinRepository) Update(checkin *model.Checkin) error {
	return r.UpdateContext(context.Background(), checkin)
}

func (r *checkinRepository) UpdateContext(ctx context.Context, checkin *model.Checkin) error {
	return r.db.WithContext(ctx).Save(checkin).Error
}

func (r *checkinRepository) Delete(id uint64) error {
	return r.DeleteContext(context.Background(), id)
}

func (r *checkinRepository) DeleteContext(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Checkin{}, id).Error
}

func (r *checkinRepository) CheckCheckinExists(userID, checkpointID uint64) (bool, error) {
	return r.CheckCheckinExistsContext(context.Background(), userID, checkpointID)
}

func (r *checkinRepository) CheckCheckinExistsContext(ctx context.Context, userID, checkpointID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Checkin{}).Where("user_id = ? AND checkpoint_id = ?", userID, checkpointID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *checkinRepository) GetCheckinCountByRoute(userID, routeID uint64) (int64, error) {
	return r.GetCheckinCountByRouteContext(context.Background(), userID, routeID)
}

func (r *checkinRepository) GetCheckinCountByRouteContext(ctx context.Context, userID, routeID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Checkin{}).Where("user_id = ? AND route_id = ?", userID, routeID).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
