#!/usr/bin/env bash
# 僅在 Docker 內執行：/src 為 KOL 唯讀來源，/repo 為 dosgolem 唯讀來源，/out 可寫。
set -eu
trap 'python3 /src/tools/oracle/write_receipt.py' EXIT
export PATH=/usr/local/go/bin:$PATH HOME=/tmp
python3 /src/tools/oracle/write_receipt.py --start
mkdir -p /tmp/kol/workplace/analysis/koltown /tmp/kol/workplace/oracle
cp -r /src/internal /tmp/kol/
cp /src/go.mod /src/go.sum /tmp/kol/
cp /src/workplace/analysis/koltown/KOLTOWN-unpacked.EXE /tmp/kol/workplace/analysis/koltown/
cp -r /src/workplace/oracle/game /tmp/kol/workplace/oracle/
mkdir -p /tmp/dosgolem
cp /repo/go.mod /tmp/dosgolem/
cp -r /repo/internal /repo/oracle /tmp/dosgolem/
cp -r /go/pkg/mod /tmp/mod
export GOMODCACHE=/tmp/mod GOCACHE=/tmp/gocache GOPATH=/tmp/gopath GOPROXY=off GOSUMDB=off GOWORK=off
case "${KOL_CHECK:-all}" in all|normal|book|editor|durability|panels|gaps|exit|terrain|quest|gear|greet|run) ;; *) echo 'KOL_CHECK 只接受 all、normal、book、editor、durability、panels、gaps、exit、terrain、quest、gear、greet 或 run' >&2; exit 2 ;; esac
cd /tmp/dosgolem
if [ "${KOL_CHECK:-all}" = all ]; then
go test ./internal/cpu/... ./internal/dos/... ./internal/dosfile/... ./internal/machine/... ./oracle/... -count=1 -timeout 180s > /out/dosgolem.log 2>&1
fi
cd /tmp/kol
go work init . /tmp/dosgolem
export GOWORK=/tmp/kol/go.work
if [ "${KOL_CHECK:-all}" = terrain ] || [ "${KOL_CHECK:-all}" = all ]; then
go test -tags oracle ./internal/oracle -run '^TestBattleFieldGenerationOriginal$' -count=1 -v -timeout 120s > /out/terrain.log 2>&1
if [ "${KOL_CHECK:-all}" = terrain ]; then exit 0; fi
fi
if [ "${KOL_CHECK:-all}" = exit ] || [ "${KOL_CHECK:-all}" = all ]; then
go test -tags oracle ./internal/oracle -run '^Test(BattleExitWounds|QuestItemRecord|MonsterRangedConversion|MonsterRangedGate)Original$' -count=1 -v -timeout 180s > /out/battle-exit.log 2>&1
if [ "${KOL_CHECK:-all}" = exit ]; then exit 0; fi
fi
if [ "${KOL_CHECK:-all}" = quest ] || [ "${KOL_CHECK:-all}" = all ]; then
go test -tags oracle ./internal/oracle -run '^TestQuest(DoneMark|PickUp|Delivery|FleeEdge)Original$' -count=1 -v -timeout 600s > /out/quest-chain.log 2>&1
if [ "${KOL_CHECK:-all}" = quest ]; then exit 0; fi
fi
if [ "${KOL_CHECK:-all}" = gear ] || [ "${KOL_CHECK:-all}" = all ]; then
go test -tags oracle ./internal/oracle -run '^Test(BeltSwitch|SpellWard|AbbeyHeal)Original$' -count=1 -v -timeout 600s > /out/gear.log 2>&1
if [ "${KOL_CHECK:-all}" = gear ]; then exit 0; fi
fi
# `KOL_CHECK=run KOL_ORACLE_RUN='^TestXxx$'`：只跑 internal/oracle 裡符合的測試（臨時探針用）。
if [ "${KOL_CHECK:-all}" = run ]; then
go test -tags oracle ./internal/oracle -run "${KOL_ORACLE_RUN:?要設 KOL_ORACLE_RUN}" -count=1 -v -timeout 600s > /out/run.log 2>&1
exit 0
fi
if [ "${KOL_CHECK:-all}" = greet ] || [ "${KOL_CHECK:-all}" = all ]; then
go test -tags oracle ./internal/oracle -run '^Test(NameCanonical|GreetingExpel|GreetingSpeaker)Original$' -count=1 -v -timeout 900s > /out/greet.log 2>&1
if [ "${KOL_CHECK:-all}" = greet ]; then exit 0; fi
fi
if [ "${KOL_CHECK:-all}" = gaps ]; then
go test -tags oracle ./internal/oracle -run '^Test(ArenaMap|WeatherDispatch|WeatherComposite|WeatherPhase)Original$' -count=1 -v -timeout 120s > /out/player-gaps.log 2>&1
exit 0
fi
if [ "${KOL_CHECK:-all}" = panels ] || [ "${KOL_CHECK:-all}" = all ]; then
go test -tags oracle ./internal/oracle -run '^TestBattle(CounterScanStart|RingSearch|EnemyShootRing|AmmunitionConsume|AmmunitionPanel|AmmunitionText|StatusBar|BodyOverlay|BodySprite|PanelSpace|TargetPanelSpace|EquipmentRows|PortraitPosition|CommandPanel|CommandSprite|CommandText|CommandDirection|CommandInspection|CommandIntellect|CommandVisibility|PlanningInspectionOrder|PlanningLimit|PlanningCancel)Original$' -count=1 -v -timeout 120s > /out/panels.log 2>&1
if [ "${KOL_CHECK:-all}" = panels ]; then exit 0; fi
fi
if [ "${KOL_CHECK:-all}" = durability ] || [ "${KOL_CHECK:-all}" = all ]; then
go test -tags oracle ./internal/oracle -run '^Test(DurabilityBoundariesOriginal|SignedItemBonusOriginal|EquipmentRecordPointerOriginal|ArmourReductionOriginal)$' -count=1 -v -timeout 120s > /out/durability.log 2>&1
if [ "${KOL_CHECK:-all}" = durability ]; then exit 0; fi
fi
if [ "${KOL_CHECK:-all}" = all ]; then
go test -tags oracle ./internal/oracle -run '^Test(FatigueRatioMatchesOriginal|SkillPenaltyMatchesOriginal|RandStreamMatchesOriginal|RollMatchesOriginal|PartySkillRowFollowsWeaponClass|TakeDamageMatchesOriginal)$' -count=1 -v -timeout 180s > /out/kol-rules.log 2>&1
fi
mkdir -p /tmp/kol/tools/oracle
cp /src/tools/oracle/normal_startup_test.go /tmp/kol/tools/oracle/
cp /src/tools/oracle/book_startup_test.go /tmp/kol/tools/oracle/
cp /src/tools/oracle/book_editor_test.go /tmp/kol/tools/oracle/
if [ "${KOL_CHECK:-all}" = editor ]; then
KOL_ORACLE_ROOT=/tmp/kol/workplace/oracle/game KOL_ORACLE_CAPTURE=/out \
go test -tags oracle ./tools/oracle -run '^TestBookBitmapEditorOriginal$' -count=1 -v -timeout "${KOL_BOOK_TIMEOUT:-600s}" > /out/book-editor.log 2>&1
exit 0
fi
if [ "${KOL_CHECK:-all}" != book ]; then
KOL_ORACLE_ROOT=/tmp/kol/workplace/oracle/game KOL_ORACLE_CAPTURE=/out \
go test -tags oracle ./tools/oracle -run '^TestNormalStartupUpdatedDosgolem$' -count=1 -v -timeout 240s > /out/normal-startup.log 2>&1
fi
if [ "${KOL_CHECK:-all}" != normal ]; then
KOL_ORACLE_ROOT=/tmp/kol/workplace/oracle/game KOL_ORACLE_CAPTURE=/out \
go test -tags oracle ./tools/oracle -run '^TestNormalBookEntry$' -count=1 -v -timeout "${KOL_BOOK_TIMEOUT:-240s}" > /out/book-startup.log 2>&1
fi
