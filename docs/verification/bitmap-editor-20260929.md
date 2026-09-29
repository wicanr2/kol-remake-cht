# 點陣圖編輯器與原版同狀態比較（2026-09-29）

## 結果

書本 EDIT PICTURE 的點陣圖編輯器（SPEC-036），remake 與原版 KOLBOOK 從同一個進場狀態送
同一串鍵、答 y 存檔，寫回的 760 byte 逐 byte 相同；進場畫面除原版左上角 DOS 滑鼠指標的
57 個像素外逐像素相同。

按鍵涵蓋：工具列換色（方向鍵）、ESC 切到畫布、`d` 落筆後移動、`u` 抬筆、四個斜走鍵、
`x` 與數字鍵盤 5 畫點、`c` 吸色、ESC 切回工具列、四支變換（右移、上移、左右鏡射、上下鏡射）
各執行一次、白框繞回第 0 格離開、答 `y`。

```
tools/run_original_parity_tests.sh        # TestOriginalEditorSameKeys
```

## 原版那一側

`tools/oracle/book_editor_test.go`（`KOL_CHECK=editor` 跑 `tools/oracle/check_updated_dosgolem.sh`）：
正常書本路徑封面 → 目錄頁 5 → 下移三格 EDIT PICTURE → 選角清單第一名（Theodoric）→ 編輯器；
進場時取工作緩衝 `word_19B52:0`，存檔後取 CHARPICS 陣列 `DS:0x1844 + word_238FC`。存檔之後
原版回到選角清單。

| 輸入 | SHA-256 |
|---|---|
| `KOLBOOK.EXE` | `2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832` |
| dosgolem | `~/cht/dosgolem-kolrand` `968ca77` |
| 收據 `workplace/bitmap-editor-20260929/bitmap-editor.json`（只留本機） | `adbde411d4683c338e463cffcdd4ccb63a2abec96b5d0acc6da6d8de74464f1c` |
| 原版進場畫格 `editor-enter.png`（只留本機） | `6c8671173aeac6dce4a14d22059b0658d6e21efa4acff5002f6fbef868df5cc2` |

## remake 那一側

- `TestOriginalEditorSameKeys`（`internal/bitmapedit`）：從收據的進場緩衝起跑同一串鍵。
- `TestEditShieldSavesAndReloads`（`cmd/kol`）：建角 → EDIT SHIELD 從預設盾牌起編 → 存檔
  失敗整筆退回 → 成功後重讀名冊帶回改過的盾牌。
- 畫面：`-scene flow -flow-script "SPACE SPACE 5 D D D RET RET"` 第 8 步與原版進場畫格逐像素比較。

## 不宣稱

- 滑鼠操作依靜態讀碼實作，沒有與原版實跑比較。
- EDIT SHIELD 的原版實跑只比了同一支編輯器的 EDIT PICTURE 那一條；兩者程式相同，只差編輯的檔案。
- 盾牌在資料頁全身立繪上的位置依靜態讀碼（`sub_18F3A:0x192D6`），沒有原版畫格。
- 游標閃爍週期是近似。
