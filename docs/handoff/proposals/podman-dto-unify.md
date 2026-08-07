# Podman DTO 统一：消除 CGO 别名文件

> 日期: 2026-08-05
> 状态: 计划阶段，待用户确认后实施
> 关联: `internal/data/runtime/podman/` 包重构

## 1. 目标

统一 CGO 和 non-CGO 构建都使用 `dto.*` 类型作为 Podman wire DTO，消除 CGO alias 文件，修复 CGO 驱动中已存在的类型不匹配编译错误。

**关键设计**：
- `dto/` 包已经是 REST 路径的规范类型定义
- CGO 路径的 SDK 类型在 `DriverBackend` 边界转换为 `dto.*` 类型
- mapper 层保持 CGO 无关
- 符合设计文档 Phase 0 的决策："两种 transport 都先转换到 adapter 内部的 Podman wire DTO，再经过同一 mapper 转成统一领域模型"

## 2. 改动范围

### 2.1 必须完成

| 文件 | 改动 |
|---|---|
| `internal/data/runtime/podman/driver_iface.go` | 更新 `DriverBackend` 接口返回类型为 `dto.*` |
| `internal/data/runtime/podman/driver_cgo.go` | 所有函数将 SDK 类型转换为 `dto.*` 类型 |
| 删除 `aliases_cgo.go` | CGO 别名文件 |
| 删除 `aliases_container_cgo.go` | CGO 别名文件 |
| 删除 `aliases_network.go` | CGO 别名文件 |

### 2.2 不得修改

- 不修改 `dto/` 包中的任何类型定义
- 不修改 `runtime/` 包中的任何领域类型
- 不修改 mapper 函数 (`mapper_*.go`)
- 不修改 REST 客户端代码 (`rest.go`, `*_rest.go`)
- 不修改 service adapters (`service_adapters.go`, `services.go`)
- 不修改 `docker/` 包

## 3. 具体转换规则

### 3.1 ListContainers

SDK 返回 `[]types.ListContainer`，转换为 `[]dto.ContainerItem`：
- `Created`: SDK 是 `time.Time`，dto 也是 `time.Time`，直接赋值
- `Ports`: SDK 返回 `[]commonTypes.PortMapping`，dto 期望 `[]dto.ContainerPort`，逐字段转换

### 3.2 ListImages

SDK 返回 `[]types.ImageSummary`，转换为 `[]dto.ImageItem`：
- 字段：`ID`、`RepoTags`、`Created`、`Size`、`Labels`、`Arch`、`IsManifestList`

### 3.3 ListNetworks

SDK 返回 `[]commonTypes.Network`，转换为 `[]dto.Network`：
- 字段：`Name`、`ID`、`Driver`、`Scope`、`Created`、`Subnets`、`IPv6Enabled`、`Internal`、`Labels`、`Options`

### 3.4 PruneNetworks

修复编译错误：`dto.NetworkPruneReportItem` 的 `Error` 字段是 `string`，需要 `fmt.Sprintf("%v", report.Error)` 而非 `json.Marshal`

### 3.5 PruneVolumes

修复编译错误：
- `dto.VolumePruneReportItem` 的 `Size` 是 `int64`，需要 `int64(report.Size)`
- `Err` 是 `string`，需要 `fmt.Sprintf("%v", report.Err)`

## 4. 实施阶段

### Wave 1: 类型基础
更新 `driver_iface.go` 使 `DriverBackend` 接口使用 `dto.*` 类型

### Wave 2: CGO 驱动转换
更新 `driver_cgo.go` 所有函数转换 SDK→dto 类型

### Wave 3: 清理与验证
删除 CGO alias 文件，验证编译和测试

## 5. 验收标准

- [ ] `CGO_ENABLED=0 go build ./...` 通过
- [ ] `CGO_ENABLED=0 go test ./internal/data/runtime/podman/...` 通过
- [ ] `driver_iface.go` 中 `DriverBackend` 接口方法返回 `dto.*` 类型
- [ ] `driver_cgo.go` 中无直接返回 SDK 类型
- [ ] `aliases_cgo.go`、`aliases_container_cgo.go`、`aliases_network.go` 已删除

## 6. 待确认问题

1. CGO 路径的 Network/Volume Prune 报告字段类型转换方式（string vs json.RawMessage, int64 vs uint64）
2. 是否需要同期实施 Podman Signal 同步（与 Stop Form 字段相关）
