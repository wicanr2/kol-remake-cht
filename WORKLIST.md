# 《武士傳說》重製工作清單

更新日期：2026-09-25

## 這份檔案怎麼讀

**未完成項的權威是 `docs/worklist.json`**，不是這裡的 checkbox。下面「未完成項」
那一節由 `tools/worklist.py render` 產生，每一條掛一個會自己開口的 `verify`；
本檔其餘部分是歷史紀錄與已完成項，散落在各節的未完成項只留一行指標。

```
tools/run_python_tool.sh tools/worklist.py verify    # 逐條核實，抓過期斷言
```

目前狀態與現行交接見 `CONTEXT.md` 頂部。使用者先選本機可玩版，隨後提高為
「所有玩家文字繁中、手冊功能與原版逐項一致」的完成閘門。先前
`v.0.2.3-20260923` AppImage 是此閘門前的本機候選包，敗北後回城保存／重載
已抽驗；重製版觀察用資料的正常按鍵勝利與繁中戰利品畫格已補收據，
但正式封包自然資料的勝利、完整戰鬥原版對拍及其他手冊功能尚待驗，不能改稱完成。
詳細缺口見 `docs/verification/manual-fidelity-audit-20260923.md`。公開發行另行審查。

涉及未決設計的分支等待使用者確認，不自行視為同意。

<!-- worklist:begin 由 tools/worklist.py render 產生，不要手改 -->

## 未完成項（權威是 `docs/worklist.json`）

這一節由 `tools/worklist.py render` 產生，**不要手改**；改 `docs/worklist.json`。
每一條掛一個 `verify`：跑起來為真＝這一條仍然未完成，為假＝訊號不見了，
要回去看它是不是做完了。核實方式：

```
tools/run_python_tool.sh tools/worklist.py verify
```

每一條在 GitHub 上有一張對應的 issue（條目後面的 `#編號`），由
`tools/sync_worklist_issues.sh` 從這份 JSON 鏡像過去；在 GitHub 上改內文或
開關不會回流，下一次同步會蓋回來。

### 玩家會撞到的功能缺口

- **g05-picture-shield-editor-spec**（#3） — G05：圖像／盾牌編輯另立規格
  EDIT PICTURE／EDIT SHIELD 未列入 SPEC-005 已完成的建角範圍，要另立規格。原版是滑鼠驅動的點陣圖編輯器、ESC 無作用，現有鍵盤 oracle 走不完。
  完成條件：規格立起來並經證據審查升 READY。
  核實：`internal/flow/flow.go` 找得到 `EDIT PICTURE／EDIT SHIELD 是滑鼠驅動的點陣圖編輯器，還沒實作` 就是還沒完

- **g07-battle-speed-options**（#5） — G07：三個加速選項的詳細規格、控制與保存
  SPEC-028 草案已核對現有接口與完成條件，狀態仍是 DRAFT，明寫「不授權先加空開關」。無訊息自動略過不等於未接敵快轉；細節與可見原型待確認。
  完成條件：SPEC-028 升 READY，三個選項可開關、會保存，並有可見原型讓使用者判斷。
  核實：`docs/spec/028-optional-battle-speed.md` 找得到 `狀態：\*\*DRAFT\*\*|未接敵|尚未回答而替使用者核准|三個已知設定及 UI 控制未接` 就是還沒完

- **a03-a04-writeback-and-contract**（#14） — A03／A04：戰後傷勢、裝備、戰利品與任務物品回寫；可變狀態保存契約
  勝利／撤退的物品與傷勢回寫、零金冠、訓練技能與設定重載已在第一批修復處理。RE-139 揭出的任務交付順序與全隊回報已依 SPEC-014 READY 接回，觀察用存檔的正常按鍵及英中狀態相同；自然取得任務物品、原版同狀態及跨戰鬥／城鎮／存檔逐欄一致仍未驗。
  完成條件：從自然接任務、取得物品、城鎮交付到保存／重載，跨戰鬥與城鎮逐欄一致，並有原版同狀態證據。
  核實：要人判——同 a01-a02：逐欄一致要跑過才知道。

- **a05-a06-a09-belt-ward-nutrition**（#15） — A05／A06／A09：腰帶轉接與武器數值交換、14 名角色 Ward 傳遞、修道院 Nutrition
  三個缺陷已在第一批修復處理，診斷測試尚未轉為回歸測試。
  完成條件：對應診斷測試轉為回歸測試並通過，再補正常玩家路徑抽測。
  核實：要人判——完成物是測試檔，而 verify 不掃測試檔（測試會提到還沒接的東西）。正常玩家路徑抽測收據落地時改綁它。

- **weather-animation-and-speed**（#20） — 接正式天候動畫及既定加速選項
  SPEC-026 已接六組天候、128 像素捲雲與正式節拍，dosgolem 四組完整合成與六組分派通過，正常出城 GUI 只在天空區改變；雨雪的正常 GUI 與原版正常野外同相位仍未宣稱。加速選項的部分與 g07-battle-speed-options 同源。
  完成條件：雨雪在正常 GUI 下與原版同相位；加速選項依 SPEC-028 READY 後接上。
  核實：`docs/spec/028-optional-battle-speed.md` 找得到 `狀態：\*\*DRAFT\*\*` 就是還沒完

- **picture-shield-bitmap-editor**（#31） — EDIT PICTURE／EDIT SHIELD 的點陣圖編輯器
  原版是滑鼠驅動、ESC 無作用，現有鍵盤 oracle 走不完；要往下做得先讓 oracle 能送滑鼠事件。這是 g05-picture-shield-editor-spec 的實作段。
  卡在：oracle 還不能送滑鼠事件
  完成條件：oracle 能送滑鼠事件，編輯器接上並對拍。
  核實：`internal/flow/flow.go` 找得到 `EDIT PICTURE／EDIT SHIELD 是滑鼠驅動的點陣圖編輯器，還沒實作` 就是還沒完

- **building-followup-dialogue**（#32） — 建築內的後續對話（訊息 10 起）
  訊息 10…13 的選法已知（聲望等第為 0 時是「訊息 10 ＋ 性別×2 ＋ 會不會法術」，否則訊息 26 ＋ 等第），但觸發條件還沒讀出來。
  完成條件：觸發條件解出並接上，正常路徑能看到後續對話。
  核實：要人判——topics.go 已有訊息 10 的選法但沒有觸發條件。接上時會多一支判斷觸發的函式，到時改成 absent 綁它。

- **text-layout-remaining-screens**（#34） — 量測城鎮、世界、戰鬥畫面的文字版面
  書本版面已機械量得；城鎮、世界與戰鬥已有部分中文安全矩形和畫格。資料名稱 387 筆在五種固定欄位新增 1,935 項寬度檢查，目前合計 9,301 項失敗 0；戰鬥訊息與其他分支仍未全覆蓋。最新手冊／中文稽核見 docs/verification/manual-fidelity-audit-20260923.md。
  完成條件：三種畫面的文字版面都量過，每個 CJK 欄位有安全矩形。
  核實：要人判——版面量測的產物是 JSON 與收據。三份版面表落地時改成 json_len 綁它們的項數。

- **cjk-remaining-screens**（#35） — 城鎮、世界、戰鬥畫面的中文化與職業名暫譯的交叉確認
  城鎮原始資料 2,121 則非空後續訊息已有 2,076 則譯文，其餘 45 則經 RE-140 判定為非文字槽位。2026-09-25 以 tools/scan_cjk_text_leaks.py 在中文模式重播 23 條玩家路徑與 9 個戰鬥情境，修正名冊所在地、教派名與頭銜、戰鬥撿丟物品名、敵名冠詞、所有格「他的的」與刻意留空譯文退回英文；戰鬥執行段訊息原本是看不見的黑塊，已修。法術名依使用者決定保留精靈語原文並加注解（SPEC-034）。顯示譯名 492 個原名、字模 2,553 字、安全矩形 9,467 項失敗 0。未涵蓋：玩家路徑沒走到的分支、畫在原版圖片裡的英文、原版同狀態；職業與專名暫譯待交叉確認。詳見 docs/verification/manual-fidelity-audit-20260923.md。
  卡在：補未涵蓋分支的中文畫格；職業名譯法定案
  完成條件：三種畫面的中文都上得去且不溢出，職業名譯法定案。
  核實：要人判——要人判：靜態譯表與文字安全矩形只能證明局部覆蓋；完整玩家路徑與動態名稱仍需收據。

- **save-across-battle-and-clock**（#47） — 存檔跨事件與戰鬥階段恢復：戰鬥中途狀態與世界時間
  任務旗標、已回報數與身上帶的東西已進存檔，並由 `full-chain` ＋ `full-chain-reload` 驗過往返。戰鬥中途的狀態與世界時間還沒進存檔——存檔格式裡連日期欄位都還沒有。
  完成條件：戰鬥中途存檔重載後接得回去；世界時間跨存讀檔一致，不會重複付款、任務或戰鬥結算。
  核實：`internal/save/save.go` 找得到 `世界時間還沒有規格` 就是還沒完

- **spell-target-figure**（#71） — 施法時目標的人形還沒畫
  `sub_27E20` 的圖庫與幾何只到 `強推論`（RE-047 那一輪標的），所以選部位那一頁的人形沒畫出來。
  完成條件：圖庫與幾何升到已證實，人形畫出來並對拍。
  核實：`docs/spec/012-combat-spells.md` 找得到 `目標的人形還沒畫` 就是還沒完


### 接了但沒驗過

- **g03-arena-vertical-chain**（#2） — G03：競技場正常取得資格與完整垂直鏈
  固定地圖已依 SPEC-025 接線，4096 格 dosgolem 解碼比較與地圖隔離回歸通過；繁中空肖像繪圖錯誤已修，受控資格存檔可經正式按鍵選人、不下注及進場（docs/verification/arena-dialog-fix-20260908.md）。剩正常取得資格與完整垂直鏈；原版特殊肖像仍未知。make_arena_save.py 只產生受控輸入，不能代替正常取得資格。
  完成條件：玩家從正常路徑取得資格、下注、打完單挑並結算，全程不靠受控存檔。
  核實：`cmd/kol/main.go`、`internal/ui/town.go` 找得到 `sub_2B454` 就是還沒完

- **v01-full-battle-parity**（#6） — V01：完整戰鬥同狀態結算、武器／護甲與保存的正常垂直鏈
  地形亂數位置轉置已依 SPEC-008 修正（三組固定種子、兩條路徑 24,576 格與亂數終態一致）；SPEC-029 修好敵人 Kind 與遠程初始化（768＋2304 組通過，加 SLEDGE 後合計 60 個初始欄位）；SPEC-027 修好戰後傷勢折算與敗北回寫（4608 組、五條結算至 v2 重載）；任務物品名稱誤認已修（258 組）；任務回報獎勵的母表共用指標已修。以上都是常式比較，正常整場仍未驗。
  完成條件：一整場戰鬥從進場到結算、保存都與原版同狀態，走的是正常玩家路徑。
  核實：要人判——常式比較過不代表整場過，沒有單一訊號能代表它。整場對拍收據落地時改成 absent 綁那份收據的路徑。與 v01-remaining-battle-chain 是同一件事的兩次登記。

- **battle-normal-seed-alignment**（#9） — 原版戰鬥同狀態：攻擊結算比較
  兩名隊員與 SLEDGE 共 60 個初始欄位、4096 格地形、SLEDGE 本體記錄 20 項都通過。**規劃段也比完了**（2026-09-11）：主指令列 5 格、第 1 格 0x29，走同一組位置之後三個決策值 Command=0x29／Style=1／Guard=2 與原版逐項相同，正反對照過——planning-parity-20260911.md，跑法 tools/run_original_parity_tests.sh。剩下的是**攻擊結算**：原版要在規劃寫回之後繼續跑到結算，remake 走同一組輸入比傷害與狀態。
  完成條件：同一 seed、同一初始狀態下，remake 與原版的攻擊結算逐項一致。
  核實：要人判——結算的對拍沒有產品碼訊號可綁。**不要綁對拍測試那句收尾自承**：verify 不掃測試檔，綁 present 會每輪回報「找不到＝可能已完成」，綁 absent 則永遠為真。規劃那一半已由 TestOriginalNormalPlanning 蓋住；結算的載入器接上時 internal/oracle 會出現產品碼落點，到時綁它。

- **inn-farewell-diff-review**（#10） — 審查原版離店告別訊息與確認流程
  SPEC-024 與告別確認已實作，dosgolem 單人／多人正常離店與繁中 GUI 已抽驗，但不是全城逐店文字驗收。要補與 remake 直接離店的玩家可見差異。
  完成條件：列出所有可見差異並各自處置，或證明沒有差異。
  核實：要人判——差異審查的產物是文件不是程式。審查報告落地時改成 absent 綁它。

- **dosgolem-gui-ai-acceptance**（#11） — 補新版正常 GUI 與玩家可見 AI 行為驗收
  最新工作樹已重跑 Olear 正常射擊 150 個輸入、確認彈藥 20→19，dosgolem Olear 旅店入口亦通過（docs/verification/fixed-seed-player-path-20260908.md）。完整 AI／原版戰鬥仍待驗。
  完成條件：玩家看得到的 AI 行為（撤退、反制掃描、側翼、撿武器）在正常 GUI 下與原版一致。
  核實：要人判——AI 行為要實跑才看得見。整場 AI 對拍收據落地時改綁它。

- **appimage-parity-continuation**（#12） — 旅店、戰鬥與存讀檔的 AppImage 同狀態對拍
  要等 remake 未完成的接線修好才能往下做。工具鏈側已可用：dosgolem 正常 KOL.EXE 入口經 DOS Enter 進 KOLBOOK，已走到 Halfway House。
  完成條件：AppImage 與原版在旅店、戰鬥、存讀檔三段同狀態一致。
  核實：要人判——卡在上游接線。對拍收據落地時改綁它。

- **a10-battle-info-panel**（#17） — A10：戰鬥左側完整資訊內容的原版同狀態
  人形／傷勢／疲勞（SPEC-018）、左上裝備／肖像與模式切換（SPEC-019）、左下命令／招式（SPEC-021）、彈藥數字（SPEC-022）都已接線；3,168 組命令參數、144 組遮罩、十二組直條與九組敵人智力檢定由 dosgolem 通過，正常 GUI 已畫出人形、兩張戰場畫格縮放比較零差異。完整裝備與攻擊後傷勢對拍仍待完成。
  完成條件：正常戰鬥能辨識目前角色、裝備與傷勢，破損後顯示同步；不能以空框截圖通過代替。
  核實：要人判——面板已接線，缺的是攻擊後的同狀態對拍。那份收據落地時改綁它。

- **roster-full-ui-oracle**（#28） — 補 16 槽全滿、重名與名字欄字元集合的正常 UI oracle
  SPEC-005 已經證據審查升 READY，但 16 槽全滿只有靜態證據，沒有正常 UI oracle。
  完成條件：三種情境各有正常 DOS UI 的畫格與輸入序列收據。
  核實：要人判——oracle 收據是檔案不是程式。那三份收據落地時改成 absent 綁它們的路徑。

- **minimal-player-chain**（#30） — 選一條最小玩家鏈：PNG → typed data → 一個互動 → UI → 存讀檔
  full-chain 等五條走到底的收據已經涵蓋這條鏈（docs/verification/player-path/），但這一條當初的用意是「先挑一條窄的走通」，而實際做法是直接走完整條。條目留著沒動。
  完成條件：確認 player-path 的收據已滿足這一條，然後收掉。
  核實：要人判——這一條很可能已經被 player-path 收據滿足了，要人看一眼再決定收掉或改寫。

- **attack-row-and-target-cursor**（#33） — 攻擊式那一列（frame-899）與選目標格畫面的對拍
  選目標格的游標目前畫成一圈白框標位置，原版是拿 sub_1C683 重畫該格的精靈，效果還沒解；繁中 renderer 只把這個既有白框等比放大，未額外宣稱原版外觀已對拍。
  完成條件：兩塊都有逐畫面對拍收據。
  核實：`internal/ui/battle.go`、`internal/ui/battle_cjk.go`、`docs/spec/010-combat-commands.md` 找得到 `確切的視覺效果還沒解出來|原版外觀尚未完全對拍` 就是還沒完

- **target-mode-icon-bar-mouse**（#94） — 選目標格那一層的圖示列怎麼接滑鼠
  選目標格時滑鼠直接指格已接（SPEC-010 §3、sub_24131:0x24332，收據 docs/verification/battle-target-mouse/receipt.json）。同一層畫面下方的圖示列原版是否也收滑鼠點擊還沒證據，remake 目前只收鍵盤，點在場地外的位置不做事。
  完成條件：以 IDA 或 dosgolem 實驗確認原版在選目標格時圖示列的滑鼠行為，remake 照著接並附收據；若原版不收滑鼠，把結論寫進 SPEC-010 並移除 input.go 的自承註解。
  核實：`cmd/kol/input.go` 找得到 `圖示列原版怎麼接滑鼠還沒證據` 就是還沒完

- **longest-cjk-visual-regression**（#40） — 最長繁中內容的視覺回歸、正常路徑截圖與輸入測試
  安全矩形檢查器是靜態的，沒有跑過最長字串的視覺回歸。
  完成條件：最長字串在正常路徑上截圖比對過。
  核實：要人判——視覺回歸的產物是收據。那份收據落地時改綁它。

- **parity-sampling-registry**（#41） — 預先登記近似完成／嚴格 parity 的抽樣母體、硬失敗與平台矩陣
  CLAUDE.md 要求先分開「可啟動」「可玩」「近似完成」「嚴格 parity」並預先登記每階段 gate，不得做到一半才選容易通過的標準。目前沒有這份登記。
  完成條件：四個階段各自的抽樣母體、硬失敗條件與平台矩陣寫成資料。
  核實：要人判——這一份本身就該是 JSON。寫出來之後改成 json_len 綁它的階段數。

- **v01-remaining-battle-chain**（#44） — V01 剩餘：完整裝備、地形、指令及攻擊結算、正常 GUI 戰後保存鏈
  與 v01-full-battle-parity 是同一件事的兩次登記（一次在現行缺口表、一次在最新驗證補充）。
  完成條件：同 v01-full-battle-parity。
  核實：要人判——同 v01-full-battle-parity。這兩條應該一起收掉，或合成一條。

- **battle-silent-steps-inventory**（#96） — 執行段裡沒有訊息就不停的那幾手
  原版每一手都印一句再等鍵（sub_1B100(5)）。remake 的 battleStep 仍保留「這一手沒有訊息就直接往下走」的迴圈，當初是因為敵人那一手與攻擊訊息都還沒接；兩者現已接上，哪些手仍會落進這條迴圈還沒盤點。
  完成條件：列出現行 Act 結果中沒有訊息的手並逐一對照原版；原版有訊息的補上，確定原版也不印的寫進 SPEC-008，最後移除或收窄這條迴圈。
  核實：`cmd/kol/main.go`、`docs/spec/008-combat-field.md` 找得到 `哪些手會落到這裡還沒盤點|迴圈尚未盤點` 就是還沒完

- **ui-details-unverified**（#64） — 五處 UI 細節還沒對拍
  程式裡自承「還沒對拍」的幾處：法術清單的標題座標與版面（`spell.go`）、戰利品畫面（`spoils.go`，要打贏一場而且擲中兩成的掉落才看得到）、選目標格的那一圈與游標樣子（`ui/battle.go` 兩處）、物品頁 `sub_153FF` 模式 8 的擺法（`ui/itempage.go`）。都是畫得出來但沒有與原版逐畫面比過。
  完成條件：五處各有一份逐畫面對拍收據。
  核實：`internal/battle/spell.go`、`internal/battle/spoils.go`、`internal/ui/battle.go`、`internal/ui/itempage.go`、`docs/spec/010-combat-commands.md` 找得到 `還沒對拍|還沒讀|的畫面還沒對` 就是還沒完

- **planning-25-not-verified**（#67） — 規劃段寫進 +0x25 的值沒逐格核對
  分派表認不得的圖號落 default 收尾時，原版寫進 `+0x25` 的是什麼還沒逐格核對過，所以 remake 記下來不動、不假裝做過。
  完成條件：逐格核對過，或確認 remake 目前的做法與原版等價。
  核實：`internal/battle/planning.go` 找得到 `還沒逐格核對過` 就是還沒完

- **planning-other-commands-unwalked**（#84） — 規劃段只走過徒手那一條
  2026-09-11 的規劃對拍只比了「主指令列第 1 格（徒手）→ 攻擊式第 2 格 → 防禦式第 3 格」這一組。射擊、施法、移動、丟撿那幾格，以及其他攻擊式／防禦式的格位都沒有逐格走過。
  完成條件：其餘指令與各攻擊式／防禦式格位逐格走過並比對，或說明為何不需要。
  核實：`docs/verification/planning-parity-20260911.md` 找得到 `沒有逐格走過` 就是還沒完

- **player-path-receipts-stale**（#88） — 正常玩家路徑的收據對不上現在的程式
  2026-09-26 重跑 25 條並逐條處理：已一致或已重生 19 條（含 town、full、quest-list、quest-take 改寫，world-clock、world-quest 依 SPEC-011 §9 重選種子）。過程中修正兩個缺陷：走出建築後隊員格仍畫白框、讀檔後選角停在名冊。剩 6 條：world-healer 原種子已擲不到治療（要驗的就是亂數遭遇，換種子需決策）；flow-quest 的前提（標題流程沿用讀檔隊伍）已不成立；full-chain、full-chain-reload、full-chain-ending、arena 從標題起跑但依賴讀檔隊伍，需改成讀檔開場並以 drive_battle_script.py 重生戰鬥段。分類與證據見 docs/verification/player-path/README.md。
  完成條件：逐條確認現在的行為是對的之後重生收據，或修回原行為；跑 check_player_paths.py --all 全數一致。
  核實：`docs/verification/player-path/README.md` 找得到 `收據過期` 就是還沒完

- **quest-turnin-original-same-state**（#93） — 任務物品交付的原版同狀態與自然玩家鏈
  RE-139 的 IDA 控制流與重製版觀察存檔 GUI 已覆蓋兩次 SPACE、第一件物品交換及全隊回報。原版同狀態畫格／狀態、自然接任務與取得物品，以及回報後保存／重載尚缺。SPEC-014 維持 READY。
  完成條件：以原版與重製版可比初態驗證兩次按鍵、原始物品記錄、六名角色旗標及回報數，再完成自然取得到保存／重載的正常玩家鏈；附版本、輸入、亂數條件與收據後才將 SPEC-014 對應範圍升 CONFORMED。
  核實：`docs/spec/014-quest-completion.md` 找得到 `同狀態畫格／狀態尚未取得前維持 READY` 就是還沒完


### 素材抽取與音訊

- **g06-audio-listening-gate**（#4） — G06：音訊候選人耳驗收、SPEC-004 升 READY 後接正式播放
  SPEC-004 仍是 DRAFT、標「是否授權實作：否」。人耳驗收是升 READY 的閘門，不能跳過。
  完成條件：人耳驗收完成、SPEC-004 升 READY，正式播放路徑接上。
  核實：`docs/spec/004-ogg-audio-catalog.md`、`README.md` 找得到 `狀態：\*\*DRAFT\*\*|還沒接上音訊路徑` 就是還沒完

- **asset-inventory-format-spec**（#22） — 唯讀輸入的資產 inventory 與格式規格
  COMBPICS、TWN0..7PICS 243 筆、MONPICS、PERPICS、MISCPICS、TITLE 五組、ENDPICS、BOOKPICS、KOLMPICS 83 筆、WEATPICS 18 筆、AWD1PICS 24 張、SHLDPICS 四組、HANDPICS 12 組、WEAPPICS 28 組、HEADPICS／TORPICS／LEGSPICS 各 54 組、CHARPICS.0 80 張、KOLPICS 14 筆與三個固定記錄表的 40×38 4-bit EGA 契約已證實；TWN 七筆圖號 1 placeholder 已排除。其餘 archive 與 PCGA／HERC 像素格式尚未完成。
  完成條件：所有 archive 與三種像素格式都有格式規格。
  核實：要人判——PCGA／HERC 還沒開始，沒有可綁的識別字。抽取器支援它們時會多出對應的解碼函式，到時改成 absent 綁那個名字。

- **atlas-and-contact-tables**（#23） — 生成 atlas／接觸表並逐批人工檢查
  769 張 PNG 已抽出並有 manifest。2026-09-10 核實：十九支 tools/probe_extract_*.py 與 extract_battle_tiles.py／extract_ega_assets.py 已經會產 atlas 與接觸表（戰場磚塊 10×8、AWD1PICS 8×3、charpics 等），但 assets/ 底下沒有留存的正式產物——那些是拋棄式探針輸出。缺的是把它們定為正式產物，並逐批人工檢查。
  完成條件：atlas 與接觸表成為版控中的正式產物，尺寸、色盤、透明色、裁切與動畫序列逐批人工檢查過。
  核實：要人判——綁不住：atlas 這個字在十九支探針工具與 manifest 的 coverage_note 裡都有，綁它會判成已完成（2026-09-10 第一版就是這樣誤判的）。正式產物進版控時，改成 absent 綁 assets/ 底下那個檔的實際路徑。

- **pit-square-wave-synth**（#24） — 實作 PIT 方波合成器並完成人耳驗收
  TITLE 的 WAV oracle 已由 DX-CAPTURE 取得（四段活動窗、codec、時長、峰值與 SHA-256 固定）；合成方向已決策為規格近似合成，不散布原版錄音。合成輸出與 DX-CAPTURE WAV 的段落結構比對後才可升 SPEC-004。
  完成條件：合成器實作完成，段落結構比對通過，人耳驗收過。
  核實：`docs/spec/004-ogg-audio-catalog.md` 找得到 `狀態：\*\*DRAFT\*\*` 就是還沒完

- **ogg-conversion-pipeline**（#25） — SPEC-004 升 READY 後建立 OGG 轉換管線
  要驗 codec、時長、非靜音、峰值、loop 接點與人耳聆聽。轉成 OGG 不改變原版素材的著作權。
  完成條件：管線可重跑，每批 OGG 有來源、參數與 SHA-256 收據。
  核實：`docs/spec/004-ogg-audio-catalog.md` 找得到 `狀態：\*\*DRAFT\*\*` 就是還沒完

- **title-thunder-audio-hookup**（#36） — 標題序列的雷聲播放接上 Ebitengine 音訊路徑
  assets/audio/title-thunder.ogg 已產出，title.go 也已標記哪一格之後放雷聲（Thunder 欄位），但 runtime 完全沒有引入 Ebitengine 的音訊套件——一個音都還沒進遊戲。
  卡在：SPEC-004 仍是 DRAFT
  完成條件：正常啟動時聽得到雷聲，且與畫面同步。
  核實：`cmd`、`internal` 找不到 `ebiten/v2/audio` 就是還沒完


### 規格與文件

- **armour-durability-ai-review**（#18） — 護甲減傷／耐久與既知 AI 差異的規格審查，競技場地形另做窄 RE
  耐久子常式九組邊界已重驗；護甲分支映射、骰數與耐久呼叫由 IDA 確認；加值 256 種輸入通過；SPEC-017 升 READY 並接近戰／射擊；84 組記錄、部位與骰值比較通過；破損與 Condition／Record／Carried 已接勝利與撤退回寫。攻擊結算／破損顯示、完整保存、武器超上限與四個其他磨損呼叫仍待驗收。
  完成條件：四個磨損呼叫各自有證據，破損顯示與保存驗過，競技場地形的窄 RE 完成。
  核實：要人判——「四個其他磨損呼叫」要逐支查。接上時會多出四個呼叫點，到時改成 json_len 綁磨損呼叫清單。

- **package-boundaries-and-injection**（#29） — 定義 package 邊界、固定 timestep、可注入 RNG／clock／input 與存檔 schema 草案
  internal/ 已分成 21 個套件，四處已有 Rand interface（title、world、charcreate、battle），存檔 schema 已到 v2。缺的是把邊界與固定 timestep 寫成文件契約——現在只在程式裡成立，沒有一份說得出「哪一層不准依賴哪一層」。
  完成條件：有一份文件說明分層契約與固定 timestep，並且違反時測得出來。
  核實：要人判——程式面已大致成立，缺的是文件與守門。硬綁 timestep 這種猜想中的識別字會永遠為真——那種條目看起來盡責，其實一次都沒真的看過。

- **spec-ready-before-implement**（#38） — 每份玩法 spec 經證據審查升 READY 後才實作，同狀態驗證後才標 CONFORMED
  這是流程規則不是一次性工作。目前 docs/spec/ 底下只有 004 與 028 兩份還是 DRAFT，其餘已升 READY 或部分 READY。
  完成條件：沒有未達 READY 的規格被實作進 production path。
  核實：要人判——流程規則沒有終點，不該用 verify 判它「完成」。留在清單上是為了每一輪被看見；真要收掉就是把它移進 CLAUDE.md 的常駐契約。

- **attack-message-chain-index**（#52） — 攻擊致死的 collapse／shatter 分支
  攻擊訊息索引規則已由 RE-088 解出並接上（SPEC-008 第 6 節）。殘留的是倒下那一句：原版依 `sub_22229` 分出 0x14C「collapse」、依怪物記錄 +0x10 bit 7 分出 0x14E「shatter into pieces!」，remake 的 `DamageLine` 一律走 die。bit 7 已抽成 `Monster.Shatters`；`sub_22229` 在這個分支的語意要先確認。
  完成條件：兩條分支依 RE 證據接上，並有 collapse 與 shatter 各一次的訊息收據；或證明原版現行資料下走不到。
  核實：`internal/battle/message.go`、`docs/spec/008-combat-field.md` 找得到 `那兩個條件的語意還沒定|攻擊致死這兩條分支尚未接上` 就是還沒完

- **sub-26c5f-damage-line**（#53） — 那一句的數字要 sub_26C5F(0x17) 才算得出來
  該常式還沒解讀，所以那一句寧可不報也不報錯的數字。
  完成條件：sub_26C5F 解出，那一句的數字報得出來。
  核實：`cmd/kol/main.go` 找得到 `那一支還沒解讀，寧可不報` 就是還沒完

- **sub-21e0b-filter-contradiction**（#95） — sub_21E0B 掃穿戴欄的過濾條件與 +0x03 設值
  RE-118 記下「種類 byte 大於 0x7F 才算有東西」，又說種類 byte 為 0 的 Fur Boots、Fur Cape、Leather Cape 走 kind 0 （設 +0x04，+0x03 為 0 就設 1）。兩句互相矛盾。若斗篷或靴子確實會設 +0x03，移動方式選單的門檻（commands.go）會變；remake 目前不設 Flag03。
  完成條件：回 IDA 重讀 sub_21E0B 的過濾分支並改正 RE-118；依結論接上或確認 Flag03 在現行資料下恆為 0，同步 SPEC-008。
  核實：`internal/battle/battle.go`、`docs/spec/008-combat-field.md`、`docs/spec/010-combat-commands.md` 找得到 `還要回 IDA 重讀|要回 IDA 重讀|參戰者記錄 \+0x03 的完整語意` 就是還沒完

- **move-unlock-tile-reason**（#57） — 磚塊 0x0C 為什麼解鎖後三個移動方式
  `MoveUnlockTile = 0x0C` 的行為已接（`0x1F88E`），但「為什麼是這一塊」還沒解出來（SPEC-010 末段）。行為正確不代表理由清楚——換一張地圖或一種地形時可能才看得出差別。
  完成條件：理由解出並寫進 SPEC-010。
  核實：`internal/battle/commands.go` 找得到 `為什麼是這一塊還沒解出來` 就是還沒完

- **actor-record-25-reader**（#58） — 參戰者記錄 +0x25 在哪裡被讀
  原版下馬時同時把 `+0x25` 清成 0，但那一格在哪裡被讀還沒解，所以 remake 沒有那一格。SPEC-010 末段記著同一項。
  完成條件：讀取端找出來，並判斷 remake 需不需要那一格。
  核實：`internal/battle/execution.go`、`docs/spec/010-combat-commands.md` 找得到 `它在哪裡被讀還沒解|`\+0x25` 在哪裡被讀還沒解` 就是還沒完

- **sub-1aae8-params**（#59） — sub_1AAE8 的參數語意
  `sub_1AAE8(3, 4, 11h, …)` 的前三個參數語意還沒解出來。SPEC-008 的未解清單列著同一項（選單協定）。
  完成條件：參數語意解出並寫進規格。
  核實：`internal/battle/spell.go` 找得到 `那一支的參數語意還沒解出來` 就是還沒完

- **word-2df2c-outdoor-switch**（#62） — word_2DF2C：城外換整張底圖的那個開關
  在城外會換成第 4 張整張（288×155），`word_2DF2C` 是那個開關，語意還沒解。
  完成條件：開關語意解出並寫進規格。
  核實：`internal/town/layout.go` 找得到 `是那個開關，語意還沒解` 就是還沒完

- **quest-item-0xb926**（#63） — 任務物品的 `[0xB926]`
  SPEC-014 記著這一格還沒解，所以那一條收據裡 `State.QuestItem` 是空的。任務物品本身的記錄與拾取已完成（RE-113／128），缺的是這一格的來源。
  完成條件：`[0xB926]` 解出，那一條的 QuestItem 不再是空的。
  核實：`docs/spec/014-quest-completion.md` 找得到 ``\[0xB926\]` 還沒解` 就是還沒完

- **world-clock-threshold-mechanism**（#65） — 世界時刻門檻的機制沒追出來
  9 步 852 剛好卡在門檻 3367 的兩側（差一），機制還沒追出來，目前靠量到的值。差一的邊界最容易在別的路徑上露餡。
  完成條件：機制追出來並寫進規格，門檻不再是量出來的常數。
  核實：`internal/world/clock.go` 找得到 `機制還沒追出來` 就是還沒完

- **chardata-writer-untraced-fields**（#66） — 角色寫回：幾個位移的來源沒反追
  那些位移目前保留樣板原值——寫進猜出來的 bytes 比留白更糟。要反追出來源才能確定它們在正常遊玩中會不會變。
  完成條件：來源反追出來，或確認那些位移在正常路徑上恆為樣板值。
  核實：`internal/chardata/writer.go` 找得到 `來源尚未反追` 就是還沒完

- **koldata-action-messages**（#68） — 幾個動作敘述還沒對上 KOLDATA
  SPEC-010 末段列了哪幾個指令的動作敘述還沒對上 KOLDATA 的記錄。
  完成條件：那幾個對上，或確認 KOLDATA 裡沒有對應的句子。
  核實：`internal/battle/commands.go` 找得到 `還沒對上 KOLDATA` 就是還沒完

- **reshuffle-behavior-address**（#69） — 重排這一輪的行為沒對上位址
  那一段的行為還沒對上原版位址，remake 取「重排這一輪」；逃走不接在這裡。
  完成條件：位址對上，行為確認。
  核實：`internal/battle/state.go` 找得到 `行為還沒對上位址` 就是還沒完

- **spec-unsolved-sections**（#70） — 各規格末尾的未解章節
  五份規格各自留了一節未解項，內容在產品碼裡多半沒有落點（沒有程式落點就沒有自承可綁），靠這一條追。目前有：`002-ega-extraction-set` 的「尚未關閉的來源」、`004-ogg-audio-catalog` 的「尚未通過的 READY gate」、`008-combat-field` 的「還沒解出來的（不得實作）」（敵人 AI 三張權重表的執行期值、`sub_1B94F`、`word_3B26A` 在城鎮被設成什麼、怯戰檢定的兩個開關、`sub_2625A` 的語意）、`010-combat-commands` 的「還沒關閉的」、`011-wilderness-ambush` 的「還沒解的」（`sub_2A39B`、`word_2FA2A` 那一條分支）。其中已經各自有條目的四項（攻擊訊息索引、參戰者記錄 +0x05、`sub_1AAE8`、`+0x25` 讀取端）不重複列。
  完成條件：五份的未解章節各自清空，或剩下的每一項有自己的條目。
  核實：`docs/spec` 找得到 `^#{2,3} .*(還沒|尚未).*(關閉|解|通過|gate)` 就是還沒完

- **actor-record-2c**（#72） — 參戰者記錄 +0x2C 仍未解
  `SpellEffectField` 的第 5、6 項指到它，但那兩個效果的來源還沒解出來。
  完成條件：+0x2C 的語意解出，兩個效果接上或確認無作用。
  核實：`docs/spec/012-combat-spells.md` 找得到 ``\+0x2C` 仍未解` 就是還沒完

- **quest-count-ff-writer**（#73） — 兩個計數怎麼變成 0xFF 還沒找到寫入點
  所以 remake 用自己的 `VerdictFled` 判定。找不到寫入點就無法確認那個判定與原版等價。
  完成條件：寫入點找出來，或確認 remake 的判定等價。
  核實：`docs/spec/014-quest-completion.md` 找得到 `還沒找到寫入點` 就是還沒完

- **spec029-wiring-review**（#74） — SPEC-029 的接線順序與取圖來源審查
  怪物遠程初始化的規則已實作並通過原版比較，但顯示取圖來源與「正式建構到第一輪」的接線順序還沒審查完，所以規格仍是 DRAFT。
  完成條件：審查完成，SPEC-029 升 READY。
  核實：`docs/spec/029-monster-ranged-initialization.md` 找得到 `仍待審查|接線順序審查` 就是還沒完


### 交付與發行

- **v02-appimage-platform-acceptance**（#7） — V02：主機 FUSE、其他平台與推廣交付的驗收
  Linux AppImage v.0.2.2-20260908 已重建：封包內版號、849 筆必要素材與文件、98 次正常建角／保存操作通過；另完成 XDG_CONFIG_HOME 與 HOME/.config 兩組各 95 次不指定 -save 的保存／重載（docs/verification/appimage-v.0.2.2-20260908.md）。主機 FUSE、其他平台與推廣交付仍待驗。
  完成條件：三平台各自驗收通過，推廣交付有影音驗收收據。
  核實：要人判——跨平台驗收沒有本機訊號。與 windows-macos-real-machine 相鄰但不同：那一條是真機，這一條含 FUSE 與推廣。

- **release-asset-separation**（#27） — 分離本機、private-repo PNG 與可公開輸出
  GitHub Release／公開 release 的 deny-list 要預設失敗即關閉，並檢查 repo visibility 沒有被改成 public。權利審查已寫在 docs/RELEASE-RIGHTS-REVIEW.md。
  完成條件：deny-list 預設失敗即關閉，visibility 檢查納入發行流程。
  核實：要人判——verify_dist.sh 已驗 release 的素材 deny-list，缺的是 visibility 檢查。那一步接上時會多一個 gh api 呼叫，到時改成 absent 綁它。

- **cross-platform-build-smoke**（#37） — Linux／Windows／macOS 可重現編譯與最小啟動 smoke
  build_dist.sh 已建出三平台（linux/amd64 原生、windows/amd64 交叉編、macOS universal 經 osxcross ＋ lipo），verify_dist.sh 驗過 manifest、授權、deny-list、Windows PE 檔頭、macOS 四道靜態驗收與 Linux 啟動。編譯這一半已經成立；缺的是 Windows 與 macOS 的最小啟動 smoke——Linux 容器裡只驗得了結構。
  完成條件：三平台各有一次真的把程式跑起來的 smoke。
  核實：要人判——Windows／macOS 的啟動要在那些平台上跑，本機沒有訊號。與 windows-macos-real-machine 相接：那一條是完整驗收，這一條只要求最小啟動。

- **windows-macos-real-machine**（#42） — Windows／macOS 真機驗收
  Linux 容器 smoke 不冒充真機結果。macOS 的 bundle 沒有簽章（_CodeSignature 在 Linux 上做不出來），首次開啟要右鍵→打開；執行檔本身的 ad-hoc 簽章有驗過。
  完成條件：兩個平台各在真機上啟動並走一次正常路徑。
  核實：`tools/verify_dist.sh` 找得到 `macOS 要真機` 就是還沒完

- **github-release**（#43） — 發 GitHub Release
  權利審查已做完（docs/RELEASE-RIGHTS-REVIEW.md），結論是 release/ 可以公開、full/ 不可以。還缺兩件：說明書譯文要不要進 release/，以及使用者對「建立第一個 Release」這個對外動作的明確同意。
  卡在：等使用者決定
  完成條件：兩件都決定，Release 發出去。
  核實：要人判——對外動作要使用者點頭，不該由 verify 判。

- **dist-rebuild-version-number**（#48） — 由乾淨輸入重建 `dist-all/<版本>/{full,release,promo}` 的正式版本號
  `tools/build_dist.sh` 已能從乾淨工作樹建出三個平台（linux/amd64 原生、windows/amd64 交叉編、macOS universal 經 osxcross ＋ lipo），`promo/` 有 59 秒推廣片與 `sources.json`。缺的只有版本號：目前用的是暫定號，產物可重生，換號碼重跑一次就好。
  卡在：版本號要使用者決定
  完成條件：版本號定案，用它重建一次並驗收。
  核實：要人判——版本號是決策不是程式狀態。`build_dist.sh` 把版本當參數收，綁不到也不該綁。


### 本輪核實為已完成（原本掛在未完成清單上）

- **g01-cjk-sheet-pages** — G01：繁中資料頁與 MIRROR／物品／隊伍／法術／任務入口：2026-09-11 接上：中文模式的角色資料頁開得起來，而且**版面直接用 `town.SheetLayout × book.CJKScale`，沒有第二份座標表**。倍率從 2 改成 3（畫布 960×600）是算出來的：中文字模格高 18，原版最小行距 6，`6 × 3` 剛好；兩倍只給得起 12，那才是原本要重排、要把抬頭六列併成四列的原因。三倍之後抬頭六列、屬性八列、武器四列、金幣三列與兩條長條全部留在原位。標籤取記錄 0x73 那十四段的譯文（早就翻好了）。畫面在 docs/screenshots/zh-town-sheet.png，倍率判準的測試在 internal/ui/sheet_cjk_test.go（帶反面對照：兩倍要放不下），verify_text_fit 6,937 項全過。
- **d01-spec-stale-clauses** — D01：逐項清除 SPEC-005、008、010～015 的過時禁作與未決條款：已完成
- **a01-a02-identity-and-save** — A01／A02：角色身分與保存規格整合、建角逐層取消、名冊保存與實際選隊：2026-09-26 v.0.2.5 AppImage 真實按鍵：建兩名、各自入隊、住宿離隊後重載各自再選進旅店、刪除不搬槽、空槽重用與滿槽拒絕皆通過；兩種預設存檔位置另通過。收據 docs/verification/identity-save-chain/receipt.json。
- **a07-a08-cjk-and-inn** — A07／A08：中文資料頁、戰鬥中文字型、開場旅店五個圖示與忠實住宿退出：2026-09-11 接上：中文模式的角色資料頁開得起來，而且**版面直接用 `town.SheetLayout × book.CJKScale`，沒有第二份座標表**。倍率從 2 改成 3（畫布 960×600）是算出來的：中文字模格高 18，原版最小行距 6，`6 × 3` 剛好；兩倍只給得起 12，那才是原本要重排、要把抬頭六列併成四列的原因。三倍之後抬頭六列、屬性八列、武器四列、金幣三列與兩條長條全部留在原位。標籤取記錄 0x73 那十四段的譯文（早就翻好了）。畫面在 docs/screenshots/zh-town-sheet.png，倍率判準的測試在 internal/ui/sheet_cjk_test.go（帶反面對照：兩倍要放不下），verify_text_fit 6,937 項全過。
- **spec-backfill-stale-clauses** — 回填 SPEC-008／010／011／012／014／015 的過時禁作條款與現況：已完成
- **spec-manual-page-links** — 讓規則 spec 連回說明書頁碼：2026-09-26 25 份規則規格（005～029）末尾加「說明書對照」表，逐段列出 PDF 頁、印刷頁與 docs/manual/chapters/ 章節連結，並寫明說明書沒寫、改以原版證據為準的部分；015、028、029 明示說明書無對應。001～004（工具與素材契約）與 030～035（中文版面）不屬規則規格。連結全數可解析。
- **runtime-format-guard-test** — 建立測試，保證 runtime 只會讀 PNG 與 OGG：2026-09-26 靜態守門（容器沒有 strace、Go 不經 libc）：main.go 旗標預設路徑須在 assets/ 或 docs/ 且為 PNG／JSON／TSV／OGG 或目錄；cmd 與 internal 非測試程式碼不得有像原版檔名的字面值；assets.Decode 拒絕非 PNG、逃出目錄與內容非 PNG。自我負例與把旗標改成原版封裝的真實負例皆失敗。
- **terminology-and-safe-rects** — 統一遊戲與說明書術語；每個 CJK 欄位定義文字安全矩形與 overflow policy：已完成
- **appimage-rebuild-latest-fixes** — 把最新怪物初始化與面板修正重建到新的本機 AppImage：2026-09-25 從 commit 5a0c0ad 重建 v.0.2.4-20260925，含 SPEC-029 怪物初始化、面板與本輪繁中修正；完整保存 GUI、封包 851 筆雜湊與兩種預設存檔位置通過，README 交付入口已指向它。封包內戰鬥勝利未重跑。
- **train-forge-entry-condition** — TRAIN 的鍛造（旗標 byte 1 的 bit3）：已決定不做：2026-09-11 依 RE-134 結案：**原版走不到**。判準是物品記錄 +0x0F 的 bit5，而三個母體（八座城商品表、ITEM.LST 60 筆、任務物品 24 筆）逐筆掃過之後，唯一叫 Ingot 的東西（城 5 的 Copper Ingot）bit5 是零，有 bit5 的全是 Cloth Pants 與 Leather Hood。說明書寫的是設計意圖，資料沒跟上。remake 維持不實作，理由從「條件不明」變成「原版走不到」。
- **dist-package-sha256-verification** — 發行包的 SHA-256 驗收：2026-09-26 verify_packages.py 新增 check_manifest：三包逐檔重算 SHA-256 與大小，並抓清單外的多餘檔與清單列了卻不存在的檔；verify_dist.sh 經由它執行，本機 AppImage 驗收也呼叫 --manifests-only。v.0.2.5 三包通過；竄改、多檔、少檔三個負例皆失敗。舊版 0.2.1 的 full 因 2026-09-04 打包工具排除所有 manifest.json，漏列 assets/ui-font/manifest.json，新檢查會報出，舊交付未改。
- **world-battle-click-regions** — 野外與戰鬥兩層的滑鼠可點區域：2026-09-11 接上：野外的圖示列、戰鬥規劃段的指令列、戰利品畫面的兩張清單與圖示列都走既有的 iconTargets／input.Row，token 與鍵盤同一套。驗收在 cmd/kol/world_battle_mouse_test.go（四支），正反對照過——拆掉「露營時不列」與「選目標格不列」兩道閘門測試會開口。**選目標格那一層故意不列**：原版那裡收的是座標直接指格（SPEC-010 §3 的 X 0x130…0x250、Y 8…0x98），語意是「指這一格」不是「搬白框幾步」，用 token 序列表達會變成另一回事；另立 battle-target-cursor-click 追。
- **inn-disband-exit** — 正式的離開遊戲：在旅店把人解散：2026-09-11 核實：town 與 flow 兩層都已接上——住宿即存檔即離隊、隊伍空了問記錄 0x36、答 YES 由 flow 轉回書本目錄（對應原版的 sub_2DB2A("kolbook.exe", 0)）。town 那一側原本就有 TestInnStayQuitsWhenPartyEmpties，這一輪補上 flow 的銜接測試 TestInnDisbandReturnsToContents（含反面對照：quit 只能取走一次）。Ctrl+Q 留著當額外退路——人不一定停在旅店裡。
- **enemy-move-direction-unhooked** — 敵人移動段的方向規則：解出來了但還沒接上：已完成
- **counterscan-style-and-retry** — 反制掃描的攻擊選項來源，與移動繞不出去的重試路徑：2026-09-11 核實：兩條**都已經在程式裡**，過期的是 enemyStrike 頭上那兩句註解。抄對手那格是 counterScan 的 bestStyle（三支測試護著）；重試路徑是 enemyStuck 的 a.Stuck++ 與 enemyStrike 的 `if a.Stuck == 0`——後者原本沒有測試，這一輪補上 TestEnemyStrikeFixesStyleAfterBeingStuck（含反面對照：沒困住過時同一顆骰子要給出不同的攻擊式）與 TestEnemyStuckCountsUpAndRetries。
- **actor-record-05-ability** — 參戰者記錄 +0x05：哪一件裝備、哪一個位元賦予的能力：已完成
- **victory-drop-sub-289ee** — 打贏之後掉落的東西：2026-09-11 核實：早就做了，過期的是 state.go 的註解。DropSpoils（spoils.go）是 sub_289EE(0) 的前半，三支測試護著，execution.go 的打贏那一段照原版順序呼叫：任務完成 → 冒險點數 → 金冠 → 掉落物。
- **town-typing-input-box** — 城鎮打字問話的輸入框：2026-09-26 互動遊玩以 Xvfb＋xdotool 真實按鍵在旅店 MOUTH 打字、退格、送出問中 ELVE、空問句，英中輸入序列相同，十二張畫格收據；修正空問句送出後游標殘留（RE-114 讀入函式收工前蓋掉橫槓）。腳本 token 版本在第 7 城民宅另驗退格與空問句，英中軌跡相同。空問句後白框位置未對拍。
- **license-into-packages** — 打包時把 LICENSE 複製進每一個發行包與 dist-all/<版本>/ 根目錄：2026-09-10 核實：tools/build_dist.sh:51-54 已把 LICENSE 複製進 dist-all 根、promo/ 與每個子目錄；tools/package_dist.sh:110 的打包清單也帶 LICENSE。
- **sheet-items-and-party-portrait** — 角色資料頁的物品（sub_16986）與隊伍立繪（sub_18D61）兩個圖示：2026-09-10 核實：internal/town/itempage.go 的 OpenItemPage 與 internal/town/portrait.go 的 OpenPartyPage 都已實作，立繪疊法在 internal/portrait/（RE-126／127／128）。殘留的 s.note(sheetIconName(...)) 只是資料缺失時的退路，不是未實作標記。自評表也已記為做完，只有這一條沒回頭改。
- **party-limit-six-boundary** — 隊伍上限六人的邊界：2026-09-10 核實：internal/town/interior.go:64 的 const MaxParty = 6 已在 recruit.go:291（招募）、interior.go:44（進門）與 save/characters.go:191（保存）三處把關。
- **extraction-docker-image** — 先辨識既有專用 Docker image；缺件才建立有鎖版來源的 Dockerfile：2026-09-10 核實：既有共用映像 eob-remake-go:1.26.7-ebiten2.9.9-audio 已由 tools/go_docker.sh 與 tools/run_python_tool.sh 使用（只讀不改）；缺件的兩個已建 docker/Dockerfile.osxcross-go 與 docker/Dockerfile.package。
- **wilderness-cloud-animation** — 野外的雲層動畫：2026-09-10 核實：已接進畫面。`internal/world/weather.go:69` 的 Composite 做旋轉，`internal/ui/world.go:45` 的 drawWeather 畫進世界畫面（144／217 兩處呼叫 CompositeGroup(s.Group, s.Phase)），`cmd/kol/main.go:937` 遞增 Phase。G02 那一條已記為做完（SPEC-026 六組天候、128 像素捲雲與正式節拍，正常出城 GUI 只在天空區改變），只有這一條沒回頭改。雨雪的正常 GUI 與原版同相位另見 weather-animation-and-speed。
- **monster-unarmed-weapon-source** — 怪物徒手那幾手改成照原版：拿手上那件：2026-09-10 依使用者決定照原版實作：`Actor.AttackWeapon` 在 `Side == SideMonster` 時一律回手上那件，`NewQuest` 不再填怪物的 UnarmedWeapon，`unarmedFrom` 連同它唯一的使用者一起移除。`TestMonsterUnarmedFromWorn` 釘的是被推翻的行為，改寫成`TestMonsterUnarmedUsesHeldWeapon`（含隊員反面對照與空手怪物邊界）。SPEC-008 與 SPEC-012 的「兩側都接上了」「不是只完成隊員一側」兩處錯誤斷言已改寫成現況。
- **save-v1-entry-replacement** — SPEC-023 的保存核心已是正式入口：2026-09-10 核實：`save.go:17` 寫明「正式遊戲使用 CompleteSchema，v1 保留為匯入入口」，`cmd/kol/complete_save.go` 用的就是 `save.Characters`。v1 是刻意保留的匯入路徑，不是還沒取代——`characters.go` 那句「尚未取代正式 v1 入口」是過期斷言，已刪除。
- **region-dat-provenance** — `region1.dat`…`region4.dat` 缺檔的來由：2026-09-11 核實：那五筆是擴充地區，由地區磁片經原版內建的 REGION INSTALLER 裝入，Origin 未曾發行擴充。執行時觀測（開檔序列、模板位址與 KOLBOOK.EXE 靜態偏移逐項對位、原版自己的訊息文字）見 docs/re/20260911-region-installer.md；探針在 tools/oracle/book_startup_test.go 的 KOL_BOOK_REGION_PROBE 模式。
- **planning-current-slot-semantics** — 規劃逐鍵時 DS:0x9E5A 的語意：2026-09-11 定案：它是走訪參戰者的**索引游標**（IDA `word_37D2A`），規劃、執行與訊息三段各自的迴圈都會設它，不是「輪到誰行動」的權威狀態。寫入監看抓到五個寫入點，其中 0x1F073 正是 SPEC-020 步驟一的收尾。remake 不需要它——輪到誰由 Round.Order 與 Cursor 表達。見 docs/re/20260911-actor-index-cursor.md。
- **battle-target-cursor-click** — 選目標格那一層的滑鼠指格：2026-09-26 依 sub_24131:0x24332..0x24543 的 IDA 傾印接上滑鼠指格：嚴格區間換格、位移走方向鍵同段夾限（mode 1 中心再走一步、夾 3…5；mode 2 夾 0…8、內圈還原）、位移為零不動、戰場外點擊不觸發 Return。射擊重播把方向鍵 X 換成 CELL:8,8，英中各自軌跡逐位元相同；近戰夾限與邊界像素由單元測試驗。
- **cjk-sheet-subpages** — 資料頁的坐騎與頂端隊員格的中文版面：docs/verification/cjk-ui/stable-six-page-receipt.json；六張繁中畫格及逐位元相同的英中軌跡。
- **screenshot-png-nondeterminism** — 重生截圖會讓無關的英文圖出現 diff：2026-09-26 鎖定映像下以 run_kol_shots.sh 把 shots.spec 全部 46 行各跑兩次，46 張逐位元相同，重生是決定性的；「內容相同但位元不同」無法重現。與 docs/screenshots 收據不同的 9 張全是內容真的改變：en-newchar-picture／stats 為擲骰對上原版後同 seed 數值不同，en-world-map／town 為 SPEC-026 捲雲位置，zh-book-cover／roster／delete、zh-town-items／sheet 為本週中文化。判準（有對拍收據者須逐位元相同）不必改；這 9 張依使用者指示待整體驗收後隨 README 一併重生。
- **path-verify-does-not-replay** — 路徑驗證只看收據自洽，不重跑：2026-09-11 做了 `tools/check_player_paths.py`：重跑到 workplace/replay-check/ 再與收據逐 byte 比，不覆寫。第一次跑就抓到 10 條對不上。`verify_full_path.py` 保留原用途（驗收據自洽），兩支分工不同。
- **cjk-battle-remaining-pages** — 補戰後戰利品清單的繁中正常路徑畫格：法術／效果兩頁及戰利品 Items Found／Characters 兩頁均已擷取並目視檢查。戰利品由 seed 39 固定輸入在重製版正常按鍵路徑第 22 回合勝利後取得；收據含輸入、程式雜湊、PNG／trace 雜湊與觀察用資料限制。本項只驗中文版面；完整戰鬥、原版同狀態與正式封包另列未完成。
- **town-untranslated-source-slot-reachability** — 城鎮 45 個未譯來源槽位的全域可達性與中文處置：2026-09-25 RE-140：訊息槽位與建築 payload 共用 DGROUP:0x1CF4，45 筆分為 40 筆馬廄坐騎記錄（已證實）、1 筆商店殘留資料（強推論）、4 筆無讀取點的任務續文（強推論）。依使用者選定處置不翻譯，列入 docs/i18n/town-nontext-slots.json 並綁原文快照；嚴格檢查 2,076 則 0 缺譯、另排除 45 筆，三個負例皆失敗。

<!-- worklist:end -->

## 現行自評與可執行缺口（本表優先於下方歷史條目）

自評 **75／100**；權重、RE／SPEC／程式落點與驗證限制見
[2026-09-08 稽核](docs/verification/spec-source-audit-20260908.md)。

> 未完成項 `g01-cjk-sheet-pages`：G01：繁中資料頁與 MIRROR／物品／隊伍／法術／任務入口。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
- [x] G02：SPEC-026 已接六組天候、128 像素捲雲及正式節拍；dosgolem 四組完整合成與六組分派通過，正常出城 GUI 只在天空區改變。雨雪正常 GUI 及原版正常野外同相位仍未宣稱。
> 未完成項 `g03-arena-vertical-chain`：G03：競技場正常取得資格與完整垂直鏈。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
- [x] G04：SPEC-024、告別確認已實作；dosgolem 單人／多人正常離店與繁中 GUI 已抽驗。不是全城逐店文字驗收。
> 未完成項 `g05-picture-shield-editor-spec`：G05：圖像／盾牌編輯另立規格。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `g06-audio-listening-gate`：G06：音訊候選人耳驗收、SPEC-004 升 READY 後接正式播放。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `g07-battle-speed-options`：G07：三個加速選項的詳細規格、控制與保存。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `v01-full-battle-parity`：V01：完整戰鬥同狀態結算、武器／護甲與保存的正常垂直鏈。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `v02-appimage-platform-acceptance`：V02：主機 FUSE、其他平台與推廣交付的驗收。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `d01-spec-stale-clauses`：D01：逐項清除 SPEC-005、008、010～015 的過時禁作與未決條款。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。

舊 A01／A02 建角與名冊保存已完成；A03～A06／A09 的已知回寫、腰帶、Ward、
Nutrition 與保存缺陷已修。A07 剩繁中資料頁；A08 開場旅店及住宿退出已接，
告別差異另列 G04。下方舊勾選不得把它們重新列為完全未實作。

本輪優先順序：功能與保存缺口 → 正常玩家流程 → 新版 AppImage 驗收。
使用者已要求停止追查 RND() 重擲與精確骰序；既有證據保留，此類追查不再列為收尾閘門。

- [x] 現行存檔設定拒絕非物件／非布林內容，重複解析不殘留舊設定；存檔模組回歸通過。
- [x] 完整角色／名冊保存：正式 v2、保留原檔的 v1 匯入、建角回填、
  逐人住宿交易及世界時間已接線。完整回歸、建置與正常 GUI 建角／重啟／住宿／
  刪除重用槽位通過；見 `docs/verification/complete-save-integration-20260908.md`。
  此項不包含新版 AppImage 打包或原版戰鬥同狀態驗收。
> 未完成項 `battle-normal-seed-alignment`：原版戰鬥同狀態：對齊地形、完整屬性與結算比較。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `inn-farewell-diff-review`：審查原版離店告別訊息與確認流程。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。

### dosgolem 更新後驗證

- [x] 敵人遇遮擋後同環續查已修，完整測試／建置通過；見 `docs/verification/ai-ring-20260908.md`。
- [x] sub_2531C 七環表、起點骰與續查已由 dosgolem 驗證；修正骰面並訂正敵人
  射擊門檻為 4，完整測試／建置通過。見 `docs/verification/ring-search-20260908.md`。
- [x] counterScan 獨立起點已核對並修正，14 組原版與回歸、完整測試及建置通過。
> 未完成項 `dosgolem-gui-ai-acceptance`：補新版正常 GUI 與玩家可見 AI 行為驗收。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。

- [x] 戰鬥丟棄／拾取／擊落／滿背包戰利品交換的 Member.Weapon 同步已補，
  撿弓後射擊指令及彈藥資訊回歸、完整測試／建置通過；正常 GUI 拾取異種武器仍待驗收。
  見 `docs/verification/weapon-sync-20260908.md`；敵人拾取及同環尋敵未因此標為完成。

- [x] SPEC-022 隊員射擊彈藥消耗已修，2,550 組原版直接比較及完整測試／建置通過。
  資訊框已接線，128 組原版顯示比較、六組文字流程及正常 GUI 150 個輸入通過，
  Olear 射擊由 20 發更新為 19 發。弩與耗盡的正常 GUI 仍待驗收，見 `docs/verification/ammunition-panel-20260908.md`。

- [x] SPEC-020 規劃門檻已接線，含空隊員槽、倒下人物與取消保持；完整測試及建置通過。正常操作另行驗收，不再要求精確骰序，見 `docs/verification/planning-roll-implementation-20260908.md`。

- [x] SPEC-019 基本肖像／裝備切換已接線，原版常式及正常 GUI 126 個輸入驗證通過；彈藥數字與完整 A10 仍未完成，見 `docs/verification/equipment-panel-implementation-20260908.md`。

- [x] 原版人形遮罩 144 組與直條十二組已直接比較正式 remake 計算，底圖路由亦通過；人形／疲勞框已接線，A10 其餘內容仍未完成。

- [x] `all` 補納護甲／耐久與資訊框直條測試，連同核心、六項規則及正常書本入口重跑通過；見 `docs/verification/dosgolem-complete-suite-20260908.md`。

- [x] 現用 `kol-force-call` 快轉至 `fd746405`，KOL 改用新版診斷欄位，保留失敗追蹤。
- [x] dosgolem 核心測試與六項 KOL 原版規則對照通過。
- [x] 正常入口 RGB 顯示：補上 BIOS 0Dh／0Eh 色盤初始化，六張 dosgolem 重生圖逐張確認。
- [x] 正常 `KOL.EXE` 啟動鏈經 DOS Enter 進入 `KOLBOOK.EXE`，確認書本封面。
- [x] 修正 DOS AH06 非阻塞讀鍵永遠回無字元的獨立缺陷；核心與六項規則再驗通過。
- [x] 依 IDA 證實並修復 BDA 檢查與 DOS AH08 取鍵的跨介面缺口；正常書本選角進入 Halfway House，截圖確認。
> 未完成項 `appimage-parity-continuation`：旅店、戰鬥與存讀檔的 AppImage 同狀態對拍。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。

更新後工具鏈狀態與限制見 [書本讀鍵修復](docs/re/20260908-kolbook-input-bridge.md)；前次色盤修復紀錄保留歷史證據。

## 目前狀態

2026-09-07 使用者確認：滑鼠可取代鍵盤，以操作結果等價為準。當前工作轉為
原始碼／RE 規格稽核，**核心功能仍有未完成接線**，不是只剩影音美化。
完整目前狀態表、九個失敗重現與逐規格矩陣見
[稽核報告](docs/verification/spec-source-audit-20260907.md)。

### 當前可執行修復項目（第一批已修，整體未完成）

2026-09-07 第一批修復與最新驗證見
[修復紀錄](docs/verification/spec-fixes-20260907.md)。建角入口／表單／取消、
同城選隊顯示、勝利／撤退物品與傷勢回寫、腰帶、Ward、修道院已修並通過回歸。
完整存檔與名冊保存已由 SPEC-023 接線；中文資料頁及旅店其餘功能驗收仍未完成。
下方大項不因局部修復而勾選；現行 AppImage 尚未更新。

2026-09-08 補修：選隊及招募建立獨立隊員資料，避免遊玩修改透過共用指標污染名冊；
招募同步戶外隊伍。開場旅店已依 SPEC-016 以角色所在建築接入正式 Session，
並補上繁中候選人清單；跨城選角亦已接通並以 Mula 正常 GUI 驗證。
中文資料頁及完整操作對拍仍待完成，見 `docs/verification/cross-town-fix-20260908.md`。
dosgolem 已補驗 Mula 跨城入口，人工確認角色、旅店、店主肖像、招呼語及圖示順序相符；
原版招募／住宿及 AppImage 封包仍須各自驗收。
招募候選人滑鼠列選擇已補接；繁中 GUI 已驗證返回圖示下點非反白列 Olear、入隊及離店。
英文與捲動清單的 GUI 抽測仍待完成。
正式翻譯字庫涵蓋檢查：容器內執行 `python3 tools/verify_cjk_coverage.py --root .`。
目前 5,109 個字串無可見缺字，隔離副本負向驗證通過；不涵蓋 Go 動態文字、墨跡與排版。

> 未完成項 `a01-a02-identity-and-save`：A01／A02：角色身分與保存規格整合、建角逐層取消、名冊保存與實際選隊。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `a03-a04-writeback-and-contract`：A03／A04：戰後傷勢、裝備、戰利品與任務物品回寫；可變狀態保存契約。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `a05-a06-a09-belt-ward-nutrition`：A05／A06／A09：腰帶轉接與武器數值交換、14 名角色 Ward 傳遞、修道院 Nutrition。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `a07-a08-cjk-and-inn`：A07／A08：中文資料頁、戰鬥中文字型、開場旅店五個圖示與忠實住宿退出。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `a10-battle-info-panel`：A10：戰鬥左側完整資訊內容的原版同狀態。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `armour-durability-ai-review`：護甲減傷／耐久與既知 AI 差異的規格審查，競技場地形另做窄 RE。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `spec-backfill-stale-clauses`：回填 SPEC-008／010／011／012／014／015 的過時禁作條款與現況。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `weather-animation-and-speed`：接正式天候動畫及既定加速選項。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。

### 前次逐系統摘要（歷史基準，不覆蓋本輪稽核）

| 分類 | 狀態 | 目前證據／限制 |
|---|---|---|
| remake | **一串按鍵從標題走到住宿存檔**，讀回來還能繼續走到任務地點與結局 | SPEC-002／005／006 READY，769 張 PNG；二十六份收據共 285 對逐畫面通過，允許區外零像素差；四條走到底的路徑收據見 `docs/verification/player-path/` |
| 戰鬥 | 一場**雙向**打得完，打贏有冒險點數、金冠與戰利品清單；丟、撿、切換、擊退／打暈、施法、射擊與逃走都做了 | SPEC-008／009／010 的已解出部分與 SPEC-012／013／014 READY；`full-chain` 在正常路徑上打贏了一場遭遇（八個回合）；**敵人傷得夠重會撤退、聰明的還會反制掃描與繞側翼、空手的會先去撿一件、繞不出去時走三條退路**（RE-117）；**看不到的動作不印訊息**（RE-124）；逃走之後掉東西原版就沒有（RE-124）|
| 原版 oracle | build 已固定；**規則可以逐支對拍了** | ZIP 與全部項目逐檔雜湊已固定；發行來源／1998 timestamp 成因未知。dosgolem 跑得動 `KOLTOWN-unpacked.EXE`：亂數 6 個種子 × 500 抽逐次相同（RE-132），戰鬥數字鏈三支規則直接呼叫原版對拍（RE-133，`internal/oracle`，`-tags oracle`）。方法是「寫記憶體設好輸入 → 釘死骰值 → 呼叫常式讀結果」，不必驅動遊戲 |
| 收據基準 | **13/15 已知失效，待重生** | 換成原版那一支亂數之後遭遇的觸發步數位移；根因與重生流程見下面「收據基準」一節。**不是 bug，也不假裝通過** |
| 玩家觀點 | 起源資料完成 | 魔旅士繁中個人心得＋《CRPG 通鑑》中文長篇評介已進 README |
| 繁中介面 | 書本、城鎮、野外的畫面；城鎮對話與三個 UI 文字庫的譯文 | 雙模式；中文字模表 2,434 字、缺字 0；文字安全矩形 6,685 項全過；UI 文字庫 486 筆譯文全過檢查器 |
| 城鎮文字母體 | 已解出，建築文字全部翻完 | 5-bit 打包，7 城 101 棟建築共 9,515 行（RE-062）；名稱、人物、進門敘述與 1,938 則建築內對話已全部翻成繁中 |
| 可選美化 | 未決策 | 不得在忠實度與操作方向確認前進入 production |
| 圖像 theme | 已決策 | 2026-09-01：首版 EGA-only；CGA／Hercules 列後續選項 |
| TITLE 音訊 | 候選資產待人耳驗收 | 規格近似合成已產出；時長與原版擷取差 0.6 ms；SPEC-004 仍 DRAFT |
| GitHub | 已建立 | private：`wicanr2/kol-remake-cht`；`main` 已推送 |
| 授權 | 已定案 | 2026-09-04：RRSAL-1.0，`LICENSE` 與 README 授權段已建立；發行包要帶條款那一項還沒做 |

## 歷史自評分數（不作現行完成判定）

**90 / 100**（2026-09-06）。

分數的算法寫在下面，好讓它可以被反駁而不是只能相信。**權重是判斷，不是量出來的**
——依據是各系統在原版遊戲裡佔玩家時間的份量，以及說明書給它的篇幅。**完成度盡量
有依據**：能指到 spec 狀態、對拍收據或程式位置的就指，指不到的寫「無」。

| 系統 | 權重 | 完成度 | 依據 | 還缺 |
|---|---:|---:|---|---|
| 標題與開場 | 3 | 95% | RE-047，18 對翻頁對拍全過 | 雷聲音訊沒接上 |
| 書本 UI | 10 | 85% | SPEC-005 READY，建角／清單／刪除／HARBOR 共 200 餘對對拍 | `EDIT PICTURE`／`EDIT SHIELD`（原版滑鼠驅動，oracle 送不了滑鼠事件） |
| 城鎮 | 18 | 99% | RE-053…RE-076，旅店／戶外／建築／對話／招募／住宿／鏡子／商店／訓練都有收據；**MOUTH 的打字問話與資料頁的任務一覽也做了**（RE-067／RE-111），**打字那一格的版面也對拍過**（RE-114，收據 `town-mouth-run1` ＋ `plan-town-mouth.json`）；**馬廄的買賣與 SCROLL 的入會、買法術也接上了**（RE-119／120／121，收據 `town-stable`、`town-scroll`）；**角色頁的物品與隊伍立繪（RE-126／127）與法術編輯器（RE-128）也做完了，這三頁都整頁對拍到允許區外零差異** | 鍛造 |
| 野外 | 9 | 98% | RE-066…RE-069，四對地圖對拍＋循路收據；隨機遭遇整條打得起來（SPEC-011，實跑收據 `battle-ambush`）；**騎走那一問與那一隻的四行敘述也做了**（RE-112，收據 `battle-rideaway`） **治療那一條也接上了**（RE-112／128：六段敘述在 `TOWN0` payload 4，答應之後全隊傷勢與飢餓歸零；實跑收據 `world-healer`） | 雲層動畫（版面已定案並驗過，RE-136，還沒接進畫面）、那六段敘述的中文譯文 |
| 戰鬥 | 25 | 99% | 一場戰鬥**雙向**打得完；戰後的冒險點數、金冠與戰利品清單都做了；丟、撿、切換三手、擊退／打暈、施法、射擊（含彈道與上弦那一手）與逃走那一問也做了；**敵人傷得夠重會撤退**（RE-081：兩個門檻來自怪物記錄，不死系永遠不逃）；打贏的冒險點數與金冠會寫回角色，競技場那一場照原版整段跳過（RE-096）；貼近打那一手會填招架姿態（`+0x24` 是「命中部位 ＋1」）；**聰明的怪物會反制掃描與繞側翼**（`sub_253E7`／`sub_25653`，RE-104）；**撿之前會問**（記錄 0x111 的 y／n，RE-113，收據 `battle-pickup`）；**空手的怪物會先去撿一件**（RE-116）；戰利品畫面整塊照原版做完（RE-115：清單座標、選中列換色、底下那兩格圖示）；**撿起來的東西會用自己的傷害骰**；**下馬那一手也做了**（RE-124：0x28 印「lands!」並清掉 +0x1E，那一格有別人就印 0x11A；上馬本來就在移動方式第 3 式）；**逃走之後掉東西那一條原版就沒有**（RE-124：掉落機率 `byte_3AC7D` 在 BSS 裡、整份反組譯零個寫入點，恆為 0）**移動段的三條退路也接上了**（RE-117：朝目標試三個方向、沒目標從隨機方向繞一圈、繞不出去就原地揮武器／撤退／重跑決策樹，`word_34A08` 非零時攻擊式固定第 1 式）；**訊息的可見性過濾也做了**（RE-124：`sub_2625A` 是「隊員 ±6 格內看得到才印」，看不到就整段跳過，連重畫都不做）**任務物品開打就擺在戰場上**（RE-113／128：座標是任務地點記錄的 `+0x27` 列、`+0x28` 欄，走到那一格才撿得到）| — |
| 角色成長 | 7 | 92% | AWARD 訓練已對拍（RE-076）；冒險點數的來源與發放已解（RE-096）；**競技場整條走得到**：選人、進賭場下注、單挑打完升一階並結算賭金（RE-108／RE-098，實跑收據 `arena`）；打贏的冒險點數與金冠會寫回角色並進存檔（RE-096）| 競技場的遭遇訊息與戰場地形沒有對拍 |
| 法術 | 10 | 97% | SPEC-012 READY：擲骰、擋招、效果、倒數、資料與兩張清單都做完並實跑（`battle-spell` 收據）；**法術的母表解出來並接上了**——SCROLL 那一格賣的八筆就是它（RE-121，收據 `town-scroll`）；**效果 7／8 是攻防技能的罰值**，一點抵十、過期不還（RE-129）；**疲勞那一條鏈整條接上了**（RE-130／RE-131：效果 5 的正負是反過來的、休息回復三成、每一手把花剩的時間額度除以十扣進疲勞，力竭會當場放倒人），八個效果全部有實作；**選部位那一次互動也做了**（RE-100：骰在問之前擲完，`PromptArea` 只認 1…6）| 選部位那一問還沒有實跑收據（要先買一筆效果 6 而且不用選目標格的法術，SPEC-012 第 8 節）|
| 任務與結局 | 8 | 97% | 任務觸發與怪物來源已解（RE-090）；完成的標記已解並實作（RE-101），帶著東西逃掉也算完成（RE-103）；**接任務、交付、回報與結局都在正常路徑上實跑通過**（RE-106／107／109／110）：`quest-take` 按 MOUTH 打字問到關鍵字就接下任務（沒換過任何輸入），`quest-list` 在資料頁上列出回報過的，`full-chain-ending` 打完最終任務那一刻播出七段結局，與 `-scene ending` 的收據逐 byte 相同 | 一覽底下那一頁（`sub_16239`）沒有對拍 |
| 音訊 | 3 | 10% | 母體已普查，TITLE 的 WAV oracle 已取得 | SPEC-004 還是 DRAFT，一個音都還沒進遊戲 |
| 繁體中文化 | 5 | 88% | 城鎮 1,938 則對話 ＋ 三個 UI 文字庫 486 筆全翻完，安全矩形 6,952 項全過、缺字 0；**ENDTXT 接上結局畫面、BOOKDATA 接上書本**（`LoadBank` 把文字庫疊到 remake 自己的字串表上），KOLDATA 畫在對話框裡的每一行也逐行量過 | 戰鬥那一段（索引 ≥ 0x100）的中文版面還沒定；每一筆記錄一次畫幾行要等接上畫面才驗得了 |
| 互動遊玩 | 3 | 90% | `kol` 不帶旗標就是互動遊玩（`-scene game`）：鍵盤與滑鼠從標題走到書本、建角、選角、旅店、城鎮、野外與戰鬥，走的是 `flow.Key`／`playToken` 這條收據驗過的 token 路徑。滑鼠翻成既有 token，可點區域取自既有版面常數。驗收 `tools/verify_interactive.sh` 在 Xvfb 裡送鍵與滑鼠比畫面雜湊，判準是**滑鼠點第 N 列要與鍵盤走到第 N 列同一張畫面**，一路驗到滑鼠點 DOOR 出門並在城鎮裡走一步；同一套驗收也直接跑過 AppImage | 開場旅店只有 DOOR 接得動（其餘五個圖示走 `f.note` 記一筆就不動）；戰鬥與野外還沒有可點區域，只有「點一下等於 Return」與滾輪；`-load` 沒有預設值 |
| 發行打包 | 2 | 95% | `tools/build_dist.sh` 從乾淨工作樹建出 `dist-all/0.2.1/`，`tools/package_dist.sh` 再把它封成**三個平台各自獨立的六個發行檔**：`full`／`release` 各有 Linux AppImage、Windows zip 與含 `武士傳說.app` 的 macOS zip；`verify_dist.sh` 驗過 manifest、授權文件、release 的素材 deny-list、macOS 的四道靜態驗收與**啟動**，`verify_packages.py` 另驗六個發行檔的結構並做過注入缺陷的負對照；**推廣片做好了**（59 秒，`promo/kol-promo.mp4`）；權利審查寫在 `docs/RELEASE-RIGHTS-REVIEW.md` | Windows 與 macOS 都沒有真機驗收；版本號仍是暫定；GitHub Release 要使用者點頭才發 |

加權總和 `3×0.95 + 10×0.85 + 18×0.99 + 9×0.98 + 25×0.99 + 7×0.92 + 10×0.97 +
8×0.97 + 3×0.10 + 5×0.88 + 3×0.90 + 2×0.95 = 95.94`，權重總和 103，換算 93.1。

**權重總和不是 100，因為這一輪多了「互動遊玩」這一列。**先前的表沒有這個維度，
而其他每一列的完成度都是用腳本餵 token 逐步截圖驗出來的——那證明狀態機正確，
不證明有人打得開來玩。兩個基數不同，先前與現在的分數不能直接相比；要不要把
權重重新分配回 100，以及總分要落在哪裡，是判斷，留給決策者定。

**這個分數不是進度百分比。**已完成的那 88 分裡，做的是基礎設施最重的部分
（抽取器、對拍框架、字模、存檔、UI 分層）；剩下的 12 分裡，**音訊一個音都還沒進遊戲**（要人耳驗收才升得了 SPEC-004），
Windows 與 macOS 的封包編得出來但沒有真機驗收。

**這一輪把任務那一整條與競技場補完了**：接任務（RE-110）、MOUTH 的打字問話
（RE-067）、任務一覽（RE-111）與競技場（RE-108，含賭場下注）各有實跑收據，
其中 `quest-take` 沒有換過任何輸入。競技場那一條把**升級**接了起來——它是唯一的入口。
除此之外的進展主要是 gate，不是新系統。完成定義要求的正常玩家路徑貫通了
（`full-chain` 等五條收據），但那是把既有的系統接起來，所以分數只動了任務與結局
那一列——接任務（RE-110）做完，結局從「只能用 `-scene ending` 直接播」變成
打完最終任務就會出來。順帶修掉兩個只有實跑才看得見的缺口：建築 typed data 的
載入器漏抄了任務欄位（整條任務交付在遊戲裡從來沒生效過，而單元測試全綠），
以及結局的觸發點原本掛在交任務物品那一步，而原版是在任務戰鬥打完那一刻。

## M0：專案建立

- [x] 撰寫 `AGENTS.md`，固定 Docker-only、證據、Go／Ebitengine、PNG-only runtime、
  OGG、中文說明書、授權與 `dist-all/` 契約。
- [x] 建立本工作清單與 README 起源骨架。
- [x] 初始化 Git，確認原版封裝與未驗證衍生素材由 `.gitignore` 排除；2026-09-01 起只允許
  已驗證且有 manifest 的 PNG 衍生物進 private repo 指定路徑。
- [x] 建立 `wicanr2/kol-remake-cht` private GitHub 儲存庫、設定 `origin` 並複核 visibility。
- [x] 提交並推送專案骨架；在 `WORKLOG.md` 記錄 commit、push 與 Docker 清理狀態。

完成條件：private 遠端可查、工作樹無誤納原版素材、`AGENTS.md` 與本清單已推送。

## M1：鎖定原版與研究來源

- [x] 在唯讀 Docker 容器中列出 ZIP 內容，記錄全部輸入大小、SHA-256、平台與已知來源；
  發行 revision／修改歷史明示為未知。
- [x] 確認 `Knights-of-legend.zip` 中的 DOS MZ executable、資料檔與圖像候選；音訊 driver
  ／資料尚未辨識，留給工具鏈指紋階段。原始檔
  不寫回、不提交。
- [x] 建立 `CONTEXT.md`、`WORKLOG.md`、`docs/re/RESEARCH-LOG.md`、`docs/spec/` 與
  `docs/verification/`。
- [x] 確認魔旅士 2010-04-16〈Knights of Legend 武士傳說〉原文，README 已記錄直接
  連結、作者、日期與摘要。
- [x] 華人玩家來源涵蓋魔旅士繁體中文個人心得與《CRPG 通鑑》中文長篇評介；逐篇記錄
  作者、網址、日期、摘要及其是玩家觀點而非規則 oracle 的限制。
- [x] 更新 README「專案起源」，只納入可直接回查的來源，不杜撰魔旅士連結。

完成條件：原版身分固定；README 每項起源敘述可回查；找不到的來源有查詢紀錄。

## M2：說明書繁體中文化

- [x] 記錄說明書與參考卡 SHA-256、頁數、頁面尺寸及 metadata 擷取工具版本；OCR 工具
  版本待正式文字擷取時記錄。
- [x] 建立 `docs/manual/README.md`、章節檔、逐頁原文定位表與 `glossary.md`；SPEC-003
  已 CONFORMED，且只證實手冊文件工作。
- [x] 逐頁文字擷取、翻譯、人工校對；玩家手冊 79 頁與 Apple II 參考卡 5 頁全部 reviewed
  （玩家手冊 76–77 為目視確認空白跨頁）。
> 未完成項 `spec-manual-page-links`：讓規則 spec 連回說明書頁碼。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
- [x] 驗證 Markdown 連結、表格、圖片權利分類與完整章節清冊；84 頁／12 輸出全套 gate 通過。

完成條件：全書與參考卡都有可回查的繁體中文 Markdown；零個未標示的 OCR／專名疑點。

## M3：可重現素材抽取

- [x] 先辨識既有專用 Docker image；缺件才建立有鎖版來源的 Dockerfile
  2026-09-10 核實已完成：既有共用映像 eob-remake-go:1.26.7-ebiten2.9.9-audio 已由 tools/go_docker.sh 與 tools/run_python_tool.sh 使用（只讀不改）；缺件的兩個已建 docker/Dockerfile.osxcross-go 與 docker/Dockerfile.package。
> 未完成項 `asset-inventory-format-spec`：唯讀輸入的資產 inventory 與格式規格。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
- [x] 勘誤 MONPICS／PERPICS 早期 PNG 探針的 AND-mask alpha 極性；依
  `sub_1BD17` 重生為 bit 1 透明、bit 0 不透明，並用 8×8 棋盤背景重驗。
- [x] 建立離線抽取器 `tools/extract_ega_assets.py`：39 個 archive、769 張 PNG 輸出到
  `assets/original-derived-png/ega/`，原子化替換、逐張重讀驗證、對照 visual probe 雜湊，
  並產生含來源與權利分類的 manifest。
- [x] 建立完整 EGA extraction set：30 個邏輯 archive／1,022 payload 全數有 disposition；
  996 筆已證實、0 筆待證實、26 筆非圖片，並有清冊一致性 verifier。
> 未完成項 `atlas-and-contact-tables`：生成 atlas／接觸表並逐批人工檢查。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
- [x] 以雙新建 IDA DB 關閉六個 executable 的直接 x86 音訊 I/O 母體；只有 TITLE
  PC Speaker 音爆，並排除 `TANDYPIN.EXE` 是音訊播放器。
- [x] 用 DOSBox-X `DX-CAPTURE /A` 在正常 TITLE 流程取得非靜音 WAV，固定四段活動窗、
  codec、時長、峰值與 SHA-256；不追逐週期 parity。
- [x] TITLE cue 的 OGG 呈現方向已決策：規格近似合成（依已證實 PIT 除數序列與公開硬體
  規格自行合成，不散布原版錄音）。
> 未完成項 `pit-square-wave-synth`：實作 PIT 方波合成器並完成人耳驗收。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
- [x] 建立 OGG catalog 的 SPEC-004 DRAFT 與機器可讀音訊 inventory；尚不授權轉換實作。
> 未完成項 `ogg-conversion-pipeline`：SPEC-004 升 READY 後建立 OGG 轉換管線。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `runtime-format-guard-test`：建立測試，保證 runtime 只會讀 PNG 與 OGG。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `release-asset-separation`：分離本機、private-repo PNG 與可公開輸出。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
- [x] 授權條款：RRSAL-1.0（`LICENSE` 整份 ＋ README 授權段）。灰色地帶依本 repo
  實際內容逐項寫明（原版衍生 PNG、原版執行畫面截圖、譯文裡的原版英文字串、
  中文說明書譯文、由原版量測的位址與數值表），第三方是 Noto Sans CJK TC（OFL 1.1）
  與 Ebitengine。
- [x] 打包時把 LICENSE 複製進每一個發行包與 dist-all/<版本>/ 根目錄
  2026-09-10 核實已完成：tools/build_dist.sh:51-54 已把 LICENSE 複製進 dist-all 根、promo/ 與每個子目錄；tools/package_dist.sh:110 的打包清單也帶 LICENSE。

完成條件：從固定雜湊原版可重生 PNG／OGG；runtime 格式守門全綠；權利分類齊全。

## M4：Go／Ebitengine 垂直切片

- [x] 建立角色建立 RE 證據表與 SPEC-005 DRAFT；固定手冊流程、16 個角色槽、768-byte
  CHARDATA stride 與 KOLBOOK 開檔路由，但不授權 production 實作。
- [x] 關閉主要屬性 RNG 的 seed、有界擲骰、獨立骰式／內部基值、骰面表、100/150 顯示
  量尺與三項衍生公式；兩個全新 IDA DB 輸出一致。
- [x] 關閉 35 個手冊職業列與 39 個 executable 記錄的 enum／名稱／eligibility 對應；
  解釋男女精靈共名及四對矮人記錄逐 byte 相同。
- [x] 關閉 39 筆起始金冠、主要／次要武器、武器技能與最小 writer 欄位索引。
- [x] 查核必要 KOLTOWN reader，並以完整 writer coverage 排除其他職業衍生起始資料。
- [x] 以雙新建 IDA DB 固定「輸入 buffer 首 byte 為零」分支、各層取消、reroll、80 張
  肖像循環與滿槽不進 writer；零 byte 不再擅自等同玩家按 Enter 的空名。
- [x] 建立畫面驅動的正常 DOS UI 擷取器（`tools/dosbox_character_ui_oracle.py`、
  `tools/x11_key_inject.py`、`tools/compare_character_ui_runs.py`），並封存名字長度、
  空名、各層 ESC 與 reroll 四個情境各兩輪的畫格、輸入序列與三檔前後 SHA-256 收據。
- [x] 完成正常 DOS UI 的接受屬性、圖像選擇與實際保存 oracle；SPEC-005 經證據審查升
  READY。16 槽全滿仍只有靜態證據。
> 未完成項 `roster-full-ui-oracle`：補 16 槽全滿、重名與名字欄字元集合的正常 UI oracle。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
- [x] 鎖定 Go 1.26.7、Ebitengine v2.9.9 與建置映像 `eob-remake-go:1.26.7-ebiten2.9.9-audio`；
  `cmd/kol` 可建置、可啟動，並有 `tools/go_docker.sh`／`tools/run_kol_headless.sh` 兩支包裝。
> 未完成項 `package-boundaries-and-injection`：定義 package 邊界、固定 timestep、可注入 RNG／clock／input 與存檔 schema 草案。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `minimal-player-chain`：選一條最小玩家鏈：PNG → typed data → 一個互動 → UI → 存讀檔。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
- [x] 建立畫面對拍工具與首批收據：`tools/verify_remake_parity.py` 的三層判準，
  82 對書本畫面全部通過（`docs/verification/remake-parity/`）。
- [x] 抽出原版內嵌點陣字型並機械量得書本文字版面；remake 以原版繪字規則重繪空白
  New Character 頁，與原版只差名字欄的 31 個原版殘留像素。
- [x] 量測角色選項、屬性數值與右頁選單的文字版面，並實作名字→種族→性別→職業→
  擲屬性的狀態機；9 對逐畫面對拍通過，殘餘差異只剩四類已查證成因（RE-045）。
- [x] 實作圖像層版面（右頁說明、`CANCEL   ACCEPT`、青色雙格外框、名牌、上下箭頭）；
  完整流程加 80 張圖像循環共 92 對逐畫面對拍通過（RE-046）。同時指認大字型的第一個
  畫面用途：codepoint 94／95 就是上下箭頭。
- [x] 角色選項頁、角色清單頁與兩段翻頁動畫（RE-049）；新增 `list-characters` 正常 DOS UI
  情境兩輪，`CHARDATA` 的槽位版面由畫面讀回的名字唯一命中確認。對拍 16 ＋ 20 對全通過。
- [x] 角色寫入的逐 byte 對拍（RE-050）：原版改動的 58 個 byte，remake 的 writer 兩個職業
  各 58/58 逐一相同，寫錯 0、多寫 0。建立→寫入→清單已閉環（清單由 14 列變 15 列）。
- [x] 起始武器樣板的來源與索引規則已反追（RE-050）：樣板是 `BOOKDATA` 第 3 項的 14 筆
  16-byte 記錄，索引 ＝ selector − 1；selector 藏在職業 metadata 第一個 byte 的兩個
  nibble。39 個職業都能寫，21 筆沒有副武器的不會誤寫副武器欄位。
- [x] 刪除角色（RE-051）：新增 `delete-character`、`delete-character-confirm` 兩個正常
  DOS UI 情境各兩輪，13/13 與 15/15 個 checkpoint 逐檔相同。刪除畫面就是角色清單頁
  加反白條，確認視窗只認 `y`／`n`；刪除只把槽位名字的第一個 byte 寫成 0，remake 的
  輸出與原版逐 byte 相同。對拍 8 ＋ 5 對全通過。
- [x] `HARBOR MASTER`（RE-052）：只安裝一個 region 時是一張三行訊息頁，1 對對拍通過。
- [x] 書本四個功能的入口已查清（RE-052）：`DELETE CHARACTER`、`EDIT PICTURE`、
  `EDIT SHIELD`、`HARBOR MASTER` 共用同一張角色選擇清單，進入後第一張畫面是同一個 PNG。
> 未完成項 `picture-shield-bitmap-editor`：EDIT PICTURE／EDIT SHIELD 的點陣圖編輯器。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
- [x] 走出書本進入城鎮（RE-053）：目錄頁 `6` → 共用角色選擇清單 → 選領隊 → 半途之家
  旅店。圖示列不吃方向鍵，改用參考卡的 `<`／`>` 移動白框、Return 執行，末格繞回；
  `tools/x11_key_inject.py` 補上 Shift 位 keysym 的送鍵。
- [x] 組隊流程已量測（RE-053）：COMPANION → 是否搜尋硬碟 → 該旅店的可招募清單 →
  上下選、Return 加入，頭像出現在頂端。招募不寫磁碟，三個角色檔全程未變。
- [x] 旅店畫面的組成已拆解（RE-054）：底圖是 `KOLMPICS[0]`（與 `BOOKPICS[3]` 同一張），
  上面疊 11 個矩形，分群之外殘餘 0 像素；六個圖示逐像素對上 KOLMPICS 圖號。
- [x] 店主肖像已定位（RE-054）：`TWN7PICS[12]` 88×64 畫在 (8,64)，88×64 像素全同。
  找法是拿畫面上一列 60 像素當指紋去搜素材，不是拿素材去搜畫面。
- [x] 旅店畫面的 remake 實作與對拍（RE-055）：`internal/town` ＋ `ui.Inn`，
  八對逐畫面對拍全過（六個圖示白框位置、單人隊伍、招募後兩人隊伍），
  允許區只有原版滑鼠游標。中文模式一併補上。
- [x] 戶外地圖已拆解（RE-056）：底圖、徽章（`TWN7PICS[0]`）、日期框、頭像格與單一
  `KOLMPICS[5]` 圖示都零差異對上；地圖視窗是 12×6 格、每格 16×16 的**捲動磚塊圖**，
  移動時以 16 px 為單位捲動。
- [x] 戶外地圖的磚圖已定位（RE-057）：在 `TOWNn.0` 城鎮資料檔裡，payload 0／1／2 各 96
  塊，格式是 16×16、四個 EGA 平面、plane-major，每塊 128 bytes。畫面上 72 塊磚有 67 塊
  逐 byte 命中，offset 全是 128 的倍數。`TOWNn.0` 之前不在 inventory 裡，是因為清冊工具
  只認名字含 `PIC` 的檔案。
- [x] 磚圖的載入路徑已讀出（RE-058）：`sub_15920` 把 `town%d` 的 payload 0／1／2 放到同一
  segment 的 0x0000／0x3000／0x6000。
- [x] 地圖格子陣列已解出（RE-059）：`sub_115D1` 是 12×6 的繪製迴圈，格子索引 `(X<<6)+Y`，
  X 下限與 Y 上限靜態就是 0 與 `3Fh`。`sub_1D1C6` 的來源段是「緩衝段 +0x100 paragraph」，
  所以段內 0x0000–0x0FFF 是 64×64 的格子陣列、0x1000–0x8FFF 是 **256 塊**磚。
  拿原版畫面上 67 格當指紋掃 64×64 個鏡頭位置，唯一命中且 67/67 全中。
- [x] 隊伍標記已解出（RE-060）：精靈是 KOLPICS payload 0 的十個 128-byte 單元，
  遮罩是把四平面 OR 成佔用位圖後左右膨脹一像素再反相算出來的（`sub_1D21D`），
  膨脹在 little-endian word 上做因此不跨位元組邊界。
- [x] `TOWNn.0` 與標記 atlas 已納入正式抽取器：八個城鎮各一張 64×64 灰階地圖與一張
  256×256 磚塊 atlas，另加一張 16×160 的標記 atlas，共 769 張 PNG。
- [x] 戶外畫面已實作並對拍：`internal/town` 的 `Grid`／`Outdoor`／`Markers` ＋
  `ui.Outdoor`／`ui.OutdoorCJK`；21 對逐畫面對拍全過，允許區只有原版滑鼠游標。
  `tools/solve_outdoor_camera.py` 從 94 張移動 oracle 反解出 22 張戶外畫格的鏡頭、
  標記偏移與動畫畫格，每一張都是唯一解。
- [x] 門口狀態已實作並對拍（RE-061）：門座標表在 payload 3，門 i 對到門面圖
  `TWN{n}PICS[2i+3]`、肖像 `[2i+2]`、文字 payload `i+4`；八座城的門數與圖片配對數
  逐一相等。門口是狀態不是位置——掃門迴圈接在移動後面。
- [x] 解碼器 `sub_1B94F` 已讀出（RE-064）：RE-062 用統計猜的七個符號全部命中，
  字元表由強推論升級為**已證實**；引號吃空白與 `[1B][1B]` ＝ `!"` 的成因一併解開。
  `sub_14234` 的迴圈上限 `40h`／步進 `10h` 也證實了「一則訊息 4 格」。
- [x] 城鎮文字編碼已解出（RE-062）：位移 128 起每 16 byte 一格，
  第 0 byte 是字數、其後 15 byte 是 24 個 5-bit 字元；小寫預設、大寫靠記號。
  7 城 101 棟建築共 9,515 行已抽成 typed data。
- [x] 建築內部畫面已實作並對拍（RE-062）：一則訊息固定 4 格 ＝ 64 byte，
  訊息 0 與 1 合起來是進門敘述的八行；畫面規則只有兩條（句號後兩個空白＋下一字母大寫、
  逗號後一個空白），套上之後與原版逐字相同。畫面完全由三份 typed data 組出來。
> 未完成項 `building-followup-dialogue`：建築內的後續對話（訊息 10 起）。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
- [x] 城鎮譯文格式已定案並上線（RE-063）：`docs/i18n/town-zh-Hant.json` 依
  「城／門／欄位」分層，譯文自己帶斷行；布瑞提爾 21 棟建築的名稱、人物與進門敘述
  共 168 行已翻好，文字安全矩形 286 項全過。
- [x] 七座城 101 棟建築的名稱、人物與進門敘述全部翻好（`docs/i18n/town-zh-Hant.json`）。
  城名由世界地圖記錄的三字母代碼配上城鎮對話的一手證據定出來（RE-066）：
  布雷特爾、赫特隆、辛博瓦德、波伊特水閘、謝勒農、奧蘭斯、霍布。
  文字安全矩形 863 項全過。
- [x] 建築內的對話機制已解出並實作（RE-067）：六個話題槽在建築 payload 位移
  0x17…0x1C，LISTEN 講沒有掛關鍵字的槽、MOUTH 用四字元字首比對關鍵字表
  （payload 20，七座城共 305 個）。招呼語依性別／會不會法術／聲望選。
  三對逐畫面對拍全過，另有 9 步玩家路徑，其中三步與對拍截圖逐 byte 相同。
  409 個非空話題槽的關鍵字與訊息交叉檢查零例外。
- [x] **七座城 101 棟建築的建築內對話全部翻好（1,938 則）**，加上城鎮與野外畫面
  會用到的 25 筆 KOLDATA 記錄。中文字模表 2,366 字，文字安全矩形 6,693 項全過。
- [x] KOLDATA／BOOKDATA／ENDTXT 三份 UI 文字庫**全部 486 筆已翻完**
  （`docs/i18n/ui-zh-Hant.json`）。譯文與原文用同一套語法（`@` 換行、
  `~Vxx`／`~Hxx` 定位、`|X` 代換），走同一套處理；
  `tools/verify_ui_translation.py` 486 筆全過，字模表重烘為 2,434 字、缺字 0，
  `verify_text_fit.py` 6,685 項全過。
- [x] BOOKDATA 與 ENDTXT 的譯文**都接上畫面了**：結局畫面走 ENDTXT，書本的標籤走
  BOOKDATA——`internal/i18n` 的 `LoadBank` 把文字庫那一份疊到
  `newchar-zh-Hant.json` 上（剝定位碼、按 `@` 切段、去尾端冒號與補齊空白，
  中英段數對不上就整筆跳過）。接起來之後有 21 個詞改用文字庫那一份，
  `HIGHWAYMAN` 的誤譯「築路工」順手改成「攔路盜」。
- [x] KOLDATA 譯文的**框內寬度**已驗證：`verify_text_fit.py` 多吃一個 `--ui-i18n`，
  把畫在城鎮對話框裡的每一行逐行量過（6,952 項全過）。**行數還沒驗**——一筆記錄
  不一定是「一次畫出來的一段」，哪一筆會一次畫幾行要等每一筆都接上畫面才知道；
  索引 ≥ 0x100 的戰鬥訊息與含 `~Vxx` 自己跳位置的記錄也還在檢查範圍外。
- [x] COMPANION 已實作並對拍（RE-072）：候選人由角色檔所在地欄比對建築名稱取得，
  清單一次八列、兩端夾住、反白到底才捲；加入之後新成員成為發言者，
  招呼語依角色記錄位移 0x1B 的傷勢換成「你受傷了」。建築內的圖示列現在吃
  `<`／`>` 與 Return，選到還沒實作的圖示會記進 `Session.Unimplemented`。
- [x] REST 已實作並對拍（RE-073／SPEC-006）：客棧住一晚的價錢是建築 payload
  位移 0x38／0x39 的 16-bit 欄位（半途之家 56，八座城 10 間是 0…120），
  修道院是「傷勢 × payload[0x39]」；金幣是角色自己的錢包（角色記錄 0x2A–0x2C 的
  24-bit 大端）。付完錢寫存檔，原版還會把角色移出隊伍——那一段做成
  `faithful_inn_stay` 開關，預設不離隊，對拍時打開。
  對拍 `town-rest` 11 對全過，玩家路徑 `town-rest` 11 步（含存檔逐欄驗證）。
- [x] MIRROR 的角色資料頁已實作並對拍（RE-074）：八個屬性的畫面值是內部值 × 100 / 150，
  身高體重查 DGROUP 三張表，武器名是「高 nibble 是字數、從第 4 個 bit 起」的 5-bit 壓縮。
  整頁 320×200 只差原版自己畫的滑鼠游標。玩家路徑 `town-mirror` 10 步。
- [x] 角色資料頁的法術一覽已實作（RE-123）：兩欄各八格，一格一個白框加名字，
  空的一格之後全部不畫；圖示列縮成只有 U-TURN。收據 `town-scroll` 第 22 步。
- [x] 角色資料頁的物品（sub_16986）與隊伍立繪（sub_18D61）兩個圖示
  2026-09-10 核實已完成：internal/town/itempage.go 的 OpenItemPage 與 internal/town/portrait.go 的 OpenPartyPage 都已實作，立繪疊法在 internal/portrait/（RE-126／127／128）。殘留的 s.note(sheetIconName(...)) 只是資料缺失時的退路，不是未實作標記。自評表也已記為做完，只有這一條沒回頭改。
- [x] 商店的商品清單與明細已實作並對拍（RE-075）：商品表在建築 payload 的 0x100 起、
  一筆 16 byte；清單與招募同伴共用同一個元件，明細的每一欄都由那 16 byte 算出來。
  買下會扣錢並放進背包（角色記錄 0xA0 起六格）。整頁只差原版的滑鼠游標。
- [x] 商店買完之後的合身與穿戴已實作並對拍（RE-075）：尺寸剛好就直接穿上，
  不合身才問要不要花錢改；穿上是把背包那一格與部位槽對調。三張係數表都在
  KOLTOWN 的 DGROUP。
- [x] 馬廄（旗標 bit5）已實作，收據 `town-stable` 17 步（RE-119／RE-120）：一筆
  0x80 byte，圖號在第 0 個 byte、名稱在 +0x70、價錢在 +0x78／+0x79（大端）、
  腳程在 +0x7A。固定四筆、不看結束標記，五間馬廄的表完全相同
  （200／400／800／1600）。買下是把 `+0x70` 起十六個 byte 抄進角色記錄 `+0x1B0`，
  折抵身上那一匹價錢的一半、**折完不大於 0 才收 50**。明細那一頁畫的是 `STEEDS.0`
  的第 `picture` 張（40×38，抽在 `assets/original-derived-png/ega/STEEDS/`），
  資料頁右上角也有一塊，名牌只印四個字元。**沒有原版對拍截圖**：五間馬廄都不在
  城 7，原版 oracle 還走不到。
- [x] TRAIN 的甲冑調整已實作（RE-075）：掃過頭、軀幹、腿三個部位，不合身的逐件講一句，
  講完問總價，付錢之後把尺寸改過來。「都合身」那一句已對拍；不合身那一條路
  原版存檔走不到，只有靜態證據。
> 未完成項 `train-forge-entry-condition`：TRAIN 的鍛造（旗標 byte 1 的 bit3）。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
- [x] AWARD 的訓練已實作並對拍（RE-076）：帳面位階與實得位階不同就先去競技場；
  學費在選技能之前就扣掉；四門技能在建築 payload 的 0x50 起、一筆 0x0C byte；
  「要幾點技能」是一格文字欄，只收 1…5 的一位數，打完要按 Return，欄尾那條
  8×1 的橫槓是原版自己畫的閃爍游標，不是字模。點數不夠那一句已對拍；
  加技能那一條路原版存檔（ddduhh 冒險點數 0）走不到，只有靜態證據與單元測試。
- [x] SCROLL 已實作，收據 `town-scroll` 19 步（RE-121）：加入這一棟代表的教派
  （智力門檻 payload `0x36`、入會費就是建築價錢、教派那十六個 byte 抄進角色
  `+0x1A0`、信物放 `+0x170` 或背包），以及跟它買法術（payload `0x100` 起固定八筆、
  價錢 `+0x0F` × 10、抄進角色 `0x200`…`0x2FF` 的空槽）。**法術的母表就是這一塊**，
  SPEC-012 第 10 節的那一項因此結案。
- [x] 選單第二列「Modify a spell」是**法術編輯器**（`sub_178E0`，RE-128），已實作：
  挑一筆抄成工作副本，十三個欄位（名稱、四格護符的類與種、屬性、射程、兩個骰式）
  各自開文字輸入或選單，選項與造價全部來自從沒被解過的 `MAGIC.0`；離開時問
  「Record changes in Spellbook?」，比過冒險點數與金冠才扣款寫回。版面來自記錄
  0x4D，教派准哪幾類在建築 payload `+0x60`（六個教派剛好把八類分完）。
  **已對過原版畫面**：整頁 320×200 允許區外零像素差異，收據
  `receipt-town-spell-edit.json`，原版那一側是 oracle 情境 `town-spell-edit`。
- [x] 城鎮已可實際操作：`town.Session` 三種模式（地圖／門口／建築內），方向鍵與斜走鍵
  走動、`<`／`>` 移動白框、Return 執行。鏡頭跟隨的規則「偏移夾在 0…1、多出來由鏡頭
  吸收」直接重播 oracle 序列通過（強推論）。14 步玩家路徑收據，其中兩步與已對拍截圖
  逐 byte 相同。
- [x] 城鎮地形擋路與地圖邊界的城門提示（RE-065）：磚塊編號的不可通行位元圖在
  `TOWNn.0` payload 3 的位移 0x00–0x1F；夾限的是鏡頭不是隊伍，夾到就跳出
  DOOR／U-TURN 提示、地圖視窗換 MISCPICS 圖 0。四對逐畫面對拍全過，
  另有 13 步玩家路徑，其中兩步與對拍截圖逐 byte 相同。
- [x] 圖示列的生命週期（RE-065）：走離門口清空、踩到門重建成兩格、
  選 U-TURN 與從建築物出來都只剩一格 DOOR。三個狀態都有原版畫格佐證。
- [x] 野外地圖畫面已實作並對拍（RE-066／RE-068）：`internal/world` 有格子陣列、
  次格移動、地形與道路查表、城鎮入口比對；畫面有地圖視窗、時段／地形標籤框與
  那個 2×2 的閃爍點（把底下的 EGA 索引與 15 互斥或）。四對逐畫面對拍全過。
- [x] 野外的雲層動畫
  2026-09-10 核實已完成：2026-09-10 核實：已接進畫面。`internal/world/weather.go:69` 的 Composite 做旋轉，`internal/ui/world.go:45` 的 drawWeather 畫進世界畫面（144／217 兩處呼叫 CompositeGroup(s.Group, s.Phase)），`cmd/kol/main.go:937` 遞增 Phase。G02 那一條已記為做完（SPEC-026 六組天候、128 像素捲雲與正式節拍，正常出城 GUI 只在天空區改變），只有這一條沒回頭改。雨雪的正常 GUI 與原版同相位另見 weather-animation-and-speed。
- [x] 城門選 DOOR 已接到野外畫面（RE-069）：落點是這座城鎮該城門在世界地圖上的
  格子，而那一欄與世界地圖城鎮記錄的座標 16 組全部相同。12 步玩家路徑，
  最後一步與野外對拍截圖逐 byte 相同。
- [x] 野外的圖示列與進城已實作並對拍（RE-070）：走到城鎮入口那一格會換成
  `TWN0PICS[記錄索引 + 8]` 的外觀圖與 DOOR／U-TURN，選 DOOR 依記錄的鏡頭、
  四個界限與入城點回到城鎮；站在路上會多出 ROAD 與 SIGNPOST 兩格。
  玩家路徑 `town-round-trip` 走完城鎮 → 野外 → 城鎮，對拍 `world-town` 七對全過。
- [x] ROAD 與 SIGNPOST 的自動循路已實作並對拍（RE-071）：道路表存的是八個方向的
  位元遮罩，ROAD 沿目前方向走、斷了往兩側各找三格，SIGNPOST 排除目前方向另挑一條；
  走正向時會用斜向把隊伍拉回格子中線，而那一步只准落在道路磚塊上。
  圖示列的三種框色（白／青／亮藍）也一併對上。玩家路徑 `world-road` 33 步，
  對拍 `world-road` 10 對全過。
- [x] 野外的時刻、日曆與天候已實作並對拍（RE-077）：每步扣 `2 × (地形成本 + 隊伍移動率)`，
  撞牆那一步不耗時間；時刻從 4200 倒數，五個門檻切出五個時段；換日補營養（表隨月份）、
  推日期（沒有閏年）；天候是每月各有上下限的隨機漫步，只在換時段時擲。
  第 14 步與第 15 步的標籤框與日期框與原版逐像素相同。
- [x] 換日的 CAMPING! 畫面已實作（RE-077／SPEC-007）：天候切第 5 組、標籤第一行換成
  記錄 0xE2、地圖視窗換成腳下地形那一張圖，按任何鍵收掉之後才推日期。
  **還沒對拍**——原版走到換日要 76 步、中間會經過四次時段變換，每一次都擲遭遇，
  兩輪量測都在半路被伏擊攔下。要等戰鬥打通、能打完再走，才拿得到同狀態畫面。
- [x] 野外的隨機遭遇的**擲點規則**已實作（`sub_2A4E5`，SPEC-011／RE-097）：換時段
  時擲一次，機率查「季節 × 地形」表再除以 2，怪物由地形與季節的位元遮罩挑，隻數與
  等級由平均位階和級距表決定，兩邊的站位來自四張固定表（RE-098），地形逐格擲出來
  （RE-099）。**全隊都有坐騎時先問「要不要騎走」**（RE-112）：那一頁印那一隻記錄
  末端的四行敘述，答是就擲 `rand(100) > 5 × 移動率`，成功回野外、失敗才開打。
  實跑收據 `docs/verification/battle-ambush/` 與 `battle-rideaway/`（後者換過名冊，
  因為原版名冊裡的角色都沒有坐騎——十六槽的 `+0x1B0` 全是零，馬廄是唯一的入口；
  馬廄已經接上了，RE-120，所以現在也走得到「買一匹再出城」那一條）。
- [x] 24 個任務地點的提示已實作並對拍（RE-079）：走進那一格而且隊上有人帶著那個任務，
  就蓋出記錄裡的八列敘述與「Will you partake?」。24 個地點的座標、怪物、站位與敘述
  都在 `quest-sites.json`。答應之後畫出滿版的怪物圖與記錄 0xE9，那一張也已對拍。
- [x] 戰場畫面已實作並對拍（RE-080／SPEC-008）：9×9 的視窗、地形取法與夾限、
  逐格拉線的視線（看不到的畫磚塊 0x4A）、參戰者疊圖（剪影往左右各推一格當遮罩）
  與目前輪到那一格的白框。收據「戰場：地形、視線與參戰者」除了原版滑鼠游標與
  十二格亂數樹叢之外逐像素相同。
- [x] 回合的規劃順序已實作（RE-080）：一個回合先規劃再執行，規劃鍵值 =
  擲骰(100) + (等級 − 1) × 10 + 洞察，隊員第二回合起先除以 2 × 回合數，
  再依傷勢折損；鍵值小的先決定，整輪決定完問「Begin next round?」。
  執行段依快捷重排、鍵值大的先動，式子已讀出但要用到負重與指令時間成本，
  那兩項還沒解完，因此不實作。
- [x] 敵人那一手的決策已解出（RE-081）：`sub_2374C` 的六條分岔（撤退、休息、
  射擊、貼近打、移動、找不到目標時繞圈）、同心環搜尋與篩選遮罩、智力擲骰、
  反制掃描 `sub_253E7` 與側翼 `sub_25653`、加權輪盤 `sub_25A3A` 的等級偏向式，
  以及三個決策全域寫回 +0x23／+0x24／+0x25 的約定。方向表 `[0x10D8]` 解出來的
  九格與 RE-080 由 `0x0B76` 反推的方向定義獨立吻合。
- [x] 玩家那一手的規則已解出並寫成 SPEC-010（RE-082）：`sub_228B7` 主指令列
  十五個圖示各自的條件、`sub_24131` 兩個 mode 的游標規則（近戰不准停中心、
  射擊不准進內圈）、四列子選單的選項換算，以及 `[0x0B6C]`／`[0x0B76]` 互為反表。
  同時更正 RE-080 標反的兩列子選單：`0x1F6BA` 是武器攻擊式、`0x1F940` 是徒手。
- [x] 指令列與選目標的規則層已實作（SPEC-010）：`internal/battle/commands.go` 的
  `CommandBar` 照原版順序組出十五格的條件，四列子選單的選項換算與兩張方向反表
  一併寫進去；`cursor.go` 是 `sub_24131` 兩個 mode 的游標（近戰不准停中心、
  射擊的內圈方向鍵乘三跳出去、滑鼠點進去還原）。二十個單元測試全過。
- [x] 指令列已接上畫面與輸入，主指令列也對拍過了（RE-083）：規劃段的選單狀態機在
  `internal/battle/planning.go`，畫面在 `internal/ui/battle.go`，輸入在
  `cmd/kol/main.go` 的 `updateBattle`／`applyBattleToken`。圖示列整區
  (32, 170) 216×22 的 4,752 個像素與 frame-886 零差異——用的是原版角色檔裡
  Poontang 真實的 Long Spear，不是反推的最小條件。
- [x] 修掉一個回歸：`tools/extract_character_roster.py` 把 `equipment` 放進了
  `sheet` 這一層，而 `internal/roster` 從槽位層讀，所以 `Worn` 一直是十格全 nil。
  單元測試抓不到（規則層自己餵 `Member`），是對拍抓到的。
- [x] 執行段已實作（RE-084／SPEC-008）：鍵值＝快捷扣傷勢、負重與指令的時間成本，
  再加擲骰；排序小到大、由尾端往回走，所以鍵值大的先動。射擊的武器種類不是
  0x12／0x16 再減半。攻擊那一手接上 SPEC-009 的結算——面前那一格用遮罩 8 找人，
  **不分敵我**；攻擊選項 0 印記錄 0x117、面前沒人揮空印 0x11F／0x120、
  休息印 0x119、怯戰把防禦選項設 2 再印 0x118。
- [x] 移動那一手已實作（RE-085）：一次只走一格，**沒有點數規則**——移動方式只
  決定要不要把 +0x1E 設起來，以及執行段鍵值裡的時間成本。走不動的四種情形
  （地形、腳下、有人、界限）都照原版判；擋路的位元圖 `[0x0FC8]` 與擋視線的
  `[0x10CE]` 是不同的兩張表。
- [x] 撿、丟與切換那三手都做了（RE-094／RE-113）：丟是把手上那件放到腳下，
  撿之前先問「要不要撿」，切換在三隻手之間換。`Act` 現在只在敵人沒有指令時
  標 `Unimplemented`。
- [x] 訊息的代換碼已解出並實作（RE-086）：`]` 換成空白，`|A`…`|S` 十一個碼分別是
  雙方的所有格／受格、行動者的武器與腰帶那一格、雙方的名字（敵人補 `The `）、
  物品名、數字與暫存名。每一手開頭先印行動者的名字（`sub_2685A`），
  所以整句是「<名字> <記錄展開>」。`internal/battle/message.go` ＋ 八個測試。
- [x] 訊息已經畫到畫面上（RE-087）：先進緩衝區、收尾才斷行，版面固定在 byte 欄 3
  （像素 24）、掃描列 0xA8、行高 6，斷行從行寬處往回找空白（第一行 30 字、之後
  33 字），每畫一個字元前先清一塊 8×8。實跑「進戰場 → 選休息 → 執行」之後用
  `read_text_from_frame.py` 從 (24, 168) 讀回 `Poontang is resting!`。
  **這是 remake 自我一致的驗證，不是 parity**——`world-quest-run1` 那一輪 oracle
  沒有走進執行段，沒有原版畫格可比。
- [x] 執行段的原版 oracle 已錄（RE-089，情境 `world-quest-execute`，969 個畫格）。
  量到五件事：訊息區整條在執行段換成白底加黃色圖案、文字白底黑字、整段轉大寫、
  起點與行高和反組譯讀出來的一致、移動成功要按兩次鍵。remake 已照著修。
- [x] 訊息區的底圖是 MISCPICS 的圖 1（一張卷軸）。remake 已改成畫它，
  `receipt-battle-exec.json` 那一對允許差異區之外零像素差（RE-090）。
- [x] 敵人的名字沒有對不上：原版那一輪根本沒走到任務地點，是路上被 1 隻 1 級
  RUFFIAN 埋伏（frame-958）。任務戰鬥的怪物確實取自任務記錄的第 0 個 byte，
  抽取器解對了（RE-090）。
- [x] 野外隨機遭遇整條打得起來了（SPEC-011）：走九十步被伏擊兩隻 1 級 BINDERAK，
  訊息、站位與地形都對。剩下遭遇圖的版面與任務怪那一條。
- [x] 攻擊那一手的訊息鏈已解出並實作（RE-088）：第一句是動作加結果、第二句是
  傷害，各等一次鍵。招架碼的訊息基底是 **0x15F** 不是 RE-080 原本寫的 0x14F，
  另外碼為 4 而且對方體型大很多時會改成 2。實跑揮空那一條讀回
  `Poontang whips his Long Spear through the air.`（兩行，斷行正確）。
- [x] 攻擊訊息的三條路都實跑過了：走進戰場、往下移動四步貼到最近那一隻
  BINDERAK，攻擊之後讀回
  `Poontang whacks his Long Spear at The BINDERAK who cannot parry the blow!`
  與 `The blow hits The BINDERAK's chest killing it instantly!`。
  **remake 現在打得死怪物**——規劃、移動、攻擊、傷害與訊息整條通了。
- [x] **怪物記錄 `+0x10` 的 bit 0 是「令人生畏」**（RE-135）：CYCLOPS、CLIFF TROLL、
  SLEDGE 與 MIST GRUB 四隻有。那一位在場時，隊員每要出手一次（武器攻擊、射擊、
  施法、撿東西、徒手）就得過一次意志檢定 `((力量 ＋ 智力) ÷ 2 × 等級) ÷ 3`，
  擲 100 擲到門檻以上就換成怯戰（指令碼 `0x0B`，印 0x118
  `is frozen in terror!`）。執行段那一段 remake 早就有了，**缺的是誰把指令碼
  改成 `0x0B`**——現在補上了（`checkFlinch`）。參戰者記錄 `+0x06` 是免疫格，
  由 `sub_21FCD(索引, 2)` 依裝備賦予，**哪一件裝備還沒解**，所以目前一律是 0。
- [x] `0x27B77` 之後的擊退與昏迷已解並實作（`internal/battle/knockback.go`）：
  推得動就把目標移過去印 0x159，推不動印 0x158，倒地再補 0x15B。
  倒下那一句的 collapse／shatter 兩個分支條件已定（RE-131：怪物記錄 `+0x10`
  的 bit 7，三十二隻裡只有 SKELETON），實作在 `fatigue.go` 的 `Shatters`。
> 未完成項 `attack-row-and-target-cursor`：攻擊式那一列（frame-899）與選目標格畫面的對拍。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
- [x] 敵人 AI 的三張基礎權重表（`byte_3496A`／`3496B`／`3496C` 移動三式、
  `[0x6AD2]` 攻擊式、`[0x1A9A]` 部位）**已由 RE-092 解出**：戰鬥初始化時從 DGROUP
  `0x1092`／`0x10A6`／`0x10AE` 複製，內容是 `6 6 5 4 6 4 4 6 6 5`、`1 3 4 4 4 4 2`
  與 `12 24 72`。
- [x] 循路一步的時刻成本已修（`world.RoadStepCost = 20`）：原版連按 ROAD 時，
  時段標籤在第 9 次才翻成 MID MORN（兩輪 oracle 一致），往回推一步扣 18…20；
  取 20 是因為 8 步 832、9 步 852 剛好卡在門檻 3367 兩側。道路那一份十對全過。
  **機制還沒追出來**：`sub_297F2` 只有一處扣時刻，呼叫前的 `word_39A54`／
  `word_2FA2C` 對不上這個值，常數旁邊已註明，之後追到再換成算式。
- [x] 走到 (55,13) 的競技場與賭場（`sub_2AAA8`）：選人、確認、進賭場下注、
  單挑打完升一階並結算賭金，實跑收據 `arena`（1,770 步）。場上只有上場那一名，
  其餘排到觀眾席（RE-098 的第三條分支）。
- [x] 野外畫面的中文模式版面已完成：`world.CJK` 把左欄重排成日期框、天候圖、
  時段／地形標籤（中文兩個字要 32 px，英文版面的 67×15 放不下），地圖與圖示列
  沿用城鎮中文版面的兩倍尺寸。五個時段與八種地形的譯文已進 `docs/i18n/town-zh-Hant.json`，
  字模表重烘後 `verify_text_fit.py` 1,789 項全過。
- [x] 隊伍上限六人的邊界
  2026-09-10 核實已完成：internal/town/interior.go:64 的 const MaxParty = 6 已在 recruit.go:291（招募）、interior.go:44（進門）與 save/characters.go:191（保存）三處把關。
- [x] remake 自己的存檔格式已定案並實作（SPEC-006）：`kol-save-v1` 是可讀的 JSON，
  `internal/save` 負責原子讀寫，`tools/import_original_characters.py` 從原版
  `CHARDATA` 離線匯入。原版檔案一律唯讀，runtime 不解析原版格式。
- [x] 讀檔回到遊戲：`-load` 把隊伍、金幣、傷勢與位置放回去；角色的靜態資料
  （頭像、屬性、武器）仍由 typed data 提供，存檔只帶會變的欄位。
  玩家路徑 `town-load` 用 `same_as_except` 證明畫面只有金幣那一欄變了。
> 未完成項 `text-layout-remaining-screens`：量測城鎮、世界、戰鬥畫面的文字版面。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
- [x] 繁中雙模式：英文模式 320×200 走原版字模與嚴格對拍；中文模式 960×600、16×18
  中文字模格。字模由 `tools/build_cjk_font.py` 從 Noto Sans CJK TC 烘成點陣表，
  執行期只讀 PNG。New Character 頁全頁中文化；文字安全矩形檢查目前 125 項全過。
- [x] 玩家可見的展示截圖：`docs/screenshots/` 22 張由 `shots.spec` 重生，有 manifest
  與就地驗證（尺寸、顏色數、與對拍收據逐 byte 相同）；README 改引用這一批。
> 未完成項 `cjk-remaining-screens`：城鎮、世界、戰鬥畫面的中文化與職業名暫譯的交叉確認。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
- [x] 標題序列：從 `TITLE.EXE` 的 `_main` 靜態讀出腳本（RE-047），remake 依腳本重播一輪，
  原版兩輪觀察的每一張畫格都逐像素對上，沒有未解釋的畫面。
- [x] 標題序列的實際播放：時間迴圈與按鍵中止已實作（`-scene title`）。
- [x] 翻書動畫：從 `KOLBOOK.EXE` 的 `sub_114F5` 讀出表驅動的逐欄捲動（RE-048），
  remake 依同一張表重播，18 對逐畫面對拍全通過。
> 未完成項 `title-thunder-audio-hookup`：標題序列的雷聲播放接上 Ebitengine 音訊路徑。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `cross-platform-build-smoke`：Linux／Windows／macOS 可重現編譯與最小啟動 smoke。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。

完成條件：不靠 debug 入口可啟動、操作、存檔與重載；沒有 runtime 原版格式依賴。

## M5：戰鬥、任務與正常玩家路徑

- [x] 戰鬥的 spec 已建立並升 READY：SPEC-008（戰場、執行順序、移動）、SPEC-009
  （近戰結算）、SPEC-010（指令列與選目標）。三份都只把**已解出**的部分標 READY，
  未解的段落留 DRAFT。
- [x] 節奏由使用者於 2026-09-03 決定：**忠實 ＋ 三個可關閉的加速選項**
  （重複上一指令、快速結算未接敵回合、動畫速度），對拍一律在全關的狀態下驗證。
  決策已定案，**三個選項本身還沒實作**——目前 `internal/save` 只有
  `faithful_inn_stay` 一個開關。
> 未完成項 `spec-ready-before-implement`：每份玩法 spec 經證據審查升 READY 後才實作，同狀態驗證後才標 CONFORMED。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
- [x] 貫通隊伍建立、城鎮、世界旅行、隨機遭遇、任務戰場、物品、訓練與勝負流程。
  `full-chain` 一串按鍵從標題走到住宿存檔，中間打贏一場遭遇；`full-chain-reload`
  讀回來走到任務地點；`full-chain-ending` 走到結局。**接任務那一側還沒實作**，
  所以結局那一條的存檔是工具造的。
> 未完成項 `save-across-battle-and-clock`：存檔跨事件與戰鬥階段恢復：戰鬥中途狀態與世界時間。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。

完成條件：正常玩家路徑不用 direct-entry 可走到預定結局；核心規則宣稱都有收據。

## M6：中文 UI、驗收與交付

> 未完成項 `terminology-and-safe-rects`：統一遊戲與說明書術語；每個 CJK 欄位定義文字安全矩形與 overflow policy。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `longest-cjk-visual-regression`：最長繁中內容的視覺回歸、正常路徑截圖與輸入測試。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `parity-sampling-registry`：預先登記近似完成／嚴格 parity 的抽樣母體、硬失敗與平台矩陣。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `dist-rebuild-version-number`：由乾淨輸入重建 `dist-all/<版本>/{full,release,promo}` 的正式版本號。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `dist-package-sha256-verification`：發行包的 SHA-256 驗收。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `windows-macos-real-machine`：Windows／macOS 真機驗收。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `github-release`：發 GitHub Release。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。

完成條件：交付包、證據、權利分類與已知限制一致；本專案 Docker 容器已清理。

## 裝備耐久還沒做

物品記錄 `+0x0B` 是耐久（RE-075）。原版在傷害超過武器上限時扣它，扣光了**那件裝備
整個清成 0**（`sub_22084`）——連帶負重、盾牌等級與徒手拿的那一格都跟著變。

remake **完全沒有這一格**：`Item` 與 `Weapon` 都沒有耐久欄位，抽取器也沒抽。

**扣除規則已經實跑量到**（RE-133）：`記錄[+0x0F]` 的 bit 5 沒設就不扣；
新耐久 ＝ `記錄[+0x0B] − 32 × 次數`；為負就把那一格 16 byte 整個清成 0，剛好 0 不清。

**觸發時機也解出來了**（2026-09-07）。六個呼叫端裡四個是穿戴欄那四格、次數固定 1；
另外兩個在攻擊結算裡，次數是區域變數——那幾個變數現在都有名字了：`var_E` 是這一擊
的傷害、`var_4` 是武器上限（`骰數 × 骰面 + 1 + 基礎`）、`var_C` 是穿戴欄位移、
`var_A` 是次數。**超過武器上限時 `sub_22084(物品格, 傷害 − 上限)`**，扣光了把
`+0x25` 清 0。

先前判定「`sub_26D76` 直接呼叫跑不完」是錯的——它**呼叫得動**，只是要傳參數
（`arg0 == 0` 才走攻擊那一條）而且命中之後要釘住 `sub_1B100`（等玩家點滑鼠）。
整條數字鏈已經對拍過（`docs/re/RESEARCH-LOG.md` 最後幾節）。

**但值不值得做要重估**（RE-134）。說明書從頭到尾**沒有提過武器或盔甲會磨損**，
`CND` 只是物品畫面的一個顯示欄位；而扣不扣的判準 `+0x0F` bit5 **與鍛造共用**，
全部物品母體裡只有 Cloth Pants 與 Leather Hood 有那一位——**原版真正會被扣耐久的
只有那兩件布皮裝**。

照抄要動抽取器、`Weapon`、`DamageRoll` 與兩處重算，而玩家能觀察到的差別只有
那兩件裝備。這是成本效益問題而不是忠實度問題，等使用者決定。

（這一項是 RE-133 那一輪「把 `Actor` 上建戰鬥時算一次的欄位逐個對回原版取用點」
找出來的第四項。前三項——負重與武器速度、盾牌等級、攻防技能那一列——都修好了。）

## 收據基準：13 份已知失效，待重生

remake 的亂數換成原版那一支之後（RE-132），舊收據記的是**用錯的亂數凍出來的畫面**，
所以整批失效。**這一節存在的理由是不讓它變成沉默的失敗**——收據還在 repo 裡，
跑起來會紅，那是正確的狀態，不是待修的 bug。

通過的兩份：`quest-list`、`quest-take`（整條不擲骰）。

失效的十三份，全部同一個根因——遭遇（`internal/battle/ambush.go:109` 的
`Roll(100)`）與治療者（同檔 `:113`）的觸發步數位移：

| 形狀 | 收據 | 表徵 |
|---|---|---|
| 遭遇不在原本那一步觸發，腳本留在野外模式 | `arena`、`full-chain`、`full-chain-ending`、`battle-flee`、`battle-rideaway`、`battle-spoils` | 「未知的野外操作 token："SPACE"」 |
| 位置與提示對不上（遭遇打斷了移動）| `flow-quest`、`full-chain-reload`、`world-healer` | 里程碑的 `world.x/y` 差一兩格、`quest_prompt`／`heal_prompt`／`camping` 沒出現 |
| 戰鬥內的擲骰變了 | `battle-ambush`、`battle-missile`、`battle-pickup`、`battle-spell` | 檢查點截圖雜湊不同 |

**沒有一項指向 bug。** 三種形狀都是同一串亂數往後推的下游效應。

### 重生卡在哪（2026-09-07 實測）

前置條件已經滿足（遭遇、天候、攻擊結算的完整鏈都對過了），流程也跑得動，
但 `arena` 重生出來的收據**還是紅的**，錯誤與重生前一模一樣
（`未知的野外操作 token："SPACE"`）。

`regen_long_path.py` 依**舊軌跡**把路徑切成「非戰鬥段照抄、戰鬥段用 driver 重打」。
`arena` 切成 7 段之後：第 1 場與第 3 場 driver 產出 **0 個 token**
（`workplace/tmp/regen/arena-1-out.txt` 與輸入同大小），只有第 2 場打出 632 個，
舊的 2973 個 token 變成 821 個。

**照抄那一段才是問題**。競技場是走到 (55, 13) 觸發的固定地點（RE-077），
路線本身是確定的；但**路上多出一場隨機遭遇**，第 77 步時隊伍不在競技場而在戰鬥中，
於是後面照抄的野外 token 全部對不上。切段依舊軌跡，遭遇的步數一變就錯位。

`regen_long_path.py` 已經改成**野外那一截交給 driver 走**（只照抄到踏進野外那一步，
之後用世界目標）。driver 確實走到了每一個目標格（43,10 → 52,13 → 55,13），
但**每一場戰鬥仍是 0 個 token**，而且走到 (55,13) 那一段跑滿了 400 輪的上限
——2973 個 token 變成 1177 個，收據還是不完整。

下一個要解的是**戰鬥段的前綴**：driver 走到目標格就停，那時還沒遇敵，所以接下來
的戰鬥段拿到的是「不在戰鬥中」的狀態，打不出東西。要嘛讓走路那一段走到「遇敵為止」
而不是「到座標為止」，要嘛把兩段合成一段交給 driver。`DRIVE_LIMIT` 也要重估。

⚠ 重生會直接改 `docs/verification/player-path/*.json`。跑之前先看 `git status`，
跑壞了 `git checkout -- docs/verification/player-path/` 還原；`regen_long_path.py`
本身也要求軌跡是舊的，被 `replay_player_path.py` 覆寫過就得先還原再跑。

重生要注意的：**換個種子救不了**。天候也吃亂數（`internal/world/session.go:222`），
而 `flow-quest` 這類收據要求檢查點與 `world-quest` 的對拍截圖逐 byte 相同——
換種子會換天候，那批對拍截圖得跟著重生。流程是
`regen_long_path.py` → `replay_player_path.py` → `remap_receipt_steps.py` →
再 `replay_player_path.py` 驗證 → 刪過期截圖。

**要等的前置（做掉一半）**：原版現在跑得動，野外遭遇這一支對完了——
**設好狀態之後兩邊算出同一個怪物、等級、隻數與站位組**
（`internal/oracle` 的 `TestAmbushOutcomeMatchesOriginal`，五組含下限的負對照）。
骰值釘死只為了讓亂數不介入；比的是機制。

原版那一側的三個前置寫在 `docs/re/RESEARCH-LOG.md` 最後一節：資料要先載入
（否則機率恆 0，那條路等於不存在）、`word_2FA2C = 20` 跳過騎走那一問、
`Call` 等不到返回是正常的（算完之後接著畫戰場）。

天候也對完了（`TestWeatherMatchesOriginal`，九組含上下限與冬天那條
「推到第 3 段再多一段」）。天候在 `word_2DF5A`，那個全域是用差分＋
「設好狀態看誰照規則走」找出來的。

還沒對的：戰鬥內的結算。**入口與出口已經找到**（`docs/re/RESEARCH-LOG.md`
最後一節）：`sub_26D76` 吃一個參數，`arg0 == 0` 才走攻擊那一條；當前行動者索引
在 `ds:9E5A`；最小狀態是兩名參戰者的「力量 ＋ 健康」不能是 0（否則 `R6003
integer divide by 0`）、加等級、敵我旗標與招架次數；結果落在 `ds:1C76`
（`0x2FB46`）與 `ds:44AC`（`0x3237C`），基礎攻擊加 50 兩者各加 5。

兩個結果的語意也問出來了（`ds:1C76` 跟 `+0x08` 走且 ÷10、`ds:44AC` 跟 `+0x09`
走且不除），但**對手還沒進到算式裡**——把整筆防禦者記錄塗成 0xFF 都沒有反應。
IDA 探針已經跑過（`tools/ida_probe_koltown_attack_resolve.py`），23 個全域的
對照表在 `docs/re/RESEARCH-LOG.md` 最後一節。**指令碼在 `word_31122`
（`ds:3252`）而不是參戰者記錄的 `+0x25`**——那是先前怎麼設都沒反應的原因。
**攻擊分數的骰前算式已經對完**（`TestAttackScoreMatchesOriginal`，七組含
力量、技能、疲勞比例、累積傷害、這一輪的傷與負重各自的負對照）。做法是把
`sub_26AC2`／`sub_23683`／`sub_24FD3` 三支釘住，中間值保持寫進去的，
驗的就只剩 `sub_26D76` 自己那段。

**防禦分數也對完了**（`TestDefenceScoreMatchesOriginal`，七組）。順帶定下三個
全域：`word_3237E` 防禦者索引、`word_2F942` 防禦分數、`word_2FB46` 攻擊分數，
以及命中判定 `word_2FB46 >= word_2F942`。`+0x24` 是防禦選項（Guard），
remake 的 `Resolve` 原本就對。

**傷害鏈也對完了**（`TestDamageMatchesOriginal`，六組：骰數、骰面、基礎傷害與
骰值各有一組只動它的負對照）。關鍵是命中之後原版先印訊息再 `sub_1B100`
**等玩家點滑鼠**，傷害在那之後才算——不釘住那一支就會看到「命中了但傷害是 0」。

**攻擊結算的數字鏈到此完成**：攻擊分數、防禦分數、命中判定、傷害。

命中部位的算式也讀出來了，三列的表與判準都與 remake 的 `locationOdds` 一致；
`sub_26D76` 的結尾是護甲減傷。肢體的扣傷在 **`sub_277F4`**（位址剛好接在結算那一支
之後），七組也對完了——一般扣傷、剛好扣光、扣過頭、累積傷害到門檻（頭胸一起廢）、
累積傷害夾在 `0x7F`。

**攻擊結算的完整鏈到此對完**：攻擊分數 → 防禦分數 → 命中判定 → 傷害 →
命中部位 → 扣傷。

### 已知的表示差異：廢肢是 −1 還是 0

廢掉的那一肢**原版寫 −1（記錄裡是 `0xFF`），remake 寫 0**。兩邊所有判準都是
「`<= 0` 就算廢」，所以戰鬥行為一致；但記錄裡的值不同，**存檔或畫面若讀那一格
就會差**。要不要改 remake 是忠實度的取捨，改之前先確認沒有別處假設那一格
非負——照 spec-gated 流程走。

對過之後才動收據，先重生等於可能做兩次。

（法術那三個洞**已經全部解掉**：效果 5 的疲勞鏈分母是「力量 ＋ 健康」（RE-129），
效果 7／8 是攻防技能的罰值一點抵十（RE-129），選部位那一次互動照原版做了
（RE-100）。怪物那一側的徒手武器也接上了——記錄 `+0x30` 起七格穿戴欄。
任務物品那一格同樣解完並接上：座標是任務地點記錄的 `+0x27` 列、`+0x28` 欄。）

之後依序：

1. **敵人 AI 最後的缺口**（見下一項）。
   **戰利品畫面整塊做完了**（RE-115）：`sub_1AAE8` 的七個參數、選中列換色而不是
   框起來、捲動與滑鼠命中的判定，以及底下那兩格圖示（YES ＋ U-TURN，連帶蓋掉
   報獎勵那張卷軸——圖示列與卷軸在原版佔同一塊）。但原版收據裡沒有戰利品畫面
   （要打贏一場而且擲中兩成的掉落），所以這一整塊只有靜態證據。
   **打字那一格已經結掉**（RE-114）——`sub_1B5C8` 的欄位、兩種字型與那條
   每 0.27 秒翻一次的橫槓都解出來了，收據 `town-mouth-run1` 與對拍
   `plan-town-mouth.json` 兩對只差原版自己畫的滑鼠游標。
   **遭遇圖那一項也早就結掉**——`[0x1370]` 不是圖號、是 RE-010 解出的尺寸表，
   三十二筆與抽出來的 PNG 逐筆相符（RE-113 的補證）。
2. **敵人 AI 沒有缺口了**。反制掃描那個 0x10 的篩選條件**在原版也恆真**
   （RE-118）：`+0x05` 只由 `sub_21FCD(i, 1)` 設，而那個 kind 是穿戴欄部位
   3／5／7／8 那件東西種類 byte 的最低兩位元——`ITEM.LST` 裡放得進那四格的
   三件（Fur Boots、Fur Cape、Leather Cape）種類 byte 都是 0，`kind = 1` 的
   全是武器而武器不在那四格。**不是缺口，是原版就沒有這條路。**
   **移動段也接上了**（RE-117）：有目標時先試往目標那一步、再試左右
   各一格；沒有目標（或三個都不通）從**隨機方向**起繞八格；八格都站不下就分三條
   ——原地揮武器／1/3 回撤退段／記一次困住再重跑決策樹，而「困住過」之後貼近打的
   攻擊式固定第 1 式。它改變戰況，`arena`、`full-chain` 與 `full-chain-ending`
   三條長路徑的固定按鍵序列因此整條重生（`tools/run_driver.sh`＋
   `regen_long_path.py`＋`remap_receipt_steps.py`），七份長收據全部重驗通過。
   **沒有武器時的指令 0x26 做完了**
   （RE-116：手上那一格空著就去撿地上的東西）。**整條在原版資料下走不到**——
   三十二隻怪全部都有武器，而 `+0x2B` 那個 byte 的兩個寫入點也都寫 0；
   照著實作是為了讓決策樹與原版逐條對得起來。撿到的東西**會用自己的傷害骰**——
   `GroundItem` 帶著那一件的武器數字，掉落物由 `ITEM.LST` 那一筆換算。

## 還沒接進遊戲的

有 RE 證據、規則層或資料，但玩家在遊戲裡碰不到：

- **`dist-all/0.1.2/` 已從乾淨工作樹重建並驗過**（`0.1.1` 那一包是戰鬥數值改動之前
  打的，行為與現在的原始碼不同，留著只作對照）。**版本號仍待使用者決定**：
  `0.1.2` 是暫定的 patch 遞增，還沒對外發布，也沒有 GitHub Release。
- [x] **任務物品開打就擺在戰場上**（RE-113／128）：座標是任務地點記錄的 `+0x27`
  （列）與 `+0x28`（欄），`sub_2B0FF:0x2B1F0` 在任務物品那十六個 byte 的第 0 格
  非零時抄進戰場物品第 0 件的那兩張表。RE-113 記著「那兩個全域被誰讀還沒解」——
  那是過濾器有洞：讀的地方走 `[bx + 2BAEh]` 這種相對定址，xref 一個都看不到。
  記錄本身改走 `good()`，所以撿起來的人拿得到傷害骰與負重。
- **音訊**：母體已普查、TITLE 的 WAV oracle 已取得，SPEC-004 還是 DRAFT，
  一個音都還沒進遊戲。
- **雲層動畫**（`sub_2A18E`）：版面與旋轉單位已定案並實跑驗過（RE-136），
  **還沒接進野外畫面**。天候 3／4 走的是另一條（三張圖輪播），也還沒做。
- **城鎮剩下的圖示**：武器店的鍛造**在原版走不到**（RE-134：說明書說金屬錠要
  「找到」不是買到，但商品表、`ITEM.LST` 60 筆與 24 個任務物品逐筆掃過，
  `+0x0F` 有 bit5 的只有 Cloth Pants 與 Leather Hood，唯一的 Copper Ingot 是
  `0x00`）。**盔甲店的 ANVIL 是另一條**——調整合身、太大的可以調掉多餘重量、
  收少量費用，說明書寫得比武器店那一條詳細，而 RE-125 沒讀那一段的碼。
  那一條在原版走得到，**是值得補的**。
  SCROLL 選單的第二列（法術編輯器 `sub_178E0`）做完也對拍完了（RE-128），
  整頁允許區外零像素差異。**資料頁那兩頁仍然沒對過原版畫面**——
  `town-sheet-items`／`town-sheet-party` 兩個 oracle 情境定義了但還沒跑。
  **資料頁那兩頁都做完了**：物品那一頁（RE-127）有立繪、十七列清單、明細與七個動作
  （拿到手上、穿上、掛腰帶、給隊友、放背包、丟掉、吃掉）；隊伍那一頁（RE-126）
  由後往前畫整隊。立繪的七層合成在 `internal/portrait`，212 張 PNG 的尺寸零處不符。
  **還沒對過原版畫面**——那兩頁的每個座標都來自反組譯，但沒有原版截圖可以逐 byte 比。
- [x] **遭遇時的治療那一條**（怪物編號 0xFF）已接上（RE-112／128）。那六段敘述的
  來源找了很久：DGROUP `0x2350` 在 BSS，xref 與 operand 兩種問法都是零個參照點。
  換一個過濾器掃 `sub_1A4FE` 的每一個載入點才看到——`TOWN0` 的 payload 4 載進
  DGROUP `0x22D0`，`0x2350` 是它的 `+0x80`。每行的第 0 個 byte 是要解幾個 5-bit
  值，不照它截斷句尾會拖出垃圾字母。實跑收據 `world-healer`（種子 23，野外第 14
  步擲到，答應之後全隊傷勢與飢餓歸零）。**那六段的中文譯文還沒做。**

等使用者決策才能動的項目（不擋上面的工作）：

1. `assets/audio/title-thunder.ogg` 的人耳驗收，通過才把 SPEC-004 升 READY。
2. 原版衍生 typed data 的權利分類（六份本機 JSON 是否比照 PNG 進 private repo）。

已於 2026-09-03 決定的三項（見 `CONTEXT.md` 的「已定案的共同決策」）：
存檔格式、戰鬥忠實度與加速開關、客棧住宿是否離隊。
# 最新驗證補充（2026-09-08）

- [x] SPEC-029 舊怪物資料缺欄及損壞值拒絕載入，13 種回歸案例通過。
- [x] 正常原版收據的初始狀態比較由 59 增為 60 欄，納入 SLEDGE 遠程能力。
> 未完成項 `v01-remaining-battle-chain`：V01 剩餘：完整裝備、地形、指令及攻擊結算、正常 GUI 戰後保存鏈。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
> 未完成項 `appimage-rebuild-latest-fixes`：把最新怪物初始化與面板修正重建到新的本機 AppImage。
> 內容與核實方式見本檔「未完成項」一節（權威是 `docs/worklist.json`）。
