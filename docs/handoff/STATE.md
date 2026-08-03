# STATE — task 任务驱动 / 当前活跃批次

> 上次更新: 2026-08-03
> 当前轮次: 1

## 批次 R1:文档真理源重组与清理

### R1.1 — 批次元信息

- **状态**:done + 已 commit(主模型审查完成,2026-08-03)
- **日期**:2026-08-02 ~ 2026-08-03
- **范围**:`docs/requirement/`(新建)、`docs/constraint/`(新建)、`docs/docs-architecture.md`(新建)、`docs/README.md` / `docs/requirements.md` / `docs/task/README.md`(更新入口)
- **不可触碰边界**:Go 代码、`internal/data/i18n/lang/*`、现有 task/ 文档(除非明确删除)
- **主模型决策**:见 [GLOBAL.md § 2026-08-02](./GLOBAL.md#2026-08-02--文档真理源重组)

### R1.2 — 子任务清单

| 子任务 | 状态 | 范围文件 | 提交者 |
|---|---|---|---|
| §8.1 目录创建 | done | `mkdir docs/requirement/R0{1..5} docs/constraint` | 小模型 |
| §8.2 README 索引(7 个) | done | `requirement/README.md` + 5 R## README + `constraint/README.md` | 小模型 |
| §8.3 constraint 文档(6 个) | done | `constraint/C0{1..6}.md` | 小模型 |
| §8.4 requirement 文档(11 个) | done | `requirement/R##-##/*.md` | 小模型 |
| §8.5 入口索引更新(3 个) | done | `docs/README.md` / `docs/requirements.md` / `docs/task/README.md` | 小模型 |
| Findings 修复(链接 + 事实) | done | R## README + R01-03 + docs/requirements.md:14 + 6 个 constraint | 小模型 |
| 6 个冗余 task 删除 | done | `rm` 6 文件 + 13 文件反向引用清理 | 小模型 |

### R1.3 — 验证

- `go build ./...`:未运行(纯文档操作,不应影响 Go 代码)
- `go test -short ./...`:未运行(同上)
- `git diff --check`:通过
- 主模型审查(2026-08-03)补充修复:
  - 子模型声明"残留引用 0"不实 — 实查发现 5 处指向已删 task 的悬空引用(`docs/task/README.md` ×5、br-039-a、br-039-code-location ×2、br-040-code-location ×2、br-041 ×1),已全部修复
  - 跨组断链 13 处:`R03-form-action/` 路径缺失 ×5、`docs/feature-design.md` 前缀 ×6、R04 `internal/data/runtime/` 相对深度 ×2,已修复
  - 全仓 markdown 链接校验:479 个内部链接全部可解析(109 个 `internal/*` 为 bugfix 文档预存代码锚点约定,非断链)
- 旧 task 文件保留:21 个(去 6 后)

### R1.4 — 主模型审查结论

- 全部新增 30 个 + 修改 13 个 + 删除 6 个文档已 review 并 commit
- 审查发现与修复记录:见上 R1.3;决策记录见 [GLOBAL.md § 2026-08-03 R1 批次审查与 commit](./GLOBAL.md#2026-08-03--r1-批次主模型审查与-commit)
- R04-01 状态 `implementing` 偏乐观,待 R4 批次实施时再校准(不阻塞 R1)

## 批次 R2:待启动

### R2.1 — 候选:BR-041 followup 实施(6 项未实施项)

- 详见 [task/br-043-form-action-bar-redesign.md](../task/br-043-form-action-bar-redesign.md)
- 路径绝对化 / 尺寸统一 / 条件字段 / Action 展示框重构 / 配置拆分 / R/C 链接键决议
- 子模型待命前需主模型决策:哪些批次进 R2,优先级如何

## 公共范围边界(所有子任务不可触碰)

- `internal/**/*.go`
- `internal/data/i18n/lang/*.jsonc`
- `internal/data/config/*.jsonc`
- 现有 task/ 文档(除非在 STATE.md 显式列入本批次范围)

## 全局链接

- [../docs-architecture.md](../docs-architecture.md) — 文档真理源架构
- [../requirement/](../requirement/) — 需求真理源
- [../constraint/](../constraint/) — 横切约束
- [./GLOBAL.md](./GLOBAL.md) — 全局决策日志
- [./COMPLETED.md](./COMPLETED.md) — 已交接历史