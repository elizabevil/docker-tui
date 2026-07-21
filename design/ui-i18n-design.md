# UI 国际化设计

## 1. 背景

dtui 当前支持 `en`、`zh` 以及兼容形式 `zh-CN` / `zh_CN`。国际化问题必须同时处理两件事：

1. 翻译资源必须可靠加载，语言切换不能静默回退到英文。
2. 翻译后的文本长度不能破坏固定轨道、表格列、帮助页和弹窗布局。

本文件描述 UI 国际化的约束、问题根因和实施方案；当前 UI 结构仍以 `current-design.md` 为权威，本文只补充语言相关规则。

## 2. 已确认问题

### 2.1 资源解析失败会伪装成"切换无效"

`internal/data/i18n/lang.go` 原先在资源读取或 JSONC 解析失败时将语言表置为空，并且没有错误提示。`zh.jsonc` 曾因缺少逗号无法解析，调用 `SetLang("zh")` 后 `Current()` 仍是 `zh`，但所有缺失键都回退到英文表，用户看到的结果像是语言切换没有生效。

约束：资源加载失败必须可诊断；语言代码的规范化和回退规则必须稳定。当前已支持 `zh`、`zh-CN`、`zh_CN`，未知语言回退 `en`。

### 2.2 Rune 数量不等于终端显示宽度

当前 `utils.VisibleLen` 主要用于去除 ANSI 后统计 Rune 数量。中文字符通常占两个终端单元格，但一个中文 Rune 只被计为一列，导致以下问题：

- 表格表头或单元格填充不足，分隔线错位。
- Header 的四列内容互相覆盖或提前换行。
- Footer、Toast、Breadcrumb 截断位置错误。
- 帮助页使用 `fmt.Sprintf("%-20s", ...)`，中文按 Go 字符串宽度补齐，视觉宽度不足。

### 2.3 固定字符串宽度假设

以下路径对英文长度有隐含假设：

- `help/view.go` 的固定 `%-20s` 快捷键列。
- Header 中 `Host`、`Runtime`、`Socket`、`Language` 等标签和固定比例列。
- Query、Footer、弹窗中的英文前缀和提示语。
- 表格自然宽度计算只以翻译后的字符串结果为输入，但缺少统一的终端单元格宽度函数。

## 3. 设计原则

### 3.1 翻译键与显示文本分离

- 数据模型、动作 ID、命令名、运行时名称保持稳定英文标识，不参与翻译。
- UI 渲染层只通过 `i18n.T(key, args...)` 获取面向用户的文本。
- 不使用已翻译文本作为状态判断、排序键或事件分支条件。
- 缺失翻译返回明确的键名，并在开发/测试阶段可检测。

### 3.2 所有布局使用终端单元格宽度

统一提供以下基础 API，页面和组件不得自行使用 `len`、Rune 数量或固定 `%-Ns` 估算显示宽度：

```go
func DisplayWidth(s string) int
func TruncateDisplay(s string, width int) string
func PadDisplay(s string, width int) string
```

实现应正确处理：

- ANSI 样式序列；
- 中日韩宽字符；
- 全角标点；
- 组合字符和零宽字符；
- 截断后的样式关闭。

如果依赖库的宽度算法与 Bubble Tea/Lip Gloss 输出不一致，应以 Lip Gloss 使用的终端宽度规则为准，并通过中英文快照测试锁定行为。

### 3.3 容器宽度先于文本渲染

每个组件遵循固定顺序：

1. 根据终端尺寸和布局 profile 确定容器宽度。
2. 以显示宽度计算列宽、内边距和间隙。
3. 对每个文本字段执行截断或换行。
4. 最后应用样式和颜色。

翻译文本不能反向改变父级轨道的尺寸；需要更多空间时应截断、换行或切换到 compact profile。

## 4. 分层方案

### 4.1 i18n 数据层

- 资源加载器返回可诊断的解析错误；生产模式使用英文回退，但保留错误状态供启动提示和测试读取。
- `Init` 只负责加载资源，`SetLang` 只负责规范化代码和切换当前表。
- 增加资源完整性测试：英文键集合是基准，中文缺失键必须显式报告。
- 增加语言切换测试：切换后 `Current()` 和代表性 UI 文本同时变化。

### 4.2 UI 文本层

- Header、Footer、Query、Dialog、Help、Breadcrumb 和空状态全部使用翻译键。
- 将 `Search:`、`Filter:`、`Pull:`、`Terminal too small`、`Select a container...` 等硬编码用户文本迁移到资源文件。
- 资源键按领域分组：`panel.*`、`table.*`、`key.*`、`hint.*`、`dialog.*`、`msg.*`、`status.*`、`toast.*`。
- 动作/事件常量继续使用稳定英文 ID，例如 `start`、`stop`、`removed`；只有显示时才翻译。

### 4.3 布局与渲染层

- 用 `DisplayWidth` 替换所有影响布局的 `len`、Rune 计数和固定格式补齐。
- Help 页按"快捷键列宽 + 描述列宽"动态分配，并对描述做显示宽度截断或换行。
- Header 标签与值分别计算宽度；标签超出列宽时使用紧凑翻译或截断，不允许覆盖相邻列。
- 表格列宽缓存的 key 必须包含 locale，因为不同语言的表头自然宽度不同。
- Toast、Footer、Breadcrumb 和 Dialog 在最终输出前统一执行单行宽度裁剪。

## 5. 实施阶段

### 阶段 A：资源可靠性

- 修复并校验全部 JSONC 资源。
- 加载错误可观测化，补充中英文切换和资源完整性测试。
- 规范化语言代码，默认 `en`。

### 阶段 B：显示宽度基础设施

- 引入终端单元格宽度函数及测试样例：ASCII、中文、全角符号、ANSI、混合文本。
- 让表格、Panel、Header、Footer、Dialog、Breadcrumb 统一使用该 API。
- 删除 UI 中的固定 `%-Ns` 补齐和基于 `len` 的布局计算。

### 阶段 C：文本覆盖与响应式布局

- 清理所有面向用户的硬编码英文。
- 为长中文翻译定义 compact/standard/wide 三档显示策略。
- 在 80、100、120 列终端宽度下分别验证中英文页面。

### 阶段 D：回归与发布约束

- 建立中英文渲染快照或关键组件宽度断言。
- 语言切换后验证：主页面、帮助页、详情页、日志页、弹窗、Toast、空状态。
- CI 中执行资源完整性检查、`go vet ./...`、`go test ./...` 和 `just check`。

## 6. 验收标准

- `--lang zh`、`--lang zh-CN` 和配置文件 `general.lang: zh` 显示中文；`--lang en` 恢复英文。
- 任意翻译资源解析失败时，不得静默显示"已切换但仍是英文"；测试能定位具体文件和键。
- 中英文在相同终端宽度下不发生列覆盖、边框错位、Footer 溢出或弹窗超宽。
- 所有文本截断按终端显示宽度执行，中文字符按实际占用列数计算。
- 动作分支、事件处理和审计仍使用稳定常量，不依赖翻译结果。

## 7. 当前状态

- 已修复中文 JSONC 解析错误。
- 已支持 `zh`、`zh-CN`、`zh_CN` 语言代码。
- 已补充语言切换测试。
- 已引入按终端单元格计算的 `DisplayWidth`、`TruncateVisible` 和 `PadVisible` 路径，并覆盖表格行前缀与 Help 快捷键列。
- 终端显示宽度的全页面覆盖、硬编码文本迁移和中英文渲染回归仍待实施。

## 8. 详情页 i18n 设计

### 8.1 翻译策略

**Docker 技术术语保持英文**：ID、PID、CPU、IP、MAC、Port、Entrypoint、Cmd 等是行业标准术语，翻译后反而不利于用户理解和与其他工具对照。

**描述性文字翻译**：Section 标题（Resources → 资源）、提示文本（Loading detail → 加载详情中...）、详情页标题等。

### 8.2 容器详情 i18n

当前 `internal/data/docker/inspect.go` 的 `InspectContainer()` 返回硬编码英文文本。需改造为使用 `i18n.T()` 翻译 Section 标题和描述性字段。

**Section 标题翻译**：

| 原始英文 | i18n key | 中文 |
|---------|----------|------|
| `── Resources ──` | `inspect.section_resources` | 资源 |
| `── Networks ──` | `inspect.section_networks` | 网络 |
| `── Mounts ──` | `inspect.section_mounts` | 挂载 |
| `── Config ──` | `inspect.section_container_config` | 配置 |
| `── Labels ──` | `inspect.section_labels` | 标签 |

**字段标签翻译**：

| 原始英文 | i18n key | 中文 |
|---------|----------|------|
| `CPUShares` | `inspect.container.cpu_shares` | CPUShares |
| `Memory` | `inspect.container.memory` | 内存 |
| `NanoCPUs` | `inspect.container.nano_cpus` | NanoCPUs |
| `NetworkMode` | `inspect.container.network_mode` | 网络模式 |
| `RestartPolicy` | `inspect.container.restart_policy` | 重启策略 |
| `StartedAt` | `inspect.container.started_at` | 启动时间 |
| `FinishedAt` | `inspect.container.finished_at` | 结束时间 |
| `RestartCount` | `inspect.container.restart_count` | 重启次数 |
| `WorkingDir` | `inspect.container.working_dir` | 工作目录 |
| `User` | `inspect.container.user` | 用户 |
| `Entrypoint` | `inspect.container.entrypoint` | 入口点 |
| `Cmd` | `inspect.container.cmd` | 命令 |
| `ExposedPorts` | `inspect.container.exposed_ports` | 暴露端口 |
| `Env` | `inspect.container.env` | 环境变量 |
| `IP` | `inspect.container.ip` | IP |
| `Gateway` | `inspect.container.gateway` | 网关 |
| `MAC` | `inspect.container.mac` | MAC |
| `Ports` | `inspect.container.ports` | 端口映射 |

**注意**：`ID`、`Name`、`Image`、`Created`、`State`、`Pid`、`Platform` 等字段名保持英文不翻译。

**详情页标题 i18n**：

| 原始硬编码 | i18n key | 中文示例 |
|-----------|----------|---------|
| `"Container Detail: "+name+" ("+id+")"` | `detail.title.container` | 容器详情: myapp (abc123) |
| `"Image Detail: "+shortID` | `detail.title.image` | 镜像详情: sha256:abc123 |

### 8.3 网络详情 i18n（新增功能）

网络详情页当前不存在。新增后需 i18n 的字段：

**Section 标题**：

| i18n key | 中文 |
|----------|------|
| `inspect.section_network_info` | 网络信息 |
| `inspect.section_network_ipam` | IP 配置 |
| `inspect.section_network_containers` | 连接容器 |
| `inspect.section_labels` | 标签 |

**字段标签**：

| i18n key | 中文 |
|----------|------|
| `inspect.network.name` | 名称 |
| `inspect.network.id` | ID |
| `inspect.network.driver` | 驱动 |
| `inspect.network.scope` | 范围 |
| `inspect.network.subnet` | 子网 |
| `inspect.network.gateway` | 网关 |
| `inspect.network.ip_range` | IP 范围 |
| `inspect.network.internal` | 内部网络 |
| `inspect.network.ipv6` | IPv6 |
| `inspect.network.labels` | 标签 |
| `inspect.network.created` | 创建时间 |

**详情页标题**：

| i18n key | 中文示例 |
|----------|---------|
| `detail.title.network` | 网络详情: bridge |

### 8.4 卷详情 i18n（新增功能）

卷详情页当前不存在（当前卷面板的"详情"是子视图，显示卷内容器列表）。新增 inspect 功能后需 i18n 的字段：

**Section 标题**：

| i18n key | 中文 |
|----------|------|
| `inspect.section_volume_info` | 卷信息 |
| `inspect.section_labels` | 标签 |
| `inspect.section_volume_options` | 选项 |

**字段标签**：

| i18n key | 中文 |
|----------|------|
| `inspect.volume.name` | 名称 |
| `inspect.volume.driver` | 驱动 |
| `inspect.volume.mountpoint` | 挂载点 |
| `inspect.volume.scope` | 范围 |
| `inspect.volume.labels` | 标签 |
| `inspect.volume.options` | 选项 |
| `inspect.volume.created` | 创建时间 |

**详情页标题**：

| i18n key | 中文示例 |
|----------|---------|
| `detail.title.volume` | 卷详情: my-volume |

### 8.5 镜像详情 i18n（已实现，补充）

镜像详情已通过 `internal/tui/ui/pages/detail/image.go` 使用 `i18n.T()` 翻译。需补充的细节：

- `buildImageDetailSections()` 中的 section 分隔符 `"── System ──"` 等仍为硬编码英文，需改为 `i18n.T()` 调用。
- `"Fields parsed: %d"` → `i18n.T("inspect.fields_parsed", len(kv))`
- `"Source: docker image inspect"` → 已有 `inspect.source` key

## 9. 源码视图（JSON/YAML 切换）

### 9.1 需求概述

详情页当前只提供 Section 分组视图。新增源码视图功能，支持 JSON 和 YAML 两种格式，展示 Docker API 返回的完整原始数据，方便用户复制使用。

### 9.2 状态管理

`internal/tui/state/app.go` 的 `AppModel` 新增字段：

```go
// DetailSourceType 控制详情页显示模式："section" | "yaml" | "json"
DetailSourceType string

// DetailRawJSON 存储 Docker API 返回的原始 JSON 字节
DetailRawJSON []byte
```

默认值：`DetailSourceType = "section"`

### 9.3 快捷键设计

在详情页（`ModeDetail`）中，按 `s` 键循环切换三种显示模式：

```
Section (默认) → YAML → JSON → Section
```

切换时重置 `DetailOffset = 0`（滚动到顶部）。

**源码视图中的快捷键**：
- `j/k` 或 方向键：滚动
- `g/G`：跳转到顶部/底部
- `Space/PgDn`：向下翻页
- `PgUp`：向上翻页
- `Esc/Enter`：返回上一级
- `s`：切换到下一个显示模式

### 9.4 Footer 提示

源码视图的 Footer 显示当前模式和操作提示：

```
1-20/150 │ s:section y:yaml j:json │ Esc:back
```

使用 i18n key：`detail.source.hint`

### 9.5 数据流改造

#### 容器 inspect 改造

当前 `InspectContainer()` 返回格式化字符串。改造为返回原始 JSON：

```go
// 改造前
func (c *Client) InspectContainer(id string) (string, error)

// 改造后
func (c *Client) InspectContainer(id string) ([]byte, error)
```

内部调用 `c.cli.ContainerInspectWithRaw(c.ctx, id)` 获取原始 JSON 字节。Section 视图通过解析 JSON 生成（新增 `buildContainerDetailSections(jsonData []byte)` 函数）。

#### 镜像 inspect 改造

当前 `InspectImageDetail()` 已调用 `ImageInspectWithRaw` 但丢弃了原始 JSON。改造为同时返回原始 JSON：

```go
// 改造前
func (c *Client) InspectImageDetail(summary ImageSummary) (*ImageDetailData, error)

// 改造后：增加返回值
func (c *Client) InspectImageDetail(summary ImageSummary) (*ImageDetailData, []byte, error)
```

#### 新增网络 inspect

```go
func (c *Client) InspectNetwork(id string) ([]byte, error)
```

内部调用 `c.cli.NetworkInspectWithRaw(c.ctx, id, ...)` 获取原始 JSON。同时返回结构化数据用于 Section 视图。

#### 新增卷 inspect

```go
func (c *Client) InspectVolume(name string) ([]byte, error)
```

内部调用 `c.cli.VolumeInspectWithRaw(c.ctx, name, ...)` 获取原始 JSON。

### 9.6 渲染逻辑

`internal/tui/ui/pages/detail/view.go` 的 `RenderView()` 根据 `DetailSourceType` 分支：

```
DetailSourceType == "section"  → 当前 Section 渲染（不变）
DetailSourceType == "yaml"     → JSON → interface{} → yaml.Marshal → 等宽渲染
DetailSourceType == "json"     → json.MarshalIndent → 等宽渲染
```

**YAML 转换流程**：
1. `sonic.Unmarshal(DetailRawJSON, &obj)` 反序列化为 `interface{}`
2. `yaml.Marshal(obj)` 转换为 YAML
3. 按行渲染，长行截断

**JSON 渲染流程**：
1. `json.MarshalIndent(DetailRawJSON, "", "  ")` 格式化
2. 按行渲染，长行截断

**源码视图标题**：

| 模式 | 标题格式 | i18n key | 中文示例 |
|------|---------|----------|---------|
| Section | `Container Detail: name (id)` | `detail.title.container` | 容器详情: myapp (abc123) |
| YAML | `Container YAML: name (id)` | `detail.title.container_yaml` | 容器 YAML: myapp (abc123) |
| JSON | `Container JSON: name (id)` | `detail.title.container_json` | 容器 JSON: myapp (abc123) |

## 10. 新增 Docker inspect 方法

### 10.1 方法清单

| 方法 | 文件 | Docker SDK 调用 | 返回 |
|------|------|----------------|------|
| `InspectContainer` 改造 | `inspect.go` | `ContainerInspectWithRaw` | `([]byte, error)` |
| `InspectImageDetail` 改造 | `images.go` | `ImageInspectWithRaw`（已用） | `(*ImageDetailData, []byte, error)` |
| `InspectNetwork` 新增 | `networks.go` | `NetworkInspectWithRaw` | `([]byte, error)` |
| `InspectVolume` 新增 | `volumes.go` | `VolumeInspectWithRaw` | `([]byte, error)` |

### 10.2 网络 inspect 结构化数据

新增 `NetworkDetailData` 结构体用于 Section 视图：

```go
type NetworkDetailData struct {
    Name       string
    ID         string
    Driver     string
    Scope      string
    Created    string
    Internal   bool
    IPv6       bool
    IPAM       NetworkIPAM
    Containers map[string]NetworkContainer
    Labels     map[string]string
}

type NetworkIPAM struct {
    Driver  string
    Config  []NetworkIPAMConfig
}

type NetworkIPAMConfig struct {
    Subnet     string
    Gateway    string
    IPRange    string
    AuxAddress map[string]string
}

type NetworkContainer struct {
    Name        string
    EndpointID  string
    MacAddress  string
    IPv4Address string
    IPv6Address string
}
```

### 10.3 卷 inspect 结构化数据

新增 `VolumeDetailData` 结构体用于 Section 视图：

```go
type VolumeDetailData struct {
    Name       string
    Driver     string
    Mountpoint string
    Scope      string
    CreatedAt  string
    Labels     map[string]string
    Options    map[string]string
    Status     map[string]interface{}
}
```

### 10.4 网络/卷详情入口

**网络详情**：在 `keyboard/network_action.go` 中新增 `doNetworkInspect` 函数，绑定到 Enter 键。调用 `InspectNetwork` 后通过 `ToDetail()` 进入详情页。

**卷详情**：在 `keyboard/volume_action.go` 中新增 `doVolumeInspect` 函数，绑定到 Enter 键（替代当前的子视图行为）。调用 `InspectVolume` 后通过 `ToDetail()` 进入详情页。

**注意**：卷的 Enter 键当前用于进入"卷内容器"子视图。改为进入 inspect 详情后，"卷内容器"功能需通过其他方式访问（如详情页中列出使用该卷的容器）。

## 11. 新增 i18n key 列表

### 11.1 容器详情新增 key

```jsonc
// Section 标题
"inspect.section_resources": "Resources"        // 资源
"inspect.section_networks": "Networks"          // 网络
"inspect.section_mounts": "Mounts"              // 挂载
"inspect.section_container_config": "Config"    // 配置
"inspect.section_labels": "Labels"              // 标签

// 字段标签
"inspect.container.cpu_shares": "CPUShares"
"inspect.container.memory": "Memory"
"inspect.container.nano_cpus": "NanoCPUs"
"inspect.container.network_mode": "NetworkMode"
"inspect.container.restart_policy": "RestartPolicy"
"inspect.container.started_at": "StartedAt"
"inspect.container.finished_at": "FinishedAt"
"inspect.container.restart_count": "RestartCount"
"inspect.container.working_dir": "WorkingDir"
"inspect.container.user": "User"
"inspect.container.entrypoint": "Entrypoint"
"inspect.container.cmd": "Cmd"
"inspect.container.exposed_ports": "ExposedPorts"
"inspect.container.env": "Env"
"inspect.container.ip": "IP"
"inspect.container.gateway": "Gateway"
"inspect.container.mac": "MAC"
"inspect.container.ports": "Ports"
"inspect.container.platform": "Platform"
"inspect.container.pid": "PID"

// 状态描述
"inspect.container.state.running": "Running"
"inspect.container.state.exited": "Exited"
"inspect.container.state.created": "Created"
```

### 11.2 网络详情新增 key

```jsonc
"inspect.section_network_info": "Network Info"      // 网络信息
"inspect.section_network_ipam": "IP Configuration"  // IP 配置
"inspect.section_network_containers": "Connected Containers"  // 连接容器

"inspect.network.name": "Name"
"inspect.network.id": "ID"
"inspect.network.driver": "Driver"
"inspect.network.scope": "Scope"
"inspect.network.subnet": "Subnet"
"inspect.network.gateway": "Gateway"
"inspect.network.ip_range": "IP Range"
"inspect.network.internal": "Internal"
"inspect.network.ipv6": "IPv6"
"inspect.network.labels": "Labels"
"inspect.network.created": "Created"
```

### 11.3 卷详情新增 key

```jsonc
"inspect.section_volume_info": "Volume Info"    // 卷信息
"inspect.section_volume_options": "Options"     // 选项

"inspect.volume.name": "Name"
"inspect.volume.driver": "Driver"
"inspect.volume.mountpoint": "Mountpoint"
"inspect.volume.scope": "Scope"
"inspect.volume.labels": "Labels"
"inspect.volume.options": "Options"
"inspect.volume.created": "Created"
```

### 11.4 源码视图新增 key

```jsonc
"detail.source.yaml": "YAML"
"detail.source.json": "JSON"
"detail.source.hint": "s:section y:yaml j:json"

// 详情页标题
"detail.title.container": "Container Detail: {0} ({1})"
"detail.title.container_yaml": "Container YAML: {0} ({1})"
"detail.title.container_json": "Container JSON: {0} ({1})"
"detail.title.image": "Image Detail: {0}"
"detail.title.image_yaml": "Image YAML: {0}"
"detail.title.image_json": "Image JSON: {0}"
"detail.title.network": "Network Detail: {0}"
"detail.title.network_yaml": "Network YAML: {0}"
"detail.title.network_json": "Network JSON: {0}"
"detail.title.volume": "Volume Detail: {0}"
"detail.title.volume_yaml": "Volume YAML: {0}"
"detail.title.volume_json": "Volume JSON: {0}"
```

### 11.5 通用新增 key

```jsonc
// 提示文本
"inspect.copied_json": "JSON copied to clipboard"
"inspect.copied_yaml": "YAML copied to clipboard"
```

## 12. 实施计划

### 阶段 1：容器详情 i18n（改造 inspect.go）

**涉及文件**：
- `internal/data/docker/inspect.go`：改造 `InspectContainer` 返回 `[]byte`，新增 `buildContainerDetailSections`
- `internal/data/i18n/en.jsonc`：新增容器详情 key
- `internal/data/i18n/zh.jsonc`：新增容器详情 key
- `internal/tui/keyboard/container_action.go`：修改 `doInspectAction` 适配新返回值
- `internal/tui/ui/pages/detail/view.go`：新增容器 Section 解析路径

**验收**：中英文切换后，容器详情页的 Section 标题和描述性字段正确翻译。

### 阶段 2：源码视图功能

**涉及文件**：
- `internal/tui/state/app.go`：新增 `DetailSourceType` 和 `DetailRawJSON` 字段
- `internal/tui/keyboard/detail.go`：新增 `s` 键循环切换逻辑
- `internal/tui/ui/pages/detail/view.go`：新增 YAML/JSON 渲染分支
- `internal/tui/ui/pages/detail/view.go`：修改 `RenderView` 支持源码视图
- `internal/data/i18n/en.jsonc`：新增源码视图 key
- `internal/data/i18n/zh.jsonc`：新增源码视图 key

**验收**：在容器/镜像详情页按 `s` 键可循环切换 Section → YAML → JSON，三种视图数据一致。

### 阶段 3：网络/卷 inspect 新功能

**涉及文件**：
- `internal/data/docker/types.go`：新增 `NetworkDetailData`、`VolumeDetailData` 结构体
- `internal/data/docker/networks.go`：新增 `InspectNetwork` 方法
- `internal/data/docker/volumes.go`：新增 `InspectVolume` 方法
- `internal/tui/keyboard/network_action.go`：新增 `doNetworkInspect`
- `internal/tui/keyboard/volume_action.go`：新增 `doVolumeInspect`，调整 Enter 键行为
- `internal/tui/ui/pages/detail/image.go`：复用或新增网络/卷 Section 渲染
- `internal/data/i18n/en.jsonc`：新增网络/卷详情 key
- `internal/data/i18n/zh.jsonc`：新增网络/卷详情 key

**验收**：在网络/卷面板按 Enter 可进入详情页，支持 Section/YAML/JSON 三种视图。

### 阶段 4：回归测试与收尾

- 补充中英文切换后所有详情页的渲染测试。
- 验证 YAML/JSON 源码视图的完整性（数据不丢失）。
- 验证 `s` 键循环切换的边界行为（快速按键、重复按键）。
- 更新 `image.go` 中遗留的硬编码英文 section 分隔符。
- CI 中执行 `go vet ./...`、`go test ./...`。
