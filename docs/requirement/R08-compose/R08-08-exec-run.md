# R08-08 一次性执行(run / exec)

## 元信息

- 状态: implementing
- 优先级: medium
- 来源: docker compose CLI 权威 + 用户评审结果
- 关联任务: [.omo/compose-todo.md 附录 A.4 / §3.4 / 补 F.1](../../../.omo/compose-todo.md)
- 关联约束:
  - [../../constraint/C01-form.md](../../constraint/C01-form.md)
  - [../../constraint/C05-path.md](../../constraint/C05-path.md)

## ⚠ 重复功能说明

| 能力 | 已实现层 | 本需求增量 |
|---|---|---|
| `docker compose exec <svc> <cmd>` | R01 容器 exec(`containerExec` + `m.Exec`) | **完全复用** — Compose-级 exec = 选 service → 第一个容器 → 走 R01 exec |
| `docker compose run <svc> <cmd>` | R01 容器运行(无 `compose run` 语义) | **R08-08 增量**:创建一次性容器,带 `com.docker.compose.oneoff=true` label,与项目其它容器同 label |

## 目标

Compose-级一次性执行。**exec 完全复用 R01**(选 service → 选 replica → R01 exec);**run 是真正的增量**(创建一次性容器)。

## 用户流程

### exec

1. Compose 面板 → 选中 service → Action Bar → Exec,或快捷键 `e`(注册)
2. **service replicas > 1**:List Dialog 让用户选具体 replica;**= 1**:直接走 R01 exec
3. 进入 R01 exec 终端(沿用 exec shell)

### run

1. Compose 面板 → 选中 service → Action Bar → Run
2. Form(参考 R02 image transfer Form 框架):
   - 命令(command line)
   - `--rm`:自动清理(默认 true)
   - `--no-deps`:不启动依赖
   - `--entrypoint`:覆盖 entrypoint
   - 标签 / 环境变量(可选,Form 字段按 R01 + R02 既有模式)
3. Confirm → 创建容器(带 compose labels)→ attach exec

## UI/UX

仅本需求特有:

- exec 子流程的 replica 选 List Dialog:沿用 [C02-dialog](../../constraint/C02-dialog.md)
- run Form:沿用 [C01-form](../../constraint/C01-form.md) 字段约束

详细版式 UI overhaul。

## 功能规则

### F1 exec 复用策略

- 在 service 多 replicas 场景提供 List Dialog(用户明确选择)
- 在单 replica 场景直接进入 exec
- audit:`compose_service.exec` trace(R01 容器 exec trace 不变 — run 才单独 trace)

### F2 run 实现

```go
type RunOptions struct {
    Service       string
    Command       []string
    RemoveAfter   bool   // --rm
    NoDeps        bool   // --no-deps
    EntrypointOverride string
    EnvOverrides  map[string]string
    LabelOverrides map[string]string
    Detach        bool   // -d
    User          string // -u
}

func (s *DockerComposeService) Run(ctx context.Context, project string, opts RunOptions) error
```

实现(不解析 yaml):

1. 通过 `ComposeService.InspectProject(project)` 拿到服务的 image,环境变量,网络,卷列表
2. 找到该 service 的**最新**容器(配置 hash +1 + `container-number=max+1`),复制其 network / volume 设置
3. 创建新容器,加 labels:
   - 复用 `com.docker.compose.project / service / config-hash` 等
   - 加 `com.docker.compose.oneoff: "true"`
   - 加 `dtui.compose.run.timestamp: <rfc3339>`(标识 run 创建)
4. 启动 + attach exec(用户 detach 时返回)
5. `--rm` 时:进程退出后调 `ContainerRemove`

### F3 ContainerNumber 自增

dtui 不解析 docker compose 的"用户配置"。`container-number` 的下一个值通过以下方式推断:

- 列出该项目所有该 service 容器的 `com.docker.compose.container-number`(int)
- 取 `max + 1`
- 边界:并发 run 时可能冲突 — 给 WARN toast,不阻塞(Docker Engine 容错)

### F4 audit

`compose_service.run` 与 `compose_service.exec` 两条 trace。run 含 oneoff label 值,便于在 audit history 中区分"运行 vs 常驻"。

### F5 entrypoint 覆盖

Form 字段允许字符串,默认空(不覆盖)。空时使用 service 默认 entrypoint(取自先前 inspect)。

## 实现设计

涉及模块:

- 新方法:`ComposeService.Run / Exec`(R08-02 文件追加)
- 修改:`internal/tui/keyboard/compose_action.go` 增加 `doComposeRun / doComposeExec`
- 新文件:`internal/tui/ui/pages/compose/run_form.go`(Form 字段按 R01 + R02 模式)
- 复用:`internal/tui/keyboard/container_action.go` 的 `ContainerExec`(compose 一侧委派)

调用链:

```text
doComposeRun(project, opts)
  -> ComposeService.Run(ctx, project, RunOptions)
     1) InspectProject → service config (image, env)
     2) ContainerList label-filter → replicas + container-number max
     3) ContainerCreate (with oneoff=true label)
     4) ContainerStart
     5) Detach=false: ContainerAttach + ContainerExec(等同 R01)
        Detach=true: return immediately, audit "detach run"
     6) --rm=true: ContainerWait + ContainerRemove

doComposeExec(project, service)
  -> replicas=1: 直接委派 R01 exec
     replicas>1: 选 list → 委派
```

## 验收标准

1. 单 replica service 的 exec 直接打开 R01 exec 终端
2. 多 replicas exec:List Dialog 显示容器标签 + IP + 状态
3. run 创建容器包含 `com.docker.compose.oneoff=true` label
4. `container-number` 自增符合 max+1
5. `--rm=true` 退出后自动删除容器
6. audit 两条 trace 区分 run / exec
7. 无效 service(项目无此 service)返回 `ErrComposeServiceNotFound`

## 非目标

- run + 同时多个命令(连续运行多个容器)— 留 future
- 自动依赖 service 启动(`docker compose run --service-ports` 的 deps)— 留 future(Yaml 不解析,无法知道 deps)

## 迁移记录

- 旧代码:无
- 复用:R01 ContainerExec 完整逻辑
- 保留信息:R01 exec 的 shell UI / audit
- 废弃信息:无
- 待确认:List Dialog 在小屏的降级
