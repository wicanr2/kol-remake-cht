#!/usr/bin/env bash
# `check_repo_visibility.sh` 的回歸測試：用假的 gh 走過每一條路，確認失敗即關閉。
#
#   tools/test_check_repo_visibility.sh
set -euo pipefail
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP:?}"' EXIT
fail=0

# 假的 gh：輸出由 FAKE_GH_OUT 決定，FAKE_GH_RC 決定結束碼。
cat > "$TMP/gh" <<'G'
#!/usr/bin/env bash
printf '%s\n' "${FAKE_GH_OUT:-}"
exit "${FAKE_GH_RC:-0}"
G
chmod +x "$TMP/gh"

run() { # 名稱 期望結束碼 [環境變數…]
  local name="$1" want="$2"; shift 2
  local rc=0
  env "$@" PATH="$TMP:$PATH" KOL_REPO_SLUG=owner/name "$REPO_ROOT/tools/check_repo_visibility.sh" >/dev/null 2>&1 || rc=$?
  if [ "$rc" != "$want" ]; then
    echo "不通過：$name（結束碼 $rc，期望 $want）" >&2; fail=1
  else
    echo "通過：$name"
  fi
}

run "public 放行" 0 FAKE_GH_OUT=public FAKE_GH_RC=0
run "private 擋下" 4 FAKE_GH_OUT=private FAKE_GH_RC=0
run "internal 擋下" 4 FAKE_GH_OUT=internal FAKE_GH_RC=0
run "空輸出擋下" 4 FAKE_GH_OUT= FAKE_GH_RC=0
run "gh 出錯擋下" 4 FAKE_GH_OUT=public FAKE_GH_RC=1

# 沒有 gh：PATH 只放腳本自己要用的 dirname，確認訊息是「找不到 gh」而不是別的錯誤。
mkdir -p "$TMP/nogh"
ln -s "$(command -v dirname)" "$TMP/nogh/dirname"
rc=0
err="$(env PATH="$TMP/nogh" KOL_REPO_SLUG=owner/name "$(command -v bash)" "$REPO_ROOT/tools/check_repo_visibility.sh" 2>&1 >/dev/null)" || rc=$?
if [ "$rc" = 4 ] && printf '%s' "$err" | grep -q '找不到 gh'; then
  echo "通過：沒有 gh 擋下"
else
  echo "不通過：沒有 gh 時結束碼 $rc、訊息「$err」" >&2; fail=1
fi

exit "$fail"
