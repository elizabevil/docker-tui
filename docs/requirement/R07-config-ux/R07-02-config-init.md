# R07-02 `dtui config init` — 生成配置文件模板

## 元信息

- 状态: planned
- 优先级: high
- 来源: BR-043 follow-up（用户反馈：配置项多、不知从何写起）
- 关联任务: ../../.omo/plans/config-ux.md（T6/T7）
- 关联约束:
  - ../../constraint/C06-i18n.md

## 目标

用户需要一个「全量分节、逐字段注释」的配置文件模板作为起点。手工记忆 8 个分节的字段名与默认值成本高；`dtui config init` 一键生成带说明的完整模板，用户在此基础上删改。

## 用户流程

```text
$ dtui config init
Wrote /home/user/.config/docker-tui/config.yml (full annotated template)
```

- 首次执行：自动创建 `~/.config/docker-tui/` 目录并写入模板，提示路径。
- 再次执行（文件已存在）：拒绝覆盖，输出错误并非 0 退出。

## UI/UX

- 纯 CLI 输出，无 TUI 渲染。
- 模板为 YAML，按 `UserConfig` 结构分节：
  - `version`
  - `app.general` / `app.ui` / `app.docker` / `app.runtime` / `app.keymap` / `app.logs` / `app.layout` / `app.commands`
  - `appearance.theme` / `appearance.overrides`
- 每个字段带注释：字段名（yaml 标签）+ 默认值说明。

## 功能规则

- 目标文件固定为 `config.ConfigFile()`（`~/.config/docker-tui/config.yml`）。
- 父目录不存在时自动创建。
- 文件已存在：不覆盖，报错退出（非 0）。
- 模板不含 `operations` 节（operations 是 embedded 只读资源，无用户覆盖路径）。

## 实现设计

- `internal/data/config/template.go`：纯函数生成模板字节，结构驱动（手写分节映射，不用反射魔法）。
- `cmd/docker-tui/main.go`：注册 `config init` 子命令（orpheus `AddSubcommand` 嵌套），调模板生成 + 写盘。

## 验收标准

- 单元测试（golden 对比）：模板含全部 8 分节 + appearance，每节有注释，不含 operations。
- 集成测试（临时 HOME）：无文件 → 生成且内容=模板；有文件 → 非 0 退出且内容不变。

## 非目标

- 不提供交互式向导。
- 不覆盖已有文件（用户需手动删除）。
- 不生成 operations 配置。

## 迁移记录

- 从 BR-043 讨论与用户 2026-08-07 决策（init 全量分节模板）合并而来。
