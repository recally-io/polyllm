# IR 模型库

该目录实现了 PolyLLM 使用的中间表示（IR）层，核心目标是提供供应商无关的请求、响应与流式事件模型，以便在不同 LLM API 之间做适配。

## 功能概览

- `request.go`：定义指令、消息、块、采样参数、约束、结构化输出等数据结构。
- `events.go`：描述流式事件模型，支持块级增量。
- `builder.go`：提供构造请求的便捷函数。
- `validate.go`：对请求与事件进行校验，包含可配置的范围检查。
- `normalize.go`：规范化请求，提供去重、合并文本等能力。
- `aggregate.go`：将流式事件聚合为非流式响应，可选择直写文本。
- `meta.go`、`errors.go`：通用元数据及错误定义。

库本身不依赖外部包，可单独在其他项目中复用。

## 示例

```go
req := ir.NewRequest("example-model")
req.AppendDirective(ir.NewDirective(ir.DirectiveSystem, ir.NewTextBlock("d1", "You are helpful")))
req.Messages = []ir.Message{
    ir.NewMessage(ir.RoleUser, "m1", ir.NewTextBlock("b1", "Hello")),
}
if err := ir.ValidateRequest(req, ir.WithTemperatureRange(0, 2)); err != nil {
    panic(err)
}
```

## 许可

与仓库其他部分一致，遵循 MIT 许可证。
