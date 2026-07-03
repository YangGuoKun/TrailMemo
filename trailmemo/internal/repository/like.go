package repository

import (
	"github.com/trailmemo/internal/config"
	"github.com/trailmemo/internal/model"
	"gorm.io/gorm"
)

type LikeRepository interface {
	Create(like *model.Like) error
	CreateWithTx(tx *gorm.DB, like *model.Like) error
	Delete(id uint64) error
	DeleteWithTx(tx *gorm.DB, id uint64) error
	FindByUserAndTarget(userID, targetID uint64, targetType string) (*model.Like, error)
	FindByUserAndTargetWithTx(tx *gorm.DB, userID, targetID uint64, targetType string) (*model.Like, error)
	CheckExists(userID, targetID uint64, targetType string) (bool, error)
	CheckExistsWithTx(tx *gorm.DB, userID, targetID uint64, targetType string) (bool, error)
	GetLikeCount(targetID uint64, targetType string) (int64, error)
	GetUserLikes(userID uint64, targetType string, page, size int) ([]*model.Like, int64, error)
}

type likeRepository struct {
	db *gorm.DB
}

func NewLikeRepository() LikeRepository {
	return &likeRepository{
		db: config.GetDB(),
	}
}

func (r *likeRepository) Create(like *model.Like) error {
	return r.db.Create(like).Error
}

func (r *likeRepository) CreateWithTx(tx *gorm.DB, like *model.Like) error {
	return tx.Create(like).Error
}

func (r *likeRepository) Delete(id uint64) error {
	return r.db.Delete(&model.Like{}, id).Error
}

func (r *likeRepository) DeleteWithTx(tx *gorm.DB, id uint64) error {
	return tx.Delete(&model.Like{}, id).Error
}

// 根据用户ID和目标ID查询点赞记录
func (r *likeRepository) FindByUserAndTarget(userID, targetID uint64, targetType string) (*model.Like, error) {
	var like model.Like
	err := r.db.Where("user_id = ? AND target_id = ? AND target_type = ?", userID, targetID, targetType).First(&like).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &like, nil
}

// 根据用户ID和目标ID查询点赞记录（事务版本）
func (r *likeRepository) FindByUserAndTargetWithTx(tx *gorm.DB, userID, targetID uint64, targetType string) (*model.Like, error) {
	var like model.Like
	err := tx.Where("user_id = ? AND target_id = ? AND target_type = ?", userID, targetID, targetType).First(&like).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &like, nil
}

// 检查点赞记录是否存在
func (r *likeRepository) CheckExists(userID, targetID uint64, targetType string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Like{}).Where("user_id = ? AND target_id = ? AND target_type = ?", userID, targetID, targetType).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// 检查点赞记录是否存在（事务版本）
func (r *likeRepository) CheckExistsWithTx(tx *gorm.DB, userID, targetID uint64, targetType string) (bool, error) {
	var count int64
	err := tx.Model(&model.Like{}).Where("user_id = ? AND target_id = ? AND target_type = ?", userID, targetID, targetType).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// 根据目标ID和目标类型查询点赞数量
func (r *likeRepository) GetLikeCount(targetID uint64, targetType string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Like{}).Where("target_id = ? AND target_type = ?", targetID, targetType).Count(&count).Error
	return count, err
}

// 根据用户ID和目标类型查询用户点赞记录
func (r *likeRepository) GetUserLikes(userID uint64, targetType string, page, size int) ([]*model.Like, int64, error) {
	var likes []*model.Like
	var total int64

	err := r.db.Model(&model.Like{}).Where("user_id = ? AND target_type = ?", userID, targetType).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("user_id = ? AND target_type = ?", userID, targetType).
		Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&likes).Error
	if err != nil {
		return nil, 0, err
	}

	return likes, total, nil
}
