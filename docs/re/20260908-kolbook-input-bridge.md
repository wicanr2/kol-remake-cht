# 書本讀鍵鏈：BIOS 檢查與 DOS 取字（2026-09-08）

## 輸入與工具

- 輸入：`KOLBOOK.EXE`，SHA-256 `2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832`。
- 工具：IDA Pro 9.4；映像 `ida-pro-9.4-idapython:locked-v1`，ID `6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780`。
- 一次性資料庫從唯讀輸入建立，辨識 191 個函式；程式段 IDA linear `0x10000`、資料段 `0x19240`。
- 可重生探針：`tools/ida_probe_kolbook_input.py`。完整 bytes／operand／xref 匯出僅本機：
  `workplace/kolbook-input-20260908/kolbook-input-private.json`，不公開反組譯 payload。

## 已證實的最小鏈

| 原始定位（IDA linear） | 附加語意與等級 | 證據 |
| --- | --- | --- |
| `sub_146CE:0x14B60` | 已證實：設 ES=0 | `33 C0 8E C0` |
| `0x14B64`、`0x14B68` | 已證實：讀實體 `0000:041A` 與 `0000:041C`，計算是否不同 | `26 A0 1A 04`、`26 2A 06 1C 04` |
| `0x14BA6`、`0x14BAF` | 已證實：非空才離開輪詢並呼叫 `_getch` | 區域變數比較／直接呼叫 xref |
| `_getch:0x187D2`、`0x187F4`、`0x187F5` | 已證實：DH=08，AX/DX 交換後 INT21，即 DOS AH08 | `B6 08`、`92`、`CD 21` |
| `0x14BC4..0x14BD5` | 已證實：第一個返回值是 0 時再次 `_getch`，第二個值左移八位 | 比較、第二個 call、shift 控制流 |

函式名稱由 IDA 原始分析／FLIRT 提供，沒有推測性改名。語意依上述 bytes 與資料流建立，名稱本身不作證據。
實體 `0000:041A/041C` 是遊戲直接引用的位址，不是 IDA linear。先前 dosgolem 的
執行期 `0110:87F5` 可作定位線索，不能直接與 IDA 位址混用。

## 對既有診斷的訂正

上輪已觀察到 BIOS 空白鍵讓封面翻到目錄，但已讀計數為 0。現在可以解釋：
遊戲直接查看 BDA，所以那一步不會增加 BIOS 呼叫計數；接著 `_getch` 走 DOS AH08，
舊 dosgolem 卻只查另一個 Stdin 佇列。先前以「有畫面推進但讀鍵計數為 0」推測讀鍵行為的
範圍不足，應以這條跨介面資料流及重跑結果為準。

先前色盤與 AH06 缺陷各自成立，但它們不是此讀鍵鏈的完整原因；不刪除舊失敗收據。
dosgolem 正式修正依 `187-bios-console-read-bridge.md` READY 契約，保留 Stdin 優先，
沒有 Stdin 時從 BDA 取鍵；擴充鍵的第二個 byte 使用既有 Stdin 保存，避免雙讀同一硬體鍵。

同一函式中的目錄滑鼠範圍也已定位，但不是本修正必要條件，未再深入無關分支。

## 修正後驗證

`workplace/kolbook-input-20260908/` 保存此次執行前輸入清冊、測試紀錄與原版擷取。
dosgolem 核心回歸與六項 KOL 規則對照均通過（規則組 142.962 秒）。
`TestNormalBookEntry` 通過（48.978 秒）：從 `KOL.EXE` 正常啟動，片頭送 DOS Enter，
書本依序送 BIOS 空白、6、Enter、空白，載入 `KOLTOWN.EXE`。
人工檢視 `book-final.png` 確認 Halfway House、Delvin Jorlack 對話及旅店圖示已顯示。
這證實本條原版正常入口已打通；尚不代表 remake 旅店、戰鬥或存讀檔已完成同狀態對拍。
紀錄仍有 `region1.dat` 至 `region4.dat` 的探測失敗，沒有因此補造檔案或宣稱全部資料已驗收。

第一次 IDA 匯出因預設 ASCII 編碼失敗，修為明確 UTF-8 後於同一映像乾淨重跑成功；
此為驗證腳本問題。一次性 IDA 與驗證容器均已移除。
