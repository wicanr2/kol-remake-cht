# 三平台 HD 更新交付

版本：`v.1.0.1-20261012`。狀態：本機交付驗收通過；private GitHub Release 已發布。

本版依使用者指令重建三平台完整版、Release 引擎包與推廣影片，包含中文修正、
首批 HD 及遊戲內 EGA／HD 切換。正式程式固定為 `343c0f87` 的 356 份 Go／module 輸入，
與[畫面切換驗收](display-switch-runtime-20261012.md)逐項相符。
不更改玩法、存檔格式與未核准的後續 HD 美術。

交付根目錄為 [dist-all/v.1.0.1-20261012/](../../dist-all/v.1.0.1-20261012/)。
本專案依 AGENTS.md 使用 `full/`、`release/`、`promo/`；`smoke/` 保存驗收證據。
前版交付保持。完整版含遊戲及首批 HD 資料，影片含原作衍生畫面，均留本機。
GitHub Release 僅上傳三平台引擎包及對應 SHA-256，不改 private 儲存庫可見度。
權利判定見[本版審查](../RELEASE-RIGHTS-REVIEW.md#v101-20261012-發行審查)。

| 項目 | 執行前固定的驗收條件 |
|---|---|
| 建置 | Linux AMD64、Windows AMD64、macOS ARM64／Intel universal；Go1.26.7／Ebitengine2.9.9 |
| 正式程式 | 356 份輸入與最近已核准版本相同；正式套件測試通過 |
| 正常抽樣 | seed1，沿前版唯一正常鍵序，戰鬥、回城住宿保存及新程序重載 |
| 狀態比較 | 只容許 battle.exec_text／encounter_text 的既有中文修正；其他 trace 相同，完整 v3 除 saved_at 相同 |
| 遊戲內切換 | 實際封包開啟 F10、切 HD、返回 EGA；Linux 檢查核准畫面、縮放、鍵鼠及住宿保存 |
| 引擎包 | 無原版素材、HD 圖片、音訊或完整手冊譯文；先核全部合法抽取資料雜湊再匯入 |
| Windows | 實際 ZIP 解壓後在 Wine9 啟動 full／匯入後 release，繪製中文 EGA／HD 並操作切換 |
| macOS | universal、最低11.0、ARM64 ad-hoc 簽章、系統相依及 app 執行位元；不宣稱真機可玩 |
| 影片 | 本版新畫格、76秒1080p、H.264／AAC、字幕安全範圍、音量、黑幀、停格及逐幕目視 |
| 發布 | 六包完整性、manifest 與 SHA-256 通過；GitHub 僅三個引擎包與其 checksum |

重建入口為 [closeout_release.py](../../tools/closeout_release.py)，每個階段必須傳入
`--version v.1.0.1-20261012`。流程依序為 `prepare`、`native`、macOS 建置、`assemble`、
`sample`、[封包執行驗證](../../tools/verify_release_runtime.py)、
[遊戲內切換驗證](../../tools/verify_display_switch.py)、
[影片合成](../../tools/compose_release_promo.py)、`verify`。

全程在有界 Docker 中執行，來源 `/src:ro`、新版本 `/out` 可寫，UID/GID1000:1000，無網路。
Linux 建置與影片沿 `kol-package:1.26.7-ebiten2.9.9-20261011-r1`；
macOS 沿 `kol-osxcross-go:20261007-r1`；Wine 視窗操作沿現有
`eob-remake-wine-verify:ubuntu-noble-20261008-r2`，因本專案 Wine 映像缺 xdotool。
Xvfb 須等待就緒，並由擁有程序的退出 trap 回收。

驗證器 `linux` 階段準備 `smoke/linux-gui/`，再將該目錄掛 `/out`、版本目錄掛 `/delivery:ro`，
在1920×1200 Xvfb中執行 `verify_display_switch.py --binary /out/run-full`。
所有來源、工具、逐檔大小與雜湊保存在交付目錄，不把封包或影片加入 Git。

| 實際驗收 | 結果 |
|---|---|
| 正式來源與測試 | 356份輸入與最近核准版本相同；正式套件通過；vet保留兩則既有未具名欄位警示 |
| 唯一正常抽樣 | 1046狀態、40PNG，完整trace相同；完整v3只差saved_at |
| 新程序重載 | 13狀態、3PNG，完整trace相同，保存檔未改 |
| AppImage實際視窗 | 十五個設定事件、縮放與鍵鼠通過；三張設定畫面與核准樣張相同；EGA／切HD後住宿v3只差saved_at |
| Release資料匯入 | 全部資料先核SHA-256；錯誤來源在寫入前拒絕；繁中EGA／HD啟動通過 |
| Windows | 實際ZIP的full／匯入後release均可F10切HD及回EGA；十張EGA／HD畫格逐像素相同，中文目視通過 |
| macOS | universal、最低11.0、ARM64 ad-hoc簽章與系統相依通過；沒有執行 |
| 六個封包 | 全部內容、必要素材雜湊、版號、條款、UTF-8、BOM、macOS執行位元及Release deny-list通過 |
| 推廣影片 | 76秒、1920×1080、25fps、H.264／AAC；峰值-9.8dB，十幕字幕與畫面目視通過，無達門檻黑幀 |

影片為截圖剪輯，各幕5至10秒停留；雷聲外的靜音依既有聲音設計保留，不稱連續實機錄影。
完整逐幕來源及技術檢查見 `promo/sources.json`、`ffprobe.json` 與各項偵測log。
六個封包及影片的大小與SHA-256以交付根目錄 `SHA256SUMS.json` 為準。

本版Linux程式SHA-256為 `36506995ad1e1ae11935be7773f6e633dbc8eeb139f07d98c83c5b4e3067c256`；
Windows為 `4291e093aba63f84f5a61e3ee92fac725235de7dfc5a782b90a802db745c3598`。
原版輸入、完整版、推廣影片及生成二進位均未加入Git。

使用者於2026-10-12明確同意將三個引擎包與 `SHA256SUMS.txt` 上傳至
`wicanr2/kol-remake-cht` 並發布 Release。正式入口為
[v.1.0.1-20261012](https://github.com/wicanr2/kol-remake-cht/releases/tag/v.1.0.1-20261012)。
發布後讀回確認 `draft=false`、`prerelease=false`，儲存庫維持 private。
遠端 tag 已取回，指向 `6d98fb0002029ab834f74f3297c6b03a8d9b8f9b`。
原先核准審查與後續授權歷程保存在 [WORKLOG.md](../../WORKLOG.md)。

四個附件的本機 SHA-256 及大小先重新核對，再與 GitHub 回傳的 `digest`、`size`
逐項比對，全部相符；附件狀態均為 `uploaded`，沒有其他附件。

| 附件 | bytes | SHA-256 |
|---|---:|---|
| `kol-v.1.0.1-20261012-release-linux-x86_64.AppImage` | 8874488 | `3115f4bd1f0076deab6a8f2b5d52261947d2396f84d644b031af18abe40574f0` |
| `kol-v.1.0.1-20261012-release-windows-amd64.zip` | 7932941 | `eb8470ec3631acd2a8fcac06739a6fcd7f5cb110761cd3ce4d616935c1c1174a` |
| `kol-v.1.0.1-20261012-release-macos-universal.zip` | 14825233 | `9170c7c7ee52702013a31c927b1afd438c6a9e212c138ea932097696be82dc6f` |
| `SHA256SUMS.txt` | 345 | `021fd51dba261899cde1fa8004d9d0ca6e0bc4454b2a8c3ebb9dff15c0c07235` |

完整版、完整手冊譯文與推廣影片留本機。這次只完成既有產物的上傳及發布，沒有重跑遊戲抽樣。
