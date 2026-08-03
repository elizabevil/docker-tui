# R01-01 容器高级动作 (Copy / Update / Diff / Export / Commit / Wait)

## 元信息

- 状态: implementing
- 优先级: high
- 来源: [../../task/br-033-task019-advanced-container-actions.md](../../task/br-033-task019-advanced-container-actions.md)、[../../task/br-033-035-review-fixes.md](../../task/br-033-035-review-fixes.md)
- 关联任务:
  - [../../task/br-033-advanced-container-ops-analysis.md](../../task/br-033-advanced-container-ops-analysis.md)
  - [../../task/br-033-small-model-implementation-prompt.md](../../task/br-033-small-model-implementation-prompt.md)
  - [../../task/br-033-035-review-fixes.md](../../task/br-033-035-review-fixes.md)
  - 运行时刷新缺陷已合并到 R01-03 + C01 + C05,原 BR-042 归档删除
- 关联约束:
  - [../../constraint/C01-form.md](../../constraint/C01-form.md)
  - [../../constraint/C05-path.md](../../constraint/C05-path.md)
  - [../../constraint/C06-i18n.md](../../constraint/C06-i18n.md)

## 目标

容器页提供 6 个 TASK-019 高级动作,允许用户从 UI 完成 Copy / Update / Diff / Export / Commit / Wait。

## 用户流程

容器页选中容器 → Action Bar(`;` 默认键) → 选动作:

- **Diff** / **Wait**: 直接执行,显示结果详情
- **Copy** / **Export**: 打开 Form 填源路径与目标 tar
- **Update**: 打开 Form 填资源参数
- **Commit**: 打开 Form 填镜像引用等参数

Submit 失败时 Form 保留并显示错误 toast。

## UI/UX

Action Bar 按 [R03-02](../R03-form-action/R03-02-action-bar.md) 分组显示,Form 遵循 [C01-form](../../constraint/C01-form.md) 与 [R03-03](../R03-form-action/R03-03-action-dialog-layout.md) 布局。

## 功能规则

- Copy / Export:本地目标必须存在父目录;已存在文件触发 Force 确认(详见 [C01-form](../../constraint/C01-form.md))
- Update:留空字段不改对应资源;restart-policy 选 `on-failure` 时 max-retries 生效
- Commit:repository 必填;tag 省略默认 `latest`
- Wait:可取消;过期响应丢弃
- Diff / Wait:失败提示走 toast + footer,成功进入详情页

## 实现设计

- 触发入口:`internal/tui/keyboard/actions.go:handleAction` 6 个 case
- 状态机:`internal/tui/state/form.go` `FormSpec` + 6 个 `FormKind`
- 表单提交流程:`internal/tui/keyboard/container_form.go` + `submitContainerForm`
- runtime 调用:`internal/data/runtime/containers.go` `ContainerService` 6 个方法已实现
- audit:`internal/tui/keyboard/audit.go` + `withAdvancedAudit`
- 涉及容器列表刷新与 detail 同步,详见 R01-03(Stats/Top/Wait 运行时刷新)

## 验收标准

- Action Bar 含 6 项,按容器状态过滤可用项
- 6 个 Cmd 工厂返回正确消息类型(`ContainerCopyDone` / `Update` / `Diff` / `Export` / `Commit` / `WaitDone`)
- 表单空字段校验正确(详见 [C01-form](../../constraint/C01-form.md))
- 覆盖目标走统一 Confirm + Force 弹层
- 双 runtime(Docker / Podman)6 个方法均通过测试

## 非目标

- 批量操作(留给批量动作 BR)
- 容器导入(不属于本需求)

## 迁移记录

- 旧文档:[task/br-033-task019-advanced-container-actions.md](../../task/br-033-task019-advanced-container-actions.md)、[task/br-033-035-review-fixes.md](../../task/br-033-035-review-fixes.md)、[task/br-033-advanced-container-ops-analysis.md](../../task/br-033-advanced-container-ops-analysis.md)
- 保留信息:6 个动作的代码索引、风险点、review 修复点
- 废弃信息:无
- 待确认状态:`implementing` 是因为 `Diff` / `Wait` 已完成,但 Form 流程在 BR-041/042 后续迭代中仍有未完成项;主模型确认后调整。
