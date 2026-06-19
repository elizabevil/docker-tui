# docker-tui

k9s 风格的 Docker 终端管理工具。键盘驱动、实时刷新、全量资源覆盖。

## 快速开始

```bash
# 构建
go build -o docker-tui ./cmd/docker-tui

# 运行（需要 Docker daemon 访问权限）
./docker-tui

# 指定配置和 Docker 主机
./docker-tui --config ~/.config/docker-tui/config.yml --host tcp://192.168.1.100:2375
```

## 项目定位

| 维度 | 说明 |
|---|---|
| 产品形态 | All-in-one 终端管理台，对标 Portainer 的 TUI 版 |
| 覆盖范围 | 容器 / 镜像 / Volume / Network / Compose |
| 目标用户 | 开发者 & SRE，日常 Docker 管理重度用户 |
| 体验对标 | k9s（键盘驱动、实时刷新、多面板、可定制） |

## 核心依赖

| 用途 | 库 | 版本 |
|---|---|---|
| TUI 框架 | `charm.land/bubbletea/v2` | v2.0.7 |
| 组件库 | `charm.land/bubbles/v2` | v2.1.0 |
| 样式 | `charm.land/lipgloss/v2` | v2.0.3 |
| Docker SDK | `github.com/docker/docker/client` | latest |
| CLI 框架 | `github.com/spf13/cobra` | v1.10.2 |
| 配置解析 | `gopkg.in/yaml.v3` | latest |

## 详细文档

- [需求规格](requirements.md) — 功能矩阵、竞品对比、MVP 分期
- [架构设计](architecture.md) — 技术选型、分层架构、数据流
- [项目结构](project-structure.md) — 代码地图、包职责
- [AI 提示指南](ai-prompts.md) — 团队 AI 协作的 prompt 模板
