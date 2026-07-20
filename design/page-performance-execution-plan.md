# dtui 页面与性能优化执行计划

> 建立日期: 2026-07-20
> 作用: 将已审批通过的页面与性能优化方案拆成可执行阶段计划。

本文不替代:

- [page-performance-approval-proposals.md](page-performance-approval-proposals.md): 审批结论
- [page-performance-optimization-design.md](page-performance-optimization-design.md): 优化设计总方案
- [bugfix-design.md](bugfix-design.md): BR-001 ~ BR-005 正式设计

本文只负责:

- 明确执行阶段
- 明确依赖关系
- 明确每阶段交付物
- 明确每阶段验收边界

## 总体执行原则

### 1. 先骨架，后模板，最后局部性能

执行优先级固定为:

1. 固定页面骨架
2. Query / Message 轨道收敛
3. 页面模板收敛
4. 统一投影层接线
5. 单页热点优化
6. 表格布局缓存

### 2. 不与既有 BR 方案冲突

必须满足:

1. `BR-001` 决定 detail page 的结构化方向
2. `BR-002 / BR-003` 决定动作系统、Footer、Help 的唯一事实源
3. `BR-004` 决定 Filter / Search 模式与 Query rail 行为
4. `BR-005` 决定 Message / Notification / Operation Log 的统一投影来源

### 3. 每阶段都必须可单独回归

每个阶段都要求:

1. 至少能单独运行并验证
2. 不要求一次性重写所有页面
3. 允许“新旧并存”，但必须有明确迁移边界

## 阶段总览

| 阶段 | 主题 | 对应审批 | 优先级 | 目标状态 |
|---|---|---|---|---|
| `EP-0` | 基线与护栏 | 全部 | P0 | 建立可回归边界 |
| `EP-1` | 固定五轨道骨架 | `PA-001` | P0 | 消除主布局跳变 |
| `EP-2` | Query / Message Rail 接线 | `PA-002` | P0 | 固定查询与消息轨道 |
| `EP-3` | Footer Rail 固定化 | `PA-001` `PA-004` | P0 | 固定底部三行结构 |
| `EP-4` | 页面模板抽象 | `PA-003` | P1 | 建立 List / Split / Detail / Log / Help 模板 |
| `EP-5` | 日志页专项优化 | `PA-006` | P1 | 修复日志页宽度与 viewport 行为 |
| `EP-6` | 统一投影层接线 | `PA-004` | P1 | Footer / Help / 标题摘要同源 |
| `EP-7` | 详情页结构化迁移 | `PA-007` `BR-001` | P1 | Detail page 切到 section model |
| `EP-8` | 表格列宽缓存 | `PA-005` | P2 | 高频表格计算收敛 |
| `EP-9` | 后续增量性能优化 | `PA-008` | P3 | 可选增强 |

## EP-0 基线与护栏

- 目标:
  在改主布局前，先建立“看得见”的回归边界。
- 依赖:
  无
- 主要工作:
  1. 明确当前页面类型清单
  2. 明确关键回归路径:
     - containers list
     - images list
     - compose split view
     - logs view
     - detail view
     - help view
  3. 记录当前已知特殊模式:
     - filter
     - search
     - command
     - mark
     - detail
     - log view
  4. 明确小终端降级边界
- 交付物:
  1. 回归检查清单
  2. 阶段验收模板
- 验收:
  1. 后续每阶段都能用同一套页面清单回归
  2. 不再凭感觉判断“页面大致没坏”

## EP-1 固定五轨道骨架

- 目标:
  将 `RenderApp()` 改为固定五轨道布局，消除 panel 高度抖动。
- 依赖:
  `EP-0`
- 对应审批:
  `PA-001`
- 主要工作:
  1. 重构主布局计算模型
  2. 引入固定轨道高度:
     - `Header Rail`
     - `Message Rail`
     - `Query Rail`
     - `Panel Rail`
     - `Footer Rail`
  3. 删除“先渲染再按行数回推 panel 高度”的路径
  4. 增加小终端降级规则
- 目标文件范围:
  - [internal/tui/ui/app/layout.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/layout.go)
  - 可能新增 `layout shell` 相关文件
- 交付物:
  1. 固定轨道布局实现
  2. 小终端降级策略
  3. 骨架层回归说明
- 验收:
  1. toast 出现/消失时 panel 高度不变
  2. `/`、`:` 激活/退出时 panel 高度不变
  3. footer 内容变化时 panel 高度不变
  4. 小终端下页面仍可用

## EP-2 Query / Message Rail 接线

- 目标:
  把查询输入和轻量消息从“插入块”改为固定轨道内容。
- 依赖:
  `EP-1`
- 对应审批:
  `PA-002`
- 主要工作:
  1. 抽离 query rail renderer
  2. 抽离 message rail renderer
  3. 将现有 `searchBar` 插入逻辑迁移到 query rail
  4. 让 toast 先落到 message rail
  5. 为 BR-004 的 Filter / Search 提示预留 message rail 接口
- 目标文件范围:
  - [internal/tui/ui/app/layout.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/app/layout.go)
  - [internal/tui/ui/component/filter.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/component/filter.go)
  - [internal/tui/ui/component/toast.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/component/toast.go)
- 交付物:
  1. Query rail
  2. Message rail
  3. 旧插入式 search / toast 路径移除或兼容桥接
- 验收:
  1. query rail 空闲时占位但不干扰 panel
  2. filter / search / command 都进入统一 query rail
  3. toast 与轻量提示进入统一 message rail

## EP-3 Footer Rail 固定化

- 目标:
  固定底部三行结构，为后续统一动作投影层做容器准备。
- 依赖:
  `EP-1`
- 对应审批:
  `PA-001`
  `PA-004`
- 主要工作:
  1. 固定 Footer Rail 为三行
  2. 将全局动作行、上下文动作行、状态/操作日志行明确分层
  3. 保持第三行始终占位
  4. 清理因 mode 切换导致的底部高度波动
- 目标文件范围:
  - [internal/tui/ui/widget/footer/footer.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/footer/footer.go)
  - [internal/tui/ui/widget/footer/oplog.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/widget/footer/oplog.go)
- 交付物:
  1. 固定三行 footer
  2. footer rail 结构约束
- 验收:
  1. mode 切换时 footer 不跳高
  2. 操作日志为空时仍保留第三行占位

## EP-4 页面模板抽象

- 目标:
  建立标准页面模板，避免各页面继续直接拼完整内容。
- 依赖:
  `EP-1`
  `EP-2`
  `EP-3`
- 对应审批:
  `PA-003`
- 主要工作:
  1. 定义五类模板:
     - `List Page`
     - `Split Page`
     - `Detail Page`
     - `Log Page`
     - `Help Page`
  2. 建立模板级 renderer 边界
  3. 将现有页面标记到模板分类
  4. 先抽象模板接口，不要求全部页面一次改完
- 交付物:
  1. 页面模板接口
  2. 页面归类表
  3. 首批模板接线点
- 验收:
  1. 后续页面优化不再直接从单页视图函数起草结构
  2. 新需求可先判断模板归属

## EP-5 日志页专项优化

- 目标:
  单独修复日志页宽度、viewport、wrap 和 search 投影问题。
- 依赖:
  `EP-2`
  `EP-4`
- 对应审批:
  `PA-006`
  `BR-004`
- 主要工作:
  1. 将日志页 wrap 宽度改为 panel body width
  2. 将 wrap / highlight 计算限制在当前 viewport
  3. 让 search 状态投影到 query rail / message rail
  4. 将日志页明确收敛到 `Log Page` 模板
- 目标文件范围:
  - [internal/tui/ui/pages/logs/view.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/pages/logs/view.go)
  - 日志相关 keyboard / search 状态文件
- 交付物:
  1. 日志页 viewport 优化
  2. Search 行为与视图投影拆分
- 验收:
  1. 宽度来源正确
  2. 长日志滚动不卡顿加剧
  3. Search 语义与 BR-004 一致

## EP-6 统一投影层接线

- 目标:
  让 Footer / Help / 标题摘要来自统一投影层。
- 依赖:
  `EP-3`
  `EP-4`
- 对应审批:
  `PA-004`
  `BR-002`
  `BR-003`
  `BR-005`
- 主要工作:
  1. 抽离 `TitleProjector`
  2. 抽离 `SummaryProjector`
  3. 抽离 `SelectionProjector`
  4. 抽离 `ShortcutProjector`
  5. Help 与 Footer 切换到同一动作来源
- 交付物:
  1. 投影层接口
  2. Footer / Help 同源接线
  3. 标题摘要统一接口
- 验收:
  1. Help 与 Footer 不再维护两套键位真相
  2. 标题摘要不再散落在页面函数里手工拼接

## EP-7 详情页结构化迁移

- 目标:
  将 detail page 优化与 `BR-001` 的 section model 正式接线。
- 依赖:
  `EP-4`
  `EP-6`
- 对应审批:
  `PA-007`
  `BR-001`
- 主要工作:
  1. 先接镜像详情页
  2. 让 detail page 消费 `ImageDetailViewModel`
  3. 从文本反解析过渡到 section model 渲染
  4. 保持 detail page 模板稳定
- 交付物:
  1. 镜像详情页结构化渲染接线
  2. detail page 模板第一版
- 验收:
  1. `History` 区段可结构化渲染
  2. detail page 不再依赖旧文本解析链作为唯一来源

## EP-8 表格列宽缓存

- 目标:
  在不引入整表缓存复杂度的前提下，先收敛宽度和表头计算成本。
- 依赖:
  `EP-4`
- 对应审批:
  `PA-005`
- 主要工作:
  1. 新增 `TableLayoutCache`
  2. 缓存列集合、profile、container width、resolved headers、column widths
  3. 明确定义失效条件:
     - resize
     - profile change
     - columns schema change
     - locale change
  4. 第一阶段不缓存整表字符串
- 目标文件范围:
  - [internal/tui/ui/component/table.go](/home/debi/IdeaProjects/docker-tui/internal/tui/ui/component/table.go)
  - 可能新增 table layout cache 文件
- 交付物:
  1. 表格宽度缓存
  2. cache key 与失效策略
- 验收:
  1. 渲染结果与缓存前一致
  2. 高频表格页面计算路径减少
  3. 不出现缓存错位或列宽过期问题

## EP-9 后续增量性能优化

- 目标:
  处理第一阶段之后仍然存在的热点，但不阻塞主干改造。
- 依赖:
  `EP-5`
  `EP-8`
- 对应审批:
  `PA-008`
- 可选工作:
  1. `VisibleLen()` 高频结果缓存
  2. render model 分层
  3. 更细粒度 dirty-path 更新
  4. 特定高频页刷新节流
- 交付物:
  后续按热点决定
- 验收:
  以 profiling 或实际卡点为依据，不做预支实现

## 建议实施顺序

1. `EP-0`
2. `EP-1`
3. `EP-2`
4. `EP-3`
5. `EP-4`
6. `EP-5`
7. `EP-6`
8. `EP-7`
9. `EP-8`
10. `EP-9`

## 建议批次划分

### 批次 A: 主骨架落地

- `EP-0`
- `EP-1`
- `EP-2`
- `EP-3`

目标:

- 固定骨架
- 固定 query / message / footer 轨道

### 批次 B: 页面模板与热点页面

- `EP-4`
- `EP-5`
- `EP-6`

目标:

- 开始从“页面函数直拼”切到模板和投影层

### 批次 C: 详情页与表格性能

- `EP-7`
- `EP-8`
- `EP-9`

目标:

- 将 BR-001 detail 方向与页面优化正式接线
- 处理高频表格性能成本

## 阻塞关系

### 硬阻塞

1. `EP-1` 未完成前，不应开始大规模页面模板改造
2. `EP-3` 未完成前，不应开始 Footer / Help 的最终投影层接线
3. `EP-4` 未完成前，不应全面铺开 detail / log / split 页面统一化

### 软阻塞

1. `EP-6` 最好在 `BR-002 / BR-003` 输入系统改造有明确注册表后推进
2. `EP-7` 最好与 `BR-001` 的 image detail builder 实现同步
3. `EP-8` 最好在页面模板稳定后再做，避免 cache key 频繁变化

## 阶段完成记录模板

```md
### EP-x

- 状态: `pending|implementing|verifying|done`
- 负责人:
- 相关审批:
- 相关 BR:
- 已完成:
- 风险:
- 验收结果:
```
