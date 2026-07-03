package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/trailmemo/internal/config"
	"github.com/trailmemo/internal/middleware"
	"github.com/trailmemo/internal/model"
	platformlogger "github.com/trailmemo/internal/platform/logger"
	"github.com/trailmemo/internal/platform/wechat"
	"github.com/trailmemo/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, username, password, phone, email string) (*model.User, error)
	Login(ctx context.Context, username, password string) (string, error)
	WechatLogin(ctx context.Context, code string) (string, error)
	GetUserInfo(ctx context.Context, userID uint64) (*model.User, error)
	UpdateUserInfo(ctx context.Context, userID uint64, nickname, avatar string) error
	ChangePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService() UserService {
	return &userService{
		repo: repository.NewUserRepository(),
	}
}

func (s *userService) Register(ctx context.Context, username, password, phone, email string) (*model.User, error) {
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "user"),
		zap.String("operation", "register"),
		zap.String("username", username),
	)

	// 检查用户名是否存在
	existing, err := s.repo.FindByUsername(username)
	if err != nil {
		log.Error("db_query_failed",
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", "db"),
			zap.Error(err),
		)
		return nil, err
	}
	if existing != nil {
		log.Warn("username_already_exists",
			zap.String("event", "username_already_exists"),
		)
		return nil, errors.New("username already exists")
	}

	// 检查手机号是否存在
	if phone != "" {
		existing, err = s.repo.FindByPhone(phone)
		if err != nil {
			log.Error("db_query_failed",
				zap.String("event", "db_query_failed"),
				zap.String("field", "phone"),
				zap.String("error_kind", "db"),
				zap.Error(err),
			)
			return nil, err
		}
		if existing != nil {
			log.Warn("phone_already_registered",
				zap.String("event", "phone_already_registered"),
				zap.String("phone", phone),
			)
			return nil, errors.New("phone already registered")
		}
	}

	// 检查邮箱是否存在
	if email != "" {
		existing, err = s.repo.FindByEmail(email)
		if err != nil {
			log.Error("db_query_failed",
				zap.String("event", "db_query_failed"),
				zap.String("field", "email"),
				zap.String("error_kind", "db"),
				zap.Error(err),
			)
			return nil, err
		}
		if existing != nil {
			log.Warn("email_already_registered",
				zap.String("event", "email_already_registered"),
				zap.String("email", email),
			)
			return nil, errors.New("email already registered")
		}
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("password_hash_failed",
			zap.String("event", "password_hash_failed"),
			zap.String("error_kind", "internal"),
			zap.Error(err),
		)
		return nil, err
	}

	// 创建用户
	user := &model.User{
		Username: username,
		Password: string(hashedPassword),
		Phone:    strPtr(phone),
		Email:    strPtr(email),
		Status:   1,
	}

	if err := s.repo.Create(user); err != nil {
		log.Error("user_create_failed",
			zap.String("event", "user_create_failed"),
			zap.String("error_kind", "db"),
			zap.Error(err),
		)
		return nil, err
	}

	log.Info("user_registered",
		zap.String("event", "user_registered"),
		zap.String("entity_type", "user"),
		zap.Uint64("entity_id", user.ID),
	)

	return user, nil
}

func (s *userService) Login(ctx context.Context, username, password string) (string, error) {
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "user"),
		zap.String("operation", "login"),
		zap.String("username", username),
	)

	var user *model.User
	var err error

	// 尝试通过用户名查找
	user, err = s.repo.FindByUsername(username)
	if err != nil {
		log.Error("db_query_failed",
			zap.String("event", "db_query_failed"),
			zap.String("field", "username"),
			zap.String("error_kind", "db"),
			zap.Error(err),
		)
		return "", err
	}

	// 如果用户名没找到，尝试通过手机号查找
	if user == nil {
		user, err = s.repo.FindByPhone(username)
		if err != nil {
			log.Error("db_query_failed",
				zap.String("event", "db_query_failed"),
				zap.String("field", "phone"),
				zap.String("error_kind", "db"),
				zap.Error(err),
			)
			return "", err
		}
	}

	// 如果手机号也没找到，尝试通过邮箱查找
	if user == nil {
		user, err = s.repo.FindByEmail(username)
		if err != nil {
			log.Error("db_query_failed",
				zap.String("event", "db_query_failed"),
				zap.String("field", "email"),
				zap.String("error_kind", "db"),
				zap.Error(err),
			)
			return "", err
		}
	}

	// 用户不存在
	if user == nil {
		log.Warn("login_failed_user_not_found",
			zap.String("event", "login_failed_user_not_found"),
		)
		return "", errors.New("invalid username or password")
	}

	// 检查用户状态
	if user.Status != 1 {
		log.Warn("login_failed_user_inactive",
			zap.String("event", "login_failed_user_inactive"),
			zap.Uint64("user_id", user.ID),
		)
		return "", errors.New("user is inactive")
	}

	// 验证密码
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		log.Warn("login_failed_invalid_password",
			zap.String("event", "login_failed_invalid_password"),
			zap.Uint64("user_id", user.ID),
		)
		return "", errors.New("invalid username or password")
	}

	// 生成token
	token, err := middleware.GenerateToken(fmt.Sprintf("%d", user.ID))
	if err != nil {
		log.Error("token_generate_failed",
			zap.String("event", "token_generate_failed"),
			zap.Uint64("user_id", user.ID),
			zap.String("error_kind", "internal"),
			zap.Error(err),
		)
		return "", err
	}

	log.Info("user_login_success",
		zap.String("event", "user_login_success"),
		zap.String("entity_type", "user"),
		zap.Uint64("entity_id", user.ID),
	)

	return token, nil
}

func (s *userService) WechatLogin(ctx context.Context, code string) (string, error) {
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "user"),
		zap.String("operation", "wechat_login"),
	)

	// 调用微信 jscode2session 换取真实 openid
	cfg := config.Get().Wechat
	wxClient := wechat.NewClient(cfg.AppID, cfg.AppSecret)
	session, err := wxClient.Code2Session(code)
	if err != nil {
		log.Error("wechat_code2session_failed",
			zap.String("event", "wechat_code2session_failed"),
			zap.String("error_kind", "external"),
			zap.Error(err),
		)
		return "", fmt.Errorf("微信登录失败，请稍后重试")
	}

	log = log.With(zap.String("openid_hash", wechat.HashOpenID(session.OpenID)))

	user, err := s.repo.FindByWechatOpenID(session.OpenID)
	if err != nil {
		log.Error("db_query_failed",
			zap.String("event", "db_query_failed"),
			zap.String("field", "wechat_openid"),
			zap.String("error_kind", "db"),
			zap.Error(err),
		)
		return "", err
	}

	if user == nil {
		user = &model.User{
			Username:     "wxuser_" + session.OpenID[:8],
			Nickname:     "微信用户",
			WechatOpenID: strPtr(session.OpenID),
			Status:       1,
		}
		if err = s.repo.Create(user); err != nil {
			log.Error("user_create_failed",
				zap.String("event", "user_create_failed"),
				zap.String("error_kind", "db"),
				zap.Error(err),
			)
			return "", err
		}
		log.Info("wechat_user_created",
			zap.String("event", "wechat_user_created"),
			zap.String("entity_type", "user"),
			zap.Uint64("entity_id", user.ID),
		)
	}

	token, err := middleware.GenerateToken(fmt.Sprintf("%d", user.ID))
	if err != nil {
		log.Error("token_generate_failed",
			zap.String("event", "token_generate_failed"),
			zap.Uint64("user_id", user.ID),
			zap.String("error_kind", "internal"),
			zap.Error(err),
		)
		return "", err
	}

	log.Info("wechat_login_success",
		zap.String("event", "wechat_login_success"),
		zap.String("entity_type", "user"),
		zap.Uint64("entity_id", user.ID),
	)

	return token, nil
}

func (s *userService) GetUserInfo(ctx context.Context, userID uint64) (*model.User, error) {
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "user"),
		zap.String("operation", "get_user_info"),
		zap.Uint64("user_id", userID),
	)

	user, err := s.repo.FindByID(userID)
	if err != nil {
		log.Error("db_query_failed",
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", "db"),
			zap.Error(err),
		)
		return nil, err
	}

	return user, nil
}

func (s *userService) UpdateUserInfo(ctx context.Context, userID uint64, nickname, avatar string) error {
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "user"),
		zap.String("operation", "update_user_info"),
		zap.Uint64("user_id", userID),
	)

	user, err := s.repo.FindByID(userID)
	if err != nil {
		log.Error("db_query_failed",
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", "db"),
			zap.Error(err),
		)
		return err
	}

	if user == nil {
		log.Warn("user_not_found",
			zap.String("event", "user_not_found"),
		)
		return errors.New("user not found")
	}

	if nickname != "" {
		user.Nickname = nickname
	}
	if avatar != "" {
		user.Avatar = avatar
	}

	if err := s.repo.Update(user); err != nil {
		log.Error("user_update_failed",
			zap.String("event", "user_update_failed"),
			zap.String("error_kind", "db"),
			zap.Error(err),
		)
		return err
	}

	log.Info("user_info_updated",
		zap.String("event", "user_info_updated"),
		zap.String("entity_type", "user"),
		zap.Uint64("entity_id", userID),
	)

	return nil
}

func (s *userService) ChangePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) error {
	log := platformlogger.FromContext(ctx).With(
		zap.String("module", "user"),
		zap.String("operation", "change_password"),
		zap.Uint64("user_id", userID),
	)

	user, err := s.repo.FindByID(userID)
	if err != nil {
		log.Error("db_query_failed",
			zap.String("event", "db_query_failed"),
			zap.String("error_kind", "db"),
			zap.Error(err),
		)
		return err
	}

	if user == nil {
		log.Warn("user_not_found",
			zap.String("event", "user_not_found"),
		)
		return errors.New("user not found")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		log.Warn("invalid_old_password",
			zap.String("event", "invalid_old_password"),
		)
		return errors.New("invalid old password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Error("password_hash_failed",
			zap.String("event", "password_hash_failed"),
			zap.String("error_kind", "internal"),
			zap.Error(err),
		)
		return err
	}

	user.Password = string(hashedPassword)
	if err := s.repo.Update(user); err != nil {
		log.Error("user_update_failed",
			zap.String("event", "user_update_failed"),
			zap.String("error_kind", "db"),
			zap.Error(err),
		)
		return err
	}

	log.Info("password_changed",
		zap.String("event", "password_changed"),
		zap.String("entity_type", "user"),
		zap.Uint64("entity_id", userID),
	)

	return nil
}

func strPtr(s string) *string {
	if s == "" { return nil }
	return &s
}
