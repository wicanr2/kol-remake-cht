# 本機 Linux AppImage 驗收

日期：2026-09-08。版本：`v.0.2.2-20260908`。本包包含原作衍生素材，僅限本機，未公開發布。

## 交付與結果

- 交付：`dist-all/v.0.2.2-20260908/full/kol-v.0.2.2-20260908-full-linux-x86_64.AppImage`。
- AppImage SHA-256：`112113cc683bd9c4621f9d6643969ef3e475e88a23e531e83e0b7f2c34e6d03f`。
- 內含執行檔 SHA-256：`a3eb7804a8133dc70883260ef6ced54a30c2a8f77b36ad629e67b032e19d61a4`。
- 從 AppImage 本身以 `--appimage-extract-and-run` 執行；98 次正常鍵盤輸入通過建角兩人、重啟選角、逐人住宿、刪角、不搬槽、重用槽位與滿槽拒絕流程。
- v1 匯入保留原檔、零值、傷勢及未知設定；受控世界時間重載後可走路及露營。這不是原版同狀態對拍。
- 解開實際 AppImage 後核對 ELF 64-bit／x86-64、程式 `-version`、849 筆必要素材與文件雜湊及啟動入口，全部通過。
- 住宿報價 60 金冠及 v1 匯入限制提示畫面已抽看。不是全部畫面或字串驗收。

## 重現與來源

既有 image：`kol-package:20260907`；Go 原始碼及必要素材從唯讀 `/src` 複製成乾淨快照，
不要求提交目前工作樹，也不帶入舊執行檔。`build-inputs.json` 記錄 1088 筆輸入雜湊，
在建置結束再次比對來源未變。執行檔以 `-trimpath -buildvcs=false` 及
`-ldflags '-X main.buildVersion=v.0.2.2-20260908'` 建置。

容器入口為 `bash /src/tools/build_dist.sh --local-linux-in-container v.0.2.2-20260908`，
要求全新空 `/out`；專案唯讀掛至 `/src`，新版本目錄寫入 `/out`，既有 Go cache 掛至 `/cache`。
容器使用 `--rm --network none --memory 4g --cpus 3 --pids-limit 512 --user 1000:1000`，
外層 `timeout 240s`。舊版本未覆蓋。AppImage 工具及 runtime 雜湊記於建置輸入收據。

驗收使用 `tools/verify_complete_save_gui.sh`，將 `/out/kol` 設為呼叫實際 AppImage 的啟動器；
存檔及圖片只寫入專用 `/out`。其後由 `tools/verify_local_appimage.py` 核對內容與操作結果。
完整紀錄、擷取畫面及逐檔收據位於 `workplace/appimage-v.0.2.2-20260908/`，
每段輸入次數為 28、32、35、0、3；種子固定為 1，只驗 remake 保存流程。

## 尚未完成

本輪未驗主機 FUSE 掛載、Windows／macOS 真機、全遊戲通關或完整 dosgolem 戰鬥同狀態。
`release/` 與 `promo/` 明示尚未交付；音訊候選未接正式播放。
繁中資料頁、圖像／盾牌編輯及加速開關仍待完成，不因新包而改標完成。
初次保存 GUI 明確指定存檔路徑；預設位置的後續抽驗見下節。

## 未指定存檔路徑的補驗

同日以 `tools/verify_appimage_default_save.py` 再從 AppImage 完成三次獨立啟動，
全程不提供 `-save`：建立 NOVA／LUNA → 明確 `-load` 重啟並逐人住宿 →
再次重啟、刪除 NOVA、以 NOVA2 重用空槽。共 95 次正常輸入（28、32、35）。

設定 `XDG_CONFIG_HOME=/out/config` 時，完整 v2 存檔寫至
`/out/config/kol-remake-cht/save.dat`；重啟仍保有 LUNA 的住宿後金額 1440，
新角色 NOVA2 使用新身分，未寫入舊測試指定的 `/out/created.json`。
AppImage 前後雜湊不變，輸出擁有權均為 1000:1000。
收據及實際展開的操作腳本保存在 `workplace/appimage-default-save-20260908/`。
這驗證使用者設定目錄的持久保存，不宣稱自動讀檔；重載仍明確傳入 `-load`。

第二組移除 `XDG_CONFIG_HOME`，以隔離的 `HOME=/out/home` 重跑相同 95 次正常輸入，
亦全部通過，存檔確實落在 `/out/home/.config/kol-remake-cht/save.dat`。
收據位於 `workplace/appimage-home-save-20260908/`；兩組使用相同 AppImage，前後雜湊均不變。
至此 Linux 兩種預設位置均完成實際正常保存／重載抽驗；不延伸為其他平台驗收。
