package ir

import "time"

// Ref 精确定位到消息或块。
type Ref struct {
	MessageID string `json:"message_id,omitempty"`
	BlockID   string `json:"block_id,omitempty"`
	Index     *int   `json:"index,omitempty"`
}

// EventType 描述流式事件类型。
type EventType string

const (
	EventStreamStart  EventType = "stream_start"
	EventStreamError  EventType = "stream_error"
	EventStreamFinish EventType = "stream_finish"

	EventMessageStart EventType = "message_start"
	EventBlockStart   EventType = "block_start"
	EventTextDelta    EventType = "text_delta"
	EventJSONDelta    EventType = "json_delta"
	EventBlockEnd     EventType = "block_end"
	EventMessageEnd   EventType = "message_end"
	EventUsageDelta   EventType = "usage_delta"
	EventFinish       EventType = "finish"
)

// Event 表示一次流式事件。
type Event struct {
	Type   EventType     `json:"type"`
	Ref    Ref           `json:"ref,omitempty"`
	Ts     time.Time     `json:"ts"`
	Text   string        `json:"text,omitempty"`
	JSON   any           `json:"json,omitempty"`
	Usage  *Usage        `json:"usage,omitempty"`
	Finish *FinishReason `json:"finish,omitempty"`
	Block  *Block        `json:"block,omitempty"`
	MsgHdr *Message      `json:"msg_hdr,omitempty"`
	Err    *StreamError  `json:"err,omitempty"`
	Meta   Meta          `json:"meta,omitempty"`
}

// StreamError 描述流式错误。
type StreamError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
