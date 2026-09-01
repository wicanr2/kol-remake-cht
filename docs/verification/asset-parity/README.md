# 抽取資產對原版執行畫面的同狀態收據

更新日期：2026-09-01

`tools/extract_ega_assets.py` 產生的 PNG 必須就是原版自己畫在螢幕上的像素，否則
remake 的畫面對拍從一開始就站不住。本目錄保存這一層的收據：拿抽取資產去對
`../character-ui/` 裡由原版執行擷取的同狀態畫面。

比對只在 RGB 通道進行，資產的透明像素不參與。

## 收據

| 檔案 | 內容 | 結果 |
|---|---|---|
| `charpics-templates-run1.json` | 80 張 `CHARPICS` 模板 對 `picture-cycle-run1` 的圖像層畫面 | 80/80 逐像素相同 |
| `charpics-templates-run2.json` | 同上，對第二輪擷取 | 80/80 逐像素相同 |
| `bookpics-cover.json` | `BOOKPICS[0]` 對書本封面畫面 | 64,000 像素中 50 個不同 |
| `chardata-write-run1.json` | 完成一次角色建立後的 `CHARDATA` 逐 offset 差異 | 只有 slot 14 變動 |
| `chardata-write-run2.json` | 同上，第二輪 | 只有 slot 14 變動 |

## 判讀

- **`CHARPICS` 80/80 完全命中。** 原版在圖像選擇畫面把 80 張模板依序畫在同一個
  40×38 矩形，blit 座標由 `tools/locate_asset_in_frame.py` 搜出為 (208, 106)、差異
  為 0。兩輪擷取各 80 張全部逐像素相同，等於同時證實 LZW → high-nibble-first 解包 →
  BIOS mode 0Dh palette → PNG 這條鏈沒有偏差。
- **`BOOKPICS[0]` 的 50 個差異像素集中在 (160, 100) 起的 8×11 區塊**，就是原版畫在
  螢幕中央的滑鼠游標。資產本身不含游標，因此這是預期差異，不是解碼誤差。
- **`CHARDATA` 差異只出現在 slot 14。** 差異清單是「值有變動的 byte」，不是「writer
  寫過的 byte」：writer 先清零整個 768-byte slot 再寫入，因此與舊值相同的欄位不會出現
  在差異裡。這份收據用來確認寫入落在單一槽位與已知欄位範圍，不用來反推 writer 覆蓋率
  （覆蓋率證據在 `docs/re/character-creation.md`）。

## 尚未涵蓋

- 其餘 29 個 archive 的資產目前只有「與已版控 visual probe 紀錄的 SHA-256 相符」這層
  驗證，還沒有各自的原版同狀態畫面收據。要補齊需要能讓原版把該素材畫出來的擷取情境。
- remake 側的畫面對拍尚未開始；本目錄只證實抽取資產對原版，不涉及 remake。
