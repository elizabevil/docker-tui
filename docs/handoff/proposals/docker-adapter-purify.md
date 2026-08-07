# Docker 适配器净化：提取共享基础设施

> 日期: 2026-08-05
> 状态: 计划阶段，待用户确认后实施
> 关联: `runtime/docker/` 包重构

## 1. 目标

`runtime/docker/` 成为纯 Docker 适配器。共享连接管理（Pool、Config、Error classification）和路由逻辑迁移到 `runtime/` 级别。Podman 特定辅助函数迁移到 `runtime/podman/`。向后兼容别名保持 35 个 TUI 文件正常工作。

**关键设计**：
- 当前 `runtime/docker/` 包混合了 Docker SDK 客户端代码与 Podman socket 检测、REST 客户端创建和运行时路由
- 这违反了预期的三层架构：`runtime/`（接口 + 连接管理）→ `runtime/docker/` 和 `runtime/podman/`（纯适配器）
- 将共享基础设施迁移到 `runtime/`，Podman 辅助函数迁移到 `runtime/podman/`，实现清晰分离
- 别名避免 35 个文件导入重写

## 2. 改动范围

### 2.1 必须完成

| 新文件 | 来源 | 说明 |
|---|---|---|
| `runtime/connection.go` | 从 `runtime/docker/` 移动 | `ConnectionSpec`, `TLSConfig`, `RuntimeType` |
| `runtime/pool.go` | 从 `runtime/docker/pool.go` 移动 | `ConnectionPool`, `PoolEntry`, `ConnState`, `NewPool()` |
| `runtime/connection_error.go` | 从 `runtime/docker/` 移动 | `ConnectionFailure`, `ClassifyConnectionError` |
| `runtime/engine_factory.go` | 从 `runtime/docker/engine_factory.go` 移动 | 路由逻辑 |
| `runtime/podman/discovery.go` | 从 `runtime/docker/` 移动 | Podman socket 发现 |

### 2.2 净化现有文件

| 文件 | 改动 |
|---|---|
| `runtime/docker/client.go` | 移除 `podmanREST` 字段，移除 `RuntimeType` 字段，简化为 Docker only |
| `runtime/docker/runtime_connection.go` | 移除 `PodmanUserEndpoint`，更新 `FromRuntimeConn` 使用 `runtime.ConnectionSpec` |
| `runtime/docker/types.go` | 保留向后兼容别名 |

### 2.3 不得修改

- 不修改 `runtime/podman/engine.go`、`services.go`、`mappers.go`、`helpers.go`（已清理）
- 不改变 `runtime.Engine` 接口或任何服务接口
- 不改变 TUI 业务逻辑（仅在别名不覆盖的地方修改导入）
- 不重写所有 35 个 TUI 文件导入（别名处理向后兼容）
- 不新增运行时检测逻辑 — 仅迁移现有代码

## 3. 实施阶段

### Wave 1: 创建共享基础设施
创建 `runtime/connection.go`、`runtime/pool.go`、`runtime/connection_error.go`、`runtime/engine_factory.go`

### Wave 2: 移动 Podman 辅助函数
将 Podman socket 发现移动到 `runtime/podman/discovery.go`

### Wave 3: 净化 Docker 适配器
更新 `runtime/docker/client.go`、`runtime/docker/runtime_connection.go`

### Wave 4: 验证
全量编译和测试通过

## 4. 验收标准

- [ ] `CGO_ENABLED=0 go build ./...` 通过
- [ ] `go test ./internal/data/runtime/...` 通过
- [ ] `go vet ./internal/data/runtime/...` 通过
- [ ] 所有 35 个 TUI 文件无需修改即可编译

## 5. 待确认问题

1. 向后兼容别名是否足够（某些可能更偏好重写所有 35 个 TUI 文件导入）
