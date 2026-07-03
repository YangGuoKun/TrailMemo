package service

import (
	"context"
	"net/http"
	"time"

	"github.com/trailmemo/internal/model"
	platformlogger "github.com/trailmemo/internal/platform/logger"
	"github.com/trailmemo/internal/repository"
	"github.com/trailmemo/pkg/apperror"
	"go.uber.org/zap"
)

type CheckinService interface {
	CreateCheckin(ctx context.Context, userID, routeID, checkpointID uint64, latitude, longitude float64,
		photoURL, content string, rating int) (*model.Checkin, error)
	GetCheckinByID(ctx context.Context, id uint64) (*model.Checkin, error)
	GetCheckinsByUserID(ctx context.Context, userID uint64, page, size int) ([]*model.Checkin, int64, error)
	GetCheckinsByRouteID(ctx context.Context, routeID uint64, page, size int) ([]*model.Checkin, int64, error)
	UpdateCheckin(ctx context.Context, id, userID uint64, photoURL, content string, rating int) error
	DeleteCheckin(ctx context.Context, id, userID uint64) error
	GetRouteProgress(ctx context.Context, userID, routeID uint64) (float64, error)
}

type checkinService struct {
	checkinRepo    repository.CheckinRepository
	checkpointRepo repository.CheckpointRepository
	routeRepo      repository.RouteRepository
}

func NewCheckinService() CheckinService {
	return &checkinService{
		checkinRepo:    repository.NewCheckinRepository(),
		checkpointRepo: repository.NewCheckpointRepository(),
		routeRepo:      repository.NewRouteRepository(),
	}
}

func (s *checkinService) CreateCheckin(ctx context.Context, userID, routeID, checkpointID uint64, latitude, longitude float64,
	photoURL, content string, rating int) (*model.Checkin, error) {
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "checkin"),
		zap.String("operation", "create_checkin"),
		zap.Uint64("user_id", userID),
		zap.Uint64("route_id", routeID),
		zap.Uint64("checkpoint_id", checkpointID),
	)

	checkpoint, err := s.checkpointRepo.FindByIDContext(ctx, checkpointID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询打卡点失败", http.StatusInternalServerError)
	}
	if checkpoint == nil {
		return nil, apperror.New(apperror.CodeCheckpointNotFound, apperror.KindNotFound, "打卡点不存在", http.StatusNotFound)
	}

	if checkpoint.RouteID != routeID {
		return nil, apperror.New(apperror.CodeCheckinConflict, apperror.KindConflict, "打卡点不属于该路线", http.StatusConflict)
	}

	exists, err := s.checkinRepo.CheckCheckinExistsContext(ctx, userID, checkpointID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询打卡状态失败", http.StatusInternalServerError)
	}
	if exists {
		return nil, apperror.New(apperror.CodeCheckinConflict, apperror.KindConflict, "该打卡点已打卡", http.StatusConflict)
	}

	checkin := &model.Checkin{
		UserID:       userID,
		RouteID:      routeID,
		CheckpointID: checkpointID,
		CheckinTime:  time.Now(),
		Latitude:     latitude,
		Longitude:    longitude,
		PhotoURL:     photoURL,
		Content:      content,
		Rating:       rating,
	}

	if err := s.checkinRepo.CreateContext(ctx, checkin); err != nil {
		return nil, apperror.Wrap(err, apperror.CodeCheckinCreateFailed, apperror.KindDB, "创建打卡失败", http.StatusInternalServerError)
	}

	if err := s.updateRouteProgress(ctx, userID, routeID); err != nil {
		log.Error("route_progress_update_failed",
			zap.String("event", "route_progress_update_failed"),
			zap.String("error_kind", apperror.From(err).Kind),
			zap.Error(err),
		)
		return nil, err
	}

	created, err := s.checkinRepo.FindByIDContext(ctx, checkin.ID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询打卡详情失败", http.StatusInternalServerError)
	}
	return created, nil
}

func (s *checkinService) GetCheckinByID(ctx context.Context, id uint64) (*model.Checkin, error) {
	checkin, err := s.checkinRepo.FindByIDContext(ctx, id)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询打卡详情失败", http.StatusInternalServerError)
	}
	return checkin, nil
}

func (s *checkinService) GetCheckinsByUserID(ctx context.Context, userID uint64, page, size int) ([]*model.Checkin, int64, error) {
	checkins, total, err := s.checkinRepo.FindByUserIDContext(ctx, userID, page, size)
	if err != nil {
		return nil, 0, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询打卡列表失败", http.StatusInternalServerError)
	}
	return checkins, total, nil
}

func (s *checkinService) GetCheckinsByRouteID(ctx context.Context, routeID uint64, page, size int) ([]*model.Checkin, int64, error) {
	checkins, total, err := s.checkinRepo.FindByRouteIDContext(ctx, routeID, page, size)
	if err != nil {
		return nil, 0, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询路线打卡列表失败", http.StatusInternalServerError)
	}
	return checkins, total, nil
}

func (s *checkinService) UpdateCheckin(ctx context.Context, id, userID uint64, photoURL, content string, rating int) error {
	checkin, err := s.checkinRepo.FindByIDContext(ctx, id)
	if err != nil {
		return apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询打卡详情失败", http.StatusInternalServerError)
	}
	if checkin == nil {
		return apperror.New(apperror.CodeCheckinNotFound, apperror.KindNotFound, "打卡记录不存在", http.StatusNotFound)
	}
	if checkin.UserID != userID {
		return apperror.New(apperror.CodeCheckinPermission, apperror.KindPermission, "无权操作该打卡记录", http.StatusForbidden)
	}

	if photoURL != "" {
		checkin.PhotoURL = photoURL
	}
	if content != "" {
		checkin.Content = content
	}
	if rating > 0 && rating <= 5 {
		checkin.Rating = rating
	}

	if err := s.checkinRepo.UpdateContext(ctx, checkin); err != nil {
		return apperror.Wrap(err, apperror.CodeCheckinUpdateFailed, apperror.KindDB, "更新打卡失败", http.StatusInternalServerError)
	}
	return nil
}

func (s *checkinService) DeleteCheckin(ctx context.Context, id, userID uint64) error {
	checkin, err := s.checkinRepo.FindByIDContext(ctx, id)
	if err != nil {
		return apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询打卡详情失败", http.StatusInternalServerError)
	}
	if checkin == nil {
		return apperror.New(apperror.CodeCheckinNotFound, apperror.KindNotFound, "打卡记录不存在", http.StatusNotFound)
	}
	if checkin.UserID != userID {
		return apperror.New(apperror.CodeCheckinPermission, apperror.KindPermission, "无权操作该打卡记录", http.StatusForbidden)
	}

	if err := s.checkinRepo.DeleteContext(ctx, id); err != nil {
		return apperror.Wrap(err, apperror.CodeCheckinDeleteFailed, apperror.KindDB, "删除打卡失败", http.StatusInternalServerError)
	}

	// 删除后同步路线完成进度，避免前端看到过期状态。
	if err := s.updateRouteProgress(ctx, userID, checkin.RouteID); err != nil {
		platformlogger.FromContext(ctx).Error("route_progress_update_failed",
			zap.String("event", "route_progress_update_failed"),
			zap.String("module", "checkin"),
			zap.String("operation", "delete_checkin"),
			zap.Uint64("user_id", userID),
			zap.Uint64("route_id", checkin.RouteID),
			zap.Uint64("checkin_id", id),
			zap.String("error_kind", apperror.From(err).Kind),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (s *checkinService) GetRouteProgress(ctx context.Context, userID, routeID uint64) (float64, error) {
	route, err := s.routeRepo.FindByIDContext(ctx, routeID)
	if err != nil {
		return 0, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询路线失败", http.StatusInternalServerError)
	}
	if route == nil {
		return 0, apperror.New(apperror.CodeRouteNotFound, apperror.KindNotFound, "路线不存在", http.StatusNotFound)
	}

	checkpoints, err := s.checkpointRepo.FindByRouteIDContext(ctx, routeID)
	if err != nil {
		return 0, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询路线打卡点失败", http.StatusInternalServerError)
	}
	if len(checkpoints) == 0 {
		return 0, nil
	}

	checkinCount, err := s.checkinRepo.GetCheckinCountByRouteContext(ctx, userID, routeID)
	if err != nil {
		return 0, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询打卡进度失败", http.StatusInternalServerError)
	}

	return float64(checkinCount) / float64(len(checkpoints)) * 100, nil
}

func (s *checkinService) updateRouteProgress(ctx context.Context, userID, routeID uint64) error {
	checkpoints, err := s.checkpointRepo.FindByRouteIDContext(ctx, routeID)
	if err != nil {
		return apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询路线打卡点失败", http.StatusInternalServerError)
	}
	if len(checkpoints) == 0 {
		return nil
	}

	checkinCount, err := s.checkinRepo.GetCheckinCountByRouteContext(ctx, userID, routeID)
	if err != nil {
		return apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询打卡进度失败", http.StatusInternalServerError)
	}

	route, err := s.routeRepo.FindByIDContext(ctx, routeID)
	if err != nil {
		return apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询路线失败", http.StatusInternalServerError)
	}
	if route == nil {
		return apperror.New(apperror.CodeRouteNotFound, apperror.KindNotFound, "路线不存在", http.StatusNotFound)
	}

	if checkinCount >= int64(len(checkpoints)) {
		route.PublishStatus = 2
	} else if checkinCount > 0 {
		route.PublishStatus = 1
	} else {
		route.PublishStatus = 0
	}

	if err := s.routeRepo.UpdateContext(ctx, route); err != nil {
		return apperror.Wrap(err, apperror.CodeRouteUpdateFailed, apperror.KindDB, "更新路线进度失败", http.StatusInternalServerError)
	}
	return nil
}
