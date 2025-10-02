package ir

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type aggregatorOptions struct {
	writer     io.Writer
	maxBuffer  int
	onDelta    func(Ref, string)
	keepBlocks bool
}

// AggregatorOption 配置聚合器。
type AggregatorOption func(*aggregatorOptions)

// WithWriter 指定增量写入目标。
func WithWriter(w io.Writer) AggregatorOption {
	return func(o *aggregatorOptions) {
		o.writer = w
	}
}

// WithMaxBufferBytes 限制内存缓冲大小。
func WithMaxBufferBytes(n int) AggregatorOption {
	return func(o *aggregatorOptions) {
		if n < 0 {
			n = 0
		}
		o.maxBuffer = n
	}
}

// WithOnDelta 注册文本增量回调。
func WithOnDelta(cb func(ref Ref, s string)) AggregatorOption {
	return func(o *aggregatorOptions) {
		o.onDelta = cb
	}
}

// WithKeepBlocks 控制是否保留块信息。
func WithKeepBlocks(keep bool) AggregatorOption {
	return func(o *aggregatorOptions) {
		o.keepBlocks = keep
	}
}

// Aggregator 将事件序列聚合为 Response。
type Aggregator struct {
	opts aggregatorOptions

	started        bool
	finished       bool
	streamFinished bool
	errored        bool

	currentMessage *Message
	currentBlock   *Block
	blockStack     []string

	messages []Message

	textBuffer     bytes.Buffer
	bufferLimitHit bool

	usage        Usage
	finishReason *FinishReason
}

// NewAggregator 构造聚合器。
func NewAggregator(opts ...AggregatorOption) *Aggregator {
	cfg := aggregatorOptions{keepBlocks: true}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Aggregator{opts: cfg}
}

// Feed 消化单个事件。
func (a *Aggregator) Feed(ev Event) error {
	if a.errored {
		return ErrStateConflict
	}
	switch ev.Type {
	case EventStreamStart:
		if a.started {
			return ErrStateConflict
		}
		a.started = true
	case EventStreamError:
		if !a.started || a.streamFinished || a.finished {
			return ErrStateConflict
		}
		a.errored = true
		return nil
	case EventStreamFinish:
		if !a.started || a.streamFinished || !a.finished {
			return ErrStateConflict
		}
		a.streamFinished = true
	case EventMessageStart:
		if !a.started || a.finished || a.currentMessage != nil {
			return ErrStateConflict
		}
		if ev.MsgHdr == nil {
			return ErrInvalidArgument
		}
		msg := *ev.MsgHdr
		msg.Meta = msg.Meta.Clone()
		msg.Blocks = nil
		a.currentMessage = &msg
	case EventMessageEnd:
		if a.currentMessage == nil || len(a.blockStack) != 0 {
			return ErrEventOrder
		}
		a.messages = append(a.messages, *a.currentMessage)
		a.currentMessage = nil
	case EventBlockStart:
		if a.currentMessage == nil {
			return ErrEventOrder
		}
		if ev.Block == nil {
			return ErrInvalidArgument
		}
		cp := copyBlockHeader(ev.Block)
		a.currentBlock = cp
		a.blockStack = append(a.blockStack, cp.ID)
	case EventBlockEnd:
		if len(a.blockStack) == 0 {
			return ErrEventOrder
		}
		last := a.blockStack[len(a.blockStack)-1]
		if ev.Ref.BlockID != "" && ev.Ref.BlockID != last {
			return ErrEventOrder
		}
		a.blockStack = a.blockStack[:len(a.blockStack)-1]
		if a.currentMessage == nil || a.currentBlock == nil || a.currentBlock.ID != last {
			return ErrStateConflict
		}
		if a.opts.keepBlocks {
			a.currentMessage.Blocks = append(a.currentMessage.Blocks, *a.currentBlock)
		}
		a.currentBlock = nil
	case EventTextDelta:
		if a.currentBlock == nil || a.currentBlock.Text == nil {
			return ErrEventOrder
		}
		if cb := a.opts.onDelta; cb != nil {
			cb(ev.Ref, ev.Text)
		}
		if a.opts.writer != nil && ev.Text != "" {
			if _, err := io.WriteString(a.opts.writer, ev.Text); err != nil {
				return fmt.Errorf("write delta: %w", err)
			}
		}
		if a.opts.keepBlocks {
			a.currentBlock.Text.Data += ev.Text
		}
		if !a.bufferLimitHit {
			if a.opts.maxBuffer > 0 && a.textBuffer.Len()+len(ev.Text) > a.opts.maxBuffer {
				remain := a.opts.maxBuffer - a.textBuffer.Len()
				if remain > 0 {
					a.textBuffer.WriteString(ev.Text[:remain])
				}
				a.bufferLimitHit = true
			} else {
				a.textBuffer.WriteString(ev.Text)
			}
		}
	case EventJSONDelta:
		return ErrUnsupported
	case EventUsageDelta:
		if ev.Usage != nil {
			a.usage.PromptTokens += ev.Usage.PromptTokens
			a.usage.CompletionTokens += ev.Usage.CompletionTokens
			a.usage.TotalTokens += ev.Usage.TotalTokens
			a.usage.BilledTokens += ev.Usage.BilledTokens
		}
	case EventFinish:
		if a.finished {
			return ErrStateConflict
		}
		if ev.Finish == nil {
			return ErrInvalidArgument
		}
		if a.currentMessage != nil || len(a.blockStack) != 0 {
			return ErrEventOrder
		}
		fr := *ev.Finish
		a.finishReason = &fr
		a.finished = true
	default:
		return ErrUnsupported
	}
	return nil
}

// Done 完成聚合并返回 Response。
func (a *Aggregator) Done(base Response) (Response, error) {
	if !a.started {
		return Response{}, ErrStateConflict
	}
	if a.errored {
		return Response{}, ErrStateConflict
	}
	if !a.streamFinished {
		return Response{}, errors.Join(ErrEventOrder, fmt.Errorf("stream_finish missing"))
	}
	if !a.finished {
		return Response{}, errors.Join(ErrEventOrder, fmt.Errorf("finish missing"))
	}
	base.Output.Messages = append([]Message(nil), a.messages...)
	base.Output.TextConcat = a.textBuffer.String()
	base.Usage = a.usage
	if a.finishReason != nil {
		base.Finish = *a.finishReason
	}
	return base, nil
}

// CurrentText 返回已聚合的文本。
func (a *Aggregator) CurrentText() string {
	return a.textBuffer.String()
}

// CurrentUsage 返回聚合的用量。
func (a *Aggregator) CurrentUsage() Usage {
	return a.usage
}

// CurrentFinish 返回已记录的结束原因。
func (a *Aggregator) CurrentFinish() *FinishReason {
	return a.finishReason
}

func copyBlockHeader(b *Block) *Block {
	cp := *b
	cp.Meta = cp.Meta.Clone()
	if b.Text != nil {
		tb := *b.Text
		tb.Data = ""
		cp.Text = &tb
	}
	if b.JSON != nil {
		jb := *b.JSON
		cp.JSON = &jb
	}
	if b.Image != nil {
		ib := *b.Image
		cp.Image = &ib
	}
	if b.Audio != nil {
		ab := *b.Audio
		cp.Audio = &ab
	}
	if b.FileRef != nil {
		fb := *b.FileRef
		cp.FileRef = &fb
	}
	if b.ToolCall != nil {
		tc := *b.ToolCall
		tc.Arguments = CloneArguments(b.ToolCall.Arguments)
		cp.ToolCall = &tc
	}
	if b.ToolResult != nil {
		tr := *b.ToolResult
		if b.ToolResult.Output != nil {
			tr.Output = CloneArguments(b.ToolResult.Output)
		}
		cp.ToolResult = &tr
	}
	return &cp
}
