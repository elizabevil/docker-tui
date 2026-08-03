# R02-02 镜像 tarball 导入 (Image Import)

## 元信息

- 状态: planned
- 优先级: medium
- 来源: 原 BR-036 已合并归档
- 关联任务: 无
- 关联约束:
  - [../../constraint/C01-form.md](../../constraint/C01-form.md)
  - [../../constraint/C05-path.md](../../constraint/C05-path.md)
  - [../../constraint/C06-i18n.md](../../constraint/C06-i18n.md)

## 目标

镜像页提供 Import tarball 动作,从扁平 tar 文件(典型来自 `docker export`)创建单层新镜像。

与现有 Load(Ctrl+L,处理 `docker save` 多层 tar)区分。

## 用户流程

镜像页 → Action Bar → Image Import → 打开 Form(源 tarball 路径 + 目标 repo:tag + commit message)→ Submit → 走 transfer 框架(进度 / 取消)→ 完成 / 失败 toast。

## UI/UX

- Form 按 [R03-01](../R03-form-action/R03-01-form-pattern.md) 模板
- 路径字段遵循 [C05-path](../../constraint/C05-path.md) — Local tarball
- 进度条走 transfer 框架(类似 Save / Load)

## 功能规则

- 源 tarball 必须是扁平单层 tar(`docker export` 输出)
- `repository:tag` 必填
- `message` 可选,默认空
- 大文件传输期间可取消
- 失败时 toast 显示 docker / podman error

## 实现设计

- 接口:`internal/data/runtime/image.go:ImageService.Import(ctx, source, ref, msg)`
- Docker:调 `cli.ImageImport(ctx, source, ref, msg)`
- Podman:调 REST `images/import`
- 入口:`internal/tui/keyboard/image_transfer.go:openImageImport(m)`
- 复用 transfer 框架:`ImageTransferRequest` / 进度 / 取消
- 表单:`internal/tui/ui/widget/dialog/image_import.go`

## 验收标准

- Action Bar 镜像页可见 Image Import 项
- 源路径通过 Tab / 路径补全
- 提交后进度显示,失败可取消
- 双 runtime 端到端测试通过

## 非目标

- 多 tar 文件批量导入
- 远程 URL 导入

## 迁移记录

- 旧文档:原 BR-036 已合并归档,内容汇总在本文档
- 保留信息:ImageService.Import 接口设计、复用 transfer 框架
- 待确认状态:`planned`,原 task 未实施;主模型确认后启动。