# 正常玩家路徑收據

這裡的截圖序列全部由 `cmd/kol -scene play -play-script …` 產生，經過 `town.Session`
這個狀態機——沒有 direct-entry、沒有座標注入、沒有測試旗標。輸入 token 對應原版鍵位：
`L R U D` 是方向鍵、`A S Z X` 是參考卡的斜走鍵、`<` `>` 移動圖示列白框、`RET` 執行。

```
tools/run_kol_shots.sh workplace/play/path.spec
python3 tools/verify_player_path.py --plan docs/verification/player-path/town.json
```

驗證器檢查三件事：

1. 每一步的截圖存在、解得開、尺寸正確；
2. **相鄰兩步不得完全相同**——每個 token 都要有可見效果。
   這一條抓到過一個真的 bug：門口圖示列的白框位置沒有畫出來，
   按 `>` 之後畫面完全沒變；
3. 指定的檢查點與**已通過對拍**的截圖逐 byte 相同。

## `town`：走動、站到門口、進建築

`L R > RET L R RET RET R R D D U`，14 張截圖：

| 步 | 操作 | 畫面 |
|---|---|---|
| 0 | （起始）| 戶外地圖，隊伍站在半途之家門口的格子上 |
| 1 | `L` | 往西一步，離開門的格子 |
| 2 | `R` | 走回門的格子 → **門口狀態**（地圖換成門面圖，圖示剩 DOOR／U-TURN）|
| 3 | `>` | 白框移到 U-TURN |
| 4 | `RET` | U-TURN → 回到戶外地圖 |
| 5–6 | `L R` | 再走回門口 |
| 7 | `RET` | DOOR → **進入建築**（肖像、標題框、八行敘述）|
| 8 | `RET` | 出來，回到戶外地圖 |
| 9–13 | `R R D D U` | 往東兩步、往南兩步、往北一步 |

檢查點：第 2 步與 `remake-parity/shots/town-door/inn-door.png` 逐 byte 相同、
第 7 步與 `remake-parity/shots/town-building/inn-entry.png` 逐 byte 相同。
**可玩路徑與對拍證據因此直接連上**——玩家走出來的畫面，就是通過逐像素對拍的那批像素。

## `full`：從標題序列走到城鎮

`SPACE SPACE 5 D RET ESC D RET ESC D D D RET SPACE ESC 6 RET > RET L R RET RET R D`，
26 張截圖。這條路徑由 `internal/flow` 串起來：它只決定「按了什麼鍵、畫面走到哪」，
每個畫面自己的狀態機維持原樣。

| 步 | 畫面 |
|---|---|
| 0–2 | 標題序列 → 書本封面 → 目錄頁 |
| 3–5 | 目錄頁 `5` → 角色選項 → 反白條移到 CREATE → New Character 頁 |
| 6–9 | ESC 回選項 → DELETE → 刪除清單 → ESC 回選項 |
| 10–14 | 反白條走過 EDIT PICTURE、EDIT SHIELD、HARBOR MASTER → 訊息頁 → 回選項 |
| 15–17 | ESC 回目錄頁 → `6` → 角色選擇清單 → 選領隊 → **旅店** |
| 18–19 | `>` 白框繞回 DOOR → **戶外地圖** |
| 20–23 | 走離門口 → 走回（門口狀態）→ DOOR 進建築 → 出來 |
| 24–25 | 往東一步、往南一步 |

檢查點：第 21 步與門口對拍截圖、第 22 步與建築內部對拍截圖，都逐 byte 相同。

`EDIT PICTURE`、`EDIT SHIELD` 與旅店的其餘五個圖示會被記進 `flow.Unimplemented`，
不會默默吞掉——測試盯著這一點。

## `town-edge`：走到地圖邊界

`R R U U U U U U U U U > RET`，13 步。路徑由 TOWN7 的格子陣列與不可通行位元圖算出，
與原版 oracle `town-map-edge-run1` 走的是同一串按鍵。

| 步 | 動作 |
|---|---|
| 1–2 | 往東兩步 |
| 3–10 | 往北八步，鏡頭一路捲到 (58,0) |
| 11 | 再往北一步：鏡頭被夾限，跳出城門提示 |
| 12 | 白框移到 U-TURN |
| 13 | 回到地圖，圖示列剩一格 DOOR |

檢查點：第 11 步與城門提示對拍截圖、第 12 步與白框在 U-TURN 的對拍截圖，都逐 byte 相同。

## `town-talk`：跟店主說話

`L R RET SPACE LISTEN SPACE SPACE LISTEN SPACE`，9 步。

| 步 | 動作 |
|---|---|
| 1–3 | 走離門口再走回來，按 DOOR 進旅店 |
| 4 | 讀完進門敘述按空白鍵：招呼語與六格圖示列 |
| 5–6 | 按 LISTEN，再按空白鍵看傳聞 |
| 7–9 | 再聽一則 |

檢查點：第 4、5、6 步與已對拍的三張截圖逐 byte 相同。

## `town-leave`：離開城鎮走到世界地圖

`R R U U U U U U U U U RET`，12 步。第 11 步是城門提示，第 12 步選 DOOR 之後
落到世界地圖 (40,23)——那是霍布入城點 3 在世界地圖上的格子。

檢查點：第 11 步與城門提示截圖、第 12 步與野外地圖對拍截圖，都逐 byte 相同。

## `town-round-trip`、`town-recruit`、`world-road`

三份較長的路徑各自對應一批對拍證據：

| 路徑 | 步數 | 內容 | 檢查點 |
|---|---:|---|---|
| `town-round-trip` | 20 | 城鎮 → 野外走到路上 → 退回入口 → 選 DOOR 進城 | 進城後的城鎮畫面 |
| `town-recruit` | 15 | COMPANION 的兩個畫面、清單捲動、加入第一位同伴 | 問句、清單、加入後三張 |
| `world-road` | 34 | 出城往北走上道路，連按七次 ROAD，再按 SIGNPOST 換另一條 | 循路的四個取樣點 |

## `town-rest`：住一晚並存檔

`L R RET SPACE > > > > RET RET`，11 步。白框走到第 5 格 REST 按下去，
對話框問「住下要 56 枚金幣」，答應之後扣錢、把留宿的建築記進角色、寫出存檔。

檢查點：第 9 步（價錢問句）與 `remake-parity/shots/town-rest/price.png` 逐 byte 相同。

這份路徑另外檢查**寫出來的存檔**：驗證器把 `kol-save-v1` 讀回來逐欄比對——
金幣 1009 − 56 ＝ 953、所在地 `HALFWAY HO`、建築編號 86、位置是城 7 入城點 3
的 (57, 8)。存檔因此也是玩家路徑的一部分，不是另一條沒人走的支線。

## 這份收據不宣稱什麼

- 不宣稱遊戲可玩到結局。目前接上的是「標題 → 書本 → 角色選項 → 旅店 → 城鎮走動、
  門口、進建築 → 野外循路 → 招募同伴 → 住宿存檔」這一條；戰鬥、任務、讀檔回到
  遊戲都還沒有。
- 角色選項頁的 EDIT PICTURE／EDIT SHIELD，以及建築內的 MIRROR、買賣、訓練、AWARD、
  SCROLL 按下去不會轉場，只會記進未實作清單。
- 不宣稱鏡頭跟隨規則已證實。dead zone 是由 22 張原版畫格反解的**強推論**，
  規則本身沒有從程式碼讀出來。
- 野外的移動、道路與進城已經接上（`world-road.json`、`town-round-trip.json`）；
  遭遇、時間推進與雲層動畫還沒有。
- 存檔只寫得出來，還讀不回遊戲：`internal/save` 的讀取與匯入工具都有測試，
  但「讀檔繼續玩」這條路徑尚未接上。
