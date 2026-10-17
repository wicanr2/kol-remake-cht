#!/usr/bin/env bash
# 發行前的儲存庫可見度閘門：預設失敗即關閉。
#
#   tools/check_repo_visibility.sh
#
# 專案契約：2026-10-17使用者明確決定現有儲存庫公開，遊戲用HD全量收錄。
# 範圍、原權利限制與歷史清冊見docs/RELEASE-RIGHTS-REVIEW.md及
# docs/verification/public-repository-20261017.json；公開收錄不授予原作權利。
# 這一支在打包與驗收之前問GitHub目前的可見度：
#
#   - 查得到而且是 public：通過（印一行，結束碼 0）。
#   - 查不到（沒有 gh、沒登入、沒網路、API 回錯）：結束碼 4。不猜，也不放行。
#   - 查得到但不是 public：結束碼 4，並指向目前公開化決定。
#
# 沒有「略過」旗標。再次改變公開方式前須依AGENTS與權利審查記錄新決定。
# 測試用：`KOL_REPO_SLUG` 換掉從 origin 推出來的 `owner/name`，`PATH` 前面放一支假的 gh。
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

slug="${KOL_REPO_SLUG:-}"
if [ -z "$slug" ]; then
  url="$(git -C "$REPO_ROOT" remote get-url origin 2>/dev/null)" || {
    echo "查不到 origin，無法確認儲存庫可見度（失敗即關閉）" >&2; exit 4; }
  slug="$(printf '%s\n' "$url" | sed -E 's#^(git@github\.com:|https://github\.com/)##; s#\.git$##')"
fi
case "$slug" in
  */*) ;;
  *) echo "無法從 origin 讀出 owner/name：$slug（失敗即關閉）" >&2; exit 4 ;;
esac

command -v gh >/dev/null 2>&1 || {
  echo "找不到 gh，無法確認 $slug 的可見度（失敗即關閉）" >&2; exit 4; }

vis="$(gh api "repos/$slug" --jq .visibility 2>/dev/null)" || {
  echo "查不到 $slug 的可見度（沒登入、沒網路或 API 出錯，失敗即關閉）" >&2; exit 4; }

if [ "$vis" != public ]; then
  echo "$slug 的可見度是「$vis」，不符合目前public決定。請核對docs/RELEASE-RIGHTS-REVIEW.md與公開化清冊，發行流程先停下來" >&2
  exit 4
fi
echo "儲存庫 $slug 可見度 public"
