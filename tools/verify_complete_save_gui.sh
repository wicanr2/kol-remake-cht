#!/usr/bin/env bash
# 僅限 Docker；/src 唯讀，/out 可寫且已有 kol。
set -euo pipefail
test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
export HOME=/tmp DISPLAY=:95
Xvfb :95 -screen 0 1600x1000x24 -nolisten tcp >/out/xvfb.log 2>&1 &
trap 'kill $(jobs -pr) 2>/dev/null || true' EXIT
cd /src
for n in $(seq 1 50); do [ -e /tmp/.X11-unix/X95 ] && break; sleep .1; done
start() {
 log="/out/$1.log";shift
 /out/kol -seed 1 -log-input "$@" >"$log" 2>&1 & pid=$!
 win=""
 for n in $(seq 1 100); do win=$(xdotool search --onlyvisible --name . 2>/dev/null | head -1 || true); [ -n "$win" ] && break; sleep .1; done
 test -n "$win" || { echo "視窗沒有出現：$log（偶發時重跑；連續失敗再查）" >&2; exit 1; }
 xdotool windowfocus "$win";sleep 3
}
key() {
 before=$(grep -c '^input scene=' "$log" || true)
 xdotool keydown --window "$win" "$1";sleep .15;xdotool keyup --window "$win" "$1"
 for n in $(seq 1 60); do after=$(grep -c '^input scene=' "$log" || true);[ "$after" -gt "$before" ] && break;sleep .1;done
 test "$after" -gt "$before";sleep .25
}
shot() { import -window "$win" "/out/$1.png"; }
stop() { xdotool keydown --window "$win" ctrl+q;wait "$pid";sleep .5; }
create() {
 key Return
 for c in "$@";do key "$c";done
 key Return;key Return;key Return;key Return;key Down;key Return;key Return
}
start creation -save /out/created.json
key Return;key Return;key 5;key Down
create N O V A
shot created-first
cp /out/created.json /out/created-first.json
create L U N A
shot created-second
stop
python3 - <<'PY'
import json
from pathlib import Path
a=json.loads(Path('/out/created-first.json').read_text());b=json.loads(Path('/out/created.json').read_text())
assert a['schema']=='kol-save-v2' and sum(x is not None for x in a['characters'])==15
assert sum(x is not None for x in b['characters'])==16
assert b['characters'][14]['name']=='NOVA' and b['characters'][15]['name']=='LUNA'
assert b['characters'][14]['id']!=b['characters'][15]['id']
assert not b['party'] and b['speaker']==-1
PY
start reload -load /out/created.json -save /out/created.json -faithful-inn-stay
shot reloaded-menu
key Escape;key 6
for n in $(seq 1 14);do key Down;done
key Return;key space
shot new-character-inn
key less;key Return;key Return
recruit_down=$(python3 - <<'PY'
import json
from pathlib import Path
s=json.loads(Path('/out/created.json').read_text())
names=[c['name'] for c in s['characters'] if c is not None and c['name']!='NOVA'
       and c['place']=='TROLLSBANE' and c['town_code']=='BRT']
print(names.index('LUNA'))
PY
)
for n in $(seq 1 "$recruit_down");do key Down;done
key Return
shot two-new-members
# REST 是倒數第二格；住宿忠實模式逐人問價錢及存檔。
key less;key less;key Return
shot stay-whole-party
key Return
shot stay-offer
key Return;key Return
cp /out/created.json /out/after-first-stay.json
shot first-retired
key Return;key Return
shot retired
stop
python3 - <<'PY'
import json
from pathlib import Path
s=json.loads(Path('/out/created.json').read_text())
assert not s['party'] and s['speaker']==-1
c=s['characters'][14]
assert c['place']=='TROLLSBANE' and c['town_code']=='BRT'
assert c['gold']<1500
assert s['characters'][15]['gold']==1440
assert s['settings']['faithful_inn_stay']
first=json.loads(Path('/out/after-first-stay.json').read_text())
assert first['party']==[15] and first['characters'][14]['gold']==1440
assert first['characters'][15]['gold']==1500
PY
echo '正常建角兩人、重啟選角與忠實住宿保存通過'

# 住宿離隊之後重新啟動，兩名新角色仍能從名冊選進旅店（#13「離開重載仍可選」）。
# 名冊第 15 格是 NOVA、第 16 格是 LUNA；選中之後應進入城鎮畫面。
for pick in 14 15; do
  start reselect-$pick -load /out/created.json
  key Escape;key 6
  for n in $(seq 1 $pick);do key Down;done
  key Return;key space
  shot reselected-$pick
  stop
  grep -q '^input scene=town' /out/reselect-$pick.log
done
echo '住宿離隊後重載，兩名新角色都能再選進旅店'

cp /out/created.json /out/before-delete.json
start delete -load /out/created.json -save /out/created.json
key Down;key Down;key Return
for n in $(seq 1 14);do key Down;done
key Return;key Y
shot deleted
cp /out/created.json /out/deleted.json
key Escape;key Up
create N O V A 2
shot reused-slot
# 全滿時不能再進建角。
key Return
shot full-roster
stop
python3 - <<'PY'
import json
from pathlib import Path
before=json.loads(Path('/out/before-delete.json').read_text())
deleted=json.loads(Path('/out/deleted.json').read_text())
after=json.loads(Path('/out/created.json').read_text())
assert deleted['characters'][14] is None and deleted['characters'][15]['name']=='LUNA'
assert after['characters'][14]['name']=='NOVA2'
assert after['characters'][14]['id']!=before['characters'][14]['id']
assert sum(c is not None for c in after['characters'])==16
PY
echo '刪除不搬槽、重用空槽產生新身分及滿槽流程通過'

# v1 相容性採明確測試輸入，不冒充原版存檔。
python3 - <<'PY'
import json
from pathlib import Path
r=json.loads(Path('/src/assets/original-derived-data/character-roster.json').read_text())
c=next(x for x in r['slots'] if x['name']=='Theodoric')
s=dict(schema='kol-save-v1',saved_at='2026-09-08T00:00:00Z',origin={'kind':'remake'},
 settings={'faithful_inn_stay':False,'future_setting':{'retained':True}},
 world={'x':40,'y':21,'sub_x':3,'sub_y':3},
 location={'kind':'town','town':7,'entry_point':3,'x':57,'y':8},
 party=[dict(name=c['name'],portrait=c['portrait'],female=c['female'],wounds=5,gold=0,
 place=c['place'],town_code=c['town'],building_id=c['building'],speaker=True,
 rank=0,adventure_points=0,quests=[0]*6,progress=[0]*5)])
Path('/out/legacy.json').write_text(json.dumps(s))
Path('/out/legacy-before.json').write_bytes(Path('/out/legacy.json').read_bytes())
PY
start import-v1 -load /out/legacy.json -save /out/legacy.json
shot imported-v1
stop
python3 - <<'PY'
import json
from pathlib import Path
assert Path('/out/legacy.json').read_bytes()==Path('/out/legacy-before.json').read_bytes()
s=json.loads(Path('/out/legacy.json.v2.json').read_text())
c=s['characters'][s['party'][0]]
assert c['gold']==0 and c['wounds']==5 and c['sheet']['rank']==0 and c['sheet']['adventure_points']==0
assert s['settings']['future_setting']=={'retained':True}
assert s['clock']==dict(ticks=4200,part=0,day=0,month=3,year=0,weather=1)
assert '無法還原' in Path('/out/import-v1.log').read_text()
PY
echo 'v1 匯入保留原檔、零值、傷勢與未知設定通過'

# 受控續玩時間案例：只驗載入後的正式走路／露營流程，不宣稱原版同狀態。
python3 - <<'PY'
import json
from pathlib import Path
s=json.loads(Path('/out/legacy.json.v2.json').read_text())
s['clock']=dict(ticks=60,part=4,day=27,month=1,year=2,weather=1)
s['location']['kind']='world'
Path('/out/clock-fixture.json').write_text(json.dumps(s))
PY
start clock-reload -load /out/clock-fixture.json -save /out/clock-fixture.json
shot clock-before
key Right;key Right
shot clock-camping
key space
shot clock-next-month
stop
echo '受控世界時間已重載並走過正常走路／露營，日期畫面另行檢視'
