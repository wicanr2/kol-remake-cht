# 圖片容器與壓縮格式

本頁只記錄固定原版 build 的格式證據。它不是 runtime 規格，也不授權 remake 讀取原版
archive。完整反組譯與原始 bytes 僅存本機 ignored 證據；版控內只保存位址、雜湊與統計。

## 固定輸入與工具

- ZIP：`Knights-of-legend.zip`，SHA-256 `2b455e3c…9f46`。
- 程式：`KOLTOWN-unpacked.EXE`，SHA-256 `01f6f9a9…d1d2`。
- IDA Pro 9.4，`ida-pro-9.4-idapython:locked-v1`。
- 位址空間：重建後 MZ 的 IDA linear effective address。
- 可重生摘要：`image-asset-inventory.json`、`image-lzw-probe-summary.json`、
  `ida-fingerprints/KOLTOWN-image-*-summary.json`。

## 容器索引契約

`sub_1A429`（`0x1a429..0x1a4e3`）與 `sub_1A4FE`（`0x1a4fe..0x1a866`）提供下列證據：

| 原始定位 | 附加語意 | 等級 | 證據 |
|---|---|---|---|
| `sub_1A4E4`，`0x1a4e4..0x1a4fd` | 從指定 segment／offset 讀 little-endian 16-bit word | 已證實 | 函式只做 `ES:[BX]` word read |
| `0x1a432..0x1a443` | 圖號轉成 `20 + index * 2` | 已證實 | 原始算術與後續 word reader consumer |
| `0x1a446..0x1a469` | byte 0 是項目數；索引 word 是 payload offset | 已證實 | count 比較與 offset 傳入解碼器 |
| `0x1a49a..0x1a4c4` | 最末項長度來自 offset 10 的 word；其餘為相鄰 offset 差 | 已證實 | 兩條長度分支均流入同一解碼器 |

清冊找到 42 個圖片候選，其中 30 個通過這個基本容器形狀，共 960 個索引項。所有 30
個容器的 offset 嚴格遞增且 payload 位於檔案範圍內；29 個恰好抵達檔尾。`AWD1PICS.0`
在索引 payload 後另有 520 bytes，內容用途為 `未知`，不得先稱為 padding 或 palette。

無副檔名 `CHARPICS`、`SHLDPICS`、`STEEDPIC`，以及部分 `TWN*PICS.1/.2` 不符合上述
容器形狀；consumer 證實前三者會被整檔讀入，但其像素佈局與是否為 cache／存檔仍未知。

## 壓縮流

`sub_1A04E`（`0x1a04e..0x1a26a`）與 bit reader `sub_1A26B`
（`0x1a26b..0x1a2ba`）共同顯示：

- 初始碼寬 9 bits，literal dictionary 為 `0x00..0xff`，下一個字典碼為 `0x101`；
- `0x100` 保留作碼寬增加標記，遇到後碼寬加一並立即再讀一碼；
- bit reader 以最低有效位元優先（LSB-first）組碼；
- 字典項保存前綴碼與尾 byte，並處理 `code == next_code` 的 LZW 特例；
- 壓縮流長度由上述索引契約提供，沒有在此證據中發現獨立 end code。

結論等級為 `強推論`：靜態控制流與資料流一致，而且一次性探針已對 30 個容器、960 個
payload 全部完成機械解碼，沒有越界或無效字典碼。但尚未證實解碼輸出的尺寸、位平面、
palette、透明色與畫面方向，因此不能把「可解出 bytes」升格成「圖片已正確抽取」。

## Loader 與 consumer

- `sub_19D98`／`sub_19E12`：整檔或指定長度讀取 helper，語意為 `強推論`；DOS runtime
  wrapper 的精確 library 名稱尚未定名。
- `sub_1A4FE`：依圖號與顯示模式選檔、選取 payload 並呼叫解碼器，`已證實`。
- `sub_19E9D`：依顯示模式做矩形 copy／blit，`強推論`。
- `sub_1BD17`：含 clipping 與 mask 的顯示模式分支 blitter，`強推論`。
- 檔名 consumer 已定位 `charpics`、`shldpics`、`steedpic`、`twn%d%s.0/.1`、
  `kolm%s.0`、`misc%s.0`、`mon%s`、`comb%s`、`per%s`；近位址 raw word 命中可能巧合，
  只有結合函式與參數資料流者才可當 consumer。

## 未關閉問題

1. 各顯示 suffix 對應的顯示模式、像素位平面與 palette。
2. 圖號到寬、高、透明規則與裁切矩形的完整映射。
3. `AWD1PICS.0` 尾端 520 bytes 的 consumer。
4. 不符合 archive 形狀之無副檔名／`.1`／`.2` 檔案的 writer 與生命週期。
5. 至少一張真實圖片的原版同狀態截圖、尺寸與像素抽樣比對。

以上問題關閉前，`SPEC-002` 保持 DRAFT；一次性解碼成功不得接入 Go／Ebitengine。
