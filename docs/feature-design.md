# dtui 应用功能设计 (App Feature Design)

> 建立日期: 2026-08-01
> 目的: 记录 dtui **最终实现目标**与原生 Docker / Podman / 竞品的功能对比,
>       作为长期路线图。
> 与本文的关系:
> - **[requirements.md](requirements.md)**: 当前范围与状态(简表)。
> - **[bugfix-requirements.md](bugfix-requirements.md)**: bug 修复台账(含修复记录)。
> - **[pending-bugs.md](pending-bugs.md)**: 当前未关闭的 BUG / 需求快照。
> - **[architecture.md](architecture.md)**: 实现架构。
>
> 本文只描述"目标是什么",不记录任务进度(进度一律走 pending-bugs.md / bugfix-requirements.md)。

## 1. 定位与目标

dtui 是一个**键盘优先**的 Docker / Podman 终端管理工具。设计目标:

1. **功能覆盖** 与 lazydocker (51k⭐, gocui) 看齐,同时保持 Bubbletea v2 现代化架构与双 runtime (Docker + Podman) 支持。
2. **终端体验**: 纯键盘可完成所有操作,鼠标支持为辅助。
3. **可发现性**: 多功能 Action Bar / 命令面板取代单字母绑定,降低快捷键记忆负担。
4. **跨 runtime 一致性**: 同一套 UI 同时支持 docker 与 podman,差异点收敛到 runtime 适配层。

## 2. 与原生 / 竞品对比矩阵

- ✅ = dtui 当前可用
- 🟡 = 设计目标(已纳入路线图,详见 §4)
- ❌ = 不实现 / 不在范围
- ⭐ = 用户在第二轮梳理中新增的目标

### 2.1 容器 (Container)

| 能力 | docker / podman 原生命令 | dtui | 状态 | 备注 |
|---|---|---|---|---|
| List | `docker ps` | 列表 + 状态色 | ✅ | |
| Start | `docker start` | `s` | ✅ | |
| Stop | `docker stop` | `Ctrl+S` | ✅ | 见 §5.2 |
| Restart | `docker restart` | `Ctrl+R` | ✅ | |
| Pause / Unpause | `docker pause` | `p` | ✅ | |
| Kill (强制) | `docker kill` | `Ctrl+K` | ✅ | 与 Stop 区别见 §5.2 |
| Remove | `docker rm` | `Ctrl+D` | ✅ | |
| Logs | `docker logs` | `l` | ✅ | 实时滚动 |
| Stats | `docker stats` | `m` | ✅ | 3s 周期 ticker,BR-031 待修体感 |
| Inspect (JSON / YAML) | `docker inspect` | `d` + `s` | ✅ | BR-007 |
| Exec (新 shell) | `docker exec -it sh` | `e` | ✅ | 页面 UI 待修 BR-038 |
| Attach (主进程 I/O) | `docker attach` | — | ⭐ | 待评估 §5.4 |
| Top (进程列表) | `docker top` | `:top` | ✅ | 仅命令面板,无直键 §5.6 |
| Port (端口映射) | `docker port` | `:port` | ✅ | 仅命令面板,无直键 §5.6 |
| Rename | `docker rename` | `:rename` | ✅ | 仅命令面板,无直键 §5.6 |
| Update (运行时改 mem/cpu/restart) | `docker update` | `Ctrl+W` | 🟡 | BR-033 |
| Diff (fs 改动) | `docker diff` | `Ctrl+?` (待重分配) | 🟡 | BR-033 |
| Export (fs → tar) | `docker export` | `Ctrl+?` (待重分配) | 🟡 | BR-033 |
| Commit (→ 新镜像) | `docker commit` | `Ctrl+?` (待重分配) | 🟡 | BR-033 |
| Wait (阻塞) | `docker wait` | `Ctrl+?` (待重分配) | 🟡 | BR-033 |
| Copy (容器 ↔ 主机) | `docker cp` | `Ctrl+?` (待重分配) | 🟡 | BR-033 §5.5 |

### 2.2 镜像 (Image)

| 能力 | docker / podman 原生命令 | dtui | 状态 | 备注 |
|---|---|---|---|---|
| List | `docker images` | 列表 | ✅ | |
| Inspect | `docker inspect` | `d` + `s` | ✅ | BR-007 |
| Pull | `docker pull` | `Ctrl+P` | ✅ | 进度条 |
| Push | `docker push` | `Ctrl+U` | ✅ | 进度条 |
| Tag | `docker tag` | `Ctrl+T` | ✅ | |
| Save (→ tar) | `docker save` | `Ctrl+E` | ✅ | §5.5 |
| Load (← tar) | `docker load` | `Ctrl+L` | ✅ | |
| Import (← tar,单层) | `docker import` | — | ⭐ | 新增 BR-036 |
| Build (← Dockerfile) | `docker build` | — | ❌ | 不在范围(参见 §3) |
| Search (Docker Hub) | `docker search` | — | ❌ | 不在范围 |
| History (顶层页) | `docker history` | `H` (镜像页) | ⭐ | BR-034,曾因性能取消,需重新设计 |
| Prune | `docker image prune` | `p` | ✅ | |
| Remove | `docker rmi` | `Ctrl+D` | ✅ | |

### 2.3 卷 (Volume)

| 能力 | docker / podman 原生命令 | dtui | 状态 | 备注 |
|---|---|---|---|---|
| List | `docker volume ls` | 列表 | ✅ | |
| Create | `docker volume create` | `c` | ✅ | 表单 |
| Inspect | `docker volume inspect` | `d` + `s` | ✅ | BR-009 部分修复 |
| Prune | `docker volume prune` | `p` | ✅ | |
| Remove | `docker volume rm` | `Ctrl+D` | ✅ | |

### 2.4 网络 (Network)

| 能力 | docker / podman 原生命令 | dtui | 状态 | 备注 |
|---|---|---|---|---|
| List | `docker network ls` | 列表 | ✅ | |
| Create | `docker network create` | `c` | ✅ | 表单 |
| Inspect | `docker network inspect` | `d` + `s` | ✅ | |
| Prune | `docker network prune` | `p` | ✅ | |
| Remove | `docker network rm` | `Ctrl+D` | ✅ | |
| Connect (容器接入网络) | `docker network connect` | — | ⭐ | BR-035 设计中,见 §5.3 |
| Disconnect (容器脱离网络) | `docker network disconnect` | — | ⭐ | BR-035 设计中,见 §5.3 |

### 2.5 系统 / Runtime

| 能力 | 原生命令 | dtui | 状态 | 备注 |
|---|---|---|---|---|
| 运行时切换 (多 daemon) | `docker context use` | `F2` | ✅ | |
| 刷新连接池 | — | `R` / `F12` | ✅ | 5s tick |
| Events 流(后台订阅) | `docker events` | 后台 | ✅ | 已订阅 |
| Events 流(独立页面) | `docker events` | `F3` | ⭐ | BR-035 配套,见 §5.7 |
| Registry Login | `docker login` | — | ⭐ | BR-037 |
| Disk usage | `docker system df` | — | ❌ | 不在范围 |
| Info / version | `docker info` | header | ✅ | |

## 3. 不在范围 / 明确不实现

以下能力**不纳入**设计目标,用户已明确决定:

| 能力 | 理由 |
|---|---|
| **Image Build** | 终端 UI 不适合 docker build 的长日志 / 多 stage 调试;走 IDE 或 CI |
| **Image Search** | Docker Hub 搜索意义不大,主流场景走私有仓库;走 CLI 即可 |
| **Volume Backup** | 业界也少做;用 `docker run --volumes-from` 临时挂载出 tar 即可 |
| **Container Attach** | 与 logs 高度重叠(只读排查场景 logs 已够用);唯一差异是 stdin 转发,在 dtui 只读场景下不常用;见 §5.4 |

新增 BR 条目时,默认不与上表冲突。

## 4. 待实现(设计目标)

每个能力详见对应 BR 条目。

| 编号 | 能力 | 优先级 | 关联 BR |
|---|---|---|---|
| 镜像 History 顶层页 | 用户浏览镜像 layer history | high | BR-034 |
| Events 独立面板 | F3 打开,浏览 runtime events 流 | medium | BR-035 |
| Image Import (tarball) | 镜像 tarball 导入(单层) | medium | BR-036 |
| Registry Login | docker / podman login 私库认证 | medium | BR-037 |
| Exec 页面 shell UI | 当前 exec 页面不像 shell 终端 | medium | BR-038 |
| Action Bar | 取代 Q 退出,每页多功能快捷面板 | medium | BR-039 |
| Dialog 风格统一 | 四周透明 + panel 居中 | medium | BR-040 |
| Network Connect / Disconnect | 容器接入/脱离网络 | medium | 待立 BR-035 |
| TASK-019 容器高级动作 | Copy/Update/Diff/Export/Commit/Wait | high | BR-033 |

## 5. 关键设计决策

### 5.1 镜像 History 顶层页 (H 键)

**历史背景**: BR-008 记录"H 键进 Help"曾因**性能问题**取消。当前 `keyboard.go:215` 把 H 绑到 `ToggleHeader`。

**取消原因**(根据上下文):镜像详情页内曾渲染完整 history,大镜像(>50 layer)下每次滚动都重新构造整段 section,导致渲染卡顿。

**重新设计**(BR-034):

- H 键在镜像页 → 进入**独立 History 顶层页**(`pages/history`),而不是在详情页内显示。
- History 顶层页只展示镜像 layer 列表(从 inspect API 拿 history 输出,见 `runtimeapi.ImageHistoryLayer`),不做详情分区。
- 复用 BR-007 / BR-016 的**按 revision 缓存文档 + 视口滚动只截可见行**机制,避免上次性能问题。
- 数据源:`Engine.Images().Inspect(...)` 的 history 字段,Podman / Docker 都已支持。
- 与详情页的约定:BR-008 第 3 条要求"镜像详情页不再渲染 History 段,仅保留其它字段"。

### 5.2 Stop vs Kill 语义(仅文档说明,无代码改动)

dtui 已实现的语义(**不需代码改动**,本节仅供用户参考与文档补充):

| | 实际行为 | runtime 调用 | docker 命令等价 |
|---|---|---|---|
| `Ctrl+S` (Stop) | SIGTERM,等待宽限期(默认 10 秒),超时自动 SIGKILL | `Engine.Containers().Stop(ctx, id, opts)` | `docker stop <ctr>` |
| `Ctrl+K` (Kill) | 立即发送 SIGKILL(或 `--signal` 自定义) | `Engine.Containers().Kill(ctx, id, signal)` | `docker kill <ctr>` |

**关键差异**:
- **Stop 优雅**:给进程机会清理(关闭文件句柄、刷写 buffer、通知其它服务),宽限期后强杀。
- **Kill 强制**:进程没有机会清理,适用于"进程卡死 / 不响应 SIGTERM / 必须立即释放资源"的场景。
- 两者都**不会**自动删除容器,只停止 PID 1。重启用 `docker start` 或 dtui 的 Start 动作。

**文档补充建议**(可选,不影响 BR):
- Help 页区分两者提示,例如:
  - `s Start`
  - `Ctrl+S Stop (SIGTERM, 10s grace)` 
  - `Ctrl+K Kill (SIGKILL 立即)`
- Footer 行动提示同样区分。
- 文档(本节已包含对比表,README / docs/navigation.md 可同步加一句)。

### 5.3 Network Connect / Disconnect

**应实现**(新增 BR 条目,与 BR-035 合并)。

Docker 语义:
```bash
docker network connect <network> <container> [--ip <ip>] [--alias <alias>]
docker network disconnect <network> <container> [--force]
```

dtui 设计:
- 在网络页(`ModeNetworks` 或 `ModeVolumes` 容器子视图):加 `+` / `-` 按钮或快捷键,弹出"选网络 / 选容器"对话框。
- 在容器详情 / inspect 视图:显示"已接入网络列表 + [断开]"交互。
- 详细 BR 条目待立项,本节先记录设计目标。

### 5.4 Container Attach vs Exec vs Logs

| | `docker logs` | `docker attach` | `docker exec` |
|---|---|---|---|
| **附加对象** | 不附加;读容器日志 | 容器 PID 1 主进程 | 容器内**新启动**的子进程 |
| **输出** | stdout/stderr 历史 | stdout/stderr 实时流 | 子进程 stdout/stderr |
| **stdin** | 不接收 | **可发送 stdin 到 PID 1** | 全新 stdin |
| **历史回看** | 有(`--since`/`--until`/`--tail`) | 无(attach 之后) | 无 |
| **多终端** | 可多次 logs | 互斥(只能一个 attach) | 可多个 exec |
| **退出语义** | Ctrl+C 中断显示,不杀容器 | Ctrl+P Ctrl+Q 脱离;`--sig-proxy=false` 时 Ctrl+C 不会杀容器 | exit 命令退出,不杀容器 |
| **dtui 已支持** | ✓ `l` 键 | ✗ | ✓ `e` 键(BR-038 待修 UI) |
| **适用场景** | 排查已发生问题 / 历史日志 | 实时追主进程 + stdin 交互(回应 `Continue? [y/N]`) | 在容器内任意命令操作 |

**关键洞察**(来自用户梳理):
- **`attach ≈ logs` 在只读场景下**:绝大多数排查只需要看容器主进程的输出,`docker logs` 已覆盖。
- **`attach` 唯一不可替代**:**向容器主进程 stdin 发输入**(回应交互式提示、输入密码等)。dtui 当前是只读排查工具,这类场景边缘。
- **`exec` 与前两者不同**:`exec` 是新启动子进程,提供"在容器内任意操作"的能力,不能被 logs 或 attach 替代。

**当前 dtui 决策**:
1. **修 Exec 页面 UI**(BR-038):当前实现不符合 shell 终端外观,用户反馈"不符合 shell 页面",需重新设计(全屏 terminal / 真实 PTY 透传 / 颜色与 ANSI 处理)。
2. **Attach 暂不实现**:与 logs 高度重叠;唯一差异(stdin 转发)在 dtui 只读排查场景下不常用;attach 的"接管 PID 1 + 信号转发"语义在 TUI 里容易让用户意外终止容器。**用户决定**:暂不纳入范围(用户已确认"attach ≈ 直接查看日志输出")。
3. 等 Exec 页面稳定、logs 已覆盖大部分场景后,**重新评估**:如果未来出现"必须向容器主进程 stdin 发输入"的需求(例如回应 apt 安装时的 y/N 提示),再立 Attach BR。

### 5.5 Ctrl+E 与 Container Export 的区别

| | `Ctrl+E` (当前) | Container Export (BR-033) |
|---|---|---|
| 绑定动作 | `ActionImageSave` | `ActionContainerExport` |
| 对象 | 镜像 | 容器运行时 fs |
| 命令等价 | `docker save -o file.tar image` | `docker export container > file.tar` |
| 保留 history | 是(layer history + tags) | 否(扁平化单层) |
| 导入方式 | `docker load` | `docker import` |

**结论**:**两者不冲突**。`Ctrl+E` 保留指向 `ImageSave`;`ActionContainerExport` 走 BR-033 的默认键位(`Ctrl+X`),文档中明确说明语义差异,避免 Help 误导。

### 5.6 Q 键退出 → 多功能 Action Bar

**当前**:
- `Q` = quit
- `Ctrl+C` = quit
- 命令面板 `:` 提供 rename/top/port/help/compose/images/containers/volumes/networks/logs

**用户决策**:取消 `Q` 作为退出键。改为每个 table 页提供**多功能 Action Bar**。

**设计**:
- 触发键:`;` (vim 风格命令模式) 或新增 `a` (action) 键 / 或扩展现有 `:` 命令面板。
- UI:弹出一个浮层,按当前面板动态显示可用动作。
- 入口分类:
  - **当前页专属**:镜像页的 `Tag/Push/Save/Load/Import/History`、容器页的 `Update/Export/Commit/Wait/Copy/Stats`、卷/网路页的 `Create/Connect/Prune` 等。
  - **通用**:Filter / Refresh / Help / Switch Runtime / Refresh Connections / Command palette。
- 每个条目显示**动作名 + 当前键位**(避免占新键,与现有绑定共存)。
- 无新动作的页面(`detail` / `logs` / `help`):**省略** Action Bar,直接显示无内容的提示。

**已有命令保留**:`:rename` / `:top` / `:port` 等命令面板入口不取消,与 Action Bar 并存(命令面板保留为"可输入"形态,Action Bar 保留为"可点击列表"形态)。

**实现位置**:新增 `internal/tui/ui/widget/actionbar/` 子包(暂定名),UI 类似 `widget/dialog/` 的浮层 + `widget/footer/` 的快捷键提示。

### 5.7 Events 流独立面板 (F3)

**当前**:`EventService.Subscribe` 已在 runtime 层实现,后台订阅并合并到资源更新路径。但**没有独立的 Events 浏览面板**。

**设计**:
- `F3` 打开 `ModeEvents` / `PanelEvents`。
- UI:类似日志页的实时流,展示最近 N 条 events(可滚动查看历史,默认保留最近 1000 条)。
- 支持 `Filter`:按类型 (container / image / network / volume)、按 action (start / stop / create / destroy / pull / push 等)、按时间范围。
- 支持 `Pause / Resume`:`space` 暂停滚动累积,新事件进入缓冲队列。
- 数据源:`Engine.Events().Subscribe(...)` 返回的 `<-chan EventItem`。
- 后端 runtime 已在 `update/update_events.go` 实现订阅,可复用。

### 5.8 Dialog 风格统一(四周透明 + panel 居中)

**用户决策**:所有 dialog(confirm / shell / filter / exec / notification / selection 等)统一为:
- **四周透明**(不是全屏覆盖)
- **中间窗口**(实际 dialog 内容)
- **位置:panel 居中**(作用域是当前 panel)

**当前实现**(分散):
- `widget/dialog/overlay.go`:overlay / 对话框可能有不同策略。
- `confirm` / `notification` / `selection` / `exec`:各自有不同的尺寸与位置计算。

**统一设计**:
- 共享 `widget/dialog/centered.go`(新文件)提供 `CenterOnPanel(panel Rect, content string, w, h int) string`。
- 所有 dialog 渲染时调用 `CenterOnPanel` 而不是各自计算坐标。
- panel 的 `bodyRect` 由 `LayoutReport.Panel` 提供(在 BR-012 wontfix 之后,`Panel` 字段可保留用于中心化)。
- 透明 = 整个 dialog 边框外不绘制背景(`Render` 不带 background fill,只画边框)。
- dialog 内部仍然按需要带背景。

**应用范围**:confirm / shell / exec / filter / selection / notification / 自定义表单(Copy / Update / Commit / Wait / Import / Login)。

### 5.9 Registry Login (BR-037)

Docker / Podman 原生命令:
```bash
docker login [SERVER] -u USER -p [PASSWORD]
# 或交互式
docker login
# 配置存储到 ~/.docker/config.json
```

dtui 设计:
- 在 Action Bar 中加入 "Login" 动作,触发表单:
  - 服务器 URL (默认 `docker.io` 或 `https://index.docker.io/v1/`)
  - 用户名
  - 密码(掩码输入)
- 提交后调 `Engine.Login(ctx, server, user, password)`(runtime 层需新增)。
- Podman 兼容:podman login 与 docker login 配置共用 `~/.docker/config.json`,可直接复用。
- 凭证存储:**不**存储密码到 dtui 配置;只调用 engine 登录并依赖引擎自身的凭证存储。

### 5.10 Image Import (BR-036)

Docker 原生命令:
```bash
docker import [OPTIONS] FILE|URL|- [REPOSITORY[:TAG]]
# 例: docker import ./container-flat.tar my-image:latest
```

与 `docker load` 区别:
- `load`:从 **docker save 导出**的多 layer tarball 恢复(保留 history)。
- `import`:从 **任意 tarball**(包括 `docker export` 的扁平 tar)导入为**单层新镜像**。

dtui 设计:
- Action Bar 中加 "Import tarball" 动作,触发表单:
  - 源 tarball 路径(`/path/to/file.tar` 或 URL)
  - 目标 repository:tag(可选)
- 提交后调 `Engine.Images().Import(ctx, source, ref)`(runtime 需新增)。
- 与 Save / Load 并列,但默认键位**不冲突**(走 Action Bar,避免 Ctrl+E / Ctrl+L 之类的键位挤兑)。

## 6. 优先级路线

| 阶段 | 目标 | 期限 |
|---|---|---|
| **P0** | TASK-019 容器高级动作 (BR-033) | 短期 |
| **P1** | Image History 顶层页 (BR-034) / Network Connect (BR-035) | 短期 |
| **P2** | Exec 页面 shell UI (BR-038) / Dialog 统一风格 (BR-040) / Image Import (BR-036) | 中期 |
| **P3** | Registry Login (BR-037) / Action Bar 取代 Q (BR-039) / Events 独立面板 (BR-035 配套) | 中长期 |

## 7. 跟踪约定

- **本文档**:长期目标与设计决策,修改需明确说明变更原因。
- **bugfix-requirements.md**:每次状态变更都要更新;修复记录(reproducer + commit hash + tests)是真相来源。
- **pending-bugs.md**:开发排期入口,BR 升级为 done / wontfix 后从本表移除。
- **requirements.md**:能力清单简表,与本文档 §2 / §4 互为指针。
- **ui-design.md** / **i18n.md** / **navigation.md**:具体子领域设计。

任何新增能力应当:
1. 先在本文档 §2 矩阵中标记 🟡(设计目标);
2. 立 BR 条目到 `bugfix-requirements.md`,状态 `open`;
3. 同步加入 `pending-bugs.md` 排序表;
4. 实现完成后:状态 → `done`,在矩阵中改回 ✅,补"修复记录"。