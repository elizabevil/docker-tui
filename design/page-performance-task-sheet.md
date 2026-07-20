# dtui 页面与性能优化任务单

> 建立日期: 2026-07-20
> 作用: 将执行计划继续细化为可直接推进的任务单与回归清单。

本文不替代:

- [page-performance-execution-plan.md](page-performance-execution-plan.md): 阶段计划
- [page-performance-approval-proposals.md](page-performance-approval-proposals.md): 审批结论
- [bugfix-design.md](bugfix-design.md): BR 正式设计

本文只负责:

- 细化阶段内任务
- 提供回归清单
- 提供每项任务完成定义

## 使用规则

- 每条任务固定归属一个执行阶段 `EP-x`。
- 每条任务必须写清输入、修改范围、完成定义、回归点。
- 优先按批次推进，不建议跨批次并行大范围修改。

## 任务状态

| 状态 | 含义 |
|---|---|
| `todo` | 尚未开始 |
| `doing` | 正在实现 |
| `verifying` | 已实现，正在回归 |
| `done` | 已完成 |
| `blocked` | 有前置依赖或外部阻塞 |

## 批次总览

| 批次 | 包含阶段 | 目标 | 状态 |
|---|---|---|---|
| `Batch-A` | `EP-0` `EP-1` `EP-2` `EP-3` | 固定主骨架与基础轨道 | `verifying` |
| `Batch-B` | `EP-4` `EP-5` `EP-6` | 建立页面模板与统一投影 | `todo` |
| `Batch-C` | `EP-7` `EP-8` `EP-9` | 详情页结构化与性能收敛 | `todo` |

## EP-0 回归清单

### 页面回归矩阵

| 编号 | 页面 | 入口动作 | 关注点 |
|---|---|---|---|
| `RG-001` | containers list | 默认启动 | 表格高度、footer、标题摘要 |
| `RG-002` | images list | `Tab` 或命令跳转 | 表格高度、子视图入口、footer |
| `RG-003` | compose split view | `Tab` 或命令跳转 | 双栏宽度、焦点切换、footer |
| `RG-004` | logs view | 容器页 `l` | query rail、wrap、search、scroll |
| `RG-005` | detail view | 镜像或资源 `d`/`Enter` | detail body 高度、section 滚动 |
| `RG-006` | help view | `?` | help 布局、footer、退出 |

### 模式回归矩阵

| 编号 | 模式 | 进入方式 | 关注点 |
|---|---|---|---|
| `RM-001` | filter | `/` on list page | query rail、退出行为、panel 高度 |
| `RM-002` | search | `/` on logs page | query rail、search 状态、panel 高度 |
| `RM-003` | command | `:` | query rail、补全、退出 |
| `RM-004` | mark | `Space` on list page | footer 第二行、操作提示 |
| `RM-005` | detail | 进入 detail page | footer 固定性、body 滚动 |
| `RM-006` | log view | 进入 logs page | footer 固定性、viewport |

### 小终端回归边界

| 编号 | 尺寸 | 目标 |
|---|---|---|
| `RS-001` | `>= 120x32` | 标准完整布局 |
| `RS-002` | `>= 100x24` | 主流降宽场景 |
| `RS-003` | `>= 80x20` | 紧凑可用场景 |
| `RS-004` | `< 80x20` | 明确降级或限制提示 |

## Batch-A 任务单

### T-001 建立回归检查模板

- 状态: `done`
- 归属阶段: `EP-0`
- 目标:
  建立后续每阶段共用的回归记录模板。
- 输入:
  - [page-performance-execution-plan.md](page-performance-execution-plan.md)
  - 本文 `RG-* / RM-* / RS-*`
- 修改范围:
  - 文档
- 主要工作:
  1. 固定回归记录格式
  2. 固定页面、模式、小终端三类检查项
  3. 固定“通过 / 失败 / 备注”记录方式
- 完成定义:
  1. 后续每阶段都能复用同一模板
  2. 不需要每次重新发明回归表
- 回归点:
  - `RG-001 ~ RG-006`
  - `RM-001 ~ RM-006`
  - `RS-001 ~ RS-004`

### T-002 梳理布局现状与轨道映射

- 状态: `done`
- 归属阶段: `EP-0`
- 目标:
  把现有 `RenderApp()` 结构明确映射到五轨道骨架。
- 输入:
  - [internal/tui/ui/app/layout.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/layout.go)
- 修改范围:
  - 文档
  - 后续可扩展为代码注释
- 主要工作:
  1. 标出现有 header / toast / search / panel / footer 实际拼接顺序
  2. 标出未来五轨道对应关系
  3. 标出当前高度回算逻辑将被替换的节点
- 完成定义:
  1. `EP-1` 可直接按轨道映射改代码
  2. 不再边改边猜当前布局链路
- 回归点:
  - 无代码回归
  - 仅要求映射清晰

### T-003 固定 Rail 高度模型

- 状态: `done`
- 归属阶段: `EP-1`
- 目标:
  将主布局从自然高度回推改为固定 rail 高度模型。
- 输入:
  - `PA-001`
  - `EP-1`
- 修改范围:
  - [internal/tui/ui/app/layout.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/layout.go)
  - 可能新增 rail 辅助文件
- 主要工作:
  1. 定义 rail 高度常量或配置
  2. 将 `RenderApp()` 改为按 rail 固定分配
  3. 删除 header / toast / footer 行数反推 panel 高度路径
- 完成定义:
  1. panel 高度与 toast/search/footer 内容解耦
  2. layout 主路径只保留固定轨道计算
- 回归点:
  - `RG-001 ~ RG-006`
  - `RM-001 ~ RM-006`
  - `RS-001 ~ RS-004`

### T-004 小终端降级策略

- 状态: `done`
- 归属阶段: `EP-1`
- 目标:
  为固定骨架建立明确的小终端退化规则。
- 输入:
  - `T-003`
- 修改范围:
  - [internal/tui/ui/app/layout.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/layout.go)
  - 可能新增文档说明
- 主要工作:
  1. 定义正常布局下限
  2. 定义紧凑布局规则
  3. 定义不可用时的最小提示输出
- 完成定义:
  1. 小终端不出现内容重叠或严重截断失控
  2. 有明确的降级行为
- 回归点:
  - `RS-001 ~ RS-004`

### T-005 抽离 Query Rail Renderer

- 状态: `done`
- 归属阶段: `EP-2`
- 目标:
  将 filter / search / command 输入统一迁到 query rail。
- 输入:
  - `PA-002`
  - `BR-004`
- 修改范围:
  - [internal/tui/ui/app/layout.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/layout.go)
  - [internal/tui/ui/component/filter.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/component/filter.go)
- 主要工作:
  1. 定义 query rail 渲染入口
  2. 将旧 `searchBar` 插入逻辑迁出 panel 上方路径
  3. 区分 `Filter / Search / Command` 三种显示语义
- 完成定义:
  1. query rail 成为唯一查询输入显示位置
  2. panel 高度不再因 query 激活变化
- 回归点:
  - `RM-001`
  - `RM-002`
  - `RM-003`
  - `RG-001`
  - `RG-004`

### T-006 抽离 Message Rail Renderer

- 状态: `done`
- 归属阶段: `EP-2`
- 目标:
  将轻量通知统一迁到 message rail。
- 输入:
  - `PA-002`
  - `BR-005`
- 修改范围:
  - [internal/tui/ui/app/layout.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/layout.go)
  - [internal/tui/ui/component/toast.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/component/toast.go)
- 主要工作:
  1. 定义 message rail 渲染入口
  2. 让现有 toast 先落到固定 rail
  3. 为 filter/search 状态提示预留接入口
- 完成定义:
  1. toast 不再改变 panel 高度
  2. message rail 成为统一短提示容器
- 回归点:
  - `RG-001 ~ RG-006`
  - `RM-001`
  - `RM-002`

### T-007 固定 Footer 三行结构

- 状态: `done`
- 归属阶段: `EP-3`
- 目标:
  将 footer 固定为三行结构，为统一投影层做容器准备。
- 输入:
  - `PA-001`
  - `PA-004`
- 修改范围:
  - [internal/tui/ui/widget/footer/footer.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/footer/footer.go)
  - [internal/tui/ui/widget/footer/oplog.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/footer/oplog.go)
- 主要工作:
  1. 固定全局动作行
  2. 固定上下文动作行
  3. 固定状态/操作日志行
  4. 保证 mode 切换不引起高度变化
- 完成定义:
  1. footer 永远三行
  2. 操作日志为空时仍有固定占位
- 回归点:
  - `RG-001 ~ RG-006`
  - `RM-004`
  - `RM-005`
  - `RM-006`

### T-008 Batch-A 综合回归

- 状态: `verifying`
- 归属阶段: `EP-0/1/2/3`
- 目标:
  对固定骨架、query rail、message rail、footer rail 做一次统一回归。
- 输入:
  - `T-003 ~ T-007`
- 修改范围:
  - 无代码
  - 回归记录
- 主要工作:
  1. 按页面矩阵回归
  2. 按模式矩阵回归
  3. 按小终端矩阵回归
  4. 记录问题归属到下一批次或返工本批次
- 完成定义:
  1. Batch-A 交付结果明确
  2. 可以决定是否进入 Batch-B
- 回归点:
  - `RG-001 ~ RG-006`
  - `RM-001 ~ RM-006`
  - `RS-001 ~ RS-004`

## Batch-B 预备任务单

### T-101 页面模板归类与接口草案

- 状态: `todo`
- 归属阶段: `EP-4`
- 目标:
  给现有页面建立模板归类表和模板接口草案。
- 完成定义:
  1. 每个页面都有模板归属
  2. 模板接口可作为后续实现入口

### T-102 日志页专项改造

- 状态: `todo`
- 归属阶段: `EP-5`
- 目标:
  将日志页改为真实 panel body width + viewport 级渲染。
- 完成定义:
  1. 宽度来源正确
  2. search / wrap / scroll 语义与 BR-004 一致

### T-103 投影层接口草案

- 状态: `todo`
- 归属阶段: `EP-6`
- 目标:
  为 title / summary / selection / shortcut 建立统一 projector 接口。
- 完成定义:
  1. Footer / Help 可切到同源
  2. 标题摘要不再分散手工拼接

## Batch-C 预备任务单

### T-201 Detail Page 接线到 BR-001

- 状态: `todo`
- 归属阶段: `EP-7`
- 目标:
  让 detail page 正式接线 image detail section model。
- 完成定义:
  1. 镜像详情页结构化渲染可用
  2. `History` 区段不再固定占位

### T-202 表格列宽缓存第一版

- 状态: `todo`
- 归属阶段: `EP-8`
- 目标:
  为高频表格建立宽度与表头缓存。
- 完成定义:
  1. 不缓存整表字符串
  2. 缓存收益明确且结果稳定

### T-203 增量性能热点收敛

- 状态: `todo`
- 归属阶段: `EP-9`
- 目标:
  对 profiling 后确认的热点做增量优化。
- 完成定义:
  1. 只处理真实热点
  2. 不做无根据的预支优化

## 回归记录模板

```md
### Batch-X / EP-x 回归记录

- 日期:
- 范围:
- 结果:
  - `RG-*`:
  - `RM-*`:
  - `RS-*`:
- 问题:
- 是否可进入下一阶段:
```

## 任务记录模板

```md
### T-xxx

- 状态: `todo|doing|verifying|done|blocked`
- 归属阶段:
- 负责人:
- 输入:
- 修改范围:
- 已完成:
- 风险:
- 下一步:
```
