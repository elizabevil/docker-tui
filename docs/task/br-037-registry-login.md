# docker / podman Registry Login (私库认证)

> 任务卡 / BR-037

## 元信息

- **关联编号**:BR-037
- **优先级**:medium
- **状态**:`open`
- **依赖**:BR-018(镜像传输 / pull 框架),BR-039(Action Bar 入口)
- **关联设计**:[docs/feature-design.md §5.9](../feature-design.md)

## 目标

为镜像 Action Bar / 命令面板增加 `Login` 动作,允许用户在 dtui 内完成 `docker login` / `podman login`,凭证由引擎自身(`~/.docker/config.json`)管理。

## 代码结构索引

### 必须读懂的文件

| 文件 | 作用 |
|---|---|
| `internal/data/runtime/image.go:118` | `ImageService` 接口(无 Login) |
| `internal/data/runtime/docker/service_image.go` | Docker `imageService` 实现 |
| `internal/data/runtime/podman/service_image.go` | Podman `imageService` 实现 |
| `internal/tui/keys/registry.go` | 现有镜像动作注册模板(Load / Save 等) |
| `internal/data/i18n/lang/*.jsonc` | 加 `image.login.title` / `image.login.server` / `image.login.user` 等 |

### 必须修改的文件

| 文件 | 改动 |
|---|---|
| `internal/data/runtime/engine.go:14` | `Engine` 接口加 `Login(ctx, server, user, password, opts) error` |
| `internal/data/runtime/docker/service_image.go` | Docker `Login`:`cli.RegistryLogin(ctx, auth)`(Docker SDK 自带) |
| `internal/data/runtime/podman/service_image.go` | Podman `Login`:`client.Login(ctx, server, user, password)`(podman REST `/auth`) |
| `internal/tui/keys/action.go` | 加 `ActionRegistryLogin` |
| `internal/tui/keys/registry.go` | 注册 `{ActionRegistryLogin, []string{...}, app}` (或 Action Bar) |
| `internal/tui/keyboard/image_action.go` | 加 `doRegistryLogin(m)` + 弹 dialog |

### 必须新增的文件

- `internal/tui/ui/widget/dialog/login.go`:Login 表单(服务器 / 用户名 / 密码(掩码) / TLS 高级选项)

## 操作链路

```text
User 在镜像页 → Action Bar 触发 Login
  → keyboard/image_action.go:doRegistryLogin(m)
      弹 widget/dialog/login.go 表单
      用户输入 server(默认 docker.io)/ user / password
      提交 → registryLoginCmd(client, server, user, password, opts)
  → tea.Cmd 异步执行
      Engine.Login(ctx, server, user, password, opts)
        Docker: cli.RegistryLogin → auth.AuthConfig{Username, Password, Serveraddress}
        Podman: REST /libpod/_ping 或 /auth (按实现)
  → 成功:toast `✓ Login <server>` + audit("resource.registry.login")
  → 失败:toast `Login failed: <err>`(不泄露密码)
  → dtui 不保存凭证(由 ~/.docker/config.json 管理)
```

## 验收标准

- [ ] Action Bar / 命令面板中"Registry Login"可点击。
- [ ] 表单可填写 server / user / password。
- [ ] 登录成功后,dtui 内 `ImagePull` (Ctrl+P) 私有仓库镜像成功。
- [ ] Docker 与 Podman 两端都接好。
- [ ] 错误凭证返回明确 toast(`Unauthorized` / `401`),**不泄露密码**。
- [ ] dtui 进程退出后,引擎凭证仍存在(由引擎自身管理 `~/.docker/config.json`)。
- [ ] audit 写入 `resource.registry.login`。
- [ ] 密码输入使用掩码(`*` / `•`)。

## 风险

| 风险 | 说明 |
|---|---|
| **密码内存安全** | 密码字符串需及时清零(Go GC 不保证);使用后 `defer func(p []byte) { for i := range p { p[i] = 0 } }(password)` 之类 |
| **凭证存储边界** | dtui 不存密码,只调 engine;需明确文档,避免后续实现误存到 dtui 配置 |
| **Podman Login API 差异** | Podman REST `/auth` 端点与 Docker 不完全兼容,需要做 capability 检查 |
| **Help 漂移** | 新增 Login 动作,Help 页 + Footer 需同步 |

## 建议任务分解

1. **TASK-BR037-A:runtime 接口 + 实现**
   - `Engine.Login` + Docker `cli.RegistryLogin` + Podman REST
   - **预估**:主模型实现,需审凭证流
2. **TASK-BR037-B:TUI handler + dialog**
   - `doRegistryLogin` + 掩码输入 widget + toast 不泄露
   - **预估**:小模型辅助,主模型审密码处理
3. **TASK-BR037-C:Action Bar 集成 (BR-039 协同)**
   - 在 Action Bar 中加"Login"
   - **预估**:等 BR-039
4. **TASK-BR037-D:i18n + Help 同步**
   - 翻译 + docs
   - **预估**:小模型机械

## 待确认项

- [ ] Podman Login REST 端点具体是哪个?(`/auth` 还是 `/system/login`?)
- [ ] 是否需要"登出"(Logout) 动作(对应 `~/.docker/config.json` 删除条目)?
- [ ] 密码字段是否走 lipgloss 掩码 widget 还是 input_edit.go 现有方案?
- [ ] 是否需要保存"最近登录的服务器"以便自动填充?