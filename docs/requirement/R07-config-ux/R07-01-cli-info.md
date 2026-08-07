# R07-01 `dtui info` — 配置路径与主题列表

## 元信息

- 状态: planned
- 优先级: high
- 来源: BR-043 follow-up（用户反馈：配置项多、找不到配置文件位置）
- 关联任务: ../../.omo/plans/config-ux.md（T4/T5）
- 关联约束:
  - ../../constraint/C06-i18n.md

## 目标

用户需要快速了解「配置写在哪里、日志在哪里、主题有哪些」。当前只有 `--list-themes` 一个 flag 且不显示任何路径；`--config`/`--host` 等 flag 的位置信息分散。提供一个 `dtui info` 命令，一次输出全部关键路径与主题列表。

## 用户流程

```text
$ dtui info
Config file:   /home/user/.config/docker-tui/config.yml
Config dir:    /home/user/.config/docker-tui/
Log dir:       /home/user/.local/state/docker-tui/logs/     (以实际实现为准)
Theme dir:     <builtin themes: 6>
Themes:        default, dark, dracula, nord, solarized, light
```

- 终端执行 `dtui info`，无交互，直接打印并退出（退出码 0）。

## UI/UX

- 纯 CLI 输出，无 TUI 渲染。
- 输出顺序固定：配置文件路径 → 配置目录 → 日志目录 → 主题列表。
- 主题列表来源为现有 `config.ListThemes()`（internal/data/config/theme.go:604）。

## 功能规则

- 输入：无参数。
- 输出：上述路径与主题，每行一项，可读格式。
- 退出码：成功 0；路径解析失败非 0。
- `--list-themes` 全局 flag 移除，其能力并入本命令。

## 实现设计

- `cmd/docker-tui/main.go`：用 orpheus `AddSubcommand` 注册顶级 `info` 命令。
- 复用 `config.ConfigDir()` / `config.ConfigFile()`（internal/data/config/config.go:14/22）与 `config.ListThemes()`。
- 日志目录常量复用既有路径常量（internal/data/config/constants_files.go:4-6 附近的日志目录常量）。

## 验收标准

- `go run ./cmd/docker-tui info` 退出码 0，输出含配置文件完整路径与全部主题名。
- `go run ./cmd/docker-tui --help` 不再显示 `--list-themes`。
- 单元测试覆盖输出内容与退出码。

## 非目标

- 不打印配置内容本身（那是 `config validate`/cat 的事）。
- 不做交互式选择器。
- 不改变 `--config`/`-f` flag 的现有语义。

## 迁移记录

- 从 BR-043 讨论（`docs/handoff/proposals/config-split-review.md`）与用户 2026-08-07 决策（`--list-themes` 并入 info）合并而来。
