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

清冊找到 44 個實體圖片候選，歸併為 30 個邏輯 archive，共 1,022 個索引項。所有 payload
都位於所屬分卷範圍內；單卷 archive 中 29 個恰好抵達檔尾。`AWD1PICS.0`
在索引 payload 後另有 520 bytes。KOLBOOK `_main` 以 `sub_137C0` 將 PICS、PCGA、HERC
三個工作區 buffer 各固定寫成 `0x0fa0 = 4000` bytes；KOLTOWN `sub_168F8` 則以
`sub_1A4FE` 只依 PICS 索引載入指定 payload。故這 520 bytes 已排除於 KOLTOWN 的索引
圖片流之外，分類為「固定大小編輯器工作區快照的未解析尾區」，等級為 `已證實`；它不是
可丟棄 padding，抽取器不得解讀或複製為 PNG 像素。

無副檔名 `CHARPICS`、`SHLDPICS`、`STEEDPIC` 均為 `0x2f80 = 12160` bytes。KOLBOOK
`sub_11820`／`sub_1293E` 固定長度寫回前兩檔，`sub_135DB` 整檔讀入；KOLTOWN
`sub_14DB1` 以角色槽位 `0x300` bytes 推導圖片槽位 `0x2f8 = 760` bytes，複製
`0x17c` words，並將三檔各固定寫回 `0x2f80` bytes。清冊因此把三檔分類為 16 筆、每筆
760 bytes 的可變角色圖片記錄表，並保存逐筆 SHA-256；記錄內的像素形狀仍是 `未知`。
`STEEDPIC` 是這條儲存流程的單數輸出；另有 `steedpics` 由 `sub_14779` 以最大
`0x7530` bytes 載入，兩者不得混為同一格式。

### TWN 多分卷契約

`sub_1A4FE`（`0x1a6b4..0x1a85d`）證實 `.0` 的前 20 bytes 同時管理最多五個分卷：

- offsets `0,2,4,6,8` 是各非末卷的累計圖數門檻；零值表示目前已到末卷；
- offsets `10,12,14,16,18` 是各卷最後一個 payload 的長度；
- 全域圖號的 16-bit offset table 從 byte 20 開始，offset 是相對於被選中的實體分卷；
- loader 比較 `image_index+1` 與累計門檻，選出 `.0` 至 `.4`，卷界的最後一項使用對應
  last-length；末卷以 table 中等於檔案大小的 sentinel 界定最終圖數。

固定輸入實際有七組兩卷與一組三卷：TWN0..6 分別為 25、34、30、26、26、30、26 張；
TWN7 為 46 張。所有卷內 offset 嚴格遞增、第一卷從 byte 200 起、續卷從 byte 0 起，且
最後 payload 精確抵達各實體檔尾。這使舊清冊漏算的 62 張 continuation payload 從 960
訂正為 1,022；v2 清冊逐張記錄 volume、offset、長度與雜湊。

## 壓縮流

`sub_1A04E`（`0x1a04e..0x1a26a`）與 bit reader `sub_1A26B`
（`0x1a26b..0x1a2ba`）共同顯示：

- 初始碼寬 9 bits，literal dictionary 為 `0x00..0xff`，下一個字典碼為 `0x101`；
- `0x100` 保留作碼寬增加標記，遇到後碼寬加一並立即再讀一碼；
- bit reader 以最低有效位元優先（LSB-first）組碼；
- 字典項保存前綴碼與尾 byte，並處理 `code == next_code` 的 LZW 特例；
- 壓縮流長度由上述索引契約提供，沒有在此證據中發現獨立 end code。

結論等級為 `強推論`：靜態控制流與資料流一致，而且一次性探針已對 30 個 archive、1,022 個
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

### TWN 全域圖號 24 的 continuation 垂直鏈

`sub_11302`（`0x11302..0x11390`）以 `twn%d%s` 建檔名後把呼叫端圖號交給
`sub_10F9C`。`sub_10F9C` 經 `sub_1A4FE` 選卷並 LZW 解碼，mode 大於 8 時再走
`sub_1C907`／`sub_1A2BB`，最後由 `sub_19E9D` blit。caller `sub_2AAA8` 在
`0x2aab1..0x2aab5` 明確傳入全域圖號 24。

圖號 24 在八組 TWN archive 全部位於 `.1`。對 town 0，`sub_11302` 設 byte-column
`x=[13,37)`、scanline `y=[58,154)`，即 192×96；town 1..7 設 `x=[1,12)`、
`y=[64,128)`，即 88×64。一次性探針對八張圖均精確展開至 consumer 像素數，索引落在
0..15，重跑 PNG 逐 byte 相同。接觸表目視為一張道路／提示畫面與七張正向人物肖像，
沒有旋轉、鏡像、空白或明顯 scanline 錯位。其 LZW、分卷選擇、二階 RLE、4-bit 索引、
尺寸與方向為 `已證實`；RGB 仍沿用 mode 0Dh 的 `hardware-spec approximation`。
安全統計見 `town-index24-png-probe-summary.json`，PNG／接觸表只留 ignored 工作區。

### PCGA／HERC 是 MAKEMODE 衍生素材

固定 ZIP 中除 `AWD1PCGA.0`／`AWD1HERC.0` 工作區快照外，沒有一般 `*PCGA*` 或
`*HERC*` 圖片檔；但它包含官方工具 `MAKEMODE.EXE`（SHA-256 `3051cc98…6c56`）。
MAKEMODE `_main` 的提示 consumer 明確提供 CGA 與 Hercules 兩種轉換選項；選項 1 把
輸出 suffix 設為 `pcga`，選項 2 設為 `herc`。

`_main` 在 loaded-MZ EA `0x1019d..0x103c4` 對 28 個 archive 基名逐一呼叫
`sub_10464`，每筆同時傳入全域圖數與 transform case。28 個圖數全部與 v2 邏輯清冊相符，
包括 TWN 分卷的 25／34／30／26／26／30／26／46；不在表內的是 AWD 編輯器工作區與
動態角色 `CHARPICS`。安全 catalog 保存在
`ida-fingerprints/MAKEMODE-display-selection-summary.json`。

`sub_10464` 讀 `basepics.0` 的 header／全域 offsets，並以 `base+suffix+volume` 寫出新
archive。靜態資料流串接 `sub_114F4` 的 LZW 解碼、依 transform case 的矩形處理、
`sub_11734` 的二階 RLE／四 EGA planes 展開、`sub_118A8` 的 4-plane 至 packed 2-bit
映射、`sub_11242` 的重壓縮，以及 `sub_1142C` 的 checked DOS write。官方「由 PICS
生成 PCGA／HERC」的來源與輸出契約為 `已證實`；完整 transform-case 幾何、CGA palette、
Hercules 單色／抖色規則及生成檔的動態 round-trip 尚未驗證，整體像素轉換仍是
`強推論`，不得直接寫成 production converter。

這形成一個產品範圍決策：remake 可只以固定安裝的 EGA PICS 作必要基準，也可把官方
生成的 CGA／Hercules 當額外原版 theme。此決定會改變抽取、驗證與發行範圍，尚未由
使用者確認，因此 SPEC-002 不代選。

## Loader 與 consumer

- `sub_19D98`／`sub_19E12`：KOLTOWN 的整檔或指定長度讀寫 helper，語意為 `強推論`。
- KOLBOOK `sub_1374E`（`0x1374e..0x137c0`）以 `__dos_open`、最大 `0xffff` bytes 的
  `__dos_read`、`__dos_close` 做整檔讀取；`sub_137C0`（`0x137c0..0x13843`）以
  `__dos_creat`、呼叫端指定長度的 `__dos_write`、短寫檢查及 `__dos_close` 寫檔，均為
  `已證實`。完整指令只留本機 private JSON，版控摘要保留原始定位與雜湊。
- `sub_1A4FE`：依圖號與顯示模式選檔、選取 payload 並呼叫解碼器，`已證實`。
- `sub_19E9D`：依顯示模式做矩形 copy／blit，`強推論`。
- `sub_1BD17`：含 clipping 與 mask 的顯示模式分支 blitter，`強推論`。
- 檔名 consumer 已定位 `charpics`、`shldpics`、`steedpic`、`twn%d%s.0/.1`、
  `kolm%s.0`、`misc%s.0`、`mon%s`、`comb%s`、`per%s`；近位址 raw word 命中可能巧合，
  只有結合函式與參數資料流者才可當 consumer。

## 未關閉問題

1. `PICS`／mode 0Dh 的 COMBPICS 與 TWN 圖號 24 垂直鏈已關閉；PCGA／HERC 與其他
   archive 圖號尚未關閉。
2. 其餘圖號到寬、高、透明規則與裁切矩形的完整映射。
3. 三個固定角色圖片記錄表內部的寬、高、位平面與透明規則。
4. 使用者尚未決定 PCGA／HERC 是必要交付 theme、可選後續 theme，或不納入產品範圍。
5. COMBPICS 已有兩張外部原版截圖交叉驗證；仍待本機 runtime framebuffer 的無損抽樣。

在完整命名、manifest 與失敗政策關閉前，`SPEC-002` 保持 DRAFT；一次性解碼成功不得
接入 Go／Ebitengine。
