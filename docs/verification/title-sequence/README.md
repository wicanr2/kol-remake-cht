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

## 序列的來源已從觀察改為靜態讀出

`TITLE.EXE` 的 `_main` 就是整段序列的程式碼，`tools/ida_probe_title_sequence.py` 已把它
與全部直接被呼叫者匯出（RE-047），重建結果在 `docs/re/title-sequence.json`。等待一律是
BIOS tick 數（全螢幕畫面之間 75 tick ≈ 4.12 秒），先前記為「機制未確認」的閃爍段是
SET0／SET1／SET2 的四圈疊圖動畫。

上面那份觀察時間軸現在的角色是**交叉核對**：靜態讀出的順序預測了它的每一個轉換點，
畫面間距的觀察值 3.80–4.08 秒與 75 tick 的 4.12 秒相差都在 0.3 秒的取樣間隔以內。

## 逐幀對拍能要求到什麼程度

序列裡有兩處亂數，種子是開機時的 BIOS tick：SET 動畫每圈的閃光用第 4 或第 5 張，
`BACKPICS[11]` 與 `[13]` 二選一。**原版自己每次播放就不一樣，因此不能要求逐幀完全一致。**
可以要求的是：資產逐像素相同、段落順序相同、tick 等待相同，以及在指定亂數序列下的
每一幀相同。

## remake 對拍結果

remake 依 `docs/re/title-sequence.json` 播放一輪，把每一個送出的畫面存成 PNG：

```
tools/run_kol_headless.sh -title-shot-dir workplace/shots/title-first  -title-choice first
tools/run_kol_headless.sh -title-shot-dir workplace/shots/title-second -title-choice second
python3 tools/verify_title_parity.py \
    --remake workplace/shots/title-first --remake workplace/shots/title-second \
    --original docs/verification/title-sequence/observe-run1/frames \
    --skip-first 4 --out docs/verification/title-sequence/parity-run1.json
```

`-title-choice first`／`second` 固定取兩處亂數的第一或第二個選項，兩次的聯集就涵蓋
序列會出現的所有畫面（各 92 張）。判準是**集合包含**：原版擷取到的每一張，remake 都要
能逐像素產生一模一樣的一張。

| 收據 | 原版畫格 | 逐像素相同 | 更新中途 | 未解釋 | 略過 |
|---|---:|---:|---:|---:|---:|
| `parity-run1.json` | 40 | 33 | 3 | 0 | 4 |
| `parity-run2.json` | 41 | 33 | 4 | 0 | 4 |

- **略過的 4 張**是 `TITLE.EXE` 開始繪製之前的畫面。
- **「更新中途」**指取樣落在整張搬移途中：畫面的一段（前綴或後綴的整數列）逐像素等於
  remake 的某一張，另一段還沒寫完。收據記下分割列與未寫完那段與最接近畫面的像素差，
  不對成因下結論。其中兩張（run1 的 frame-019、run2 的 frame-031）未寫完的那段差到
  一萬多個像素，代表取樣落在兩次更新重疊的時候，這個分類對它們的說服力較弱。
- **逐像素相同的 66 張涵蓋整段序列**：閃電、SET0／SET1／SET2 的每一張疊圖與兩種閃光、
  `BACKPICS[7]`–`[13]` 的全螢幕畫面，以及 FACE 的六張疊圖與兩次全螢幕閃光。

remake 側每一張畫面的來源記在 `remake-frames-first.json` 與 `remake-frames-second.json`，
每一筆都帶原版送出該畫面的呼叫位址與停留的 tick 數。畫面 PNG 由上面的命令重生，不入庫。

## 實際播放

`cmd/kol -scene title` 依同一份腳本實際播放：以 BIOS tick 推進、按鍵中止。
`-title-timeline` 匯出一輪的時間軸（`remake-timeline.json`）。

一輪的**程式碼時間**是 70.49 秒；原版在 DOSBox-X 觀察到的一輪約 77.5 秒
（第一次閃電 7.98 秒，下一輪 85.48 秒）。差距的來源是可指認的，不是版本不同：

- 原版每一段等待都是「先記下 tick，載入或繪製，再等到 tick ≥ 記錄值 + N」。載入與繪製
  發生在記錄與等待之間，所以實際停留是「載入時間」與「N 個 tick」取大的那個。
  在 DOSBox-X 的固定 3000 cycles 下，載入一張全螢幕圖要花掉可觀的時間。
- 雷聲長度不是 tick 數（見上），remake 取的是擷取量測的平均值。
- 有兩格的程式碼沒有等待，原版在那裡停留多久完全取決於載入速度。remake 給了一個
  最小停留（1 tick），畫面才不會一閃而過。

換句話說 remake 重現的是**設計時序**，原版觀察到的較長時間是機器速度造成的。
要對拍實際秒數必須先固定模擬器速度並量出載入時間，那是另一件事。

## 尚未關閉

- 雷聲的長度：`sub_10F75` 的步進間隔是以開機期校正值算出的忙等，不是 tick 數，
  靜態讀不出實際秒數。remake 目前只標記「這一格之後放雷聲」，沒有播放。
- 開機期的校正常式 `sub_118D1` 與 `sub_11AAC` 的作用尚未讀出。
- 畫格時間戳是主機牆鐘、取樣間隔 0.3 秒，只能核對段落順序與量級。要做逐幀時序對拍，
  必須換成能鎖定畫面更新時點的擷取方式。
- remake 目前只重播畫面與 tick 等待，還沒有實際播放的時間迴圈與音效。
