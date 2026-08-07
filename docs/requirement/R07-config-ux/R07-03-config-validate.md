# R07-03 `dtui config validate` — 校验配置文件

## 元信息

- 状态: planned
- 优先级: high
- 来源: BR-043 follow-up（用户反馈：配置写错后 TUI 启动失败，报错不直观）
- 关联任务: ../../.omo/plans/config-ux.md（T8）
- 关联约束:
  - ../../constraint/C06-i18n.md

## 目标

用户改配置后希望在不启动 TUI 的情况下验证配置是否合法，且错误信息能精确定位到字段。`dtui config validate` 复用现有完整加载链，输出字段路径级错误。

## 用户流程

```text
$ dtui config validate
OK: /home/user/.config/docker-tui/config.yml

$ dtui config validate          # 配置有错时
Error: app.general.logLevel: unknown value "debug2" (valid: debug, info, warn, error)
(退出码 1)

$ dtui config validate          # 配置文件缺失时
No config file at /home/user/.config/docker-tui/config.yml — run `dtui config init`
(退出码 2)
```

## UI/UX

- 纯 CLI 输出，无 TUI 渲染。
- 错误按字段路径输出（如 `app.general.<field>`、`app.runtime.connections[0].endpoint`）。

## 功能规则

- 复用完整 `LoadResolved`（含 app 语义校验 + theme 加载 + appearance overrides 应用 + theme 校验），不新写校验逻辑。
- 退出码约定：成功 0 / 校验失败 1 / 配置文件缺失 2。
- 配置文件缺失时提示 `dtui config init`。

## 实现设计

- `cmd/docker-tui/main.go`：注册 `config validate` 子命令，调 `config.LoadResolved(config.LoadOptions{...})`。
- 错误输出：将返回 error 的字段路径部分提取并逐行打印。
- 缺失检测：`os.Stat(config.ConfigFile())` 不存在 → 退出码 2。

## 验收标准

- 单元测试（临时 HOME）：合法配置 → 退出 0 + OK 提示；坏配置 fixture → 退出 1 + 错误含字段路径；缺失文件 → 退出 2 + init 提示。

## 非目标

- 不新写任何校验规则（完全复用 LoadResolved）。
- 不修改 LoadResolved 签名。
- 不做配置修复/自动纠错。

## 迁移记录

- 从 BR-043 讨论与用户 2026-08-07 决策（validate 复用 LoadResolved、退出码 0/1/2）合并而来。
