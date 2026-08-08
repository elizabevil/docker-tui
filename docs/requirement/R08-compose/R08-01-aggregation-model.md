# R08-01 容器标签聚合模型(Docker / Podman 分轨适配)

## 元信息

- 状态: implementing
- 优先级: high
- 来源: 用户评审结果 2026-08-07 + [.omo/compose-requirements-discussion.md §2.4 / §4.1](../../../.omo/compose-requirements-discussion.md)
- 关联任务: [.omo/compose-todo.md 附录 E.5](../../../.omo/compose-todo.md)
- 关联约束:
  - [../../constraint/C04-keybinding.md](../../constraint/C04-keybinding.md)
  - 与 R08-03(发现)耦合

## 目标

dtui 的 Compose 视图依赖容器 label 聚合项目 / 服务。现状是 Docker 与 Podman 适配器各写一份 label 读取,**形态不一致**(`if ok { ... = project }` vs 直接赋值 `summary.ComposeProject = c.Labels[...]`),且只覆盖 `com.docker.compose.project` / `com.docker.compose.service` 两个键,其它 `com.docker.compose.*` 官方 label 全部丢弃。

需要:

1. 抽象出"Compose-标签集"概念,包含**完整 Docker Compose label**(`com.docker.compose.*`)+ **Podman 兼容 label**(`io.podman.compose.*`)的并集。
2. Docker 适配器与 Podman 适配器**独立**实现标签提取,不允许共享工具函数。
3. 聚合器(Compose UI 层)消费统一 `ComposeLabels` 结构,不做运行时特化判断。

## 用户流程

无直接用户流程(数据层需求)。UI 行为见 R08-03 / R08-12。

## UI/UX

不涉及(后续 R08-12 / UI overhaul 时再决定 label 信息如何展示在右栏头部)。

## 功能规则

### F1 标签集定义(Docker 侧)

```go
// docker 包内独立定义,不允许导出到 runtime 包外
type composeLabels struct {
    Project         string   // com.docker.compose.project
    Service         string   // com.docker.compose.service
    ConfigHash      string   // com.docker.compose.config-hash
    ContainerNumber string   // com.docker.compose.container-number
    Oneoff          bool     // com.docker.compose.oneoff
    Version         string   // com.docker.compose.version
    WorkingDir      string   // com.docker.compose.project.working_dir
    ConfigFiles     []string // com.docker.compose.project.config_files (解析为多行)
}
```

### F2 标签集定义(Podman 侧)

Podman Compose 同时回写两组 label,故 Podman 适配器需要**先**查 `io.podman.compose.*`,再回退 `com.docker.compose.*`(后者实际上是用 docker-compose / podman-compose 时也存在)。两组等价。

```go
// podman 包内独立定义,不允许导出
type podmanComposeLabels struct {
    Project     string // io.podman.compose.project 或 com.docker.compose.project
    Service     string // io.podman.compose.service 或 com.docker.compose.service
    ProjectBy   string // 哪个键命中("io.podman.compose.*" / "com.docker.compose.*")
    ServiceBy   string
}
```

### F3 共享聚合接口

```go
// 在 runtime 包导出,runtime-agnostic
type ComposeLabels struct {
    Project         string
    Service         string
    WorkingDir      string   // 来自 Docker label;Podman 留空(无此 label)
    ConfigFiles     []string // 来自 Docker label;Podman 留空
    Version         string   // 来自 Docker label
    ConfigHash      string
    ContainerNumber string
    Oneoff          bool
    Source          string   // "docker" | "podman" | "unknown"
}

// runtime 包共享的 label 常量与填充函数(实际实现,2026-08-08 核对):
//   internal/data/runtime/compose_labels.go
//   ComposeLabelProject / ComposeLabelService / ComposeLabelImage /
//   ComposeLabelContainerNumber / ComposeLabelConfigHash / ComposeLabelOneoff /
//   ComposeLabelVersion / ComposeLabelWorkingDir / ComposeLabelConfigFiles
//   func ApplyComposeLabels(labels map[string]string, summary *ContainerSummary)
//
// 各 adapter 独立读取容器 label 后调用 ApplyComposeLabels 填充字段:
//   docker: internal/data/runtime/docker/containers.go(独立适配)
//   podman: internal/data/runtime/podman/mappers.go(独立适配)
```

聚合器在 `internal/tui/ui/pages/compose/view.go:36 gatherComposeProjects` 处调用聚合结果,不做运行时特化分支。

### F4 边界

- Project 为空 + Service 不为空 → 视为 `Unknown Service`(原代码已这样做)
- Project 不为空 + Service 为空 → `Unknown Service` 子组
- WorkingDir / ConfigFiles / Version 在 Podman 侧**留空**(Podman Compose 不设置这些),UI 需要降级显示

## 实现设计

涉及模块:

- 新文件:`internal/data/runtime/compose_labels.go` — 共享 label 常量 + `ApplyComposeLabels`(已落地)
- 修改:`internal/data/runtime/docker/containers.go` 调用 `ApplyComposeLabels` 并写入新增字段(已落地)
- 修改:`internal/data/runtime/podman/mappers.go` 同上(已落地)
- 修改:`internal/data/runtime/containers.go` `ContainerSummary` 增加 `WorkingDir / ConfigFiles / Version / ConfigHash / ContainerNumber / Oneoff / ComposeLabelSource` 字段(已落地)
- 修改:`internal/tui/keyboard/compose_action.go:30 / :43`(`composeProjectVolumes` / `composeProjectNetworks`)按 `ComposeLabels.Project` 字段比较(从 `Labels` map 改字段访问)

### R08-14 CoLocatedGroup 字段注入

`ContainerSummary` 同步增 2 个字段(详见 [R08-14](./R08-14-co-located-group.md) §F1 / §I1):

```go
type ContainerSummary struct {
    // ... existing
    CoLocatedGroupID  string  // R08-14
    CoLocatedGroupSrc string  // R08-14: "podman_default_pod" | "docker_network_mode_service" | ...
}
```

填充路径:

- Podman adapter 在 list 阶段读 `c.Pod`(podman REST 直接包含)→ `CoLocatedGroupID = "pod_<project>"`,`Src = podman_default_pod`
- Docker adapter 在 list 阶段不识别(需 inspect),仅在 drill-in / detail 时通过 `identifyDockerCoLocated()` 设置这两字段

调用链:

```text
runtime.ContainerList (docker) | libpod.ContainersList (podman)
  -> Adapter.*ContainerList
     -> Adapter.extractComposeLabels (per-container)
        -> ComposeLabels struct (runtime)
           -> runtime.ContainerSummary (新增字段)
              -> gatherComposeProjects(view.go)
```

## 验收标准

1. Docker 与 Podman 各自 adapter 实现 `Extract`,**互不调用**
2. Podman 路径:优先 `io.podman.compose.*`,回退 `com.docker.compose.*`
3. Docker 路径:读取所有列出的 `com.docker.compose.*` label,WorkingDir / ConfigFiles 多值按 `\\n` 解析为 slice
4. ContainerSummary 增加字段不破坏现有 JSON 序列化(`omitempty` + 兼容空值)
5. 单元测试覆盖两组 adapter 的"project / service 命中"主路径与回退路径
6. 双 runtime container JSON testdata 中显式存在 `com.docker.compose.project.working_dir` 等字段,需新增对应 fixture

## 非目标

- 不解析 `compose.yaml` 内容(R08-03 处理)
- 不读 Podman 侧不存在但 docker-compose 写入的额外字段(语义对齐即可,无需可逆)
- 不处理 label value 中的 shell 转义、超长字段截断(R08-03 / R08-13 关注)

## 迁移记录

- 旧文档:[.omo/compose-requirements-discussion.md §2.4 / §4.1](../../../.omo/compose-requirements-discussion.md)
- 旧代码:`internal/data/runtime/docker/containers.go:84-89`、`internal/data/runtime/podman/mappers.go:81-82`、`internal/tui/keyboard/compose_action.go:30 / :43`
- 保留信息:`com.docker.compose.project` / `com.docker.compose.service` 主键语义
- 废弃信息:无
- 待确认:**已定(2026-08-08)**。实际实现未引入 `ComposeLabelExtractor` 接口 —— label 常量与 `ApplyComposeLabels` 共享于 `runtime` 包,docker / podman adapter 各自独立读取容器 label 后调用填充函数(遵守"独立适配"原则:读取路径独立,填充逻辑共享)。
