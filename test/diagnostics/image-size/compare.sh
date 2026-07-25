#!/usr/bin/env bash
# compare.sh — 对比 TUI FormatSize (1024-base) vs podman CLI (1000-base)
#
# Usage: just test-image-size-live
#
# 通过 Docker-compatible API 获取原始 Size 字节值，
# 然后分别用 1024-base (TUI) 和 1000-base (CLI) 格式化，
# 与 podman images 的输出对比。

set -euo pipefail

SOCK="${XDG_RUNTIME_DIR:-/run/user/1000}/podman/podman.sock"

if [ ! -S "$SOCK" ]; then
  echo "ERROR: Podman socket not found at $SOCK"
  exit 1
fi

echo "=== 1. 通过 API 获取原始 Size（Docker-compatible API）==="
curl -sf --unix-socket "$SOCK" http://localhost/v5.0.0/images/json > /tmp/podman_api_images.json
echo "  OK ($(wc -c < /tmp/podman_api_images.json) bytes)"

echo "=== 2. podman images CLI 输出 ==="
podman images --format '{{.ID}}\t{{.Repository}}:{{.Tag}}\t{{.Size}}' > /tmp/podman_cli_images.txt
cat /tmp/podman_cli_images.txt

echo ""
echo "=== 3. 对比表 ==="
printf "%-55s | %10s | %10s | %10s | %10s | %s\n" \
  "Repo:Tag" "Raw Bytes" "CLI显示" "TUI(1024)" "SI(1000)" "匹配"
printf "%-55s-|-%10s-|-%10s-|-%10s-|-%10s-|-%s\n" \
  "-------------------------------------------------------" "----------" "----------" "----------" "----------" "------"

python3 << 'PYEOF'
import json

with open('/tmp/podman_api_images.json') as f:
    raw = json.load(f)

api_map = {}
for img in raw:
    size = img.get('Size', 0)
    for tag in img.get('RepoTags', []):
        api_map[tag] = size
    img_id = img.get('Id', '')
    api_map[img_id] = size

def format_1024(b):
    if b < 1024:
        return f'{b} B'
    div = 1024
    exp = 0
    n = b // 1024
    while n >= 1024:
        div *= 1024
        exp += 1
        n //= 1024
    return f'{b/div:.1f} {"KMGTPE"[exp]}B'

def format_1000(b):
    units = ['B', 'kB', 'MB', 'GB', 'TB']
    val = float(b)
    i = 0
    while val >= 1000 and i < len(units) - 1:
        val /= 1000
        i += 1
    if i == 0:
        return f'{int(val)} {units[i]}'
    s = f'{val:.2f} {units[i]}'
    if s[-6:] == '.00 ':
        s = s[:-6] + ' ' + s[-3:]
    return s

with open('/tmp/podman_cli_images.txt') as f:
    for line in f:
        line = line.strip()
        if not line:
            continue
        parts = line.split('\t')
        if len(parts) < 3:
            continue
        img_id = parts[0][:12]
        tag = parts[1]
        cli_size_str = parts[2]

        raw_bytes = api_map.get(tag) or api_map.get(parts[0], 0)

        if raw_bytes == 0:
            print(f'{tag[:55]:55s} | {"N/A":>10s}')
            continue

        tui_str = format_1024(raw_bytes)
        si_str = format_1000(raw_bytes)

        if tui_str == cli_size_str:
            match = 'TUI'
        elif si_str == cli_size_str:
            match = 'CLI'
        else:
            match = 'NEITHER'

        print(f'{tag[:55]:55s} | {raw_bytes:>10d} | {cli_size_str:>10s} | {tui_str:>10s} | {si_str:>10s} | {match}')
PYEOF

echo ""
echo "=== 4. 结论 ==="
echo "TUI FormatSize 使用 1024-base (binary): 1 MB = 1048576 bytes"
echo "podman CLI    使用 1000-base (SI):      1 MB = 1000000 bytes"
python3 -c "print(f'差异程度: {abs(1-1000/1024)*100:.2f}% (KB), {abs(1-1000**2/1024**2)*100:.2f}% (MB), {abs(1-1000**3/1024**3)*100:.2f}% (GB)')"

rm -f /tmp/podman_api_images.json /tmp/podman_cli_images.txt
