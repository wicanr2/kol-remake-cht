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
```

remake 以 Go 1.26.7 ＋ Ebitengine v2.9.9 建置於固定映像
`eob-remake-go:1.26.7-ebiten2.9.9-audio`，在容器內以 Xvfb ＋ Mesa llvmpipe 軟體算繪，
輸出的是 320×200 原版解析度離屏畫布，因此不受視窗縮放與 GPU 差異影響。

原版側畫面來自 `../character-ui/`，是 `KOL.EXE` 正常玩家路徑的擷取。

## 目前結果

`book-screens.json`：82 對，82 通過，0 失敗。

| 對拍項目 | 允許差異區實際差異 | 結果 |
|---|---|---|
| 書本封面 | 原版滑鼠游標 50 px | 通過 |
| 目錄頁 | 原版滑鼠游標 55 px | 通過 |
| New Character 角色圖像 selector 0–79（80 對） | 游標 55 px、左頁文字 6,524 px、右頁文字與選單 | 通過 |

- 書本封面與目錄頁在游標以外**完全相同**：這兩個畫面在原版是整張烘好的 BOOKPICS
  素材，remake 直接畫同一張抽取 PNG。
- 80 張 New Character 畫面的角色圖像 blit 矩形 (208, 106, 40×38) 列為必須相同區，
  80 對全部逐像素相同。

## 這份收據不宣稱什麼

- 不宣稱 remake 已實作角色建立玩法。目前只畫底圖與角色圖像，文字、選單、輸入與
  狀態機都還沒接上，因此文字區才會整片列為允許差異。
- 不宣稱動畫對拍。TITLE 標題序列與翻頁動畫尚未實作，也還沒建立逐幀時間軸的比較方式。
- 不宣稱字型決策已完成。原版書本字模仍未從 `KOLBOOK.EXE` 取出；繁中字型、文字安全
  矩形與是否保留一個英文字模模式供嚴格像素比對，都還沒決定。
