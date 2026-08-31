# SPEC-002：PNG 素材抽取契約

- 狀態：`DRAFT`
- 日期：2026-08-31
- 範圍：由固定原版輸入重生本機 PNG 衍生素材；不定義玩法
- RE 來源：`docs/re/image-formats.md`、`docs/re/image-asset-inventory.json`、
  `docs/re/image-lzw-probe-summary.json`

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

- 每個可顯示圖像輸出獨立 PNG；命名必須可回查來源 archive、圖號與顯示模式。
- manifest 至少記錄來源 ZIP／entry SHA-256、來源圖號、壓縮範圍、尺寸、色彩型別、
  palette 來源、透明規則與 PNG SHA-256。
- 抽取器不得把不明尾端 bytes、越界索引、未知 palette 或猜測尺寸靜默 fallback。
- remake 的素材載入測試必須拒絕非 PNG 圖像格式與原版 archive。

## 已有證據

- [x] 42 個圖片候選的大小與 SHA-256 已固定。
- [x] 30 個容器的 20-byte 固定區、2-byte offset table、count、payload bounds 已盤點。
- [x] 30 個容器共 960 個 payload 通過一次性 LZW 機械解碼。
- [x] IDA 已定位索引 reader、LZW core、bit reader、檔名 consumer 與 blitter 候選。
- [ ] suffix／顯示模式、尺寸、位平面、palette 與透明規則尚未證實。
- [ ] 尚無 PNG 通過原版同狀態畫面與像素抽樣。

## READY gate

- [ ] 以 IDA consumer 或原版執行期證據固定至少一個完整格式分支：來源 → 解壓 → 尺寸 →
  位平面／palette → blit。
- [ ] 解釋或排除 `AWD1PICS.0` 索引 payload 後的 520 bytes，不得稱作無害 padding。
- [ ] 對無副檔名 `CHARPICS`／`SHLDPICS`／`STEEDPIC` 建立 writer／reader 與格式分類。
- [ ] 可丟棄 probe 產生至少一張真實 PNG，並通過 bounds、預期尺寸、非空像素、palette、
  畫面方向及原版同狀態視覺抽樣。
- [ ] 預先定義全部 archive 的失敗政策、命名與 manifest schema，完成權利審查。
- [ ] 由第二條獨立證據（原版截圖、runtime buffer 或另一 consumer）交叉驗證像素結果。

## 實作狀態

無 production implementation。`tools/probe_image_lzw.py` 是 RE 階段一次性機械探針，
其輸出明示未驗證尺寸、像素與 palette，不能被 runtime import。
