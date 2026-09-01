# 刪除角色的逐 byte 對拍

更新日期：2026-09-01

建立角色的屬性是擲骰出來的，原版與 remake 不可能寫出同一份檔案，因此那邊的判準是
「原版改動的每個 byte 都要寫對」。刪除沒有這個問題：同一份 `CHARDATA`、刪同一個
槽位，兩邊應該產出**完全相同的檔案**。這裡的判準就是全檔相等。

## 原版怎麼刪

原版 DELETE CHARACTER 只寫一個 byte：把該槽位名字的第一個 byte 寫成 0。

兩輪獨立執行（`../character-ui/delete-character-confirm-run{1,2}`）都刪掉名冊第 2 名
`Poontang`，12288 bytes 的 `CHARDATA` 只有 slot 1 的 offset `0x000` 由 `0x50` 變成
`0x00`；所在地、屬性、金幣、武器欄位都沒有動，`CHARPICS` 與 `SHLDPICS` 的 SHA-256
也完全沒變。畫面上名冊會往上補齊，那是「只列出已建立的槽位」造成的，不是資料搬移。

## 產生方式

```
tools/run_character_ui_oracle.sh <目錄> --scenario delete-character-confirm \
    --save-character-files
tools/run_kol_headless.sh -chardata-delete-slot 1 \
    -chardata-template <原版 CHARDATA> -chardata-out <remake 產出>
python3 tools/verify_chardata_delete.py --before <原版 CHARDATA> \
    --original <DOS 執行後> --remake <remake 產出> --slot 1 \
    --out docs/verification/chardata-delete/delete.json
```

`CHARDATA` 含原版 bytes，只留在本機；入庫的是收據裡的位移與雜湊。

## 目前結果

| 收據 | 原版來源 | 原版改動 | remake 改動 | 全檔相同 |
|---|---|---:|---:|---|
| `delete.json` | `delete-character-confirm-run1` | 1 | 1 | 是 |
| `delete-run2.json` | `delete-character-confirm-run2` | 1 | 1 | 是 |

`internal/chardata` 的單元測試另外檢查刪除確實只動那一個 byte：對一份沒有任何 0 的
樣板刪一個槽位，全檔只有 `slot × 768 + 0x00` 變成 0。
