# dtui AI 协作配置

## 一、安装 OpenCode

```bash
# 安装 OpenCode CLI
npm install -g @opencode/cli
# 或
curl -sSf https://opencode.ai/install.sh | sh
```

## 二、项目集成配置

创建 `opencode.json` 在项目根目录：

```json
{
  "version": "1.0",
  "project": {
    "name": "dtui",
    "language": "go",
    "build": "go build -o dtui ./cmd/docker-tui",
    "test": "go test ./...",
    "vet": "go vet ./..."
  },
  "commands": {
    "build": {
      "command": "go build -o dtui ./cmd/docker-tui && echo '✓ dtui built'",
      "description": "Build project"
    },
    "test": {
      "command": "go test -count=1 ./internal/... && echo '✓ tests pass'",
      "description": "Run unit tests"
    },
    "preview": {
      "command": "cd docs/ui-preview && python3 -m http.server 8080",
      "description": "Start UI preview server"
    },
    "lint": {
      "command": "golangci-lint run ./...",
      "description": "Run linter"
    }
  },
  "agents": {
    "ui-designer": {
      "model": "anthropic/claude-opus-4-7",
      "instructions": "You are a TUI/terminal UI designer. Focus on layout, color, spacing, and keyboard interaction patterns. Output design documents and HTML preview code.",
      "skills": [
        "opencode-ensemble"
      ]
    },
    "go-dev": {
      "model": "anthropic/claude-sonnet-4-7",
      "instructions": "You are a Go developer implementing a Bubbletea TUI application. Follow the existing patterns: MVU architecture, lipgloss styles, tea.Cmd async patterns.",
      "skills": [
        "opencode-ensemble"
      ]
    },
    "reviewer": {
      "model": "anthropic/claude-opus-4-7",
      "instructions": "Review Go code for correctness, MVU pattern adherence, goroutine safety, and error handling.",
      "skills": [
        "code-reviewer"
      ]
    }
  }
}
```

## 三、可用的 AI 能力

### 3.1 并行团队 (Ensemble)

适用于 dtui 的多模块并行开发：

```typescript
// 示例：并行实现容器详情 + Compose 面板
team_create({name: "dtui-features"})

team_tasks_add({
    tasks: [
        {content: "实现容器详情 Inspect 的展开/折叠分区", priority: "high"},
        {content: "实现 Compose 面板的项目列表+服务子表", priority: "high"},
        {content: "添加容器详情和 Compose 面板的单元测试", priority: "medium"},
    ]
})
```

### 3.2 专业技能 (Skills)

当前项目最适用的技能：

| Skill               | 用途                              |
|---------------------|---------------------------------|
| `opencode-ensemble` | 多 agent 并行协调                    |
| `go-dev`            | Go 代码实现 (Bubbletea, Docker SDK) |
| `ui-designer`       | TUI 布局、色彩、交互设计                  |
| `code-reviewer`     | Go 代码审查 (MVU 模式、goroutine 安全)   |
| `test-master`       | Go 测试 (model 测试、view 测试)        |
| `spec-miner`        | 从现有代码反向提取规格                     |
| `legacy-modernizer` | 代码重构、模块拆分                       |

### 3.3 典型工作流

```
1. 设计阶段
   Agent: ui-designer
   输入: "设计容器详情页的折叠分区布局"
   输出: design/features/container-detail.md + HTML 预览

2. 实现阶段
   Agent: go-dev
   输入: "按设计实现容器详情 Inspect 分区，7个分区支持 1-7键跳转"
   输出: internal/docker/inspect.go + internal/tui/view/xxx.go

3. 审查阶段
   Agent: reviewer
   输入: "审查 inspect.go 的 goroutine 安全和错误处理"
   输出: review 报告

4. 测试阶段
   Agent: test-master
   输入: "为 inspect.go 添加测试，mock Docker API"
   输出: internal/docker/inspect_test.go
```

## 四、MCP 服务集成

如果需要与外部系统集成，可配置 MCP 服务器：

```json
{
  "mcpServers": {
    "playwright": {
      "command": "npx",
      "args": [
        "@anthropic-ai/playwright-mcp"
      ],
      "env": {}
    },
    "github": {
      "command": "gh",
      "args": [
        "mcp"
      ],
      "env": {}
    },
    "filesystem": {
      "command": "npx",
      "args": [
        "@anthropic-ai/filesystem-mcp"
      ],
      "env": {}
    }
  }
}
```

常用 MCP:

- **playwright**: 打开 HTML 预览截图验证 UI
- **github**: Issue/PR 管理、Code Review
- **filesystem**: 文件操作权限扩展

## 五、提示词模板

### UI 设计

```
你是一个 TUI 设计师。设计 [页面名]。
- 框架: Bubbletea v2 + Lip Gloss v2
- 风格: k9s 紧凑风格
- 输出: design/ 文档 + docs/ui-preview/ 更新
```

### 功能实现

```
在 docker-tui 项目中实现 [功能名]。
- MVU 模式: tea.Cmd → Msg → Update → View
- Docker 调用通过 internal/docker/ 封装
- 新增消息类型定义在 internal/tui/model/xxx.go
- 渲染函数放在 internal/tui/view/xxx.go
- 键盘绑定放在 internal/tui/keys.go
```
