# Image Size 差异分析结论

## 根因

**TUI `FormatSize` 使用 1024-base（二进制），而 `podman images` CLI 使用 1000-base（SI 十进制）。**

|        | TUI `FormatSize`    | `podman images` CLI |
|--------|---------------------|---------------------|
| 1 KB = | 1,024 bytes         | 1,000 bytes         |
| 1 MB = | 1,048,576 bytes     | 1,000,000 bytes     |
| 1 GB = | 1,073,741,824 bytes | 1,000,000,000 bytes |

## 数据链路

两者从同一个数据源读取 Size 字段：

```
Podman libpod API  /libpod/images/json
  → JSON.Size (int64, 原始字节数)
    → TUI:    FormatSize(raw)   → 1024-base  → "40.1 MB"
    → CLI:    podman images     → 1000-base  → "42 MB"
```

两次测试均确认：API 返回的 `Size` 原始字节值是相同的。差异仅在于格式化时的进制选择。

## 差异量级

| 量级 | 差异     | 示例                           |
|----|--------|------------------------------|
| KB | ~2.34% | 100 KB (SI) = 97.7 KiB (bin) |
| MB | ~4.63% | 100 MB (SI) = 95.4 MiB (bin) |
| GB | ~6.87% | 1.1 GB (SI) = 1.03 GiB (bin) |

## 在测试机器上的实际对比

```
Repo:Tag                                        | Raw Bytes    | CLI 显示 | TUI(1024) | SI(1000)  | 匹配
------------------------------------------------|--------------|----------|-----------|-----------|------
alpine:3.19                                     |    7,687,286 |  7.69 MB |    7.3 MB |  7.69 MB  | CLI ✓
redis:7-alpine                                  |   42,017,397 |   42 MB  |   40.1 MB | 42.02 MB  | (rounding)
nginx:alpine                                    |   50,143,357 |  50.1 MB |   47.8 MB | 50.14 MB  | (rounding)
postgres:18.4-alpine                            |  288,681,577 |  289 MB  |  275.3 MB | 288.68 MB | (rounding)
```

"匹配"列：alpine 精确匹配 SI 值；其余 CLI 做了整数舍入（如 42.02 → 42）。

## 验证方式

```bash
# Shell 测试（需 Podman socket）
bash test/diagnostics/image-size/compare.sh

# Go 测试
go test ./test/diagnostics/image-size/ -v
```

## 配置方式

程序内嵌默认 1024-base（binary）。用户可在 `~/.config/docker-tui/config.yml` 中覆盖：

```yaml
general:
  sizeFormat: "si"    # "binary"（默认，1024-base）或 "si"（1000-base，与 CLI 一致）
```

设置 `"si"` 后，TUI 显示的镜像大小将使用 1000-base（MB/GB），与 `podman images` / `docker images` CLI 输出一致。

## 实现变更

| 文件                                   | 修改                                                                              |
|--------------------------------------|---------------------------------------------------------------------------------|
| `internal/data/config/types.go`      | `GeneralConfig` 新增 `SizeFormat string`                                          |
| `internal/data/config/default.jsonc` | `general.sizeFormat` 默认 `"binary"`                                              |
| `internal/tui/utils/format.go`       | `FormatSize` 支持根据 `SetSizeFormat()` 切换 1024/1000-base，输出 KiB/MiB/GiB 或 kB/MB/GB |
| `cmd/docker-tui/main.go`             | 启动时根据 `cfg.General.SizeFormat` 调用 `utils.SetSizeFormat()`                       |
