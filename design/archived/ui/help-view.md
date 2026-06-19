# Help View — 帮助页面

## 状态

- [x] 已实现

## 触发

- 按 `?`
- 替换主内容区域

## 布局

```
┌──────────────────────────────────────────────────────────────────┐
│                                                                  │
│    ██████╗ ████████╗██╗   ██╗██╗                                │
│    ██╔══██╗╚══██╔══╝██║   ██║██║                                │
│    ██║  ██║   ██║   ██║   ██║██║                                │
│    ██║  ██║   ██║   ██║   ██║██║                                │
│    ██████╔╝   ██║   ╚██████╔╝██║                                │
│    ╚═════╝    ╚═╝    ╚═════╝ ╚═╝                                │
│    Docker & Podman TUI  v0.2.0                                   │
│                                                                  │
│  Navigation                                                       │
│    ↑↓ / j k           Move cursor                                │
│    Tab / S-Tab        Switch panel                               │
│    Enter              Select / Image detail                      │
│    Esc                Go back / Close                            │
│                                                                  │
│  Containers                                                       │
│    s                  Start                                      │
│    S                  Stop                                       │
│    R                  Restart                                    │
│    K                  Kill                                       │
│    d / x              Remove                                     │
│    l                  View logs                                  │
│    e                  Exec shell hint                            │
│    i                  Inspect                                    │
│    m                  Toggle stats                               │
│                                                                  │
│  Images                                                           │
│    P                  Pull image                                 │
│    p                  Prune dangling                             │
│    d / x              Remove image                               │
│    o                  Sort by column                             │
│    Enter              Image detail                               │
│                                                                  │
│  Volumes & Networks                                               │
│    d / x              Remove (Volume/Network)                    │
│                                                                  │
│  Global                                                           │
│    /                  Filter / Search                            │
│    Space              Mark for bulk                              │
│    r                  Refresh all                                │
│    ?                  Toggle help                                │
│    q / Ctrl+C         Quit                                       │
│                                                                  │
│  Themes (-t flag)                                                 │
│    default            Dark cyan theme                            │
│    dark               High contrast dark                         │
│    light              Clean light theme                          │
│    nord               Arctic blue palette                        │
│    dracula            Purple dark theme                          │
│    solarized          Precision colors                           │
│                                                                  │
│  Press '?' or Esc to close help                                  │
└──────────────────────────────────────────────────────────────────┘
```
