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
- 圖片 RE 清冊已固定 44 個實體候選、30 個邏輯 archive；TWN `.1/.2` 為 continuation
  volume，總計 1,022 payload 通過 LZW 機械探針。
  SPEC-002 已建立為 DRAFT；大多數 archive 的像素、尺寸與 palette 尚未證實。
- COMBPICS EGA 窄分支已證實 296×144、二階 RLE、4-bit 索引與 BIOS mode 0Dh palette；
  兩張 ignored 測試 PNG 對上魔旅士文章的同場景 JPEG，仍沒有 production PNG。
- AWD 三變體已證實為固定 4000-byte 工作區快照；PICS 尾端 520 bytes 不屬索引圖片流，
  但仍是 opaque workspace tail。三個無副檔名角色圖片檔已證實為 16×760-byte 記錄表。
- TWN 全域圖號 24 已完成 `.1` continuation → LZW → 二階 RLE → 4-bit EGA → blit 垂直鏈；
  town 0 為 192×96，其餘七城為 88×64，八張 ignored PNG 接觸表方向正常。
- MONPICS 已完成 32 組「4-bit EGA 彩圖＋1-bit 遮罩」垂直鏈；64 個 payload 與 IDA
  尺寸表逐組吻合，ignored RGBA PNG 接觸表的方向與透明邊界正常。
- MAKEMODE 已證實是官方 CGA／Hercules 轉換器；28 組 PICS 圖數與 v2 清冊逐組相符，
  會生成 `PCGA`／`HERC` 分卷。完整像素規則未關閉，首版 theme 範圍待使用者決定。
- SPEC-003 已 READY；玩家手冊 79 頁與參考卡 5 頁的逐頁 Poppler 雜湊清冊可重生。
  玩家手冊實體頁 1–5 已 reviewed；76–77 為目視確認空白頁，其他 77 頁仍 todo。

## 證據入口

| 主題 | 位置 | 說明 |
|---|---|---|
| 目前真相 | `CONTEXT.md` | 不用歷史工作清單反向覆蓋 |
| 原版清冊 | `docs/re/original-inventory.json` | 雙容器重生逐 byte 相同；只有雜湊與 metadata |
| 研究結論 | `docs/re/RESEARCH-LOG.md` | RE-001 至 RE-010 |
| 圖片格式 | `docs/re/image-formats.md` | 索引、LZW、consumer 與未知項 |
| entry 原始 bytes | `docs/re/ida-fingerprints/KOLTOWN-entry.json` | 僅本機，Git 精確排除 |

## 尚待處理

| 優先 | 類型 | 工作 | 下一個精確動作 | 阻塞 |
|---|---|---|---|---|
| 1 | product decision | theme 範圍 | 使用者選 EGA-only、CGA／Hercules 後續，或三者首版必備 | 等待使用者 |
| 2 | documentation | 中文手冊 | 依 SPEC-003 翻譯並校對玩家手冊實體頁 6–10 | 77 頁仍 todo |
| 3 | asset RE | SPEC-002 READY 證據 | MONPICS 已關閉；追下一個 EGA PICS consumer 圖號 | 多數像素契約未知 |

## 執行中程序與暫存物

無。每輪結束仍須重新檢查本專案 Docker 容器。

## 不要重複

- 不要把 1998 timestamp 的檔案當成已證實原始發行媒體。
- 不要因檔名含 `PICS` 就宣稱圖像格式已解碼。
- 不要把 1,022 個 LZW payload 機械成功冒充全部正確像素；AWD 尾區雖已排除於圖片流，仍不是
  可任意丟棄的 padding。
- 不要在各格式／玩法 spec READY 前開始對應的 Go runtime。
- 不要把 `KOLTOWN-entry.json` 的線性 32 條指令當成完整控制流。
- 不要因 SPEC-001 READY 就開始猜玩法；它只授權 derivative 作研究 oracle。
