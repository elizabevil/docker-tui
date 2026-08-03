# R01-03 Stats / Top / Wait 运行时刷新体验

## 元信息

- 状态: planned-review
- 优先级: medium
- 来源: 原 BR-042 中运行时刷新缺陷条目已合并入本文档 + C01 + C05;BR-042 归档删除
- 关联任务: 无
- 关联约束:
  - [../../constraint/C03-table.md](../../constraint/C03-table.md)
  - [../../constraint/C04-keybinding.md](../../constraint/C04-keybinding.md)

## 目标

Stats / Top / Wait 三个运行时刷新类动作,提供统一的取消、过期响应丢弃、面板退出清理等体验。

Stats 当前是容器列表的列(可由开关切换显示),不是独立 Stats 页面。
Top 是独立页面(`ModeTop`)。
Wait 是 Action Bar 上的独立动作,带可取消的 context。

## 用户流程

- Stats:容器页按 `m` 键或 Action Bar 切换 stats 列的显示
- Top:Action Bar → 进入 Top 顶层页(实时刷新,支持排序)
- Wait:Action Bar → Wait 动作 → 进入等待状态,容器达成条件自动退出
- 任一模式下用户按 `Esc` / `Ctrl+C` 取消并清理后台订阅

## UI/UX

### Stats

- 容器列表的列;开关状态在 `state.Resources.Containers.StatsEnabled` 或等价字段
- 列内容:CPU% / Memory usage / Net I/O

### Top

- 复用 [C03-table](../../constraint/C03-table.md) 表格布局
- 顶部状态栏显示 `[refreshing]`
- 排序快捷键 `O` / `Ctrl+O`

### Wait

- 弹层显示等待状态与已等待时长
- 达成条件自动退出,失败 / 用户取消走 toast

## 功能规则

### 通用

- 后台订阅必须可取消(`context.Context`)
- 过期响应丢弃(`generation` token)
- Mode 退出 / runtime 切换时清理后台 goroutine

### Stats

- 定时轮询当前容器列表(不是只轮询单个聚焦容器)
- 关闭开关后停止轮询

### Top

- 仅在 running 容器可用
- 已退出 / 已停止容器给空表 + 提示

### Wait

- 容器从未运行 / 已停止状态给明确提示
- 等待超时给可配置上限

## 实现设计

### Stats

- 列定义在 `internal/tui/ui/component/container_table.go` 或类似
- 轮询任务由 `state.Containers.StatsTimer` 调度
- 取消与清理:`state.Containers.StatsTimer.Stop()`

### Top

- 状态:`state.Top`(`AppModel` 顶级字段或 `Resources.Containers.Top`)
- Mode:`state.ModeTop`
- View:`internal/tui/ui/pages/top/view.go`
- runtime 调用:`engine.Containers().Top(ctx, id)`(对应 `ContainerService.Top`)

### Wait

- 由 `state.AppModel.ContainerWait`(`state.ContainerWaitState`) 承载 generation 与 cancel context,完成消息为 `keyboard.ContainerWaitDone`;**不参与** `state.ExecState`(ExecState 是 Exec shell 的状态)
- Cmd:`containerWaitCmd` 已有(详见 [R01-01](./R01-01-advanced-ops.md) 的 `Wait` 动作)
- cancellation:Context cancel propagation

## 验收标准

- Stats 列开关切换后 1 秒内开始 / 停止轮询
- Top 独立页可正常进入退出
- Wait 可取消,过期响应被丢弃
- 快速切换 runtime 时旧 runtime 响应不污染新视图

## 非目标

- 多容器并发 Stats / Top
- 导出 / 持久化刷新历史

## 迁移记录

- 旧文档:BR-042 运行时刷新缺陷条目已合并入本文档 + C01 + C05,BR-042 归档删除
- 保留信息:运行时刷新缺陷列表与修复方向
- 废弃信息:旧文档混入了 ExecState 等无关内容,本需求仅与 Stats / Top / Wait 相关
- 待确认状态:`planned-review`,原 task 把这些条目列入"修正"清单,但未独立标记为新需求;主模型确认 R01-03 边界后调整。
