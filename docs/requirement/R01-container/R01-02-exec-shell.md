# R01-02 容器 Exec 页面 shell UI

## 元信息

- 状态: planned
- 优先级: medium
- 来源: 原 BR-038 已合并归档
- 关联任务: 无
- 关联约束:
  - [../../constraint/C04-keybinding.md](../../constraint/C04-keybinding.md)
  - [../../constraint/C03-table.md](../../constraint/C03-table.md)

## 目标

为容器页 Exec 动作提供独立的 shell UI 页面,支持交互式输入输出与多行编辑。

## 用户流程

容器页选中容器 → `e` 键 / Action Bar → `ModeExec` → 终端式 UI → 用户输入命令 → 看到实时输出 → `Esc` 退出(回到原页面)。

## UI/UX

- 全屏或主区域占满的 shell 视图
- 输出区可滚动(回看历史)
- 输入区底部高亮,支持左右移动光标编辑当前行
- 快捷键提示在底部:`Esc back`、`Ctrl+C cancel command`

## 功能规则

- 透传按键到容器内的 shell(`charmbracelet/x/term` 或类似方案)
- 进程生命周期与模式绑定:Mode 退出时关闭连接
- 输出追加到滚动缓冲区(类似 logs 页)
- 当前命令未结束前 `Ctrl+C` 终止当前命令,不退出模式

## 实现设计

- 状态:`internal/tui/state/exec.go` `ExecState` 已有 `ModeExec` 与 `ModeExecShell`、`ModeExecPassthrough`
- 页面:`internal/tui/ui/pages/exec/` 现有 shell 渲染
- 键盘:`internal/tui/keyboard/keyboard.go` `ModeExec` / `ModeExecPassthrough` 分发已存在
- 与 [R03-02](../R03-form-action/R03-02-action-bar.md) Action Bar 入口整合

## 验收标准

- `e` 键从容器页进入 Exec
- 输入字符实时送达容器 shell
- 输出实时回显,带颜色
- `Esc` 安全关闭连接并退出
- `Ctrl+C` 中断当前命令但不退出

## 非目标

- 多窗口 / 多 tab
- 命令历史持久化

## 迁移记录

- 旧文档:原 BR-038 已合并归档,内容汇总在本文档
- 保留信息:ModeExec 状态、键盘分发模式
- 待确认状态:`planned`,原 task 标记 incomplete;主模型确认后启动实施。