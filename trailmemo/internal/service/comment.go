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

type CommentService interface {
	CreateComment(ctx context.Context, userID, postID, parentID uint64, content string) (*model.Comment, error)
	GetCommentByID(ctx context.Context, id uint64) (*model.Comment, error)
	GetCommentsByPostID(ctx context.Context, postID uint64, page, size int) ([]*model.Comment, int64, error)
	GetCommentsByUserID(ctx context.Context, userID uint64, page, size int) ([]*model.Comment, int64, error)
	UpdateComment(ctx context.Context, id, userID uint64, content string) error
	DeleteComment(ctx context.Context, id, userID uint64) error
}

type commentService struct {
	commentRepo repository.CommentRepository
	postRepo    repository.PostRepository
}

func NewCommentService() CommentService {
	return &commentService{
		commentRepo: repository.NewCommentRepository(),
		postRepo:    repository.NewPostRepository(),
	}
}

// 创建评论
func (s *commentService) CreateComment(ctx context.Context, userID, postID, parentID uint64, content string) (*model.Comment, error) {
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "comment"),
		zap.String("operation", "create_comment"),
		zap.Uint64("user_id", userID),
		zap.Uint64("post_id", postID),
	)

	if content == "" {
		log.Warn("validation_failed",
			zap.String("event", "validation_failed"),
			zap.String("error_code", apperror.CodeInvalidParams),
			zap.String("error_kind", apperror.KindValidation),
		)
		return nil, apperror.New(apperror.CodeInvalidParams, apperror.KindValidation, "content is required", http.StatusBadRequest)
	}

	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		log.Error("db_query_failed",
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return nil, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询帖子失败", http.StatusInternalServerError)
	}
	if post == nil {
		log.Warn("post_not_found",
			zap.String("event", "post_not_found"),
		)
		return nil, apperror.New(apperror.CodePostNotFound, apperror.KindNotFound, "post not found", http.StatusNotFound)
	}

	comment := &model.Comment{
		UserID:   userID,
		PostID:   postID,
		ParentID: parentID,
		Content:  content,
	}

	if err := s.commentRepo.Create(comment); err != nil {
		log.Error("comment_create_failed",
			zap.String("event", "comment_create_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return nil, apperror.Wrap(err, apperror.CodeCommentCreateFailed, apperror.KindDB, "创建评论失败", http.StatusInternalServerError)
	}

	err = s.postRepo.IncreaseCommentCount(postID)
	if err != nil {
		log.Warn("increase_comment_count_failed",
			zap.String("event", "increase_comment_count_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
	}

	created, err := s.commentRepo.FindByID(comment.ID)
	if err != nil {
		log.Error("db_query_failed",
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return nil, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询评论详情失败", http.StatusInternalServerError)
	}

	log.Info("comment_created",
		zap.String("event", "comment_created"),
		zap.String("entity_type", "comment"),
		zap.Uint64("entity_id", comment.ID),
	)

	return created, nil
}

// 根据评论ID查询评论
func (s *commentService) GetCommentByID(ctx context.Context, id uint64) (*model.Comment, error) {
	comment, err := s.commentRepo.FindByID(id)
	if err != nil {
		platformlogger.FromContext(ctx).Error("db_query_failed",
			zap.String("module", "comment"),
			zap.String("operation", "get_comment_by_id"),
			zap.Uint64("comment_id", id),
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return nil, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询评论失败", http.StatusInternalServerError)
	}
	return comment, nil
}

// 根据帖子ID查询帖子评论
func (s *commentService) GetCommentsByPostID(ctx context.Context, postID uint64, page, size int) ([]*model.Comment, int64, error) {
	comments, total, err := s.commentRepo.FindByPostID(postID, page, size)
	if err != nil {
		platformlogger.FromContext(ctx).Error("db_query_failed",
			zap.String("module", "comment"),
			zap.String("operation", "get_comments_by_post_id"),
			zap.Uint64("post_id", postID),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return nil, 0, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询评论列表失败", http.StatusInternalServerError)
	}
	return comments, total, nil
}

// 根据用户ID查询用户评论
func (s *commentService) GetCommentsByUserID(ctx context.Context, userID uint64, page, size int) ([]*model.Comment, int64, error) {
	comments, total, err := s.commentRepo.FindByUserID(userID, page, size)
	if err != nil {
		platformlogger.FromContext(ctx).Error("db_query_failed",
			zap.String("module", "comment"),
			zap.String("operation", "get_comments_by_user_id"),
			zap.Uint64("user_id", userID),
			zap.Int("page", page),
			zap.Int("size", size),
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return nil, 0, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询用户评论列表失败", http.StatusInternalServerError)
	}
	return comments, total, nil
}

// 更新评论
func (s *commentService) UpdateComment(ctx context.Context, id, userID uint64, content string) error {
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "comment"),
		zap.String("operation", "update_comment"),
		zap.Uint64("comment_id", id),
		zap.Uint64("user_id", userID),
	)

	comment, err := s.commentRepo.FindByID(id)
	if err != nil {
		log.Error("db_query_failed",
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询评论失败", http.StatusInternalServerError)
	}
	if comment == nil {
		log.Warn("comment_not_found",
			zap.String("event", "comment_not_found"),
		)
		return apperror.New(apperror.CodeCommentNotFound, apperror.KindNotFound, "comment not found", http.StatusNotFound)
	}
	if comment.UserID != userID {
		log.Warn("permission_denied",
			zap.String("event", "permission_denied"),
			zap.Uint64("comment_user_id", comment.UserID),
		)
		return apperror.New(apperror.CodeCommentPermission, apperror.KindPermission, "permission denied", http.StatusForbidden)
	}

	if content != "" {
		comment.Content = content
	}

	if err := s.commentRepo.Update(comment); err != nil {
		log.Error("comment_update_failed",
			zap.String("event", "comment_update_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return apperror.Wrap(err, apperror.CodeCommentUpdateFailed, apperror.KindDB, "更新评论失败", http.StatusInternalServerError)
	}

	log.Info("comment_updated",
		zap.String("event", "comment_updated"),
		zap.String("entity_type", "comment"),
		zap.Uint64("entity_id", id),
	)

	return nil
}

func (s *commentService) DeleteComment(ctx context.Context, id, userID uint64) error {
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "comment"),
		zap.String("operation", "delete_comment"),
		zap.Uint64("comment_id", id),
		zap.Uint64("user_id", userID),
	)

	comment, err := s.commentRepo.FindByID(id)
	if err != nil {
		log.Error("db_query_failed",
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询评论失败", http.StatusInternalServerError)
	}
	if comment == nil {
		log.Warn("comment_not_found",
			zap.String("event", "comment_not_found"),
		)
		return apperror.New(apperror.CodeCommentNotFound, apperror.KindNotFound, "comment not found", http.StatusNotFound)
	}
	if comment.UserID != userID {
		log.Warn("permission_denied",
			zap.String("event", "permission_denied"),
			zap.Uint64("comment_user_id", comment.UserID),
		)
		return apperror.New(apperror.CodeCommentPermission, apperror.KindPermission, "permission denied", http.StatusForbidden)
	}

	err = s.postRepo.DecreaseCommentCount(comment.PostID)
	if err != nil {
		log.Warn("decrease_comment_count_failed",
			zap.String("event", "decrease_comment_count_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
	}

	if err := s.commentRepo.Delete(id); err != nil {
		log.Error("comment_delete_failed",
			zap.String("event", "comment_delete_failed"),
			zap.String("error_kind", apperror.KindDB),
			zap.Error(err),
		)
		return apperror.Wrap(err, apperror.CodeCommentDeleteFailed, apperror.KindDB, "删除评论失败", http.StatusInternalServerError)
	}

	log.Info("comment_deleted",
		zap.String("event", "comment_deleted"),
		zap.String("entity_type", "comment"),
		zap.Uint64("entity_id", id),
	)

	return nil
}
