# R08-07 服务级联查询(top / port / stats)

## 元信息

- 状态: implementing
- 优先级: medium
- 来源: 用户评审结果 + docker compose CLI 权威
- 关联任务: [.omo/compose-todo.md 附录 A.2 / §3.2 / 补 G.4](../../../.omo/compose-todo.md)
- 关联约束:
  - [../../constraint/C03-table.md](../../constraint/C03-table.md)
  - [../../constraint/C04-keybinding.md](../../constraint/C04-keybinding.md)

## ⚠ 重复功能说明

| 能力 | 已实现层 | 本需求增量 |
|---|---|---|
| `docker compose top <svc>` | R01 容器 `top`(`containerTopCmd`) | R08-07 增量:按 service 聚合展示(每容器 1 子表),不需 Tab → 容器 |
| `docker compose port <svc> <port>` | R01 容器 `:port` 命令(`doContainerPort`) | R08-07 增量:compose-级 port 自动查 service 名 + 端口号 |
| `docker compose stats <svc>` | R01 容器 `stats`(`containerStatsCmd`) | R08-07 增量:按 service 聚合 / project 聚合 + 时间序列同步刷新 |
| `docker compose ps` (服务级 + 状态) | 已实现(view.go:renderServicePanel) | 不重复实现 |

## 目标

Compose-级查询类动作:`top / port / stats`。三个动作可由 R01 容器动作循环实现,但在 compose 层有**聚合价值**:

1. 用户在 compose 面板直接触发,不需要 Tab → 容器面板找到那个容器
2. 一次显示整个 service 的 top / stats,而不是单容器
3. project 级 stats 是聚合数据

## 用户流程

### top

1. Compose 面板 → 选中 service → 按 `Ctrl+T` (注册在 R08-12)
2. 弹出 Top 子视图(弹层),含该 service 所有容器的 process 表
3. `j` / `k` 上下选择容器,`Enter` 进入容器 Top(R01 已有)
4. `Esc` 返回 Compose 面板

### port

1. Compose 面板 → 选中 service → 按 `,`(注册)
2. 弹出 Port 子视图,展示 service 的 port mappings(自动聚合 replicas 的 port 列表)
3. 每行:`host:container:protocol`
4. `j` / `k` 翻行,`Esc` 返回

### stats

1. Compose 面板 → 选中 service 或项目 → 按 `m`(`Ctrl+M` 是否重绑由 R08-12 决定)
2. 弹出 Stats 子视图,展示聚合 stats:
   - 服务级:每 replica 一行(`cpu% / mem usage / mem limit / net io / block io`)
   - 项目级:每 service 一行 + 总和
3. 默认 follow(每 3s 刷新,沿用 R01 stats 周期)
4. `p` 暂停 / `r` 刷新 / `Esc` 返回

## UI/UX

仅本需求特有:

- Top 子视图:列表 + Tab 切换容器 + 紧凑布局
- Port 子视图:列表 + 表格列布局(`Host`、`Container`、`Protocol` 沿用 R01 端口列)
- Stats 子视图:表格列沿用 R01 stats(`CPU% / MEM / MEM% / NET I/O / BLOCK I/O`),列头来自 [constraint/C03-table](../../constraint/C03-table.md)

完整版式 UI overhaul。

## 功能规则

### F1 Service-level Top

```go
// ComposeService.Top(project, service) -> per-container ContainerProcesses
func (s *DockerComposeService) Top(ctx context.Context, project, service string) (map[string]runtime.ContainerProcesses, error)
```

实现:列出 `project=<p> AND service=<s>` 容器,循环 `ContainerTop`。

UI 渲染:每容器一个标题 + 子表,**缓存**每个容器的 top 结果,避免每次刷新抓全部。

### F2 Service-level Port

```go
func (s *DockerComposeService) Port(ctx context.Context, project, service string) ([]PortMapping, error)

type PortMapping struct {
    ContainerID string
    ContainerPort int
    Protocol    string
    HostIP      string
    HostPort    int
}
```

实现:解析 `ContainerInspect.NetworkSettings.Ports` 字段(已有,见 R01 容器 `:port` 命令实现),聚合 replicas。

### F3 Project-level Stats

```go
// stats 是 high-frequency 拉取,本方法返回 stream
func (s *DockerComposeService) Stats(ctx context.Context, project string) (<-chan ComposeStats, error)

type ComposeStats struct {
    Timestamp   time.Time
    PerService  map[string][]ContainerStat // service -> container stats
    PerProject  ContainerStat              // 累加
}
```

实现:

1. 列出项目容器,起 N 个 `ContainerStats(ctx, id)` API goroutine
2. 合并到 channel,每 T 间隔
3. 失败容器显示 `(lost connection)` 占位,不阻塞汇总

### F4 拉取频率 / 取消

- 默认周期:沿用 `config.Docker.StatsPollSec`(默认 3s)
- 取消:用户按 `Esc` 时退订 channel,R08-07 实现要确保 goroutine 退出

### F5 audit

- `top / port / stats` **不**触发 audit(读操作,沿用 R01 容器级"只读不审计"政策)
- exception:用户做了修改类动作(如 `compose port` 之后改了 compose port 暴露)— 不在本需求范围

## 实现设计

涉及模块:

- 新增方法:`ComposeService.Top / Port / Stats`(R08-02 文件追加)
- 新文件:`internal/tui/ui/pages/compose/top_subview.go`
- 新文件:`internal/tui/ui/pages/compose/port_subview.go`
- 新文件:`internal/tui/ui/pages/compose/stats_subview.go`(3 个子视图)
- 修改:`internal/tui/keyboard/compose_action.go` 加 `doComposeTop / doComposePort / doComposeStats`
- 修改:`internal/tui/keyboard/compose_nav.go` 注册快捷键(R08-12 决定)
- 复用 R01:`internal/tui/keyboard/container_action.go:ContainerTop / Port / Stats`

调用链:

```text
doComposeStats(project)
  -> ComposeService.Stats(ctx, project)
     -> 多 ContainerStats goroutine
        -> channel ComposeStats
           -> UI stats_subview 渲染
```

## 验收标准

1. Top 子视图:scale=3 的 service 展示 3 张子表,每个展示进程标题 + rows
2. Port 子视图:replicas 的 port 列表去重(`0.0.0.0:8080 -> 80/tcp` 一次)
3. Stats 子视图:每秒更新,Esc 终止
4. 失败容器显示 `(lost connection)` 占位行,不中断其它容器
5. audit 不增加条目

## 非目标

- project stats 跨 runtime(Docker + Podman 同主机)— 留 R04
- 历史 stats 回放(`docker stats` 无此语义)
- 警示阈值(`mem > 90%` 警告)— 留 future enhancement

## 迁移记录

- 旧代码:无(R08-07 全新)
- 复用:R01 容器 top / port / stats 的 runtime 调用
- 保留信息:R01 stats 的轮询 / 取消机制
- 废弃信息:无
- 待确认:`m` 键是否复用为 stats(可能与 R01 容器 stats 冲突 — 由 R08-12 决策)
