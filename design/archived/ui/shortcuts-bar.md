# Shortcuts Bar — 底部快捷键栏

## 状态

- [x] 已实现 (常驻显示)

## 设计原则

- 始终可见 (k9s 风格)
- 根据当前面板/模式动态切换内容
- 分隔符 `│` 用表面色渲染

## 各模式下快捷键

### 容器面板 (默认)

```
j↓ Down │ k↑ Up │ Tab Next │ Space Mark │ s Start │ S Stop │ R Restart │ l Logs │ m Stats │ d Detail │ Ctrl+D Del │ c Connect │ / Filter │ ? Help │ q Quit
```

### 镜像面板

```
j↓ Down │ k↑ Up │ Tab Next │ Space Mark │ P Pull │ p Prune │ o Sort │ Enter Containers │ d Detail │ D Debug │ E Export │ Ctrl+D Del │ c Connect │ / Filter │ ? Help │ q Quit
```

### 卷/网络面板

```
j↓ Down │ k↑ Up │ Tab Next │ Space Mark │ Ctrl+D Del │ c Connect │ / Filter │ ? Help │ q Quit
```

### 日志视图

```
Esc Back │ j/k Scroll │ PgUp/Dn Page │ g/G Top/Bottom
```

### 详情视图

```
Esc/Enter Back
```

### 确认模式

```
y Confirm │ n Cancel
```

### 帮助视图

```
? Close
```

## 有标记项时

当 `MarkedIDs` 不为空时，追加批量删除按钮:

```
... │ Ctrl+D Del 3 │ ...
```

## 样式

| 元素 | 样式 |
|---|---|
| 背景 | 主题 Dark 色 |
| 快捷键 (字母) | ShortcutKeyStyle (白色) |
| 描述 | ShortcutDescStyle (灰色) |
| 分隔符 | ShortcutSepStyle (主题 Surface 色) |
