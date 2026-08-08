# R08-06 聚合日志

## 元信息

- 状态: planned
- 优先级: high
- 来源: 用户评审结果 + [.omo/compose-todo.md 补 E.5 / 补 G.3](../../../.omo/compose-todo.md)
- 关联任务: [.omo/compose-todo.md §C.1 / §3.2 `logs -f` / 补 E.5 `doComposeLogs` 单容器问题](../../../.omo/compose-todo.md)
- 关联约束:
  - [../../constraint/C04-keybinding.md](../../constraint/C04-keybinding.md)
  - 历史 R05 事件流的 logs 复用

## ⚠ 重复功能说明

| 能力 | 已实现层 | 本需求增量 |
|---|---|---|
| `docker compose logs`(全项目) | R01 单容器 logs(`m.Log.Open(id)`) | **R08-06 增量**:聚合日志,按 service / time 排序 + service 前缀 |
| `docker compose logs <service>` | R01 容器 logs(选中特定容器) | **R08-06 增量**:按 service 名聚合,即使 scale=3 也合并到一行时间轴 |
| `docker compose logs -f` (follow) | R01 容器 logs(可 follow) | **R08-06 复用 follow 机制**,聚合 follow 多个 stream |

## 目标

当前 `doComposeLogs`(`internal/tui/keyboard/compose_action.go:241-269`)只打开 `containers[0]` 的 logs,看不到其它 service 的输出,scale>1 也只看到 1 个。R08-06 改为聚合所有匹配容器的日志流,带 service 前缀与可选 follow。

三个具体能力:

1. **项目级 logs**:打开项目时拉所有容器日志,合并输出
2. **服务级 logs**:打开 service 时拉该 service 全部 replicas 聚合
3. **`-f` follow**:复用 R01 follow 机制(订阅 stream + tick 拉取),额外实现"多 stream 复用一行时间轴"

## 用户流程

1. Compose 面板左栏/右栏选中目标(project / service)
2. 按 `l` / Action Bar(Logs)
3. 弹屏 / Logs 视图:
   - 标题:`Compose Logs: <project>` 或 `Compose Logs: <project>/<service>`
   - 内容:每行前缀 `[service-name] ` 或 `[service-name.1] [service-name.2]`(replicas 时)
   - 默认 `-f true` 开启 follow
4. `Esc` 返回 Compose 面板

## UI/UX

仅限本需求特有:

- 时间戳排序:多 stream 拉到时按时间戳合并(monotonic CLOCK_REALTIME 不可靠,用 line sequence + 内部时间戳)
- 高亮不同 service 用 6 色 token(light theme 区分)
- 配色与日志系统已有 colors 一致(沿用 R01 容器 logs 的 color mapping)

完整版式 UI overhaul 阶段。

## 功能规则

### F1 聚合策略

```text
1. 列出该范围(project 或 service)的所有容器
2. 对每个容器,起一个 fetch 流,接收 lines
3. 在内存按时间戳排序输出
4. follow:true 时,新 line 到达即写入
```

### F2 内存 vs 文件

R01 当前做法:fetch log batch → 放进 ring buffer → 显示。R08-06 沿用同一 buffer,但 buffer 持有的是行数组 + per-line service tag。

```go
type composeLogLine struct {
    Service string
    Replica int // 0 = single
    ContainerID string
    Timestamp time.Time
    Line     string
}
```

### F3 scale 合并

同 service 多 replicas:

- 用户选项:`separate`(每 replica 一行)/ `merged`(同名 service 合并 + 时间戳)
- 默认 `separate`(避免幻觉)
- merged 时:R08-06 UI 在 header 显示 `[merged]`

### F4 service 前缀格式

```
[14:23:18.123] [web.1] GET /api 200
[14:23:18.231] [db.1] INSERT 12 rows
[14:23:18.456] [web.2] GET /api 200
```

(实现在 buffer 行格式 + 渲染层)

### F5 follow 终止

按 Esc:停止所有子 stream + 关闭当前 logs view。

若是 R05 事件流派生的 follow(留 R08-10 评估是否走 docker events 流 vs socket stream)。

### F6 partial fetch 失败

单个容器的 fetch 失败容错:

- 显示 `(web.1: connection lost, retries: 2)`,其它容器继续
- toast 报告:`⚠ web.1 stream lost`

## 实现设计

涉及模块:

- 修改:`internal/tui/keyboard/compose_action.go:241 doComposeLogs`,改为聚合调用
- 修改:`internal/tui/state/log.go`(`m.Log`)添加 `MultiSourceFollow` 字段
- 修改:`internal/tui/keyboard/log.go`(`LogBuffer`),扩展支持多 source 标注
- 修改:`internal/tui/ui/pages/logs/`(`LogView`)支持 multi-service 前缀
- 多 stream 调度:`internal/tui/keyboard/compose_action.go` 新增 `composeLogsCmd(project, services []string, opts)`

调用链:

```text
doComposeLogs(project, service "")
  -> composeLogsCmd(project, services=ALL)
     -> 多 containerFetchLogs + 内部 merge goroutine
        -> m.Log.OpenMulti() -> UI 渲染多 stream buffer
doComposeLogs(project, service "web")
  -> composeLogsCmd(project, services=[web])
```

## 验收标准

1. 选中项目后按 l → logs 中能看到该项目所有容器输出,每行带 service 前缀
2. 选中具体 service → 仅该 service 的容器输出
3. scale>1 的 service,默认 separate(每 replica 一行),用户选项 merged 走 footer 显示
4. follow 开启时,新行到达即写入,关闭 logs view 时 stream 全部终止
5. 单个 fetch 失败不阻塞其它容器,toast 报告失败容器
6. audit:`compose_service.logs_view` trace 在打开时触发一次,关闭不触发

## 非目标

- 跨项目聚合(超出项目概念)
- 持久化日志(`/var/lib/docker/containers/.../*.log` 文件直读)— dtui 走 socket API
- 日志全文搜索(含 nav 键 `n` / `Ctrl+N`)移交给 R01 容器 logs 的搜索机制,R08-06 复用

## 迁移记录

- 旧代码:`internal/tui/keyboard/compose_action.go:241 doComposeLogs`,当前只 fetch 1 个容器
- 保留信息:R01 容器 logs 的 ring buffer / follow / search
- 废弃信息:无(doComposeLogs 替换为聚合实现)
- 待确认:merge 在内存还是落盘(内存足够 — 单 buffer 限制 N 行,与 R01 一致)
