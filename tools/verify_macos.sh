#!/usr/bin/env bash
# macOS 交叉編產物的靜態驗收。
#
#   tools/verify_macos.sh <binary>
#
# **Linux 上執行不了 macOS binary**，所以這裡只驗結構。全過只代表「不會因為結構
# 問題開不起來」，**不代表功能正常**——真機驗收要另外做。
#
# 四道：
#   1. 雙弧（arm64 ＋ x86_64）。
#   2. arm64 必須有 `LC_CODE_SIGNATURE`。ld64 連結 arm64 時會自己補 ad-hoc 簽章，
#      少了它使用者那邊會看到 `Killed: 9`（x86_64 沒這限制）。
#   3. 最低系統版本讀得出來。
#   4. 動態相依只能落在 `/usr/lib/` 與 `/System/Library/`；其他都是連到編譯機。
#      `otool -L` 對 fat binary 會為每個架構印檔名標頭，所以要先 `lipo -thin`。
set -euo pipefail

BIN="${1:?用法：verify_macos.sh <binary>}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${KOL_MAC_IMAGE:-kol-osxcross-go:20260907}"
case "$BIN" in /*) REL="${BIN#"$REPO_ROOT"/}" ;; *) REL="$BIN" ;; esac

exec timeout 300 docker run --rm --network none \
  --memory 2g --cpus 2 --pids-limit 128 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" -e HOME=/tmp -e "KOL_BIN=$REL" \
  -v "$REPO_ROOT:/src:ro" -w /src \
  --tmpfs "/tmp:exec,uid=$(id -u),gid=$(id -g),size=512m" \
  "$IMAGE" \
  bash -c '
    set -euo pipefail
    eval "$(osxcross-conf)"
    T=x86_64-apple-$OSXCROSS_TARGET
    B=/src/$KOL_BIN
    info=$($T-lipo -info "$B")
    echo "$info"
    for a in arm64 x86_64; do
      case "$info" in *" $a"*) ;; *) echo "缺少架構 $a" >&2; exit 1 ;; esac
    done
    for a in arm64 x86_64; do
      $T-lipo -thin $a "$B" -output /tmp/$a
      # ⚠ 先整份收下來再比，**不要 `otool | grep -q`**：`grep -q` 找到就結束，
      # 上游的 otool 收到 SIGPIPE 回 141，`pipefail` 把整條管線判成失敗——
      # 於是「簽章明明在」也會噴錯。機器閒的時候 otool 早就寫完不會踩到，
      # 一忙就必中，是最難查的那種偶發。
      load=$($T-otool -l /tmp/$a)
      if [ "$a" = arm64 ]; then
        case "$load" in
          *LC_CODE_SIGNATURE*) ;;
          *) echo "arm64 沒有 LC_CODE_SIGNATURE（使用者會看到 Killed: 9）" >&2; exit 1 ;;
        esac
      fi
      echo "-- $a $(printf "%s\n" "$load" | grep -m1 minos || echo "（沒有 minos）")"
      libs=$($T-otool -L /tmp/$a)
      bad=$(printf "%s\n" "$libs" | tail -n +2 | awk "{print \$1}" \
            | grep -vE "^(/usr/lib/|/System/Library/)" || true)
      if [ -n "$bad" ]; then
        echo "$a 連到系統目錄以外的動態庫：" >&2; echo "$bad" >&2; exit 1
      fi
    done
    echo "四道全過（結構驗收；不等於功能驗收）"
  '
