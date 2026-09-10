# `region1.dat` 的來歷：地區擴充片（2026-09-11）

證據等級 `已證實`（執行時觀測＋原版位元組對位＋原版自己的訊息文字）。

## 結論

`region%d.dat` 是**擴充地區**的檔名。原版支援最多五個地區（`region0`…`region4`），
`REGION0.DAT` 是隨遊戲附的第一個；其餘四個由另外發行的「地區磁片」帶來，玩家用
遊戲內建的 **REGION INSTALLER** 裝進來，之後在遊戲裡選要航向哪一個地區。

Origin 沒有發行過擴充地區，所以 `region1.dat`…`region4.dat` 從來就不存在。書本
啟動時依序試開這五個檔，是在**清點裝了幾個地區**——開不到是預期結果，原版對這個
情況有自己的訊息：`ONLY ONE REGION IS INSTALLED.`

缺這四個檔不是素材不全，也不是執行器的問題，不需要補造。

## 證據

輸入：`KOLBOOK.EXE` SHA-256 `2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832`、
`REGION0.DAT` SHA-256 `21d1291a41792688268af02bbeabcf56a5402c614d3c0d86da2adbe55a9e6ca4`。
執行器 dosgolem，載入 `KOL.EXE`，正常啟動鏈進書本。重跑：

```
KOL_CHECK=book KOL_BOOK_SELECT_DOWN=2 KOL_BOOK_REGION_PROBE=1 \
  tools/oracle/check_updated_dosgolem.sh
```

### 誰在開、什麼時候開

開檔序列出現在**書本階段**（`KOLBOOK.EXE` 已載入、`KOLTOWN.EXE` 尚未載入）：
先開 `REGION0.DAT` 成功，再依序試 `region1.dat`…`region4.dat`，四次全部找不到。
五次都從同一個呼叫點 `0110:9157` 發出（該處是 `KOLBOOK.EXE` 的 CRT `open()`）。

### 格式字串的來源

執行時掃描找到三處 `region%d.dat` 模板，位址與 `KOLBOOK.EXE` 的靜態偏移逐項對上
（載入基底段 `0x0110`、MZ header `0x200`，換算式 `檔案偏移 = linear - 0x1100 + 0x200`）：

| 執行時 | 換算 | `KOLBOOK.EXE` 檔案偏移 |
|---|---|---:|
| `0000:a902` | `0xa902 - 0x1100 + 0x200` | `0x09a02` |
| `0000:a948` | `0xa948 - 0x1100 + 0x200` | `0x09a48` |
| `0000:aa30` | `0xaa30 - 0x1100 + 0x200` | `0x09b30` |

三處都落在一張檔名模板表裡，鄰居是 `bookdata`、`book%s.0`、`disk%c3`、
`awd%dpics.0`、`awd%dpcga.0`、`awd%dherc.0`、`chardata`、`charpics`、`shldpics`。
`KOLTOWN.EXE` 與 `KOL.EXE` 沒有這個模板。

組好的檔名字串在 `0000:b714`，正是缺檔記錄的 `DS:DX`（`0a34:13d4` → linear `0xb714`）。

### 原版自己的說法

同一時刻記憶體裡的訊息表（`@` 是換行、`.` 是字串結尾、`~V`／`~H` 是排版控制碼）：

```
REGION INSTALLER

INSERT DISK 3 OF THE NEW REGION AND PRESS ANY KEY OR PRESS ESC TO ABORT.

NOW INSERT A BLANK FORMATED DISK IF YOU WISH TO MAKE A NEW CHARACTER DISK
OR A CHARACTER DISK WHICH WILL BE UPDATED FOR THE NEW REGION.  PRESS ANY KEY.

ONLY ONE REGION IS INSTALLED.  PRESS ANY KEY.

SELECT THE REGION YOU WISH TO SAIL TO...
```

原版的 `INSTALL.BAT`／`INSUB.BAT` 只安裝 A1–A5 五片遊戲本體，沒有任何 region
步驟——擴充地區走的是遊戲內建的安裝器，不是磁片安裝腳本。

### `REGION0.DAT` 的內容

48 byte，長度前綴字串的地名表：

```
0000 0e "TROLLSBANEBRT" 0002 0000 2b32 0e26 2a62 003f 3200 3f38
0a "ASHTALAREA" 00...
```

## 觀測方法上的一個坑

**讀原版記憶體要在事件當下讀，不能事後補讀。** 這次的四次缺檔發生在
`KOLBOOK.EXE` 階段，但測試結束時 `KOLTOWN.EXE` 已經取代它佔住同一個段
（`0x0110`）。兩支是同一套 compiler／library 連出來的，CRT 落在**相同的段內
偏移**，所以事後把記憶體 dump 出來拿去比對原版執行檔，會精確命中「錯的那一支」，
而且比對本身完全成立——32 byte 逐位元組相同，看不出任何異狀。

`FileAccess` 記的 `CS:IP`／`DS:DX` 是當下值，可信；用那組值在事後去讀記憶體，
讀到的是後來者的內容。同理，BP 鏈回溯在 CRT 的 `open()` 會斷（那層沒建 frame），
拿到的 caller 只能當線索。

## 不宣稱

- 上限五個地區來自實測（試到 `region4` 就停）與 `ONLY ONE REGION IS INSTALLED`
  的存在，沒有反組譯出迴圈上界常數。
- 沒有查明那張訊息表存在哪個原版檔：原版檔案裡沒有明文，也不是單一 XOR 或
  加常數，是壓縮存放的，載體與格式未解。
- 擴充地區的資料格式未知——沒有實物可查。
