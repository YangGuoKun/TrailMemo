package repository

import (
	"github.com/trailmemo/internal/config"
	"github.com/trailmemo/internal/model"
	"gorm.io/gorm"
)

type CommentRepository interface {
	Create(comment *model.Comment) error
	FindByID(id uint64) (*model.Comment, error)
	FindByPostID(postID uint64, page, size int) ([]*model.Comment, int64, error)
	FindByUserID(userID uint64, page, size int) ([]*model.Comment, int64, error)
	Update(comment *model.Comment) error
	Delete(id uint64) error
	IncreaseLikeCount(id uint64) error
	DecreaseLikeCount(id uint64) error
	IncreaseLikeCountWithTx(tx *gorm.DB, id uint64) error
	DecreaseLikeCountWithTx(tx *gorm.DB, id uint64) error
	GetCommentCountByPostID(postID uint64) (int64, error)
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository() CommentRepository {
	return &commentRepository{
		db: config.GetDB(),
	}
}

// 创建评论
func (r *commentRepository) Create(comment *model.Comment) error {
	return r.db.Create(comment).Error
}

// 根据ID查询评论
func (r *commentRepository) FindByID(id uint64) (*model.Comment, error) {
	var comment model.Comment
	err := r.db.Preload("User").First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// 根据帖子ID查询评论列表
func (r *commentRepository) FindByPostID(postID uint64, page, size int) ([]*model.Comment, int64, error) {
	var comments []*model.Comment
	var total int64

	err := r.db.Model(&model.Comment{}).Where("post_id = ?", postID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Preload("User").Where("post_id = ?", postID).
		Order("created_at ASC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&comments).Error
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// 根据用户ID查询评论列表
func (r *commentRepository) FindByUserID(userID uint64, page, size int) ([]*model.Comment, int64, error) {
	var comments []*model.Comment
	var total int64

	err := r.db.Model(&model.Comment{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	// 根据创建时间降序查询评论
	err = r.db.Preload("User").Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&comments).Error
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func (r *commentRepository) Update(comment *model.Comment) error {
	return r.db.Save(comment).Error
}

func (r *commentRepository) Delete(id uint64) error {
	return r.db.Delete(&model.Comment{}, id).Error
}

//需要保持数据的一致性，所以需要在事务中执行

// 增加评论点赞次数
func (r *commentRepository) IncreaseLikeCount(id uint64) error {
	return r.db.Model(&model.Comment{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error
}

// 减少评论点赞次数
func (r *commentRepository) DecreaseLikeCount(id uint64) error {
	return r.db.Model(&model.Comment{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).Error
}

// 增加评论点赞次数（事务版本）
func (r *commentRepository) IncreaseLikeCountWithTx(tx *gorm.DB, id uint64) error {
	return tx.Model(&model.Comment{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error
}

// 减少评论点赞次数（事务版本）
func (r *commentRepository) DecreaseLikeCountWithTx(tx *gorm.DB, id uint64) error {
	return tx.Model(&model.Comment{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).Error
}

// 根据帖子ID查询评论数量
func (r *commentRepository) GetCommentCountByPostID(postID uint64) (int64, error) {
	var count int64
	err := r.db.Model(&model.Comment{}).Where("post_id = ?", postID).Count(&count).Error
	return count, err
}
