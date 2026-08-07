# Podman 适配器重新定位

> 日期: 2026-08-05
> 状态: ✅ 已完成

## 1. 目标

将分散在 `runtime/` 级别的 12 个 Podman 适配器文件移动到 `runtime/podman/` 包中，形成与 `runtime/docker/` 对称的结构。修复 `docker/` 适配器对旧的 `runtime/podman`（现在是 `internal/driver/podman`）的过时导入。

**结果**：三层架构 — 接口 → 适配器 → 驱动

## 2. 实施内容

### 2.1 文件移动

| 来源 | 目标 |
|---|---|
| `runtime/podman_*.go` (12 个文件) | `runtime/podman/` |
| `runtime/*_podman_test.go` (3 个测试文件) | `runtime/podman/` |

### 2.2 包声明与导入

- 所有移动文件包声明从 `runtime` 改为 `podman`
- 添加 `runtimeapi "...internal/data/runtime"` 导入

### 2.3 Docker 适配器修复

| 文件 | 修复内容 |
|---|---|
| `runtime/docker/client.go` | 导入：`internal/data/runtime/podman` → `internal/driver/podman` |
| `runtime/docker/engine_factory.go` | `runtimeapi.NewPodmanEngine(...)` → `podmanadapter.NewEngine(...)` |
| `runtime/docker/engine_factory_test.go` | 类型断言更新为新包 |

### 2.4 清理

- 删除旧的 `runtime/podman/` 目录（空的占位符）
- 删除旧的 `*_podman.go` 文件

## 3. 验收标准

- [x] `runtime/podman/` 包实现 `runtime.Engine` 和所有服务接口
- [x] `runtime/` 仅包含接口定义和共享类型（无 `*podman.go` 文件）
- [x] `runtime/docker/` 正确导入 `internal/driver/podman` 和 `runtime/podman`
- [x] `internal/driver/podman/` 未修改
- [x] 全量编译通过：`CGO_ENABLED=0 go build ./...`

## 4. 相关提案

- 本提案是 `runtime-driver-refactor.md` 的一部分
- 与 `podman-dto-unify.md` 配合使用（统一 DTO 类型）
