# 选择框设计 — 删除/覆盖类操作的确认 UI

> **状态**: 设计稿已确认(2026-08-05)— 用户已决策采用 Option B(Form 组件模式)
> **关联**: docs/constraint/C01-form.md §待确认项 #4(覆盖确认:FORCE 是 Form 内嵌字段还是 ConfirmState 模式) — 已收敛到 Form 内嵌字段方案
> **关联**: docs/discussion/batch-operations-design.md
> **关联 explore**: `bg_bacd2ebb` 调查 + `bg_99c13c2e` 已完成的 ActionBar 修复残留(working tree,未 commit)

---

## 0. 用户决策(2026-08-05 第二轮)

- **采用 Option B(Form 组件模式)** — 不再保留 Option A/C 的探索
- **Force 独立 FormBool 字段**,不与其他选项打包
- **暴露所有 SDK 参数**(Force / RemoveVolumes / RemoveLinks / PruneChildren / Platforms 等)作为 FormBool / FormSelect / FormText 字段,**默认值 false**
- **Cancel / Delete 是 Form 自带按钮**,不是字段选项
- **类似 Form 组件** — 与 Update Resources / Copy / Export / Commit / Save / Load 6 个 Form 风格统一
- C01-form.md §待确认项 #4 随之收敛到 Form 内嵌字段方案

---

## 0bis. 视觉形态参考

镜像删除 Form(示意):
```
┌─ Remove Image ──────────────────────────┐
│                                         │
│   [ ] Force                             │
│         remove even if containers       │
│         reference this image            │
│                                         │
│   [ ] Prune children                    │
│         also delete untagged parents    │
│                                         │
│   [ ] Platforms (advanced)              │
│         linux/amd64, linux/arm64        │
│                                         │
│   [Cancel]                  [Delete]    │
└─────────────────────────────────────────┘
```

容器删除 Form(示意):
```
┌─ Remove Container ───────────────────────┐
│                                         │
│   [ ] Force                             │
│   [ ] Remove anonymous volumes          │
│   [ ] Remove links                      │
│                                         │
│   [Cancel]                  [Remove]     │
└─────────────────────────────────────────┘
```

卷删除 Form(示意):
```
┌─ Remove Volume ──────────────────────────┐
│                                         │
│   [ ] Force                             │
│                                         │
│   [Cancel]                  [Remove]    │
└─────────────────────────────────────────┘
```

Network Remove 不进 Form(SDK 无参数),保持 2 选项 ChoiceDialog `[Cancel] [Delete]`。

---

## 1. 现状与背景

### 1.1 用户报告

> 删除镜像只有两种情况(取消,删除),实际上应该是三种:取消,删除,强制删除。
> 页面实现应该是"取消,删除(Force 强制删除)"。统一使用 Form 表单组件。

### 1.2 全部删除/覆盖类操作的现状

来自 explore agent `bg_bacd2ebb` 的全量调查(Docker SDK v28.5.2):

| 操作 | SDK 是否支持 Force | 当前 UI 类型 | 是否暴露 Force 选项 | 备注 |
|---|---|---|---|---|
| Container Remove(单) | ✅ `container.RemoveOptions{Force, RemoveVolumes, RemoveLinks}` | ChoiceDialog(2 选项) | ❌ 硬编码 `true`(`mark_action.go:86`) | 还支持 `RemoveVolumes`/`RemoveLinks`,当前被丢弃 |
| **Image Remove(单)** | ✅ `image.RemoveOptions{Force, PruneChildren, Platforms}` | ChoiceDialog(2 选项) | ❌ 硬编码 `true`(`mark_action.go:88`)| 还支持 `PruneChildren`,当前 `false`(零值)= 保留未打 tag 父镜像 |
| Volume Remove(单) | ✅ `cli.VolumeRemove(ctx, id, force bool)`(API ≥1.25) | ChoiceDialog(2 选项) | ❌ 硬编码 `true`(`mark_action.go:90`) | — |
| Network Remove(单) | ❌ SDK 无 Force | ChoiceDialog(2 选项) | n/a | 故意不加 Force |
| Bulk Delete(任意面板) | varies | ChoiceDialog(3 选项 Cancel/Force 或 Cancel/Delete) | ✅(`mark_action.go:56-64`) | 是当前唯一正确模式 |
| Compose `down` | varies | 无 confirm(静默执行) | ❌ 硬编码 `true` | 容器/卷 force,网络不 force |

**关键 bug 摘要**:
- 所有**单条**删除都是 2 选项 ChoiceDialog,且 `doConfirmYes` 中 `Force` 写死为 `true`
- 用户无法选择是否 Force
- Image Remove 还丢失 `PruneChildren` 控制权
- Container Remove 还丢失 `RemoveVolumes`/`RemoveLinks` 控制权

### 1.3 两个组件的能力对比

#### A. ChoiceDialog(当前)

```go
// state/confirm.go:22-29
s.Options = []ChoiceOption{
    {ID: "cancel", Label: "Cancel"},
    {ID: "confirm", Label: "Confirm"},
}
```

- 简单、平铺的选项列表
- 渲染: `widget/dialog/choice.go:18-46` → `[Cancel]   [▶ Confirm]` 横向布局
- 交互: Tab/Shift+Tab 切换,Enter 提交,Esc 取消
- 支持 Description(辅助描述,例如 `bulkDeleteOptions` 的 `"delete even if active or in use"`)
- 没有 FormField 概念,直接是按钮列表
- 多用于"纯确认 + 可选 Force"

#### B. Form 组件(已用于 6 种"复杂操作")

现有 FormKind(`state/form.go:14-22`):
- `FormContainerCopy`, `FormContainerUpdate`, `FormContainerExport`, `FormContainerCommit`
- `FormImageSave`, `FormImageLoad`

**没有** `FormImageRemove` / `FormContainerRemove` / `FormVolumeRemove`。

FormField 能力(`state/form.go:71-133`):
- `Kind`: FormText, FormInt, FormSelect, FormMultiSelect, FormBool, FormPath
- `Options []string` + `DisplayOptions []string`(`Options` 是 wire key,`DisplayOptions` 是本地化标签)
- `HelperText`, `Unit`, `Placeholder`, `Min/Max *float64`, `Required`, `DependsOn/DependsEq`, `Hidden`
- `Index int`(FormSelect 当前选中)
- `Touched bool` — 提交时区分"用户动过 vs 默认"

FormDialog 按钮行(`widget/dialog/form.go:49-56`):
- 内置 Cancel(左)+ Confirm(右)按钮
- 已经在 Submit/Cancel 路径覆盖

参考实现 — Restart policy(`container_form.go:45-84`):
```go
const restartPolicyUnchanged = "unchanged"

var restartPolicyChoices = []restartChoice{
    {Key: "unchanged",         Label: i18n.T("container.update.form.restart.unchanged")},
    {Key: "no",                Label: i18n.T("container.update.form.restart.no")},
    {Key: "always",            Label: i18n.T("container.update.form.restart.always")},
    {Key: "unless-stopped",    Label: i18n.T("container.update.form.restart.unless_stopped")},
    {Key: "on-failure",        Label: i18n.T("container.update.form.restart.on_failure")},
}

func buildRestartField() state.FormField {
    options := make([]string, len(restartPolicyChoices))
    display := make([]string, len(restartPolicyChoices))
    for i, c := range restartPolicyChoices {
        options[i] = c.Key      // wire value
        display[i] = c.Label    // localized display
    }
    return state.FormField{
        Key:            fieldRestartPolicy,
        Label:          i18n.T("container.update.form.restart"),
        Kind:           state.FormSelect,
        Options:        options,
        DisplayOptions: display,
    }
}
```

---

## 2. 设计选项对比

### Option A — ChoiceDialog 3 选项模式(最小改动) — ❌ 不再采纳

~~**形态**:`[ Cancel ]  [ Delete ]  [ Force delete ]` 横向 3 按钮~~

> **已废弃** — 用户已确认采用 Option B,此方案不再考虑。保留作为对比参考。

**形态**:`[ Cancel ]  [ Delete ]  [ Force delete ]` 横向 3 按钮

#### 改动范围

1. `helpers.go:264-268` `confirmAction` 抽出 `confirmActionWithOptions(m, action, target, message, opts)` 重载
2. 每个删除触发点传自定义 Options:
   ```go
   confirmActionWithOptions(m, keys.ShowImageRemove, img.ID, msg, []state.ChoiceOption{
       {ID: keys.ShowOptionCancel,  Label: i18n.T("key.cancel")},
       {ID: keys.ShowOptionConfirm, Label: i18n.T("image.remove.confirm"),         Description: i18n.T("image.remove.confirm_desc")},
       {ID: keys.ShowOptionForce,   Label: i18n.T("image.remove.confirm_force"),   Description: i18n.T("image.remove.confirm_force_desc")},
   })
   ```
3. `confirm.go:handleConfirmKeys` 扩展 `ShowOptionForce` switch:
   ```go
   case keys.ShowOptionForce:
       switch m.Confirm.ConfirmAction {
       case keys.ShowBulkDelete:      m.Confirm.ConfirmAction = keys.ShowBulkDeleteForce
       case keys.ShowImageRemove:     m.Confirm.ConfirmAction = keys.ShowImageRemoveForce
       case keys.ShowContainerRemove: m.Confirm.ConfirmAction = keys.ShowContainerRemoveForce
       case keys.ShowVolumeRemove:    m.Confirm.ConfirmAction = keys.ShowVolumeRemoveForce
       }
       return doConfirmYes(m)
   ```
4. `mark_action.go:doConfirmYes` 增加 Force 变体 action:
   ```go
   case action == keys.ShowImageRemove, action == keys.ShowImageRemoveForce:
       force := action == keys.ShowImageRemoveForce
       return m, withImageAudit(imageRemoveCmd(m.Connection.Engine, target, force), trace)
   ```

#### 优点
- 改动小(4 文件,~80 行)
- 与 Bulk Delete 现有 3 选项模式一致
- 视觉上"Cancel 与 Delete/Force 同级",符合用户描述的"3 种情况"
- 不引入新 FormKind
- 不破坏 Podman 驱动(`pruneChildren` 等暂不暴露)

#### 缺点
- 仅支持单一 Force 开关
- 无法扩展为 `PruneChildren` / `RemoveVolumes` 等多参数(用户说"同样参数不包括格外参数如 Force"暗示未来可能加 PruneChildren 等)
- 与现有 6 个 Form 操作(Copy/Update/Export/Commit/Save/Load)模式不一致

### Option B — Form 组件模式(对齐现有复杂操作)— ⭐ 已确认采用

**形态**: FormDialog 顶部多个 FormBool / FormSelect / FormText 字段(Force、RemoveVolumes、RemoveLinks、PruneChildren、Platforms 等),底部 Form 自带的 Cancel / Confirm 按钮。

**核心原则**:
- **Force 是独立 FormBool 字段**,不是与其他选项打包的 "Delete / Force Delete" 单选
- **暴露所有 SDK 参数** — 每个 bool 参数一个 FormBool,默认 false(与 Docker CLI 默认一致)
- **Cancel / Delete 是 Form 自带的按钮**(FormDialog 已内置),不作为 FormField

#### 改动范围

1. `state/form.go` 新增 FormKind:
   ```go
   const (
       ...
       FormImageRemove
       FormContainerRemove
       FormVolumeRemove
   )
   ```
2. `keyboard/image_action.go` 新增 `openImageRemoveForm(m, img)`(镜像 `container_form.go:buildRestartField` 模式)
3. `keyboard/container_form.go` `submitContainerForm` 新增 `case state.FormImageRemove:` 分支,读 `m.Form.Get(fieldImageRemoveMode).Option() == "force"`
4. i18n 新增 keys(参考 restart-policy):
   - `image.remove.form.title`
   - `image.remove.form.mode`
   - `image.remove.form.mode.default`
   - `image.remove.form.mode.default_desc`
   - `image.remove.form.mode.force`
   - `image.remove.form.mode.force_desc`
   - 平行 `container.remove.form.*`, `volume.remove.form.*`
5. Podman 驱动扩展 `prune-children`(若要暴露 ImageRemove.PruneChildren,见下)

#### 优点
- 与 6 个现有 Form 操作(Copy/Update/Export/Commit/Save/Load)风格一致
- 同一 Form 可叠加 `FormBool PruneChildren` / `FormBool RemoveVolumes` 等,无需重建 UI
- 未来加参数不用动对话框结构,只加 FormField 即可
- 触发 PathSaveFile 时可复用覆盖确认逻辑(C01-form.md §待确认项 #4 关联)

#### 缺点
- 改动较大(5-6 文件,~200-300 行)
- 新增 FormKind、submit 分支、submitContainerForm 已很大
- 用户描述"3 种情况"在视觉上是 3 个同级按钮,FormSelect 是 1 个下拉 + 2 个 Form 按钮 — 形式不直接等同
- "Cancel 是 Form 按钮,Delete/Force 是字段"语义上不对等

### Option C — 折中(本期最小 + 留接口) — ❌ 不再采纳

~~**形态**: 先 Option A(3 选项 ChoiceDialog)修复当前 bug;Form 模式留待 C01-form.md §待确认项 #4 决议后做。~~

> **已废弃** — 用户已确认直接采用 Option B。

#### 优点
- 本期工作量小
- 不影响 C01-form.md 待定结论
- 现有 Bulk Delete 模式已是 3 选项 ChoiceDialog,扩到单条删除最自然

#### 缺点
- 仍然无法扩展 PruneChildren 等多参数(若用户后续提)

---

## 3. Option B vs 其他方案对比

| 维度 | B: Form 组件模式(已确认) | A: 3 选项 ChoiceDialog(已废弃) |
|---|---|---|
| 改动行数 | ~200-300 | ~80 |
| 视觉形态 | FormBool 字段 + Form 按钮 | 3 个同级按钮 |
| 与现有 Form 操作一致性 | ✅ 完全一致 | ❌ |
| 暴露 SDK 参数 | ✅ 全部(Force/RemoveVolumes/RemoveLinks/PruneChildren/Platforms) | ❌ 仅 Force |
| 可扩展性 | ✅ 加参数不改 UI | ❌ |
| 与 Bulk Delete 模式一致性 | N/A(各操作独立 Form) | ✅ 相似 |
| 工作量(人时) | 2-4 | 0.5-1 |
| 是否解决 C01-form.md §待确认项 #4 | ✅ 收敛到 Form | ❌ |

---

## 4. 操作清单与现状对照表

### 4.1 受影响的删除/覆盖操作

| 操作 | 现状 | 推荐迁移 | 备注 |
|---|---|---|---|
| Container Remove(单) | 2 选项,hardcoded true | A 或 B | 还有 RemoveVolumes/RemoveLinks 待暴露 |
| Image Remove(单) | 2 选项,hardcoded true | A 或 B(用户报告 case)| 还有 PruneChildren/Platforms 待暴露 |
| Volume Remove(单) | 2 选项,hardcoded true | A 或 B | — |
| Network Remove(单) | 2 选项,SDK 无 force | 保持 2 选项 | 注意:不要加 Force(SDK 忽略) |
| Bulk Delete | 已 3 选项 | 不变 | Container/Image/Volume 用 Force,Network 用 Delete |
| Compose `down` | 无 confirm,hardcoded true | 增加 ChoiceDialog(同 Bulk Delete 模式) | — |

### 4.2 字段暴露策略(Option B 确认)

| 操作 | SDK 字段 | 暴露字段(均 FormBool 默认 false) | 备注 |
|---|---|---|---|
| Container Remove | `Force`, `RemoveVolumes`, `RemoveLinks` | 3 个 FormBool 全暴露 | Force 默认 false — 普通删除对运行中容器会拒绝,需 Force |
| Image Remove | `Force`, `PruneChildren`, `Platforms` | 3 个字段全暴露 | `Platforms` 是 FormText 而非 FormBool(逗号分隔列表) |
| Volume Remove | `force bool`(pos) | 1 个 FormBool | API ≥1.25 需要 |
| Network Remove | n/a | — | 保持 2 选项 ChoiceDialog |

---

## 5. i18n key 候选(若选 B)

镜像删除(参考 restart-policy 模式):
```
image.remove.form.title:            Remove image
image.remove.form.mode:             Mode
image.remove.form.mode.default:     Delete
image.remove.form.mode.default_desc: remove if no containers are using this image
image.remove.form.mode.force:       Force delete
image.remove.form.mode.force_desc:  remove even if containers reference this image
```

容器删除:
```
container.remove.form.title:        Remove container
container.remove.form.mode:         Mode
container.remove.form.mode.default: Delete
container.remove.form.mode.default_desc: stop and remove (refuses if running)
container.remove.form.mode.force:     Force delete
container.remove.form.mode.force_desc: SIGKILL and remove immediately
container.remove.form.volumes:      Also remove anonymous volumes
container.remove.form.links:        Also remove links
```

卷删除:
```
volume.remove.form.title:           Remove volume
volume.remove.form.mode:            Mode
volume.remove.form.mode.default:    Delete
volume.remove.form.mode.default_desc: remove if no containers reference this volume
volume.remove.form.mode.force:      Force delete
volume.remove.form.mode.force_desc: remove even if containers reference this volume
```

---

## 6. 运行时契约扩展(若选 B 并要暴露多参数)

当前 `runtimeapi.LifecycleOptions`(`data/runtime/actions.go:62-71`):
```go
type LifecycleOptions struct {
    Force   bool
    Name    string
    Signal  string
    Timeout time.Duration
}
```

要支持 `PruneChildren` / `RemoveVolumes` / `RemoveLinks`,两个选择:
- (a) 给 `LifecycleOptions` 加 bool 字段(简单,但膨胀)
- (b) 引入新 `RemoveOptions struct { Force bool; PruneChildren bool; RemoveVolumes bool; RemoveLinks bool }` 平铺到 `runtimeapi.ActionOptions`(更结构化)

Podman 驱动需同步:`internal/driver/podman/images_actions.go:18-23` 当前仅 `force`,若 B + 多参数,Podman 端需扩展 `query{}`。

---

## 7. 用户已决策 + 剩余开放问题

### 已决策
- ✅ Option B(Form 组件模式)
- ✅ Force 独立 FormBool,默认 false
- ✅ 暴露所有 SDK 参数,默认 false
- ✅ Cancel / Delete 用 Form 自带按钮
- ✅ Network Remove 保持 2 选项 ChoiceDialog(SDK 无 Force)

### 剩余开放
1. **Platforms 字段是否本期暴露?** (Image Remove) — 默认空,advanced 选项,可能延后
2. **Compose `down` 是否一并迁移?** (当前是 hardcoded true,无 confirm) — 建议同期改造
3. **PruneChildren 字段是否同期扩 Podman driver?** (若 Podman REST 支持) — 待查
4. **i18n 文案是否需要新增中文/日文翻译** — 建议同期完成

---

## 8. 后续行动(待用户决定后)

1. 根据用户选择(A / B / C)实施
2. 新增 i18n keys 并提交 en/zh/ja
3. 更新 Podman 驱动若选 B + 多参数
4. 单 commit:`fix(dialog): delete operations 选择框 — Force 选项 + Force 暴露` (按用户偏好)
5. PTY light 主题视觉验证