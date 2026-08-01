# 镜像 tarball 导入功能 (Image Import)

> 任务卡 / BR-036

## 元信息

- **关联编号**:BR-036
- **优先级**:medium
- **状态**:`open`
- **依赖**:BR-018(镜像传输工作流),BR-039(Action Bar 入口)
- **关联设计**:[docs/feature-design.md §5.10](../feature-design.md)

## 目标

为镜像页增加 `Import tarball` 动作,从扁平 tar 文件(典型来自 `docker export`)创建单层新镜像。与现有 `Load` (Ctrl+L) 区分:`Load` 处理 `docker save` 的多层 tar,`Import` 处理单层扁平 tar。

## 代码结构索引

### 必须读懂的文件

| 文件 | 作用 |
|---|---|
| `internal/data/runtime/image.go:118` | `ImageService` 接口(只含 List / Inspect,**无 Import**) |
| `internal/data/runtime/docker/service_image.go` | Docker `imageService` 实现(只 List/Inspect) |
| `internal/data/runtime/podman/service_image.go` | Podman `imageService` 实现 |
| `internal/data/runtime/image_transfer.go` | 现有 `ImageTransferRequest` / `ImageTransferService` 框架(进度 / toast / 取消) |
| `internal/tui/keyboard/image_transfer.go` | 现有 transfer dialog / 进度处理(`openImageWorkflow` / `handleImageWorkflowInput` / `beginImageTransfer`) |
| `internal/tui/state/image_transfer.go`(若存在) | 现有 transfer 状态 |
| `internal/tui/keys/registry.go:78` | `ActionImageLoad ← Ctrl+L`(注意:Import 走 Action Bar 不直接绑键) |
| `internal/data/i18n/lang/*.jsonc` | 加 `image.import.title` / `image.import.path` / `image.import.repository` |

### 必须修改的文件

| 文件 | 改动 |
|---|---|
| `internal/data/runtime/image.go:118` | `ImageService` 接口加 `Import(ctx, source, ref, msg) (*ImageSummary, error)` |
| `internal/data/runtime/docker/service_image.go` | Docker `imageService.Import` 实现:`cli.ImageImport(ctx, source, ref, msg)` |
| `internal/data/runtime/podman/service_image.go` | Podman `imageService.Import` 实现:REST `images/import` |
| `internal/tui/keyboard/image_transfer.go` | 加 `openImageImport(m)` + 走现有 transfer 框架(进度 / 取消) |

### 必须新增的文件

- `internal/tui/ui/widget/dialog/image_import.go`:Import 表单(源 tarball 路径 + 目标 repo:tag + commit message)

## 操作链路

```text
User 在镜像页 → Action Bar (BR-039) 触发 ImageImport
  → keyboard/image_transfer.go:openImageImport(m)
      弹 widget/dialog/image_import.go 表单
      用户输入 source(本地 tar 路径或 URL) / ref(repo:tag) / message(可选)
      提交 → beginImageImport(m, req)
  → 构造 ImageTransferRequest{Operation: ImageTransferImport, Source: source, Destination: ref, Message: msg}
  → state.ImageTransfer.Begin(req, trace)
  → tea.Cmd: imageImportCmd(client, req) → Cmd
  → Engine.Images().Import(ctx, source, ref, msg)
      Docker: cli.ImageImport
      Podman: REST images/import
  → 返回 ImageSummary → 更新镜像列表(Refresh)
  → toast + audit("resource.image.import")
```

## 验收标准

- [ ] 镜像页 Action Bar 中"Import tarball"可点击。
- [ ] 触发表单,可填写源路径 / 目标 repo:tag / commit message。
- [ ] 提交后新镜像出现在镜像列表,inspect 可见,但只有 1 层(history 为空)。
- [ ] 与 Load (Ctrl+L) 路径独立不混淆:Load 处理多层 tar(保留 history),Import 处理扁平 tar(单层)。
- [ ] 进度反馈 / 取消机制复用现有 `ImageTransfer` 框架。
- [ ] Podman 端行为一致。
- [ ] 错误时显示明确 toast(路径不存在 / tar 文件损坏 / 引擎拒绝)。
- [ ] Audit 写入 `resource.image.import`。

## 风险

| 风险 | 说明 |
|---|---|
| **runtime API 兼容性** | Docker `cli.ImageImport` 接受 source 字符串(可远程 URL);Podman REST 接受 multipart upload,签名差异需谨慎 |
| **Load vs Import 误用** | UI 需明确区分:"多层 tar"vs"扁平 tar",Help 文案必须说清 |
| **路径注入** | 源路径来自用户输入,需要校验(避免 path traversal) |
| **大量镜像事件** | 导入后是否触发 Engine events?需要 Refresh 来更新列表 |

## 建议任务分解

1. **TASK-BR036-A:runtime 接口 + 实现**
   - `ImageService.Import` + Docker + Podman
   - **预估**:主模型实现
2. **TASK-BR036-B:TUI handler + dialog**
   - `openImageImport` + widget + 与 transfer 框架集成
   - **预估**:小模型辅助,主模型审 transfer 状态
3. **TASK-BR036-C:Action Bar 集成 (BR-039 协同)**
   - 在镜像 Action Bar 中加 "Import tarball"
   - **预估**:等 BR-039
4. **TASK-BR036-D:i18n + Help 同步**
   - 翻译文案 + 区分 Load / Import
   - **预估**:小模型机械同步

## 待确认项

- [ ] Import 是否支持 URL 源(如 `https://example.com/container.tar`)?
- [ ] Import 后 commit message 是否必填?
- [ ] 源路径是否允许 stdin(`-`)?
- [ ] 是否需要"导入后立即 inspect 详情"流程?