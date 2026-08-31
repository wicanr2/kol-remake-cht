# 位址空間與 KOLTOWN 搬移映射

## 位址空間

| 名稱 | 基準 | 用途 |
|---|---|---|
| packed file offset | 原始 `KOLTOWN.EXE` 檔頭 byte 0 | MZ header、packed stream、stub、relocation table |
| packed IDA EA | IDA load base `0x10000` | 原始 packed EXE 的 loader／stub 分析 |
| runtime relocated EA | packed IDA EA 加 stage 1 搬移差額 | 描述 stub 在 DOS 記憶體搬移後的執行位置 |
| unpacked file offset | 衍生 EXE 檔頭 byte 0 | CC0 unexepack 重建輸出 |
| unpacked IDA EA | 衍生 EXE 的 IDA load base `0x10000` | 真正程式 entry、函式、xref 與後續 RE |

後續證據必須寫出空間名稱，不能只寫 `0x2b90c`。

## Stage 1 搬移

輸入 SHA-256：`c2442c…b221`。IDA Pro 9.4 packed EA：

| 欄位 | 值 | 等級 |
|---|---|---|
| MZ entry | `0x2f522` | 已證實 |
| stub base | `0x2f510` | 已證實 |
| copy source | `[0x2f510, 0x3082f)` | 已證實 |
| copy destination | `[0x2f920, 0x30c3f)` | 已證實 |
| copy bytes | `0x131f`（4,895） | 已證實 |
| mapping delta | `+0x410` | 已證實 |
| stage 2 source representation | `0x2f548` | 已證實 |
| stage 2 runtime location | `0x2f958` | 已證實 |

`STD; REP MOVSB` 表示由區塊尾端向前迭代，以安全處理互相重疊且目的位址較高的搬移；
它不是「把整個區塊搬到較低位址」。

## Stage 2 解壓後入口

- 程式相對入口：`CS:IP=1b8f:001c`。
- 在 unpacked derivative、IDA load base `0x10000` 下：linear EA `0x2b90c`。
- Stack：`SS:SP=2b48:07d0`。
- packed IDA `0x2f548..0x2f627` 是解壓 stub 的來源表示，不是解壓後產品函式地址。
- 解壓後所有函式與資料語意只能引用 unpacked derivative SHA-256
  `01f6f9…d1d2` 的 unpacked IDA EA；不得把 packed 與 unpacked EA 並列成同一位址。
