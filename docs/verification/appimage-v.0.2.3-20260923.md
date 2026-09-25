# 本機 Linux AppImage 驗收

日期：2026-09-23。版本：`v.0.2.3-20260923`。目標是使用者選定的「本機可玩版」；
本包含原作衍生素材，只留本機，沒有公開散布權。這是 Linux x86_64 候選包，
不宣稱近似完成、完整通關或嚴格原版對拍。

## 交付與封包

- 本機檔案：`dist-all/v.0.2.3-20260923/full/kol-v.0.2.3-20260923-full-linux-x86_64.AppImage`。
- AppImage SHA-256：`3413ec68699eb4ff06044c014bec0c6a720234fcd958d056a8d467b29d9f7a4e`。
- 內含執行檔 SHA-256：`1ea6ea15e2f7bed66a0263f1bbd7497da06af54d128b658ddecf1cb69e83c613`。
- `build-inputs.json` 記錄 1,107 筆來源雜湊；唯讀抽驗封包、逐檔 manifest、
  `SHA256SUMS`、版本、ELF 64-bit／x86-64、入口與 849 筆打包素材／文件雜湊通過。
  `full/` 有 AppImage 與 SHA256SUMS；`release/` 和 `promo/` 只有明示未交付的文字及 manifest。
- 建置使用既有 `kol-package:20260907` 映像、唯讀原始碼與新建空輸出目錄；
  容器用 UID/GID `1000:1000`、`--rm --network none` 與有界資源。

## 正常玩家操作抽樣

- AppImage 的實際入口在 Docker／Xvfb 下執行。建角兩人、重啟選角、逐人住宿、
  刪角與槽位重用、滿槽拒絕、v1 匯入、受控世界時間重載均通過。
  操作輸出和逐檔雜湊在 `workplace/appimage-v.0.2.3-20260923/receipt.json`；
  這個 `workplace/` 收據是本機產物，不是公開交付物。
- 未指定 `-save` 時，`XDG_CONFIG_HOME` 與僅設定 `HOME` 的兩個隔離位置各通過
  三次啟動、95 次正常輸入的建立、住宿保存、重載及刪角流程。收據分別在
  `workplace/appimage-default-save-v.0.2.3-20260923/`、
  `workplace/appimage-home-save-v.0.2.3-20260923/`。重載使用明確 `-load`。
- 從標題、選隊、旅店、城鎮、野外到戰鬥指令列的 GUI 抽樣通過。
  另用原版名冊的六名角色、seed `1`、正常鍵盤腳本進入 WALBAR 遭遇；
  第 26 回合 `verdict=1` 是**敗北**，畫面明寫隊伍遭擊敗。戰後仍能走回
  城 7 半途之家、住宿扣 Poontang 的金冠 `214→158`，寫出 `kol-save-v2`，
  再由同一 AppImage 載入；六名隊員、傷勢、金冠與城鎮位置仍在。
  戰後存檔 SHA-256：`7e42be4e3cc44f17b0d8b12cf9c94d5f640f359ec09b34dbd4534ba63662ce92`。
  完整輸入腳本 SHA-256：`f35fd420fab9bb246a70463d7c32140ca84ea69b159765d280660e44f2b4e0e9`，
  留在 `workplace/appimage-battle-v.0.2.3-20260923/`。
- 單人 Poontang、seed `3` 從城鎮走到世界 `(48,21)` 的第 5 號任務地點，
  正常提示後進入 BINDERAK 戰鬥；這只驗證任務入口與戰鬥開始，不驗證任務完成。

## 判讀與限制

最初把 `verdict=1` 誤讀為勝利，後以戰果文字和 `internal/battle/victory.go`
的列舉值核對並訂正：`1=敗北`、`2=勝利`、`3=逃走`。這場戰鬥沒有冒險點與
金冠獎勵屬敗北結果，不能據此宣稱獎勵回寫缺陷或勝利驗收通過。
單人 Poontang 對 GOBLIN、Moraine 對 GOBLIN 的另兩條探索路徑也敗北；
這些是已知抽樣結果，不代表所有戰鬥無法勝利。

現行封包尚缺正常玩家路徑的勝利、獎勵與戰利品繁中畫格；#90 仍開放。
完整戰鬥同狀態、音訊、圖像／盾牌編輯、加速選項、其他平台真機、全程通關及
公開發行均未驗收。上列單元測試和現代封包抽樣不能當成原版 parity。

重現使用 `tools/build_dist.sh --local-linux-in-container v.0.2.3-20260923`、
`tools/verify_complete_save_gui.sh`、`tools/verify_local_appimage.py`、
`tools/verify_appimage_default_save.py`；正常戰鬥腳本由
`tools/drive_battle_script.py` 在固定 seed 與原版名冊 v1 匯入存檔上產生。
`workplace/` 的探索輸出只留本機。容器和檔案擁有權於每輪收尾另記 `WORKLOG.md`。
