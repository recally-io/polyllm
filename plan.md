# PolyLLM API 格式扩展实现计划

## 目标
- 将 PolyLLM 的内部数据交换模型迁移到统一的 IR（中间表示）库，以便统一适配多家厂商的对话式/流式 LLM API。
- 在保持现有功能（CLI、HTTP Server、Go SDK）的前提下，扩展对更多 API 格式（OpenAI Responses、Anthropic Messages 等）的兼容能力。
- 为未来新增厂商提供可复用的适配框架与测试基线。

## 总体策略
1. **引入独立 IR 模块**：按照设计文档在仓库内新增 `ir/` 目录，实现零依赖的请求/响应/事件模型、校验与聚合能力。
2. **实现 PolyLLM ↔ IR 的桥接层**：在现有 `llm` 抽象中增加从 PolyLLM 自身模型到 IR 的转换工具，并提供反向转换。
3. **分阶段迁移与适配**：优先实现 OpenAI Chat/Responses、Anthropic Messages 等主流 API 的 IR 适配器，逐步替换/兼容现有实现，确保每一步可回滚。
4. **完善测试与文档**：构建单元测试、集成测试及端到端示例，更新 README/示例配置，指导使用者扩展。

## 里程碑
| 里程碑 | 内容 | 交付物 |
| --- | --- | --- |
| M1 | IR 库落地 | `ir/` 数据结构、校验、聚合器及 README；基础单元测试 |
| M2 | PolyLLM 内核桥接层 | `internal/irbridge`（示例命名）模块，封装 IR ↔ PolyLLM 的转换逻辑；对 Request/Response 结构的兼容性测试 |
| M3 | 主流 API 适配 | OpenAI Chat + Responses、Anthropic Messages 适配器，新增 provider 测试用例、stream 聚合验证 |
| M4 | CLI/HTTP Server 改造 | CLI & Server 在默认路径上使用 IR；保持旧配置兼容；提供迁移指南 |
| M5 | 扩展与文档 | 针对其他现有厂商（Groq、Gemini 等）完成适配；补充性能/压力测试脚本与用户文档 |

## 详细任务拆分

### M1：IR 库实现
- **数据结构落地**：
  - 按设计实现 Directives、Message、Block、Sampler、Constraints、OutputSchema 等结构体以及不变式。
  - 为 Block 子类型提供构造函数，并保证 JSON 序列化标签与规范一致。
- **流式事件模型**：
  - 定义 `Event`、`Ref`、`StreamError` 及事件类型常量。
  - 编写事件序列的状态机校验器 `ValidateEvents`，覆盖顺序约束与错误处理。
- **构建器/规范化/校验**：
  - 实现 `NewRequest`、`NormalizeRequest`、`ValidateRequest` 等纯函数。
  - 提供校验选项与错误类型（`FieldError`、`ErrEventOrder` 等）。
- **聚合器**：
  - 完成事件到非流式响应的 `Aggregator`；实现 Writer/MaxBuffer/OnDelta 等选项。
- **测试与文档**：
  - 针对每个模块编写单元测试（Go test）；覆盖成功路径与异常路径。
  - 在 `ir/README.md` 和 `doc.go` 中撰写使用说明与示例。

### M2：PolyLLM ↔ IR 桥接层
- **内部模型盘点**：
  - 梳理 `llm.go`、`llm_models.go` 等结构，明确与 IR 对应字段（模型名、指令、消息、工具调用等）。
- **转换工具**：
  - 在 `internal/irbridge` 新建包，提供：
    - `FromPolyRequest`：将 PolyLLM 的请求结构映射为 `ir.Request`。
    - `ToPolyResponse`：将 `ir.Response` 转换为 PolyLLM 当前的响应结构。
    - 辅助函数：处理 StopSequences、工具调用、Meta 等。
- **流式桥接**：
  - 设计 `EventSink` 接口，将 IR `Event` 推送到 PolyLLM 现有的流式回调。
  - 确保聚合器与现有 streaming API 协同工作。
- **兼容层测试**：
  - 针对转换函数编写单元测试，模拟多模态/工具调用等情况。

### M3：主流 API 适配
- **OpenAI Chat/Responses**：
  - 重构现有 OpenAI Provider，使其基于 IR 进行请求拼装与响应解析。
  - 支持新的 Responses API（指令、内容块、stream）。
- **Anthropic Messages**：
  - 新增/重构 Anthropic Provider，将 Messages API 映射为 IR。
  - 处理不同事件粒度（block delta、message stop 等）。
- **回归测试**：
  - 使用本地 mock 或录制的 HTTP fixture 验证适配器正确性。
  - 针对 stream 与非流式分别断言 `Aggregator` 输出。

### M4：CLI 与 HTTP Server 改造
- **请求路径统一**：
  - CLI/Server 在构造请求时通过桥接层生成 `ir.Request`。
  - 流式输出通过 IR 事件驱动，实现一致的 `--stream` 行为。
- **配置兼容**：
  - 确保现有 JSON 配置与环境变量解析逻辑无需调整或仅需最少改动。
  - 对新增字段（如 OutputSchema）设计配置项并提供默认值。
- **文档更新**：
  - README 中新增 IR 适配说明、示例代码与流式事件示例。
  - 提供从旧版迁移到 IR 的 FAQ。

### M5：扩展与文档完善
- **多厂商扩展**：
  - 按优先级逐步为 Groq、Gemini、DeepSeek、OpenRouter 等补充 IR 适配。
  - 编写统一的 provider 适配器模板与贡献指南。
- **性能/稳定性**：
  - 构建长文本、工具调用、结构化输出等场景的压力测试脚本。
  - 记录内存占用、流式延迟，验证 `WithMaxBufferBytes` 等选项效果。
- **最终文档**：
  - 在 `docs/`（如需新建）内整理开发者指南、测试建议与示例。
  - 更新 `builtin-llm-providers.json`，标注哪些提供者支持高级特性。

## 风险与缓解
- **迁移风险**：大规模重构可能影响现有用户。→ 采取分支发布、保持旧接口兼容、编写迁移指南。
- **事件模型差异**：不同厂商流式粒度不一致。→ 利用 IR `Ref` 与可选事件，编写适配层单元测试。
- **性能压力**：聚合长文本可能导致内存飙升。→ 默认启用 `WithMaxBufferBytes`，并在文档中推荐使用。
- **测试复杂度**：多厂商、多模式测试庞大。→ 引入 mock 上游/录制回放，尽量使用表驱动测试和 golden files。

## 成功标准
- 新增 IR 模块通过 `go test ./...`。
- 至少两个主流 API（OpenAI Responses、Anthropic Messages）通过 IR 成功交互，包含流式与工具调用场景。
- CLI/HTTP Server 在启用 IR 后通过集成测试，无破坏性变更。
- 文档清晰描述如何扩展新的 API 格式，并提供示例。

