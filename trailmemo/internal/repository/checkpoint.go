package repository

import (
	"context"
	"errors"

	"github.com/trailmemo/internal/config"
	"github.com/trailmemo/internal/model"
	"gorm.io/gorm"
)

type CheckpointRepository interface {
	Create(checkpoint *model.Checkpoint) error
	CreateWithTx(tx *gorm.DB, checkpoint *model.Checkpoint) error
	CreateBatch(checkpoints []*model.Checkpoint) error
	CreateBatchWithTx(tx *gorm.DB, checkpoints []*model.Checkpoint) error
	CreateBatchWithTxContext(ctx context.Context, tx *gorm.DB, checkpoints []*model.Checkpoint) error
	FindByRouteID(routeID uint64) ([]*model.Checkpoint, error)
	FindByRouteIDContext(ctx context.Context, routeID uint64) ([]*model.Checkpoint, error)
	FindByID(id uint64) (*model.Checkpoint, error)
	FindByIDContext(ctx context.Context, id uint64) (*model.Checkpoint, error)
	Update(checkpoint *model.Checkpoint) error
	Delete(id uint64) error
	DeleteByRouteID(routeID uint64) error
	DeleteByRouteIDContext(ctx context.Context, routeID uint64) error
}

type checkpointRepository struct {
	db *gorm.DB
}

func NewCheckpointRepository() CheckpointRepository {
	return &checkpointRepository{
		db: config.GetDB(),
	}
}

func (r *checkpointRepository) Create(checkpoint *model.Checkpoint) error {
	return r.db.Create(checkpoint).Error
}

func (r *checkpointRepository) CreateWithTx(tx *gorm.DB, checkpoint *model.Checkpoint) error {
	return tx.Create(checkpoint).Error
}

func (r *checkpointRepository) CreateBatch(checkpoints []*model.Checkpoint) error {
	return r.db.Create(&checkpoints).Error
}

func (r *checkpointRepository) CreateBatchWithTx(tx *gorm.DB, checkpoints []*model.Checkpoint) error {
	return tx.Create(&checkpoints).Error
}

func (r *checkpointRepository) CreateBatchWithTxContext(ctx context.Context, tx *gorm.DB, checkpoints []*model.Checkpoint) error {
	return tx.WithContext(ctx).Create(&checkpoints).Error
}

func (r *checkpointRepository) FindByRouteID(routeID uint64) ([]*model.Checkpoint, error) {
	return r.FindByRouteIDContext(context.Background(), routeID)
}

func (r *checkpointRepository) FindByRouteIDContext(ctx context.Context, routeID uint64) ([]*model.Checkpoint, error) {
	var checkpoints []*model.Checkpoint
	err := r.db.WithContext(ctx).Where("route_id = ?", routeID).Order("sequence ASC").Find(&checkpoints).Error
	if err != nil {
		return nil, err
	}
	return checkpoints, nil
}

func (r *checkpointRepository) FindByID(id uint64) (*model.Checkpoint, error) {
	return r.FindByIDContext(context.Background(), id)
}

// FindByIDContext 使用请求上下文查询打卡点，让 GORM 日志能携带 request_id。
func (r *checkpointRepository) FindByIDContext(ctx context.Context, id uint64) (*model.Checkpoint, error) {
	var checkpoint model.Checkpoint
	err := r.db.WithContext(ctx).First(&checkpoint, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &checkpoint, nil
}

func (r *checkpointRepository) Update(checkpoint *model.Checkpoint) error {
	return r.db.Save(checkpoint).Error
}

func (r *checkpointRepository) Delete(id uint64) error {
	return r.db.Delete(&model.Checkpoint{}, id).Error
}

func (r *checkpointRepository) DeleteByRouteID(routeID uint64) error {
	return r.db.Delete(&model.Checkpoint{}, "route_id = ?", routeID).Error
}

func (r *checkpointRepository) DeleteByRouteIDContext(ctx context.Context, routeID uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Checkpoint{}, "route_id = ?", routeID).Error
}
