# TITLE 標題序列的原版畫面時間軸

更新日期：2026-09-01

`KOL.EXE` 在沒有按鍵時會執行 `TITLE.EXE`，播放一段會循環的標題序列。動畫對拍要先
知道原版每一幀畫了什麼，因此這裡保存兩輪各 100 秒的純觀察擷取，以及用抽取資產逐張
比對得到的資產對照表。

## 產生方式

```
tools/run_character_ui_oracle.sh <輸出目錄> --scenario prelude --observe 100
python3 tools/identify_frame_assets.py --frames <目錄>/frames \
    --archives BACKPICS,KOLMPICS,KOLPICS,MISCPICS,BOOKPICS --out asset-map-runN.json
```

觀察模式全程不送任何按鍵，因此擷取到的就是原版自己的循環播放。資產對照只回報逐像素
完全命中的結果，不做近似匹配。

## 已證實的結構

兩輪各擷取到 40／41 個去重畫格，其中 9 個與 `BACKPICS` 的整張全螢幕素材逐像素相同：
圖號 4、5、6、7、8、9、10、11、12。

第一輪的時間軸顯示序列為：

| 時間（秒） | 畫面 |
|---:|---|
| 7.98 | `BACKPICS[6]` |
| 10.1–16.2 | 在三個畫格之間反覆切換的閃爍段 |
| 17.30 | `BACKPICS[6]` |
| 19.13 | `BACKPICS[7]` |
| 23.06 | `BACKPICS[8]` |
| 27.3–33.8 | 閃爍段 |
| 34.77 | `BACKPICS[6]` |
| 36.67 | `BACKPICS[9]` |
| 40.47 | `BACKPICS[10]` |
| 44.9–52.6 | 閃爍段 |
| 53.52 | `BACKPICS[6]` |
| 55.40 | `BACKPICS[11]` |
| 59.48 | `BACKPICS[12]` |
| 63.9–79.8 | 閃爍段，其間出現 `BACKPICS[4]`、`BACKPICS[5]` |
| 81.41 | `BACKPICS[10]` |
| 85.48 | `BACKPICS[6]` |
| 94.07 | `BACKPICS[6]`（下一輪起點） |
| 95.96 | `BACKPICS[7]` |

`BACKPICS[6]` 反覆出現在各段落之間；序列在約 86 秒後回到同一畫面重新開始。

## 尚未關閉

- 閃爍段的機制未確認：那些畫格在少數幾張之間高頻交替，可能是重繪、色盤動畫或
  逐段清除，成因要由 `TITLE.EXE` 的 IDA 證據判定，不能由畫格外觀推定。
- 畫格時間戳是主機牆鐘，且擷取取樣間隔 0.3 秒，因此上表的時間是觀察值而非原版的
  設計時序。要做逐幀時序對拍必須另外固定 DOSBox 週期與取樣方式。
- remake 尚未實作標題序列，本目錄只有原版側證據。
