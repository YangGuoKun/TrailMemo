package repository

import (
	"context"
	"errors"

	"github.com/trailmemo/internal/config"
	"github.com/trailmemo/internal/model"
	"gorm.io/gorm"
)

type RouteRepository interface {
	Create(route *model.Route) error
	CreateWithContext(ctx context.Context, route *model.Route) error
	CreateWithTx(tx *gorm.DB, route *model.Route) error
	CreateWithTxContext(ctx context.Context, tx *gorm.DB, route *model.Route) error
	FindByID(id uint64) (*model.Route, error)
	FindByIDContext(ctx context.Context, id uint64) (*model.Route, error)
	FindByUserID(userID uint64, page, size int) ([]*model.Route, int64, error)
	FindByUserIDContext(ctx context.Context, userID uint64, page, size int) ([]*model.Route, int64, error)
	FindPublic(page, size int) ([]*model.Route, int64, error)
	FindPublicContext(ctx context.Context, page, size int) ([]*model.Route, int64, error)
	Update(route *model.Route) error
	UpdateContext(ctx context.Context, route *model.Route) error
	Delete(id uint64) error
	DeleteContext(ctx context.Context, id uint64) error
	IncrementReuseCount(id uint64) error
	IncrementReuseCountContext(ctx context.Context, id uint64) error
	IncreaseFavoriteCount(id uint64) error
	DecreaseFavoriteCount(id uint64) error
	IncreaseFavoriteCountWithTx(tx *gorm.DB, id uint64) error
	DecreaseFavoriteCountWithTx(tx *gorm.DB, id uint64) error
	Begin() *gorm.DB
	BeginTx(ctx context.Context) *gorm.DB
}

type routeRepository struct {
	db *gorm.DB
}

func NewRouteRepository() RouteRepository {
	return &routeRepository{
		db: config.GetDB(),
	}
}

func (r *routeRepository) Begin() *gorm.DB {
	return r.db.Begin()
}

// BeginTx 使用请求上下文开启事务，后续 SQL 日志会自动带上 request_id。
func (r *routeRepository) BeginTx(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Begin()
}

func (r *routeRepository) Create(route *model.Route) error {
	return r.db.Create(route).Error
}

func (r *routeRepository) CreateWithContext(ctx context.Context, route *model.Route) error {
	return r.db.WithContext(ctx).Create(route).Error
}

func (r *routeRepository) CreateWithTx(tx *gorm.DB, route *model.Route) error {
	return tx.Create(route).Error
}

func (r *routeRepository) CreateWithTxContext(ctx context.Context, tx *gorm.DB, route *model.Route) error {
	return tx.WithContext(ctx).Create(route).Error
}

// FindByID 根据 ID 查找路线，不预加载路点。
func (r *routeRepository) FindByID(id uint64) (*model.Route, error) {
	return r.FindByIDContext(context.Background(), id)
}

// FindByIDContext 根据 ID 查找路线，预加载路点。
func (r *routeRepository) FindByIDContext(ctx context.Context, id uint64) (*model.Route, error) {
	var route model.Route
	if err := r.db.WithContext(ctx).Preload("Checkpoints").First(&route, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &route, nil
}

func (r *routeRepository) FindByUserID(userID uint64, page, size int) ([]*model.Route, int64, error) {
	return r.FindByUserIDContext(context.Background(), userID, page, size)
}

func (r *routeRepository) FindByUserIDContext(ctx context.Context, userID uint64, page, size int) ([]*model.Route, int64, error) {
	var routes []*model.Route
	var total int64

	db := r.db.WithContext(ctx)
	err := db.Model(&model.Route{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Preload("Checkpoints").Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&routes).Error
	if err != nil {
		return nil, 0, err
	}

	return routes, total, nil
}

func (r *routeRepository) FindPublic(page, size int) ([]*model.Route, int64, error) {
	return r.FindPublicContext(context.Background(), page, size)
}

func (r *routeRepository) FindPublicContext(ctx context.Context, page, size int) ([]*model.Route, int64, error) {
	var routes []*model.Route
	var total int64

	db := r.db.WithContext(ctx)
	err := db.Model(&model.Route{}).Where("is_public = 1").Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Preload("Checkpoints").Where("is_public = 1").
		Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&routes).Error
	if err != nil {
		return nil, 0, err
	}

	return routes, total, nil
}

func (r *routeRepository) Update(route *model.Route) error {
	return r.db.Save(route).Error
}

func (r *routeRepository) UpdateContext(ctx context.Context, route *model.Route) error {
	return r.db.WithContext(ctx).Save(route).Error
}

func (r *routeRepository) Delete(id uint64) error {
	return r.db.Delete(&model.Route{}, id).Error
}

func (r *routeRepository) DeleteContext(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Route{}, id).Error
}

func (r *routeRepository) IncrementReuseCount(id uint64) error {
	return r.db.Model(&model.Route{}).Where("id = ?", id).UpdateColumn("reuse_count", gorm.Expr("reuse_count + ?", 1)).Error
}

func (r *routeRepository) IncrementReuseCountContext(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.Route{}).Where("id = ?", id).UpdateColumn("reuse_count", gorm.Expr("reuse_count + ?", 1)).Error
}

func (r *routeRepository) IncreaseFavoriteCount(id uint64) error {
	return r.db.Model(&model.Route{}).Where("id = ?", id).UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", 1)).Error
}

func (r *routeRepository) DecreaseFavoriteCount(id uint64) error {
	return r.db.Model(&model.Route{}).Where("id = ?", id).UpdateColumn("favorite_count", gorm.Expr("favorite_count - ?", 1)).Error
}

// 增加收藏次数（事务版本）
func (r *routeRepository) IncreaseFavoriteCountWithTx(tx *gorm.DB, id uint64) error {
	return tx.Model(&model.Route{}).Where("id = ?", id).UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", 1)).Error
}

// 减少收藏次数（事务版本）
func (r *routeRepository) DecreaseFavoriteCountWithTx(tx *gorm.DB, id uint64) error {
	return tx.Model(&model.Route{}).Where("id = ?", id).UpdateColumn("favorite_count", gorm.Expr("favorite_count - ?", 1)).Error
}
