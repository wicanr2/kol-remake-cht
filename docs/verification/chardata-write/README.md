# 角色寫入的逐 byte 對拍

更新日期：2026-09-01

原版建立角色時只改動槽位裡的一部分 bytes。這裡的判準不是「檔案雜湊相同」——屬性是
擲骰出來的，兩邊不可能一樣——而是：**給定同一個角色，原版改動的每一個 byte，remake 的
writer 都要寫出相同的值**，而且不能寫錯、也不能多寫。

## 產生方式

```
tools/run_character_ui_oracle.sh <目錄> --scenario accept-and-save --save-character-files
tools/run_kol_headless.sh -chardata-write workplace/chardata-write.json \
    -chardata-template <原版 CHARDATA> -chardata-slot 14 \
    -chardata-name zed -chardata-stats "136,91,104,115,101,94,99"
python3 tools/verify_chardata_write.py --before <寫入前> --after <寫入後> \
    --record workplace/chardata-write.json --out docs/verification/chardata-write/barbarian.json
```

`CHARDATA` 含原版樣板 bytes，只留在本機；入庫的是收據裡的位移與統計。

## 目前結果

| 收據 | 職業 | 原版改動 | 逐 byte 相同 | 寫錯 | 多寫 |
|---|---|---:|---:|---:|---:|
| `barbarian.json` | BARBARIAN（記錄 0） | 58 | 58 | 0 | 0 |
| `ranger.json` | RANGER（記錄 1） | 58 | 58 | 0 | 0 |

`internal/chardata` 的單元測試另外逐筆檢查 39 個職業都寫得出武器樣板，
沒有副武器的 21 筆不會誤寫副武器欄位。

兩筆是不同職業、不同起始武器、不同擲骰結果，因此不是同一組數字碰巧對上。

## 一個被推翻的推論

第一輪只有 BARBARIAN 一筆資料時，把寫進槽位的武器 bytes 拿去搜尋，在 `ITEM.LST` 裡
唯一命中記錄 7；GS(4)→7、HL(8)→11 兩點都成立，看起來就是「`ITEM.LST` 索引 ＝
selector ＋ 3」。跑第二個職業才發現 BR(9)→5、LB(3)→25 不符。

真正的來源不是 `ITEM.LST`：武器樣板來自 `BOOKDATA` 的第 3 個項目（14 筆 × 16 bytes），
記錄索引 ＝ selector − 1。同樣的武器記錄同時出現在兩個檔案裡，所以位元組搜尋會命中一個
**正確但無關**的位置。**位元組相同不等於來源相同**——確定來源只能讀定址的那段程式碼
（`sub_11820` 的 `[bx+79EEh]`）。規則讀出來之後，39 筆職業都能寫，其中 21 筆沒有副武器，
原版對那些職業不動 `0x57` 與 `0x5f`。

## 已修正的欄位語意

`KOLBOOK-character-writer-summary.json` 把位移 `0x2b` 記成「starting gold crowns
little-endian」。實際寫入的 bytes 是 `05 DC`：職業表裡是 little-endian 的 `dc 05`
（＝1500），寫進槽位時變成 big-endian 的 `05 DC`。**記錄裡的金冠是 big-endian**，
與原版讀圖片張數時的 `xchg al, ah` 一致。

## 建立 → 寫入 → 清單 的閉環

`roster-after-write.png` 是把 remake 用腳本建立的角色（`kai`，BARBARIAN）寫進槽位 14
之後重新讀出的角色清單：原本 14 列變成 15 列，新角色的所在地是 `TROLLSBANE BRT`。

## 這份收據不宣稱什麼

- 不宣稱整份 `CHARDATA` 逐 byte 相同。屬性是擲骰的，兩邊必然不同。
- 不宣稱 remake 有正式存檔格式。這個 writer 是為了對拍而寫的原版格式輸出；
  remake 自己的存檔 schema 仍是共同決策項，尚未決定（SPEC-005）。
- 39 個職業都能寫出武器樣板（`internal/chardata` 的單元測試逐筆檢查），但只有
  BARBARIAN 與 RANGER 有原版實際寫入可以對照，其餘 37 筆沒有原版側的對照組。
- 不宣稱槽位其餘欄位的語意已知。REGION 中繼資料與武器樣板內部欄位仍未反追。
