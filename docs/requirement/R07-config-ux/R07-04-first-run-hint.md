# R07-04 首次运行提示（toast）

## 元信息

- 状态: planned
- 优先级: medium
- 来源: BR-043 follow-up（用户反馈：不知道有 `dtui config init` 可生成模板）
- 关联任务: ../../.omo/plans/config-ux.md（T9/T10）
- 关联约束:
  - ../../constraint/C06-i18n.md（文案三语 en/zh/ja）

## 目标

用户首次运行 dtui（尚无配置文件）时，TUI 内以 toast 提示存在配置模板命令，引导用户探索配置能力。**不自动生成配置文件**——自动生成会改写用户环境，属越权行为；只提示、不动作。

## 用户流程

- 首次运行（`~/.config/docker-tui/config.yml` 不存在）→ TUI 启动后显示一次 toast：
  - en: `No config file — run 'dtui config init' to create one`
  - zh: `未找到配置文件 — 运行 'dtui config init' 生成`
  - ja: `設定ファイルがありません — 'dtui config init' で作成できます`
- 已有配置文件 → 无 toast。

## UI/UX

- 复用现有 Feedback toast 机制（`m.model.Feedback.ToastGeneration/ToastTimer`）。
- toast 位于 TUI 现有 toast 位置（右上角），不阻塞交互，可被常规 toast 操作清除。

## 功能规则

- 触发条件：启动时 `config.ConfigFile()` 不存在。
- 不创建任何文件、不修改任何目录。
- 不阻塞 TUI 启动（toast 异步显示）。
- 文案通过 i18n key（如 `toast.firstRunHint`）三语提供。

## 实现设计

- `internal/data/i18n/lang/{en,zh,ja}.jsonc`：新增 `toast.firstRunHint`。
- `cmd/docker-tui/main.go` runTUI：加载 resolved 配置前先 stat ConfigFile，缺失则在 Feedback 中设置 toast。

## 验收标准

- 测试：ConfigFile 缺失 → toast 被设置；存在 → 不设置。
- 三语 key 一致性测试通过。

## 非目标

- 不自动生成配置文件（显式非目标）。
- 不做「首次运行向导/多步引导」。
- 不做运行后定时重复提醒（只在启动时检测）。

## 迁移记录

- 从用户 2026-08-07 决策（首次运行不自动生成、仅 toast 提示）合并而来。
