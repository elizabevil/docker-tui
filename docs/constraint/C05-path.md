# C05 Path 横切约束

## 适用范围

任何要求输入文件路径的字段(Local destination / source / container path)。

## 来源文档

- [../task/br-041-unified-form-path-completion.md](../task/br-041-unified-form-path-completion.md)
- [../task/br-043-form-action-bar-redesign.md](../task/br-043-form-action-bar-redesign.md)
- `internal/tui/state/path.go`

## 规则

### Local vs Container

- `Local`:`dtui` 进程所在机器的文件系统。即使连接远程 daemon,Local destination 仍写本机。
- `Container`:选中容器内部路径。

### 默认目录

- Local 路径的默认目录是 dtui 启动时 shell 的 `cwd`(由 `os.Getwd()` 取得),**不是 binary 所在目录**。
- 相对路径基于该 cwd 展开。

### 路径展开与绝对化 — 当前实现

- 提交前必须绝对化,且父目录必须存在(`normalizeSaveDestination`)
- 提交后用户输入的相对路径会被展开成绝对路径再写入
- 渲染层 `renderEditableValue` 直接显示 `f.Input.Text`(用户输入的相对路径原样显示,提交后才绝对化)

### 路径展开与绝对化 — 目标规范(BR-043)

- 路径字段每次 keystroke 后,若输入文本不是绝对路径,运行 `state.Absolute(text, cwd)` 写回 `f.Input.Text` 并相应调整 `Cursor`
- 渲染层始终显示绝对路径形式
- 详见 [../requirement/R03-form-action/R03-01-form-pattern.md](../requirement/R03-form-action/R03-01-form-pattern.md)

### Tab / Arrow / Enter 语义(Path)

- `Tab` 单候选直接补全 / 多候选补公共前缀+打开候选弹层 / 无候选 toast
- 弹层已开时 `Tab` 前移 / `Shift+Tab` 反移
- `Enter` 单候选应用,多候选确认弹层选中项,无候选下一字段
- `Ctrl+Space` 显式打开候选弹层

### 默认文件名

- Copy:`<cwd>/<container-name>-<source-base>-<YYYYMMDD-HHMMSS>.tar`
- Export:`<cwd>/<container-name>-filesystem-<YYYYMMDD-HHMMSS>.tar`
- Image Save:`<image-name>-<tag>-<timestamp>.tar`(后续 R02-image)
- 用户手工修改后,源路径变化不得覆盖

### 隐藏文件规则

- basename 以 `.` 开头时显示,否则隐藏

## 与需求文档的引用

- [../requirement/R01-container/R01-01-advanced-ops.md](../requirement/R01-container/R01-01-advanced-ops.md): Copy / Export
- [../requirement/R02-image/R02-02-import-tarball.md](../requirement/R02-image/R02-02-import-tarball.md): Import
- [../requirement/R03-form-action/R03-01-form-pattern.md](../requirement/R03-form-action/R03-01-form-pattern.md): 公共 Form 模式

## 待确认项

1. Container Path 补全(第二阶段)统一接口未确定,见 [../requirement/R04-runtime/R04-01-docker-podman-capabilities.md](../requirement/R04-runtime/R04-01-docker-podman-capabilities.md)。
2. Image Save / Load / Import 路径字段是否完全复用 Local 规则还是需要独立 mode。
3. 绝对化时机:keystroke 时 vs 提交前 — 当前为提交前,目标规范为 keystroke 时;cursor 调整需测试。