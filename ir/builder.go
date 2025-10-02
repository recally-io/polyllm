package ir

import "encoding/json"

// NewRequest 创建一个指定模型的 Request。
func NewRequest(model string) *Request {
	return &Request{Model: model}
}

// NewDirective 构造 Directive。
func NewDirective(kind DirectiveKind, blocks ...Block) Directive {
	return Directive{Kind: kind, Blocks: append([]Block(nil), blocks...)}
}

// NewTextBlock 构造文本块。
func NewTextBlock(id, text string) Block {
	return Block{
		ID:   id,
		Type: BlockText,
		Text: &TextBlock{Data: text},
	}
}

// NewJSONBlock 构造 JSON 块。
func NewJSONBlock(id string, data any) Block {
	return Block{
		ID:   id,
		Type: BlockJSON,
		JSON: &JSONBlock{Data: data},
	}
}

// NewMessage 构造 Message。
func NewMessage(role Role, id string, blocks ...Block) Message {
	return Message{
		ID:     id,
		Role:   role,
		Blocks: append([]Block(nil), blocks...),
	}
}

// AppendDirective 追加指令。
func (r *Request) AppendDirective(d Directive) {
	if r == nil {
		return
	}
	r.Directives = append(r.Directives, d)
}

// AppendUserText 追加用户文本消息。
func (r *Request) AppendUserText(id, text string) {
	if r == nil {
		return
	}
	r.Messages = append(r.Messages, NewMessage(RoleUser, id, NewTextBlock(id+"_b", text)))
}

// AppendAssistantText 追加助手文本消息。
func (r *Request) AppendAssistantText(id, text string) {
	if r == nil {
		return
	}
	r.Messages = append(r.Messages, NewMessage(RoleAssistant, id, NewTextBlock(id+"_b", text)))
}

// CloneArguments 返回工具参数的深拷贝。
func CloneArguments(raw *json.RawMessage) *json.RawMessage {
	if raw == nil {
		return nil
	}
	dup := make(json.RawMessage, len(*raw))
	copy(dup, *raw)
	return &dup
}
