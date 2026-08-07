# Runtime 驱动层重构：领域层与传输层分离

> 日期: 2026-08-05
> 状态: 计划阶段，待用户确认后实施
> 关联: `podman-adapter-relocate.md`, `podman-dto-unify.md`, `docker-adapter-purify.md`

## 1. 目标

驱动层（`podman/`、`docker/`）只保留纯传输逻辑（REST/CGO/SDK），所有映射、适配、错误处理归属领域层（`runtime/`）。依赖方向反转：`runtime/` → `podman/`（构造+映射），`podman/` 零上游依赖。

**关键设计**：
- mapper、service adapter、error mapping、filter 等非传输逻辑从 podman/ 迁移到 runtime/
- 驱动层成为纯粹的传输层
- ContainerInspectJSON 移入 dto/ 统一原生类型位置
- docker/ 从独立目录迁移到 runtime/docker/ 实现对称结构

## 2. 改动范围

### 2.1 必须完成

| 阶段 | 任务 |
|---|---|
| Wave 1 | ContainerInspectJSON 迁移到 dto/；简化 aliases.go |
| Wave 2 | 5 个 mapper 文件迁移到 runtime/ |
| Wave 3 | Service/Adapter/Client 迁移到 runtime/ |
| Wave 4 | Error/Filter/Progress/Event 迁移到 runtime/ |
| Wave 5 | docker/ 从 internal/data/docker/ 迁移到 internal/data/runtime/docker/ |
| Wave 6 | 测试迁移 + 编译验证 |

### 2.2 详细文件清单

#### Wave 1: 基础准备

| 文件 | 改动 |
|---|---|
| `podman/inspect_types.go` | 移动到 `podman/dto/inspect_types.go` |
| `podman/aliases.go` | 删除不再需要的类型别名 |

#### Wave 2: Mapper 迁移

| 来源 | 目标 |
|---|---|
| `podman/mapper_container.go` | `runtime/mapper_podman_container.go` |
| `podman/mapper_image.go` | `runtime/mapper_podman_image.go` |
| `podman/mapper_volume.go` | `runtime/mapper_podman_volume.go` |
| `podman/mapper_network.go` | `runtime/mapper_podman_network.go` |
| `podman/inspect_mapper.go` | `runtime/mapper_podman_inspect.go` |

#### Wave 3: Service/Adapter/Client 迁移

| 来源 | 目标 |
|---|---|
| `podman/service_adapters.go` | `runtime/service_adapters_podman.go` |
| `podman/services.go` | `runtime/service_podman.go` |
| `podman/client_methods.go` | `runtime/client_methods_podman.go` |

#### Wave 4: Error/Filter/Progress/Event 迁移

| 来源 | 目标 |
|---|---|
| `podman/error_mapping.go` | `runtime/error_mapping_podman.go` |
| `podman/filter_post.go` | `runtime/filter_post_podman.go` |
| `podman/progress.go` | `runtime/progress_podman.go` |
| `podman/event_util.go` | `runtime/event_util.go` |

#### Wave 5: Docker 层迁移

| 来源 | 目标 |
|---|---|
| `internal/data/docker/` | `internal/data/runtime/docker/` |

### 2.3 不得修改

- 不修改 dto/ 内部结构
- 不修改 runtime/ 的领域类型定义（ContainerSummary, ImageSummary 等）
- 不修改 UI 层代码（internal/tui/）
- 不修改 REST 客户端实现（rest.go, containers_rest.go 等）
- 不修改 Docker SDK 调用逻辑
- 不新增功能或 API

## 3. 实施顺序

1. **Wave 1**: InspectTypes 迁移、别名清理
2. **Wave 2**: 5 个 mapper 文件迁移
3. **Wave 3**: Service/Adapter/Client 迁移
4. **Wave 4**: Error/Filter/Progress/Event 迁移
5. **Wave 5**: Docker 层迁移
6. **Wave 6**: 测试迁移 + 编译验证

## 4. 验收标准

- [ ] `podman/` 不再导入 `runtime/` — 依赖方向反转
- [ ] `docker/` 位于 `internal/data/runtime/docker/` — 对称结构
- [ ] 所有 mapper、service adapter、error mapping、filter 在 `runtime/` 中
- [ ] `CGO_ENABLED=0 go build ./...` 通过
- [ ] `go vet ./...` 通过
- [ ] 所有测试通过

## 5. 相关提案

- `podman-adapter-relocate.md` — Podman 适配器重新定位
- `podman-dto-unify.md` — Podman DTO 统一
- `docker-adapter-purify.md` — Docker 适配器净化

## 6. 待确认问题

1. ContainerInspectJSON 移入 dto/ 的具体子类型拆分
2. engine.go 是否留在 podman/（用于构造 Engine）
3. exec.go 是否留在 podman/（传输层逻辑）
