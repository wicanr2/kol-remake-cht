# 正常玩家路徑收據

這裡的截圖序列由 `cmd/kol` 的操作腳本產生，經過各畫面自己的狀態機——沒有
direct-entry、沒有座標注入、沒有測試旗標。輸入 token 對應原版鍵位：
`L R U D` 是方向鍵、`A S Z X` 是參考卡的斜走鍵、`<` `>` 移動圖示列白框、`RET` 執行、
`SPACE` 翻下一段文字。

短路徑（十幾二十步）逐步存圖，用 `verify_player_path.py` 驗；
從標題走到存檔的長路徑有上千步，改存**狀態軌跡**加幾張檢查點截圖，
用 `verify_full_path.py` 驗。

```
tools/run_kol_shots.sh workplace/play/path.spec
python3 tools/verify_player_path.py --plan docs/verification/player-path/town.json
python3 tools/verify_full_path.py   --plan docs/verification/player-path/full-chain.json
```

長路徑那幾份**從 plan 就能一鍵重生**：`tools/replay_player_path.py <名字>`（或 `--all`）
把腳本、種子、要讀的存檔與換過的資料檔組成規格檔跑一遍，再驗一次。
plan 因此要帶齊重跑所需的每一個旗標——`full-chain-ending` 的 `-quest-sites`
一度只寫在散文裡，改了戰鬥數值之後重跑打的是原版那十二隻 CYCLOPS，怎麼跑都跑不完。

> ⚠ **2026-09-11：從標題走起的那幾條收據過期了。** `e118596` 把選角之後的
> `SceneInn` 改成 `SceneTown`（開場旅店接進正式 `town.Session`），畫面節奏因此變了
> 而按鍵沒有重生——`flow-quest` 的 109 步在英文模式重跑之後停在城鎮戶外，收據裡
> 那一步是遭遇畫面，trace 的第一個分歧就在第 17 步。受影響的是 `full`、
> `full-chain`、`full-chain-reload`、`full-chain-ending`、`flow-quest`、`arena`。
>
> 底下那兩支驗證**不重跑**，比的是收據目錄裡那幾張圖與軌跡自不自洽，所以程式改了
> 它們照樣通過。要抓這種回歸得重跑再比，而 `replay_player_path.py` 會寫回這個目錄，
> 不能當檢查用——`tools/check_player_paths.py` 才是：它重跑到 `workplace/replay-check/`
> 再與這個目錄裡同名那幾張逐 byte 比。第一次跑的結果：17 條短路徑裡 5 條一致
> （town-edge、town-mirror、town-recruit、town-rest、town-talk）、10 條對不上、
> 2 條因為 plan 缺旗標跑不動。
>
> **有六條 plan 的描述寫「讀存檔開場」而 plan 沒有 `load` 欄位**（town-load、
> town-scroll、town-shop、town-stable、town-train、world-quest），`full` 要走 flow
> 而沒有 `scene` 欄位——那正是這份 README 上面說過的坑：重跑所需的每一個旗標都要
> 寫進 plan，寫在散文裡機器讀不到。見未完成項 `player-path-receipts-stale`。

**改了戰鬥數值之後戰鬥那一段的按鍵就作廢了**（打幾個回合、每一手的訊息都不同）。
`tools/drive_battle_script.py` 用來重生它：跑一次腳本 → 讀軌跡的最後一筆 →
依規劃段停在哪一層決定下一批 token → 再跑，產出的仍是一串固定按鍵。走路與城裡
那幾段沿用原腳本的前後綴，只換中間那一段。

`docs/verification/battle-*/` 那幾份短收據同理，重跑用
`tools/replay_battle_receipts.py <名字>`（或 `--all`）：跑各自的 `replay.spec`，
再拿輸出的 PNG 對收據 `frames` 的 SHA-256 與版控裡那一張。

`verify_player_path.py` 檢查三件事：

1. 每一步的截圖存在、解得開、尺寸正確；
2. **相鄰兩步不得完全相同**——每個 token 都要有可見效果。
   這一條抓到過一個真的 bug：門口圖示列的白框位置沒有畫出來，
   按 `>` 之後畫面完全沒變；
3. 指定的檢查點與**已通過對拍**的截圖逐 byte 相同。

`verify_full_path.py` 檢查的是里程碑與檢查點：軌跡（`-trace` 寫出來的 JSON）在
指定的那幾步必須停在指定的畫面、座標與戰鬥階段，檢查點截圖必須與已對拍的那一批
逐 byte 相同，路徑寫出來的存檔必須逐欄對得上。軌跡是遊戲自己報的狀態，不是從畫面猜的。

## 四條走到底的路徑

完成定義要的是「啟動 → 建立隊伍 → 城鎮／世界探索 → 戰鬥規劃與結算 → 任務 →
存檔 → 離開 → 重載 → 結局」走在**同一串按鍵**上。四條合起來蓋滿它：

| 路徑 | 步數 | 走到哪 |
|---|---:|---|
| `full-chain` | 938 | 標題 → 書本 → 角色選項 → 旅店 → 城鎮 → 出城 → 野外遇伏 → 打贏 → 回城 → 住一晚並存檔 |
| `full-chain-reload` | 119 | 讀那份存檔 → 資料頁確認金幣 → 出城 → 任務地點 → 答應參戰 |
| `flow-quest` | 109 | 同一條 flow 路徑走到任務地點，兩張與已對拍的截圖逐 byte 相同 |
| `quest-take` | 34 | 走進一戶人家，按 MOUTH 打字問話，接下一個任務 |
| `quest-list` | 29 | 資料頁按第 19 個圖示，看回報過的任務排成八欄三列 |
| `arena` | 1770 | 走到世界地圖 (55,13) 的競技場，選人、進賭場下兩注、打完升一階並結算賭金 |
| `full-chain-ending` | 1003 | 出城 → 打贏一場伏擊 → 最終任務的地點 → 打完 → 結局七段 |

### `full-chain`：從標題走到存檔

隊伍是原版名冊裡六名帶武器的角色（Poontang、Moraine、Fighting Bull、Kenneth、
Sharakan、Olear），用 `tools/import_original_characters.py` 匯入。路徑上發生的事：

| 步 | 內容 |
|---|---|
| 0–18 | 標題序列 → 書本 → 目錄頁 → 角色選項（走過 CREATE、DELETE、HARBOR）→ 角色選擇清單 → 旅店 |
| 19–30 | DOOR → 城鎮戶外地圖 → 往東兩步、往北九步走到地圖上緣，跳出城門提示 |
| 31 | 選 DOOR 離開城鎮，落到霍布入城點 3 在世界地圖上的 (40,23) |
| 45 | 往東走兩格被伏擊：滿版的 ZOMBIE 圖（SPEC-011） |
| 46–791 | 八個回合的規劃與執行：移動靠近、揮武器、推每一手的訊息 |
| 792 | `You have defeated The ZOMBIES!` |
| 794–804 | 回野外，走回入城點 |
| 805–928 | 進城，走到半途之家門口（城裡的路是邊走邊撞出來的，撞到建築就換一條） |
| 929–938 | DOOR 進門 → REST → 「住下要 56 枚金幣」→ 答應：扣錢、記下留宿的建築、寫出存檔 |

**這條路徑釘住了亂數種子。** 野外每次時段變換都會擲一次遭遇，走到哪一格遇上什麼
由種子決定；戰鬥裡的攻防與規劃順序同樣要重播。

寫出來的存檔逐欄比對，其中冒險點數證明了戰後獎勵真的寫回角色：Poontang 從角色檔
的 2,316 變成 2,376，那一場每人分到 60（RE-096）。

### `full-chain-reload`：存檔的迴圈

讀 `full-chain` 寫出來的存檔，先進半途之家開角色資料頁——上面的金幣是住宿扣掉 56
之後的 158，不是角色檔裡的 214。接著出城走到第 5 個任務地點，中途經過換日的
CAMPING!，踏上任務地點時蓋出敘述與「Will you partake?」，答應之後是滿版的怪物圖。

### `flow-quest`：與對拍證據連上

`full` 的前 19 步接 `world-quest` 的 90 步，存檔與種子沿用 `world-quest`
（單人 Poontang、seed 3）。兩個檢查點與 `remake-parity/shots/world-quest/` 的
截圖逐 byte 相同——從標題按下去和從城鎮直接開場，走到任務地點時畫的是同一批像素。

### `quest-take`：接下一個任務

走進霍布城西的一戶人家（TOWN7 的第 7 扇門），白框走到 MOUTH 按下去，
在對話框裡打 `nobj` 四個字元再按 Return——主人就把任務講給你聽：敘述分兩頁各八行，
讀完問「Will you do this?」，答應之後第 6 個任務的旗標從「沒接」變成「接了」
（RE-067／RE-110）。**沒有換過任何輸入**——隊伍是原版名冊裡的六名角色，
存檔只指定開場站在哪一格。

接任務的入口是問話的**第 5 槽**，而且要這一棟的 `icon_flags[1] & 4`——
與收任務物品同一位，同一棟給也收。打字那一格讀入最多 23 個字元、整串轉小寫，
這兩件事有證據；**游標的樣子與座標沒有對拍**，收據裡標了不宣稱。

### `quest-list`：接了哪些、完成了哪些

角色資料頁上按第 19 個圖示打開任務一覽：二十四個任務排成八欄三列，
**只列回報過的**，一格是 AWDnPICS 的一張 24×21 圖加一個綠框，一組一頁
（RE-111）。那一格圖示本身是條件出現的——一件都沒回報過就不會有這一格。

**這一組換過輸入**：原版名冊裡的角色都是全新的，一件任務都沒回報過，
所以用工具把存檔改成「回報過第 4 個任務、已回報數 5」。

### `arena`：升階唯一的入口

世界地圖 (55,13) 那一格與城鎮入口走同一個版面：DOOR 進場、U-TURN 留在野外。
進去先看誰夠格——帳面位階與實得位階不同、而且帳面還沒到 24；隊伍多人就讓玩家在
清單上選，選好再問「Are you sure you wish to enter?」。

多人隊伍答應之後先換到賭場（`A Gambling Booth／The Bet Taker`）：沒上場的人可以押
那一位會贏還是會輸，金額讀兩位數（上限 99），押他會輸寫負值再多講一句
「Good friends, eh?」；答否才開打。這一條路上押了兩注——Poontang 押 25 賭
Moraine 會贏、Fighting Bull 押 10 賭她會輸。

**場上只有上場那一名**：其餘五名排到觀眾席、記錄 `+0x01` 設成 2 不參戰，上場那一名
與唯一那隻對手各站一格固定的（RE-098 的第三條分支）。對手一隻，等級是上場那一名的
實得位階加一。第五回合打贏，印「you have gained a level!」，帳面位階加一——**一次
只加一級**，所以帳面追上實得要打好幾場；賭金接在升階那一句後面結算，最後一頁是
三行：升階、`Poontang wins a bet!`、`Fighting Bull loses a bet!`，錢包 214 → 248 與
442 → 441（RE-108／RE-103）。路上還打贏了兩場伏擊。

路上那兩場伏擊佔掉整條路徑的八成四——第一場 582 步、第二場 905 步，競技場那一場
只有 94 步。敵人傷得夠重會撤退、被打時會招架、聰明的還會反制掃描與繞側翼
（RE-081／RE-104），追著逃跑的怪補刀本來就是原版的節奏。

**這一組換過輸入**：原版名冊裡的角色兩個位階都是齊的（他們是遊戲開始前的狀態），
照原樣走進去只會聽到「None of you are ready」，所以用工具把兩名的帳面位階降到 0。

### `full-chain-ending`：打完最終任務看結局

**結局的觸發點是任務戰鬥打完那一刻**（RE-106）：原版在 `sub_2B0FF` 的尾端先看
`word_34968`——那是「這一場真的把任務標成完成了」的旗標（RE-101）——成立才逐名
檢查已回報數等於 0x17。這條路徑從標題走到霍布另一個城門出城、路上打贏一場伏擊、
走到最終任務的地點打完那一場，畫面接著換成結局。七張圖與七段敘事都與
`-scene ending` 那一組收據逐 byte 相同。

**這一組換過兩處輸入**：存檔要「已回報 23 件、最終任務已接」——接任務雖然做完了
（`quest-take`），要在遊戲裡真的回報 23 件仍然不現實；最終任務原本是十二隻 10 級
CYCLOPS，六人隊打不贏，所以壓成一隻 1 級。收據裡列了換過什麼、不宣稱什麼。

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

## `town-mirror`：角色資料頁

`L R RET SPACE > RET > < RET`，10 步。白框走到第 2 格 MIRROR 打開資料頁，
在頁面上移動白框，再選 U-TURN 回到建築內。

檢查點：第 6 步與 `remake-parity/shots/town-mirror/sheet.png` 逐 byte 相同——
那一張整頁只和原版差原版自己畫的滑鼠游標。

## `town-load`：從存檔開場

`L R RET SPACE > RET`，7 步，用 `-load` 讀 `town-rest` 寫出來的那一份存檔。

檢查點用的是 `same_as_except`：第 6 步與**沒讀檔**的資料頁比對，
金幣那一欄（第 29 欄起 9 格）之內必須不同、之外必須逐像素相同。
收據記下的是內 36 個像素不同、外 0 個——1009 變成 953，其餘一個像素都沒動。

這一份要先跑 `town-rest` 才有存檔可讀；兩份合起來把存檔的迴圈閉起來。

## `town-shop`：買一件東西

`L RET SPACE > > > > RET D U RET RET SPACE`，14 步，從存檔開場站在城 1 魯德曼甲冑鋪門邊。
走進去按 TRADE 開商品清單，上下移動反白，選第一件看明細，答應買下——
扣掉 32 枚金幣；毛皮帽的尺寸剛好合身，店家直接幫他穿上。

檢查點：第 8 步（清單）、第 11 步（明細）與第 12 步（穿上）都與已對拍的截圖
逐 byte 相同。
存檔用 `tools/import_original_characters.py --names ddduhh --location 1,0,24,52`
產生，`--location` 就是為了讓這種固定起點可以重生。

## `town-stable`：買一匹坐騎

`U RET SPACE > > > > RET D RET RET SPACE > RET > RET`，17 步，從存檔開場站在城 1
Holcroft Stables 門邊。進門按 TRADE 開四匹的清單（永遠四筆、沒有捲動箭頭），
選輕型戰馬看明細——圖在 (112, 86)、名字在右邊、問句在下面——答應買下扣掉
400 枚金幣，店家講「Take these reins」。離開商店開 MIRROR：資料頁右上角多一塊
坐騎，名牌只印四個字元（`LGT`），移動率從用走的 `U`（20）換成 `E`（4）。

檢查點只有描述，**沒有 identical_to**：五間馬廄都不在城 7，而原版 oracle 目前
只走得到城 7 的半途之家，所以這一頁還沒有原版對拍截圖。版面數字全部來自
RE-120 的反組譯（`sub_1452D` 把游標設回 (0x0D, 0x56)，之後三段都從那裡推）。

存檔用 `tools/make_stable_save.py --in workplace/saves/quest.json
--out workplace/saves/stable.json --gold 3000` 產生：原版名冊的角色都住城 7，
錢也不夠買最便宜那一匹。

## `town-scroll`：加入教派、買一筆法術

`U RET SPACE > > > > RET RET SPACE RET D RET RET > RET > RET > > > RET`，23 步，從存檔開場
站在城 7 The Hostel 門邊。按 SCROLL 先問要不要加入「Blue Gem Order」——入會費
450、智力門檻 90；答應之後扣錢，教派與頭銜寫進角色。選單只有一列
「Learn a spell」，因為這名角色還沒學過法術。進去挑第二筆 Tyatwelon（330）買下，
回清單。離開之後開 MIRROR：資料頁最下面那一列是「Blue Gem  Illusionist」，
圖示列也多出法術那一格——按下去開法術一覽（兩欄各八格），剛買的 Tyatwelon
在左上那一格。

檢查點只有描述，**沒有 identical_to**：原版收據裡沒有人入過會，這一頁還沒有
對拍截圖。版面與流程數字全部來自 RE-121 的反組譯。

存檔用 `tools/make_stable_save.py --in workplace/saves/quest.json
--out workplace/saves/hostel.json --gold 2000 --town 7 --x 59 --y 28 --entry 7`
產生：原版角色的錢不夠付入會費加法術。`--entry` 選的是入城點 7，
因為城 7 的鏡頭夾限是逐 16×16 區塊的，選錯玩家會被推回別的區。

## 這份收據不宣稱什麼

- 不宣稱可以從新遊戲一路玩到結局。接任務與回報各自走得通（`quest-take`、
  `full-chain` 那條路上的交付），但要在遊戲裡累積到 23 件仍然不現實，
  所以結局那一段的存檔是工具造的。`full-chain-ending` 證明的是
  「打完最終任務那一刻結局會出來」，不是「玩得到那一刻」。
- MOUTH 的輸入框已經做了，但**版面沒有對拍**：讀入上限與轉小寫有證據，
  游標的樣子與座標沒有。
- 不宣稱長路徑上的每一張畫面都與原版一致。`full-chain` 與 `full-chain-reload`
  的隊伍是六個人，原版沒有取到對應的畫格；連上對拍證據的是 `flow-quest` 那兩張，
  以及各段自己的短路徑收據。
- 打輸之後怎麼處置還沒解（RE-091 只讀到那一句訊息），這一版打贏打輸都回野外。
- 角色選項頁的 EDIT PICTURE／EDIT SHIELD，以及建築內的 SCROLL、馬廄商品表、
  角色資料頁上除了 U-TURN 的三個圖示按下去不會轉場，只會記進未實作清單。
- 角色資料頁只有英文模式。中文版面還沒做，中文模式按 MIRROR 只會記一筆。
- 任務一覽底下那一頁（`sub_16239` 畫的版面）與原版的疊法沒有對拍；
  格子的座標與框的顏色有證據（RE-111）。
- 競技場的遭遇訊息與戰場地形都沒有對拍：原版在開打前把 `word_34968` 設成 1
  （走任務地點那一條），這一版照那個旗標推了訊息、地形則沿用隨機遭遇的逐格擲。
  下注那一段還沒實作。
- 不宣稱鏡頭跟隨規則已證實。dead zone 是由 22 張原版畫格反解的**強推論**，
  規則本身沒有從程式碼讀出來。
- 野外的雲層動畫（`sub_2A18E` 的版面）還沒做。
