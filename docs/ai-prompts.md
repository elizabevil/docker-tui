# AI 协作提示指南

用于团队在 AI 辅助开发时保持一致性。

## 通用开发流程

```
1. 分析 → 2. 设计 → 3. 实现 → 4. 验证
```

每步对应不同的 prompt 模板。

---

## 1. 需求分析 Prompt

```
你是一个 TUI 工具的产品经理。分析以下功能需求，输出：
1. 用户故事
2. 验收标准
3. 边界条件
4. 与现有行为的交互点

功能: [描述要做的功能]
涉及模块: [涉及的 internal/ 包]
```

**使用场景**：新功能或改动前，确保理解透彻再动手。

---

## 2. 架构/设计 Prompt

```
你是一个 Go TUI 应用的架构师。项目使用：
- TUI 框架: charm.land/bubbletea/v2
- 样式: charm.land/lipgloss/v2
- Docker SDK: github.com/docker/docker/client
- 项目结构: internal/{config,docker,tui/{model,view}}/

我需要设计 [功能描述]。

请输出：
1. 新增/修改的文件列表
2. 关键数据流（消息类型 → Update → View）
3. 与现有 Model 的交互方式
4. 错误处理策略
5. 性能考虑

参考现有模式：
- 容器列表: model.ContainerListModel + FetchContainers + renderContainerList
- 异步操作: tea.Cmd 封装 Docker API 调用 → XxxMsg → Update 分支
```

**使用场景**：复杂功能或跨多文件改动前。

---

## 3. 实现 Prompt（单文件改动）

```
你是一个 Go TUI 开发者。请在 [文件路径] 中实现/修改：
[具体需求]

约束：
- 遵循 Bubbletea v2 MVU 模式（Model → Update → View）
- Docker API 调用通过 internal/docker/ 封装
- 不要用 as any 或类型断言跳过类型安全
- 新增消息类型定义在 internal/tui/model/containers.go
- 新增渲染函数放在 internal/tui/view/layout.go
- 异常情况返回 error，不 panic

参考已有代码：
[相关代码片段或文件路径]
```

**使用场景**：具体文件改动，输出即可直接应用的代码。

---

## 4. 实现 Prompt（多文件 + 新功能）

```
你是一个 Go TUI 开发者。在 docker-tui 项目中实现 [功能]。

架构参照：
- Bubbletea v2 的 tea.Cmd 执行异步 Docker API 调用
- 结果通过自定义 Msg 类型返回，在 Update() 中处理
- View() 根据 Model 状态渲染

需要改动的文件：
1. internal/docker/client.go — 添加 [方法]
2. internal/tui/model/containers.go — 添加消息类型
3. internal/tui/update.go — 添加 FetchXxx Cmd + Update 分支
4. internal/tui/view/layout.go — 添加渲染

请按上述顺序逐一实现，每完成一个文件给出 diff。
```

**使用场景**：跨多文件的完整功能，让 AI 按文件粒度输出。

---

## 5. Bug 修复 Prompt

```
在 docker-tui 项目中遇到 Bug：
[现象描述]
[复现步骤]

相关代码上下文：
[文件路径 + 关键行]

已尝试的排查：
[已做检查]

请输出：
1. 根因分析
2. 修复方案
3. 修改的代码

约束：只修 Bug，不做额外重构。
```

**使用场景**：定位和修复问题，防止 AI 顺手做无关改动。

---

## 6. Code Review Prompt

```
请审查以下代码，关注：
1. Bubbletea MVU 模式正确性（Msg → Update → View 单向流）
2. 异步操作是否正确封装为 tea.Cmd
3. 错误处理（Docker API 调用失败时的降级）
4. goroutine 生命周期管理（ctx.Done() 检查）
5. Lip Gloss 样式一致

代码：
[paste code diff or file content]
```

**使用场景**：PR 审查或完成功能后自查。

---

## 7. 测试 Prompt

```
为 docker-tui 项目的 [功能/文件] 编写测试。

测试策略：
- Model 测试：构造消息 → 调用 Update → 断言新状态
- Docker 层：mock client 接口
- View 测试：构造 Model → RenderApp → 断言输出包含关键文本

被测代码：
[paste code]

输出：完整的 test 文件。
```

**使用场景**：需要补测试时。

---

## 8. 项目上下文速查（给 AI 的初始化 prompt）

```
项目: docker-tui
语言: Go
TUI: Bubbletea v2 (charm.land/bubbletea/v2)
样式: Lip Gloss v2 (charm.land/lipgloss/v2)
Docker: github.com/docker/docker/client (官方 SDK)
CLI: github.com/spf13/cobra

架构模式:
  - MVU (Model-View-Update)
  - Docker 异步操作 → tea.Cmd → XxxMsg → Update 处理
  - 事件驱动（Docker Events API 代替轮询）
  - 离线优先（Docker 不可达仍可启动）

包结构:
  cmd/docker-tui/main.go      入口 + Cobra
  internal/config/            配置（YAML）
  internal/docker/            Docker SDK 封装
  internal/tui/model/         Bubbletea Models + Msg 定义
  internal/tui/view/          渲染函数
  internal/tui/update.go      消息路由
  internal/tui/styles.go      主题
  internal/tui/keys.go        键盘绑定

关键约定:
  - 所有异步操作返回 tea.Cmd
  - View 只读，不修改状态
  - Docker SDK 调用通过 internal/docker/ 封装层
  - 新资源类型需添加 Model + Msg + FetchXxx + renderXxxList
```

**使用场景**：在 AI 对话开始时注入，快速建立上下文。
