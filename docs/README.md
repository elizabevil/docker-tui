# dtui 文档索引

`dtui` 是一个面向 Docker / Podman 的终端管理工具,代码入口在 `cmd/docker-tui/main.go`,CLI `Use` 名称当前为 `dtui`。

## 最小运行

```bash
go build -o ./dist/dtui ./cmd/docker-tui
./dist/dtui
```

如果需要覆盖嵌入版本号,可以设置 `DTUI_VERSION`,例如:

```bash
DTUI_VERSION=v0.2.0 go build -o ./dist/dtui ./cmd/docker-tui
```

指定配置或主机:

```bash
./dist/dtui --config ~/.config/docker-tui/config.yml --host unix:///var/run/docker.sock
```

如果使用仓库内置任务,`just build` 和 `just run` 也都基于 `./dist/dtui`。发布 workflow 在 tag `v*` 上会生成 Linux / macOS / Windows 产物。程序内部 `cobra.Use` 名称仍是 `dtui`。

## 先看哪里

按 **UI 设计 / 功能设计 / 修复流程 / 实现参考 / 用户参考** 分类组织。

### 设计与目标

- [feature-design.md](feature-design.md):**功能设计总览** — 最终实现目标与原生 Docker / Podman / 竞品的功能对比矩阵;设计决策与路线;明确不在范围的能力。
- [feature-todo-list.md](feature-todo-list.md):**功能实现待办清单** — TASK 历史台账 + 设计驱动需求摘要 + 已取消决议;新任务不再使用 TASK-xxx 编号,统一改用 BR-xxx / FR-xxx。

### 需求真理源

- [requirement/README.md](requirement/README.md):**需求树** — 大需求(R##)与子需求(R##-##)的索引;每个需求文档包含目标、用户流程、UI/UX、功能规则、验收标准、迁移记录。
- [requirement/R01-container/](requirement/R01-container/README.md):容器大需求(高级动作、Exec shell、Stats/Top/Wait)。
- [requirement/R02-image/](requirement/R02-image/README.md):镜像大需求(History 顶层页、Import tarball、Registry Login)。
- [requirement/R03-form-action/](requirement/R03-form-action/README.md):表单与动作展示(Form 公共模式、Action Bar、Action 展示框布局)。
- [requirement/R04-runtime/](requirement/R04-runtime/README.md):运行时能力(Docker / Podman 差异评估)。
- [requirement/R05-events-network/](requirement/R05-events-network/README.md):事件流与网络(Events 面板、Network Connect/Disconnect)。
- [requirement/R06-operation/](requirement/R06-operation/README.md):Operation 一等公民(横切)。
- [requirement/R07-config-ux/](requirement/R07-config-ux/README.md):配置系统便利化(CLI info / config init / validate / 首次运行提示)。
- [requirement/R08-compose/](requirement/R08-compose/README.md):**Compose 项目级与编排**(2026-08-07 立项) — 聚合模型、ComposeService 候选、生命周期、镜像/日志/查询/执行、Action Bar/Keymap/i18n 集成;旧 [.omo/compose-todo.md](../../.omo/compose-todo.md) 保留作研究底稿。

### 横切约束

- [constraint/README.md](constraint/README.md):**横切约束索引** — 多需求共享的 UI/交互/工程规则。
- [constraint/C01-form.md](constraint/C01-form.md):Form 字段类型、焦点、Tab 补全、条件显示、错误提示。
- [constraint/C02-dialog.md](constraint/C02-dialog.md):Dialog 居中、尺寸、Action 展示框布局。
- [constraint/C03-table.md](constraint/C03-table.md):表格列布局、选中态、滚动。
- [constraint/C04-keybinding.md](constraint/C04-keybinding.md):快捷键分层、Action Bar 角色、链接键。
- [constraint/C05-path.md](constraint/C05-path.md):Local / Container 路径语义、补全、绝对化。
- [constraint/C06-i18n.md](constraint/C06-i18n.md):i18n key 命名空间、终端宽度。

> 需求文档引用约束时只用链接,不复制通用段落。

### 修复流程

- [bugfix-requirements.md](bugfix-requirements.md):**历史 bug 修复需求台账**,单独跟踪修复项;包含修复记录。
- [pending-bugs.md](pending-bugs.md):当前未关闭的 BUG / 需求快照,从 bugfix-requirements.md 同步生成的待办清单。

### 实现参考

- [architecture.md](architecture.md):当前实现架构,已按代码核对。
- [navigation.md](navigation.md):当前默认交互与快捷键。
- [project-structure.md](project-structure.md):目录与包职责,已按代码核对。
- [requirements.md](requirements.md):产品范围和路线图,包含规划项,等于已实现清单。
- [docs-architecture.md](docs-architecture.md):**本文档**架构说明 — 真理源、约束、实施卡的层级关系。

### 用户参考

- [ui-design.md](ui-design.md):UI 设计系统(布局 / 组件 / 表格 / 样式),从 `design/current-design.md` 提取的开发者参考。
- [i18n.md](i18n.md):国际化设计(翻译资源 / 终端宽度 / 语言切换),从 `design/ui-i18n-design.md` 提取的开发者参考。
- [ai-prompts.md](ai-prompts.md):面向当前代码结构的 AI 协作提示。

### 设计原始档案

`design/` 目录保留完整设计原始档案(UI 权威 + 功能/适配 + 修复流程 + 已归档),见 [../design/README.md](../design/README.md)。

## 设计文档

- [../design/current-design.md](../design/current-design.md):当前唯一权威 UI 设计文档
- [../design/bugfix-design.md](../design/bugfix-design.md):历史 bug 修复设计与变更约束
- [../design/podman-capabilities-analysis.md](../design/podman-capabilities-analysis.md):Docker / Podman 能力差异分析
- [../design/ui-i18n-design.md](../design/ui-i18n-design.md):UI 国际化详细设计
- [../design/README.md](../design/README.md):设计文档入口

## 当前代码验证结论

- 主程序入口:`cmd/docker-tui/main.go`
- 默认配置文件:`~/.config/docker-tui/config.yml`
- 资源面板:`containers`、`images`、`volumes`、`networks`、`compose`
- 覆盖层/模式:`logs`、`detail`、`help`、`filter`、`command`、`exec`
- 运行时支持:Docker / Podman,启动时通过连接池依次尝试本地 Podman 和 Docker
- 主题与语言:内置主题加载与 `zh` / `en` i18n 已接入
- 快捷键自定义:`keymap.*` 已接入统一动作注册表,Help 与 Footer 显示当前有效绑定
- 用户操作审计:资源操作使用统一 trace,终态投影到通知与 Footer,并写入按日 JSONL

## 文档约定

- "实现现状"以代码为准。
- "设计意图"以 `design/current-design.md`(UI) / `docs/feature-design.md`(功能) 为准。
- "需求/路线图"允许包含尚未实现的能力,但应与实现文档分开表述(`docs/requirements.md`)。
- "历史 bug 修复需求"统一记录在 `bugfix-requirements.md`,不要混入路线图。
- "待办"分两层:`pending-bugs.md`(BUG 级)与 `feature-todo-list.md`(功能级)。
- 设计已排除 / 取消的能力,统一记录在 `feature-todo-list.md §3` 与 `feature-design.md §3`,不要在新需求中重新提出。