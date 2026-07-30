# dtui 文档索引

`dtui` 是一个面向 Docker / Podman 的终端管理工具，代码入口在 `cmd/docker-tui/main.go`，CLI `Use` 名称当前为 `dtui`。

## 最小运行

```bash
go build -o ./dist/dtui ./cmd/docker-tui
./dist/dtui
```

如果需要覆盖嵌入版本号，可以设置 `DTUI_VERSION`，例如：

```bash
DTUI_VERSION=v0.2.0 go build -o ./dist/dtui ./cmd/docker-tui
```

指定配置或主机：

```bash
./dist/dtui --config ~/.config/docker-tui/config.yml --host unix:///var/run/docker.sock
```

如果使用仓库内置任务，`just build` 和 `just run` 也都基于 `./dist/dtui`。发布 workflow 在 tag `v*` 上会生成 Linux / macOS / Windows 产物。程序内部 `cobra.Use` 名称仍是 `dtui`。

## 先看哪里

- [architecture.md](architecture.md)：当前实现架构，已按代码核对
- [project-structure.md](project-structure.md)：目录与包职责，已按代码核对
- [navigation.md](navigation.md)：当前默认交互与快捷键
- [ui-design.md](ui-design.md)：UI 设计系统（布局/组件/表格/样式），从 design/current-design.md 提取的开发者参考
- [i18n.md](i18n.md)：国际化设计（翻译资源/终端宽度/语言切换），从 design/ui-i18n-design.md 提取的开发者参考
- [requirements.md](requirements.md)：产品范围和路线图，包含规划项，不等于已实现清单
- [bugfix-requirements.md](bugfix-requirements.md)：历史 bug 修复需求台账，单独跟踪修复项
- [ai-prompts.md](ai-prompts.md)：面向当前代码结构的 AI 协作提示

## 设计文档

- [../design/current-design.md](../design/current-design.md)：当前唯一权威 UI 设计文档
- [../design/bugfix-design.md](../design/bugfix-design.md)：历史 bug 修复设计与变更约束
- [../design/README.md](../design/README.md)：设计文档入口
- [ui-preview/README.md](ui-preview/README.md)：浏览器预览系统说明

## 当前代码验证结论

- 主程序入口：`cmd/docker-tui/main.go`
- 默认配置文件：`~/.config/docker-tui/config.yml`
- 资源面板：`containers`、`images`、`volumes`、`networks`、`compose`
- 覆盖层/模式：`logs`、`detail`、`help`、`filter`、`command`、`exec`
- 运行时支持：Docker / Podman，启动时通过连接池依次尝试本地 Podman 和 Docker
- 主题与语言：内置主题加载与 `zh` / `en` i18n 已接入
- 快捷键自定义：`keymap.*` 已接入统一动作注册表，Help 与 Footer 显示当前有效绑定
- 用户操作审计：资源操作使用统一 trace，终态投影到通知与 Footer，并写入按日 JSONL

## 文档约定

- “实现现状”以代码为准。
- “设计意图”以 `design/current-design.md` 为准。
- “需求/路线图”允许包含尚未实现的能力，但应与实现文档分开表述。
- “历史 bug 修复需求”统一记录在 `bugfix-requirements.md`，不要混入路线图。
