# 《武士傳說》重製目前狀態

更新日期：2026-09-10

## 目前狀態（2026-09-10，以本節為準）

已依使用者指示恢復。自評 **75／100**，不是完成百分比或原版一致率。

<!-- worklist:stats:begin 由 tools/worklist.py render 產生，不要手改 -->
- **未完成項的權威是 `docs/worklist.json`**，不是 `WORKLIST.md` 的 checkbox。
  核實用 `tools/run_python_tool.sh tools/worklist.py verify`；`WORKLIST.md` 的
  那一節與這幾行數字都由 `render` 產生。
- 目前 **71 條未完成**（33 條沒有機器訊號、標 `manual`），
  另有 16 條已核實完成——後者跑的是回歸檢查，訊號不見了就開口。
  未完成項的 verify 種類：`present` 35、`manual` 33、`absent` 2、`required_in_all` 1。
<!-- worklist:stats:end -->
- **dosgolem 對拍工作區是 `~/cht/dosgolem-kolrand` 的 `kol-oracle` 分支**
  （base 為 `origin/main`，`5f9a9a1`，已推送到 `origin/kol-oracle`）。舊的 `kol-force-call` 已刪除——它的 commit
  全在 `origin/main` 裡，工作樹上的 12 個未提交檔案已移植過來。KOL 靠的三份規格在
  dosgolem 的號碼是 **194**（EGA BIOS 預設色盤）、**195**（DOS 直接讀鍵）、
  **196**（BIOS 讀鍵橋接）。
- 新 base 上 `KOL_CHECK=all` 全部通過；書本入口要帶 `KOL_BOOK_SELECT_DOWN=12`
  才走得到 KOLTOWN（預設 0 會選到空槽，那是測試參數不是回歸）。

- 已完成並有局部對拍證據：怪物遠程初始化與肖像索引分離、地形生成與灑點轉置、
  戰後傷勢折算、任務物品完整記錄比較及獎勵隔離。完整 Go 回歸與建置已通過。
- 同狀態基礎：兩名隊員與 SLEDGE 共 60 個初始欄位通過；地形三種固定種子、
  兩條路徑共 24,576 格及亂數終態與 dosgolem 一致。
- 最後原版正常遇敵流程 84.44 秒通過，實際地形收據在
  `workplace/battle-terrain-normal-fixed-20260908/battle-terrain.json`；尚未接入 remake 完整比較。
- 現行本機交付仍為 `dist-all/v.0.2.2-20260908/full/`；AppImage **不含最新怪物及地形修正**。
  程式異動及研究收據留在工作樹，未提交或發布新版。
- 未完成：完整戰鬥指令／結算與正常戰後保存、競技場完整流程、繁中資料頁、
  圖像／盾牌編輯、正式音訊、加速選項及其他平台／推廣交付。詳細依 WORKLIST 各閘門。

下一步接上新地形快照，完成同狀態指令／結算及正常戰後保存，再將累積修正重建
為新版 AppImage 並驗收。繁中資料頁版面及其他未決選項仍待確認，不能自行視為同意。
下方逐輪紀錄供追溯，不得覆蓋本節現況。

最新修正：怪物面板已將肖像索引與遠程能力值分離，完整測試及建置通過。
SPEC-029 能力初始化已接入首次規劃，32 種怪物資料已重生；正式實作 768＋2304 組
原版比較、一次性初始化及 AI 分支回歸、完整測試與建置通過。
見 [能力初始化收據](docs/verification/monster-ranged-initialization-20260908.md)。
現行 AppImage 尚未包含本次能力及面板修正，完整正常戰鬥仍待驗。

目前唯一自評與缺口入口：[RE／SPEC／實作稽核](docs/verification/spec-source-audit-20260908.md)。
玩家可用性與交付準備度自評 **75／100**，不是完成百分比或原版一致率。
最新交付補驗：Linux AppImage 的 XDG_CONFIG_HOME 與 HOME/.config 兩種預設保存位置，
各以 95 次正常輸入跨三次啟動通過；封包交付由 2／5 提高至 3／5，其他分項不變。
主要未完成：繁中資料頁、圖像／盾牌編輯、
音訊與已決加速選項；各項 READY 範圍不同，不得一概當作已核准而漏做。
三個加速選項已整理為 SPEC-028 DRAFT，核對既有規劃、推進、播放器與保存接口；
尚未以未決細節加入正式開關。繁中資料頁版面問題已再次呈現，仍等待使用者選擇。
完整保存已實作；新版 Linux AppImage `v.0.2.2-20260908` 已從封包通過 98 次正常輸入及保存驗收，
見 [交付收據](docs/verification/appimage-v.0.2.2-20260908.md)。完整戰鬥同狀態、其他平台與推廣交付仍待完成。
正常入口補驗見 [隊員數值收據](docs/verification/battle-party-state-20260908.md)：
Olear／Theodoric 共 40 個初始欄位相符，原版正常 SLEDGE 等級為 3；完整狀態仍待驗。
最新原版補證：SPEC-027 已修共同出口傷勢折半及敗北角色回寫，4608 組
dosgolem 比較、完整回歸及五條結算至 v2 重載通過。正常 GUI／完整戰鬥、
額外掉落與敗北返回仍待驗，不因局部計算正確而升為完整對拍。
見 [共同出口收據](docs/verification/battle-exit-wounds-20260908.md)。
最新任務修正：拾取提示及攜物逃走改比完整 16-byte 記錄，避免同名誤認。
258 組 dosgolem 比較、完整回歸及建置通過；正常取得、逃走與回報鏈仍待驗，
見 [任務物品收據](docs/verification/quest-item-record-20260908.md)。
同日補驗全部 24 個實際任務的資料載入、拾取、攜物逃走與完成旗標，
確認完整記錄及其他任務旗標保留；位置受控，正常 GUI 完整鏈仍待驗。
最新回報修正：獎勵物品不再與建築母表及其他角色共用實例；四種交付位置、
完整回歸及建置通過，見 [獎勵隔離收據](docs/verification/quest-reward-isolation-20260908.md)。
下方較早的「尚未」敘述只保留交接脈絡，與本表衝突時以本表及最新收據為準。

最新補缺：SPEC-024 告別已實作並驗單人直接離店、多人等待確認；SPEC-025 已將
競技場改接原版第 25 張固定地圖，4096 格 dosgolem 比較與完整測試／建置通過。
競技場受控資格進場已抽驗，正常取得資格與完整結算鏈仍待驗。
SPEC-026 天氣已修正合成、雨雪輪播、換組及正常節拍；
dosgolem 四組合成與六組分派通過，正常出城 GUI 2928 個變動像素全在天空，其他區域零差。
繁中資料頁的分區原型已展示，版面選擇待答；不因未回覆就視為核准。
本輪中間產物在 `workplace/player-gaps-20260908/`，AppImage 尚未更新。
完整範圍與失敗勘誤見 [補缺收據](docs/verification/player-gaps-20260908.md)。

後續補修：競技場對話原先固定使用英文繪製器，現已依語言切入既有 `InnCJK`；
完整回歸及建置通過。首次 GUI 揭露繁中空肖像讀取錯誤，現已修正並乾淨重跑，
九張畫面、選人及不下注進場通過；收據見
[競技場對話修正](docs/verification/arena-dialog-fix-20260908.md)。
最新執行檔與雜湊在 `workplace/arena-gui-fixed-20260908/`，不是 AppImage。
前次建置另存
`workplace/player-gaps-20260908/arena-cjk/`，不可把前輪 `summary.json` 的 GUI
執行檔雜湊套用於本版。SPEC-005、012、015 已訂正部分過時限制，D01 尚未全數完成。

最新保存進度：正式遊戲已使用 v2，完成 v1 匯入、建角回填、逐人住宿交易與
世界時間保存／還原。完整測試及建置通過，正常 GUI 建角、重啟選角、忠實住宿、
刪除及重用槽位通過；受控世界時間由 1102 年 2 月 28 日推進到 3 月 1 日。
v1 原檔保留，未保存變更的限制會在畫面提示；零值與空裝備不再被原始名冊補回。
詳見 `docs/verification/complete-save-integration-20260908.md`。AppImage 尚未重打包，
本輪不宣稱原版戰鬥同狀態完成。

最新對齊結果：以 dosgolem 正常戰鬥快照建立 remake 受控情境，兩名隊員與一名敵人
的角色／站位、目前角色 Theodoric、鏡頭中心 (31,30) 與地圖標記通過；共同規劃點
seed=1。原版正常路徑測試及 remake 完整戰鬥模組回歸通過。尚未驗地形、完整屬性、
規劃順序或攻擊結算，不宣稱完整戰鬥同狀態。見 `docs/verification/battle-actor-alignment-20260908.md`。

最新決定：使用者核准 v2 支援匯入 v1；不再等待相容性選擇。保留舊檔並明示舊格式
未保存的資料無法還原。SPEC-023 已完成審查並接入正式保存流程。

最新原版進度：dosgolem 正常選 Olear、招募 Theodoric、確認離店告別、出城、露營，
第二天遭 SLEDGE 伏擊並進入戰鬥規劃。原版 seed=1 固定於旅店招呼後；91.08 秒測試
通過，49 個輸出來源未變。原版先手 Theodoric，既有 remake 收據先手 Olear，尚未完成
角色／站位／seed 設定時點對齊，不宣稱戰鬥同狀態完成。見
`docs/verification/battle-normal-seed-20260908.md`。
完整保存欄位與交易邊界依 SPEC-023 READY 實作；原版離店告別訊息仍另列
玩家可見差異，待規格審查。

最新驗證：目前工作樹完整測試／建置通過，固定 seed=1 的 150 個正常 GUI 輸入
確認 Olear 射擊彈藥 20→19；dosgolem 另由正常入口選 Olear 進旅店，角色、肖像、
招呼語與六圖示順序相符。原版該腳本未設定遊戲 seed，不宣稱戰鬥同狀態完成。
見 `docs/verification/fixed-seed-player-path-20260908.md`；AppImage 仍未更新。

現行存檔設定讀取已補 SPEC-006 型別契約：拒絕 `settings: null` 及非布林住宿設定，
成功解析時整份替換，避免沿用上一份設定；未知鍵仍保留。存檔模組回歸通過。
完整角色／名冊保存已改為 v2 並支援 v1 匯入；封包尚未重打包。

本輪使用者決定：停止追查 RND() 的重擲次數與精確骰序，不再把這類差異當成
收尾阻塞。保留已驗證的修正與歷史證據，後續優先補功能缺漏、正常玩家路徑及
AppImage 驗收；滑鼠替代鍵盤仍以操作結果等價為準。
後續明確化：涉及 RND() 的測試須由 dosgolem 固定原版 seed，remake 也固定測試 seed
做可重播對拍；不要求未受控的真實 RND() 執行逐次一致。完整鐵則見 AGENTS.md。

反制搜尋起點後續已核對並修正，14 組原版與對應回歸通過，完整測試及建置通過。
收據 `workplace/counter-scan-20260908/oracle/manifest.json` 記錄來源未變；
最新中間執行檔為 `workplace/counter-scan-20260908/build/kol`，SHA-256：
`e853ceba58bf7ac35c9bec5a495ae4d4467d43929d48f73bf3ed952ece8f67cd`。
這不是新版 AppImage，也不代表完整 AI 或正常玩家流程已驗收。

敵人同環尋敵已依 SPEC-013 READY 修復：先遇到的隊員被牆擋住時繼續同圈搜尋，
不因續查重擲。後續 dosgolem 七環表驗證推翻舊索引偏移說法，已修搜尋骰面為
格數減一、敵人射擊門檻為 4；完整測試／建置通過。counterScan 獨立起點後續已核對，
正常戰鬥對拍仍待完成，見 `docs/verification/ring-search-20260908.md` 與上方最新狀態。

最新修復：戰鬥拾取／丟棄／擊落／滿背包戰利品交換漏同步 Member.Weapon，導致
攻擊與選單／彈藥資訊使用不同武器數值。四路已補，撿弓後射擊指令及資訊框回歸、
完整測試與建置通過；正常 GUI 異種武器拾取仍待驗收，見 `docs/verification/weapon-sync-20260908.md`。

最新修復：隊員射擊原本未扣彈藥，已依 SPEC-022 READY 接入共同收尾，含射空與撞牆；
裝填不扣、最後一發後停止。完整測試、建置與 2,550 組 dosgolem 直接比較通過。
彈藥資訊框後續已接線，128 組原版顯示比較與六組文字流程通過；正常 GUI 150 個輸入
確認 Olear 射擊後由 ARROWS:20 更新為 ARROWS:19。弩與耗盡的正常 GUI、原版整場
同狀態及 AppImage 仍待驗收，見 `docs/verification/ammunition-panel-20260908.md`。

最新新版 dosgolem 整合重跑：`KOL_CHECK=all` 全部通過，包含完整面板與規劃檢定。
來源未變、19 個輸出雜湊與擁有權均核對通過，容器已清理；詳見
`docs/verification/dosgolem-current-integration-20260908.md`。AppImage 尚未重打包，
完整操作等價仍未驗收。

SPEC-021 已接左下命令、招式、部位、防禦及方向面板，完整測試、建置及 3,168 組
原版參數比較通過。正常 GUI 148 個輸入已顯示隊友 OPT／休息圖示，取消整張恢復零差異；
其他圖示及原版同狀態尚待驗證，見 `docs/verification/command-panel-implementation-20260908.md`。

SPEC-020 已接規劃位置一次性門檻及空槽／倒下人物的擲骰順序；完整測試、建置與
生命週期回歸通過，仍為 READY。正常 GUI 126 個輸入與面板切回零差異已驗證；
完整戰鬥仍待完成；依本輪決定不再要求精確骰序作為收尾閘門，見
`docs/verification/planning-roll-implementation-20260908.md`。

最新修復：NewQuest 遺漏兩方初始智力，已補上 Member／Monster 到 Actor 的轉接。
初始及法術恢復後的判定一致性、完整測試與建置通過；見
`docs/verification/intellect-initialization-20260908.md`。SPEC-020 與命令面板目前狀態以上方摘要為準；
AppImage 尚未更新，不以此局部修正宣稱整體完成。

新版 dosgolem 整合入口已補納護甲／耐久與資訊框直條測試；`all` 乾淨重跑全部通過，
正常選角進 Halfway House 圖片已檢視。輸入未變、19 個輸出雜湊及擁有權核對通過，
KOL 一次性容器已清理。見 `docs/verification/dosgolem-complete-suite-20260908.md`。
這只更新工具鏈整合驗證，不改變下列功能缺口與 AppImage 尚未更新的狀態。

正常戰鬥入口新增查核：實際視窗從標題選 Theodoric、招募 Poontang、離店及出城，
步行後進入戰場。原版解析度戰場直接畫在繁中畫布上造成左上四分之一顯示，已補共用
離屏縮放出口並通過完整測試／建置。正常 GUI 重跑通過，123 個輸入一致，兩張戰場
畫格依預期縮放後像素差異為零；見 `docs/verification/normal-battle-entry-20260908.md`。
當時確認左側裝備與身體框沒有內容，列 A10；不能以戰場可啟動宣稱資訊完整。
後續已由 IDA 定位左上顯示模式及六個裝備位移，左下直條讀疲勞而非生命值；
dosgolem 十二組直條與人形參數測試通過。後續人形遮罩 144 組通過，確認六部位矩形、
截斷及色號 7／8／15 轉 12；見 `docs/re/20260908-body-overlay.md`。
後續底圖與顯示角色已補證，SPEC-018 經審查為 READY；正式人形、傷勢遮罩、疲勞條
及被檢視者狀態已接通。156 組 remake／原版參數、底圖路由、完整測試與建置通過。
正常 GUI 戰場已顯示人形，裁切與來源 PNG 放大後像素差異為零。
再以 127 個正常輸入驗證選目標清框與取消恢復，全畫面恢復後像素差異零。
當時左上模式／裝備／肖像及左下命令圖示仍未完成；SPEC-018 仍未 CONFORMED，
見 `docs/verification/body-panel-implementation-20260908.md`。

左上框後續已依 SPEC-019 READY 接通基本肖像、六列隊員／五列敵人裝備與空白鍵切換。
新版 dosgolem 原版模式、空格、裝備列及肖像位置測試通過，完整 remake 測試及建置通過。
正常 GUI 的 126 個輸入完成，肖像與裝備切回後各自整張像素差異零；最新執行檔及收據見
`docs/verification/equipment-panel-implementation-20260908.md`。28 筆 IDA 分級註記已重生核對。
彈藥標題／數字及正常原版戰鬥同狀態驗證仍待完成，不宣稱 A10 或整體對拍完成。
左下命令已依 SPEC-021 接線，3,168 組原版參數、文字與 PNG 路由及九組真實敵人
智力檢定比較通過；正常 GUI 已驗隊友休息圖示。詳見
`docs/verification/command-panel-implementation-20260908.md`。

護甲查核最新狀態：既有 IDA 資料庫證實攻擊常式會經護甲與耐久處理後呼叫扣傷，
已訂正 RE-133 的相反結論。新版 dosgolem 加值全 256 種輸入與九組耐久邊界通過；
這不代表完整攻擊已對拍。後續已確認 18 個記錄指標與 84 組護甲後段行為，
SPEC-017 已審查升為 READY，近戰／射擊與勝利／撤退回寫已接通；84 組 remake／原版後段
比較及完整測試、建置通過。正常 GUI 戰鬥、完整保存與其餘磨損呼叫仍待完成，
見 [護甲控制流訂正](docs/re/20260908-armour-control-flow.md)。

後續修復：開場選角已依 SPEC-016 接到 `town.Session` 的正式建築流程，驗證城鎮代碼、
建築編號與所在地後才進旅店；跨城載入已接通，正常 GUI 選 Mula、招募 ddduhh 及離店通過。
詳見 [跨城驗證](docs/verification/cross-town-fix-20260908.md)。繁中候選人清單漏畫已補上。
同日 dosgolem 正常選 Mula 重跑通過，原版／remake 的角色、旅店、肖像、招呼語及
六個圖示順序相符；只涵蓋跨城入口可見狀態，不代表招募、住宿或完整存讀檔已對拍。
招募滑鼠接線另補候選人列命中範圍：點名字以既有上下移動及確認指令選取該列，
不再退回確認目前反白者。正常繁中 GUI 已驗證返回圖示下點第二列 Olear、正確加入及離店；
英文／捲動清單的 GUI 與封包對拍仍待驗證，詳見開場旅店驗證紀錄的後續段落。
同時修正選隊／招募共用名冊可變資料的缺陷：兩條入口共用獨立隊員複製，涵蓋技能、
裝備、任務、原始物品記錄及法術防護陣列；招募後同步戶外與室內隊伍。
這些修正不更改保存格式，也不代表中文資料頁、完整保存或 AppImage 對拍已完成。

工具鏈最新狀態：依使用者授權將 `dosgolem-kolrand/kol-force-call` 快轉至
`fd746405`，並完成 KOL 新版診斷介面適配；dosgolem 本機補回檔案軌跡失敗與定位基準欄位。
核心測試與六項 KOL 規則對照通過。後續已依 BIOS 契約修正 RGB 全黑，六張片頭擷取逐張確認，
並由正常 `KOL.EXE` 入口以 DOS Enter 進入書本封面；**完整 AppImage 對拍仍未驗收**。
最新工具鏈證據見 [色盤修復紀錄](docs/verification/dosgolem-palette-fix-20260908.md)，
先前全黑失敗鏈保留於 [整合紀錄](docs/verification/dosgolem-integration-20260908.md)。
同輪另修 DOS AH06 永遠回無字元的缺陷。後續 IDA 證實 KOLBOOK 先查 BDA 再用 DOS AH08 取鍵，
已補上跨介面取鍵；核心、六項規則與正常書本入口測試通過，擷取確認已進入 Halfway House。
最新證據見 [書本讀鍵修復](docs/re/20260908-kolbook-input-bridge.md)。
下一步回到 remake 開場旅店等未完成接線，再以此原版正常入口對拍。
兩個儲存庫的本輪修正尚未提交／推送；不影響下列遊戲修復與未完成項目。

最新進度：已依使用者「先修正再檢驗」實作第一批修復；目前真相以
[修復與再驗證紀錄](docs/verification/spec-fixes-20260907.md) 為準。
完整既有測試與七項新增回歸通過；正常 GUI 驗證建角輸入／取消與第二名角色入隊顯示；
dosgolem 六項具名規則對照通過。保存格式 v1／v2 決定待回覆，未更改 schema；
完整建角保存與中文資料頁仍未完成。開場旅店已接通同城／跨城選隊、招募及離店，
正常 GUI 確認，見 [本輪驗證](docs/verification/starting-inn-fix-20260908.md)。現行 AppImage 尚未更新。
以下九個失敗為修復前稽核基準，不能反向覆蓋這份最新進度。

**現況：核心玩家流程仍有阻塞缺口。** 2026-09-07 原始碼／規格稽核重現九個失敗，
包含建角、選隊、戰後回寫、保存、腰帶與法術目標類別。先前 90/100 是歷史自評，
不再作為目前完成判定。唯一目前狀態與證據見
[規格與程式稽核](docs/verification/spec-source-audit-20260907.md)。

## 目前真相

- 本輪使用者確認：滑鼠可取代鍵盤，只要操作結果等價於原版即可；不要求裝置或
  按鍵序列相同。此後優先查正式資料與功能接線，前輪裝置差異不單獨作為否決理由。
- 本輪實測：既有測試全過，但九個跨層稽核測試全失敗。修復入口與完成條件以
  `WORKLIST.md` 頂部及上述稽核 A01…A09 為準；沒有更改正式程式或 spec 授權狀態。
  下列較早「做完」的描述只保留其局部程式／收據意義，不能覆蓋本輪找出的垂直鏈缺口。

- 2026-09-07 實測既有 `0.2.1` 完整版 AppImage 與 dosgolem：基本鍵鼠流程通過，
  **操作感一致尚未通過**。地圖點擊送確認而非移動；dosgolem 正常啟動能進入 KOLBOOK
  載入階段，但未走到城鎮。城鎮點擊證據目前僅有初始化輔助實驗，長按節奏、戰鬥與
  存讀檔未完成正常同狀態對照。見 `docs/verification/appimage-dosgolem-audit-20260907.md`。

- 本專案是 *Knights of Legend* 的繁體中文乾淨重寫，正式 runtime 使用 Go／Ebitengine、
  PNG 圖像與 Ogg Vorbis 音訊。
- GitHub `wicanr2/kol-remake-cht` 已建立為 private。原版 ZIP／EXE／archive／PDF 與其他 raw
  素材不得提交；使用者已於 2026-09-01 特別允許已驗證、可重生且有 provenance manifest
  的 PNG 衍生物入庫：抽取器產生的素材 PNG 放 `assets/original-derived-png/`，模擬器
  擷取的原版執行畫面 PNG 與其收據同目錄放 `docs/verification/`。這不授權 GitHub Release
  或公開散布。
- Go／Ebitengine 垂直切片已啟動：`cmd/kol` 以 Go 1.26.7＋Ebitengine v2.9.9 建於固定
  映像 `eob-remake-go:1.26.7-ebiten2.9.9-audio`，只讀 `assets/original-derived-png/ega/`
  的 manifest 與 PNG，畫在 320×200 原版解析度離屏畫布上，並提供截圖模式供對拍。
  書本的六個畫面（封面、目錄頁、角色建立、角色清單、刪除、HARBOR MASTER）、兩段
  翻頁動畫、TITLE 標題序列與城鎮的旅店、戶外地圖（含地形擋路、門口與地圖邊界的
  城門提示、進建築、招呼語與 LISTEN／MOUTH 對話、招募同伴、客棧住宿與修道院治療、
  MIRROR 的角色資料頁、商店買賣、AWARD 訓練、存檔與讀檔）
  與野外地圖畫面（含 ROAD／SIGNPOST 自動循路、進城、時刻與日曆推進）都已實作並對拍。
  任務地點的戰鬥走得完一輪：戰場的 9×9 視窗、逐格拉線的視線、參戰者疊圖與白框，
  指令列、選目標、四列子選單、兩段回合順序、移動、近戰結算與逐句訊息都實作了，
  實跑打得死怪物（RE-079…090／SPEC-008、009、010）。
  **一場戰鬥雙向打得完了**（RE-091／RE-092）：敵人會照原版的 AI 接近、挑目標、
  選攻擊式與命中部位；每個回合結束數兩邊倒下的人，依四個條件判勝負，
  打輸與逃走各有自己的訊息。
  **戰後結算做完了**（RE-096）：冒險點數是每一個參戰敵人的攻防技能加總除以 2
  再除人數，金冠只讀第一隻怪的骰式，掉落物與戰鬥中丟在地上的東西一起排成
  「Items Found:」清單讓玩家一件一件領走。整條路徑實跑過
  （`docs/verification/battle-spoils/`，用了觀察用的替代輸入）。
  **野外的隨機遭遇打得起來了**（SPEC-011／RE-097…099／RE-112）：時段一變就擲一次，
  遇到就進戰場——機率、挑怪、隻數、等級、騎走、兩邊站位與逐格擲出來的地形
  都對得上原版。**全隊都有坐騎時會先問「要不要騎走」**：那一頁印的是那一隻記錄
  末端的四行敘述（`0xBB26` 就是記錄的 `+0x1C0`），答是就擲 `rand(100) > 5 × 移動率`。
  實跑收據 `docs/verification/battle-ambush/` 與 `battle-rideaway/`。
  **施法、射擊與逃走也做完了**（SPEC-012／013／014）：法術從角色記錄讀、
  兩張清單走原版的版面；射擊拉一條 Bresenham 彈道，樹扣命中值、牆停下、半路的人
  各扣體型的一半，敵人在近距離也會射；每個回合結束若有人帶著任務要的東西站到戰場
  最外圈就問「要逃嗎」，答應就走——**帶著東西逃掉一樣記任務完成**（RE-101…104）。
  **任務那一整條接完了**（RE-067／101／106…111／SPEC-015）：走進給任務的建築、
  按 MOUTH 打字問到那一戶的關鍵字就接得下來；打贏任務戰鬥標成完成；
  走進收任務物品的建築就把身上那件交出去、旗標變成「回報過」、已回報數加一；
  角色資料頁上的第 19 個圖示列出回報過的那些。**結局的觸發點是任務戰鬥打完那一刻**
  ——這一場真的標到完成、而且有人的已回報數等於 0x17，畫面就換成結局的七段。
  **敵人傷得夠重會撤退了**（RE-081）：兩個門檻來自怪物記錄的 `+0x0D`／`+0x0E`
  ——傷勢達到前者就考慮撤退，撤退中降到後者以下就回到正常判斷。旁邊有隊員貼著
  就往外逃，八個方向都沒人就原地喘息。SKELETON、GOLEM、ZOMBIE 這些是 100／0，
  傷勢到不了 100，所以**不死系與元素系永遠不逃**。同一批修掉了兩個「這一手」的
  計數從來沒有歸零的問題（挨傷計數與招架次數，`0x20E64`），並補上貼近打那一手的
  招架姿態——`+0x24` 填的是「命中部位 ＋1」，而同一格在這一名被攻擊時就是防禦選項，
  所以先前的敵人被打時不會招架。**聰明的怪物還會反制掃描與繞側翼**
  （`sub_253E7`／`sub_25653`，RE-104）：智力擲骰過關就繞著自己那一環算優先度，
  把貼著的隊員的攻防選項抄過來當成自己這一手；對方要移動時另外繞到它的側面。
  打贏那一場的冒險點數
  與金冠也開始寫回角色了，而**競技場那一場整段獎勵照原版跳過**（`0x2124E` 的閘門，
  RE-096）——它給的是位階，不是點數與錢。冒險點數跟著進了存檔。
  **競技場整條走得到了**（RE-108）：世界地圖 (55,13) 那一格選 DOOR 進場，
  夠格的條件是帳面位階與實得位階不同、而且帳面還沒到 24；多人隊伍選好人、確認之後
  先換到賭場，沒上場的人可以押那一位會贏還是會輸（金額兩位數、上限 99，
  負值就是賭他會輸），答否才開打。對手一隻，等級是上場那一名的實得位階加一；
  打贏帳面位階加一——**一次只加一級**——賭金接在升階那一句後面結算。
  **場上只有上場那一名**：其餘隊員排到觀眾席、記錄 `+0x01` 設成 2 不參戰，
  上場那一名與唯一那隻對手各站一格固定的（RE-098 的第三條分支）。
  實跑收據 `arena`（1,770 步）。**它是升級唯一的入口。**
  **一條按鍵走得到底的正常玩家路徑有收據了**：`full-chain`（567 步）從標題走到住宿
  存檔，中間打贏一場野外遭遇；`full-chain-reload` 讀回來走到任務地點；
  `full-chain-ending` 走到結局。四條都在 `docs/verification/player-path/`。
  任務旗標、已回報數與身上帶的東西已經進存檔，重載之後進度還在。
  SPEC-001／002／003／005／006／007／008／009／010／011／012／013／014／015 為 READY
  或 CONFORMED（008、009、010 只有已解出的段落 READY），SPEC-004（音訊）仍為 DRAFT。
- **開起來就能玩**：`kol` 不帶旗標進的是互動遊玩（`-scene game`，新的預設值），
  從標題序列開始，鍵盤與滑鼠直接走到書本、建角、選角、旅店、城鎮、野外與戰鬥。
  它借的是 `flow` 那條路徑的全部初始化，分派與 `-flow-script` 完全相同——人在
  城裡是 `flow.Key`，走出城門交給 `playToken`——所以互動遊玩經過的是二十六份
  對拍收據驗過的同一組狀態機，不是另一條平行的輸入處理。
  滑鼠翻成既有的鍵盤 token（點選單第三列送出 `D D RET`），可點區域取自各畫面
  既有的版面常數；右鍵是 Escape，滾輪是上下。介面預設繁體中文，`-lang en`
  切回原版字模的英文對拍版面。住宿存檔預設寫到 `os.UserConfigDir` 底下。
  驗收是 `tools/verify_interactive.sh`：Xvfb 裡送鍵與滑鼠、截圖比雜湊，判準是
  **滑鼠點第 N 列必須與鍵盤走到第 N 列產生同一張畫面**，一路驗到選角、進旅店、
  滑鼠點 DOOR 出門、在城鎮裡走一步。
  尚未接上的是開場旅店那一格的其餘五個圖示（MIRROR／LISTEN／TALK／REST／
  COMPANION，城鎮裡的建築內部功能是齊的），以及戰鬥與野外的可點區域。
- **三個平台各自獨立封裝**：`tools/package_dist.sh` 產 `dist-all/<版本>/packages/`
  六個發行檔——`full`／`release` 各有 Linux 的 AppImage、Windows 的 zip 與
  含 `武士傳說.app` 的 macOS zip。`.app` 由 Finder 啟動時工作目錄是 `/`、
  AppImage 掛在暫時目錄，所以程式在工作目錄找不到資源時會換到執行檔旁邊
  （含 `.app` 的 `../Resources`）。`tools/verify_packages.py` 驗六個檔的結構
  （執行位元、`.app` 的 Info.plist、full 要有素材、release 不得有原作衍生位元組），
  並已做過注入缺陷的負對照。`0.2.1` 全部通過，**Windows 與 macOS 仍然沒有真機驗收**。
- 原版內嵌點陣字型已抽出：`BOOKDATA.0`／`KOLDATA.0` 的第一個 payload 是同一份
  1,303-byte 區塊（SHA-256 `819e4f82…d318`），前 744 bytes 是 93 個 8×8 字模、
  `0x2e8` 起是 93 個 8×6 字模，都由 ASCII 32 起算涵蓋 32–124。
  `tools/extract_fonts.py` 產出 `assets/original-derived-png/fonts/`。書本內文用小字型。
- 書本文字版面已機械量得並版控於 `docs/re/text-layout.json`；原版繪字規則為
  逐字模格清成白底、但空白字元不清底。原版字模格是 8×6、字距 8，繁體中文字形放不進
  同一格，因此採雙模式：英文模式維持 320×200 與原版字模供嚴格對拍，中文模式 960×600、
  16×18 中文字格——**倍率三是算出來的**：原版最小行距 6 px，`6 × 3` 剛好等於
  字模格高 18，兩倍只給得起 12 放不下（SPEC-030）。城鎮的文字則用 8×8 大字型，不是書頁的小字型。
- remake 目前畫面的展示圖在 `docs/screenshots/`，由 `tools/run_kol_shots.sh` 依
  `shots.spec` 在鎖定容器內一次重生，`MANIFEST.json` 記下產生指令、輸入雜湊與 PNG
  SHA-256，並要求有對拍收據的畫面與那張逐 byte 相同。README 的畫面區塊引用同一批圖。
- remake 對拍收據已建立（`docs/verification/remake-parity/`）：二十六份收據 285 對畫面
  全部通過（書本畫面 3、角色建立 9、圖像層 92、三段翻頁 54、刪除流程 13、
  HARBOR MASTER 1、旅店畫面 8、戶外地圖 22、建築進門 1、城門提示 4、建築對話 3、
  招募同伴 9、住宿存檔 11、角色資料頁 8、商店 9、訓練 6、野外地圖 4、野外時刻 4、
  任務地點 4、道路 10、進城 7）。野外那一份用 `compare` 指明每一對涵蓋的矩形，天候圖列在
  `not_claimed`——雲層動畫的版面已經解完並實跑驗過（RE-136），但還沒接進畫面，
  那一塊不宣稱，也不當成允許差異區。
  書本封面與目錄頁在原版滑鼠游標以外與 DOS EGA 原版逐像素相同；80 張 New Character
  畫面的角色圖像 blit 矩形 (208,106,40×38) 全部逐像素相同；空白 New Character 頁
  以原版字型與版面重繪後，64,000 個像素中只有 31 個不同且全部落在名字欄的原版殘留
  字元。對拍判準分成「素材與版面必須相同」、「明列允許差異區」與「必須相同區」三層。
- `Knights-of-legend.zip` 已在唯讀容器中通過 ZIP CRC 檢查，共 76 個中央目錄項目；可見
  6 個 MZ executable、安裝批次檔、城鎮／角色／物品資料及多組 `*PICS` 圖像候選。
- ZIP 多數遊戲檔時間為 1989，但 `KOLTOWN.EXE`、`CHARDATA`、`CHARPICS`、`SHLDPICS`、
  `STEEDPIC` 等顯示 1998。這份封裝是可運行 DOS 集合的 `強推論`，是否為未修改原始
  發行版仍是 `未知`。
- IDA Pro 9.4 `locked-v1` 已對六個 executable 完成第一層指紋。全部由 IDA loader 判定
  為 16-bit MS-DOS EXE；`KOL`、`KOLBOOK`、`MAKEMODE`、`TITLE` 出現同一個 Microsoft
  Quick C／MSC 5.1 runtime FLIRT 候選，這只證實可套用的 library signature 候選，不能
  鎖定精確 compiler 版本。`TANDYPIN` 尚未辨識函式，仍屬未知。
- `KOLTOWN` entry 已證實會由區塊尾端向前迭代，把 `0x131f` bytes 搬到較高的重疊
  segment，隨後 `retf` 至新 segment 的 `IP=0x0038`；這是 EXEPACK 前置搬移。檔案相對
  MZ 宣告長度多出的 81 bytes 已證實全為 `0x1A`，屬 DOS EOF padding，不是 overlay payload。
- `KOLTOWN` 已證實使用 Microsoft EXEPACK；精確版本未知。以固定 SHA-256 的 CC0
  `unexepack.c` 產生本機 derivative `01f6f9…d1d2`，IDA 9.4 在 unpacked EA `0x2b90c`
  建立真正入口後辨識 337 個函式。解包 EXE、`.i64` 與完整 disassembly 僅留本機 ignored。
- DOSBox-X 第二次乾淨啟動以 `KOLTOWN.EXE A` 在 runtime `23B3:001C`、C startup 前命中；
  2,281 個 relocation 全通過，正規化後 129,312-byte load image 與 derivative 零差異。
- 魔旅士 2010-04-16 的玩家心得與《CRPG 通鑑》中文版已進入 README；只作專案起源，
  不作玩法 oracle。
- 中文手冊 SPEC-003 已 CONFORMED：Poppler 22.12.0 對 79 頁玩家手冊與 5 頁參考卡建立逐頁
  雜湊清冊；82 個內容頁都有文字層，玩家手冊 76、77 經頁面 PNG 目視為空白跨頁。
  玩家手冊 79 頁與 Apple II 參考卡 5 頁已全部 reviewed；84 頁總 gate 通過。
- 圖片 RE：44 個實體候選歸併為 30 個邏輯 archive；TWN `.1/.2` 已證實是 continuation
  volume，總數由 960 訂正為 1,022 個 payload，全部通過 LZW 解碼。索引／分卷契約、
  尺寸、位平面、palette 與透明規則均已逐 archive 證實。
- `AWD1PICS.0` 是 KOLBOOK 固定寫出的 4000-byte 工作區快照；KOLTOWN 只讀其中索引
  payload。尾端 520 bytes 已排除於圖片流，但仍是未解析工作區資料，不是 padding。
- `CHARPICS`／`SHLDPICS`／`STEEDPIC` 已由 KOLBOOK／KOLTOWN 讀寫資料流證實為 16 筆、
  每筆 760 bytes 的固定可變角色圖片記錄表。KOLBOOK `sub_15AEC` 與 4 倍顯示
  `sub_15B6E` 已證實每筆是 40×38、high-nibble-first 的 4-bit EGA 圖；48 張 ignored
  PNG 與接觸表已重生。逐槽語意與未使用槽內容仍未知。
- `CHARPICS.0` 全 80 筆已由 KOLBOOK `sub_13D7D` 索引／解碼鏈與 RE-027 下游 consumer
  證實為 40×38、high-nibble-first 的 4-bit EGA 完整角色模板；每筆第 761 byte 為
  `0x00` 且不進像素。80 張 ignored PNG 與接觸表已重生；逐圖職業／裝備名稱仍未知。
- COMBPICS 全七筆已關閉：0..4 為 296×144 戰鬥場景，5..6 為 88×64 肖像，皆使用
  二階 RLE、4-bit 色彩與 BIOS mode 0Dh palette；圖號 0、1 另對上魔旅士文章同場景 JPEG。
  這只證實 COMBPICS archive，不代表全部圖片格式完成。
- TWN0..7PICS 全 243 筆已關閉：TWN0 全 25 筆為 192×96；TWN1..7 圖號 0 為
  56×57 彩圖加共享 AND-mask、圖號 1 為七筆一-byte placeholder、偶數圖號 2 起為
  88×64、奇數圖號 3 起為 192×96。236 張 ignored PNG 與八張接觸表已兩輪重生；
  TWN0、TWN1、TWN7 人工抽查通過，個別 NPC／地點語意仍未知。
- MONPICS 已關閉 32 組怪物 sprite 垂直鏈：64 個 payload 依序配成 4-bit EGA 彩圖與
  MSB-first 1-bit 遮罩；IDA 尺寸表、逐張像素數與透明接觸表均吻合。遮罩有效資料後多出的
  單一 LZW byte 不作像素，仍標為壓縮終止／padding 的強推論。
- PERPICS 已關閉六組人物圖垂直鏈：前六個 payload 是 4-bit EGA 彩圖，後六個依
  `[i,i+6]` 配成 MSB-first 1-bit 遮罩；六組 IDA 矩形、像素數與透明接觸表全部吻合。
- KOLMPICS 全 83 筆已分類：71 筆 active、12 筆一-byte placeholder。圖號 0 是
  320×200、2 是 296×32；68 筆共同展開為 24×20 UI 圖示，其中非固定 caller 的個別幾何
  位置／裁切不宣稱 caller-exact；canonical extraction shape 已證實。28 是 56×57
  AND-mask。68 圖示接觸表方向與序列均連貫。
- WEAPPICS 全 56 筆已關閉為 28 組角色武器彩圖／AND-mask：`sub_18F3A` 由裝備欄位、
  六值 selector 基表與 family 5 的 variant 2／3 交換規則產生偶數圖號；28 組尺寸、
  像素數與透明棋盤接觸表全部吻合。family／variant 只是導覽標籤，尚未推定遊戲內分類名稱。
- HEADPICS／TORPICS／LEGSPICS 各 108 筆已關閉為各 54 組角色頭部、軀幹與腿部
  彩圖／AND-mask；共享 selector 是 `9*renderer_class+variant`，三份 6×9 接觸表與
  324 筆像素／mask 長度均吻合。指令允許 variant 0..15，實際合法資料範圍仍待另證。
- KOLPICS 全 14 筆已分類：0／8／12 是 `sub_19B29` 轉換的 raw-planar 資料，1／2／4／5／13
  是 direct-blit RLE 圖，3／6／7／9／10／11 是一-byte placeholder。RE-030 已證實圖號 0
  是五個 converter block 的 256×5 renderer-data PNG；8／12 由 40 個完整 stride 與
  `y<<6+x` 固定為 64×160 grid PNG。這些不是 original display rectangles。
- MISCPICS 六筆已完整分類：image 0／1／3／4／5 是五張 direct-blit 4-bit EGA 彩圖；
  image 2 是唯一一-byte、全程未引用的非圖片 placeholder，不產生 PNG。
- 固定 ZIP 一般圖片只有 EGA `PICS`，但 `MAKEMODE.EXE` 已證實是官方 CGA／Hercules
  轉換器：它對 28 組 PICS archive 生成 `PCGA`／`HERC` 分卷，28 組圖數與現行清冊完全
  相符。使用者已於 2026-09-01 決定首版只做 EGA，CGA／Hercules 列為後續選項；
  MAKEMODE 的完整轉換像素規則保留為待辦 RE，不進首版交付範圍。
- SPEC-002 已 READY。完整 EGA extraction set 為 30 個邏輯 archive、1,022 payload：
  996 筆來源已證實、26 筆非圖片 placeholder、0 筆 unknown，形成 752 個 PNG 單位。
  `tools/extract_ega_assets.py` 已把全部 769 張 PNG 產到 `assets/original-derived-png/ega/`，
  逐張重讀驗證並與已版控 visual probe 紀錄的 SHA-256 核對，原子化替換、輸出根目錄受限。
  原版同狀態 parity：80 張 CHARPICS 模板在 blit 座標 (208,106) 兩輪各 80/80 逐像素相同；
  BOOKPICS[0] 對書本封面畫面只差螢幕中央 8×11 的滑鼠游標。
- 音訊 RE 已建立 SPEC-004 DRAFT。固定 ZIP 沒有獨立音訊檔；TITLE `sub_10F75` 已證實為
  兩段、每段 100 次隨機 PIT 除數更新的 PC Speaker cue。`TANDYPIN.EXE` 已證實是 DOS
  `AH=31h` 常駐記憶體程式，不是音訊播放器。六個 executable 的雙新建 IDA DB 直接 I/O
  普查只有 TITLE 命中音訊埠；六個未解析 DX 點均已人工排除為 EGA renderer。正常啟動的
  DOSBox-X `DX-CAPTURE /A` 已取得 38.229333 秒 WAV，四段非靜音窗及 peak 均已固定。
  使用者已於 2026-09-01 決定 OGG 採規格近似合成：依已證實的 PIT 除數序列與公開硬體
  規格自行合成方波再轉 OGG，不散布原版錄音。合成器實作與人耳驗收尚未完成。
- 角色建立已建立 SPEC-005 DRAFT。手冊固定名字→種族→性別→職業→reroll→圖像→保存
  狀態機；`CHARDATA` 為 16×768 bytes、`CHARPICS` 為 16×760 bytes。IDA 固定 KOLBOOK
  `chardata` EA `0x1945a`、路由 `sub_134C2` 與五個立即 arg=0 caller。屬性 RNG 已證實為
  BIOS tick seed、`rand()%N+1` wrapper、八值骰面表、獨立骰式／內部基值與 100/150 顯示
  量尺；三個衍生值均為內部屬性的向下取整平均。`BOOKDATA` 的 39 筆記錄、四種族、兩個
  secondary 值及 33 個 menu names 已建立完整 selector；35 個手冊列與 39 筆記錄的差異由
  男女精靈共名及矮人雙名稱同記錄解釋。39 筆起始金冠、主要／次要武器及武器技能已與
  中文手冊逐列相符；最小 writer offset、CHARPICS／SHLDPICS 同槽寫入及三檔保存順序亦已
  固定。完整 writer coverage 已排除其他職業衍生起始技能／物品，KOLTOWN 最小 reader 亦
  交叉確認名字、種族、職業、金冠與主要武器。KOLBOOK 靜態控制流另已固定輸入 buffer
  首 byte 為零時重試、各層取消返回、選項 0 reroll、80 張肖像循環與 16 槽全滿時在 writer
  前退出。
- SPEC-005 已 READY。正常 DOS 玩家路徑的書本與城鎮開場 UI 有 22 個情境、32 輪的可重生
  收據（`docs/verification/character-ui/`，43 個 run、6,515 個去重畫格；核心情境各兩輪）。
  入口鏈為 `KOL.EXE` → TITLE
  循環標題序列 → 任意鍵 → 書本封面 → 任意鍵 → 目錄頁 → `5` → Character options →
  `CREATE CHARACTER`。名字欄容量 16 個可見字元，輸入第 16 個字元即直接前進種族層、
  不需 Return；空名按 Return 不前進；取消目的地為種族→名字、性別→種族、職業→性別、
  屬性→職業，回到名字層時欄位清空；`REROLL STATS` 重擲九個顯示值；圖像 selector 恰好
  循環 80 張；接受圖像後才寫檔，兩輪都只寫 `CHARDATA` 的 slot 14。跨輪差異只出現在
  TITLE 動畫落點、名字欄游標與屬性數值三類畫面且差異量已量化。16 槽全滿、重名處理、
  名字欄字元集合與大小寫保存方式仍未由 UI 證實，實作必須標為未驗證分支。

## 固定輸入

| 檔案 | Bytes | SHA-256 | 分類 |
|---|---:|---|---|
| `Knights-of-legend.zip` | 1,699,452 | `2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46` | 僅本機原版封裝 |
| `knightslegend-manual.pdf` | 35,942,482 | `4c80f5d4c7221f9d3376546d249d52e2fc7fc0f73b48ee9daf6e20ce87d7086a` | 僅本機英文說明書 |
| `knightslegend-refcard.pdf` | 2,253,388 | `4815bc6c3a51ff7abaa610a071f08e05eb2dd180e5130b0674de8b4a94b2abd3` | 僅本機參考卡 |
| `KOL-MAP-001a4s2.jpg` | 92,661 | `9528ca50e14d09d746d2ecce523ac56f78baeaec5cdd4ba5e3351679bd52dff9` | 僅本機參考地圖 |
| `knightslegend.jpg` | 433,061 | `76b08e4b6a6da9b616d7f57ce3e77c2973591379c8897779d849126555defb9b` | 僅本機盒面 |
| `knightslegend-back.jpg` | 525,582 | `10d3e63358d6604b807d4afae3b53acf9a2962ce10871390dd815e325e53b194` | 僅本機盒背 |

## 文件入口

- 執行規範：`AGENTS.md`
- 唯一待辦：`WORKLIST.md`
- 工作歷程：`WORKLOG.md`
- 研究證據：`docs/re/RESEARCH-LOG.md`
- RE 知識庫索引：`docs/re/README.md`
- 重製計畫：`REMAKE-PLAN.md`
- 驗證矩陣：`docs/verification/VERIFICATION-MATRIX.md`
- 交接：`HANDOFF.md`

## 下一個唯一工作

**用點擊序列把原版推進戰鬥**（2026-09-06 起）。原版現在跑得到可互動的遊戲畫面：
`KOL.EXE` 是主程式，設好視訊模式 `0Dh` 之後用 `int 21h AH=4Bh AL=03h`（載入
overlay）把 `TITLE.EXE`／`KOLBOOK.EXE`／`KOLTOWN.EXE` 搬進自己配好的記憶體再
far call 進去——**不是 EXEC 子程式**。城鎮載入 35 個檔之後停在 `sub_1B100` 輪詢
`int 33h` 等玩家操作，在對話框上點一下文字就推進到下一段，底部長出六個圖示的
指令列。`Click` 要帶 `NoCursorWait()` 與 `Frame((*Oracle).Tandy16)`。
證據在 `docs/re/RESEARCH-LOG.md` 最後一節；dosgolem 那一側的三個修正
（記憶體服務的界線、載入 overlay、handle 號碼重用）在 `kol-force-call` 分支。

**選部位那一問的實跑收據**：互動已經做了（使用者 2026-09-06 決定回到原版），
但十六名原版角色裡只有 Warder 有法術、而且走不到這一問，所以要先在 SCROLL 買一筆
效果 6 且不用選目標格的（The Hostel 第一筆 Tyanalon 340）。路徑摸到一半——
買得到、出得了城、種子 11 在野外第 14 步遇到 GHOUL——卡在戰鬥那一段的按鍵序列
（SPEC-012 第 8 節）。

**徒手攻擊拿的是穿戴品**（RE-131）：`sub_26AC2` 依指令碼挑「這一手用哪一件東西」，
徒手那一條查 `[0x0FD2 + 攻擊選項] + 0x100`，除以一格 0x10 之後正好是穿戴欄的
FEET／SHIELD／HEAD／HANDS——四式就是踢、盾擊、頭槌、拳。傷害骰與額度的速度項
都從那一格來，remake 之前一律拿手上那件在算。怪物那一側還沒接（三個位移落在
怪物記錄的已知版面之外）。

另外兩件要原版畫面當 oracle：競技場的遭遇訊息與戰場地形、`sub_16239` 那一頁。

**書本的中文標籤改走 UI 文字庫**：`internal/i18n` 的 `LoadBank` 把
`ui-zh-Hant.json` 的 BOOKDATA 那一份疊到 `newchar-zh-Hant.json` 上，兩份譯文從此
只有一個來源；`verify_text_fit.py` 多驗了 KOLDATA 畫在城鎮對話框裡的每一行
（6,952 項全過）。
**城鎮的兩格圖示接上了**：馬廄的買賣（RE-119／120，收據 `town-stable`；買下的坐騎
會換掉隊伍腳程，「全隊都有坐騎才問得出來」的騎走那一頁因此走得到）與 SCROLL 的
入會與買法術（RE-121，收據 `town-scroll`）。後者順帶解掉 SPEC-012 的「法術的母表」
——法術就是 SCROLL 賣的那八筆——並更正了 RE-074：角色記錄 `+0x1A0`／`+0x1A8`
是所屬教派與頭銜，不是盔甲與盾牌。
**戰鬥又補了三個洞**（RE-124）：下馬那一手接上了（0x28 印「lands!」並清掉 `+0x1E`，
上馬本來就在移動方式第 3 式）；**逃走之後掉東西那一條原版就沒有**（掉落機率在 BSS
裡、零個寫入點，恆為 0）；`sub_227C0` 不是騎乘，是「這一格被別人佔著嗎」。
**敵人 AI 的兩個缺口也補完了**：移動段繞不出去時的三條退路（RE-117）與訊息的
可見性過濾（RE-124：怪物那一側要有隊員在 ±6 格內而且看得到才印，看不到就整段
跳過）。兩項都改變戰況，`arena`、`full-chain` 與 `full-chain-ending` 三條長路徑的
固定按鍵序列因此整條重生——重生本身做成了工具（`tools/run_driver.sh` 把 driver
整段跑在同一個容器裡，`regen_long_path.py` 只重打戰鬥段，`remap_receipt_steps.py`
把里程碑對到新步號），七份長收據、七份戰鬥收據與十七份短路徑收據全部重驗通過。
鍛造那一條也讀完了（RE-125），**但入口條件對不上原版資料**，先不做。
**角色立繪的合成解出來了**（RE-126）：資料頁那兩頁把體型、腿、身、頭、盾、武器、
手七層疊起來，版面來自 KOLTOWN 資料段一串首尾相接的表，212 張抽出來的 PNG 尺寸
零處不符。合成規則已經做成 `internal/portrait`，還缺那兩頁的背景與外框——
原版畫面沒有 oracle，沒有證據不先畫。
**資料頁那兩頁都做完了**：物品那一頁（RE-127）有立繪、十七列清單
（HEAD…BELT／WEAPON／POCKET ×6）、明細與七個動作（拿到手上、穿上、掛腰帶、
給隊友、放背包、丟掉、吃掉）；隊伍那一頁（RE-126）由後往前畫整隊。那一頁順帶證明
**腰帶與手上那件是兩格**（`0x190` 與 `0x30`），typed data 與 `town.Member` 因此改成
穿戴欄十格 ＋ `InHand`。**這兩頁也對過原版畫面了**：各自允許區外零像素差異（收據
`receipt-town-sheet-items.json`、`receipt-town-sheet-party.json`）。對拍逼出四處
更正：`sub_1BA9D` 是「填內部 ＋ 畫框 ＋ 投影」的複合所以每個框都自帶一圈四像素
黑帶、`sub_1103F(0)` 是整頁圖而不是把框內填黑、隊伍那一頁的圖示列縮成一格
SPACEBAR，以及**腰帶那一格另外橫放在畫面左下角**（只有 dx 是 0 時才畫，所以
隊伍那一頁看不到）。
**野外的治療遭遇也接上了**（RE-112／128）：怪物編號 0xFF 那一條不是任務怪，是路上
遇見誦經的僧侶或披斗篷的療者，答應之後全隊的傷勢與飢餓一起清零。那六段敘述在
`TOWN0` 的 payload 4——DGROUP `0x2350` 在 BSS 且零個參照點，掃 `sub_1A4FE` 的每一個
載入點才找到。實跑收據 `world-healer`。
**任務物品也擺到戰場上了**（RE-113／128）：座標是任務地點記錄的 `+0x27` 列與
`+0x28` 欄。RE-113 記著「那兩個全域被誰讀還沒解」——那是過濾器有洞，讀的地方走
相對定址，xref 一個都看不到。
**亂數換成原版那一支，而且逐次對拍過了**（RE-132）：原版用的是 Microsoft C 的
`rand()`（`種子 × 0x343FD + 0x269EC3`，回高 word 的低 15 位元），remake 之前用的是
Numerical Recipes 那一組，**兩邊的序列從來沒有對過**。現在是 `internal/rng`。

判準不是我對指令的解讀，是原版自己在執行期回的值：[dosgolem](https://github.com/wicanr2/dosgolem)
把 `KOLTOWN-unpacked.EXE` 跑在同一個 Go 行程裡，壓好種子**直接呼叫**原版的
`rand()` 與 `rand(n)`——6 個種子 × 500 抽逐次相同，`rand(n)` 十種面數各 40 次
逐次相同，收尾的 32 位元種子也相同（`internal/oracle`，`-tags oracle`，0.02 秒）。
正對照先行：`_main` 進入時種子讀到 MS C 的初值 1，一次證明映射、DS 與種子位址都對。

強制呼叫這條路是這一輪在 dosgolem 加的（`Call`，分支 `kol-force-call`）。原本想等
遊戲自己抽，跑滿 2.6 億道指令、送 12 批鍵是 0 次；而**中間任何一步錯了，看到的都是
「一次都沒抽」**，與「原版根本不抽」長得一模一樣。**這一層驗的是常式的實作，不是
呼叫時機**——誰在什麼時候抽、抽幾次還沒驗過，那要 dosgolem 第四層（`apps/kol/`）。

順帶量到原版 `rand(0)` 會**除零當機**（R6003，回傳碼 255）。`rng.Roll(0)` 回 0 是
故意的不一致：收到 0 表示上游算錯了，不是安全的預設值。

代價是**十三份收據基準失效**（遭遇與治療者的觸發步數位移）。根因、形狀與重生流程
記在 [WORKLIST 的「收據基準」](WORKLIST.md#收據基準13-份已知失效待重生)一節。
跑起來會紅，那是正確的狀態。

**疲勞那一整條鏈接上了**（RE-131）：一手的時間額度 `word_30D06` 不是行動點數——
沒有任何地方拿它當「還能不能動」的閘門，只是每一手結算一次的計量。額度扣掉挨的傷、
指令成本與施法成本之後除以十扣進疲勞，**為負就是往上加**；額度裡有
「累積傷害 × 10」，所以傷得越重疲勞漲得越快。這條迴路是 RE-081 那一整套撤退判斷
唯一的驅動來源。順帶解出**力竭會當場放倒人**（比例 ≥ 100 就把頭與胸設成 0xFF），
三句話依隊員／怪物／骷髏分開（`+0x10` 的 bit 7）。
**戰鬥數學修掉一個對錯**（RE-130）：`sub_23683` 算的是
`50 × 疲勞 ÷ (力量 ＋ 健康)`，remake 原本實作成 `50 × 累積傷害 ÷ 昏迷門檻`，
等於替原版加了一條「傷得越重打得越差」的乘法惩罰——規劃鍵值、執行鍵值、攻防值、
力量加成與敵人 AI 的撤退門檻五處都吃它。判準是全庫掃 `[bx+322Eh]` 的寫入點：
傷害落到身上那一支不在名單裡，只有法術效果 5 會加。`Actor` 因此加了
`Fatigue`／`Stamina` 兩格，`+0x2D` 那一格改名 `Hurt`。效果 5 與休息回復一併接上。
`full-chain` 的戰鬥少打了兩回合、按鍵序列失效，用 driver 重生（957 → 1067 個
token），十五份收據全部重跑通過。
**攻防技能的罰值也接上了**（RE-129）：法術效果 7、8 寫的參戰者記錄 `+0x2E`／
`+0x2F` 是攻防技能的罰值，一點抵十、法術過期不還。順帶查出原版**沒有**為這兩格把
正負反過來（效果 5 的疲勞有反），所以「正面」的法術反而扣對方的技能——remake 照抄
不代它修。疲勞的分母同一輪解出來是「力量 ＋ 健康」，RE-122 兩個洞都填掉。
**法術編輯器那一頁也做完了**（RE-128）：SCROLL 選單第二列挑一筆法術，十三個欄位
（名稱、四格護符的類與種、屬性、射程、兩個骰式）改到滿意，離開時比過冒險點數與
金冠才扣款寫回。選項與造價全部來自 `MAGIC.0`——**這個原版資料檔先前從沒被解過**，
28 個 SUBCLASS 名字有 27 個一字不差地出現在 `monsters.json` 裡。它的骰式表順帶
交叉驗證了戰鬥那邊的骰面表。教派准法術對哪幾類生效寫在建築 payload `+0x60`，
六個教派剛好把八類分完。**這一頁已經對過原版畫面**：整頁 320×200 允許區外零像素
差異（收據 `receipt-town-spell-edit.json`），路上順帶更正了兩張成本表的歸屬
（`0x600`／`0x700` 是金冠）、原版畫面把那兩行標籤印反這件事，以及 RE-121 的
「買完法術直接離開 SCROLL」。
**打字那一格的版面已經對拍完**（RE-114：`sub_1B5C8` 的欄位、兩種字型與那條每
0.27 秒翻一次的橫槓；收據 `town-mouth-run1` ＋ `plan-town-mouth.json`），
**清單元件的參數語意也解完了**（RE-115：`sub_1AAE8` 的七個參數、選中列換色、
捲動與滑鼠命中）。
完整清單見 `WORKLIST.md`。

## 交付現況（2026-09-07）

`dist-all/0.2.0/` 由乾淨工作樹重建，**三個平台都在裡面**：linux/amd64 原生、
windows/amd64 交叉編、macOS universal（arm64 ＋ x86_64，osxcross ＋ CGO）。
`promo/` 有 59 秒的推廣片，畫面全部是實跑截圖、沒有背景音樂（原版本來就沒有）。

**版本號 `0.2.0` 是暫定的**，仍待使用者決定；產物可重生，換號碼重跑一次就好。

**還沒發 Release。** 權利審查已經做完並寫在 `docs/RELEASE-RIGHTS-REVIEW.md`：
`release/` 可以公開、`full/` 不可以。缺的是兩個決定——說明書譯文要不要進
`release/`，以及「建立第一個 GitHub Release」這個對外、不可逆的動作要使用者
明確同意。Windows 與 macOS 都**沒有真機驗收**，容器裡的靜態檢查不冒充它。

## 已定案的共同決策

| 日期 | 議題 | 決定 |
|---|---|---|
| 2026-09-07 | 輸入方式等價 | remake 可用滑鼠取代鍵盤，以操作結果等價於原版為準；不要求相同裝置或按鍵序列。不因此省略功能、改變規則或放寬資料保存要求。 |
| 2026-09-04 | 授權條款 | **RRSAL-1.0**（復古重製 source-available 授權條款 1.0，SPDX `LicenseRef-RRSAL-1.0`）。非商業免費含修改與再散布、實況與影片明示允許（含平台分潤）、貢獻回授、商業要洽談。`LICENSE` 與 README 授權段已建立；**不涵蓋原版素材**，第 2 條逐項寫明界線。這不是開源授權，對外一律寫 source-available。 |
| 2026-09-03 | 戰鬥忠實度 | **忠實 ＋ 可關閉的加速**。規則、傷害計算與回合結構完全照原版；另提供「重複上一指令」「快速結算未接敵回合」「動畫速度」三個可關閉的選項，預設全開。**對拍一律在三個選項全關的狀態下驗證**，加速只影響玩家的等待，不影響任何結果。 |
| 2026-09-03 | 存檔格式 | **remake 自有格式 ＋ 匯入原版**。存檔是可讀的 JSON，含 remake 才有的狀態（世界時間、任務進度、設定）；另提供從原版 `CHARDATA` 匯入角色。原版檔案一律唯讀，不寫回。已落實為 SPEC-006、`internal/save` 與 `tools/import_original_characters.py`。 |
| 2026-09-03 | 客棧住宿是否離隊 | **預設不離隊，設定開關 `faithful_inn_stay` 切回原版**。原版住宿是「收費 → 寫存檔 → 角色離隊 → 隊伍空了問是否離開遊戲」，那是一人一檔的磁片設計。開關打開就完全照原版跑，**與原版對拍一律在打開的狀態下進行**。 |
| 2026-09-01 | 圖像 theme | 首版 EGA-only；CGA／Hercules 列為後續選項。 |
| 2026-09-01 | PNG 入庫例外 | 已驗證、可重生且有 provenance manifest 的 PNG 衍生物可進 private repo；不含 ZIP／EXE／archive／PDF，也不等於公開散布權。 |

待使用者決策：`assets/original-derived-data/character-creation.json`（39 筆職業骰式、
內部基值與名稱表）是原版衍生的 typed data，不是 PNG。2026-09-01 的 private-repo
例外明文只涵蓋 PNG，因此這份檔案目前留在本機並由 `.gitignore` 排除，改由
`tools/extract_character_data.py` 從固定原版輸入重生（抽取器本身有進版控）。
要不要把例外擴及 typed data，屬權利邊界決策。

未關閉的邊界：16 槽全滿、重名、名字欄字元集合；展開跨頁貼圖殘影的貼圖矩形切法；
其餘 29 個 archive 的原版同狀態畫面收據；PCGA／HERC 轉換規則。
# 2026-09-08 最新補充：怪物資料契約與初始狀態

怪物遠程欄位的 13 種載入案例通過，包含舊資料缺欄、null、負數、超界與錯誤型別，
並保留合法 0／255。正常原版收據的 remake 側比較已由 59 增為 60 欄，補回 SLEDGE
遠程能力；完整 Go 測試與建置通過。來源與限制見
`docs/verification/battle-party-state-20260908.md` 最新追加節。
本次沒有重跑 dosgolem、沒有重建 AppImage；自評維持 75/100，完整戰鬥驗收仍未完成。
# 2026-09-08 最新補充：戰場亂數地形索引

依 SPEC-008 修正一般地形生成及固定／一般地圖灑點的轉置：原版 `Y*64+X`
位置轉為正式 `X*64+Y`。六組 dosgolem 固定種子比較共 24,576 格及亂數終態全部一致，
完整 Go 測試與建置通過。收據：`docs/verification/battle-field-index-20260908.md`。
正常整場、完整戰後保存及新版 AppImage 仍待完成；自評維持 75/100。
# 暫停交接（2026-09-08，使用者要求）

目前暫停開發與新測試，自評維持 75/100。最後已啟動的 dosgolem 正常流程已自然
結束：`TestNormalBookEntry` 84.44 秒通過，地形收據已產生於
`workplace/battle-terrain-normal-fixed-20260908/battle-terrain.json`。
這份收據尚未接入 remake 的完整同狀態比較，不能宣稱整場戰鬥完成。

本輪修正匯出工具：生成階段暫存指標 DS:2B9A／2B9C 不能沿用到規劃畫面；
IDA 證實繪製直接讀 DS:2DBA 地形段。首次失敗保留在
`workplace/battle-terrain-normal-20260908/`；訂正見
`docs/re/20260908-battle-field-index.md`。
最近正式程式修正仍為怪物遠程初始化及地形索引轉置，完整 Go 測試與建置已通過。
現行 AppImage v.0.2.2-20260908 尚未包含這些修正；未提交、未發布新版本。
恢復時優先接上新地形快照，再驗指令／結算及正常戰後保存；其他 G01～G07 缺口仍依 WORKLIST。
