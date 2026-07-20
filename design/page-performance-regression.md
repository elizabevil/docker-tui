# dtui 页面与性能优化回归记录

> 建立日期: 2026-07-20
> 作用: 固定页面布局优化的现状映射、回归格式与阶段结果。

## 固定五轨道模型

| 轨道 | 固定高度 | 当前内容来源 | 空闲行为 |
|---|---:|---|---|
| `Header Rail` | 4 | `header.Render()` | 保留 4 行空轨道 |
| `Message Rail` | 1 | `AppModel.ToastMessage` | 保留 1 行空轨道 |
| `Query Rail` | 3 | `ModeFilter / ModeCommand` | 保留 3 行空轨道 |
| `Panel Rail` | 剩余空间 | 当前页面 renderer | 最低 4 行 |
| `Footer Rail` | 3 | `footer.Render()` | 三行始终存在 |

Panel body 高度只由 `Panel Rail - border(2) - title(1)` 决定，不再读取 Header、Toast、Query 或 Footer 的实际输出行数。

## 旧布局到新布局映射

| 旧节点 | 旧行为 | 新归属 | 已替换路径 |
|---|---|---|---|
| `header.Render()` | 与 toast 合并后测量自然高度 | `Header Rail` | 固定 4 行 |
| `renderToast()` | 插入 header 下方并增加 top 高度 | `Message Rail` | 固定 1 行、单行截断 |
| `renderSearchBar()` | 激活时插入 panel 上方 | `Query Rail` | 固定 3 行，区分 Filter/Search/Command |
| `renderMiddlePanel()` | 使用权重高度再次计算 body | `Panel Rail` | 直接接收 rail 高度并计算 body |
| `footer.Shortcuts() + StatusBar()` | 内容行数决定 footer 高度 | `Footer Rail` | 全局动作/上下文动作/状态三行 |
| `sectionHeights()` | 按 top/content/bottom 权重分配 | 已移除 | `calculateRailHeights()` |

## 小终端策略

- `>= 120x32`: 标准布局。
- `>= 100x24`: 保留完整五轨道，缩减 Panel 可用空间。
- `>= 80x20`: 最小可用布局，关闭百分比边距后仍不足时由 Panel 保底。
- `< 80x20`: 不渲染业务页面，显示终端最小尺寸提示。

## 回归记录模板

```md
### Batch-X / EP-x

- 日期:
- 代码范围:
- 自动化验证:
- 页面矩阵: `RG-001 ~ RG-006`
- 模式矩阵: `RM-001 ~ RM-006`
- 尺寸矩阵: `RS-001 ~ RS-004`
- 失败项:
- 遗留风险:
- 结论: `通过 / 返工 / 继续验证`
```

## Batch-A / EP-0 ~ EP-3

- 日期: 2026-07-20
- 代码范围: 固定五轨道、Query Rail、Message Rail、三行 Footer Rail。
- 自动化验证: `CGO_ENABLED=0 go test ./...` 通过。
- 自动化覆盖:
  - 固定轨道总高度与 Panel body 高度。
  - Normal/Filter/Message 状态切换时应用输出高度不变。
  - Filter/Search/Command 查询语义识别。
  - Footer 在 Normal/Mark/Detail 下固定三行。
  - `< 80x20` 最小尺寸规则。
- 页面矩阵: `RG-001 ~ RG-006` 待真实终端交互回归。
- 模式矩阵: `RM-001 ~ RM-006` 核心高度约束通过，具体交互待真实终端回归。
- 尺寸矩阵: 尺寸分配自动化通过，`RS-001 ~ RS-004` 视觉结果待真实终端回归。
- 失败项: 无自动化失败。
- 遗留风险: Footer 长动作行会按终端宽度截断，完整动作发现依赖后续 Help 统一投影层。
- 结论: `继续验证`。
