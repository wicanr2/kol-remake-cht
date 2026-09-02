#!/usr/bin/env bash
# 在鎖定的 IDA Pro 9.4 image 裡跑一支 IDAPython 探針。
#
#   tools/run_ida_probe.sh tools/ida_probe_koltown_world_map.py workplace/.../out.json
#
# 只有輸出的 JSON 才是證據；`idat` 的離開碼不能當判準（見 kb retro/ida-pro-9.4）。
# `HOME` 必須是 image 裡那個帶 `.idapro` 的家目錄，否則 IDAPython 與授權都起不來，
# 而 `idat` 只會在 `-L` 指定的 log 裡講，標準輸出是空的。
set -euo pipefail

script=${1:?用法：run_ida_probe.sh <探針.py> <輸出.json>}
out=${2:?用法：run_ida_probe.sh <探針.py> <輸出.json>}

root=$(cd "$(dirname "$0")/.." && pwd)
work=$root/workplace/analysis/koltown
image=${IDA_IMAGE:-ida-pro-9.4-idapython:locked-v1}

cp "$root/$script" "$work/$(basename "$script")"
rm -f "$work/probe-out.json"
timeout 1800 docker run --rm --network none \
  --memory 6g --cpus 2 --pids-limit 256 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" -e HOME=/home/ubuntu \
  -v "$work:/work" -w /work "$image" \
  idat -A -L/work/ida.log \
  -S"/work/$(basename "$script") /work/probe-out.json /work/KOLTOWN-unpacked.EXE" \
  /work/KOLTOWN-unpacked-ida94-derived.i64 || true

if [ ! -s "$work/probe-out.json" ]; then
  echo "探針沒有產出 JSON；IDA log 尾端：" >&2
  tail -20 "$work/ida.log" >&2 || true
  exit 1
fi
mkdir -p "$(dirname "$root/$out")"
mv "$work/probe-out.json" "$root/$out"
echo "$out"
