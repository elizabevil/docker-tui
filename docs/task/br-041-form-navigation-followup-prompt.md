# BR-041 小模型执行指令：Form 导航与选择弹层收尾

> 日期: 2026-08-02  
> 类型: 允许修改代码  
> 提交规则: 不 commit、不 push  
> 范围: 仅完成 BR-041 第一阶段新增交互要求

## 1. 开始前必须读取

- `docs/ai-prompts.md`
- `docs/task/br-041-unified-form-path-completion.md`
- `internal/tui/state/form.go`
- `internal/tui/state/path.go`
- `internal/tui/keyboard/container_form.go`
- `internal/tui/keyboard/input_edit.go`
- `internal/tui/keyboard/keyboard.go`
- `internal/tui/ui/component/formrow.go`
- `internal/tui/ui/widget/dialog/form.go`
- `internal/tui/ui/widget/dialog/form_popup.go`
- 上述文件对应的全部测试

当前工作区包含用户、主模型和其他小模型的未提交修改。必须保留，不得 reset、checkout、clean、覆盖或恢复无关文件。

## 2. 当前事实

已经实现:

- 公共 FormState、FormPath、FormSelect、FormMultiSelect、FormBool。
- 本地路径 provider、默认 Copy/Export 文件名、覆盖 Force 确认。
- 两列 Form 基础布局。
- Select/MultiSelect/Path 候选弹层。

仍需完成:

- Tab 仍有旧的组件切换语义，需要彻底移除。
- Form 主页面缺少完整方向键空间导航。
- Select/MultiSelect/Path 弹层缺少 Home/End、PgUp/PgDn。
- 弹层 cursor、checkbox/type、文本列排版需要统一。
- 当前行需要稳定背景高亮，不能只依赖前景色。
- 需要补小窗口和长内容渲染测试。

## 3. 必须实现的键盘状态机

### 3.1 Form 主页面

| 按键 | 必须行为 |
|---|---|
| `Up` | 移动到上一个字段；从首字段进入按钮区 |
| `Down` | 移动到下一个字段；从末字段进入按钮区 |
| `Left/Right` | Text/Int/Path 中移动文本光标；按钮区切换 Cancel/Confirm |
| `Space` | Bool 切换；Text/Path 中输入空格；Select/MultiSelect 不直接横向切换 |
| `Enter` | 普通字段进入下一字段；Select/MultiSelect 打开弹层；按钮执行 |
| `Tab` | 仅对支持补全的字段执行补全；不得切换字段或按钮 |
| `Shift+Tab` | 不切换字段或按钮；有候选弹层时可反向移动候选 |
| `Esc` | 无子弹层时关闭 Form |

按钮区:

- Cancel 在左，Confirm 在右。
- 默认焦点为 Cancel。
- Left/Right 切换按钮。
- Up 返回最后一个字段。
- Down 保持按钮区，不执行、不循环。
- Enter 执行当前按钮。

### 3.2 Text / Int / Path

- Left/Right 调用现有 QueryInput 光标移动能力。
- Home/End 移到输入开头/结尾。
- Backspace/Delete 保持现有编辑语义。
- Up/Down 只移动字段焦点，不修改文本。
- 长输入必须保持光标可见。

### 3.3 Path 补全

- Tab 请求/应用补全。
- 单候选直接补全。
- 多候选补全公共前缀并打开候选弹层。
- 弹层已打开时重复 Tab 向下移动候选。
- Shift+Tab 向上移动候选。
- 无候选时保持输入和焦点，并显示非阻塞提示。
- 目录候选 Enter 后进入该目录并继续显示其候选。
- 文件候选 Enter 后写入字段并关闭弹层。

### 3.4 Select 弹层

- Up/Down 移动一项。
- Home/End 到首项/末项。
- PgUp/PgDn 按可见行数分页。
- Space 或 Enter 确认当前项并关闭。
- Esc 关闭并保留打开前的值。
- Left/Right 不切换选项。

### 3.5 MultiSelect 弹层

- Up/Down、Home/End、PgUp/PgDn 导航。
- Space 切换当前项工作副本。
- Enter 提交全部选择并关闭。
- Esc 丢弃工作副本并关闭。

## 4. 必须实现的布局

### 4.1 Form

- Dialog 整体相对当前 panel 居中。
- 标题左对齐。
- Label 列右对齐。
- Value 列左对齐，所有 Value 左边界一致。
- 当前字段高亮不得改变宽度或列位置。
- 字段错误与 Value 列左边界一致。
- 长路径不撑破 Dialog。
- Cancel / Confirm 按钮组可整体居中，但组内 Cancel 左、Confirm 右。

### 4.2 Select

固定列结构:

```text
>   unchanged
    no
    always
```

- cursor 列固定宽度。
- 文本列左对齐。
- 当前行使用背景色和 cursor 标记。
- 选项不能居中。

### 4.3 MultiSelect

固定列结构:

```text
> [x] read
  [ ] write
  [x] inspect
```

- cursor、checkbox、文本分别为固定列。
- 当前 cursor 行仍必须显示 `[x]` 或 `[ ]`。
- 文本左对齐。

### 4.4 Path 候选

固定列结构:

```text
> [DIR]  backups/
  [DIR]  exports/
  [FILE] nginx.tar
```

- cursor、类型、basename 分列。
- 类型列固定宽度。
- basename 左对齐。
- 当前行使用背景色。
- 完整父路径放在标题或底部说明，不横向撑宽候选行。
- 最多显示固定数量可见行，cursor 移动时窗口跟随。

## 5. 实现边界

允许修改:

- `internal/tui/state/form.go`
- `internal/tui/keyboard/container_form.go`
- `internal/tui/keys/keys.go`（仅必要按键常量）
- `internal/tui/ui/component/formrow.go`
- `internal/tui/ui/widget/dialog/form.go`
- `internal/tui/ui/widget/dialog/form_popup.go`
- 上述文件对应测试
- `internal/data/i18n/lang/en.jsonc`
- `internal/data/i18n/lang/zh.jsonc`
- `internal/data/i18n/lang/ja.jsonc`
- `docs/task/br-041-unified-form-path-completion.md`（仅同步实际完成状态）

禁止修改:

- `internal/data/runtime/`
- Docker / Podman adapter 和 driver
- Image Save / Load / Import workflow
- Action Bar 产品边界
- 全局非 Form 快捷键
- 其他页面布局
- 无关文档

若实现需要超出允许范围，停止并报告原因，不自行扩大范围。

## 6. 测试要求

必须新增或修正测试:

1. 默认 Cancel，Left/Right 在按钮间切换。
2. Up/Down 从按钮区进入字段并遍历全部字段。
3. Tab/Shift+Tab 在所有非 Path 字段不改变焦点。
4. Tab 在 Path 字段覆盖无候选、单候选、多候选。
5. 候选弹层重复 Tab/Shift+Tab 正反移动。
6. Text/Path Left/Right/Home/End 移动光标。
7. Select Home/End/PgUp/PgDn/Space/Enter/Esc。
8. MultiSelect Space 工作副本、Enter 提交、Esc 回滚。
9. 路径目录与文件 Enter 行为不同。
10. 当前 MultiSelect 行仍显示 checkbox。
11. cursor/type/checkbox/text 列左边界一致。
12. 英文、中文 Label 的 Value 起点一致。
13. 长路径光标保持可见。
14. 40x16、80x24、160x40 三组 viewport 下不发生边框、字段、按钮重叠。

运行:

```bash
GOCACHE=/tmp/dtui-go-cache go test ./internal/tui/state ./internal/tui/keyboard ./internal/tui/ui/component ./internal/tui/ui/widget/dialog
GOCACHE=/tmp/dtui-go-cache go test -short ./...
GOCACHE=/tmp/dtui-go-cache go vet ./...
git diff --check
```

## 7. 完成条件

- Tab/Shift+Tab 不再承担 Form 焦点切换。
- Form 可完全通过方向键访问所有字段和按钮。
- 三种候选弹层均为稳定的左对齐列表布局。
- 所有新增状态机行为有 focused tests。
- 全仓短测试和 vet 通过。
- 没有修改 runtime、Image workflow 或其他任务范围。
- 没有 commit、没有 push。

## 8. 最终输出

1. 修改文件清单。
2. 最终键盘状态机说明。
3. 三种弹层布局说明。
4. 测试命令和结果。
5. 未完成项或需要主模型确认的风险。
6. 明确声明未 commit、未 push。
