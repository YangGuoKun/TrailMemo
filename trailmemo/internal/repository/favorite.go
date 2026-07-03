package repository

import (
	"github.com/trailmemo/internal/config"
	"github.com/trailmemo/internal/model"
	"gorm.io/gorm"
)

// FavoriteRepository 路由收藏仓库接口

type FavoriteRepository interface {
	Create(favorite *model.Favorite) error
	CreateWithTx(tx *gorm.DB, favorite *model.Favorite) error
	Delete(userID, routeID uint64) error
	DeleteWithTx(tx *gorm.DB, userID, routeID uint64) error
	FindByUserAndRoute(userID, routeID uint64) (*model.Favorite, error)
	CheckExists(userID, routeID uint64) (bool, error)
	CheckExistsWithTx(tx *gorm.DB, userID, routeID uint64) (bool, error)
	GetFavoriteCount(routeID uint64) (int64, error)
	GetUserFavorites(userID uint64, page, size int) ([]*model.Favorite, int64, error)
}

type favoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository() FavoriteRepository {
	return &favoriteRepository{
		db: config.GetDB(),
	}
}

func (r *favoriteRepository) Create(favorite *model.Favorite) error {
	return r.db.Create(favorite).Error
}

func (r *favoriteRepository) CreateWithTx(tx *gorm.DB, favorite *model.Favorite) error {
	return tx.Create(favorite).Error
}

func (r *favoriteRepository) Delete(userID, routeID uint64) error {
	return r.db.Where("user_id = ? AND route_id = ?", userID, routeID).Delete(&model.Favorite{}).Error
}

func (r *favoriteRepository) DeleteWithTx(tx *gorm.DB, userID, routeID uint64) error {
	return tx.Where("user_id = ? AND route_id = ?", userID, routeID).Delete(&model.Favorite{}).Error
}

// 根据用户ID和路由ID查询路由收藏记录
func (r *favoriteRepository) FindByUserAndRoute(userID, routeID uint64) (*model.Favorite, error) {
	var favorite model.Favorite
	err := r.db.Where("user_id = ? AND route_id = ?", userID, routeID).First(&favorite).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &favorite, nil
}

// 检查路由收藏记录是否存在
func (r *favoriteRepository) CheckExists(userID, routeID uint64) (bool, error) {
	var count int64
	err := r.db.Model(&model.Favorite{}).Where("user_id = ? AND route_id = ?", userID, routeID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// 检查路由收藏记录是否存在（事务版本）
func (r *favoriteRepository) CheckExistsWithTx(tx *gorm.DB, userID, routeID uint64) (bool, error) {
	var count int64
	err := tx.Model(&model.Favorite{}).Where("user_id = ? AND route_id = ?", userID, routeID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// 根据路由ID查询路由收藏数量
func (r *favoriteRepository) GetFavoriteCount(routeID uint64) (int64, error) {
	var count int64
	err := r.db.Model(&model.Favorite{}).Where("route_id = ?", routeID).Count(&count).Error
	return count, err
}

// 根据用户ID查询用户收藏记录
func (r *favoriteRepository) GetUserFavorites(userID uint64, page, size int) ([]*model.Favorite, int64, error) {
	var favorites []*model.Favorite
	var total int64

	err := r.db.Model(&model.Favorite{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Preload("Route").Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&favorites).Error
	if err != nil {
		return nil, 0, err
	}

	return favorites, total, nil
}
