# 镜像 History 顶层页 (H 键)

> 任务卡 / BR-034 / 历史页

## 元信息

- **关联编号**:BR-034 + BR-008(部分,H 键路由)
- **优先级**:high
- **状态**:`open`
- **依赖**:BR-007(详情页 revision 缓存)、BR-016(详情页 ClampVisibleOffset)
- **历史**:曾因性能问题取消;BR-008 把 H 键暂时让给 BR-007 详情页 ClampVisibleOffset 收尾
- **关联设计**:[docs/feature-design.md §5.1](../feature-design.md)、[bugfix-requirements.md BR-034](../bugfix-requirements.md)

## 目标

在**镜像页**(不是详情页内)按 `H` 键进入**独立顶层页** `ModeHistory`,浏览当前选中镜像的 layer history。解决历史性能问题(大镜像 >50 layer 滚动卡顿)。

## 代码结构索引

### 必须读懂的文件

| 文件 | 作用 |
|---|---|
| `internal/data/runtime/image.go` | `ImageHistoryLayer` 数据结构(line 57) + `ImageService.Inspect` 返回 `*ImageDetail`(含 history) |
| `internal/data/runtime/docker/images.go:107` | `InspectImageDetailContext` history 字段填充逻辑 |
| `internal/data/runtime/podman/service_image.go` | Podman `Inspect` 实现 history 来源 |
| `internal/tui/state/images.go` | 镜像列表 state + `Selected()` 方法 |
| `internal/tui/state/detail.go` | 现有 `DetailState` 结构(若复用 `InspectImageDetail` + revision 缓存机制,见 BR-007) |
| `internal/tui/ui/pages/detail/image.go:14` | 镜像详情页 section 构造(已移除 history,见 BR-008) |
| `internal/tui/keyboard/keyboard.go:215` | 当前 `KeyH` 入口,执行 `ToggleHeader`(需改路由) |
| `internal/tui/ui/app/layout.go:104` | `resolveStandardLayout` 面板几何(参考 History 页面放哪里) |
| `internal/tui/ui/app/mode_router.go`(若存在) | Mode 路由(新增 ModeHistory) |
| `internal/tui/keyboard/navigate.go` | ToXxxDetail 模式入口模板 |

### 必须修改的文件

| 文件 | 改动 |
|---|---|
| `internal/tui/state/history.go`(或 `state/image_history.go`,新文件) | `HistoryState` 结构:`ImageID` / `Layers []ImageHistoryLayer` / `Cursor` / `ViewOffset` / `Filter` / `Loading` |
| `internal/tui/state/navigation.go` | 加 `ModeHistory` 枚举 + `PanelImages` 时按 H 进入 |
| `internal/tui/keyboard/keyboard.go:215` | 把 `KeyH` 从 ToggleHeader 改为按 `ModeImages + History` 路由 |
| `internal/tui/ui/pages/history/`(新目录) | 新增 `pages/history/view.go` + `state.go`(类似 detail) |
| `internal/tui/ui/app/page_templates.go` | 注册 `ModeHistory` → 渲染 `history/view.RenderView(...)` |
| `internal/tui/ui/action/registry.go` | 加 History 帮助页条目 + Footer 提示 |

### 必须新增的文件

| 文件 | 作用 |
|---|---|
| `internal/tui/ui/pages/history/view.go` | History 顶层页渲染:镜像名/Tag 头部 + Created/Size/CreatedBy/Comment 表格 + filter 输入 |
| `internal/tui/ui/pages/history/view_test.go` | 单元测试:渲染 / 滚动 / 过滤 |
| `internal/tui/ui/pages/history/state.go` | `HistoryState` 定义 |
| `internal/tui/state/history.go` | 状态:Loading / Error / Layers / Cursor / ViewOffset / Filter |
| `internal/tui/keyboard/history_keys.go` | `handleHistoryKeys`:`j/k/PgUp/PgDn/g/G//Filter/Esc/s source` |
| `internal/data/i18n/lang/zh.jsonc` + `en.jsonc` + `ja.jsonc` | 加 `history.title` / `history.layer` / `history.created_by` / `history.size` 等 key |
| `test/benchmarks/history_bench_test.go` | 大镜像(>50 layer)滚动 benchmark(对应 BR-016 性能验证) |

### 不应触碰的文件

- `internal/data/runtime/`(image detail 已包含 history,无新 API)
- 详情页 sections / YAML/JSON 视图(沿用 BR-007 缓存机制即可,**不重渲染** history)
- 其它 panel 的 Mode 路由

## 操作链路

```text
User 在 ModeImages 按 H
  → keyboard/keyboard.go:215 case keys.KeyH
      if m.Navigation.ActivePanel == state.PanelImages && c == 'H' (小写)
        触发 History 入口
  → keyboard/navigate.go:ToImageHistory(m)
      m.Navigation.Mode = state.ModeHistory
      m.History.Open(img.ID, m.Resources.Images.Selected().RepoTags[0])
      tea.Cmd: historyFetchCmd(client, img.ID) → HistoryLoadedMsg
  → update_events.go:handleHistoryLoaded(m, msg)
      m.History.Layers = msg.Layers
      m.History.Loading = false
      m.History.Cursor = 0; m.History.ViewOffset = 0
  → ui/pages/history/view.go:RenderView(m, panelHeight, panelWidth)
      渲染层表(沿用 BR-007 缓存:按 revision 缓存 + 只渲染可见行)
  → keyboard/history_keys.go:handleHistoryKeys
      j/k → Cursor ±1 + ClampVisibleOffset (复用 BR-016 机制)
      PgDn/PgUp → Cursor ±20
      g → top; G → bottom
      / → ToFilter
      Esc → BackFromHistory
      (无 s source,只展示 layer 表,不渲染 raw source)
  → 过滤:用户输入 / 字符串 → Filter 生效 → 表格只显示 CreatedBy 匹配行
```

## 验收标准

- [ ] 镜像页按 `H` 键,弹出 History 顶层页,标题为 `<image>:<tag> history`。
- [ ] 行数等于镜像 layer 数(inspect API 返回)。
- [ ] 每行展示:`Created`(相对时间)/ `Size`(人类可读)/ `CreatedBy`(摘要)/ `Comment`。
- [ ] 选中镜像 > 50 layer,上下滚动 / PgDn / j 翻页无卡顿(性能:验证 BR-016 缓存机制复用)。
- [ ] 顶部固定显示镜像 `Repository:Tag` 与总 layer 数。
- [ ] `/` 触发 Filter,输入字符串后 CreatedBy 模糊匹配,行数实时减少。
- [ ] `Esc` 返回镜像页,ModeHistory → ModeImages,State 已清理。
- [ ] 详情页不再渲染 History 段(沿用 BR-008 决策;不破坏 BR-007)。
- [ ] `s` 键在 ModeHistory 下不切 YAML/JSON(History 页没有 raw source)。
- [ ] 大镜像 benchmark(类似 BR-016):`BenchmarkRenderCachedHistory` 通过。
- [ ] 镜像不存在 / inspect 失败:显示明确错误占位,不静默。
- [ ] Help 页 `/ F1` 增加"镜像页 H = History"提示。

## 风险

| 风险 | 说明 |
|---|---|
| **历史性能问题复现** | BR-008 注明曾因"每次滚动重建 section"卡顿取消。新实现必须按 revision 缓存 + 只渲染可见行,**继承 BR-016 的 ClampVisibleOffset**;否则性能问题会复发 |
| **H 键路由与 BR-008 冲突** | BR-008 H 键期望进 Help;本任务 H 键期望进 History。需要用户决策二选一,或模式路由:H 全局进 Help,镜像页 H 进 History。**当前设计(§5.1)选择后者** |
| **inspect 数据大小** | Docker 镜像 history 数据结构 `ImageHistoryLayer` 字段长度可能很长(CreatedBy 完整命令行);需要在 view 层做截断 |
| **wide chars / ANSI 残留** | CreatedBy 字段可能含 shell 转义序列;需要 ANSI 长度计算正确 |
| **表格布局** | History 页是否走 table layout?若走需新增 `tables/history.jsonc`;若走简单列表则更轻量。建议简单列表 + 4 列(类似 detail page) |
| **i18n** | Created / Size / CreatedBy / Comment 列名需多语言 |
| **多个镜像 history** | 历史记录是否包含已删除 tag 的镜像?若 `Selected()` 返回 nil,需 no-op |

## 建议任务分解

1. **TASK-BR034-A:状态 + 数据加载**
   - `state/history.go` + `HistoryFetchCmd`
   - `HistoryLoaded` Msg + `update_events.go:handleHistoryLoaded`
   - **预估**:小模型独立可完成
2. **TASK-BR034-B:渲染 + 滚动**
   - `pages/history/view.go` + `view_test.go`
   - 沿用 BR-007 的按 revision 缓存 + ClampVisibleOffset
   - **预估**:小模型独立可完成
3. **TASK-BR034-C:键盘 + 过滤**
   - `keyboard/history_keys.go` + Filter 集成
   - **预估**:小模型独立可完成
4. **TASK-BR034-D:H 键路由 + Mode 路由 + i18n**
   - 改 `keyboard/keyboard.go:215` `KeyH` 入口
   - 加 `ModeHistory` / `PanelHistory`(若走 panel) / `Navigate ToHistory`
   - **预估**:小模型辅助,主模型决策 H 键冲突
5. **TASK-BR034-E:benchmark + 性能验证**
   - `BenchmarkRenderCachedHistory` 加入 test/benchmarks
   - 验证 >50 layer 滚动流畅
   - **预估**:小模型独立可完成

## 待确认项

- [ ] **H 键冲突**:BR-008 期望 H 进 Help,本任务期望 H 进 History(模式路由)。需要主模型决策:**全局 H = Help + 镜像页 H = History**,还是**全局 H = History(详情页 / 其它页 / Mirror 页统一)**?详见 [feature-design.md §5.1](../feature-design.md)
- [ ] **History 页 vs 详情页关系**:History 是独立顶层页(`ModeHistory`),还是详情页的子模式(`ModeDetail + DetailSourceHistory`)?独立页更清晰,占用一个完整 panel
- [ ] **表格 vs 简单列表**:参考 detail 页的表格 layout 还是自定义渲染?简单列表更轻
- [ ] **History 页与 Compose 子视图等其它详情页的滚动状态共享**:`state.Detail.Documents map[DetailSource]DetailDocument` 已经在详情页用,History 是否复用 `Documents` map(以 `DetailSource = "history"` 为 key)
- [ ] **Filter 是否影响 Cursor / ViewOffset**:filter 缩小行集后 Cursor 越界如何 clamp

## 参考

- [docs/feature-design.md §5.1 镜像 History 顶层页 (H 键)](../feature-design.md)
- [docs/bugfix-requirements.md BR-008](../bugfix-requirements.md)(H 键路由历史决策)
- [docs/bugfix-requirements.md BR-016](../bugfix-requirements.md)(ClampVisibleOffset 机制)
- [docs/bugfix-requirements.md BR-007](../bugfix-requirements.md)(按 revision 缓存文档)