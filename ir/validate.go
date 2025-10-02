package ir

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

type validateCfg struct {
	tempRange *floatRange
	topPRange *floatRange
	stopRune  int
}

type floatRange struct {
	min float32
	max float32
}

// ValidateOption 配置请求校验行为。
type ValidateOption func(*validateCfg)

// WithTemperatureRange 指定温度范围。
func WithTemperatureRange(min, max float32) ValidateOption {
	return func(cfg *validateCfg) {
		cfg.tempRange = &floatRange{min: min, max: max}
	}
}

// WithTopPRange 指定 top_p 范围。
func WithTopPRange(min, max float32) ValidateOption {
	return func(cfg *validateCfg) {
		cfg.topPRange = &floatRange{min: min, max: max}
	}
}

// WithMaxStopRune 限制 stop 序列最大 rune 数。
func WithMaxStopRune(n int) ValidateOption {
	return func(cfg *validateCfg) {
		cfg.stopRune = n
	}
}

// ValidateRequest 校验请求结构。
func ValidateRequest(req *Request, opts ...ValidateOption) error {
	if req == nil {
		return ErrInvalidArgument
	}
	cfg := &validateCfg{}
	for _, opt := range opts {
		opt(cfg)
	}
	if req.Model == "" {
		return &FieldError{Field: "model", Reason: "empty"}
	}
	blockIDs := make(map[string]struct{})
	messageIDs := make(map[string]struct{})

	checkBlocks := func(prefix string, blocks []Block) error {
		for i, b := range blocks {
			if b.ID == "" {
				return &FieldError{Field: fmt.Sprintf("%s.blocks[%d].id", prefix, i), Reason: "empty"}
			}
			if _, ok := blockIDs[b.ID]; ok {
				return &FieldError{Field: fmt.Sprintf("%s.blocks[%d].id", prefix, i), Reason: "duplicate"}
			}
			blockIDs[b.ID] = struct{}{}

			// 确保仅有一个载荷字段非空。
			count := 0
			if b.Text != nil {
				count++
			}
			if b.JSON != nil {
				count++
			}
			if b.Image != nil {
				count++
			}
			if b.Audio != nil {
				count++
			}
			if b.FileRef != nil {
				count++
			}
			if b.ToolCall != nil {
				count++
			}
			if b.ToolResult != nil {
				count++
			}
			if count != 1 {
				return &FieldError{Field: fmt.Sprintf("%s.blocks[%d]", prefix, i), Reason: "invalid payload"}
			}

			switch b.Type {
			case BlockText:
				if b.Text == nil {
					return &FieldError{Field: fmt.Sprintf("%s.blocks[%d].text", prefix, i), Reason: "missing_payload"}
				}
			case BlockJSON:
				if b.JSON == nil {
					return &FieldError{Field: fmt.Sprintf("%s.blocks[%d].json", prefix, i), Reason: "missing_payload"}
				}
			case BlockImage:
				if b.Image == nil {
					return &FieldError{Field: fmt.Sprintf("%s.blocks[%d].image", prefix, i), Reason: "missing_payload"}
				}
			case BlockAudio:
				if b.Audio == nil {
					return &FieldError{Field: fmt.Sprintf("%s.blocks[%d].audio", prefix, i), Reason: "missing_payload"}
				}
			case BlockFileRef:
				if b.FileRef == nil {
					return &FieldError{Field: fmt.Sprintf("%s.blocks[%d].file_ref", prefix, i), Reason: "missing_payload"}
				}
			case BlockToolCall:
				if b.ToolCall == nil {
					return &FieldError{Field: fmt.Sprintf("%s.blocks[%d].tool_call", prefix, i), Reason: "missing_payload"}
				}
			case BlockToolResult:
				if b.ToolResult == nil {
					return &FieldError{Field: fmt.Sprintf("%s.blocks[%d].tool_result", prefix, i), Reason: "missing_payload"}
				}
			default:
				return &FieldError{Field: fmt.Sprintf("%s.blocks[%d].type", prefix, i), Reason: "unsupported"}
			}

			if b.ToolCall != nil {
				if b.ToolCall.CallID == "" {
					return &FieldError{Field: fmt.Sprintf("%s.blocks[%d].tool_call.call_id", prefix, i), Reason: "empty"}
				}
				if b.ToolCall.Name == "" {
					return &FieldError{Field: fmt.Sprintf("%s.blocks[%d].tool_call.name", prefix, i), Reason: "empty"}
				}
			}
			if b.ToolResult != nil {
				if b.ToolResult.CallID == "" {
					return &FieldError{Field: fmt.Sprintf("%s.blocks[%d].tool_result.call_id", prefix, i), Reason: "empty"}
				}
			}
			if b.FileRef != nil && b.FileRef.URI == "" {
				return &FieldError{Field: fmt.Sprintf("%s.blocks[%d].file_ref.uri", prefix, i), Reason: "empty"}
			}
		}
		return nil
	}
	for i, d := range req.Directives {
		if d.Kind == "" {
			return &FieldError{Field: fmt.Sprintf("directives[%d].kind", i), Reason: "empty"}
		}
		if err := checkBlocks(fmt.Sprintf("directives[%d]", i), d.Blocks); err != nil {
			return err
		}
	}

	for i, msg := range req.Messages {
		if msg.ID == "" {
			return &FieldError{Field: fmt.Sprintf("messages[%d].id", i), Reason: "empty"}
		}
		if _, ok := messageIDs[msg.ID]; ok {
			return &FieldError{Field: fmt.Sprintf("messages[%d].id", i), Reason: "duplicate"}
		}
		messageIDs[msg.ID] = struct{}{}
		if msg.Role == "" {
			return &FieldError{Field: fmt.Sprintf("messages[%d].role", i), Reason: "empty"}
		}
		if err := checkBlocks(fmt.Sprintf("messages[%d]", i), msg.Blocks); err != nil {
			return err
		}
	}

	if cfg.tempRange != nil && req.Sampler.Temperature != nil {
		t := *req.Sampler.Temperature
		if t < cfg.tempRange.min || t > cfg.tempRange.max {
			return &FieldError{Field: "sampler.temperature", Reason: "out_of_range"}
		}
	}
	if cfg.topPRange != nil && req.Sampler.TopP != nil {
		t := *req.Sampler.TopP
		if t < cfg.topPRange.min || t > cfg.topPRange.max {
			return &FieldError{Field: "sampler.top_p", Reason: "out_of_range"}
		}
	}
	if req.Constraints.MaxOutputTokens != nil && *req.Constraints.MaxOutputTokens <= 0 {
		return &FieldError{Field: "constraints.max_output_tokens", Reason: "must_be_positive"}
	}
	seenStop := make(map[string]struct{})
	for i, stop := range req.Constraints.StopSequences {
		if stop == "" {
			return &FieldError{Field: fmt.Sprintf("constraints.stop_sequences[%d]", i), Reason: "empty"}
		}
		if cfg.stopRune > 0 && utf8.RuneCountInString(stop) > cfg.stopRune {
			return &FieldError{Field: fmt.Sprintf("constraints.stop_sequences[%d]", i), Reason: "too_long"}
		}
		if _, ok := seenStop[stop]; ok {
			return &FieldError{Field: fmt.Sprintf("constraints.stop_sequences[%d]", i), Reason: "duplicate"}
		}
		seenStop[stop] = struct{}{}
	}
	switch req.Output.Kind {
	case "", "none", "text":
		// ok
	case "json_schema":
		if req.Output.JSONSchema == nil {
			return &FieldError{Field: "output.json_schema", Reason: "required"}
		}
	default:
		return &FieldError{Field: "output.kind", Reason: "unsupported"}
	}
	return nil
}

// ValidateEvents 校验流式事件顺序。
func ValidateEvents(seq []Event) error {
	var started, finished, streamFinished, errored bool
	var messageOpen bool
	blockStack := make([]string, 0, 4)
	blockTypes := make(map[string]BlockType)

	for idx, ev := range seq {
		switch ev.Type {
		case EventStreamStart:
			if started || finished || errored {
				return fmt.Errorf("event[%d]: unexpected stream_start", idx)
			}
			started = true
		case EventStreamError:
			if !started || finished || streamFinished {
				return fmt.Errorf("event[%d]: unexpected stream_error", idx)
			}
			errored = true
		case EventStreamFinish:
			if !started || streamFinished || errored {
				return fmt.Errorf("event[%d]: unexpected stream_finish", idx)
			}
			if !finished {
				return fmt.Errorf("event[%d]: finish before stream_finish missing", idx)
			}
			streamFinished = true
		case EventMessageStart:
			if !started || finished || errored {
				return fmt.Errorf("event[%d]: unexpected message_start", idx)
			}
			if messageOpen {
				return fmt.Errorf("event[%d]: message already open", idx)
			}
			if ev.MsgHdr == nil || ev.MsgHdr.ID == "" {
				return fmt.Errorf("event[%d]: message header missing id", idx)
			}
			messageOpen = true
		case EventMessageEnd:
			if !messageOpen || errored {
				return fmt.Errorf("event[%d]: unexpected message_end", idx)
			}
			if len(blockStack) != 0 {
				return fmt.Errorf("event[%d]: block still open", idx)
			}
			messageOpen = false
		case EventBlockStart:
			if !messageOpen || errored {
				return fmt.Errorf("event[%d]: unexpected block_start", idx)
			}
			if ev.Block == nil || ev.Block.ID == "" {
				return fmt.Errorf("event[%d]: block header missing id", idx)
			}
			blockStack = append(blockStack, ev.Block.ID)
			blockTypes[ev.Block.ID] = ev.Block.Type
		case EventBlockEnd:
			if len(blockStack) == 0 || errored {
				return fmt.Errorf("event[%d]: unexpected block_end", idx)
			}
			last := blockStack[len(blockStack)-1]
			if ev.Ref.BlockID != "" && ev.Ref.BlockID != last {
				return fmt.Errorf("event[%d]: block_end ref mismatch", idx)
			}
			blockStack = blockStack[:len(blockStack)-1]
		case EventTextDelta:
			if len(blockStack) == 0 || errored {
				return fmt.Errorf("event[%d]: text_delta without block", idx)
			}
			last := blockStack[len(blockStack)-1]
			if bt := blockTypes[last]; bt != BlockText && bt != "" {
				return fmt.Errorf("event[%d]: text_delta in non-text block", idx)
			}
		case EventJSONDelta:
			if len(blockStack) == 0 || errored {
				return fmt.Errorf("event[%d]: json_delta without block", idx)
			}
			last := blockStack[len(blockStack)-1]
			if bt := blockTypes[last]; bt != BlockJSON && bt != "" {
				return fmt.Errorf("event[%d]: json_delta in non-json block", idx)
			}
		case EventUsageDelta:
			if !started || finished || errored {
				return fmt.Errorf("event[%d]: unexpected usage_delta", idx)
			}
		case EventFinish:
			if !started || finished || errored {
				return fmt.Errorf("event[%d]: unexpected finish", idx)
			}
			if messageOpen || len(blockStack) != 0 {
				return fmt.Errorf("event[%d]: finish while scope open", idx)
			}
			finished = true
		default:
			return fmt.Errorf("event[%d]: unknown type %q", idx, ev.Type)
		}
		if errored {
			if idx != len(seq)-1 {
				return fmt.Errorf("event[%d]: stream_error must be terminal", idx)
			}
		}
	}
	if started && !errored {
		if !finished {
			return errors.Join(ErrEventOrder, fmt.Errorf("finish missing"))
		}
		if !streamFinished {
			return errors.Join(ErrEventOrder, fmt.Errorf("stream_finish missing"))
		}
	}
	return nil
}
