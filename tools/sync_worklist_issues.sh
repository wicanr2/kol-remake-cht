#!/usr/bin/env bash
# 把 docs/worklist.json 鏡像到 GitHub issue：每一條一張，已完成的關掉。
#
#   tools/sync_worklist_issues.sh            # 同步
#   DRY_RUN=1 tools/sync_worklist_issues.sh  # 只印要做什麼，不碰 GitHub
#
# **權威是 JSON**。issue 的標題、內文、label（只限 worklist／分層／manual 這一組）
# 與開關狀態都照 JSON 蓋過去；別人手加的 label 與留言不動。在 GitHub 上改內文
# 不會回流。
#
# 分工：內容由 `worklist.py issues` 在容器裡產生、建好的編號也由它寫回 JSON
# （條目的 `issue` 欄位）；這支只在主機上呼叫 `gh`——GitHub 控制面本來就在主機。
#
# 中途斷掉可以重跑：每建好一張就先記進 `created.tsv`，下一次開頭先把它寫回 JSON，
# 所以不會重複開同一條。
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"
WORK="workplace/worklist-issues"
PLAN="$WORK/plan.json"
CREATED="$WORK/created.tsv"
REPO="${KOL_ISSUE_REPO:-$(gh repo view --json nameWithOwner -q .nameWithOwner)}"
DRY="${DRY_RUN:-}"
PAUSE="${KOL_ISSUE_PAUSE:-1}"   # 每次寫入之間停一下，避開 GitHub 的次級速率限制

command -v gh >/dev/null && command -v jq >/dev/null || { echo "要有 gh 與 jq" >&2; exit 2; }
mkdir -p "$WORK"

py() { KOL_PY_TIMEOUT=120 tools/run_python_tool.sh tools/worklist.py "$@"; }

adopt() {
  if [ -s "$CREATED" ]; then
    py issues --adopt "$CREATED"
    rm -f "$CREATED"
  fi
}

run() {
  if [ -n "$DRY" ]; then
    printf '  (dry-run)'; printf ' %q' "$@"; echo
  else
    "$@"
    sleep "$PAUSE"
  fi
}

adopt   # 上一次斷在半路時建好的那幾張
py issues --out "$PLAN"
echo "同步到 $REPO"

# label：分層、manual、worklist。--force 會把顏色與說明更新成 JSON 的版本。
mapfile -t MANAGED < <(jq -r '.repo_labels[].name' "$PLAN")
jq -c '.repo_labels[]' "$PLAN" | while read -r l; do
  run gh label create "$(jq -r .name <<<"$l")" --repo "$REPO" --force \
    --color "$(jq -r .color <<<"$l")" --description "$(jq -r .description <<<"$l")"
done

body="$WORK/body.md"
created=0; updated=0
while read -r row; do
  id=$(jq -r .id <<<"$row")
  num=$(jq -r '.number // empty' <<<"$row")
  title=$(jq -r .title <<<"$row")
  state=$(jq -r .state <<<"$row")
  mapfile -t want < <(jq -r '.labels[]' <<<"$row")
  jq -r .body <<<"$row" >"$body"

  if [ -z "$num" ]; then
    args=(gh issue create --repo "$REPO" --title "$title" --body-file "$body")
    for l in "${want[@]}"; do args+=(--label "$l"); done
    if [ -n "$DRY" ]; then
      run "${args[@]}"
      continue
    fi
    url=$("${args[@]}")
    num="${url##*/}"
    printf '%s\t%s\n' "$id" "$num" >>"$CREATED"
    echo "  #$num ← $id"
    created=$((created + 1))
    sleep "$PAUSE"
    [ "$state" = closed ] && run gh issue close "$num" --repo "$REPO" --reason completed
    continue
  fi

  # 已經有編號：標題、內文、label、開關狀態逐項對齊。
  [ -n "$DRY" ] && { echo "  (dry-run) 更新 #$num ← $id"; continue; }
  cur=$(gh issue view "$num" --repo "$REPO" --json state,labels,title,body)
  edit=(gh issue edit "$num" --repo "$REPO")
  [ "$(jq -r .title <<<"$cur")" != "$title" ] && edit+=(--title "$title")
  [ "$(jq -r .body <<<"$cur")" != "$(cat "$body")" ] && edit+=(--body-file "$body")
  mapfile -t have < <(jq -r '.labels[].name' <<<"$cur")
  for l in "${want[@]}"; do
    [[ " ${have[*]} " == *" $l "* ]] || edit+=(--add-label "$l")
  done
  for l in "${have[@]}"; do
    [[ " ${MANAGED[*]} " == *" $l "* ]] || continue        # 別人加的不動
    [[ " ${want[*]} " == *" $l "* ]] || edit+=(--remove-label "$l")
  done
  if [ "${#edit[@]}" -gt 6 ]; then
    run "${edit[@]}"
    updated=$((updated + 1))
  fi
  now=$(jq -r .state <<<"$cur" | tr '[:upper:]' '[:lower:]')
  if [ "$now" != "$state" ]; then
    if [ "$state" = closed ]; then
      run gh issue close "$num" --repo "$REPO" --reason completed
    else
      run gh issue reopen "$num" --repo "$REPO"
    fi
    updated=$((updated + 1))
  fi
done < <(jq -c '.issues[]' "$PLAN")
rm -f "$body"

adopt
[ -z "$DRY" ] && py render >/dev/null
echo "新開 $created 張、更新 $updated 處"
