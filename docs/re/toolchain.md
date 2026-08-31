# KOLTOWN 工具鏈與 Packer 指紋

## Microsoft EXEPACK

輸入：原始 `KOLTOWN.EXE` SHA-256 `c2442c…b221`。

| 證據 | 原始定位 | 結論 |
|---|---|---|
| MZ relocation count 為 0、非零高段 entry | MZ header | packer 候選；單獨不足以命名 |
| `Packed file is corrupt` | packed IDA stub base `+0x117` | Microsoft EXEPACK marker |
| `0xB0/0xB1` 填值、`0xB2/0xB3` 複製，low bit 結束 | packed IDA `0x2f58c..0x2f5ac` | EXEPACK backward decompression contract |
| 16 段 relocation table，逐 offset 加 load segment | packed IDA `0x2f5ae..0x2f5ec` | EXEPACK relocation reconstruction |
| 重設 `SS:SP` 後 indirect far jump | packed IDA `0x2f5ee..0x2f610` | 交接到解壓後程式 |
| CC0 `unexepack` 成功產生 2,281 relocations，IDA 可辨識 337 函式 | 衍生映像 `01f6f9…d1d2` | 獨立 consumer 驗證 |

結論：`KOLTOWN.EXE` 使用 Microsoft EXEPACK，`已證實`。精確 EXEPACK／linker／compiler
版本仍是 `未知`；不能由 marker 或相容解包器推定版本號。

## Compiler／runtime

- `KOL.EXE`、`KOLBOOK.EXE`、`MAKEMODE.EXE`、`TITLE.EXE`：IDA FLIRT 提出 Microsoft
  Quick C 1.0／2.01 與 MSC 5.1 DOS runtime／graphics 候選，等級為 `強推論`。
- `KOLTOWN.EXE`：packed 原檔無法直接套用函式簽章。後續應在 unpacked derivative 對
  startup ABI、runtime copyright、至少三種 codegen idiom 與 library callers 交叉確認。
- `TANDYPIN.EXE`：工具鏈與功能仍 `未知`；檔名不能作證據。

## Remake 邊界

EXEPACK stub、DOS relocation、compiler startup 與 runtime helper 不移植到 Go remake。
它們只用來取得可追溯的原版產品程式。後續研究會從 337 個函式中先排除工具鏈，再按
玩家可見 consumer 排序；函式數量不是完成度。

