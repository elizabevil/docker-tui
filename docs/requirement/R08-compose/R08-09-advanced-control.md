# R08-09 高级控制(pause / unpause / kill / scale / rm / prune)

## 元信息

- 状态: implemented
- 优先级: low
- 来源: docker compose CLI 权威 + 用户评审结果
- 关联任务: [.omo/compose-todo.md 附录 A.1 / A.5 / A.6 / §3.6](../../../.omo/compose-todo.md)
- 关联约束:
  - [../../constraint/C01-form.md](../../constraint/C01-form.md)
  - [../../constraint/C02-dialog.md](../../constraint/C02-dialog.md)
  - [../../constraint/C04-keybinding.md](../../constraint/C04-keybinding.md)

## ⚠ 重复功能说明

| 能力 | 已实现层 | 本需求增量 |
|---|---|---|
| `docker compose pause / unpause / kill` | R01 容器(`containerPause / Unpause / Kill`) | R08-09:Compose-级 = 该 service 或项目所有容器的批量执行 |
| `docker compose scale <svc>=N` | R01 容器无 scale 概念 | **R08-09 增量**:scale = 增加 replicas,符合 `com.docker.compose.container-number` label 语义 |
| `docker compose rm` | R01 容器 remove(form) | R08-09:Compose-级 = 删该项目所有 stopped 容器(service alive 单容器不动) |
| `docker compose prune` | R04 runtime prune(无 compose 范围) | **R08-09 增量**:Compose-级 prune = 删 stopped / orphan 容器 + 项目命名卷(可选) |

## 目标

Compose-级"杂项"动作集:`pause / unpause / kill / scale / rm / prune`。前三个**完全依赖 R01 容器级 API**,只是批处理入口;scale 与 R08-04 类似需要新建容器;rm / prune 是项目级清理。

## 用户流程

各动作的入口与触发见 R08-11 Action Bar 注册。本节只列**用户可见的关键差异**:

### scale

1. Form 输入(参考 R01 ContainerForm 模式):
   - 服务名(input)
   - 目标副本数(spinner int)
   - `--no-deps`
2. Confirm → 视情况创建 / 删除 replicas
3. 增加 replicas:走 R08-08 的 `RunOptions` 路径(创建容器 + start)
4. 减少 replicas:ContainerStop + ContainerRemove(保留最新 N 个)

### rm

1. Confirm Dialog(默认 `--all=false`,只删 stopped)
2. 选项:
   - `[ ] --all` 包括 running
   - `[x] --stopped-only` 默认
3. Confirm 后:批量删容器,返回 BatchActioned

### prune

1. Confirm Dialog(默认行为友好)
2. 行为:对**所有**该项目命名空间(`com.docker.compose.project=<project>`)的 stopped 容器 / orphan 容器 + 项目命名卷(可选)批量删
3. 用户可选项:`--include-volumes`
4. Confirm 后:BatchActioned 报告删除数量

## UI/UX

仅本需求特有:

- scale Form:沿用 [C01-form](../../constraint/C01-form.md),数值字段带 spinner
- rm / prune Confirm Dialog:沿用 [C02-dialog](../../constraint/C02-dialog.md) + 已有 confirmAction 组件
- toast 报告:**scoped 计数**(per-service 增长或减少 / pruned 数量 / 跳过的 named volume)

完整版式 UI overhaul。

## 功能规则

### F1 pause / unpause / kill

```go
// 全部委派到容器级
func (s *DockerComposeService) Pause(ctx, project, services []string) error
func (s *DockerComposeService) Unpause(ctx, project, services []string) error
func (s *DockerComposeService) Kill(ctx, project, services []string, signal string) error
```

- `services` 空 → 项目所有容器;非空 → 仅匹配 `service ∈ services` 的容器
- audit:`compose_service.pause / unpause / kill` 一条 trace,内含 counts

### F2 scale

```go
type ScaleOptions struct {
    Service    string
    Replicas   int // 目标副本数
    NoDeps     bool
}

func (s *DockerComposeService) Scale(ctx, project string, opts ScaleOptions) error
```

实现语义:

1. 当前 replicas = `count(com.docker.compose.project=<p>, service=<s>)`
2. 目标 > 当前:`replicas = target - current` 次创建(沿用 RunOptions,不增 oneoff,不带 `--rm`)
3. 目标 < 当前:保留最新的 `target` 个(按 container-number 排序),删除其余
4. audit:`compose_service.scale` trace + 每个被创建 / 删除的容器 trace

### F3 rm

```go
type RmOptions struct {
    StoppedOnly bool   // 默认 true
    Force       bool   // -f
    All         bool   // --all
}

func (s *DockerComposeService) Rm(ctx, project string, opts RmOptions) error
```

### F4 prune

```go
type PruneOptions struct {
    IncludeVolumes bool
    DryRun         bool
}

func (s *DockerComposeService) Prune(ctx, project string, opts PruneOptions) (ComposePruneReport, error)
```

`ComposePruneReport` 报告每个资源类型的删除数量:

```go
type ComposePruneReport struct {
    Containers int
    Networks   int
    Volumes    int
}
```

`DryRun=true` 时:列出**会**被删的资源但**不**删,UI 给 Confirm。

## 实现设计

涉及模块:

- 新增方法:`ComposeService.Pause / Unpause / Kill / Scale / Rm / Prune`(R08-02 追加)
- 修改:`internal/tui/keyboard/compose_action.go` 加 `doComposePause / Unpause / Kill / Scale / Rm / Prune`
- 新文件:`internal/tui/ui/pages/compose/scale_form.go`
- 复用:`internal/data/runtime/docker/containers.go` 的 `ContainerRemove / ContainerPause` 等

## 验收标准

1. service 级 pause 影响该项目该 service 的所有容器
2. scale=2 → 3 时创建 1 个新容器,`com.docker.compose.container-number` label 为 `max+1`
3. scale=2 → 1 时保留 container-number 最大的容器,删除其他
4. rm 只删 stopped 容器,默认;加 `--all` 删 running(强制 case 由 `--force` 标志)
5. prune 在 DryRun 时只列出,确认后才删
6. audit 链路:`compose_service.prune` + 每个被删资源一条 trace,便于 history 查询
7. Action Bar 注册:`OperationScopeCompose` 含 6 项(见 R08-11)

## 非目标

- **prune 跨项目**(留 future enhancement 或单独 BR)
- **kill signal 自定义**(`-s SIGTERM` 等)— Phase 后置,先用 `SIGKILL`
- **global prune**(整体 daemon,不是项目级别)
- **与系统 disk usage 的联动**(如 `docker system df`)

## R08-14 集成

`ScaleOptions` 在 CoLocated group 触发时的行为(详见 [R08-14 §F5](./R08-14-co-located-group.md)):

| 运行时 | group 内 service scale 行为 |
|---|---|
| Docker(network_mode: service: 触发的 group) | scale 强制 = 1,UI 灰显 scale Action + toast 提示 "Docker compose 限制 network_mode: service: 的 service 不能 scale>1" |
| Podman(podman-compose 默认 1 项目 1 pod) | scale 允许独立,只 share namespace 不冲突 host port |

实现:

- `doComposeScale` 接收可选 `GroupID`,启用 group 模式时根据 `Runtime` 字段判定是否 UI 拦截
- Action Bar 项 `compose_project.scale` 视 context 自动提示 "(Docker group 锁定为 1)" 或默认允许
- audit trace:group scale 时 trace 含 `GroupID`,普通 scale 不含

## 迁移记录

- 旧代码:无
- 复用:R01 容器动作 / R05 events-network 的 prune 思路
- 保留信息:audit trace 模式
- 废弃信息:无
- 待确认:scale 时 `--no-deps` 的 deps 概念 — 没有 yaml 解析,所以这个标志**仅作 Form 占位**,实际行为 = 全部启动无论 deps

### 2026-08-08 阶段 C 落地

**已完成**(本轮):

- `doComposePause / Unpause / Kill / Rm`(compose_action.go):共享 `doComposeProjectLifecycle(verb, perContainerFn)` helper,迭代项目容器,聚合为单条 BatchActioned。
- `doComposePrune`:在 lifecycle 流程基础上加卷清理(`m.Resources.Volumes.Items` 中带 project label 的)。
- `doComposeScale`(compose_action.go + container_form.go):
  - `state.FormComposeScale` 新增 FormKind。
  - `openComposeScaleForm`:Form 字段 = `IntField(target replicas)` + `BoolField(--no-deps)`,默认 0..100 范围。
  - `executeComposeScale`:target > current 走 `ComposeService.Run` 新增 replicas;target < current 按 container-number 升序删最早的。共一条 BatchActioned + 一条 audit trace。
  - `ResourceComposeService`(state)新增 ResourceType 常量。

**保留/已知缺口**:
- `--no-deps` 标志未生效(dtui 不解析 yaml,实际行为=全启)。

**已完成**(本轮):
- CoLocated group 拦截(Docker group scale = 1):`openComposeScaleForm` 在 runtime=Docker 且选中 service 的任一容器 `CoLocatedGroupID` 非空时,toast "✕ Docker compose 限制 network_mode: service: 的 service 不能 scale>1" 且不开 Form;Podman 正常开 Form(R08-14 算法落地后启用)。

**测试**:35 packages 全 PASS。
