# 現行三平台建置與 Windows Wine 正常保存

日期：2026-10-01。此收據驗證目前未提交工作樹的中間建置，版本為 `development`。
完整來源、工具、畫格與雜湊見[正式收據](platform-smoke-20261001/receipt.json)；
逐項結果見[核對報告](platform-smoke-20261001/checks.json)。

| 範圍 | 結果 | 尚未涵蓋 |
|---|---|---|
| Linux amd64 | 兩次編譯相同，仍為現行 `f969a63f`；正常互動及預設保存、重載通過 | 正式新封包與主機 FUSE |
| Windows amd64 | 兩次編譯相同；`304890ea` 在 Wine 9 實際啟動，預設繁中與 `%APPDATA%` 保存、重載通過 | Windows 真機與正式 ZIP |
| macOS Intel／Apple Silicon | 通用檔兩個架構與各自單架構檔相同；最低 11.0、arm64 簽章區塊與系統相依通過 | 實際啟動、真機、完整簽章與正式 `.app` |

兩個執行平台都以啟動前固定的 `-seed 1`，從標題正常招募六人，走到第七城酒館，
購買「吉拉克赫爾巴克」後帶走，檢視第一背包格，再正常回旅店住宿保存。
各 198 個 X11 真實鍵盤事件都由程式的 `-log-input` 確認；沒有指定 `-scene`、
`-lang` 或 `-save`，因此實際走互動入口與預設中文、預設保存位置。
Windows 保存於 `%APPDATA%\kol-remake-cht\save.dat`；Linux 保存於
`$XDG_CONFIG_HOME/kol-remake-cht/save.dat`。資源工作目錄唯讀，保存輸出可寫。

金冠依序為 1,009→974→918，扣食品 35 與住宿 56。兩側 v3 保存除 `saved_at` 外
完全相同，含六人隊伍與食品的原始識別記錄。重新啟動時明寫 `-load` 指向本次正常保存，
各 15 個互動按鍵再次檢視食品；原存檔雜湊未變。此處沒有驗證自動續玩選單。

批次正常流程各 199 筆、重載各 16 筆狀態，與既有
[初始物品收據](inventory-empty-slots-20261001/receipt.json)完全相同。
40 組完整畫格比較差異為零，包括互動對批次、兩平台互動及中文封面。
室內與物品畫面逐像素比較；即時標題與戶外隊伍標記會動，不列入靜止整圖比較。
Windows 的商品、背包、住宿與重載明細已目視可讀，Linux 對應靜止畫格相同。

首輪批次未指定 `-save`，兩側購物完成但沒有保存；查明批次旗標排除互動初始化，
因此改驗真實互動入口，沒有修改產品。Linux 首輪退出時，擷取器遇到 X11 關閉競速；
保留原 198 鍵與保存，修擷取器後以新設定目錄乾淨重跑，程式退出碼為 0。
首輪控制快照及失敗位置均登記於正式收據。原始存檔、軌跡、執行檔、Wine 環境
與 SDK 只留在 `workplace/platform-smoke-20261001/`；36 張正式 PNG 僅適用既有
private GitHub 特例，沒有公開散布授權。

重核使用既有 `kol-package:20260907` 映像，在 Docker 中執行
[核對器](platform-smoke-20261001/controls/kol_platform_verify.py)。它需唯讀掛載儲存庫至
`/work`，只將已存在的 `workplace/platform-smoke-20261001/` 掛為可寫；使用目前
UID/GID、`--rm --network none --memory 1g --cpus 2 --pids-limit 64` 與外層 90 秒逾時。
執行前確認兩個掛載來源均為目錄。完整編譯與 GUI 控制來源保存在同一收據的
`controls/`，首輪快照則在 `controls-attempt1/`；重播需新的乾淨輸出與設定目錄，
不得覆寫本次證據或拿前次保存當成本次結果。

此驗證沒有建立 `dist-all/`、正式版號或 Release。Wine 結果不等於 Windows 真機；
macOS 靜態結構不等於啟動。#37、#42 與 #7 保持開放，本地仍 23 項未完成、83 項完成。
音訊人耳、正式封包、完整戰鬥／任務及原版同狀態依各自閘門繼續驗收。
