# R02-03 docker / podman Registry Login(私库认证)

## 元信息

- 状态: planned
- 优先级: medium
- 来源: 原 BR-037 已合并归档
- 关联任务: 无
- 关联约束:
  - [../../constraint/C01-form.md](../../constraint/C01-form.md)
  - [../../constraint/C06-i18n.md](../../constraint/C06-i18n.md)

## 目标

为 docker / podman registry login 提供 UI 入口,允许用户在 dtui 中输入凭据完成私库认证,无需切换到命令行。

## 用户流程

镜像页 → Action Bar → Registry Login → Form(server / username / password / [storing]) → Submit → 调 runtime login → toast 反馈成功 / 失败。

## UI/UX

- Form 按 [R03-01](../R03-form-action/R03-01-form-pattern.md) 模板
- password 字段屏蔽显示
- storing 布尔(是否持久化凭据)

## 功能规则

- 必填:server、username、password
- password 字段 `EchoMode` 设为 mask
- storing=true 走 docker login(写 ~/.docker/config.json)或 podman 同等行为
- 失败时 toast 显示 error

## 实现设计

- 接口:`internal/data/runtime/auth.go:RegistryService.Login(ctx, server, user, pass, store)`
- Docker:`cli.RegistryLogin(ctx, auth)`
- Podman:对应 auth API
- 入口:`internal/tui/keyboard/auth.go:openRegistryLogin(m)`
- 表单:`internal/tui/ui/widget/dialog/registry_login.go`

## 验收标准

- Action Bar 镜像页可见 Registry Login 项
- password 屏蔽
- storing 选项控制持久化
- 双 runtime 测试通过

## 非目标

- 凭据管理 / 列表 / 删除
- OAuth / token 认证
- 多 registry 批量认证

## 迁移记录

- 旧文档:原 BR-037 已合并归档,内容汇总在本文档
- 保留信息:RegistryLogin interface 设计
- 待确认状态:`planned`,原 task 未实施;主模型确认后启动。