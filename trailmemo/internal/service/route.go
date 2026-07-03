package service

import (
	"context"
	"net/http"

	"github.com/trailmemo/internal/model"
	platformlogger "github.com/trailmemo/internal/platform/logger"
	"github.com/trailmemo/internal/repository"
	"github.com/trailmemo/pkg/apperror"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RouteService interface {
	CreateRoute(ctx context.Context, userID uint64, title, description, coverImage, startCity, endCity string,
		totalDistance, estimatedHours float64, isPublic int, checkpoints []*model.Checkpoint) (*model.Route, error)
	GetRouteByID(ctx context.Context, id uint64) (*model.Route, error)
	GetRoutesByUserID(ctx context.Context, userID uint64, page, size int) ([]*model.Route, int64, error)
	GetPublicRoutes(ctx context.Context, page, size int) ([]*model.Route, int64, error)
	UpdateRoute(ctx context.Context, id, userID uint64, title, description, coverImage, startCity, endCity string,
		totalDistance, estimatedHours float64, isPublic int) error
	DeleteRoute(ctx context.Context, id, userID uint64) error
	CopyRoute(ctx context.Context, userID, routeID uint64, isPublic int) (*model.Route, error)
}

type routeService struct {
	routeRepo      repository.RouteRepository
	checkpointRepo repository.CheckpointRepository
}

func NewRouteService() RouteService {
	return &routeService{
		routeRepo:      repository.NewRouteRepository(),
		checkpointRepo: repository.NewCheckpointRepository(),
	}
}

func (s *routeService) CreateRoute(ctx context.Context, userID uint64, title, description, coverImage, startCity, endCity string,
	totalDistance, estimatedHours float64, isPublic int, checkpoints []*model.Checkpoint) (*model.Route, error) {
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "route"),
		zap.String("operation", "create_route"),
		zap.Uint64("user_id", userID),
	)

	if title == "" {
		return nil, apperror.New(apperror.CodeInvalidParams, apperror.KindValidation, "路线标题不能为空", http.StatusBadRequest)
	}
	if startCity == "" {
		return nil, apperror.New(apperror.CodeInvalidParams, apperror.KindValidation, "起点城市不能为空", http.StatusBadRequest)
	}
	if endCity == "" {
		return nil, apperror.New(apperror.CodeInvalidParams, apperror.KindValidation, "终点城市不能为空", http.StatusBadRequest)
	}

	route := &model.Route{
		UserID:         userID,         // 用户ID
		Title:          title,          // 路线标题
		Description:    description,    // 路线描述
		CoverImage:     coverImage,     // 路线封面图片
		StartCity:      startCity,      // 起点城市
		EndCity:        endCity,        // 终点城市
		TotalDistance:  totalDistance,  // 总距离
		EstimatedHours: estimatedHours, // 估计时间
		IsPublic:       isPublic,       // 是否公开
	}

	// 开启事务，确保路线和路点数据一致性
	tx := s.routeRepo.BeginTx(ctx) // 开启事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	log.Debug("tx_begin", zap.String("event", "tx_begin"))

	// 创建路线
	if err := s.routeRepo.CreateWithTxContext(ctx, tx, route); err != nil {
		tx.Rollback()
		log.Error("tx_rollback",
			zap.String("event", "tx_rollback"),
			zap.String("error_kind", "db"),
			zap.Error(err),
		)
		return nil, apperror.Wrap(err, apperror.CodeRouteCreateFailed, apperror.KindDB, "创建路线失败", http.StatusInternalServerError)
	}

	// 关联路线ID到路点
	for _, cp := range checkpoints {
		cp.RouteID = route.ID
	}

	// 批量创建路点
	if len(checkpoints) > 0 {
		if err := s.checkpointRepo.CreateBatchWithTxContext(ctx, tx, checkpoints); err != nil {
			tx.Rollback()
			log.Error("tx_rollback",
				zap.String("event", "tx_rollback"),
				zap.String("error_kind", "db"),
				zap.Error(err),
			)
			return nil, apperror.Wrap(err, apperror.CodeRouteCreateFailed, apperror.KindDB, "创建路线打卡点失败", http.StatusInternalServerError)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Error("tx_commit_failed",
			zap.String("event", "tx_commit_failed"),
			zap.String("error_kind", "db"),
			zap.Error(err),
		)
		return nil, apperror.Wrap(err, apperror.CodeRouteCreateFailed, apperror.KindDB, "创建路线失败", http.StatusInternalServerError)
	}

	return s.routeRepo.FindByIDContext(ctx, route.ID)
}

func (s *routeService) GetRouteByID(ctx context.Context, id uint64) (*model.Route, error) {
	route, err := s.routeRepo.FindByIDContext(ctx, id)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询路线失败", http.StatusInternalServerError)
	}
	return route, nil
}

func (s *routeService) GetRoutesByUserID(ctx context.Context, userID uint64, page, size int) ([]*model.Route, int64, error) {
	routes, total, err := s.routeRepo.FindByUserIDContext(ctx, userID, page, size)
	if err != nil {
		return nil, 0, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询路线列表失败", http.StatusInternalServerError)
	}
	return routes, total, nil
}

func (s *routeService) GetPublicRoutes(ctx context.Context, page, size int) ([]*model.Route, int64, error) {
	routes, total, err := s.routeRepo.FindPublicContext(ctx, page, size)
	if err != nil {
		return nil, 0, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询公开路线失败", http.StatusInternalServerError)
	}
	return routes, total, nil
}

func (s *routeService) UpdateRoute(ctx context.Context, id, userID uint64, title, description, coverImage, startCity, endCity string,
	totalDistance, estimatedHours float64, isPublic int) error {

	route, err := s.routeRepo.FindByIDContext(ctx, id)
	if err != nil {
		return apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询路线失败", http.StatusInternalServerError)
	}
	if route == nil {
		return apperror.New(apperror.CodeRouteNotFound, apperror.KindNotFound, "路线不存在", http.StatusNotFound)
	}
	if route.UserID != userID {
		return apperror.New(apperror.CodeRoutePermission, apperror.KindPermission, "无权操作该路线", http.StatusForbidden)
	}

	if title != "" {
		route.Title = title
	}
	if description != "" {
		route.Description = description
	}
	if coverImage != "" {
		route.CoverImage = coverImage
	}
	if startCity != "" {
		route.StartCity = startCity
	}
	if endCity != "" {
		route.EndCity = endCity
	}
	if totalDistance > 0 {
		route.TotalDistance = totalDistance
	}
	if estimatedHours > 0 {
		route.EstimatedHours = estimatedHours
	}
	if isPublic >= 0 {
		route.IsPublic = isPublic
	}

	if err := s.routeRepo.UpdateContext(ctx, route); err != nil {
		return apperror.Wrap(err, apperror.CodeRouteUpdateFailed, apperror.KindDB, "更新路线失败", http.StatusInternalServerError)
	}
	return nil
}

func (s *routeService) DeleteRoute(ctx context.Context, id, userID uint64) error {
	route, err := s.routeRepo.FindByIDContext(ctx, id)
	if err != nil {
		return apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询路线失败", http.StatusInternalServerError)
	}
	if route == nil {
		return apperror.New(apperror.CodeRouteNotFound, apperror.KindNotFound, "路线不存在", http.StatusNotFound)
	}
	if route.UserID != userID {
		return apperror.New(apperror.CodeRoutePermission, apperror.KindPermission, "无权操作该路线", http.StatusForbidden)
	}

	if err := s.checkpointRepo.DeleteByRouteIDContext(ctx, id); err != nil {
		return apperror.Wrap(err, apperror.CodeRouteDeleteFailed, apperror.KindDB, "删除路线打卡点失败", http.StatusInternalServerError)
	}

	if err := s.routeRepo.DeleteContext(ctx, id); err != nil {
		return apperror.Wrap(err, apperror.CodeRouteDeleteFailed, apperror.KindDB, "删除路线失败", http.StatusInternalServerError)
	}
	return nil
}

func (s *routeService) CopyRoute(ctx context.Context, userID, routeID uint64, isPublic int) (*model.Route, error) {
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "route"),
		zap.String("operation", "copy_route"),
		zap.Uint64("user_id", userID),
		zap.Uint64("route_id", routeID),
	)

	sourceRoute, err := s.routeRepo.FindByIDContext(ctx, routeID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询路线失败", http.StatusInternalServerError)
	}
	if sourceRoute == nil {
		return nil, apperror.New(apperror.CodeRouteNotFound, apperror.KindNotFound, "路线不存在", http.StatusNotFound)
	}
	if sourceRoute.IsPublic != 1 {
		return nil, apperror.New(apperror.CodeRoutePermission, apperror.KindPermission, "只能复用公开路线", http.StatusForbidden)
	}

	checkpoints, err := s.checkpointRepo.FindByRouteIDContext(ctx, routeID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeDBError, apperror.KindDB, "查询路线打卡点失败", http.StatusInternalServerError)
	}

	newRoute := &model.Route{
		UserID:         userID,
		Title:          sourceRoute.Title + " (复用)",
		Description:    sourceRoute.Description,
		CoverImage:     sourceRoute.CoverImage,
		StartCity:      sourceRoute.StartCity,
		EndCity:        sourceRoute.EndCity,
		TotalDistance:  sourceRoute.TotalDistance,
		EstimatedHours: sourceRoute.EstimatedHours,
		IsPublic:       isPublic,
	}

	// 开启事务，确保复制操作的原子性
	tx := s.routeRepo.BeginTx(ctx)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	log.Debug("tx_begin", zap.String("event", "tx_begin"))

	// 创建新路线
	if err := s.routeRepo.CreateWithTxContext(ctx, tx, newRoute); err != nil {
		tx.Rollback()
		log.Error("tx_rollback", zap.String("event", "tx_rollback"), zap.String("error_kind", "db"), zap.Error(err))
		return nil, apperror.Wrap(err, apperror.CodeRouteCopyFailed, apperror.KindDB, "复用路线失败", http.StatusInternalServerError)
	}

	// 创建新路点
	newCheckpoints := make([]*model.Checkpoint, 0, len(checkpoints))
	for _, cp := range checkpoints {
		newCP := &model.Checkpoint{
			RouteID:      newRoute.ID,
			Name:         cp.Name,
			Description:  cp.Description,
			Latitude:     cp.Latitude,
			Longitude:    cp.Longitude,
			Address:      cp.Address,
			City:         cp.City,
			Sequence:     cp.Sequence,
			ArriveTime:   cp.ArriveTime,
			StayDuration: cp.StayDuration,
			PhotoURL:     cp.PhotoURL,
		}
		newCheckpoints = append(newCheckpoints, newCP)
	}

	// 批量创建路点
	if len(newCheckpoints) > 0 {
		if err := s.checkpointRepo.CreateBatchWithTxContext(ctx, tx, newCheckpoints); err != nil {
			tx.Rollback()
			log.Error("tx_rollback", zap.String("event", "tx_rollback"), zap.String("error_kind", "db"), zap.Error(err))
			return nil, apperror.Wrap(err, apperror.CodeRouteCopyFailed, apperror.KindDB, "复用路线打卡点失败", http.StatusInternalServerError)
		}
	}

	// 更新原路线复用计数
	if err := tx.Model(&model.Route{}).Where("id = ?", routeID).UpdateColumn("reuse_count", gorm.Expr("reuse_count + ?", 1)).Error; err != nil {
		tx.Rollback()
		log.Error("tx_rollback", zap.String("event", "tx_rollback"), zap.String("error_kind", "db"), zap.Error(err))
		return nil, apperror.Wrap(err, apperror.CodeRouteCopyFailed, apperror.KindDB, "更新路线复用次数失败", http.StatusInternalServerError)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Error("tx_commit_failed", zap.String("event", "tx_commit_failed"), zap.String("error_kind", "db"), zap.Error(err))
		return nil, apperror.Wrap(err, apperror.CodeRouteCopyFailed, apperror.KindDB, "复用路线失败", http.StatusInternalServerError)
	}

	return s.routeRepo.FindByIDContext(ctx, newRoute.ID)
}
