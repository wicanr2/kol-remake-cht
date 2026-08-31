# 圖片容器與壓縮格式

本頁只記錄固定原版 build 的格式證據。它不是 runtime 規格，也不授權 remake 讀取原版
archive。完整反組譯與原始 bytes 僅存本機 ignored 證據；版控內只保存位址、雜湊與統計。

## 固定輸入與工具

- ZIP：`Knights-of-legend.zip`，SHA-256 `2b455e3c…9f46`。
- 程式：`KOLTOWN-unpacked.EXE`，SHA-256 `01f6f9a9…d1d2`。
- IDA Pro 9.4，`ida-pro-9.4-idapython:locked-v1`。
- 位址空間：重建後 MZ 的 IDA linear effective address。
- 可重生摘要：`image-asset-inventory.json`、`image-lzw-probe-summary.json`、
  `combpics-visual-oracle.json`、`monpics-ega-visual-probe.json`、
  `ida-fingerprints/KOL*-video-io-summary.json`、
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
`control-0x7f` 個 literal。每八個色彩索引拆成四個 EGA bitplane bytes。

`sub_288D4` 的全部五個 caller 傳入圖號 3、2、0、4、1，完整覆蓋 0..4；五筆均精確展開
為 `42,624 = 37×8×144 = 296×144` pixels。另一 consumer `sub_2B454` 在
`0x2b45d..0x2b46f` 固定矩形 byte-column `[1,12)`、scanline `[64,128)`，兩個 caller
分別傳圖號 6 與 5；兩筆各精確展開為 `5,632 = 11×8×64 = 88×64` pixels。兩組 caller
集合完整覆蓋 archive 0..6。

七張接觸表目視為五張戰鬥場景與兩張肖像，均無鏡像、旋轉、plane 錯排或裁切。故
COMBPICS 全七筆的 LZW、二階 RLE、4-bit indices、尺寸、方向及 direct-blit 政策為
`已證實`；其中圖號 0、1 另有魔旅士文章同場景截圖交叉驗證。

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

### MONPICS 彩圖與遮罩配對

固定 KOLTOWN derivative 的 DGROUP offset `0x1370`（IDA linear EA `0x2f240`）保存
32 組兩 byte 記錄。`sub_22EFC` 在 `0x22f3b..0x22f73` 以
`2 * word_3B262` 查表；第一 byte 參與水平範圍，且 EGA blit 以 byte-column 計算，故是
8-pixel 寬度單位，第二 byte 是 scanline 高度。它在 `0x22fb3..0x22fbe` 又以同一
`2 * word_3B262` 當 `mon%s` 圖號呼叫 `sub_10F9C`。另一條玩家路徑 `sub_2A975`
在 `0x2a9ac..0x2aa14` 使用同一尺寸表，並把圖號範圍
`[2 * monster_index, 2 * monster_index + 1]` 交給 `sub_16FA6`。

實檔 `MONPICS.0` 有 64 個 payload，正好是 32 組。每組偶數 payload 經 LZW 與
`sub_1A2BB` 二階 RLE 後，精確展開為 `width_bytes * 8 * height` 個 4-bit EGA index；
奇數 payload 的 LZW 輸出前 `width_bytes * height` bytes 是逐 scanline 的 1-bit
遮罩，每 byte 由 MSB 到 LSB 對應八個像素。一次性探針將兩者合成 32 張 RGBA PNG，
全部尺寸、色彩範圍與遮罩長度通過；接觸表人工抽查可辨識 32 個完整怪物，無鏡像、反相、
位元錯位或彩圖／遮罩邊界偏移。

彩圖／遮罩配對、32 組尺寸、MSB-first 遮罩與可見方向為 `已證實`。奇數 payload 在有效
遮罩後各多一個 LZW 解碼 byte；它不參與 `width_bytes * height` 的玩家可見遮罩，現分類為
`強推論：壓縮流終止／padding artifact`，抽取器必須記錄但不得當像素。安全統計見
`monpics-ega-visual-probe.json`；32 張 PNG 與接觸表只留 ignored 工作區。

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

### PERPICS 六組人物彩圖／遮罩

`sub_18F3A` 在 loaded-MZ EA `0x18fdd` 以 DGROUP offset `0x17d8` 的 `per%s` 與目前
顯示 suffix 建立檔名，接著於 `0x18ff3` 把 `[portrait_index, portrait_index + 6]`
交給 `sub_16FA6`。後者在 `0x16fe3`、`0x1706b` 兩次呼叫共用 loader `sub_1A4FE`，最後
於 `0x17087` 呼叫遮罩 blitter `sub_1BD17`。舊清冊只憑 near-offset 把 `per%s` 歸到
`sub_18F3A`；本次已用參數資料流證實，而 `sub_1CBA9` 則排除為媒體／磁碟提示流程。

DGROUP offset `0x0168` 的六筆三-byte 記錄為 `(x0, y0, x1_exclusive)`，固定
`y1_exclusive=154`；mode 0Dh 下寬度為 `(x1-x0)*8`。六組尺寸依序是 72×103、64×111、
72×64、112×148、72×103、56×110。前六個 payload 的二階 RLE／4-bit EGA 展開像素數
全數精確相符；後六個 payload 各為對應 `width_bytes*height+1` bytes，前段以 MSB-first
展開為 1-bit 遮罩。六張 RGBA PNG 接觸表目視為完整直立人物，方向與透明邊界正確。

因此 EGA `PERPICS.0` 的六組尺寸、前半彩圖／後半遮罩配對、索引、方向與透明邊界為
`已證實`。每筆遮罩最後一 byte 均為零且不進入 consumer 矩形；其形成原因仍是
`強推論：壓縮流終止／padding artifact`，不得當像素輸出。

### KOLMPICS 圖號 28 的 EGA AND-mask

`sub_10000` 在 loaded-MZ EA `0x1037a` 以 DGROUP offset `0x165a` 的 `kolm%s.0`
建立檔名，並於 `0x1038e` 呼叫 `sub_19D98` 把整個 archive 讀入 `word_39820`。
`sub_14047` 在 `0x14150/0x14154` 傳入同一 buffer、以 `0x14158` 的常數 `0x1c` 選圖號
28，再於 `0x1415c` 呼叫索引解碼器 `sub_1A429`。來源矩形固定為 byte-column `[0,7)`、
scanline `[0,57)`，即 mode 0Dh 下 56×57。

圖號 28 的 LZW 輸出為 400 bytes；前 399 bytes 以 MSB-first 展開後正好填滿 56×57，
末端單一零 byte 不進入矩形。`sub_1BD17` 的 EGA 路徑在 `0x1bf4c..0x1bf53` 執行
`destination = (destination AND mask) OR color`，故 mask bit 1 表示保留目的背景（透明），
bit 0 表示先清除再寫入彩圖（不透明）。bit-map PNG 目視為右上切角與圓弧下緣，方向連續。

以上 archive 載入、圖號、幾何、MSB-first 位序、AND-mask 語意與方向為 `已證實`；尾端
零 byte 成因維持 `強推論：壓縮流終止／padding artifact`。這只關閉 KOLMPICS 圖號 28，
不得泛化到其他 82 個 payload，也不得把純 mask 冒充可獨立顯示的彩圖。

### MISCPICS 五張 EGA 彩圖與 placeholder

`sub_10000:0x1060d` 以 DGROUP `0x1678` 的 `misc%s.0` 建檔名，`0x10621` 呼叫
`sub_19D98` 把 archive 讀入 `word_2FBB8`。IDA 全域 data refs 只有啟動配置 `0x105ec`、
清理釋放 `0x10e62` 與唯一 consumer `sub_1103F:0x111c0`；不存在另一條直接讀取端。

`sub_1103F` 的 request-code 映射與矩形為：34 → image 0（192×96）、35 → image 1
（296×32）、71／72／73 → image 3／4／5（各 56×70）。`0x111c7` 以
`mapped_request_code-6` 算索引，證實集合是 `{0,1,3,4,5}`；image 2 永遠不會被選中。
`sub_11262:0x11288` 呼叫 `sub_1A429` 解碼，`0x112f4` 用 `sub_19E9D` 直接 blit。

五個被引用 payload 經 LZW、二階 RLE 後均精確得到矩形所需的 4-bit EGA pixels；接觸表
目視為一張場景、一張橫幅與三張人物／標記圖，無鏡像、旋轉或 plane 錯排。直接 blit
沒有獨立 mask，故索引 0 是圖像色彩，不得擅自轉透明。image 2 解碼後只有一 byte，且
全程沒有 consumer，可分類為 `已證實：未引用非圖片 placeholder`，manifest 必須用
`excluded_non_image` 且不得產生 PNG。

### TITLE 的 SET0／SET1／SET2／FACE 動畫素材

固定 `TITLE.EXE`（SHA-256 `8a2879d9…ad6`）的 `_main` 依序以 `set0%s.0`、
`set1%s.0`、`set2%s.0`、`face%s.0` 建檔名，並在 loaded-MZ EA `0x1021b`、
`0x103e0`、`0x10558`、`0x10723` 呼叫整檔 reader `sub_11058`。`sub_11666`
以 `20+index*2` 讀取 archive offset，`sub_1170E` 是 little-endian word reader，最後由
`sub_11133` 依 mode 0Dh 的 40-byte scanline、四個 EGA planes 做矩形 blit。

`sub_10AF0` 對 SET0／1／2 完整消費圖號 0..5，矩形依序是 88×70、104×84、
160×125；FACE 主流程的 18 個呼叫涵蓋圖號 0..5，矩形固定 112×80。四個 archive 共
24 筆經 LZW 與二階 RLE 後，全部精確得到對應矩形的 4-bit 色彩索引，範圍均為 0..15。
代表性 ignored PNG 目視為三組逐步擴大的山谷騎士場景及正向臉部影格，沒有鏡像、旋轉、
plane 錯排或裁切。來源、完整索引集合、尺寸、像素契約、方向與 direct-blit 政策為
`已證實`；RGB 仍沿用 mode 0Dh 的 `hardware-spec approximation`。

### TITLE 的 BACKPICS 全螢幕背景

固定 `TITLE.EXE` 的 `sub_10E9F`（loaded-MZ EA `0x10e9f..0x10f33`）以 `back%s`
建立檔名，並呼叫 `sub_11728` 選取 archive payload。IDA 列出的全部 30 個 caller 只來自
`_main` 與 `sub_10C86`，每個呼叫點皆以立即值傳圖號，集合完整覆蓋 0..13。

`sub_11291` 是本 executable 的 9-bit 起始、LSB-first LZW decoder；`sub_114F8` 的控制碼、
high-nibble-first 色彩流與四 EGA planes 寫入邏輯，和既有二階 RLE 契約一致。14 個 payload
全部精確展開為 64,000 個 4-bit indices，即 320×200，且索引皆在 0..15。代表性 ignored
PNG 為山谷雷雨動畫、閃電過場與製作名單，無鏡像、旋轉、plane 錯排或裁切。

BACKPICS 0..13 的 selector、完整 caller 集、LZW、二階 RLE、尺寸、方向與無透明全螢幕
direct-blit 政策為 `已證實`；RGB 仍是 mode 0Dh `hardware-spec approximation`。

### KOLTOWN 的 ENDPICS 結局圖片

固定 KOLTOWN derivative 的 `sub_2B80A`（loaded-MZ EA `0x2b80a..0x2b845`）以
`end%s` 建檔名並把呼叫端圖號交給共用 loader `sub_10F9C`。它的全部七個 caller 都在
`sub_2B4A7`，call EA `0x2b525/0x2b57d/0x2b5d5/0x2b62d/0x2b695/0x2b6ed/0x2b755`
依序傳入 0..6，完整覆蓋 archive。

caller 在每次載入前設定 mode 0Dh byte-column／scanline 矩形；七張尺寸依序為 168×122、
192×111、296×93、176×97、168×123、192×102、184×98。七個 payload 經共用
`sub_1A4FE` LZW、`sub_1C907`／`sub_1A2BB` 二階 RLE 與 EGA planes 後，像素數全數
精確吻合，索引均為 0..15。七張 ignored PNG 目視為結局敘事人物／場景，無鏡像、旋轉、
plane 錯排或裁切。

ENDPICS 0..6 的來源、完整 caller 集、尺寸、LZW、二階 RLE、4-bit EGA、方向與無透明
direct-blit 政策為 `已證實`；RGB 仍採 mode 0Dh `hardware-spec approximation`。

### KOLBOOK 的 BOOKPICS 書本介面

固定 `KOLBOOK.EXE`（SHA-256 `2cada028…5832`）的 `_main` 以 `book%s.0` 建檔名，於
loaded-MZ EA `0x102ac` 以整檔 reader `sub_1374E` 載入 archive。`sub_1137B`
（`0x1137b..0x114f5`）把圖號交給 `sub_13D7D` 做索引／LZW 解碼，再經 `sub_15702`
二階 RLE／EGA planes 與 `sub_13843` rectangle blit。

`sub_1137B` 的全部 caller 只傳常數 0..13 與特殊值 100；`0x11396/0x1139f` 證實
100 會先映為圖號 9，所以實際 archive 索引完整且只限 0..13。幾何分支為：0..3 各
320×200；4..12 各 112×31（9、11 只是右側位置）；13 為 248×22。

14 個 payload 的 LZW／二階 RLE 像素數全部精確吻合，索引均為 0..15。ignored PNG
目視為書本封面／目錄／頁面、章節標題與圖示列，無鏡像、旋轉、plane 錯排或裁切。
BOOKPICS 全 archive 的來源、consumer、alias、索引、幾何、像素格式、方向與無透明
direct-blit 政策為 `已證實`；RGB 是 mode 0Dh `hardware-spec approximation`。

### WEATPICS 分層合成

`sub_2A0C6:0x2a0d5..0x2a11d` 以 `3*group+[0,1,2]` 選擇三層；六組均是
64×62。group 0／1／2／5 的第三層是 496-byte MSB-first 1-bit AND mask，group 3／4
則是二階 RLE 後的四平面 AND mask。`sub_2A18E:0x2a1de..0x2a1e8` 的原始運算是
`(base AND mask) OR overlay`。因此 mask PNG 的像素值是運算資料，不是 alpha。
六張合成圖目視支持晴空／雲／風暴／雨／雪／夜景的分類（`強推論`）；像素契約與公式為 `已證實`。

## 未關閉問題

1. `PICS`／mode 0Dh 的 COMBPICS、TITLE SET0／1／2／FACE／BACK、ENDPICS、BOOKPICS、WEATPICS、TWN 圖號 24、MONPICS、PERPICS、MISCPICS 與 KOLMPICS 圖號 28
   垂直鏈已關閉；PCGA／HERC 與其他
   archive 圖號尚未關閉。
2. 其餘圖號到寬、高、透明規則與裁切矩形的完整映射。
3. 三個固定角色圖片記錄表內部的寬、高、位平面與透明規則。
4. 使用者尚未決定 PCGA／HERC 是必要交付 theme、可選後續 theme，或不納入產品範圍。
5. COMBPICS 全 archive 已關閉，且兩張有外部原版截圖交叉驗證；仍待本機 runtime
   framebuffer 的無損抽樣。

在完整命名、manifest 與失敗政策關閉前，`SPEC-002` 保持 DRAFT；一次性解碼成功不得
接入 Go／Ebitengine。
