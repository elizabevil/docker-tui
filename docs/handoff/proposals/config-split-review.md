# 待住模型审查：BR-043 §3.4 配置拆分方案

> 提交日期: 2026-08-03
> 提交人: 子模型（执行）
> 关联 commit: `b79662b`（已实施基础拆分）
> 关联决策: GLOBAL.md 2026-08-03 §3.4 配置拆分实施
> 关联 R03: `R03-04-path-popup.md`（设计文档）
> 关联 BR: BR-043 §3.4

## 1. 已实施现状

`internal/data/config/default.jsonc`（189 行单文件）已拆分为 `internal/data/config/styles/*.jsonc` 8 个嵌入分片：

```
general.jsonc / ui.jsonc / docker.jsonc / runtime.jsonc
logs.jsonc / layout.jsonc / commandTemplates.jsonc / keymap.jsonc
```

加载机制（`internal/data/config/loader.go`）：
1. `//go:embed styles/*.jsonc` → `StylesFS() fs.FS`
2. `LoadSplit(userDir)` 按字母序加载嵌入分片
3. 用户覆盖 `~/.config/docker-tui/styles/*.jsonc` 同结构叠加
4. 合并 `map[string]json.RawMessage` → unmarshal 到 `Config` → `Validate()`
5. `config.go Load` 改用 `LoadSplit` 作为底座，再叠 CLI YAML `config.yml`

测试：`TestLoadSplit*` 5/5 pass + 全包无回归。

## 2. 5 项设计权衡（待决策）

### 权衡 1：深合并 vs 浅合并

**当前**：浅合并（顶层 section 粒度）。

**问题**：用户文件改 `runtime.health.intervalSec` 时若只写 `{"runtime":{"health":{"intervalSec":5}}}`，未指定字段被零值覆盖，`runtime.default = ""` 违反 `Validate()`。

**推荐**：**A. 字段级深合并**（JSON → `map[string]any` → 递归合并；数组按覆盖/追加策略）。

**影响**：`loader.go` 替换 `mergedToConfig`；增加数组处理逻辑；测试覆盖字段级合并。

### 权衡 2：YAML + JSONC 双格式共存

**当前**：YAML 走 `config.yml`，JSONC 走 `styles/*.jsonc`，同时生效。

**推荐**：**保持现状**。JSONC 用于细粒度字段级覆盖，YAML 用于整体配置（如连接列表）。语义互补不冲突。

**影响**：无。

### 权衡 3：路径不一致

**当前**：
- 应用配置：`internal/data/config/styles/*.jsonc`（data 包）
- Dialog 配置：`internal/tui/ui/widget/dialog/dialog.jsonc`（UI 包）
- 主题：`internal/data/config/themes/*.jsonc`（data 包）

**推荐**：**将 Dialog 配置搬到 `internal/data/config/styles/dialog.jsonc`**，统一"配置/样式"到一处。

**影响**：
- `dialog.jsonc` 移位
- `view.go` 改用 `config.StylesFS()` + `utils.LoadFromString`
- 加载逻辑加 dialog section（可选 — dialog 仍可走独立 `LoadDialogConfig` 路径）

### 权衡 4：主题独立目录

**当前**：`themes/` 与 `styles/` 独立，前者用 `ThemeFS()`，不参与 `LoadSplit`。

**推荐**：**保持现状**。主题是运行时切换，配置是启动期合并，语义不同。**仅文档说明**：在 R03-04 或 README 注明两个目录的用途差异。

**影响**：无。

### 权衡 5：配置版本演进

**当前**：merged map 强制注入 `configVersion = CurrentConfigVersion`，无 migration 机制。

**推荐**：**保持现状**。当前 v1 无兼容负担；将来 bump 到 v2 时再加 migration（`v1ToV2()` 函数）。

**影响**：无（当前）。

## 3. 推荐决策汇总

| # | 权衡 | 推荐 | 优先级 |
|---|------|------|--------|
| 1 | 深合并 vs 浅合并 | **A. 字段级深合并** | P1（UX 缺陷） |
| 2 | YAML + JSONC | 保持现状 | — |
| 3 | 路径一致性 | Dialog 搬到 `styles/dialog.jsonc` | P2 |
| 4 | 主题独立 | 保持现状 + 文档说明 | — |
| 5 | 配置版本 | 保持现状 | — |

## 4. 实施计划（若批准 P1+P2）

### 阶段 1：字段级深合并（约 80 行 + 60 行测试）

修改 `loader.go`：

```go
// 替换 mergeJSONCSection
func deepMerge(dst, src map[string]json.RawMessage) error {
    for k, v := range src {
        existing, ok := dst[k]
        if !ok {
            dst[k] = v
            continue
        }
        // both sides are objects → recurse
        var dstObj, srcObj map[string]json.RawMessage
        if err := json.Unmarshal(existing, &dstObj); err == nil {
            if err := json.Unmarshal(v, &srcObj); err == nil {
                if err := deepMerge(dstObj, srcObj); err != nil { return err }
                merged, _ := json.Marshal(dstObj)
                dst[k] = merged
                continue
            }
        }
        // src wins (leaf or type mismatch)
        dst[k] = v
    }
    return nil
}
```

测试：
- `TestLoadSplitDeepMergesRuntimeHealth`（只覆盖 `intervalSec`，其他 runtime 字段保留）
- `TestLoadSplitDeepMergesArrayOverride`（数组整体替换，不追加）

### 阶段 2：Dialog 配置统一（约 30 行 + 10 行测试）

- 移动 `internal/tui/ui/widget/dialog/dialog.jsonc` → `internal/data/config/styles/dialog.jsonc`
- `view.go LoadDialogConfig` 改为读 `config.StylesFS()` 中的 `dialog.jsonc`
- 或保持 Dialog 独立加载（更简单），仅文档说明

## 5. 风险评估

| 风险 | 级别 | 缓解 |
|------|------|------|
| 深合并破坏现有用户覆盖语义 | 中 | 测试覆盖 `TestLoadSplitAppliesPartialOverride`（验证只覆盖 `intervalSec` 不破坏 `default`） |
| Dialog 路径迁移导致加载失败 | 低 | 保持 `LoadDialogConfig` 作为 fallback |
| 性能：深合并递归开销 | 低 | 分片仅 ~8 个 section，单次启动开销 <1ms |

## 6. 待住模型决策

1. ☐ **P1**：是否实施字段级深合并（权衡 1 → 方案 A）？
2. ☐ **P2**：是否将 Dialog 配置搬到 `internal/data/config/styles/`（权衡 3）？
3. ☐ **N/A**：YAML/JSONC 双格式、主题独立、配置版本 — 确认保持现状即可

---

## 7. 审查响应模板

```text
决策 1 (深合并): [ ] A 实施  [ ] B 浅合并  [ ] C 零值填充 + seed  [ ] 维持现状
决策 2 (Dialog 路径): [ ] 搬  [ ] 维持现状 + 文档说明  [ ] 维持现状不改
决策 3 (现状确认):  [ ] 同意
```

---

**审查响应回填后**，子模型将按 §4 计划实施；无需住模型参与代码。