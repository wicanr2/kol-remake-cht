#!/usr/bin/env bash
# 推廣片的分鏡與合成。在 `game-video:latest` 容器裡跑，由 `tools/make_promo.sh` 呼叫。
# 掛載：`/docs`（唯讀）、`/assets`（唯讀）、`/out`（寫）。
set -eu
. /docs/promo/theme.sh

SHOT=/docs/verification
TMP=/tmp/promo; rm -rf "$TMP"; mkdir -p "$TMP"

# --- 版面函式（五種，輪流用，避免十二段長得都一樣）-------------------------

bg() { # $1 out —— 赤石底：暗紅漸層 ＋ 交丁砌的石縫（母題取自原版城鎮的磚牆邊框）
  local args="" y x row=0
  for y in 0 120 240 360 480 600; do
    args="$args -draw \"rectangle 0,$((y+117)) ${W},$((y+119))\""
    # 交丁砌：奇數列的縱縫錯開半塊。
    if [ $((row % 2)) -eq 0 ]; then x=0; else x=-107; fi
    while [ $x -lt $W ]; do
      [ $x -ge 0 ] && args="$args -draw \"rectangle $x,$y $((x+2)),$((y+117))\""
      x=$((x+214))
    done
    row=$((row+1))
  done
  eval convert -size ${W}x${H} '"gradient:${BG_LITE}-${BG_DEEP}"' -fill '"#00000026"' $args "$1"
}

card() { # $1 out $2 中標 $3 英標 $4 副標 —— 鎏金浮雕標題卡
  bg "$TMP/_bg.png"
  convert "$TMP/_bg.png" -gravity center \
    -font "$FONT_TITLE" \
    -fill "$ACCENT_SH" -pointsize 86 -annotate +3+3 "$2" \
    -fill "$ACCENT"    -pointsize 86 -annotate +0+0 "$2" \
    -font "$FONT_BODY" \
    -fill "$TEXT" -pointsize 40 -annotate +0+90 "$3" \
    -fill "$DIM"  -pointsize 26 -annotate +0+170 "$4" \
    "$1"
}

slide_frame() { # $1 out $2 圖 $3 字幕 —— 金框置中，底部字幕條
  bg "$TMP/_bg.png"
  convert "$2" -filter point -resize x520 -bordercolor "$ACCENT" -border 3 "$TMP/_sc.png"
  convert "$TMP/_bg.png" "$TMP/_sc.png" -gravity north -geometry +0+40 -composite \
    -fill "#00000099" -draw "rectangle 0,624 ${W},720" \
    -font "$FONT_BODY" -fill "$TEXT" -gravity south -pointsize 32 -annotate +0+34 "$3" \
    "$1"
}

slide_full() { # $1 out $2 圖 $3 字幕 —— 滿版填滿，下三分之一壓字幕條
  convert "$2" -filter point -resize ${W}x${H}^ -gravity center -extent ${W}x${H} \
    -fill "#000000cc" -draw "rectangle 0,600 ${W},720" \
    -font "$FONT_BODY" -fill "$TEXT" -gravity south -pointsize 34 -annotate +0+42 "$3" \
    "$1"
}

split_ba() { # $1 out $2 左圖 $3 右圖 $4 左標 $5 右標 —— 左右對照
  bg "$TMP/_bg.png"
  convert "$2" -filter point -resize 560x400 -bordercolor "$ACCENT" -border 2 "$TMP/_l.png"
  convert "$3" -filter point -resize 560x400 -bordercolor "$ACCENT" -border 2 "$TMP/_r.png"
  convert "$TMP/_bg.png" \
    "$TMP/_l.png" -gravity northwest -geometry +40+150 -composite \
    "$TMP/_r.png" -gravity northeast -geometry +40+150 -composite \
    -font "$FONT_BODY" -fill "$DIM" -gravity northwest -pointsize 30 -annotate +60+100 "$4" \
    -font "$FONT_BODY" -fill "$TEXT" -gravity northeast -pointsize 30 -annotate +60+100 "$5" \
    -font "$FONT_TITLE" -fill "$ACCENT" -gravity center -pointsize 56 -annotate +0-30 "▶" \
    "$1"
}

quote() { # $1 out $2 上句 $3 下句 —— 巨型引號的金句卡
  bg "$TMP/_bg.png"
  convert "$TMP/_bg.png" \
    -font "$FONT_TITLE" -fill "#ffffff14" -gravity northwest -pointsize 200 -annotate +48+40 '「' \
    -font "$FONT_TITLE" -fill "#ffffff14" -gravity southeast -pointsize 200 -annotate +48+40 '」' \
    -font "$FONT_BODY" -fill "$TEXT"   -gravity center -pointsize 44 -annotate +0-28 "$2" \
    -fill "$ACCENT" -draw "rectangle 540,382 740,384" \
    -font "$FONT_BODY" -fill "$ACCENT" -gravity center -pointsize 32 -annotate +0+58 "$3" \
    "$1"
}

# --- 一段影片 ---------------------------------------------------------------
seg() { # $1 png $2 mp4 $3 秒
  FO=$(awk "BEGIN{print $3-0.6}")
  ffmpeg -y -loglevel error -loop 1 -i "$1" -t "$3" -r $FPS \
    -vf "fade=t=in:st=0:d=0.6,fade=t=out:st=$FO:d=0.6,format=yuv420p" \
    -threads 2 -c:v libx264 -preset veryfast -pix_fmt yuv420p "$2"
}

# --- 分鏡 -------------------------------------------------------------------
n=0
add() { # $1 png $2 秒
  m=$(printf '%02d' $n)
  seg "$1" "$TMP/v_$m.mp4" "$2"
  echo "file '$TMP/v_$m.mp4'" >> "$TMP/list.txt"
  n=$((n+1))
}
: > "$TMP/list.txt"

card "$TMP/00.png" '武士傳說' 'Knights of Legend' '1989 · Origin Systems　繁體中文重製'
add  "$TMP/00.png" "$PACE_CARD"

slide_full "$TMP/01.png" "$SHOT/player-path/full-chain/step-000.png" \
  '一支隊伍，一片被遺忘了三十幾年的大陸'
add "$TMP/01.png" "$PACE_SLIDE"

slide_frame "$TMP/02.png" "$SHOT/cjk-ui/shots/outdoor.png" \
  '西元 1100 年四月一日 —— 時刻、季節與天候都照原版在走'
add "$TMP/02.png" "$PACE_SLIDE"

slide_frame "$TMP/03.png" "$SHOT/cjk-ui/shots/inn.png" \
  '七座城鎮 101 棟建築、1,938 則對話，全部譯成繁體中文'
add "$TMP/03.png" "$PACE_SLIDE"

slide_frame "$TMP/04.png" "$SHOT/cjk-ui/shots/step-005.png" \
  '建角色的每一欄都看得懂：種族、職業、九項屬性'
add "$TMP/04.png" "$PACE_SLIDE"

split_ba "$TMP/05.png" "$SHOT/player-path/quest-list/step-029.png" \
  "$SHOT/cjk-ui/shots/roster.png" '英文模式 320×200' '繁中模式 640×400'
add "$TMP/05.png" "$PACE_SLIDE"

quote "$TMP/06.png" '英文模式逐畫面對拍原版，繁中模式把畫布放大兩倍' \
  '同一份規則，兩種讀法'
add "$TMP/06.png" "$PACE_CARD"

slide_frame "$TMP/07.png" "$SHOT/battle-ambush/field.png" \
  '戰鬥仍是原版那套：先規劃一整輪，再一手一手執行'
add "$TMP/07.png" "$PACE_SLIDE"

slide_frame "$TMP/08.png" "$SHOT/player-path/town/step-013.png" \
  '城門、馬廄、商店、訓練場 —— 走的是玩家自己的那條路'
add "$TMP/08.png" "$PACE_SLIDE"

quote "$TMP/09.png" '每一條規則都要先有反組譯證據，才寫得進程式' \
  'RE 證據 → 規格 → 實作 → 同狀態驗證'
add "$TMP/09.png" "$PACE_CARD"

card "$TMP/10.png" '武士傳說' 'Knights of Legend' \
  '繁體中文重製 · 需自備原版磁片產生素材'
add "$TMP/10.png" "$PACE_CARD"

# --- 串接 -------------------------------------------------------------------
ffmpeg -y -loglevel error -f concat -safe 0 -i "$TMP/list.txt" \
  -threads 2 -c:v libx264 -preset veryfast -pix_fmt yuv420p "$TMP/silent.mp4"
DUR=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$TMP/silent.mp4")

# --- 音軌：只有遊戲自己的雷聲，其餘留白（見檔頭的素材界線）-------------------
# 開場那一下雷落在標題卡上，第二下落在夜景。沒有背景音樂是刻意的。
#
# ⚠ 音軌要鋪滿整片，靠的是**第三路 `anullsrc` 拿 `-t` 定長**再 `amix`。
# 兩條走不通的路都試過了：`apad` 沒有長度就是無限產生，而 `atrim` 只丟掉超過的
# 畫格、不會讓上游 EOF，整支 ffmpeg 一直跑下去（輸出永遠停在 48 byte）；
# 改成 `apad=whole_dur=$DUR` 不會卡住，但音軌只有 8.2 秒——它管的是「補到多長」
# 而不是「輸出多長」，前面的 `adelay` 已經把時間軸推掉了。
ffmpeg -y -loglevel error \
  -i "$TMP/silent.mp4" -i /assets/audio/title-thunder.ogg \
  -f lavfi -t "$DUR" -i anullsrc=r=44100:cl=stereo \
  -filter_complex "\
    [1:a]adelay=900|900,volume=0.9[t1]; \
    [1:a]adelay=7200|7200,volume=0.55[t2]; \
    [t1][t2][2:a]amix=inputs=3:duration=longest:normalize=0,\
afade=t=out:st=$(awk "BEGIN{print $DUR-2}"):d=2[a]" \
  -map 0:v -map "[a]" -threads 2 \
  -c:v libx264 -preset veryfast -pix_fmt yuv420p -c:a aac -b:a 160k \
  -movflags +faststart "/out/kol-promo.mp4"

# --- 收據 -------------------------------------------------------------------
V=$(ffprobe -v error -select_streams v:0 -show_entries stream=width,height,duration \
      -of csv=p=0 "/out/kol-promo.mp4")
A=$(ffprobe -v error -select_streams a:0 -show_entries stream=codec_name,duration \
      -of csv=p=0 "/out/kol-promo.mp4")
# volumedetect 印在 **info** 等級，`-v error` 會把它一起吃掉（量到空字串）。
LOUD=$(ffmpeg -hide_banner -v info -i "/out/kol-promo.mp4" -af volumedetect -f null /dev/null 2>&1 \
      | grep -E "mean_volume|max_volume" | sed 's/.*\] //' | tr '\n' ' ')
cat > /out/sources.json <<JSON
{
  "schema": "kol-promo-sources-v1",
  "theme": "$THEME_NAME",
  "video": {"probe": "$V", "audio": "$A", "loudness": "$LOUD"},
  "picture_policy": "全部是 remake 的實機截圖，來自 docs/verification/ 的可重播收據；沒有 mockup、沒有重畫、不調色。",
  "audio_policy": "沒有背景音樂。KOL 的 DOS 版全庫普查後只有 TITLE 的 PC Speaker 音爆，原版沒有可用的配樂；rulebook/93 的鐵則是找不到就回報，不自己合成頂替。片中兩下雷聲是 assets/audio/title-thunder.ogg，依公開的 PIT channel 2 規格產生，權利分類為 remake 自有。",
  "shots": [
    "docs/verification/player-path/full-chain/step-000.png",
    "docs/verification/cjk-ui/shots/outdoor.png",
    "docs/verification/cjk-ui/shots/inn.png",
    "docs/verification/cjk-ui/shots/step-005.png",
    "docs/verification/player-path/quest-list/step-029.png",
    "docs/verification/cjk-ui/shots/roster.png",
    "docs/verification/battle-ambush/field.png",
    "docs/verification/player-path/town/step-013.png"
  ],
  "audio_source": "assets/audio/title-thunder.ogg"
}
JSON
# 抽四幀供讀圖檢查。
for t in 3 18 33 60; do
  ffmpeg -y -loglevel error -ss $t -i "/out/kol-promo.mp4" -frames:v 1 "/out/frame-$t.png"
done
echo "$V | $A | $LOUD"
