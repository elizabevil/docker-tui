# 取消 Q 退出,改为每页多功能 Action Bar

> 任务卡 / BR-039

## 元信息

- **关联编号**:BR-039
- **优先级**:medium
- **状态**:`done`
- **依赖**:所有 panel 已有功能(否则 Action Bar 内容为空)
- **关联设计**:[docs/feature-design.md §5.6](../feature-design.md)

## 目标

取消 `Q` 作为退出应用的快捷键,增加按当前 panel 动态生成的 Action Bar。Action Bar 专门承载 Rename、Import、Login 等无直接快捷键的复杂/多步操作;已有直接快捷键与全局命令不重复展示。命令面板 (`:` 命令)继续保留。

## 代码结构索引

### 必须读懂的文件

| 文件 | 作用 |
|---|---|
| `internal/tui/keys/registry.go` | `ActionQuit ← KeyCtrlC`;`ActionActionBar ← KeySemicolon` |
| `internal/tui/keyboard/actions.go` | `case ActionQuit: tea.Quit` 与 Action Bar 动作分发 |
| `internal/tui/keyboard/command.go` | 命令面板 `:` 入口(`executeCommand`) |
| `internal/tui/keys/commands.go` | 命令列表(`:rename` / `:top` / `:port` 等) |
| `internal/tui/ui/action/registry.go` | 帮助页 / footer 投影框架 |
| `internal/tui/ui/widget/dialog/` | dialog 复用(CenterOnPanel / BR-040 配套) |
| `internal/tui/keyboard/container_action.go` | 容器 ActionBar 内容来源 |
| `internal/tui/keyboard/image_action.go` | 镜像 ActionBar 内容来源 |

### 必须修改的文件

| 文件 | 改动 |
|---|---|
| `internal/tui/keys/registry.go` | `{ActionQuit, []string{KeyCtrlC}, app}` 并注册 Action Bar |
| `internal/tui/keyboard/actions.go` | Action Bar 与 typed panel 动作分发 |
| `internal/tui/ui/action/registry.go` | Footer / Help 投影 Action Bar 快捷键 |
| `internal/tui/ui/app/layout.go` | 标准与紧凑布局渲染 Action Bar 浮层 |

### 必须新增的文件

- `internal/tui/ui/widget/actionbar/`(新子包)
  - `actionbar.go`:ActionBar 浮层渲染组件(类似 vim command palette)
- `internal/tui/actionbar/registry.go`:按 `ActivePanel` 动态生成 typed 动作列表
- `internal/tui/state/actionbar.go`:ActionBar state
- `internal/tui/keyboard/actionbar_keys.go`:`handleActionBarKeys`(j/k 选择 / Enter 执行 / Esc 关闭 / / 过滤)

### 不应触碰的文件

- 全局键盘分发路径(只新加 `;` 入口,不破坏其它)
- 已有命令面板 `:`(`:rename` / `:top` / `:port` 保留)
- 退出流程(仍由 Ctrl+C + EscPending 双段确认)

## 操作链路

```text
User 在任意 panel 按 ;
  → keys/registry.go:ActionActionBar 命中
  → keyboard/actions.go:handleActionBar(m)
      m.Navigation.Mode = state.ModeActionBar
      m.ActionBar.Open(panel) // 加载该 panel 的动作列表
  → ui/widget/actionbar/actionbar.go:RenderBar(m)
      按 ActivePanel 加载动作条目:
        镜像页: Pull / Tag / Push / Save / Load / Import / History / Prune / Remove / Refresh / Filter ...
        容器页: Start / Stop / Kill / Exec / Logs / Stats / Top / Port / Update / Diff / Export / Commit / Wait / Copy / Filter ...
        卷 / 网络 / Compose 页: 各自动作
        通用: Login / Switch Runtime / Refresh Connections / Help / Events (F3) / Quit (Ctrl+C)
      渲染为浮层列表(类似 fuzzy finder)
  → keyboard/actionbar_keys.go:handleActionBarKeys
      j/k → 选择上 / 下
      / → filter 输入
      Enter → 执行选中动作(转调 keyboard/actions.go:case 已有)
      Esc → 关闭 ActionBar
      数字键 1-9 → 跳到第 N 个动作
```

## 验收标准

### 取消 Q

- [x] 按 Q 不再退出应用,保持未分配且无提示。
- [ ] Help / Footer 不再显示 `Q Quit`。
- [ ] Ctrl+C 仍可退出(双段确认沿用)。

### Action Bar

- [ ] 任意 panel 按 `;`(或约定键)弹出 Action Bar 浮层。
- [ ] 浮层显示当前 panel 可用动作 + 当前键位(`Tag (Ctrl+T)` / `Save (Ctrl+E)`)。
- [ ] j/k 选择 / Enter 执行 / Esc 关闭 / `/` filter / 数字键跳转。
- [ ] 选中动作后行为与原直键一致。
- [ ] 选中"无键位"动作(如 BR-033 暂未分配键位的高级动作)也能执行。
- [ ] detail / logs / help 页不强行弹 Action Bar(空列表或直接 no-op)。
- [ ] 命令面板 `:` 仍可用,所有原命令保留(`:rename` 等不取消)。
- [ ] Action Bar 与命令面板的交互不冲突(各自独立状态)。
- [ ] Action Bar 内 j/k 选中后,某些动作(需要表单的)直接打开对应 dialog。
- [ ] Quit (Ctrl+C) 在 Action Bar 列表中可见。

### 跨页一致性

- [ ] 切换 panel 后,Action Bar 内容动态更新(镜像页的 Action Bar 不显示容器动作)。
- [ ] Help 页新增 "Action Bar (`;`)" 段落。

## 风险

| 风险 | 说明 |
|---|---|
| **入口键冲突** | `;` 是否与现有命令冲突?vim 风格可接受;若用户偏好,可在 docs/navigation.md 让用户决策 |
| **动作列表过长** | 镜像页可能有 10+ 动作;需 filter 或分类滚动 |
| **动态发现** | 动作列表是硬编码还是动态生成(从 `ActionRegistry` 推导)?前者简单,后者灵活 |
| **Cmd+C 二段确认** | 退出保留 Ctrl+C 双段确认(沿用 `EscPending` 机制),不破坏现有 |
| **Help 漂移** | 现有 Help / Footer 大批更新;`navigation.md` 同步 |
| **空页面** | detail / logs / help 页无 panel 级动作,Action Bar 行为需明确(显示"No additional actions"或直接不打开) |

## 建议任务分解

1. **TASK-BR039-A:Action Bar widget + 触发键 + 状态**
   - `widget/actionbar/` 子包 + `state.ActionBar` + `keyboard/actionbar_keys.go`
   - **预估**:小模型独立
2. **TASK-BR039-B:每页动作列表**
   - 容器 / 镜像 / 卷 / 网络 / Compose 页 + 通用动作列表
   - **预估**:小模型辅助,主模型审内容完整性
3. **TASK-BR039-C:取消 Q + Help 同步**
   - `registry.go` 移除 KeyQ + Help 页更新
   - **预估**:小模型机械
4. **TASK-BR039-D:i18n + 文档**
   - 所有动作名的 i18n key + navigation.md 同步
   - **预估**:小模型机械

## 待确认项

- [ ] **触发键**:用户说"不占用已有快捷键",建议候选:`;`(vim 风格)/ `:` 扩展 / 新增 `a` / `\` 等。需主模型确认
- [ ] **动作列表来源**:硬编码 vs 从 `ActionRegistry` + `keymap` 动态生成。推荐动态生成
- [ ] **多 tab**:Action Bar 是否支持多 tab(同时挂多个 panel 的动作)?建议不,避免复杂度
- [ ] **空页面行为**:detail / logs 触发后是空 list + "No actions" 还是直接 no-op?建议空 list
- [ ] **Q 键重映射提示**:Q 按了是提示"按 ; 打开 Action Bar"还是 no-op?用户已确认改为 Action Bar,故 Q 应有引导反馈
- [ ] **Quit 在 Action Bar 内的位置**:Quit 列在最后,与 Ctrl+C 双段确认逻辑兼容

## 与其他 BR 的协同

- **BR-033 (TASK-019 高级动作)**:这些动作在 Action Bar 中列出,即使没快捷键也能触发
- **BR-036 (Image Import)**:Image Import 通过 Action Bar 触发,不直接绑键
- **BR-037 (Registry Login)**:Registry Login 通过 Action Bar 触发
- **BR-040 (Dialog 统一)**:Action Bar 是 dialog 的一种,应走 `CenterOnPanel` 统一接口

## 实现结果

- `Q` 已解除退出绑定,普通模式下静默无操作;`Ctrl+C` 保持直接退出。
- `;` 打开当前 panel 的 Action Bar,支持 j/k、方向键、Enter、Esc、1-9 与 `/` 过滤。
- `internal/tui/actionbar` typed registry 当前提供 Containers 的 Rename、Top、Port;其它 panel 等待各自复杂操作落地后接入。
- 全局动作与已有直接快捷键动作不进入 Action Bar;无快捷键动作显示 `—`。
- Action Bar 使用 BR-040 的 panel body 几何,在标准和紧凑布局中均限制在 panel 内。
- Footer / Help 显示 `; Action Bar`,不再显示 `Q Quit`;默认 keymap 改为 `actionBar: [";"]`、`quit: ["ctrl+c"]`。
- BR-033、BR-035、BR-036、BR-037 尚未实现的业务动作由各自任务完成后追加到 typed registry,不属于 BR-039 基础设施缺口。
