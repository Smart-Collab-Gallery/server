package v1

import (
	"github.com/go-kratos/kratos/v2/errors"
)

// Error 辅助函数

func ErrorParamsError(format string, args ...interface{}) *errors.Error {
	return errors.New(400, ErrorReason_PARAMS_ERROR.String(), format)
}

func ErrorAccountTooShort(format string, args ...interface{}) *errors.Error {
	return errors.New(400, ErrorReason_ACCOUNT_TOO_SHORT.String(), format)
}

func ErrorPasswordTooShort(format string, args ...interface{}) *errors.Error {
	return errors.New(400, ErrorReason_PASSWORD_TOO_SHORT.String(), format)
}

func ErrorPasswordNotMatch(format string, args ...interface{}) *errors.Error {
	return errors.New(400, ErrorReason_PASSWORD_NOT_MATCH.String(), format)
}

func ErrorAccountDuplicate(format string, args ...interface{}) *errors.Error {
	return errors.New(409, ErrorReason_ACCOUNT_DUPLICATE.String(), format)
}

func ErrorSystemError(format string, args ...interface{}) *errors.Error {
	return errors.New(500, ErrorReason_SYSTEM_ERROR.String(), format)
}

func IsParamsError(err error) bool {
	return errors.Reason(err) == ErrorReason_PARAMS_ERROR.String()
}

func IsAccountTooShort(err error) bool {
	return errors.Reason(err) == ErrorReason_ACCOUNT_TOO_SHORT.String()
}

func IsPasswordTooShort(err error) bool {
	return errors.Reason(err) == ErrorReason_PASSWORD_TOO_SHORT.String()
}

func IsPasswordNotMatch(err error) bool {
	return errors.Reason(err) == ErrorReason_PASSWORD_NOT_MATCH.String()
}

func IsAccountDuplicate(err error) bool {
	return errors.Reason(err) == ErrorReason_ACCOUNT_DUPLICATE.String()
}

func IsSystemError(err error) bool {
	return errors.Reason(err) == ErrorReason_SYSTEM_ERROR.String()
}

func ErrorUserNotExistOrPasswordError(format string, args ...interface{}) *errors.Error {
	return errors.New(401, ErrorReason_USER_NOT_EXIST_OR_PASSWORD_ERROR.String(), format)
}

func IsUserNotExistOrPasswordError(err error) bool {
	return errors.Reason(err) == ErrorReason_USER_NOT_EXIST_OR_PASSWORD_ERROR.String()
}

func ErrorAccountError(format string, args ...interface{}) *errors.Error {
	return errors.New(400, ErrorReason_ACCOUNT_ERROR.String(), format)
}

func IsAccountError(err error) bool {
	return errors.Reason(err) == ErrorReason_ACCOUNT_ERROR.String()
}

func ErrorPasswordError(format string, args ...interface{}) *errors.Error {
	return errors.New(400, ErrorReason_PASSWORD_ERROR.String(), format)
}

func IsPasswordError(err error) bool {
	return errors.Reason(err) == ErrorReason_PASSWORD_ERROR.String()
}

func ErrorNotLoginError(format string, args ...interface{}) *errors.Error {
	return errors.New(401, ErrorReason_NOT_LOGIN_ERROR.String(), format)
}

func IsNotLoginError(err error) bool {
	return errors.Reason(err) == ErrorReason_NOT_LOGIN_ERROR.String()
}

func ErrorInvalidToken(format string, args ...interface{}) *errors.Error {
	return errors.New(401, ErrorReason_INVALID_TOKEN.String(), format)
}

func IsInvalidToken(err error) bool {
	return errors.Reason(err) == ErrorReason_INVALID_TOKEN.String()
}

func ErrorUserNotFound(format string, args ...interface{}) *errors.Error {
	return errors.New(404, ErrorReason_USER_NOT_FOUND.String(), format)
}

func IsUserNotFound(err error) bool {
	return errors.Reason(err) == ErrorReason_USER_NOT_FOUND.String()
}

func ErrorNoAuthError(format string, args ...interface{}) *errors.Error {
	return errors.New(403, ErrorReason_NO_AUTH_ERROR.String(), format)
}

func IsNoAuthError(err error) bool {
	return errors.Reason(err) == ErrorReason_NO_AUTH_ERROR.String()
}
