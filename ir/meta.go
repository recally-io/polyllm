package ir

// Meta 允许调用方附带任意元数据。建议仅使用可 JSON 序列化的值。
type Meta map[string]any

// Clone 返回 Meta 的浅拷贝。若 m 为空，返回 nil。
func (m Meta) Clone() Meta {
	if len(m) == 0 {
		return nil
	}
	out := make(Meta, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
