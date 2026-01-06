package v1

import (
	"github.com/go-kratos/kratos/v2/errors"
)

// Error 辅助函数

func ErrorPictureNotFound(format string, args ...interface{}) *errors.Error {
	return errors.New(404, ErrorReason_PICTURE_NOT_FOUND.String(), format)
}

func ErrorPictureUploadFailed(format string, args ...interface{}) *errors.Error {
	return errors.New(500, ErrorReason_PICTURE_UPLOAD_FAILED.String(), format)
}

func ErrorPictureDeleteFailed(format string, args ...interface{}) *errors.Error {
	return errors.New(500, ErrorReason_PICTURE_DELETE_FAILED.String(), format)
}

func ErrorPictureUpdateFailed(format string, args ...interface{}) *errors.Error {
	return errors.New(500, ErrorReason_PICTURE_UPDATE_FAILED.String(), format)
}

func ErrorPictureNoAuth(format string, args ...interface{}) *errors.Error {
	return errors.New(403, ErrorReason_PICTURE_NO_AUTH.String(), format)
}

func ErrorPictureFileTooLarge(format string, args ...interface{}) *errors.Error {
	return errors.New(400, ErrorReason_PICTURE_FILE_TOO_LARGE.String(), format)
}

func ErrorPictureFormatError(format string, args ...interface{}) *errors.Error {
	return errors.New(400, ErrorReason_PICTURE_FORMAT_ERROR.String(), format)
}

func ErrorParamsError(format string, args ...interface{}) *errors.Error {
	return errors.New(400, ErrorReason_PARAMS_ERROR.String(), format)
}

func ErrorInvalidArgument(format string, args ...interface{}) *errors.Error {
	return errors.New(400, ErrorReason_INVALID_ARGUMENT.String(), format)
}

func ErrorUnauthorized(format string, args ...interface{}) *errors.Error {
	return errors.New(401, ErrorReason_UNAUTHORIZED.String(), format)
}

// Is 辅助函数

func IsPictureNotFound(err error) bool {
	return errors.Reason(err) == ErrorReason_PICTURE_NOT_FOUND.String()
}

func IsPictureUploadFailed(err error) bool {
	return errors.Reason(err) == ErrorReason_PICTURE_UPLOAD_FAILED.String()
}

func IsPictureDeleteFailed(err error) bool {
	return errors.Reason(err) == ErrorReason_PICTURE_DELETE_FAILED.String()
}

func IsPictureUpdateFailed(err error) bool {
	return errors.Reason(err) == ErrorReason_PICTURE_UPDATE_FAILED.String()
}

func IsPictureNoAuth(err error) bool {
	return errors.Reason(err) == ErrorReason_PICTURE_NO_AUTH.String()
}

func IsPictureFileTooLarge(err error) bool {
	return errors.Reason(err) == ErrorReason_PICTURE_FILE_TOO_LARGE.String()
}

func IsPictureFormatError(err error) bool {
	return errors.Reason(err) == ErrorReason_PICTURE_FORMAT_ERROR.String()
}

func IsParamsError(err error) bool {
	return errors.Reason(err) == ErrorReason_PARAMS_ERROR.String()
}

func IsInvalidArgument(err error) bool {
	return errors.Reason(err) == ErrorReason_INVALID_ARGUMENT.String()
}

func IsUnauthorized(err error) bool {
	return errors.Reason(err) == ErrorReason_UNAUTHORIZED.String()
}
