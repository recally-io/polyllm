package ir

import (
	"encoding/json"
	"time"
)

// DirectiveKind 描述指令类型。
type DirectiveKind string

const (
	DirectiveSystem    DirectiveKind = "system"
	DirectiveDeveloper DirectiveKind = "developer"
)

// Directive 承载 system/developer 等指令信息。
type Directive struct {
	Kind   DirectiveKind `json:"kind"`
	Blocks []Block       `json:"blocks"`
	Meta   Meta          `json:"meta,omitempty"`
}

// Role 表示对话消息的角色。
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message 是对话轮次的承载单元。
type Message struct {
	ID     string  `json:"id"`
	Role   Role    `json:"role"`
	Blocks []Block `json:"blocks"`
	Meta   Meta    `json:"meta,omitempty"`
}

// BlockType 表示内容块类型。
type BlockType string

const (
	BlockText       BlockType = "text"
	BlockJSON       BlockType = "json"
	BlockImage      BlockType = "image"
	BlockAudio      BlockType = "audio"
	BlockFileRef    BlockType = "file_ref"
	BlockToolCall   BlockType = "tool_call"
	BlockToolResult BlockType = "tool_result"
)

// Block 是多模态内容的基本单位。
type Block struct {
	ID   string    `json:"id"`
	Type BlockType `json:"type"`

	Text       *TextBlock       `json:"text,omitempty"`
	JSON       *JSONBlock       `json:"json,omitempty"`
	Image      *ImageBlock      `json:"image,omitempty"`
	Audio      *AudioBlock      `json:"audio,omitempty"`
	FileRef    *FileRefBlock    `json:"file_ref,omitempty"`
	ToolCall   *ToolCallBlock   `json:"tool_call,omitempty"`
	ToolResult *ToolResultBlock `json:"tool_result,omitempty"`

	Meta Meta `json:"meta,omitempty"`
}

// TextBlock 为文本内容。
type TextBlock struct {
	Data string `json:"data"`
	Mime string `json:"mime,omitempty"`
}

// JSONBlock 承载 JSON 结构。
type JSONBlock struct {
	Data any `json:"data"`
}

// ImageBlock 描述图像内容。
type ImageBlock struct {
	Source ImageSource `json:"source"`
	Alt    string      `json:"alt,omitempty"`
	Mime   string      `json:"mime,omitempty"`
}

// ImageSource 提供图像数据或 URL。
type ImageSource struct {
	URL  string `json:"url,omitempty"`
	Data []byte `json:"-"`
}

// AudioBlock 描述音频内容。
type AudioBlock struct {
	URL  string `json:"url,omitempty"`
	Data []byte `json:"-"`
	Mime string `json:"mime,omitempty"`
}

// FileRefBlock 引用外部文件。
type FileRefBlock struct {
	URI  string `json:"uri"`
	Mime string `json:"mime,omitempty"`
	Name string `json:"name,omitempty"`
}

// ToolCallBlock 描述工具调用。
type ToolCallBlock struct {
	CallID    string           `json:"call_id"`
	Name      string           `json:"name"`
	Arguments *json.RawMessage `json:"arguments"`
}

// ToolResultBlock 描述工具返回结果。
type ToolResultBlock struct {
	CallID  string           `json:"call_id"`
	Success bool             `json:"success"`
	Output  *json.RawMessage `json:"output,omitempty"`
	Text    string           `json:"text,omitempty"`
	IsError bool             `json:"is_error,omitempty"`
}

// Sampler 包含采样相关的参数。
type Sampler struct {
	Temperature       *float32        `json:"temperature,omitempty"`
	TopP              *float32        `json:"top_p,omitempty"`
	TopK              *int            `json:"top_k,omitempty"`
	FrequencyPenalty  *float32        `json:"frequency_penalty,omitempty"`
	PresencePenalty   *float32        `json:"presence_penalty,omitempty"`
	RepetitionPenalty *float32        `json:"repetition_penalty,omitempty"`
	LogitBias         map[int]float32 `json:"logit_bias,omitempty"`
	Seed              *int64          `json:"seed,omitempty"`
}

// Constraints 描述输出限制。
type Constraints struct {
	MaxOutputTokens *int     `json:"max_output_tokens,omitempty"`
	StopSequences   []string `json:"stop_sequences,omitempty"`
}

// OutputSchema 描述结构化输出。
type OutputSchema struct {
	Kind       string           `json:"kind"`
	JSONSchema *json.RawMessage `json:"json_schema,omitempty"`
	Strict     bool             `json:"strict,omitempty"`
}

// Request 描述一次完整的模型调用。
type Request struct {
	Model      string      `json:"model"`
	Directives []Directive `json:"directives,omitempty"`
	Messages   []Message   `json:"messages,omitempty"`

	Sampler     Sampler      `json:"sampler,omitempty"`
	Constraints Constraints  `json:"constraints,omitempty"`
	Output      OutputSchema `json:"output,omitempty"`

	Meta Meta `json:"meta,omitempty"`
}

// Usage 记录 token 用量。
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
	BilledTokens     int `json:"billed_tokens,omitempty"`
}

// FinishReason 描述结束原因。
type FinishReason string

const (
	FinishStop        FinishReason = "stop"
	FinishLength      FinishReason = "length"
	FinishStopSeq     FinishReason = "stop_sequence"
	FinishToolUse     FinishReason = "tool_use"
	FinishContentFilt FinishReason = "content_filter"
	FinishOther       FinishReason = "other"
)

// Output 描述最终生成的消息集合。
type Output struct {
	Messages   []Message `json:"messages"`
	TextConcat string    `json:"text_concat,omitempty"`
}

// Response 是非流式聚合结果。
type Response struct {
	ID      string       `json:"id"`
	Model   string       `json:"model"`
	Created time.Time    `json:"created"`
	Usage   Usage        `json:"usage"`
	Finish  FinishReason `json:"finish"`
	Output  Output       `json:"output"`
	Meta    Meta         `json:"meta,omitempty"`
}
