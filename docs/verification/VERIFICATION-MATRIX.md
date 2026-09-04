# 驗證矩陣

每一列的「結果」以 **code 與收據**為準，不以工作清單的勾選為準。沒有收據的一律
寫「尚無」，不寫「應該可以」。

| 子系統 | 單元／真實資料測試 | 原版 oracle | 視覺 oracle | 玩家路徑 | 結果 |
|---|---|---|---|---|---|
| 原版輸入清冊 | ZIP CRC、逐檔 SHA-256、雙容器重生 byte compare | 不適用 | 不適用 | 不適用 | 通過 |
| KOLTOWN 解包研究映像 | header、2,281 relocations、IDA entry／337 functions | startup 前 129,312 bytes、五錨點全吻合 | 不適用 | `KOLTOWN.EXE A` 正常啟動至 entry | SPEC-001 READY |
| 標題序列 | `internal/title` 的腳本重播測試 | `TITLE.EXE` 的 `_main` 靜態腳本（RE-047） | 兩輪觀察的每一張畫格逐像素對上 | 標題 → 按鍵中止 | 已對拍；雷聲未接（見音訊列） |
| 書本畫面與翻頁 | `internal/book`、`internal/flow` | `KOLBOOK.EXE` `sub_114F5` 的表驅動捲動（RE-048） | 書本 3 ＋ 翻頁 54 對 | `full` 路徑前段 | 已對拍 |
| 角色建立 UI | 22 個情境、32 輪，檢查點畫面 SHA-256 交叉比對 | `KOL.EXE` 正常入口；DOSBox-X 2026.06.02 固定設定 | 建立 9 ＋ 圖像層 92 對 | 目錄頁 → 建立 → 接受屬性 → 接受圖像 → 保存 | SPEC-005 READY；16 槽全滿、重名、字元集合未涵蓋 |
| 角色寫入與刪除 | `DeleteFlow`／`DeleteSlot`；writer 逐 byte | 改動的 58 個 byte 兩個職業各 58/58 相同；刪除只寫一個 byte | 刪除流程 13 對 | 角色選項 → 刪除 → 確認 | SPEC-005 涵蓋；ESC、大寫 Y／N、空名冊未涵蓋 |
| 書本其他功能 | 尚無 | 四個功能共用同一張角色選擇清單；HARBOR MASTER 為三行訊息頁 | HARBOR MASTER 1 對 | 角色選項 → Harbor Master | `EDIT PICTURE`／`EDIT SHIELD` **未實作**：原版滑鼠驅動，鍵盤 oracle 走不完 |
| 城鎮移動、門口與邊界 | `Grid`／`Outdoor`／`Markers`／`Session` | 94 張移動 oracle 中 22 張畫格的鏡頭全部唯一解 | 戶外 22 ＋ 城門提示 4 對 | `town` 14 步、`town-edge` 14 步 | 已對拍 |
| 建築內部與對話 | `EnterBuilding`、`Topics` | 進門敘述八行逐格讀回與 5-bit 解碼逐字相同；409 個話題槽交叉檢查零例外 | 進門 1 ＋ 對話 3 對 | `town-talk` 10 步 | 已對拍。**訊息 10 起的後續對話觸發條件仍未解** |
| 招募同伴 | `Inn` 的清單捲動與名牌 | 候選人由角色檔所在地欄比對建築名稱（RE-072） | 旅店 8 ＋ 招募 9 對 | `town-recruit` 15 步 | 已對拍 |
| 住宿、治療與存讀檔 | `internal/save` 原子讀寫；`Rest` 三條路 | 價錢取自建築 payload `0x38`／`0x39`（RE-073） | 住宿 11 對 | `town-rest` 11 步（含存檔逐欄驗證）、`town-load` 7 步、`full-chain` ＋ `full-chain-reload` | SPEC-006 READY。任務旗標、已回報數與身上帶的東西已進存檔（RE-101／107／109）。**戰鬥中途的狀態與世界時間仍未進存檔** |
| 角色資料頁 | `Sheet` 的量尺與 5-bit 武器名 | 屬性是內部值 × 100 / 150；身高體重查三張 DGROUP 表（RE-074） | 鏡子 8 對 | `town-mirror` 10 步 | 已對拍。**物品、換人、法術三個圖示未實作** |
| 商店、背包與穿戴 | `Shop` 的商品表與合身判定 | 商品表在建築 payload `0x100` 起，一筆 16 byte（RE-075） | 商店 9 對 | `town-shop` 14 步 | 已對拍。**馬廄的商品表（一筆 0x80 byte）未實作** |
| 訓練 | `Train` 的位階閘門與技能表 | 四門技能在建築 payload `0x50` 起（RE-076） | 訓練 6 對 | `town-train` 15 步 | 已對拍。**加技能與鍛造兩條路原版存檔走不到，只有靜態證據** |
| 世界地圖、時刻與道路 | `internal/world` 的格子、次格、道路遮罩 | 城鎮記錄 16 組座標全部相同（RE-069）；循路一步扣 20 個時刻 | 野外 4 ＋ 時刻 4 ＋ 循路 10 ＋ 進城 7 對 | `world-road` 34 步、`world-clock` 28 步、`town-leave` 13 步、`town-round-trip` 20 步 | 已對拍。**雲層動畫列在 `not_claimed`**。隨機遭遇已實作（SPEC-011）：野外那兩條路徑因此**釘住亂數種子**，被伏擊就走不到終點——原版量測也一樣 |
| 任務與戰鬥 | `internal/battle` 十九個測試檔 | 規劃、指令列、選目標、執行順序、移動、近戰結算、訊息鏈、敵人 AI、擊退、勝負、戰後結算、施法、射擊、任務旗標與逃走逐條反組譯（RE-079…104） | 任務與戰場 4 ＋ 執行段訊息 1 對 | `world-quest` 91 步；戰後結算 `battle-spoils/`、施法 `battle-spell/`、射擊 `battle-missile/`、逃走 `battle-flee/`（皆為觀察，非對拍） | SPEC-008／009／010 的已解出部分與 SPEC-012／013／014 READY。一場**雙向**打得完，戰後的冒險點數、金冠與戰利品清單都實跑過；敵人在近距離會射擊。**敵人的撤退與反制掃描、投擲、逃走之後掉東西未實作**；戰利品框的內部版面未對拍 |
| 結局 | `internal/ending` 一個測試檔 | `sub_2B4A7` 逐段抄出圖號、記錄與矩形；四個版面全域換算出來的七組寬高與 RE-017 從像素數反推的完全吻合（RE-106） | 無 | `docs/verification/ending/` 七張（觀察，非對拍）、`player-path/full-chain-ending` 28 步 | SPEC-015 READY。**已接進玩家路徑**：交出最後一件任務那一刻已回報數 22→23，畫面直接換成結局，七張與 `-scene ending` 的收據逐 byte 相同。接任務那一側（`sub_14564`）還沒實作，所以那條路徑的存檔是工具造的 |
| 圖像／主題 | production 抽取器逐張重讀，769 張 PNG 全部對上 visual probe 雜湊 | IDA consumer ＋ BIOS mode 0Dh | 80 張角色模板在 (208,106) 逐像素相同 | 由 `picture-cycle` 正常路徑取得對照 | SPEC-002 READY；首版 EGA-only |
| 在地化／字型 | 兩套原版字型各 93 字模抽出並重生 | HUMAN 五字模逐 byte 命中壓縮 payload | 空白 New Character 頁 64,000 px 只差 31 px | 由正常路徑取得對照 | 中文字模表 **2,434 字**、缺字 0；安全矩形 **6,685 項**全過；UI 文字庫 **486 筆**譯文全過 `verify_ui_translation.py` |
| 音訊 | 尚無 | 六個 executable 的 I/O 母體已窮舉；TITLE 的 WAV oracle 已取得 | 不適用 | 尚無 | **SPEC-004 仍 DRAFT，遊戲裡一個音都沒有**。卡在人耳驗收 |
| 封包 | 尚無 | 不適用 | 不適用 | 尚無 | 未開始；`dist-all/` 還不存在 |

## 截圖 metadata

原版執行畫面的逐畫面 PNG 由 `tools/dosbox_character_ui_oracle.py` 產生，
43 個 run、6,515 個去重畫格連同收據、跨輪比對與 provenance manifest 存於
`docs/verification/character-ui/`；抽取資產對原版畫面的收據見
`docs/verification/asset-parity/`。權利分類為 private-repo 可保存、不得公開散布。
remake 同狀態截圖存於 `docs/verification/remake-parity/shots/`，對拍計畫與收據同目錄
（二十五份收據 283 對，允許差異區之外零像素差）。

## 正常玩家路徑

`player-path/` 有十九份可重播收據。短路徑逐步存圖，要求每一步有可見效果、指定的
檢查點與已通過對拍的截圖逐 byte 相同；四條走到底的長路徑改存狀態軌跡加檢查點截圖
（`verify_full_path.py`），軌跡是遊戲自己報的狀態，不是從畫面猜的：

| 路徑 | 步數 | 內容 |
|---|---:|---|
| `town` | 14 | 城鎮內走動、門口、進建築 |
| `town-edge` | 14 | 地圖邊界的林地圖與城門提示、U-TURN 回地圖 |
| `town-talk` | 10 | 建築內的圖示列、LISTEN 的傳聞與 MOUTH 的問話 |
| `town-leave` | 13 | 在城門選 DOOR 離開城鎮，落到世界地圖 |
| `town-round-trip` | 20 | 城鎮 → 野外 → 退回入口 → 進城 |
| `town-recruit` | 15 | COMPANION 招募：搜尋硬碟的問句、候選人清單、加入 |
| `town-rest` | 11 | REST 住一晚：付 56 枚金冠、寫出存檔並逐欄驗證 |
| `town-mirror` | 10 | MIRROR 的角色資料頁與 U-TURN |
| `town-load` | 7 | 從 `town-rest` 的存檔開場，只有金幣那一欄不同 |
| `town-shop` | 14 | TRADE：反白、明細、買下、扣錢、穿上 |
| `town-train` | 15 | AWARD 訓練的位階閘門與技能選擇 |
| `world-road` | 34 | 出城上道路、ROAD 自動循路、SIGNPOST 換向 |
| `world-clock` | 28 | 野外的時刻推進與換日 |
| `world-quest` | 91 | 走到任務地點、遭遇畫面、進戰場、規劃與執行 |
| `full` | 26 | 標題 → 書本 → 角色選項 → 旅店 → 城鎮 → 建築，由 `internal/flow` 串起來 |
| `full-chain` | 567 | 標題 → … → 出城 → 野外遇伏 → 四個回合打贏 → 回城 → 住一晚並存檔 |
| `full-chain-reload` | 119 | 讀那份存檔 → 資料頁的金幣是住宿後的 158 → 出城 → 任務地點 → 參戰 |
| `flow-quest` | 109 | 同一條 flow 路徑走到任務地點，兩張與已對拍的截圖逐 byte 相同 |
| `full-chain-ending` | 28 | 交出最後一件任務 → 已回報數 22→23 → 結局七段（觀察，存檔是工具造的） |

尚未實作的功能會被記進 `flow.Unimplemented` 或 `battle.ActResult.Unimplemented`，
不會默默吞掉。

## 發行 gate

- [ ] 全庫測試與靜態分析
- [~] 翻譯／編碼檢查（城鎮 1,938 則與 UI 文字庫 486 筆已翻完並有檢查器；
      ENDTXT 的譯文已接上結局畫面，BOOKDATA 還沒接上書本）
- [~] 無除錯捷徑的正常玩家路徑（`full-chain` 一串按鍵從標題走到住宿存檔，中間打贏
      一場遭遇；`full-chain-reload` 讀回來繼續走到任務地點；`full-chain-ending`
      走到結局。**結局那一條的存檔是工具造的**——接任務那一側還沒實作）
- [~] 可寫存檔與 pristine original 不變檢查（存檔迴圈已閉合，任務旗標、已回報數與
      身上帶的東西都進存檔；戰鬥中途的狀態與世界時間還沒有。
      原版 `CHARDATA` 一律唯讀，匯入是離線工具）
- [~] 原版／remake 同狀態截圖審查（283 對通過；音訊與未實作的部分沒有可審的畫面）
- [~] 封包啟動 smoke（`dist-all/0.1.1/` 的 linux/amd64 與 windows/amd64 已由
      `tools/verify_dist.sh` 驗過 manifest、授權與啟動；macOS 與 Windows 真機未驗）
