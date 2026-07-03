package repository

import (
	"github.com/trailmemo/internal/config"
	"github.com/trailmemo/internal/model"
	"gorm.io/gorm"
)

type PostRepository interface {
	Create(post *model.Post) error
	FindByID(id uint64) (*model.Post, error)
	FindByUserID(userID uint64, page, size int) ([]*model.Post, int64, error)
	FindByRouteID(routeID uint64, page, size int) ([]*model.Post, int64, error)
	FindAllPublic(page, size int) ([]*model.Post, int64, error)
	Update(post *model.Post) error
	Delete(id uint64) error
	IncreaseViewCount(id uint64) error
	IncreaseLikeCount(id uint64) error
	DecreaseLikeCount(id uint64) error
	IncreaseLikeCountWithTx(tx *gorm.DB, id uint64) error
	DecreaseLikeCountWithTx(tx *gorm.DB, id uint64) error
	IncreaseCommentCount(id uint64) error
	DecreaseCommentCount(id uint64) error
}

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository() PostRepository {
	return &postRepository{
		db: config.GetDB(),
	}
}

// 创建帖子
func (r *postRepository) Create(post *model.Post) error {
	return r.db.Create(post).Error
}

// 根据ID查询帖子
func (r *postRepository) FindByID(id uint64) (*model.Post, error) {
	var post model.Post
	err := r.db.Preload("User").First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// 根据用户ID查询帖子
func (r *postRepository) FindByUserID(userID uint64, page, size int) ([]*model.Post, int64, error) {
	var posts []*model.Post
	var total int64

	err := r.db.Model(&model.Post{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Preload("User").Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// 根据路由ID查询帖子
func (r *postRepository) FindByRouteID(routeID uint64, page, size int) ([]*model.Post, int64, error) {
	var posts []*model.Post
	var total int64

	err := r.db.Model(&model.Post{}).Where("route_id = ?", routeID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	// 根据创建时间降序查询帖子
	err = r.db.Preload("User").Where("route_id = ?", routeID).
		Order("created_at DESC"). //
		Offset((page - 1) * size).
		Limit(size).
		Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// 查询所有公开帖子
func (r *postRepository) FindAllPublic(page, size int) ([]*model.Post, int64, error) {
	var posts []*model.Post
	var total int64

	err := r.db.Model(&model.Post{}).Where("status = ?", 1).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Preload("User").Where("status = ?", 1).
		Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

func (r *postRepository) Update(post *model.Post) error {
	return r.db.Save(post).Error
}

func (r *postRepository) Delete(id uint64) error {
	return r.db.Delete(&model.Post{}, id).Error
}

//Todo:
//这里会出现并发问题，需要使用事务来解决，
// 例如：
// tx := r.db.Begin()
// tx.Model(&model.Post{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))
// tx.Commit()

// 增加帖子查看次数
func (r *postRepository) IncreaseViewCount(id uint64) error {
	return r.db.Model(&model.Post{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// 增加帖子点赞次数
func (r *postRepository) IncreaseLikeCount(id uint64) error {
	return r.db.Model(&model.Post{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error
}

// 减少帖子点赞次数
func (r *postRepository) DecreaseLikeCount(id uint64) error {
	return r.db.Model(&model.Post{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).Error
}

// 增加帖子点赞次数（事务版本）
func (r *postRepository) IncreaseLikeCountWithTx(tx *gorm.DB, id uint64) error {
	return tx.Model(&model.Post{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error
}

// 减少帖子点赞次数（事务版本）
func (r *postRepository) DecreaseLikeCountWithTx(tx *gorm.DB, id uint64) error {
	return tx.Model(&model.Post{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).Error
}

func (r *postRepository) IncreaseCommentCount(id uint64) error {
	return r.db.Model(&model.Post{}).Where("id = ?", id).UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).Error
}

// 减少帖子评论次数
func (r *postRepository) DecreaseCommentCount(id uint64) error {
	return r.db.Model(&model.Post{}).Where("id = ?", id).UpdateColumn("comment_count", gorm.Expr("comment_count - ?", 1)).Error
}
