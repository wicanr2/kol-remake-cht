# 交接

- 日期：2026-08-31
- 儲存庫：`wicanr2/kol-remake-cht`（private）
- HEAD：以 `git rev-parse HEAD` 為準
- 分支：`main`
- 目標：完成 RE／spec 後再進入素材抽取與 remake 實作。

## 已完成並驗證

- 專案規範與 GitHub 骨架已建立。
- 六個根目錄輸入已固定大小與 SHA-256。
- ZIP 共 76 項，CRC 通過；6 個 executable 皆有 MZ header。
- PDF 頁數／尺寸與 JPEG 尺寸已記錄。
- 六個 executable 已由 IDA Pro 9.4 匯出第一層指紋；四檔有 Microsoft runtime signature
  候選，`KOLTOWN`／`TANDYPIN` 仍未知。
- RE 知識庫已建立；KOLTOWN 的 Microsoft EXEPACK、真正入口與 337-function unpacked IDA
  derivative 已固定。SPEC-001 已 READY，startup 前 runtime load image 完整零差異核對通過。
- 圖片 RE 清冊已固定 42 個候選；30 個索引容器共 960 payload 通過 LZW 機械探針。
  SPEC-002 已建立為 DRAFT；像素、尺寸與 palette 尚未證實，沒有 production PNG。

## 證據入口

| 主題 | 位置 | 說明 |
|---|---|---|
| 目前真相 | `CONTEXT.md` | 不用歷史工作清單反向覆蓋 |
| 原版清冊 | `docs/re/original-inventory.json` | 雙容器重生逐 byte 相同；只有雜湊與 metadata |
| 研究結論 | `docs/re/RESEARCH-LOG.md` | RE-001 至 RE-005 |
| 圖片格式 | `docs/re/image-formats.md` | 索引、LZW、consumer 與未知項 |
| entry 原始 bytes | `docs/re/ida-fingerprints/KOLTOWN-entry.json` | 僅本機，Git 精確排除 |

## 尚待處理

| 優先 | 類型 | 工作 | 下一個精確動作 | 阻塞 |
|---|---|---|---|---|
| 1 | asset RE | SPEC-002 READY 證據 | 沿一個 consumer 固定尺寸／位平面／palette，產生同狀態 PNG probe | 像素契約未知 |
| 3 | documentation | 中文手冊 | 建頁面索引、OCR 草稿與術語表 | M1 後開始 |

## 執行中程序與暫存物

無。每輪結束仍須重新檢查本專案 Docker 容器。

## 不要重複

- 不要把 1998 timestamp 的檔案當成已證實原始發行媒體。
- 不要因檔名含 `PICS` 就宣稱圖像格式已解碼。
- 不要把 960 個 LZW payload 機械成功冒充正確像素；`AWD1PICS.0` 尾端 520 bytes 未分類。
- 不要在各格式／玩法 spec READY 前開始對應的 Go runtime。
- 不要把 `KOLTOWN-entry.json` 的線性 32 條指令當成完整控制流。
- 不要因 SPEC-001 READY 就開始猜玩法；它只授權 derivative 作研究 oracle。
