package service

import (
	"context"
	"net/http"

	platformlogger "github.com/trailmemo/internal/platform/logger"
	"github.com/trailmemo/pkg/apperror"
	"go.uber.org/zap"

	"github.com/trailmemo/internal/model"
	"github.com/trailmemo/internal/repository"
)

type PostService interface {
	CreatePost(ctx context.Context, userID, routeID uint64, title, content, images string) (*model.Post, error)
	GetPostByID(ctx context.Context, id uint64) (*model.Post, error)
	GetPostsByUserID(ctx context.Context, userID uint64, page, size int) ([]*model.Post, int64, error)
	GetPostsByRouteID(ctx context.Context, routeID uint64, page, size int) ([]*model.Post, int64, error)
	GetAllPublicPosts(ctx context.Context, page, size int) ([]*model.Post, int64, error)
	UpdatePost(ctx context.Context, id, userID uint64, title, content, images string) error
	DeletePost(ctx context.Context, id, userID uint64) error
}

type postService struct {
	postRepo repository.PostRepository
}

func NewPostService() PostService {
	return &postService{
		postRepo: repository.NewPostRepository(),
	}
}

// 创建帖子
func (s *postService) CreatePost(ctx context.Context, userID, routeID uint64, title, content, images string) (*model.Post, error) {
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "post"),
		zap.String("operation", "create_post"),
		zap.Uint64("user_id", userID),
		zap.Uint64("route_id", routeID),
	)

	if title == "" || content == "" {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
		)
		return nil, apperror.New(apperror.CodeInvalidParams, apperror.KindValidation, "title and content are required", http.StatusBadRequest)
	}

	post := &model.Post{
		UserID:  userID,
		RouteID: routeID,
		Title:   title,
		Content: content,
		Images:  images,
	}

	if err := s.postRepo.Create(post); err != nil {
		log.Error("post_create_failed",
			zap.String("event", "post_create_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return nil, apperror.Wrap(err, apperror.CodePostCreateFailed, apperror.KindDB, "创建帖子失败", http.StatusInternalServerError)
	}

	created, err := s.postRepo.FindByID(post.ID)
	if err != nil {
		log.Error("db_query_failed",
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return nil, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询帖子详情失败", http.StatusInternalServerError)
	}

	log.Info("post_created",
		zap.String("event", "post_created"),
		zap.String("entity_type", "post"),
		zap.Uint64("entity_id", post.ID),
	)

	return created, nil
}

func (s *postService) GetPostByID(ctx context.Context, id uint64) (*model.Post, error) {
	post, err := s.postRepo.FindByID(id)
	if err != nil {
		platformlogger.FromContext(ctx).Error("db_query_failed",
			zap.String("module", "post"),
			zap.String("operation", "get_post_by_id"),
			zap.Uint64("post_id", id),
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return nil, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询帖子失败", http.StatusInternalServerError)
	}

	if err := s.postRepo.IncreaseViewCount(id); err != nil {
		platformlogger.FromContext(ctx).Warn("increase_view_count_failed",
			zap.String("module", "post"),
			zap.String("operation", "get_post_by_id"),
			zap.Uint64("post_id", id),
			zap.String("event", "increase_view_count_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
	}

	return post, nil
}

func (s *postService) GetPostsByUserID(ctx context.Context, userID uint64, page, size int) ([]*model.Post, int64, error) {
	posts, total, err := s.postRepo.FindByUserID(userID, page, size)
	if err != nil {
		platformlogger.FromContext(ctx).Error("db_query_failed",
			zap.String("module", "post"),
			zap.String("operation", "get_posts_by_user_id"),
			zap.Uint64("user_id", userID),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return nil, 0, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询用户帖子列表失败", http.StatusInternalServerError)
	}
	return posts, total, nil
}

func (s *postService) GetPostsByRouteID(ctx context.Context, routeID uint64, page, size int) ([]*model.Post, int64, error) {
	posts, total, err := s.postRepo.FindByRouteID(routeID, page, size)
	if err != nil {
		platformlogger.FromContext(ctx).Error("db_query_failed",
			zap.String("module", "post"),
			zap.String("operation", "get_posts_by_route_id"),
			zap.Uint64("route_id", routeID),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return nil, 0, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询路线帖子列表失败", http.StatusInternalServerError)
	}
	return posts, total, nil
}

// 获取所有公开帖子
func (s *postService) GetAllPublicPosts(ctx context.Context, page, size int) ([]*model.Post, int64, error) {
	posts, total, err := s.postRepo.FindAllPublic(page, size)
	if err != nil {
		platformlogger.FromContext(ctx).Error("db_query_failed",
			zap.String("module", "post"),
			zap.String("operation", "get_all_public_posts"),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return nil, 0, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询公开帖子列表失败", http.StatusInternalServerError)
	}
	return posts, total, nil
}

func (s *postService) UpdatePost(ctx context.Context, id, userID uint64, title, content, images string) error {
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "post"),
		zap.String("operation", "update_post"),
		zap.Uint64("post_id", id),
		zap.Uint64("user_id", userID),
	)

	post, err := s.postRepo.FindByID(id)
	if err != nil {
		log.Error("db_query_failed",
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询帖子失败", http.StatusInternalServerError)
	}
	if post == nil {
		log.Warn("post_not_found",
			zap.String("event", "post_not_found"),
		)
		return apperror.New(apperror.CodePostNotFound, apperror.KindNotFound, "post not found", http.StatusNotFound)
	}
	if post.UserID != userID {
		log.Warn("permission_denied",
			zap.String("event", "permission_denied"),
			zap.Uint64("post_user_id", post.UserID),
		)
		return apperror.New(apperror.CodePostPermission, apperror.KindPermission, "permission denied", http.StatusForbidden)
	}

	if title != "" {
		post.Title = title
	}
	if content != "" {
		post.Content = content
	}
	if images != "" {
		post.Images = images
	}

	if err := s.postRepo.Update(post); err != nil {
		log.Error("post_update_failed",
			zap.String("event", "post_update_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return apperror.Wrap(err, apperror.CodePostUpdateFailed, apperror.KindDB, "更新帖子失败", http.StatusInternalServerError)
	}

	log.Info("post_updated",
		zap.String("event", "post_updated"),
		zap.String("entity_type", "post"),
		zap.Uint64("entity_id", id),
	)

	return nil
}

func (s *postService) DeletePost(ctx context.Context, id, userID uint64) error {
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "post"),
		zap.String("operation", "delete_post"),
		zap.Uint64("post_id", id),
		zap.Uint64("user_id", userID),
	)

	post, err := s.postRepo.FindByID(id)
	if err != nil {
		log.Error("db_query_failed",
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询帖子失败", http.StatusInternalServerError)
	}
	if post == nil {
		log.Warn("post_not_found",
			zap.String("event", "post_not_found"),
		)
		return apperror.New(apperror.CodePostNotFound, apperror.KindNotFound, "post not found", http.StatusNotFound)
	}
	if post.UserID != userID {
		log.Warn("permission_denied",
			zap.String("event", "permission_denied"),
			zap.Uint64("post_user_id", post.UserID),
		)
		return apperror.New(apperror.CodePostPermission, apperror.KindPermission, "permission denied", http.StatusForbidden)
	}

	if err := s.postRepo.Delete(id); err != nil {
		log.Error("post_delete_failed",
			zap.String("event", "post_delete_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return apperror.Wrap(err, apperror.CodePostDeleteFailed, apperror.KindDB, "删除帖子失败", http.StatusInternalServerError)
	}

	log.Info("post_deleted",
		zap.String("event", "post_deleted"),
		zap.String("entity_type", "post"),
		zap.Uint64("entity_id", id),
	)

	return nil
}
