# dtui 国际化设计

## 架构

```
┌─────────────────────────────────────────────────────────────┐
│  应用层                                                      │
│  i18n.T("panel.containers")  →  "容器" / "Containers"       │
├─────────────────────────────────────────────────────────────┤
│  翻译引擎 (internal/i18n/)                                   │
│  T(key, args) → 当前语言 → 英文备选 → key本身降级            │
├─────────────────────────────────────────────────────────────┤
│  数据层                                                      │
│  go:embed zh.json + en.json → map[string]string            │
└─────────────────────────────────────────────────────────────┘
```

## 文件结构

```
internal/i18n/
├── embed.go               // go:embed *.json
├── lang.go                // Init(), T(), SetLang(), Current()
├── en.json                // 英文 (默认)
└── zh.json                // 中文
```

## 翻译引擎

```go
package i18n

// Init 初始化翻译引擎，code 为 "zh"/"en"
func Init(code string)

// T 翻译 key，{0} {1} 替换位置参数
// 降级链: 当前语言 → 英文 → key本身
func T(key string, args ...any) string

// SetLang 运行时切换语言
func SetLang(code string)

// Current 返回当前语言代码
func Current() string
```

## 翻译文件

扁平 key + 位置参数:

```json
{
  "panel.containers": "Containers",
  "panel.images": "Images",
  "toast.started": "Started {0}",
  "status.connected": "● {0} │ {1} containers",
  "key.start": "Start",
  "confirm.delete_title": "Delete {0}",
  "confirm.force": "Force remove"
}
```

## 命名空间

```
panel.xxx        面板标题
key.xxx          快捷键描述
toast.xxx        操作反馈 Toast
status.xxx       状态栏
confirm.xxx      确认弹窗
search.xxx       搜索
sort.xxx         排序
table.xxx        表头列名
help.xxx         帮助页
err.xxx          错误信息
```

## 使用方式

```go
import "docker-tui/internal/i18n"

// 初始化 (在 main.go)
i18n.Init("zh")                  // 或 cfg.General.Lang

// 各处使用
panelTitle := i18n.T("panel.containers")
toastMsg   := i18n.T("toast.deleted", containerName)
status     := i18n.T("status.connected", engineLabel, count)
```

## 布局兼容

- 列宽按英文计算，中文占用更少空间，自动适配
- `truncate()` 兜底超长翻译
- 中文翻译长度控制在英文 1.5 倍以内

## 配置

```yaml
# config.yml
general:
  lang: "zh"           # 省略时默认 en
```

```bash
dtui --lang zh              # CLI 参数优先
dtui --config config.yml    # 配置文件
```

## 批次替换计划

| 批次    | 范围                      | 内容                               |
|-------|-------------------------|----------------------------------|
| 第 1 批 | model/app.go, keys.go   | PanelLabel, HelpEntries (安全无副作用) |
| 第 2 批 | view/*.go               | 表头、状态栏、快捷键栏、帮助页                  |
| 第 3 批 | handler_*.go, update.go | Toast 消息、确认弹窗、错误消息               |
