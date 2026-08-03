# R02 镜像(Image)

> 范围: 镜像资源的浏览、详情、导入、Registry 认证、传输

## 子需求

| 编号 | 标题 | 状态 | 优先级 | 来源 |
|---|---|---|---|---|
| [R02-01](./R02-01-history.md) | 镜像 History 顶层页 (H 键 / Action Bar) | implementing | high | [task/br-034-image-history-top-page.md](../../task/br-034-image-history-top-page.md) |
| [R02-02](./R02-02-import-tarball.md) | 镜像 tarball 导入 (Image Import) | planned | medium | 原 BR-036 已归档 |
| [R02-03](./R02-03-registry-login.md) | docker / podman Registry Login | planned | medium | 原 BR-037 已归档 |

## 关联约束

- [constraint/C01-form.md](../../constraint/C01-form.md) — Save / Load / Import 表单字段
- [constraint/C05-path.md](../../constraint/C05-path.md) — Image Save/Load 本地路径语义
- [constraint/C03-table.md](../../constraint/C03-table.md) — History 表列布局
- [constraint/C06-i18n.md](../../constraint/C06-i18n.md) — 文案

## 关联任务卡

- [task/br-034-image-history-top-page.md](../../task/br-034-image-history-top-page.md) — History 顶层页
- [task/br-034-runtime-analysis.md](../../task/br-034-runtime-analysis.md) — Runtime 分析
- [task/br-034-ui-state-analysis.md](../../task/br-034-ui-state-analysis.md) — UI/状态分析
- [task/br-034-implementation-completed.md](../../task/br-034-implementation-completed.md) — 实施完成记录
- BR-036 / BR-037 已合并入 R02-02 / R02-03,原 task 文件归档删除

## 待确认

- R02-02 Image Import 与现有 Image Load(Ctrl+L)的区分在用户层是否清晰,需主模型 review。
- Registry Login 是否纳入 R02-01 的 Image 多表单或独立 R02-04。