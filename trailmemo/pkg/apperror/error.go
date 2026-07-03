package apperror

import (
	"errors"
	"net/http"
)

// AppError 是业务层统一错误类型：对外给安全消息，对内保留原始 Cause。
type AppError struct {
	Code       string `json:"code"`
	Kind       string `json:"kind"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Cause      error  `json:"-"`
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	return e.Code + ": " + e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func New(code, kind, message string, status int) *AppError {
	return &AppError{
		Code:       code,
		Kind:       kind,
		Message:    message,
		HTTPStatus: status,
	}
}

func Wrap(err error, code, kind, message string, status int) *AppError {
	return &AppError{
		Code:       code,
		Kind:       kind,
		Message:    message,
		HTTPStatus: status,
		Cause:      err,
	}
}

// From 把任意 error 转成 AppError，兜底为 INTERNAL_ERROR。
func From(err error) *AppError {
	if err == nil {
		return nil
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return Wrap(err, CodeInternalError, KindInternal, "服务内部错误", http.StatusInternalServerError)
}

func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}
