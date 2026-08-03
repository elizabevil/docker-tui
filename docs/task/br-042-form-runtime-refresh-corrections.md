# BR-042 FORM 与容器运行时刷新修正

## 元信息

- 编号: BR-042
- 优先级: high
- 状态: implemented, awaiting live Podman verification
- 依赖: BR-033、BR-041
- 范围: FORM 通用交互、Podman Commit/Wait、Update 默认值、Top/Stats 刷新

## 目标

修正高级容器操作中已经复现的运行时兼容和交互问题，并把 FORM 的焦点、布尔值、文本光标规则收敛到公共状态与公共渲染层。页面不得自行定义同类字段的按键语义或焦点样式。

## 问题与方案

### 1. Podman Commit 返回 404

当前代码请求 `POST /libpod/containers/{id}/commit`。Podman Libpod Commit 实际为集合级端点 `POST /libpod/commit`，容器 ID 通过 `container` query 参数传入。

- `ContainerCommitPath()` 固定返回 `/libpod/commit`。
- `ContainerCommit()` 设置 `container=<id>`、`repo`、`tag`、`comment`、`author`、`pause`。
- Repository 与 Tag 分别传递，不拼接成单个 `repo:tag`。
- 驱动测试断言 method、path 和全部 query。

### 2. FORM 焦点层级不正确

路径候选框或字段获得焦点时，下方 `Esc Cancel` 仍呈选中样式。原因是按钮渲染只判断 `OnConfirm`，没有先判断 `FieldFocus == -1`。

- 字段焦点、Popup 焦点和按钮焦点互斥。
- 仅 `FieldFocus == -1` 时允许 Cancel/Confirm 使用选中样式。
- Popup 打开时由 Popup 独占焦点；底层 FORM 不显示活动焦点。
- Cancel 是 FORM 初始默认按钮，但编辑字段时不持续高亮。

### 3. Update 默认资源值误导并可能被重复提交

`Memory=0`、`NanoCPUs=0` 是“不限制”，不是显式限制 0。当前提交还只检查输入是否为空，没有检查字段是否由用户修改。

- Memory、CPU 仅在 inspect 值大于 0 时显示；0 保持空白，表示 unchanged/unlimited。
- Maximum retries 仅在当前策略为 `on-failure` 时显示。
- Restart policy 可展示当前有效值，但未触碰时不提交。
- Update request 只包含 `Touched == true` 的字段。
- 用户显式输入 `0` 仍是合法修改。

### 4. FORM 文本字段需要闪烁光标

范围为 `FormText`、`FormInt`、`FormPath`；Bool、Select、MultiSelect 不显示文本光标。

- `AppModel` 持有全局 cursor blink phase，应用只启动一个固定周期 tick。
- FORM、筛选/搜索/命令栏、Action Bar 过滤、重命名/创建对话框和 Exec 输入框共享同一相位。
- 输入、移动光标、切换字段时立即显示光标。
- 非输入页面只更新一个布尔相位，不创建页面级 ticker。
- 光标隐藏阶段保留一个终端单元宽度，避免文字左右抖动。

### 5. Podman Wait 被 10 秒 HTTP 超时中断

Wait 是长期阻塞请求，但复用了 `http.Client.Timeout=10s` 的普通 `Post()`。

- 增加由调用方 context 控制生命周期的无 client timeout JSON 请求入口。
- `ContainerWait()` 使用该入口；普通请求继续保留 10 秒保护。
- Esc、切换 runtime 和退出仍通过 `ContainerWaitState.Stop()` 取消 wait。
- 测试证明 wait 超过普通 client timeout 仍可等待，并可由 context 取消。

### 6. Top 与容器 STATISTICS 不自动刷新

Top 目前只在打开或按 `r` 时请求。Stats ticker 离开 Containers 后可能停止调度，返回页面没有可靠重启入口。

- Top 使用 generation 绑定的 2 秒周期刷新；离开 Top 后旧 tick 自动失效。
- Top 请求进行中不并发叠加；刷新保留 cursor 和 viewport。
- Stats ticker 在 engine 存在时持续调度，但只在需要统计的容器列表视图发请求。
- 页面切换不能终止 ticker 链；runtime 断开时停止，重新连接或首次加载时重启。
- 单个容器失败不阻止其他容器和下一轮刷新。

## FORM 公共规则

| 字段类型 | 编辑键 | 激活/确认 | Popup | 光标 |
|---|---|---|---|---|
| Text | Left/Right/Home/End、Backspace/Delete、普通字符 | Enter 下一字段 | 否 | 闪烁文本光标 |
| Int | 同 Text，只接受合法数值字符 | Enter 下一字段 | 否 | 闪烁文本光标 |
| Path | Text 编辑键、Tab 补全 | Enter 确认候选或下一字段 | 候选列表 | 闪烁文本光标 |
| Bool | Space 切换 | Enter 下一字段 | 否 | 无 |
| Select | 无内联文本编辑 | Enter 打开，Enter 确认 | 单选 | 无 |
| MultiSelect | 无内联文本编辑 | Enter 打开，Space 勾选，Enter 确认 | 多选 | 无 |

统一约束:

- Bubble Tea 的命名 `space` 和字面空格统一为 `keys.KeySpace`。
- 字段原语行为只由 `handleFormFieldEditKey` 分发。
- Popup 打开后独占 Up/Down/Left/Right/Enter/Space/Esc 的语义。
- FORM 初始焦点为 Cancel；进入字段后 Cancel 取消高亮。
- Tab 仅在 Update FORM 中切换焦点；Path FORM 中保持 shell 补全语义。

## 代码结构索引

- `internal/driver/podman/containers_advanced.go`: Commit/Wait 请求。
- `internal/driver/podman/rest.go`: 普通与长期请求的 timeout 边界。
- `internal/tui/state/form.go`: FORM 焦点、字段状态和光标 generation。
- `internal/tui/keyboard/container_form.go`: FORM 按键和 Update 提交语义。
- `internal/tui/ui/widget/dialog/form.go`: 字段、按钮和文本光标渲染。
- `internal/tui/state/process.go`: Top 刷新 generation/loading。
- `internal/tui/update/update.go`: Top、FORM cursor tick 消息分发。
- `internal/tui/update/update_tick.go`: Stats/Top/FORM 周期调度。

## 验收标准

- [x] Podman Commit 构造 `/libpod/commit?container=<id>` 请求并由驱动测试覆盖。
- [x] FORM 字段或 Popup 活动时，Cancel/Confirm 均不显示选中态。
- [x] 未设置限制的 Update 表单中 Memory/CPU 不显示 `0`。
- [x] 未修改 Update 字段时不发送对应参数。
- [x] FORM、查询栏、Action Bar 和普通文本对话框共享稳定且不引起布局移动的闪烁光标。
- [x] Podman Wait 不继承普通 HTTP client timeout，并可通过 context 取消。
- [x] Top 页面周期刷新，离开后停止。
- [x] Containers STATISTICS ticker 跨页面保持，返回后恢复请求。
- [ ] 使用真实 Podman 5.4.2 验证 Commit 和超过 10 秒的 Wait。
- [x] `go test -short ./...`、`go vet ./...`、`git diff --check` 通过。

## 非目标

- 不重新设计表格列宽。
- 不增加 Wait 条件选择 FORM。
- 不改变 Docker adapter 的公开接口。
- 不在非相关页面持续请求容器统计数据。
