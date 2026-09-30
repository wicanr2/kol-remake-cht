# 天候動畫的節拍與相位（2026-09-30）

## 結果

野外天空的雨雪輪播與雲層捲動，remake 照原版 `sub_11390` 的計數狀態機推進，同一串
「呼叫、換組、地圖重畫」送進原版與 remake，53 次逐次一致：每一次是等一個 BIOS tick 還是畫一格、
畫的是雨雪第幾張或雲層轉到第幾格都相同（RE-164）。

對拍與實跑找出並修正的差異：

| 項目 | 原版 | remake 舊版 |
|---|---|---|
| 雨雪畫格索引 | 整局延續，換組不歸零 | 換組歸零 |
| 雲層第一次畫 | 先轉一格再畫 | 顯示未旋轉 |
| 地圖重畫 | 計數清零、立刻畫一格 | 沒有這一步 |
| 推進時機 | 只在野外地圖等方向鍵時 | 野外所有狀態（含 CAMPING!、提示頁） |
| 進野外 | 畫面天候組設成 0（晴空）再重畫 | 沿用離城前的組別 |

```
# 原版逐次對拍（容器內，/src 與 /repo 唯讀）
KOL_CHECK=gaps bash /src/tools/oracle/check_updated_dosgolem.sh     # TestWeatherPhaseOriginal
# 原版野外實跑觀察
KOL_CHECK=book KOL_BOOK_SELECT_DOWN=2 KOL_BOOK_BATTLE=1 KOL_BOOK_WALK=1 KOL_BOOK_NEXT_DAY=1 \
KOL_BOOK_WEATHER=1 bash /src/tools/oracle/check_updated_dosgolem.sh
```

## 原版實跑觀察

次日野外把畫面組改成雨（組 3）後送 20 步方向鍵：每一步約 3 個 BIOS tick，期間 4～6 次天候
更新、1～2 張雨畫格，索引跨步延續；20 步裡 1 步有地圖重畫，那一步多畫一格。之後進入訊息畫面，
空等 120 tick 沒有任何天候更新。收據 `workplace/weather-phase-20260930/weather-phase.json`
（只留本機）。

| 輸入 | SHA-256 |
|---|---|
| `KOLTOWN-unpacked.EXE` | `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2` |
| dosgolem | `~/cht/dosgolem-kolrand` `968ca77` |

## 不宣稱

- 節拍以名義 BIOS tick（每秒約 18.2 次）換算 remake 的畫面更新，屬 `hardware-spec approximation`；
  實機 wall-clock 與每一步處理按鍵花的時間不宣稱相同。
- remake 只在離城時做「地圖重畫」那一步；原版野外捲動視窗時的重畫沒有逐一對應，差距在 2 個
  tick 以內。
- 換組之後、下一次繪製之前那不到 2 個 tick，原版畫面還是舊的一張，remake 已顯示新組。
- 城鎮頭盔動畫與天候共用同一個計數，離城那一刻帶過來的計數差距不在比較範圍（進野外會重畫、
  清零）。
- 加速選項沒有做，併入 g07-battle-speed-options（#5）。
