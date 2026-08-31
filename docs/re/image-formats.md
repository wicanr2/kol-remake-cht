# 圖片容器與壓縮格式

本頁只記錄固定原版 build 的格式證據。它不是 runtime 規格，也不授權 remake 讀取原版
archive。完整反組譯與原始 bytes 僅存本機 ignored 證據；版控內只保存位址、雜湊與統計。

## 固定輸入與工具

- ZIP：`Knights-of-legend.zip`，SHA-256 `2b455e3c…9f46`。
- 程式：`KOLTOWN-unpacked.EXE`，SHA-256 `01f6f9a9…d1d2`。
- IDA Pro 9.4，`ida-pro-9.4-idapython:locked-v1`。
- 位址空間：重建後 MZ 的 IDA linear effective address。
- 可重生摘要：`image-asset-inventory.json`、`image-lzw-probe-summary.json`、
  `combpics-visual-oracle.json`、`ida-fingerprints/KOL*-video-io-summary.json`、
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

### COMBPICS 的二階像素流

`sub_288D4`（`0x288d4..0x289ee`）提供一條完整 consumer：

1. 以 `comb%s` 選檔並呼叫 `sub_1A4FE` 做索引選取與 LZW 解碼；
2. 固定矩形 byte-column `x=[1,38)`、scanline `y=[8,152)`；
3. mode 大於 8 時讀 LZW 輸出開頭的 big-endian 16-bit 分界，呼叫 `sub_1C907`；
4. `sub_1C907` 對 mode 9 呼叫 `sub_1C80F`，其他模式呼叫 `sub_1A2BB`；
5. 完成後以 `sub_19E9D` 把轉換結果 blit 到上述矩形。

`sub_1A2BB`（`0x1a2bb..0x1a428`）把第一段 byte 當控制流、第二段當 high-nibble-first
的 4-bit 色彩流。控制碼 `<0x80` 重複一個色彩 `control+1` 次；`>=0x80` 複製
`control-0x7f` 個 literal。每八個色彩索引拆成四個 EGA bitplane bytes。前五個
`COMBPICS.0` 項目均精確展開為 `42,624 = 37×8×144 = 296×144` pixels，與 consumer
矩形獨立吻合。此分支的 LZW、二階 RLE、4-bit 索引、方向與尺寸為 `已證實`；其餘
archive 仍維持前述 `強推論`。

### 顯示模式、suffix 與色盤

- KOLTOWN `0x10032..0x10053` 讀 BIOS Data Area `0040:0049`；mode 4 把 suffix index
  加一，mode 7 設為二。DGROUP offset `0x0670` 的固定表為 `pics\0pcga\0herc\0`。
- KOL `_main`（KOL loaded-MZ EA `0x100d1..0x1011d`）先以 BIOS `INT 10h/AH=12h`
  查 EGA；成功時選 `AX=0x000d`，再於 `0x10219` 呼叫 `INT 10h` 設定模式。mode 0Dh
  是 320×200、40 byte-columns、4-plane EGA；KOLTOWN 的 `*PICS` 即此一般分支。
- KOL／TITLE／KOLTOWN 的 IDA-defined code I/O 清冊沒有 palette register 寫入；TITLE
  只重設 video mode、操作 speaker／timer 及 EGA plane，KOLTOWN 的直接 port I/O 均是
  EGA sequencer／graphics controller plane、mask 與 blit 操作。`MAKEMODE` 清冊沒有
  video I/O。因此 mode 0Dh palette 由 BIOS mode setup 繼承。
- 外部平台契約採 [DOSBox-X commit `63a4357`](https://github.com/joncampbell123/dosbox-x/commit/63a435738b537d720c2db2ed0944fde3d88009c9)
  的 [`int10_modes.cpp`](https://github.com/joncampbell123/dosbox-x/blob/63a435738b537d720c2db2ed0944fde3d88009c9/src/ints/int10_modes.cpp)（SHA-256
  `42044a6d…4060`）：mode table 把 `0x00d` 定義為 320×200 `M_EGA`，`ega_palette`
  使用 6-bit RGB `00/15/2a/3f`，換算為 PNG 的 `00/55/aa/ff`。這是
  `hardware-spec approximation`，不冒稱原始類比螢幕逐波形色彩。

測試 PNG `COMBPICS[0]`／`[1]` 分別對上[魔旅士文章](https://morrisjfwong.com/blog/knights-of-legend-%E6%AD%A6%E5%A3%AB%E5%82%B3%E8%AA%AA-dos-game/)
中的 `kol-flee-001a-s.jpg` 與
`war-lose-001a-s.jpg`。文章 288×189 JPEG 的遊戲區以固定 crop `(7,16)-(273,145)`
對應 296×144 圖；BICUBIC RGB MAE 正配為 `20.12`／`19.17`，交叉錯配為
`62.00`／`62.88`，正配均小於錯配一半。這項外部 oracle 證實場景身分、幾何、方向與
色彩索引排列；JPEG 壓縮與縮放不能證明 byte-exact RGB。下載圖片只在 ignored 工作區，
Git 只保存 URL、SHA-256 與量化摘要 `combpics-visual-oracle.json`。

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

1. `PICS`／mode 0Dh 的 COMBPICS 垂直鏈已關閉；PCGA／HERC 與其他 archive 尚未關閉。
2. 其餘圖號到寬、高、透明規則與裁切矩形的完整映射。
3. `AWD1PICS.0` 尾端 520 bytes 的 consumer。
4. 不符合 archive 形狀之無副檔名／`.1`／`.2` 檔案的 writer 與生命週期。
5. COMBPICS 已有兩張外部原版截圖交叉驗證；仍待本機 runtime framebuffer 的無損抽樣。

以上問題關閉前，`SPEC-002` 保持 DRAFT；一次性解碼成功不得接入 Go／Ebitengine。
