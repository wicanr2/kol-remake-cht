# remake 對 DOS EGA 原版的畫面對拍

更新日期：2026-09-01

繁體中文重製不可能在文字上逐像素等於原版英文字模，因此對拍分兩層，而不是一個
「像素相同／不同」的是非題：

- **素材與版面層**：允許差異區以外的每個像素都必須相同。差一個像素就是失敗。
- **允許差異區**：由計畫明文列出並附理由（原版畫的滑鼠游標、原版自繪的文字區）。
  允許不等於忽略——收據會記錄每個允許區內實際差了多少像素，數字異常變動就看得出來。
- **必須相同區**：優先於允許差異區，用來在「文字可以不同」的大片區域裡仍然釘住素材
  必須逐像素相同的矩形，例如角色圖像的 blit 位置。

## 產生方式

```
tools/run_kol_headless.sh -shot-dir docs/verification/remake-parity/shots
python3 tools/verify_remake_parity.py \
    --plan docs/verification/remake-parity/plan-book-screens.json \
    --out docs/verification/remake-parity/book-screens.json

tools/run_kol_headless.sh -script "t e s t RET RET RET RET" \
    -shot-dir docs/verification/remake-parity/shots/newchar-steps
python3 tools/verify_remake_parity.py \
    --plan docs/verification/remake-parity/plan-newchar-steps.json \
    --out docs/verification/remake-parity/newchar-steps.json
```

remake 以 Go 1.26.7 ＋ Ebitengine v2.9.9 建置於固定映像
`eob-remake-go:1.26.7-ebiten2.9.9-audio`，在容器內以 Xvfb ＋ Mesa llvmpipe 軟體算繪，
輸出的是 320×200 原版解析度離屏畫布，因此不受視窗縮放與 GPU 差異影響。

原版側畫面來自 `../character-ui/`，是 `KOL.EXE` 正常玩家路徑的擷取。

## 目前結果

`book-screens.json`：83 對，83 通過，0 失敗。
`newchar-steps.json`：9 對，9 通過，0 失敗。

| 對拍項目 | 允許差異區實際差異 | 結果 |
|---|---|---|
| 書本封面 | 原版滑鼠游標 50 px | 通過 |
| 目錄頁 | 原版滑鼠游標 55 px | 通過 |
| 空白 New Character 頁 | 名字欄殘留字元 31 px | 通過 |
| New Character 角色圖像 selector 0–79（80 對） | 游標 55 px、左頁文字 6,524 px、右頁文字與選單 3,979 px | 通過 |

- 書本封面與目錄頁在游標以外**完全相同**：這兩個畫面在原版是整張烘好的 BOOKPICS
  素材，remake 直接畫同一張抽取 PNG。
- 80 張 New Character 畫面的角色圖像 blit 矩形 (208, 106, 40×38) 列為必須相同區，
  80 對全部逐像素相同。
- **空白 New Character 頁是第一個含文字的逐像素對拍**：remake 以抽出的原版字型、
  機械量得的版面座標與原版的繪字規則（逐字模格清白底、空白字元不清底）重繪整頁。

## 角色建立流程逐畫面對拍

remake 以固定腳本 `t e s t RET RET RET RET`（seed 1）重播名字→種族→性別→職業→
擲屬性五層，九張畫面對原版 `character-ui/reroll-run1` 的同狀態畫格。文字、選單、
標籤高亮與輸入游標全部由狀態機畫出來，不是預先烘好的圖。

| 步驟 | 原版畫格 | 允許差異區實際差異 | 結果 |
|---|---|---|---|
| 進入 New Character 頁 | frame-065 | 頁尾殘影 112 px | 通過 |
| 名字輸入 T／TE／TES／TEST | frame-068／071／074／077 | 頁尾殘影 112 px、名字游標 8 px | 通過 |
| 種族層 | frame-079 | 頁尾殘影 112 px、滑鼠 55 px | 通過 |
| 性別層 | frame-080 | 頁尾殘影 112 px、滑鼠 55 px | 通過 |
| 職業層 | frame-081 | 頁尾殘影 112 px、滑鼠 55 px | 通過 |
| 擲屬性層 | frame-082 | 頁尾殘影 112 px、滑鼠 55 px、數值欄 163 px | 通過 |

四類允許差異的成因都查證過（RE-045）：

- **頁尾殘影**是原版行為，不是 remake 缺漏。原版把展開跨頁貼到前一個書本畫面之上時，
  左頁底緣有一條未被覆蓋的細帶，帶內每一個像素都等於 `BOOKPICS[0]`。差異是雙向的，
  所以不是素材解碼錯誤；貼圖矩形怎麼切還沒反組譯確認。
- **名字輸入游標**會閃爍，取樣落在哪一相位由畫面靜止的時點決定；游標的格子位置本身
  仍受驗證——跑到別的格子就會落在允許區外。
- **滑鼠指標**是 DOSBox 畫的，remake 沒有這個元素。
- **屬性數值欄**是亂數：原版以 BIOS tick 播種，remake 以固定 seed 重播，骰式相同、
  點數不同。標籤與冒號不在允許區內，因此擲屬性層十列標籤由青轉棕的高亮規則仍逐像素
  受檢。

## 這份收據不宣稱什麼

- 不宣稱角色建立已完成。存檔寫入、16 槽管理、重名與名字欄字元集合都還沒對拍。
- 不宣稱動畫對拍。TITLE 標題序列與翻頁動畫尚未實作，也還沒建立逐幀時間軸的比較方式。
- 不宣稱繁中排版已決策。目前 remake 畫的是原版英文字模，不是譯文；雙模式（英文字模
  嚴格對拍模式＋中文大字格模式）已決策方向但尚未實作。
