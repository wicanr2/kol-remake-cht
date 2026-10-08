# RE-259：遭遇名稱分列後的長任務與平台回歸

日期：2026-10-06。狀態：現行來源有限回歸通過。

沿[RE-258](20261006-ambush-name-frame.md)正式336Go／338輸入及Linux1f31007c，
使用RE-257既定5971鍵、seed1與200個明示遭遇骰100。途中沒有重播種、狀態代換或新測試捷徑。
本輪沒有修改正式程式、玩法、資料契約、操作、HD素材、版號或發行。

| 驗證 | 結果 | 限制 |
|---|---|---|
| 雙語長任務 | 11944正常狀態與26真實新程序重載通過 | 固定控制遭遇的有限任務鏈 |
| 完整trace | 雙語完整trace與RE-257相同 | 不把顯示欄位差異隱去後稱整份相同 |
| 抽樣整圖 | 40／40張PNG與RE-257逐byte相同 | 差分逐圖另留quest-png-comparison.json |
| 實際v3保存 | 完整保存只排saved_at與RE-257相同，重載未改檔 | 不宣稱原版整條任務parity |
| Windows AMD64 | PE32+、離線建置、音訊依賴與Wine9啟動通過 | 不是Windows真機 |
| macOS | AMD64、ARM64及universal結構、最低11.0與系統相依通過 | 尚無macOS實際執行 |
| Wine正常首場 | 雙語872狀態、16PNG與RE-258現行Linux相同 | seed2、435正常鍵、十四回合，沒有遭遇覆寫 |
| Wine真正保存重載 | 雙語26狀態、6PNG與本輪Linux相同，保存檔未改 | 各自從本輪實際v3檔以新程序重載 |

第4781步、第69回合Fighting Bull取得Ship Wheel；第109回合逃離，第5583步回世界。
六人各回報一次，最終金冠0／0／1721／0／386／0，傷勢3／1／26／1／1／10。
Olear與Fighting Bull各付56住宿，其他四人依正常不足提示略過，保存佇列結束。
Treasure Map與所有物品記錄涵蓋於完整v3比對，沒有只比物品名稱。
跨語言比較只排四個既定顯示欄：encounter_text、exec_text、prompt_lines、spoil_lines。
200明示遭遇骰100沿既定計畫，不當自然完整旅程、全24任務或原版整條任務對拍。

## 執行檔與保存檔身分

| 執行檔 | SHA-256 |
|---|---|
| kol-linux-amd64 | 1f31007c19ceac875a12df9b68d8629296764d01194983cf98d10cc8dcdbd274 |
| kol-macos-amd64 | d1c907b4d581f972870e9aee3ec4931ef255ed21927da1c70cd15dcd04bb1b4a |
| kol-macos-arm64 | 1b7fe5ed05ed4fc35cb3e50192bad33fd1ab568bd62d80a0113f3bcf1a96cadb |
| kol-macos-universal | 80e4873fccc9949d94973955d4dc1cdc6434bacb0109d255dbea1057fab9d7a8 |
| kol-windows-amd64.exe | f6e1710179ed0f2a98484def6614e82e827974f021766777095042cfe1a4f365 |

Linux沿已驗正式執行檔逐byte身分，本輪未另編譯Linux。Windows／macOS由同338來源離線重建。
en實際v3 SHA：6d3aea6d960237dce4471bd996c472d4a28cf9c7068c46925e5beaf81b2016f0。
zh-Hant實際v3 SHA：6457de06b2df90614f253285dbbe72857ef33c9df16ad1c377b31854fa9bfd18。

Go1.26.7、Ebitengine2.9.9與音訊依賴保持鎖版，完整module info保存platform-build-info.json。
三個固定映像與ID保存tool-images.txt：eob-remake-go:1.26.7-ebiten2.9.9-audio、
eob-remake-macos:1.26.7-ebiten2.9.9-audio、coab-wine-smoke:ubuntu-noble-20260826。
Wine正常與重載兩階段的啟動、版本與Xvfb logs分開保留。

## 收據與重現

工作區`workplace/ambush-current-regression-20261006/`。
入口prepare.py固定338來源與既定計畫；run.py／run.sh跑雙語長任務、實際保存及重載。
review.py逐欄核對RE-257，只有已核原版的英文伏擊分列可異，保存只排saved_at。
build-windows.sh、mac-build.sh、check-platform.py建置與驗結構；run-wine.py／run-wine.sh
分normal及reload兩階段比對現行Linux。review-remote.py核對主機gh取得的完整Issue快照。
finalize.py合併Wine收據、核對所有UID後只移除本輪wineprefix-current並更新目前文件。
seal.py封存，verify.py重算來源、附件、完整PNG CRC及解碼、trace／保存比對與擁有權。
重驗時來源唯讀掛/src、此工作區唯讀掛/out，同工作區以可寫別名掛/audit，
設定KOL_RE259_OUTPUT=/audit，在固定Go映像內執行python3 /out/verify.py。
採UID/GID1000:1000、network none、2GiB／2CPU／64程序與120秒；只回寫audit.json。
每批有界Docker，Xvfb就緒與清理trap、Wine服務清理、容器與檔案擁有權另存收據。
本輪圖片與保存資料只留本機，沒有公開散布或正式封包變更。

## 遠端與未完成項

依AGENTS.md GitHub例外在主機先核auth status，再唯讀取得106張worklist Issue與private狀態。
24張OPEN／82張CLOSED，本機18未完成／89完成、14人工保持。
八張本機已有完成證據仍OPEN，建角邊界與HD尚未建Issue，差異保存remote-review.json。
沒有Issue寫入、憑證搬運、提交、推送、版號、封包或發行。

正式338來源未變，十九套件沿RE-258，不再重跑相同Go測試。
本輪只解除新來源長任務與平台有限回歸缺項；Windows真機、macOS執行與正式交付仍待驗。
受阻持物原型DRAFT、完整UI／操作、勝利S結尾GUI、舊H、42別名、全50-byte與其他玩家分支仍未閉合。
HD手繪媒材已核准，解析度、介面與人物造型仍待確認，所有完成條件保持。
