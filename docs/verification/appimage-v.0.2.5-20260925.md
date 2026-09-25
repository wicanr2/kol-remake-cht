# 本機 Linux AppImage 驗收

日期：2026-09-25。版本：`v.0.2.5-20260925`。本包含原作衍生素材，只留本機，
沒有公開散布權。它是 Linux x86_64 本機候選包，在 `v.0.2.4-20260925` 之上再收進
漏譯掃描後的修正：名冊所在地、教派名、戰鬥物品名與冠詞、戰鬥執行訊息可讀性、
法術名注解（SPEC-034）與法術編輯器中文頁（SPEC-035）；不宣稱全遊戲繁中、手冊功能與原版一致、完整通關或嚴格原版對拍。

## 交付與封包

- 本機檔案：`dist-all/v.0.2.5-20260925/full/kol-v.0.2.5-20260925-full-linux-x86_64.AppImage`。
- AppImage SHA-256：`8bcb7814ebce17a4b52e85fedbe386c8f1278656d23719e2c857cdeb558fc27f`。
- 內含執行檔 SHA-256：`05608cf69036a8f43682e3eb8f061b71002cc8d5de5449e8206a8e438663002b`。
- 來源為 commit `c62bef2` 的乾淨工作樹。`build-inputs.json` 記錄 1,119 筆輸入雜湊；
  封包核對 851 筆打包素材／文件雜湊、版本、入口與禁止副檔名，`SHA256SUMS` 相符。
  `release/`、`promo/` 只有明示未交付的文字及 manifest。
- 建置使用既有 `kol-package:20260907` 映像、唯讀原始碼與新建空輸出目錄；
  容器用 UID/GID `1000:1000`、`--rm --network none` 與有界資源。

## 正常玩家操作抽樣

- 實際 AppImage 以 `--appimage-extract-and-run` 在 Docker／Xvfb 下執行，固定 seed `1`。
  建角兩人、重啟選角、逐人住宿、刪角與槽位重用、滿槽拒絕、v1 匯入、受控世界時間
  重載均通過；各段正常輸入 28、32、35、0、3 次。雙擊啟動未指定語言時介面為繁中，
  野外日期顯示「二月二十八日／西元 1102」。收據在
  `workplace/appimage-v.0.2.5-20260925/receipt.json`。
- 未指定 `-save` 時，`XDG_CONFIG_HOME` 與僅設定 `HOME` 的兩個隔離位置各通過
  建角、重啟 `-load`、住宿、刪角及空槽重用。收據在
  `workplace/appimage-default-save-v.0.2.5-20260925/`、
  `workplace/appimage-home-save-v.0.2.5-20260925/`。

`workplace/` 的收據是本機產物，不是公開交付物。

## 限制

本包沒有重跑野外戰鬥、勝利獎勵或任務鏈的封包內正常路徑；這些仍以重製版
觀察資料收據（`docs/verification/cjk-ui/`）為局部證據。45 個城鎮來源槽位仍缺譯（#92），
完整戰鬥同狀態、音訊、圖像／盾牌編輯、加速選項、其他平台真機及公開發行均未驗收。

重現使用 `tools/build_dist.sh --local-linux-in-container v.0.2.5-20260925`、
`tools/verify_complete_save_gui.sh`（stdout 另存 `/out/verification.log`）、
`tools/verify_local_appimage.py`、`tools/verify_appimage_default_save.py`。
