package ir

import "errors"

var (
	// ErrInvalidArgument 表示请求参数非法。
	ErrInvalidArgument = errors.New("ir: invalid argument")
	// ErrUnsupported 表示某功能尚未支持。
	ErrUnsupported = errors.New("ir: unsupported")
	// ErrEventOrder 表示流式事件顺序不合法。
	ErrEventOrder = errors.New("ir: invalid event order")
	// ErrStateConflict 表示状态机冲突，例如重复结束。
	ErrStateConflict = errors.New("ir: state conflict")
)

// FieldError 描述字段校验失败的原因。
type FieldError struct {
	Field  string
	Reason string
}

func (e *FieldError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Field == "" {
		return e.Reason
	}
	return e.Field + ": " + e.Reason
}
