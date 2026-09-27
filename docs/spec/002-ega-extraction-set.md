# SPEC-002 EGA extraction set

- 狀態：`READY`
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
| KOLMPICS | 83 | 71 | 0 | 12 | 71 | color；圖號 28 是 AND-mask data PNG | `kolmpics-all-ega-visual-probe.json` |
| KOLPICS | 14 | 8 | 0 | 6 | 8 | direct color＋renderer-data PNG | `kolpics-ega-visual-probe.json` |
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
| **總計** | **1,022** | **996** | **0** | **26** | **752** |  |  |

## 來源關閉狀態

固定 EGA 母體目前為 0 筆。KOLMPICS 的 68 筆圖示已由 480-index 全批證據、兩條固定
24×20 consumer 與完整接觸表共同關閉 canonical extraction shape；動態 caller 的螢幕
位置／裁切仍是 UI parity 限制，不再混入素材格式 gate。
## 已關閉的 26 筆非圖片來源

- `KOLMPICS`：12 筆一-byte placeholder。
- `KOLPICS`：6 筆一-byte placeholder。
- `MISCPICS`：1 筆一-byte placeholder。
- `TWN1PICS` 至 `TWN7PICS`：各 1 筆圖號 1 placeholder，共 7 筆。

## 升級條件

- 本表的 archive 名稱、payload 數與總和必須由 verifier 對清冊重算，禁止手工漂移。
- 待證實來源降為 0，或由使用者明確縮小 extraction set 並接受排除，才可通過
  「每種預定輸出格式皆有已證實契約」的 READY gate。
- 顯示 theme 已由使用者於 2026-09-01 決定為首版 EGA-only；本表即首版 extraction set，
  CGA／Hercules 若日後納入必須另立母體與 gate。
- production 抽取器必須原子化輸出、寫 manifest、重讀 PNG 並執行權利 deny-list。
  `tools/extract_ega_assets.py` 已滿足這些條件並產出全部 752 個 PNG 單位，逐張與
  visual probe 紀錄的 SHA-256 相符。
