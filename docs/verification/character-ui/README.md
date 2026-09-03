# 書本與城鎮的正常 DOS 玩家路徑收據

更新日期：2026-09-01

本目錄保存 KOLBOOK 書本流程與 KOLTOWN 城鎮開場在正常玩家路徑下的動態 oracle 收據。所有結論來自
`KOL.EXE` 啟動、鍵盤操作、逐畫面擷取與三個角色檔的前後 SHA-256，不使用除錯入口、
座標注入或測試旗標。

## 產生方式

```
tools/run_character_ui_oracle.sh <輸出目錄> --scenario <情境> [--save-character-files]
tools/compare_character_ui_runs.py <第一輪目錄> <第二輪目錄> --out cross-run/<情境>.json
```

- 容器映像 `rich2-dosbox-dbg:latest`（DOSBox-X 2026.06.02 SDL2），`docker run --rm
  --network none`，`--memory 1g --cpus 1 --pids-limit 256`。
- 原版 ZIP 唯讀掛載，SHA-256 不符即失敗關閉；遊戲目錄每次由 ZIP 重建於容器 tmpfs。
- 機器設定固定為 `machine=ega`、`memsize=4`、`core=normal`、`cputype=386`、
  `cycles=fixed 3000`，設定檔本身的 SHA-256 記在每份收據內。
- 按鍵以 X11 XTEST 送出（`tools/x11_key_inject.py`）。送鍵時機不是盲打時序，而是
  等畫面靜止到指定秒數才送下一鍵，因此原版磁碟解碼時間的變動不會讓序列走偏。
- 情境定義送出的是小寫字母；畫面與 `CHARDATA` 中的名字也是小寫，因此本批收據不對
  大小寫保存行為下任何結論。
- 畫面由 Xvfb 的 `-fbdir` framebuffer 取樣，依 DOSBox-X 視窗實際幾何裁切，逐畫格
  去重後存成 PNG。

## 目錄結構

每個情境有兩個獨立執行的目錄 `<情境>-run1/`、`<情境>-run2/`，各含 `receipt.json`
與 `frames/`；收據以相對路徑索引畫格。`cross-run/` 是兩輪檢查點畫面的比對結果，
`MANIFEST.json` 記錄來源、產生工具、雜湊入口與權利分類。

3,117 個去重畫格 PNG 已依使用者 2026-09-01 的決定提交至此 private 儲存庫。畫面內容
衍生自原版執行檔與資料，著作權仍屬原權利人；此決定不涵蓋 public repo、public fork、
GitHub Release 或公開 `dist-all/.../release/`。畫格編號是擷取順序，不是原版 archive 圖號。

## 情境與結果

| 情境 | 檢查點 | 兩輪相同 | 三檔變動 |
|---|---:|---:|---|
| `name-length` | 26 | 25 | 無 |
| `blank-name` | 10 | 10 | 無 |
| `escape-layers` | 28 | 21 | 無 |
| `reroll` | 20 | 12 | 無 |
| `accept-and-save` | 16 | 12 | `CHARDATA` |
| `picture-cycle` | 99 | 10 | 無 |
| `delete-character` | 13 | 13 | 無 |
| `delete-character-confirm` | 15 | 15 | `CHARDATA` |
| `edit-picture-probe`／`edit-shield-probe`／`harbor-master-probe` | 10／11／12 | 單輪 | 無 |
| `edit-picture-enter`／`edit-shield-enter`／`harbor-master-enter` | 12／13／14 | 單輪 | 無 |
| `play-the-game`／`town-icon-bar`／`town-icon-move`／`town-recruit-companion` | 8／12／13／16 | 單輪 | 無 |
| `town-outdoor-move` | 14 | 單輪 | 無 |
| `town-rest-save`／`town-rest-two` | 14／19 | 單輪 | `CHARDATA`／無 |
| `town-mirror` | 13 | 單輪 | 無 |
| `town-shop` | 66／72 | 兩輪 | 無 |

兩輪之間的差異全部落在三類本來就不可能逐 byte 相同的畫面：

| 差異來源 | 不同 byte 比例 |
|---|---|
| TITLE 標題序列（循環動畫，固定等待落點不同） | 13.98% |
| 名字欄閃爍游標 | 0.0125% |
| 屬性頁數值（BIOS tick 播種的擲骰） | 0.1%–0.25% |

`picture-cycle` 的 99 個檢查點中有 89 個落在屬性數值仍顯示於左頁的圖像層，因此逐 byte
不同；該情境真正要驗的角色圖像區塊另由
`../asset-parity/charpics-templates-run{1,2}.json` 確認兩輪都 80/80 逐像素相同。

## 已證實的行為

- **名字欄容量 16 個字元，且輸入第 16 個字元會直接前進到種族層，不需 Return。**
  連續送出 17 個字母時，畫面在第 15 個字元顯示 `abcdefghijklmno`，第 16 個字元顯示
  `abcdefghijklmnop` 並同時開啟種族選單；第 17 個字元與其後兩次 Backspace 都落在
  種族層、對畫面沒有作用。
- **空名字不會前進。** 在名字欄不輸入任何字元直接按 Return，畫面停在名字層。
  這與 KOLBOOK 靜態控制流「輸入 buffer 首 byte 為零時回到同一輸入點」一致。
- **各層取消的返回目的地**：種族→名字、性別→種族、職業→性別、屬性→職業，
  再往上依序回到性別、種族、名字層。回到名字層時，名字與右頁摘要欄位都已清空。
- **屬性頁的 REROLL STATS 會重新產生整組數值。** 同一個 HUMAN／MALE／BARBARIAN
  角色連續兩次 reroll，力量 88 → 85 → 89、體型 74 → 72 → 73、健康 75 → 75 → 74。
- **角色圖像 selector 恰好循環 80 張。** 由 selector 0 向左一次得到的畫面，與由 0
  連續向右 80 次得到的畫面相同；再向右一次回到 selector 0。方向鍵瀏覽、Return 接受。
- **走出書本的鏈已量測。** 目錄頁按 `6`（Play the game）→ 同一張角色選擇清單 →
  選領隊 → 城鎮（半途之家旅店）。城鎮的圖示列**不吃方向鍵**，要用參考卡指定的
  `<`／`>` 移動白框（走到最後一格會繞回第一格），Return 執行。
- **組隊流程**：COMPANION 圖示 → 先問是否搜尋硬碟角色 → 列出該旅店裡不在隊伍中的
  角色 → 上下方向鍵選、Return 加入；加入後頭像出現在頂端。**招募不寫磁碟**，
  三個角色檔全程未變（RE-053）。
- **書本四個功能共用同一張角色選擇清單。** `DELETE CHARACTER`、`EDIT PICTURE`、
  `EDIT SHIELD`、`HARBOR MASTER` 進入後的第一張畫面是同一個 PNG（四份收據的 SHA-256
  完全相同）；由清單 ESC 回到選單（反白重置），再 ESC 回到目錄頁。
- **HARBOR MASTER 在只安裝一個 region 時是一張訊息頁**，按鍵回到角色選項。
  `EDIT PICTURE` 與 `EDIT SHIELD` 是同一個滑鼠驅動的點陣圖編輯器，ESC 沒有作用，
  現有的鍵盤 oracle 走不完（RE-052）。
- **DELETE CHARACTER 用的就是角色清單頁。** 版面與 `LIST CHARACTERS` 完全相同，
  差別只有顏色語意：反白那一列黑字、其餘青字，左右兩欄同時反白；反白條起點是第 0 列，
  不繞圈。Return 叫出 (94,58) 的 132×24 確認視窗（黑框白底、右下 4 px 投影），
  視窗內兩行暗紅字 `ARE YOU SURE?` 與 `YES/NO`；視窗開著時清單沒有任何一列反白。
- **確認視窗不吃方向鍵與 Return。** 送出這兩個鍵之後畫面逐像素沒有變化；`y` 刪除、
  `n` 取消，兩者都回到刪除清單並把反白條重置為第 0 列。刪除只把該槽位名字的第一個
  byte 寫成 0（見 `../chardata-delete/`），名冊往上補齊是「只列已建立槽位」的結果。
- **接受圖像是保存點。** 走完「ACCEPT STATS → 接受圖像」之後畫面回到 Character
  options，`CHARDATA` 出現變動；本批兩輪都只寫入 slot 14，`CHARPICS` 與 `SHLDPICS`
  的 SHA-256 未變。逐 offset 差異見 `../asset-parity/chardata-write-run{1,2}.json`。
- 其餘四個情境全程沒有接受屬性或角色圖像，三個角色檔的前後 SHA-256 皆未變動。

## 尚未涵蓋

- 16 槽全滿時的訊息與退出行為。固定 ZIP 的 `CHARDATA` 十六槽都不是全零，前十四槽
  有名字，第 15、16 槽名字為空但記錄非零；本批的保存落在 slot 14，支持「以名字首
  byte 判斷槽位可用」的假說，但滿槽分支本身仍只有靜態控制流證據。
- 重名處理、名字欄接受的字元集合與大小寫保存方式。
- 選擇非 selector 0 的圖像時 `CHARPICS` 的寫入內容；本批兩輪都取 selector 0，該槽
  原有內容與寫入值相同，因此檔案雜湊沒有變化。
