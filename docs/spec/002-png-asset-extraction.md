# SPEC-002：PNG 素材抽取契約

- 狀態：`DRAFT`
- 日期：2026-08-31
- 範圍：由固定原版輸入重生本機 PNG 衍生素材；不定義玩法
- RE 來源：`docs/re/image-formats.md`、`docs/re/image-asset-inventory.json`、
  `docs/re/image-lzw-probe-summary.json`、`docs/re/town-index24-png-probe-summary.json`、
  `docs/re/monpics-ega-visual-probe.json`、`docs/re/perpics-ega-visual-probe.json`、
  `docs/re/kolmpics-index28-visual-probe.json`、`docs/re/miscpics-ega-visual-probe.json`、
  `docs/re/combpics-all-ega-visual-probe.json`、`docs/re/title-set-face-ega-visual-probe.json`、
  `docs/re/title-backpics-ega-visual-probe.json`、`docs/re/koltown-endpics-ega-visual-probe.json`、
  `docs/re/kolbook-bookpics-ega-visual-probe.json`

## 目的

建立離線、可重生、失敗即關閉的圖片抽取流程。remake runtime 最終只可讀 PNG；不得讀
原版 ZIP、`*PICS*`、無副檔名素材或 DOS executable。本 spec 在 READY 前只允許清冊、
IDA 證據匯出與可丟棄解碼探針，不得建立 production importer 或 Go runtime 路徑。

## 輸入與權利邊界

1. 輸入 ZIP SHA-256 必須是 `2b455e3c…9f46`；不符即建立新 build 身分並停止。
2. 所有原始與解碼後像素資料只可位於本機 ignored 工作區，禁止加入 Git／GitHub Release。
3. 受版控內容只含工具、格式契約、檔名、大小、雜湊與不含 payload 的統計。
4. 工具一律在無網路、非 root、資源受限的一次性 Docker 容器內執行。

## 預定輸出契約

- 每個已驗證的可顯示圖像輸出獨立 PNG，路徑固定為
  `png/<正規化來源檔名>/<四位十進位圖號>.png`。正規化只允許 ASCII 小寫、數字、點、
  底線與連字號；兩個來源若正規化後碰撞，必須失敗，禁止自行覆寫或加隨機尾碼。
- manifest 必須符合 [png-asset-manifest.schema.json](schemas/png-asset-manifest.schema.json)，
  記錄來源 ZIP／entry SHA-256、來源種類與圖號、顯示模式、證據等級、處置、壓縮範圍、
  解碼契約、尺寸、色彩型別、palette 來源、透明規則、權利分類與 PNG SHA-256。
- 未證實項目也必須有一筆 manifest，`disposition` 設為 `excluded_unknown`，且不得產生
  `output_png`。只有 `evidence_status=confirmed` 且欄位完整者可設為 `png`。
- 抽取器不得把不明尾端 bytes、越界索引、未知 palette 或猜測尺寸靜默 fallback。
- remake 的素材載入測試必須拒絕非 PNG 圖像格式與原版 archive。

## 失敗政策

下列任一條件立即非零退出，且不得留下半套現行輸出：輸入或 entry 雜湊不符、重複或
非遞增 offset、payload 越界、解壓錯誤、像素數不符、色彩索引越界、未知 palette／透明
規則卻要求輸出、命名碰撞、schema 驗證失敗、輸出 PNG 重讀不一致。工具先寫入同一輸出
根目錄下的新暫存批次，整批驗證後才替換該版本；舊批次與其他專案不得刪除。

未知或尚未支援的格式不是整份清冊錯誤：它必須明列 `excluded_unknown`，不可被 runtime
引用。若使用者要求的 extraction set 含任何 excluded 項目，該批次仍失敗，以免把「部分
成功」冒充「素材已全部抽取」。

## 權利審查

- 原版 archive、解碼像素與其 PNG 衍生物一律標為 `local_original_derivative`，只留本機
  ignored 工作區或明確授權的 private 交付位置，不加入 Git、GitHub Release 或公開包。
- Git 可保存不含 payload 的工具、schema、來源雜湊、位址、統計與 manifest 範本。
- 公開 `dist-all/<版本>/release/` 只能使用另行證明可散布的素材，或提供合法原版資料的
  本機匯入流程；private GitHub 本身不會把原版素材轉成可公開授權。

## 已有證據

- [x] 44 個圖片候選的大小與 SHA-256 已固定；新增涵蓋 AWD 的 PCGA／HERC 變體。
- [x] 30 個容器的 20-byte 固定區、2-byte offset table、count、payload bounds 已盤點。
- [x] 30 個邏輯 archive 共 1,022 個 payload 通過一次性 LZW 機械解碼；`.1/.2` 已依
  KOLTOWN 分卷 consumer 納入，不再被誤列為未知獨立格式。
- [x] IDA 已定位索引 reader、LZW core、bit reader、檔名 consumer 與 blitter 候選。
- [x] `COMBPICS.0` 全七筆已證實：0..4 是 296×144、5..6 是 88×64；全部使用 mode
  0Dh／PICS、二階 RLE、4-bit indices 與 EGA palette，consumer 集合完整覆蓋 0..6。
- [x] TITLE 的 `SET0PICS.0`、`SET1PICS.0`、`SET2PICS.0`、`FACEPICS.0` 共 24 筆已
  完整證實；每包圖號 0..5，尺寸依序為 88×70、104×84、160×125、112×80。
- [x] TITLE 的 `BACKPICS.0` 全 14 筆已證實為 320×200、4-bit EGA、無透明全螢幕背景；
  全部 30 個 caller 的常數圖號集合完整覆蓋 0..13。
- [x] KOLTOWN 的 `ENDPICS.0` 全七筆已證實；圖號 0..6、七種 caller 矩形、4-bit EGA
  與無透明 direct-blit 政策均已關閉。
- [x] KOLBOOK 的 `BOOKPICS.0` 全 14 筆已證實；0..3 為 320×200、4..12 為
  112×31、13 為 248×22，特殊 argument 100 已證實映為圖號 9。
- [x] KOLTOWN 的 `WEATPICS.0` 全 18 筆已證實分為六組 64×62 圖層；每組依序為
  `base_color`、`or_color`、`and_mask`，以 `(base AND mask) OR overlay` 合成。第 3、4 組
  遮罩為四平面 4-bit，其餘為 MSB-first 1-bit；六張合成圖目視支持野外天候／時段背景分類。
- [x] `AWD1PICS.0` 全 24 筆已證實為 24×21、4-bit EGA、無獨立透明遮罩的 direct-blit
  圖示；consumer 完整覆蓋 0..23，接觸表與手冊 `Medal` 命令交叉支持任務勳章分類。
- [x] `SHLDPICS.0` 八筆已證實為四組彩圖／AND-mask 盾牌；尺寸為 40×29、
  56×45、48×63、48×61，mask bit 1 透明、bit 0 不透明。它與無副檔名
  `SHLDPICS` 的 16×760-byte 角色記錄表是不同格式。
- [x] `HANDPICS.0` 全 24 筆已證實為六個 renderer class×兩個相對位置的 12 組
  手部彩圖／AND-mask；尺寸只取 16×6、16×8、24×7、24×8，alpha 為 `1-mask`。
- [x] `WEAPPICS.0` 全 56 筆已證實為 28 組角色武器的「偶數 4-bit EGA 彩圖＋奇數
  MSB-first AND-mask」；裝備欄位 selector、六值基表、特殊 variant 交換、28 組尺寸、
  六類位置表、像素數與 `alpha=1-mask` 均由 IDA／棋盤接觸表證實。family／variant
  只作資料流導覽，不宣稱遊戲內分類語意。
- [x] `TWN0..7PICS` 全域圖號 24 均跨入 `.1`，已證實 town 0 為 192×96、其餘為
  88×64，並通過八張 continuation PNG 的幾何、方向、4-bit bounds 與接觸表目視。
- [x] `MONPICS.0` 的 64 個 payload 已證實是 32 組「偶數 4-bit EGA 彩圖＋奇數 1-bit
  MSB-first 遮罩」；逐組尺寸表、像素數、透明邊界與方向均經接觸表抽查。
- [x] `PERPICS.0` 的 12 個 payload 已證實是六組「前六筆 4-bit EGA 彩圖＋後六筆 1-bit
  MSB-first 遮罩」；`[i,i+6]` 配對、六組尺寸、像素數、方向與透明邊界均已抽查。
- [x] `KOLMPICS.0` 圖號 28 已證實為 56×57、MSB-first EGA AND-mask；bit 1 保留背景、
  bit 0 讓彩圖覆寫。它只能作合成 mask，禁止冒充獨立彩圖；其餘 82 筆仍未知。
- [x] `MISCPICS.0` 已完整分類：image 0／1／3／4／5 是五張 direct-blit 4-bit EGA
  彩圖；image 2 是全程未引用的一-byte placeholder，必須 `excluded_non_image`、無 PNG。
- [x] 兩張 COMBPICS 測試 PNG 已對上魔旅士原版文章中的同場景截圖；量化摘要保留正配
  與交叉錯配矩陣，圖片本身不入 Git。
- [ ] 其他 archive／PCGA／HERC 的尺寸、位平面、palette 與透明規則尚未證實。

## READY gate

- [x] 以 IDA consumer 或原版執行期證據固定至少一個完整格式分支：來源 → 解壓 → 尺寸 →
  位平面／palette → blit。
- [x] `AWD1PICS.0` 尾端 520 bytes 已排除於 KOLTOWN 索引圖片流，分類為固定 4000-byte
  編輯器工作區快照的未解析尾區；不得稱作 padding 或輸出為 PNG 像素。
- [x] `CHARPICS`／`SHLDPICS`／`STEEDPIC` 已由 writer／reader 資料流分類為 16 筆、
  每筆 760 bytes 的固定可變記錄表；內部像素佈局仍未知，未授權正式抽取。
- [x] 可丟棄 probe 產生至少一張真實 PNG，並通過 bounds、預期尺寸、非空像素、palette、
  畫面方向及原版同狀態視覺抽樣。
- [x] 已預先定義全部 archive 的失敗政策、可回查命名、manifest schema 與權利審查。
- [ ] 每種預定輸出格式皆已有已證實像素契約；其餘明列排除，且完整 extraction set 不得
  把排除項目冒充成功。
- [ ] 使用者確認顯示 theme 範圍：EGA-only、EGA 基準加後續 CGA／Hercules，或三者同為
  首版必要交付；未確認前不得把建議固化成正式 extraction set。
- [x] 由第二條獨立證據（兩張原版文章截圖＋DOSBox-X mode／palette 契約）交叉驗證像素結果。

## 實作狀態

無 production implementation。`tools/probe_image_lzw.py` 是 RE 階段一次性機械探針，
其批次輸出仍明示未驗證多數圖片的尺寸、像素與 palette，不能被 runtime import；
`tools/probe_extract_comb_png.py`、`tools/probe_extract_monpics_png.py` 與
`tools/probe_extract_perpics_png.py` 只關閉各自的窄分支，也不得作 runtime importer；
`tools/probe_extract_kolmpics_index28_png.py` 同樣只驗證單一純 mask，不能直接成為正式素材；
`tools/probe_extract_miscpics_png.py` 驗證 MISCPICS 全 archive，但仍是 RE probe，不是正式
抽取器；`tools/probe_extract_comb_all_png.py` 與 `tools/probe_extract_title_set_face_png.py`
以及 `tools/probe_extract_title_backpics_png.py` 同樣只作各自 archive 的 RE 驗證。
`tools/probe_extract_koltown_endpics_png.py` 也僅驗證 ENDPICS，不是 production importer。
`tools/probe_extract_kolbook_bookpics_png.py` 也僅驗證 BOOKPICS，不是 production importer。
`tools/probe_extract_koltown_weatpics_png.py` 僅驗證 WEATPICS 圖層與合成契約；
遮罩 PNG 儲存的是 AND 資料，不是 RGBA 透明度，不得作為獨立美術圖顯示。
`tools/probe_extract_awdpics_png.py` 僅關閉 bundled `AWD1PICS`；consumer 可建立
`AWD1..AWD5` 檔名不代表其餘角色檔可以未驗證資料代替。
`tools/probe_extract_shldpics_png.py` 只讀索引 archive `SHLDPICS.0`，並以 `1-mask`
產生 PNG alpha；禁止把無副檔名記錄表交給它。MONPICS／PERPICS 早期探針直接把
AND-mask bit 當 alpha，與已證實的 `sub_1BD17` 極性契約衝突；已改為 `1-mask`
並用 8×8 棋盤背景重生全部 PNG、接觸表與安全摘要。
`tools/probe_extract_handpics_png.py` 同樣將 mask 當作 AND 運算資料，並以棋盤接觸表
驗證透明輪廓；manifest 只能使用「第一／第二位置」，未有更強證據前不命名解剖學左／右手。
`tools/probe_extract_weappics_png.py` 只關閉 `WEAPPICS.0` 的 28 組武器；它依安全 IDA
摘要取得尺寸並以 `1-mask` 生成 RGBA，不能據此泛化角色其他身體部件或成為 runtime importer。
官方 `MAKEMODE.EXE` 已證實可由 PICS 生成 PCGA／HERC archive，但完整轉換規則尚未關閉，
且產品 theme 範圍待使用者決策，故不得先接入正式抽取流程。
