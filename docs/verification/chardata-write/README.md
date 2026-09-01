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

兩筆是不同職業、不同起始武器、不同擲骰結果，因此不是同一組數字碰巧對上。

## 一個被推翻的推論

第一輪只有 BARBARIAN 一筆資料時，武器樣板的來源看起來是「`ITEM.LST` 索引 ＝ 武器
selector ＋ 3」：GS(4)→7、HL(8)→11 兩點都成立。跑第二個職業才發現 BR(9)→5、LB(3)→25，
關係根本不是算術的，EXE 內也找不到對應的位元組表。**通則至今未知**，因此
`item-templates.json` 只登記實際觀測過的職業，writer 對沒觀測過的職業直接失敗——
猜一個索引會寫出「自洽但錯」的存檔。

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
- 不宣稱 39 個職業都能寫。只有 BARBARIAN 與 RANGER 的起始武器物品索引被觀測過。
- 不宣稱槽位其餘欄位的語意已知。REGION 中繼資料與武器樣板內部欄位仍未反追。
