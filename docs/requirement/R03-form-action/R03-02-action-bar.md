# R03-02 Action Bar

## 元信息

- 状态: implementing
- 优先级: high
- 来源: [../../task/br-039-action-bar-replace-q.md](../../task/br-039-action-bar-replace-q.md)、[../../task/br-039-a-implementation-design.md](../../task/br-039-a-implementation-design.md)、[../../task/br-039-code-location-analysis.md](../../task/br-039-code-location-analysis.md)
- 关联任务: [../../task/br-039-main-model-feedback.md](../../task/br-039-main-model-feedback.md)
- 关联约束:
  - [../../constraint/C04-keybinding.md](../../constraint/C04-keybinding.md)
  - [../../constraint/C03-table.md](../../constraint/C03-table.md)

## 目标

用 Action Bar 取代旧 Q 退出键,承载每个页面的多动作入口(容器页 6 动作、镜像页 Save/Load/Import、卷页 Create 等)。

## 用户流程

任意 ModeNormal 页面 → 触发 Action Bar 键(`:` 或 Action Bar 入口键)→ 列出当前页可用动作 → 上下选择 → Enter 触发。

## UI/UX

- 居中弹层,按 [C02-dialog](../../constraint/C02-dialog.md) 尺寸规则
- 列表显示:`▶` cursor + 动作名 + 简短描述
- 不可用动作灰显
- 底部状态栏显示 Action Bar 入口键

## 功能规则

- 容器状态过滤(running / stopped / paused)
- 选中目标资源为空时灰显相关动作
- Esc 关闭,Enter 触发
- 不承载全局命令(走 `:` command palette)

## 实现设计

- 状态:`internal/tui/ui/widget/actionbar/` 已有 `ModeActionBar`
- 注册:`internal/tui/actionbar/registry.go` 按 panel 注册动作
- 触发:`internal/tui/keyboard/actionbar_keys.go` 按 Mode / Panel 分发
- 帮助投影:`internal/tui/ui/action/registry.go`
- 与 [C04-keybinding](../../constraint/C04-keybinding.md) 集成

## 验收标准

- 每页至少 1 个 Action Bar 入口
- 容器页 6 动作 + 镜像页 Save/Load/Import 可见且可用
- 状态过滤正确(running / stopped)
- 弹层 Esc 关闭不破坏页面状态

## 非目标

- 自定义动作(用户扩展)
- 多级菜单

## 迁移记录

- 旧文档:[task/br-039-action-bar-replace-q.md](../../task/br-039-action-bar-replace-q.md) 等 4 个 br-039 文件
- 保留信息:Action Bar 注册表 + 触发链 + Help 投影
- 废弃信息:Q 退出键的兼容实现
- 待确认状态:`implementing`,Action Bar 主路径已就绪,各页动作注册完整性需主模型逐项确认。