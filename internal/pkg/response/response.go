package response

import (
	"encoding/json"
	"net/http"

	"github.com/go-kratos/kratos/v2/errors"
)

// Response 统一响应结构
type Response struct {
	Code     int                    `json:"code"`
	Message  string                 `json:"message"`
	Metadata map[string]interface{} `json:"metadata"`
	Reason   string                 `json:"reason"`
	Data     interface{}            `json:"data"`
}

// Success 创建成功响应
func Success(data interface{}) *Response {
	return &Response{
		Code:     200,
		Message:  "",
		Metadata: make(map[string]interface{}),
		Reason:   "",
		Data:     data,
	}
}

// SuccessWithMessage 创建带消息的成功响应
func SuccessWithMessage(data interface{}, message string) *Response {
	return &Response{
		Code:     200,
		Message:  message,
		Metadata: make(map[string]interface{}),
		Reason:   "",
		Data:     data,
	}
}

// Error 创建错误响应
func Error(code int, message, reason string, metadata map[string]interface{}) *Response {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	return &Response{
		Code:     code,
		Message:  message,
		Metadata: metadata,
		Reason:   reason,
		Data:     nil,
	}
}

// ResponseEncoder 自定义响应编码器，统一响应格式
// 所有成功的响应都会经过这个函数处理
func ResponseEncoder(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if v == nil {
		// 如果响应为空，也返回统一格式
		v = map[string]interface{}{}
	}

	// 检查是否已经是统一响应格式
	if resp, ok := v.(*Response); ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		return json.NewEncoder(w).Encode(resp)
	}

	// 将普通响应（如 protobuf message）包装成统一格式
	resp := Success(v)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(resp)
}

// ErrorEncoder 自定义错误编码器，统一错误响应格式
// 所有错误响应都会经过这个函数处理
func ErrorEncoder(w http.ResponseWriter, r *http.Request, err error) {
	se := errors.FromError(err)

	// 将 map[string]string 转换为 map[string]interface{}
	metadata := make(map[string]interface{})
	for k, v := range se.Metadata {
		metadata[k] = v
	}

	resp := &Response{
		Code:     int(se.Code),
		Message:  se.Message,
		Metadata: metadata,
		Reason:   se.Reason,
		Data:     nil,
	}

	w.Header().Set("Content-Type", "application/json")
	// HTTP 状态码统一返回 200，业务错误码在 response.code 中体现
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
