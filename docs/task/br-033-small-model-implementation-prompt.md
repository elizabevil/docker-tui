# BR-033 小模型完整实现指令

你是 dtui 项目的实现模型。请在当前工作区完成 BR-033 剩余功能，不要只做分析或输出方案。

## 开始前必须读取

- `docs/ai-prompts.md`
- `docs/task/br-033-task019-advanced-container-actions.md`
- `docs/task/br-033-advanced-container-ops-analysis.md`
- `docs/task/br-033-035-review-fixes.md`
- `internal/data/runtime/containers.go`
- `internal/tui/keyboard/container_advanced.go`
- `internal/tui/keyboard/actions.go`
- `internal/tui/actionbar/registry.go`
- `internal/tui/state/dialog.go`
- `internal/tui/ui/widget/dialog/`
- `internal/tui/update/update.go`

源码是最终事实来源。当前工作区存在其他模型和用户的未提交修改，必须保留，不得 reset、checkout、clean 或覆盖无关变更。

## 当前事实

- Diff、Wait 已有 Action Bar 入口、Cmd、结果处理和测试，不要删除或重新设计。
- Copy、Update、Export、Commit 目前只有部分 Cmd / Msg / 流保存核心，没有可用 UI。
- 6 个高级动作统一通过 Action Bar 发现，不分配直接快捷键。
- Action Bar 只承载复杂或多步操作，不加入全局命令。
- Dialog 默认焦点必须是 Cancel；需要确认或 Force 时提供独立选项，不使用不断扩张的枚举确认框。

## 必须完成

### 1. Copy

- 容器 Action Bar 增加 `Copy`。
- 打开参数表单，至少输入容器内源路径和主机目标 tar 路径。
- 提交后调用现有 `containerCopyCmd`。
- 成功显示写入路径与字节数；失败写入统一错误提示和 audit。
- reader 必须关闭；失败不得破坏已有目标文件。

### 2. Update

- 容器 Action Bar 增加 `Update`。
- 表单支持 Memory、NanoCPUs、RestartPolicy、RestartMaxRetries。
- 空字段表示不修改；数值和 restart policy 必须校验，非法输入不得调用 runtime。
- 提交后调用现有 `containerUpdateCmd`。
- 成功、失败均结束 audit，并提供 toast / 统一错误提示。

### 3. Export

- 容器 Action Bar 增加 `Export`。
- 表单输入主机目标 tar 路径。
- 提交后调用现有 `containerExportCmd`。
- 成功显示路径与字节数；失败不得留下部分 tar 或破坏已有文件。

### 4. Commit

- 容器 Action Bar 增加 `Commit`。
- 表单支持 repository、tag、author、comment、pause。
- repository 必填；tag 可选；pause 使用 checkbox / 二态字段，默认采用 runtime 当前约定。
- 提交后调用现有 `containerCommitCmd`。
- 成功显示新 image ID，并触发镜像资源刷新；失败进入统一错误提示和 audit。

### 5. 公共集成

- 在 `keyboard/actions.go` 接入四个 Action case。
- 在 `update` 层处理四种 Done Msg，不能静默吞错误。
- Action Bar 中最终必须包含 Rename、Top、Port、Diff、Wait、Copy、Update、Export、Commit。
- 参数表单尽量复用现有 dialog/input/state 组件；确有共同逻辑时可提取通用多字段表单，但不要大范围重构。
- 增加 en / zh / ja i18n；用户可见文案不得只硬编码英文。
- 同步 Help / Footer 时遵守：Action Bar 动作不作为直接快捷键重复显示。
- 同步 BR-033 任务文档，只有所有验收项完成后才能标记 done。

## 允许修改范围

- `internal/tui/actionbar/`
- `internal/tui/keyboard/`
- `internal/tui/state/`
- `internal/tui/update/`
- `internal/tui/ui/widget/dialog/`
- `internal/tui/ui/app/`（仅 dialog 渲染接线）
- `internal/tui/ui/action/`（仅必要提示同步）
- `internal/data/i18n/lang/`
- `docs/task/br-033-task019-advanced-container-actions.md`
- `docs/navigation.md`（仅实际交互同步）

不要修改 runtime 接口或 Docker / Podman adapter，除非当前源码证明已有接口无法完成任务；遇到这种情况先停止并报告，不要自行扩大边界。

## 测试要求

至少覆盖：

- 四个 Action Bar 条目及 disabled 状态。
- 四种表单默认焦点、取消、字段编辑、校验和提交。
- 无 engine / 无选中容器时不执行。
- 四种 Done Msg 的成功、失败与 audit 结束状态。
- Copy / Export reader 关闭、原子替换及失败清理。
- Commit 成功触发镜像刷新。
- Docker / Podman 通过同一 runtime interface 调用，不依赖具体 adapter 类型。

完成后运行：

```bash
GOCACHE=/tmp/dtui-go-cache go test ./internal/tui/...
GOCACHE=/tmp/dtui-go-cache go test -short ./...
GOCACHE=/tmp/dtui-go-cache go vet ./...
git diff --check
```

## 禁止事项

- 不要 git commit，不要 push。
- 不要删除、恢复或格式化任务范围之外的文件。
- 不要重新加入冲突快捷键。
- 不要把未接入 UI 的 Cmd 声称为功能完成。
- 不要只写完成报告而不实现代码。

## 最终输出

1. 修改文件清单。
2. Copy / Update / Export / Commit 各自的完整交互链路。
3. 测试命令与结果。
4. 仍未完成或需要主模型判断的问题。
5. 明确声明没有 commit、没有 push。
