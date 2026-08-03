# BR-041: 统一 Form 布局、选择组件与路径补全

> 更新日期: 2026-08-02  
> 状态: `partial` - 公共 Form、Local Path、Image Save/Load 已完成；Image Import 待对应需求实现
> 优先级: high  
> 依赖: BR-033、BR-040  
> 执行方式: 小模型分批实现，主模型 review、集成和提交

> FORM 焦点、闪烁光标以及 Podman Commit/Wait、Top/Stats 刷新修正统一记录在 [R01-03 Stats/Top/Wait](../requirement/R01-container/R01-03-stats-top.md) + [constraint/C01-form.md](../constraint/C01-form.md) + [constraint/C05-path.md](../constraint/C05-path.md)(原 BR-042 已归档删除)，后续实现不得在页面内重复定义规则。

## 1. 背景

BR-033 为 Copy / Update / Export / Commit 增加了通用 Form，但当前实现仍有以下问题:

1. Copy 的容器路径和本地目标路径不支持补全。
2. Copy / Export 的目标路径没有根据当前工作目录和容器信息生成默认文件名。
3. `Host destination` 容易被理解为 Docker daemon 所在主机；实际文件写入 dtui 进程所在机器。
4. Form 使用字符串逐行拼接，Label 与 Value 起点不一致，不具备表格式两侧对齐。
5. Select 把全部选项横向放在同一行，内容多时排版混乱。
6. Select / MultiSelect / Bool 没有统一的 Web Form 风格交互。
7. 当前 `Tab` 用于切换字段，与 shell 风格路径补全产生冲突。
8. Image Save / Load / Import 等其他路径输入页面也需要复用相同能力。

本任务目标不是单独修改 Copy，而是建立唯一的公共 Form 状态、渲染和补全来源。

## 2. 设计原则

1. Form 的字段结构、焦点、校验、候选状态只能由公共组件维护。
2. 页面只声明字段配置和提交逻辑，不自行渲染输入框或选择列表。
3. Label 列与 Value 列按外部 Dialog 可用宽度从外到内计算。
4. 路径必须区分 dtui 本地文件系统与容器内部文件系统。
5. 不允许使用 `exec ls/find` 作为容器路径补全的唯一实现。
6. 不支持补全时仍允许手工输入，不阻断原有操作。
7. 危险覆盖操作默认选择 Cancel，并提供独立 Force 开关或确认项。
8. Action Bar 继续只承载复杂操作；本任务不新增直接快捷键。

## 3. 路径语义

### 3.1 Local path

`Local path` 指运行 dtui 进程的机器文件系统。

- Docker / Podman 即使连接远程 daemon，本地目标仍是 dtui 所在机器。
- UI 字段统一使用 `Local destination` / `Local source`，不使用模糊的 `Host path`。
- 相对路径基于 dtui 启动时的当前工作目录。
- 支持 `~`、相对路径和环境变量展开。
- 提交前将路径清理为绝对路径。

### 3.2 Container path

`Container path` 指选中容器内部路径。

- Copy 首批仍允许手工输入任何绝对路径。
- 容器路径候选必须来自统一 `ContainerPathProvider`。
- provider 不可用时显示非阻塞状态，不得禁止提交。
- 不允许直接假设容器正在运行或包含 shell、`ls`、`find`。

## 4. 默认文件名

### 4.1 Copy

当用户填写容器源路径后，若本地目标尚未被用户修改，自动生成:

```text
<cwd>/<container-name>-<source-base>-<YYYYMMDD-HHMMSS>.tar
```

示例:

```text
/home/user/dtui-test-nginx-nginx.conf-20260802-153012.tar
```

规则:

- `container-name` 去除前导 `/`，非法文件名字符替换为 `-`。
- `source-base` 使用源路径 basename；根目录使用 `rootfs`。
- 空容器名回退到 short ID。
- 时间使用 dtui 本地时区。
- 用户手工修改目标后，后续源路径变化不得覆盖用户输入。

### 4.2 Export

打开表单时直接生成:

```text
<cwd>/<container-name>-filesystem-<YYYYMMDD-HHMMSS>.tar
```

### 4.3 其他页面

后续迁移时使用一致规则:

- Image Save: `<image-name>-<tag>-<timestamp>.tar`
- Image Load / Import: 默认从当前工作目录开始选择文件，不自动生成文件名。
- 其他导出类动作: `<resource-name>-<operation>-<timestamp>.<ext>`。

## 5. Form 字段模型

公共字段类型至少包含:

```go
type FormFieldKind int

const (
    FormText FormFieldKind = iota
    FormInt
    FormPath
    FormSelect
    FormMultiSelect
    FormBool
)
```

建议字段配置:

```go
type FormField struct {
    Key         string
    Label       string
    Kind        FormFieldKind
    Input       QueryInputState
    Required    bool
    Disabled    bool
    Error       string

    PathSource  PathSource
    PathMode    PathMode
    Suggestions []PathEntry

    Options     []FormOption
    Selected    map[string]bool
    Toggle      bool
}
```

实际字段名称可按现有代码风格调整，但不得为每个页面重复定义 Select、Bool 或 Path 状态。

### 5.1 PathSource

```go
const (
    PathLocal PathSource = iota
    PathContainer
)
```

### 5.2 PathMode

```go
const (
    PathAny PathMode = iota
    PathFile
    PathDirectory
    PathSaveFile
)
```

## 6. 路径 Provider

建议接口:

```go
type PathProvider interface {
    Complete(ctx context.Context, request PathCompletionRequest) ([]PathEntry, error)
}

type PathCompletionRequest struct {
    Input       string
    ContainerID string
    Mode        PathMode
}

type PathEntry struct {
    Name  string
    Path  string
    IsDir bool
}
```

### 6.1 LocalPathProvider

首批必须实现:

- 展开 `~` 与环境变量。
- 基于输入的目录部分读取本地目录。
- 目录候选末尾显示路径分隔符。
- 目录优先、文件其次，同组按名称排序。
- 隐藏文件仅在 basename 以 `.` 开始时显示。
- `PathSaveFile` 允许不存在的最终文件名，但父目录必须存在。
- 文件系统错误进入字段错误或提示栏，不导致 panic。
- 使用 `filepath`，兼容 Linux / macOS / Windows 路径分隔符。

### 6.2 ContainerPathProvider

第二阶段实现:

- 先核验 Docker 与 Podman 是否有无需 shell 的 archive/stat/list 能力。
- Docker / Podman archive API 不适合列出根目录或大目录，不能作为交互式补全来源。
- 运行中 Linux 容器通过公共 `ExecService` 异步执行 `/bin/sh` glob 脚本，只枚举当前目录的直接子项。
- 脚本通过位置参数接收目录，不拼接用户输入，不使用 `ls/find`。
- 停止容器、无 `/bin/sh`、权限不足或 runtime exec 不可用时显示错误并保留手工输入。
- 请求携带 container ID、field key 和原始输入；返回时丢弃过期结果，避免覆盖后续编辑。

## 7. 键盘交互

### 7.1 普通 Form

| 按键 | 行为 |
|---|---|
| `Up/Down` | 在 Form 字段和按钮区之间移动焦点 |
| `Left/Right` | 输入字段移动光标；Bool/按钮区按控件语义切换 |
| `Enter` | 普通字段进入下一字段；Select/MultiSelect 打开选择框；按钮执行 |
| `Esc` | 先关闭候选/选择框；无子弹层时关闭 Form |
| `Space` | Bool 切换；MultiSelect 选中/取消；Select 选中当前项 |
| `Tab` | 仅用于当前字段的补全；不得切换字段或按钮 |
| `Shift+Tab` | 不承担反向焦点切换；保留给未来反向候选循环 |

### 7.2 Path 字段

| 按键 | 行为 |
|---|---|
| `Tab` | 请求或应用路径补全 |
| `Ctrl+Space` | 显式打开路径候选框，作为 Tab 的备用入口 |
| `Enter` | 有候选框时应用候选；无候选框时进入下一字段 |
| `Esc` | 先关闭候选框，不直接关闭整个 Form |

补全规则:

- 无候选: 保持输入，显示非阻塞提示。
- 单一候选: 直接补全。
- 多个候选: 先补全公共前缀；仍有多个时打开候选框。
- 选中目录: 更新输入并继续停留在 Path 字段。
- 选中文件: 更新输入，可按 Enter 进入下一字段。

## 8. Select / MultiSelect / Bool

### 8.1 Select

- Form 主页面只显示当前值与下拉标识。
- Enter 打开单选弹层。
- 弹层使用单列表格，每项一行，不横向拼接全部选项。
- Up/Down 移动；Space 或 Enter 确认；Esc 取消并恢复原值。

示例:

```text
Restart policy      unchanged  v

┌──────────────────────────┐
│ > unchanged              │
│   no                     │
│   always                 │
│   unless-stopped         │
│   on-failure             │
└──────────────────────────┘
```

### 8.2 MultiSelect

- Enter 打开列表。
- Space 切换当前项。
- 使用 `[ ]` / `[x]` 显示状态。
- Enter 完成选择，Esc 放弃本次弹层修改。
- 主页面显示已选摘要；宽度不足时按可见宽度截断。

### 8.3 Bool

- 不打开弹层。
- 使用 `[ ]` / `[x]`。
- Space 切换。
- Label 与 checkbox 仍遵守两列表格对齐。

## 9. 表格式 Form 布局

目标视觉结构:

```text
Container path      /etc/nginx/nginx.conf
Local destination   /home/user/nginx-nginx.conf-20260802.tar
Overwrite           [ ]
```

布局计算必须从外到内:

1. 取得 Dialog 内部可用宽度。
2. 扣除 border、padding 和固定列间隔。
3. Label 列宽取最长 Label 的显示宽度，并受最大占比约束。
4. Label 统一右对齐。
5. Value 列占满剩余宽度，所有输入起点一致。
6. 输入内容过长时进行水平视口滚动，使光标始终可见。
7. 非聚焦值按显示单元格宽度截断，不得撑宽 Dialog。
8. 字段错误显示在下一行，并与 Value 列左边界对齐。
9. Confirm / Cancel 位于底部独立区域，默认焦点为 Cancel。
10. 不允许 UI 内容与按钮、提示或 Dialog 边框重叠。

不得通过多个空格手工拼出列宽；应新增公共 row layout helper，并使用 ANSI / 宽字符可见宽度函数。

### 9.1 对齐规则

Dialog 内部不能统一使用居中对齐。不同内容必须按语义对齐:

| 内容 | 对齐方式 |
|---|---|
| Dialog 标题 | 左对齐 |
| Form Label | 右对齐 |
| 输入值、路径、数值 | 左对齐 |
| 字段错误 | 与 Value 列左边界对齐 |
| Select / MultiSelect 当前值 | 左对齐 |
| Select / MultiSelect 弹层选项 | 左对齐 |
| 路径候选 | 左对齐 |
| 说明文字 | 左对齐并按可用宽度换行 |
| Confirm / Cancel 按钮组 | 整组位于底部；组内按按钮语义排列 |

居中只允许用于:

- Dialog 整体相对当前 panel 居中。
- 底部按钮组作为一个整体居中。
- 单独的短状态图标。

不得居中路径、选项名称、字段值或多行说明。

### 9.2 Form 示例

`Copy file from container` 应表现为 Form，不是居中的选择菜单:

```text
Copy file from container

       Container path   /etc/nginx/nginx.conf
    Local destination   /home/user/nginx-nginx.conf-20260802.tar
             Overwrite  [ ]

                    Cancel     Confirm
```

- 标题左对齐。
- Label 右对齐。
- 所有 Value 左边界一致。
- 当前字段使用行背景色或 Label / cursor 强调，不移动列位置。
- 焦点变化不得导致行宽、按钮宽度或 Dialog 尺寸变化。

### 9.3 方向键导航

Form 主页面使用方向键进行空间导航:

| 按键 | Form 主页面行为 |
|---|---|
| `Up` | 移动到上一个字段；首字段继续到按钮区 |
| `Down` | 移动到下一个字段；末字段继续到按钮区 |
| `Left/Right` | Text/Int/Path 中移动光标；Bool 切换；按钮区在 Cancel/Confirm 间移动 |
| `Enter` | 普通字段进入下一字段；Select/MultiSelect 打开弹层；按钮执行 |
| `Space` | Bool 切换；其他文本字段输入空格 |
| `Tab` | 对支持补全的字段请求/应用补全；不切换组件 |
| `Shift+Tab` | 不切换组件；可保留为反向遍历补全候选 |

方向键不得全部映射为同一种行为。输入字段的 Left/Right 必须保留文本光标语义。

对于 `Copy file from container`、Export、Image Save / Load / Import 等路径驱动的表单，Tab 语义必须与 shell 一致:

- Tab 只补全当前路径或命令字段。
- 单候选直接补全。
- 多候选首次 Tab 只补全公共前缀；再次 Tab 打开候选列表。
- 重复 Tab 在候选列表中循环或保持列表打开。
- 路径候选使用 Up/Down 在当前目录同级移动，Right 进入选中目录，Left 返回父目录，Enter 确认文件或目录。
- Path Popup 打开后保持固定宽高；切换目录和异步加载只替换固定行内容，不重新挂载不同尺寸窗口。
- 无候选时保持当前输入并显示提示。
- Tab 和 Shift+Tab 都不得把焦点移动到其他输入框或按钮。
- 字段、选择框和按钮之间的移动统一使用 Up/Down/Left/Right。

按钮区规则:

- `Cancel` 在左，`Confirm` 在右。
- 默认焦点是 `Cancel`。
- Left/Right 切换按钮。
- Up 返回最后一个字段。
- Down 可保持在按钮区，不循环触发操作。

### 9.4 Select 弹层布局

Select 弹层是左对齐的单列表格，不是居中的按钮集合:

```text
Restart policy

> unchanged
  no
  always
  unless-stopped
  on-failure
```

- 标题左对齐。
- 选择标记使用固定宽度列。
- 所有选项文字左边界一致。
- 当前行使用背景色和 `>` 标记，不能只靠文字颜色。
- Up/Down 移动，Home/End 到首尾，PgUp/PgDn 分页。
- Space 或 Enter 确认单选。
- Esc 关闭并恢复打开前的值。
- Left/Right 不用于切换单选项，避免与 Form 主页面语义混淆。

### 9.5 MultiSelect 弹层布局

```text
Capabilities

> [x] read
  [ ] write
  [x] inspect
```

- cursor 列、checkbox 列、文本列分别固定宽度。
- checkbox 与选项文字左对齐。
- Up/Down 移动。
- Space 切换当前项。
- Enter 提交全部选择。
- Esc 放弃本次修改。

### 9.6 路径候选布局

```text
Local destination

> [DIR] backups/
  [DIR] exports/
  [FILE] nginx.tar
```

- 候选名称左对齐。
- 目录 / 文件使用固定类型列或统一图标列。
- 不把完整绝对路径全部居中展示。
- 候选较长时优先保留 basename；完整路径可放在底部说明行。
- Up/Down、Home/End、PgUp/PgDn 导航。
- Enter 进入目录或选择文件。
- Esc 只关闭候选框，返回原字段。

## 10. 覆盖确认

`PathSaveFile` 提交时若目标存在:

1. 打开统一确认框。
2. 左侧默认 `Cancel`。
3. 提供独立 `Force` checkbox / toggle。
4. 未启用 Force 时不得覆盖。
5. Copy / Export 保持临时文件写入和原子替换语义。

Form 状态必须保留原目标 resource ID；弹出确认、资源列表刷新或选择行变化后，不得操作其他资源。

## 11. 第一阶段实施范围

小模型第一批只实现以下内容:

- [x] 公共 Form 两列表格布局。
- [x] `FormSelect` 单选弹层。
- [x] `FormMultiSelect` 状态与弹层，即使当前页面暂未使用也需有 focused tests。
- [x] `FormBool` checkbox 交互与对齐。
- [x] `FormPath` 状态和本地路径 provider。
- [x] Tab / Ctrl+Space 路径补全状态机。
- [x] Copy 的 Local destination 自动文件名。
- [x] Export 的 Local destination 自动文件名。
- [x] Copy / Export 目标覆盖确认与 Force。
- [x] en / zh / ja i18n。

### 第一阶段主模型审查修复（2026-08-02）

- 修正 Tab 多候选时错误切换字段，现会应用公共前缀并打开候选弹层。
- 修正 `~` 被展开到 CWD；环境变量、相对路径在提交前统一转为本地绝对路径。
- 保存目标在调用 runtime 前校验父目录。
- MultiSelect 使用工作副本：Space 切换、Enter 提交、Esc 回滚。
- Select 支持 Space 或 Enter 确认，当前多选行正确显示 checkbox。
- 修正 Label 伪右对齐和长路径光标被截断的问题。
- Copy 源路径变化会更新尚未被用户编辑的默认目标文件名。
- 候选弹层限制可见行并使用独立 modal，避免撑破 Form。
- 覆盖确认不能使用 `Y` 绕过默认 Cancel，必须显式选择 Force。
- 增加 Bubble Tea `Ctrl+Space` 实际键名回归测试。

第一阶段原定不做（第二阶段已处理其中部分）:

- 容器内部路径自动补全的 runtime / adapter 实现。
- Image Save / Load / Import 的迁移（Save / Load 已在第二阶段迁移）。
- runtime 接口修改。
- 新直接快捷键。

## 12. 第二阶段实施范围

- [x] 只读分析 Docker / Podman 容器目录枚举能力。
- [x] 主模型确认统一 `ContainerPathProvider` 方案。
- [x] Copy 源路径使用 `FormPath/PathContainer`；运行中 Linux 容器支持异步 Tab / Ctrl+Space 目录补全，失败时保留手工输入。
- [x] 迁移 Image Save / Load 到公共 Form、LocalPathProvider 和统一覆盖确认。
- [ ] Image Import 尚未实现，待 BR-036 对应 workflow 落地后直接使用公共 Form。
- [ ] 同步 navigation / feature design / bugfix requirements。

### 第二阶段能力结论（2026-08-02）

- Docker 与 Podman 的 archive/stat API 可以读取指定路径或导出 tar，但没有稳定、低成本的目录枚举接口。
- 递归读取 archive 来模拟目录列表可能传输大量容器数据，不适合作为交互式 Tab 补全。
- 采用公共 `ExecService` 和 POSIX shell glob 枚举当前目录，避免 `ls/find` 工具差异、命令注入和递归传输。
- 补全命令异步运行，不阻塞 Bubble Tea UI；单候选直接补全，多候选应用公共前缀并打开弹层。
- 该增强仅适用于运行中且包含 `/bin/sh` 的 Linux 容器；其他场景明确降级为手工输入，不阻断 Copy。
- Image Save 自动生成绝对 tar 路径；Image Load 可浏览本地目录，提交时校验为已存在的普通文件。

### Update Form 补充（2026-08-02）

- 打开 Update 后异步 Inspect 当前容器，显示 Memory、CPU、Restart policy 和 Max retries。
- Inspect 返回前已经编辑的字段不会被异步结果覆盖。
- Update 使用 Tab/Shift+Tab 切换字段和按钮；路径类 Form 继续保留 Tab 补全语义。

### Commit Form 补充（2026-08-02）

- Repository、Tag、Author、Comment、Archive destination 均提供默认值。
- Bool 切换统一由公共 `FormField.ToggleBool()` 定义，所有 Form 只响应 Space；Pause 默认 true，Export image as tar 默认 false。
- Form 原始按键先统一归一化：Bubble Tea 的 `space` 与字面空格都转换为 `keys.KeySpace`；字段级规则只由 `handleFormFieldEditKey` 分发。
- 启用 Export image as tar 后，Commit 成功会自动启动 Image Save 到本地 tar。
- tar 目标已存在时仍默认 Cancel，必须显式选择 Force。

## 13. 候选文件

第一阶段允许修改:

- `internal/tui/state/form.go`
- `internal/tui/state/app.go`（仅必要 Form 状态）
- `internal/tui/keyboard/container_form.go`
- `internal/tui/keyboard/input_edit.go`（仅复用，不破坏其他输入模式）
- `internal/tui/keyboard/keyboard.go`（仅 Form 子状态分发）
- `internal/tui/ui/widget/dialog/form.go`
- `internal/tui/ui/widget/dialog/` 下新增公共 form/select 文件
- `internal/tui/ui/component/` 下新增通用行布局 helper
- `internal/data/i18n/lang/en.jsonc`
- `internal/data/i18n/lang/zh.jsonc`
- `internal/data/i18n/lang/ja.jsonc`
- 对应测试文件

禁止修改:

- `internal/data/runtime/` 接口与 adapter。
- Docker / Podman driver。
- 与 Form 无关的表格页面。
- Action Bar 产品边界与全局快捷键。
- 无关文档。

## 14. 测试要求

### 状态与键盘

- 默认焦点为 Cancel。
- Enter 在普通字段间前进，不意外提交或取消。
- Up/Down 正确移动字段焦点，Tab/Shift+Tab 不切换组件。
- Bool Space 切换。
- Select Enter 打开、Up/Down、Space/Enter 确认、Esc 回滚。
- MultiSelect Space 多选、Enter 完成、Esc 回滚。
- Path Tab 单候选、多候选、无候选、目录继续补全。
- 子弹层 Esc 只关闭子弹层。
- Form 保留打开时的 resource ID。

### 路径

- `~`、环境变量、相对路径、绝对路径。
- 空目录、权限错误、不存在父目录。
- 隐藏文件显示规则。
- Linux / Windows 分隔符相关纯函数测试。
- 默认文件名字符清理和固定时间测试。
- 用户修改目标后不再被自动默认值覆盖。
- 已存在文件进入覆盖确认，默认 Cancel。

### 渲染

- Label 右对齐、Value 起点一致。
- 中英文 Label 使用可见宽度计算。
- 长路径不会撑破 Dialog。
- 小窗口不发生字段、按钮和边框重叠。
- Select/MultiSelect 每项独占一行。

### 验证命令

```bash
GOCACHE=/tmp/dtui-go-cache go test ./internal/tui/state ./internal/tui/keyboard ./internal/tui/ui/widget/dialog ./internal/tui/ui/component
GOCACHE=/tmp/dtui-go-cache go test -short ./...
GOCACHE=/tmp/dtui-go-cache go vet ./...
git diff --check
```

## 15. 小模型执行约束

1. 先读当前源码；当前工作区有未提交 BR-033 / BR-035 修改，必须保留。
2. 不得 `git reset`、`git checkout`、`git clean`。
3. 不 commit，不 push。
4. 第一阶段不得修改 runtime 或 adapter。
5. 不得使用 shell `ls/find` 完成容器路径补全。
6. 修改前先输出将修改的函数和文件。
7. 如果需要超出允许文件范围，停止并记录原因。
8. 完成后输出修改清单、测试结果、残余风险和第二阶段建议。

## 16. 验收标准

- [x] Form 视觉上表现为稳定的两列表格，所有 Value 左边界一致。
- [x] Select/MultiSelect 不再横向铺开，使用独立列表弹层。
- [x] Bool 可用 Space 勾选，状态清晰。
- [x] Copy / Export 打开时有合理的本地绝对目标路径和自动文件名。
- [x] 本地 Path 字段支持 shell 风格 Tab 补全。
- [x] 运行中 Linux 容器的 Container Path 支持异步 Tab 补全和目录候选弹层。
- [x] 目标文件存在时默认不覆盖，Force 语义明确。
- [x] 远程 daemon 场景明确区分 Local 与 Container path。
- [x] 不支持容器路径补全时仍可手工输入并执行 Copy。
- [x] 所有新增状态机、路径函数和渲染行为有 focused tests。
- [x] 全仓短测试和 vet 通过。
