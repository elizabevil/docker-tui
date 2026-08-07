# Config UX CLI 便利化

> 日期: 2026-08-05
> 状态: 计划阶段，待用户确认后实施

## 1. 目标

为用户提供便利的 CLI 工具来管理配置文件，包括查看配置路径、初始化配置模板、校验配置有效性。

**关键设计**：
- 顶级 `dtui info` 命令显示配置路径和可用主题
- `dtui config init` 生成带注释的分节配置模板
- `dtui config validate` 校验配置文件
- 首次运行无配置时 toast 提示

## 2. 改动范围

### 2.1 必须完成

| 组件 | 描述 |
|---|---|
| `dtui info` | 顶级命令，打印配置/日志/主题目录路径 + 可用主题列表 |
| `dtui config init` | 生成配置文件模板，按类型分节（样式/快捷键/日志/命令等） |
| `dtui config validate` | 校验配置文件，复用 ValidateApp，输出字段路径错误 |
| 首次运行提示 | 无 config.yml 时 toast 提示 `dtui config init` |
| 文档更新 | README Configuration 章节 + docs/navigation.md |

### 2.2 详细设计

#### 2.2.1 dtui info

```
Config File: /home/user/.config/docker-tui/config.yml
Log Directory: /home/user/.config/docker-tui/logs/
Theme Directory: /home/user/.config/docker-tui/themes/

Available Themes:
  - default
  - dark
  - light
  - nord
  - dracula
  - solarized
```

#### 2.2.2 dtui config init

- 生成 config.yml 到 `ConfigFile()` 路径
- 已存在则报错不覆盖
- 模板按 UserConfig 真实结构分节，每节带注释说明

#### 2.2.3 dtui config validate

- 复用 `LoadResolved` 进行完整校验
- 无文件时提示"未找到配置，运行 dtui config init"
- 退出码区分（0=有效，1=无效，2=缺失）

### 2.3 不得修改

- 不实现配置热加载/热更新
- 不实现应用内 Settings 页/配置编辑器
- 不实现应用内打开 config.yml 动作
- 不实现主题视觉编辑器
- 不实现 shell 补全（列为未来需求）
- 不移除现有的 --theme 全局 flag

## 3. 验收标准

- [ ] `dtui info` 输出配置路径和主题列表
- [ ] `dtui config init` 生成带注释的分节模板
- [ ] `dtui config validate` 正确校验并输出字段路径错误
- [ ] 首次运行无配置时 toast 提示
- [ ] 所有测试通过

## 4. 相关需求

- R07: Config UX CLI 便利化

## 5. 待确认问题

无
