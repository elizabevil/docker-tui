# AI 协作提示指南

本文保留面向当前代码结构的最小上下文，避免继续引用旧路径。

## 项目上下文速查

```text
项目: dtui
语言: Go
CLI: cobra
TUI: Bubble Tea v2
样式: Lip Gloss v2
容器引擎: Docker SDK + Podman 兼容接入

入口:
  cmd/docker-tui/main.go

当前包结构:
  internal/data/config
  internal/data/docker
  internal/data/i18n
  internal/tui/state
  internal/tui/keyboard
  internal/tui/keys
  internal/tui/ui/app
  internal/tui/ui/pages
  internal/tui/ui/component
  internal/tui/update*.go
```

## 协作约束

```text
- 先读代码，再判断实现位置。
- 所有异步容器操作优先走 tea.Cmd。
- 状态定义集中在 internal/tui/state。
- 键盘入口集中在 internal/tui/keyboard。
- 主布局从 internal/tui/ui/app/layout.go 进入。
- 设计规范优先参考 design/current-design.md。
```

## 需求分析 Prompt

```text
你在 dtui 项目中做需求分析。

请基于当前代码而不是历史文档，输出：
1. 现状
2. 目标行为
3. 需要改动的包
4. 风险和回归点

功能: [需求描述]
建议先读:
- cmd/docker-tui/main.go
- internal/tui/state/app.go
- internal/tui/keyboard/
- internal/tui/ui/pages/
```

## 设计 Prompt

```text
你在 dtui 项目中设计一个新功能。

约束：
- 使用 Bubble Tea MVU
- 数据接入放在 internal/data/docker
- 状态放在 internal/tui/state
- 交互放在 internal/tui/keyboard
- 渲染放在 internal/tui/ui/pages 或 ui/component

输出：
1. 变更文件列表
2. 新增消息流
3. 状态字段变更
4. 键盘交互方案
5. UI 影响面
```

## 实现 Prompt

```text
在 dtui 中实现 [功能]。

要求：
- 先说明要改哪些文件
- 再逐文件给出修改
- 不要引用不存在的旧目录，如 internal/config 或 internal/tui/model
- 如果涉及 UI，说明是在 pages、component 还是 widget 层修改
- 如果涉及异步操作，明确 tea.Cmd -> Msg -> Update 的链路
```

## Bug 修复 Prompt

```text
定位 dtui 中的一个问题。

请输出：
1. 根因
2. 最小修复方案
3. 需要补的测试
4. 是否影响 keyboard / state / update / ui 的其他路径

问题描述:
[粘贴现象]
```

## Review Prompt

```text
请以 dtui 的代码审查视角审查以下改动，重点关注：
- Bubble Tea 消息流是否正确
- 是否错误修改了 state 与 view 的职责边界
- 容器引擎调用是否仍集中在 internal/data/docker
- 键盘分支是否与现有模式冲突
- 是否引入了文档与代码再次漂移
```
