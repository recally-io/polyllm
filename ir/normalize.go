package ir

import "strings"

// NormalizeOption 控制规范化行为。
type NormalizeOption struct {
	TrimWhitespace             bool
	MergeConsecutiveTextBlocks bool
	DedupeStopSequences        bool
}

// NormalizeRequest 根据选项返回规范化后的请求副本。
func NormalizeRequest(req *Request, opt *NormalizeOption) *Request {
	if req == nil {
		return nil
	}
	o := NormalizeOption{DedupeStopSequences: true}
	if opt != nil {
		o = *opt
	} else {
		o.DedupeStopSequences = true
	}
	out := *req
	out.Directives = cloneDirectives(req.Directives, o.TrimWhitespace, o.MergeConsecutiveTextBlocks)
	out.Messages = cloneMessages(req.Messages, o.TrimWhitespace, o.MergeConsecutiveTextBlocks)
	if req.Sampler.LogitBias != nil {
		out.Sampler.LogitBias = make(map[int]float32, len(req.Sampler.LogitBias))
		for k, v := range req.Sampler.LogitBias {
			out.Sampler.LogitBias[k] = v
		}
	}
	if o.DedupeStopSequences {
		out.Constraints.StopSequences = dedupeStrings(req.Constraints.StopSequences)
	} else {
		out.Constraints.StopSequences = append([]string(nil), req.Constraints.StopSequences...)
	}
	if req.Output.JSONSchema != nil {
		out.Output.JSONSchema = CloneArguments(req.Output.JSONSchema)
	}
	out.Meta = req.Meta.Clone()
	return &out
}

func cloneDirectives(ds []Directive, trim, merge bool) []Directive {
	if len(ds) == 0 {
		return nil
	}
	out := make([]Directive, len(ds))
	for i, d := range ds {
		out[i] = Directive{
			Kind:   d.Kind,
			Blocks: normalizeBlocks(d.Blocks, trim, merge),
			Meta:   d.Meta.Clone(),
		}
	}
	return out
}

func cloneMessages(ms []Message, trim, merge bool) []Message {
	if len(ms) == 0 {
		return nil
	}
	out := make([]Message, len(ms))
	for i, m := range ms {
		out[i] = Message{
			ID:     m.ID,
			Role:   m.Role,
			Blocks: normalizeBlocks(m.Blocks, trim, merge),
			Meta:   m.Meta.Clone(),
		}
	}
	return out
}

func normalizeBlocks(blocks []Block, trim, merge bool) []Block {
	if len(blocks) == 0 {
		return nil
	}
	out := make([]Block, 0, len(blocks))
	var pending *Block
	flush := func() {
		if pending != nil {
			out = append(out, *pending)
			pending = nil
		}
	}
	for _, b := range blocks {
		cp := copyBlockHeader(&b)
		if cp.Text != nil {
			cp.Text.Data = b.Text.Data
			if trim {
				cp.Text.Data = strings.TrimSpace(cp.Text.Data)
			}
			if merge {
				if pending != nil && pending.Type == BlockText {
					pending.Text.Data += cp.Text.Data
					continue
				}
				if pending != nil {
					flush()
				}
				pending = cp
				continue
			}
		}
		flush()
		pending = cp
		flush()
	}
	flush()
	return out
}

func dedupeStrings(ss []string) []string {
	if len(ss) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(ss))
	out := make([]string, 0, len(ss))
	for _, s := range ss {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
