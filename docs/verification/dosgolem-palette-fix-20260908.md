# 原版色盤修復與正常書本入口（2026-09-08）

## 目前結果

- dosgolem 現用 `kol-force-call` 為 `fd746405` 加本機介面與色盤修正，尚未提交／推送。
- 0Dh／0Eh 的 BIOS 色盤初始化已補齊；相同片頭路徑的六張 RGB 圖片全部恢復顯示，已逐張檢視。
- dosgolem 核心套件通過；KOL 六項規則對照通過（123.786 秒）。
- 未修改的 `KOL.EXE` 經 DOS Enter 正常載入 `KOLBOOK.EXE`，確認看到書本封面。
- **尚未證明正常旅店、完整玩家路徑或 AppImage 操作等價。** 存檔、中文資料頁等原有缺口仍需完成。
- 後續另修正 DOS `AH=06h/DL=FFh` 永遠回「無字元」的缺陷，核心套件及六項 KOL 規則再驗通過（118.650 秒）；正常書本選入旅店仍未通過，沒有把獨立缺陷冒稱成書本卡住的已證實根因。

本頁取代前輪「RGB 全黑仍未修復」的目前狀態，原失敗證據與形成原因保留於
[上輪紀錄](dosgolem-integration-20260908.md)，不重寫歷史。

## 原因與修正

上輪畫面索引非零，但 DAC 色盤全為黑色。`SetVideoMode` 重設平面控制器後，沒有執行
BIOS 預設色盤載入。`PlanarRGB()` 正確地查詢實際 DAC，因此輸出全黑。

本輪依 [SeaBIOS 的模式表與模式設定流程](https://github.com/coreboot/seabios/blob/master/vgasrc/stdvgamodes.c)
建立 dosgolem `docs/spec/194-ega-bios-default-palette.md` 的 READY 契約（2026-09-10 移植到 dosgolem `main` 時由 185 改編為 194：main 已用掉 185，而且是兩份），再以獨立色彩位元公式實作：
0Dh／0Eh 載入 CGA 相容的 64 色 DAC 與屬性色號映射；保留遊戲後續 DAC 改寫，重新設定模式恢復預設。
未移植 SeaBIOS 程式，也未替截圖事後套色。

來源快照 SHA-256 為 `8ae58eb660e0b002b19a64c9c98992408a65d55d1379e7383d376f4868cb8411`，
存取日期 2026-09-08，原始來源及來源授權記錄僅本機保存。
這是有來源支持的 BIOS 相容契約，不宣稱原版硬體逐週期一致；其他模式的完整預設色盤不在本切片驗收範圍。

## 驗證

新增機器層回歸覆蓋棕色、深灰、白色、埠改寫 DAC、模式重設與非目標模式邊界。
修復後原版片頭六段各 2,000 萬指令，畫面索引雜湊與修復前逐段相同；RGB 已能辨識雷電、
Origin 標誌與製作人員文字。不是以「檔案存在」或單純非黑計數取代逐張檢視。

正常書本探針由 `KOL.EXE` 開始，不改遊戲記憶體、不替換 EXE、不跳入口：

1. 硬體 Enter 測試至十億指令仍循環片頭，失敗日誌及畫面保留。
2. 改用 dosgolem 的 DOS 主控台 Enter，於 2.5 億指令階段看到 `KOLBOOK.EXE` 開檔，
   再執行 500 萬指令後截圖確認書本封面；此探針 26.871 秒通過。
3. DOS 主控台 SPACE／6／Enter／SPACE 未使書本進入旅店。這說明輸入介面需要按階段驗證，
   不能從片頭能退出推論書本也會讀取同一佇列。
4. BIOS 空白鍵已實際翻到目錄，但計數仍是「已讀 0」，證明計數不足以否定畫面觀測。
   2,000 萬指令時的畫面尚在翻頁；改成 8,000 萬指令後目錄穩定，數字 6 仍未選入。
   滑鼠點擊目錄第六列也沒有推進。失敗日誌與翻頁中間畫面保留，不再用增加等待當作修復。

## DOS AH06 的獨立修正與下一個最小動作

`conOut` 的 AH06 輸入分支原先完全不讀 Stdin，永遠設定 ZF。
依 [Microsoft MS-DOS 4.0 Programmer’s Reference](https://www.pcjs.org/documents/books/mspl13/msdos/dosref40/)
的 Function 06H，補齊有字元時取出一 byte、回 AL 並清 ZF；空佇列才回 AL=0／ZF=1，
不回顯、不阻塞。dosgolem READY 規格為 `186-dos-direct-console-input.md`，新增回歸覆蓋
一般字元、NUL、Ctrl-C 字元、順序、空佇列及軌跡。沒有順便合併 BIOS／DOS 佇列。

原版正常路徑在修復後仍未到旅店（65.646 秒，FAIL），因此下一個最小動作是：
使用既有 KOLBOOK IDA 資料庫定位目錄讀鍵常式，對照執行期 `0110:87F5` 附近的
實際輸入服務與暫存器／BDA 讀寫；先取得讀鍵鏈再修執行器。上述 `0110:87F5`
是 dosgolem 執行期段位址，不是 IDA 線性位址，不可直接混用。

## 收據與重跑

本機 `workplace/dosgolem-palette-20260908/` 保留來源、PNG、日誌、失敗歷史與 manifest。
舊全黑圖片仍在 `workplace/dosgolem-integration-20260908/`。
AH06 修復後日誌與圖片在 `workplace/dosgolem-console-20260908/`。本輪較早的 manifest
是執行後工作樹清冊，已明示不能單獨證明歷史執行來源；新腳本會在複製來源前固定
input manifest，結束後檢查來源是否變動。測試判定仍以日誌與逐張畫面檢視為準。

在既有 `kol-package:20260907` 容器中，依 `tools/oracle/check_updated_dosgolem.sh` 的掛載契約執行；
`KOL_CHECK=normal` 為片頭六段，`KOL_CHECK=book` 為正常書本與後續玩家輸入探針。
所有原始遊戲檔案唯讀，Docker 斷網、限額、非 root、外層逾時與 `--rm`；只有一次公開平台來源下載開網路。

本輪容器均已移除；新增檔案與本機輸出擁有權 1000:1000。未重打包、提交或推送。
