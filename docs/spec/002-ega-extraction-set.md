# SPEC-002 EGA extraction set

- 狀態：`DRAFT`
- 日期：2026-09-01
- 母體：`docs/re/image-asset-inventory.json` 的 30 個邏輯 archive、1,022 個 payload
- 適用範圍：固定原版 ZIP 的 EGA `PICS`；不含 PCGA／HERC、無副檔名固定記錄表或玩法資料

本表是 SPEC-002 的完整 EGA 母體，不是 production manifest。`已證實來源` 表示來源
payload 已有足夠契約；多筆彩圖／遮罩可合成一張 runtime PNG，所以不必等於 PNG 數。
`待證實來源` 一律使用 `excluded_unknown`，在證據升級前不得輸出 production PNG。
`非圖片來源` 使用 `excluded_non_image`，不得將 placeholder 當素材。

| 邏輯 archive | payload | 已證實來源 | 待證實來源 | 非圖片來源 | 已證實 PNG 單位 | 轉換／合成 | 主要證據 |
|---|---:|---:|---:|---:|---:|---|---|
| AWD1PICS | 24 | 24 | 0 | 0 | 24 | direct color | `awd1pics-ega-visual-probe.json` |
| BACKPICS | 14 | 14 | 0 | 0 | 14 | direct color | `title-backpics-ega-visual-probe.json` |
| BOOKPICS | 14 | 14 | 0 | 0 | 14 | direct color | `kolbook-bookpics-ega-visual-probe.json` |
| CHARPICS | 80 | 80 | 0 | 0 | 80 | direct color | `charpics-archive-ega-visual-probe.json` |
| COMBPICS | 7 | 7 | 0 | 0 | 7 | direct color | `combpics-all-ega-visual-probe.json` |
| ENDPICS | 7 | 7 | 0 | 0 | 7 | direct color | `koltown-endpics-ega-visual-probe.json` |
| FACEPICS | 6 | 6 | 0 | 0 | 6 | direct color | `title-set-face-ega-visual-probe.json` |
| HANDPICS | 24 | 24 | 0 | 0 | 12 | color＋AND-mask | `handpics-ega-visual-probe.json` |
| HEADPICS | 108 | 108 | 0 | 0 | 54 | color＋AND-mask | `bodypics-ega-visual-probe.json` |
| KOLMPICS | 83 | 3 | 68 | 12 | 3 | color；圖號 28 是 AND-mask data PNG | `kolmpics-all-ega-visual-probe.json` |
| KOLPICS | 14 | 5 | 3 | 6 | 5 | direct color；raw-planar 暫排除 | `kolpics-ega-visual-probe.json` |
| LEGSPICS | 108 | 108 | 0 | 0 | 54 | color＋AND-mask | `bodypics-ega-visual-probe.json` |
| MISCPICS | 6 | 5 | 0 | 1 | 5 | direct color | `miscpics-ega-visual-probe.json` |
| MONPICS | 64 | 64 | 0 | 0 | 32 | color＋AND-mask | `monpics-ega-visual-probe.json` |
| PERPICS | 12 | 12 | 0 | 0 | 6 | color＋AND-mask | `perpics-ega-visual-probe.json` |
| SET0PICS | 6 | 6 | 0 | 0 | 6 | direct color | `title-set-face-ega-visual-probe.json` |
| SET1PICS | 6 | 6 | 0 | 0 | 6 | direct color | `title-set-face-ega-visual-probe.json` |
| SET2PICS | 6 | 6 | 0 | 0 | 6 | direct color | `title-set-face-ega-visual-probe.json` |
| SHLDPICS | 8 | 8 | 0 | 0 | 4 | color＋AND-mask | `shldpics-ega-visual-probe.json` |
| TORPICS | 108 | 108 | 0 | 0 | 54 | color＋AND-mask | `bodypics-ega-visual-probe.json` |
| TWN0PICS | 25 | 25 | 0 | 0 | 25 | direct color | `twn-archives-ega-visual-probe.json` |
| TWN1PICS | 34 | 33 | 0 | 1 | 33 | 圖號 0 color＋共享 mask | `twn-archives-ega-visual-probe.json` |
| TWN2PICS | 30 | 29 | 0 | 1 | 29 | 圖號 0 color＋共享 mask | `twn-archives-ega-visual-probe.json` |
| TWN3PICS | 26 | 25 | 0 | 1 | 25 | 圖號 0 color＋共享 mask | `twn-archives-ega-visual-probe.json` |
| TWN4PICS | 26 | 25 | 0 | 1 | 25 | 圖號 0 color＋共享 mask | `twn-archives-ega-visual-probe.json` |
| TWN5PICS | 30 | 29 | 0 | 1 | 29 | 圖號 0 color＋共享 mask | `twn-archives-ega-visual-probe.json` |
| TWN6PICS | 26 | 25 | 0 | 1 | 25 | 圖號 0 color＋共享 mask | `twn-archives-ega-visual-probe.json` |
| TWN7PICS | 46 | 45 | 0 | 1 | 45 | 圖號 0 color＋共享 mask | `twn-archives-ega-visual-probe.json` |
| WEAPPICS | 56 | 56 | 0 | 0 | 28 | color＋AND-mask | `weappics-ega-visual-probe.json` |
| WEATPICS | 18 | 18 | 0 | 0 | 18 | 三層 data PNG；另可重生六張 preview | `koltown-weatpics-layered-visual-probe.json` |
| **總計** | **1,022** | **925** | **71** | **26** | **681** |  |  |

## 尚未關閉的 71 筆

1. `KOLMPICS` 68 筆 24×20 圖示：全部機械展開為 480 pixels，且共用 helper 已證實
   24×20；但動態 caller 尚未逐圖號封閉，因此目前逐筆 `geometry_grade` 是
   `strong inference`。下一窄任務是證實 router 的 active index 集合確實都走同一
   24×20 consumer，或將無法證實者保留排除。
2. `KOLPICS` 圖號 0、8、12：raw-planar 轉換已證實；圖號 8／12 的 64 欄已由
   `y<<6+x` 證實，但 160 列仍只由有效資料長度與目視推得。圖號 0 的 256×5 亦未有
   完整矩形 consumer。三筆在正式抽取器中必須先排除。

## 已關閉的 26 筆非圖片來源

- `KOLMPICS`：12 筆一-byte placeholder。
- `KOLPICS`：6 筆一-byte placeholder。
- `MISCPICS`：1 筆一-byte placeholder。
- `TWN1PICS` 至 `TWN7PICS`：各 1 筆圖號 1 placeholder，共 7 筆。

## 升級條件

- 本表的 archive 名稱、payload 數與總和必須由 verifier 對清冊重算，禁止手工漂移。
- 待證實來源降為 0，或由使用者明確縮小 extraction set 並接受排除，才可通過
  「每種預定輸出格式皆有已證實契約」的 READY gate。
- 顯示 theme 仍須使用者決定；本表只證實 EGA 母體，不替使用者排除或承諾 CGA／Hercules。
- production 抽取器必須原子化輸出、寫 manifest、重讀 PNG 並執行權利 deny-list；
  本表本身不授權實作。
