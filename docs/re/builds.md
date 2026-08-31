# Build 與衍生映像台帳

## 原始封裝

| 身分 | SHA-256 | Bytes | 權利／位置 |
|---|---|---:|---|
| `Knights-of-legend.zip` | `2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46` | 1,699,452 | 僅本機、唯讀 |
| ZIP 內 `KOLTOWN.EXE` | `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221` | 133,760 | 僅本機、唯讀 |

完整 76 項中央目錄與逐檔 hash 見 `original-inventory.json`。

## KOLTOWN 解包衍生映像

| 項目 | 值 |
|---|---|
| 輸出 | `workplace/analysis/koltown/KOLTOWN-unpacked.EXE`（Git ignored） |
| SHA-256 | `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2` |
| Bytes | 138,528 |
| 解包器來源 | `/home/anr2/openkb/openkb-code/src/tools/unexepack.c` |
| 解包器來源 SHA-256 | `0e1da11a1d5a4aa5722dc83dd16c683b4c2c25cec31d3741ee71039840a6b03c` |
| 解包器授權 | 檔頭明示 CC0 1.0；未把來源複製進本專案 |
| 建置環境 | `gcc:13`、`-std=c99 -O2 -Wall -Wextra`、無網路容器、UID/GID 1000 |
| 解包器 binary SHA-256 | `998569497d3e29c137821e79635b2a411f04c1d236bbf2a66b7d52929ca12bec` |
| 限制 | 上游明示 garbage bytes 不保證與 Microsoft 原解包結果逐 byte 相同；只作研究衍生映像 |

### 解包結果不變量

- MZ header：`CS:IP=1b8f:001c`、`SS:SP=2b48:07d0`。
- Header paragraphs：576；relocation entries：2,281。
- 解壓資料長度：129,312 bytes；解壓寫入 16,421 bytes，消耗 15,381 packed bytes。
- IDA Pro 9.4 重新載入後，entry linear EA 為 `0x2b90c`；強制建立入口函式後資料庫辨識
  337 個函式，證明衍生映像可供 xref／函式邊界研究。

### 已知建置失敗

上游 C99 原始碼在 GCC 13 加 `-Werror` 時因舊式 `%d`／pointer signedness／未使用變數
警告而失敗。正式成功命令保留 `-Wall -Wextra` warnings，但不修改上游來源；輸出仍須由
header、hash 與 IDA 重載三層驗證。
