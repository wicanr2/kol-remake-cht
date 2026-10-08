#!/usr/bin/env bash
set -euo pipefail
export PATH=/usr/local/go/bin:$PATH
export GOMODCACHE=/tmp/mod GOCACHE=/tmp/gocache GOPATH=/tmp/gopath
export GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local GOWORK=off
mkdir -p /tmp/dosgolem /tmp/kol/tools/oracle
git -c safe.directory=/repo -C /repo archive 968ca774705853276ae55c8d1163c369481ac71f go.mod internal oracle | tar -x -C /tmp/dosgolem
cp /src/go.mod /src/go.sum /tmp/kol/
cp /src/tools/oracle/book_startup_test.go /tmp/kol/tools/oracle/
cp -r /src/internal /tmp/kol/
cp -r /go/pkg/mod /tmp/mod
python3 /scratch/kol_cursor_gui_prepare.py
cd /tmp/kol
go work init . /tmp/dosgolem
export GOWORK=/tmp/kol/go.work
export KOL_ORACLE_ROOT=/out/${KOL_CURSOR_ATTEMPT:-gui1}/game KOL_ORACLE_CAPTURE=/out/${KOL_CURSOR_ATTEMPT:-gui1}
export KOL_BOOK_BATTLE=1 KOL_BOOK_WALK=1 KOL_BOOK_NEXT_DAY=1 KOL_BOOK_ENTER_FIGHT=1 KOL_BOOK_PLAN=1
export KOL_BOOK_PLAN_KEYS=${KOL_CURSOR_KEYS:-Return,Right,Return,Right,Down,Left,Up,Return,Return}
go test -tags oracle ./tools/oracle -run '^TestNormalBookEntry$' -count=1 -v -timeout 600s > /out/${KOL_CURSOR_ATTEMPT:-gui1}/run.log 2>&1
python3 /scratch/kol_cursor_gui_finish.py
