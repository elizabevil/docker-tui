# 小模型下一步执行简报

本文给小模型读取使用。当前 `docs/task/` 已经完成任务卡梳理,下一步目标是从任务卡进入可执行实现闭环。

## 当前判断

优先不继续扩展需求列表,先从已有任务卡中选择一个低风险任务进入实现准备。

推荐优先级:

1. `BR-040 Dialog 风格统一:四周透明 + panel 居中`
   - 推荐先做。
   - 原因: 范围相对独立,和确认窗口、后续 Action Bar 都有关。
   - 适合小模型参与: 只读定位、测试补齐、机械迁移。
   - 主模型保留: 最终 CenterOnPanel 几何方案和合并判断。

2. `BR-039 取消 Q 退出,改为每页多功能 Action Bar`
   - 推荐第二个做。
   - 原因: TASK-019、Image Import、Registry Login 都需要它作为低冲突入口。
   - 适合小模型参与: 动作清单整理、i18n/docs/help/footer 同步、diff 巡检。
   - 主模型保留: 全局键盘语义、Quit 行为、Action Bar 状态流。

3. `BR-034 镜像 History 顶层页`
   - 推荐 BR-039 后做。
   - 原因: H 键存在冲突,Action Bar 落地后快捷键压力更小。
   - 适合小模型参与: 状态/渲染只读定位、focused tests、文档同步。
   - 主模型保留: History 是独立页还是详情子模式的决策。

暂缓:

- `BR-033 TASK-019 容器高级动作`: 依赖 Action Bar 降低快捷键冲突。
- `BR-036 镜像 tarball 导入`: 需要 runtime 接口和 Action Bar 入口。
- `BR-037 Registry Login`: 涉及凭证流和 Podman 端点确认。
- `BR-038 Exec Shell UI`: PTY、ANSI、detach、resize 风险高。
- `BR-035 Events + Network Connect`: 实际是两个功能,建议后续拆开。
- `TASK-020 Podman 能力评估`: 这是决策任务,不是当前编码任务。

## 小模型当前任务

请先执行 `BR-040 Dialog 风格统一` 的只读代码定位,不要修改文件。

读取任务卡:

- `docs/task/br-040-dialog-style-center-on-panel.md`

重点读取源码:

- `internal/tui/ui/widget/dialog/`
- `internal/tui/ui/app/layout.go`
- `internal/tui/ui/app/rails.go`
- `internal/tui/ui/app/compact.go`
- `internal/tui/ui/component/dialog.go`
- `internal/tui/state/confirm.go`
- `internal/tui/keyboard/confirm.go`

如果存在测试,一并读取:

- `internal/tui/ui/widget/dialog/*_test.go`
- `internal/tui/ui/app/*_test.go`
- `internal/tui/ui/component/*dialog*_test.go`

## 输出要求

只输出分析报告,不要修改代码。

报告必须包含:

1. 当前 dialog / overlay 渲染链路
   - ModeConfirm
   - selection dialog
   - exec dialog
   - image workflow / transfer dialog
   - filter / search / command 是否属于 dialog

2. 当前居中依据
   - 是按 terminal 居中
   - 是按 panel 居中
   - 是否存在 compact layout 特殊路径

3. 建议的 `CenterOnPanel` 最小实现边界
   - 必改文件
   - 可选文件
   - 不应触碰文件

4. 测试建议
   - 几何计算单测
   - overlay 保留底层内容测试
   - 窄宽度 fallback 测试
   - compact layout 测试

5. 风险点
   - dialog 覆盖 header / footer
   - 双栏 Compose panel
   - terminal 太小时的 fallback
   - ANSI 宽度计算
   - 与 BR-039 Action Bar 的复用关系

6. 需要主模型确认的问题
   - panel rect 来源
   - 垂直居中范围
   - dialog 超出 panel 时如何处理
   - 多 dialog 是否需要支持

## 明确禁止

- 不要修改代码。
- 不要新增文件。
- 不要提交 git commit。
- 不要把 BR-039 Action Bar 一起实现。
- 不要重新设计全部布局系统。
- 不要引用旧路径,例如 `internal/config`、`internal/tui/model`、`internal/data/docker`。

## 建议使用 Prompt

```text
你是 dtui 项目的代码定位助手。

只读分析,不要修改文件。

目标需求:
BR-040 Dialog 风格统一:四周透明 + panel 居中。

请读取:
- docs/task/br-040-dialog-style-center-on-panel.md
- internal/tui/ui/widget/dialog/
- internal/tui/ui/app/layout.go
- internal/tui/ui/app/rails.go
- internal/tui/ui/app/compact.go
- internal/tui/ui/component/dialog.go
- internal/tui/state/confirm.go
- internal/tui/keyboard/confirm.go

请输出:
1. 当前 dialog / overlay 渲染链路。
2. 当前居中依据。
3. CenterOnPanel 最小实现边界。
4. 建议测试。
5. 风险点。
6. 需要主模型确认的问题。

要求:
- 只输出分析报告。
- 不要修改文件。
- 不要实现 BR-039。
- 不要引用不存在的旧路径。
```
