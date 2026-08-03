# R03-04 路径补全 popup 设计

> 状态: implementing(v5 commits, 5 已知问题待修)
> 关联 BR: BR-041 §3.3/§6、BR-043 §3.1
> 实现 commits: 7da5b10 / 9c87403 / 640f9eb / 645f35d

## 1. 设计目标

FormPath 字段的补全 popup 应提供 **类 Linux shell + 文件管理器** 的体验:

- **eza -l 风格详情**:每行显示 类型/权限/属主/组/大小/时间/名(类似 `ls -la` / `eza -l`)
- **shell-like BrowseMode**:popup 打开后仍可继续输入字符过滤,↑↓ 切换候选
- **文件管理器导航**:Enter on dir 自动 drill-down 加 `/`、← 父目录、→ 子目录
- **可滚动详情**:选中项详情框显示完整信息,长路径 wrap 换行

参考: `eza -l` 输出格式、`fish` / `zsh` 的补全体感、`ranger` / `lf` 的目录导航。

## 2. 布局规范

### 2.1 popup 整体结构

```
┌─ Container path ────────────────────────┐  ← 字段标签
│ /etc/nginx/                              │  ← 面包屑(青色)
│ T MODE       OWNER  GROUP   SIZE   DATE        NAME   │  ← 列头
│ D drwxrwxr-x root   root    4.0K  Jul 01 14:30 ./     │
│ > D drwxrwxr-x root   root    4.0K  Jul 01 14:30 ../    │  ← 选中(青色高亮)
│   D drwxr-xr-x root   root    4.0K  Jul 01 14:30 conf/   │
│   F -rw-r--r-- root   root    1.2K  Jul 01 14:30 passwd  │
│   L lrwxrwxrwx root   root       7  Jun 08  2024 man → share/man │
├─────────────────────────────────────────┤
│ drwxr-xr-x root:root 4.0K Jul 01 2026 14:30 │  ← 详情框 line 1
│ Name: conf/                              │  ← 详情框 line 2
└─────────────────────────────────────────┘
```

### 2.2 列定义

| 列  | 宽度 | 类型 | 说明 |
|-----|------|------|------|
| T   | 1    | char | F(文件) / D(目录) / L(链接) |
| MODE| 10   | str  | `drwxrwxr-x` 格式(空填 `----------`)|
| OWNER| 5   | str  | 属主(空填 `-`,超长截断)|
| GROUP| 5   | str  | 组(空填 `-`)|
| SIZE | 7   | str  | human-readable,K/M/G/T(空填 `0`)|
| DATE | 12  | str  | `Jul 01 14:30`(空填空白)|
| NAME | rest| str  | 文件名;目录加尾 `/`;链接 `name → target` |

### 2.3 详情框

```
{permissions} {owner}:{group} {size} {date}
Name: {name}[ → {link_target}]
```

- 第一行:mode owner:group size date
- 第二行:Name 加链接 target(仅 symlink)
- 长 Name wrap 换行(超 innerW 宽度)

## 3. 交互规范

### 3.1 打开 popup

| 触发条件 | 行为 |
|---------|------|
| `Tab` 首次按下 | 发送 completion 请求(openPopup=false) |
| `Tab` 第二次按下 | popup 打开(等待 async 返回)|
| `Ctrl+Space` | 直接打开 popup(openPopup=true)|
| `Enter`(BrowseMode)| 高亮首匹配 + 应用 |

### 3.2 BrowseMode(popup 打开时)

| 键 | 行为 |
|---|------|
| 任意可打印字符 | 编辑字段 + 触发 completion(staleness check 去重) |
| `Backspace` / `Delete` | 同上 |
| `Ctrl+H` | 切换隐藏文件(ShowHidden) + 立即重触发 completion |
| `↑` / `↓` | 切换候选 |
| `Tab` / `Shift+Tab` | 切换候选 |
| `Home` / `End` | popup cursor 首/尾 |
| `PgUp` / `PgDn` | popup cursor 翻页 |
| `←` | 父目录(navigatePathPopupParent)|
| `→` | 子目录(navigatePathPopupChild,仅目录)|
| `Enter` | 应用候选;目录 → drill-down(追加 `/` + 新 completion) |
| `Esc` | 关闭 popup,字段内容保留 |

### 3.3 debounce

不显式实现 timer debounce,依赖 `HandleContainerPathCompleted` 的 staleness check:
```go
if f.Input.Text != msg.Input || !f.PathLoading {
    return m, nil  // 旧请求被丢弃
}
```

每次按键触发 completion,旧请求自然失效。实际效果等价 debounce。

## 4. 数据模型

### 4.1 PathEntry(state/path.go)

```go
type PathEntryType rune  // 'F' / 'D' / 'L'

type PathEntry struct {
    Name       string         // basename,不带 /
    Path       string         // 完整路径
    IsDir      bool           // 是否目录(冗余,等价 Type==PathEntryDir)
    Type       PathEntryType  // F/D/L
    Mode       string         // "drwxr-xr-x"
    Owner      string         // 属主
    Group      string         // 组
    Size       int64          // 字节
    Mtime      time.Time      // 修改时间
    LinkTarget string         // 软链目标(仅 symlink)
}
```

### 4.2 FormField 扩展(state/form.go)

```go
ShowHidden bool   // Ctrl+H 切换
PathError  string // 最后一次 completion 错误(popup 红 ⚠ 显示)
```

### 4.3 完成来源

| 来源 | 命令/调用 | 字段填充 |
|------|----------|----------|
| Container | shell `stat -c "kind\|name\|%A\|%U\|%G\|%s\|%Y\|target"` | 全部 8 字段 |
| Local | `os.Lstat` + `os.Readlink` | 全部 8 字段 |

## 5. 已知问题(5 项待修)

用户反馈"都有问题",需要逐项修复。

### 5.1 列对齐错位(空字段时)

**症状**:当 Owner/Group/Mtime 为空时,Mode 列后的空格数量变化,导致后续列左移。

**根因**:`utils.PadVisible` 用 `DisplayWidth` 计算宽度,但 mode 为空时 fallback 到 `----------` 仍占 10 字符;Owner/Group 空时 fallback 到 `-`(1 字符)而非 5 字符空格,总宽度变化。

**修复方向**:统一使用 fallback 字符串(`-`)而非空格填充,列宽固定由后续空格决定;或者改用 lipgloss table 风格的列对齐。

### 5.2 链接 → target 截断

**症状**:`man → share/man` 在窄宽度下被裁,只显示 `man → share/m...`。

**根因**:`formatPathRow` 的 name 列受 bodyW 限制,链接 target 占用 name 空间。

**修复方向**:
- 选项 A:链接显示纯 name,target 只在详情框显示(当前已实现)
- 选项 B:name 列允许链接时占用更多宽度,超长 wrap
- 倾向 A(已实现),需验证 popup 渲染逻辑不再在 name 列追加 `→ target`

### 5.3 详情框与列表间距

**症状**:详情框紧贴列表最后一行,无视觉分隔。

**根因**:`renderPathPopup` 只插入 `""` 空行作为分隔。

**修复方向**:插入 `─` 横线分隔符或 `·` 装饰,增加视觉层次。

### 5.4 面包屑换行

**症状**:面包屑路径过长时换行,占用额外行,popup 高度变化。

**根因**:`component.FormRow("", 0, innerW, path)` 仅做横向截断,不处理换行。

**修复方向**:
- 选项 A:限制面包屑单行,超长截断(`Jul 01 14:30` 风格)
- 选项 B:面包屑支持 wrap(类似 fish 的提示)
- 倾向 A(简单、与单行字段标签对齐)

### 5.5 Size 数字宽度对齐偏差

**症状**:`4.0K` 与 `1.2M` 右对齐时小数点不对齐,视觉不齐。

**根因**:`utils.FormatBytes` 返回 `%.1f%s`,整数位宽度不一致(`4.0K` 5 字符,`1.2M` 5 字符 — 实际长度一致,问题可能在于 mono 字体下小数点位置视觉不齐)。

**修复方向**:
- 选项 A:改用 mono 字体(M 字段),或加 ANSI 控制强制对齐
- 选项 B:移除小数,统一 `4K` / `1234K`(右对齐更稳)
- 倾向 B,需修改 `HumanSizeBytes`

## 6. 修复优先级

| 优先级 | 问题 | 影响 |
|-------|------|------|
| P1 | 5.1 列对齐错位 | 阅读体验严重受损 |
| P1 | 5.5 Size 对齐偏差 | 数据难以快速对比 |
| P2 | 5.4 面包屑换行 | popup 高度不稳定 |
| P2 | 5.3 详情框间距 | 视觉层次不清 |
| P3 | 5.2 链接 target | 信息完整但展示位置已迁详情框 |

## 7. 测试覆盖

- `TestRenderFormPopupPathRows`:path 候选必须列出(`backup.tar` + `sub/`)
- `TestPathPopupTypeColumnAligned`:D/F/L 前缀位置对齐
- `TestPathPopupGeometryStaysFixedAcrossDirectoryLoads`:popup 高度在候选数变化时稳定

需补充测试:
- 空 Owner/Group/Size/Mtime 时的列对齐
- 长路径的 wrap 行为
- 链接在详情框的显示
- Ctrl+H 切换 ShowHidden 后重触发 completion
- BrowseMode 字符输入触发 completion

## 8. 关联文档

- [R03-01-form-pattern.md](./R03-01-form-pattern.md) — FormField 状态机
- [R03-03-action-dialog-layout.md](./R03-03-action-dialog-layout.md) — FormDialog 整体布局
- [C05-path.md](../../constraint/C05-path.md) — Path 字段语义
- [../../task/br-043-form-action-bar-redesign.md](../../task/br-043-form-action-bar-redesign.md) — BR-043 决策来源