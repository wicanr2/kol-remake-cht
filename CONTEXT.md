# 《武士傳說》重製目前狀態

更新日期：2026-10-13。收尾改採 SPEC-040 受控場景與中文畫面驗收，指定場景已通過。
新版`v.1.0.2-20261013`首次候選的Windows HD最高差2／255。使用者先選擇維持一色階並暫緩發布。固定整數雙線性取樣已接入；正式三環境三十組及實際封包二十組，EGA零差、HD最多一色階。六包、正常1046狀態及13新程序重載、F9／F10、實際音訊與76秒影片均通過。使用者在最終確認後以go解除暫緩，授權指定private Release的三個引擎包與校驗碼；正在發布與遠端核對。原失敗保留。入口：[新版交付及診斷](docs/verification/release-v1.0.2-20261013.md)。
HD 首批門面與六人肖像已核准並接入1920×1200遊戲，首批有限CONFORMED；其他批次維持DRAFT。
遊戲內「畫面 F10」設定已核准並接入正式遊戲，切換與正常住宿存讀檔驗收通過。
操作介面及人物頭像已獲使用者「ok 同意開工」核准並[接入正式Go呈現](docs/verification/hd-interface-runtime-20261012.md)，SPEC-039此批有限CONFORMED。此批為開發版，既有Release保持。
四首原創室內樂已由使用者接受並接入開發版；[配樂收據](docs/verification/hd-music-runtime-20261013.md)為有限CONFORMED。配樂預設開啟，F9及下緣按鈕獨立切換，換EGA／HD持續播放；本次遊玩有效，不改v3。標題／旅店跨輪、準備世界至戰鬥、缺件、正常住宿保存及七狀態重載通過。現行Release與封包保持。
本檔保存目前狀態、有效決策與下一閘門；歷程見 [WORKLOG.md](WORKLOG.md)，
可執行項目及完成條件以 [docs/worklist.json](docs/worklist.json) 為準。

## 目前作業方向

使用者要求操作介面與人物頭像 HD 化、同一主題的對比圖，以及遊戲中切換主題。
[候選比較頁](docs/verification/hd-ui-theme-20261012/review.html)沿核准六人與門面，加入紅黑底、銅框、HD 中文、操作圖示與旅店老闆頭像。
提供旅店、門面、戰鬥三組目前 HD／候選對比，亦可看 EGA／HD；兩側來源場景及封存 trace 相同。
[瀏覽器檢查](docs/verification/hd-ui-theme-20261012/browser-check.json)通過42項實際操作檢查，含四尺寸切換、鍵鼠及中文安全矩形自檢。
這是可丟棄原型，未執行規則或寫保存；其他 NPC、組合裝備、書本、地圖及傷勢遮罩未完成。
本批已依[SPEC-039契約](docs/spec/039-hd-art-ui.md#操作介面與人物頭像候選)接入原生HD字型、面板、操作列與旅店頭像。
七路徑71狀態與中文紀錄相同，EGA畫格保持；15個實際F10事件、鍵鼠、兩視窗及缺件回退通過。
正常住宿v3只差saved_at，另起程序重載7狀態相同。原型250字模與正式字集像素、字距相同，現行2608字完整涵蓋。
四首原創室內樂及獨立F9控制已核准並接入，見[配樂驗收](docs/verification/hd-music-runtime-20261013.md)。目前接續[三平台新版交付](docs/verification/release-v1.0.2-20261013.md)，舊版保持；完整版與影片留本機。
既有F10決定有效；本輪重播封存輸入，沒有新增原版抽樣或重建交付。比較證據已[封存](docs/verification/local/hd-interface-runtime-20261012/retention.json)，工作區已核對後清理。

使用者於2026-10-12核准已展示版面並要求開始，遊戲內畫面設定已接入正式入口。
點「畫面 F10」或按F10，選EGA／HD立即預覽，Esc或返回按鈕留在原場景。
[正式驗收](docs/verification/display-switch-runtime-20261012.md)確認三張畫面與核准樣張逐像素相同，
滑鼠、鍵盤、右鍵、兩種視窗與缺素材中文提示通過。十五個切換事件的場景及隊伍trace保持，
切換後正常住宿的完整v3除saved_at與未切換相同，新程序重載的七個狀態及畫面相同。
[SPEC-039設定契約](docs/spec/039-hd-art-ui.md#遊戲內畫面設定)升有限CONFORMED；整體HD仍為DRAFT。
最新交付為 `v.1.0.1-20261012`，三平台完整版已包含此功能，前版保持。
六包與76秒影片已通過[本版驗收](docs/verification/release-v1.0.1-20261012.md)，發行來源6d98fb00已推送。
使用者已明確授權指定目的地與四個附件，[GitHub Release](https://github.com/wicanr2/kol-remake-cht/releases/tag/v.1.0.1-20261012)已發布。
四個遠端附件的大小與SHA-256相符，tag指向6d98fb00；儲存庫維持private，完整版與影片留本機。

使用者明確修改驗收規則：允許鎖血、提高攻擊與準備場景，重製版抵達對應場景，
確認中文與畫面正確就通過。依 [SPEC-040](docs/spec/040-controlled-scene-acceptance.md)，
[受控收據](docs/verification/controlled-scenes-20261011.md)已確認六種服務／任務頁、
戰鬥規劃、執行、勝利、戰利品及七段結局；書本、建角、七城建築重用前批固定畫面。
新戰鬥採 seed1、隊伍鎖血、命中後傷害20倍，英中各33狀態非顯示差異0，英文33張圖不變。
結局中文原版小座標及兩處段落行數造成的重疊已修，七段真實譯文的畫布／圖框／文字檢查通過。
舊缺輸入、長流程逾時與競技場里程碑差異保留在[前批收據](docs/verification/cjk-closeout-20261011.md)，
不再作本輪收尾或 HD 開工門檻，不將它們改寫成原版對拍通過。

使用者於2026-10-11核准首批[比較原型](docs/verification/hd-first-slice-20261011/review.html)並要求開始。
門面1774×887與六人圖集1536×1024原樣接入，可由遊戲內畫面設定啟用1920×1200呈現，亦保留`-graphics hd -lang zh-Hant`啟動選項。
按實際肖像RGB指紋匹配，不依姓名或隊伍位置；其他人物及自訂圖回退EGA，保存格式不變。
[首批接入驗收](docs/verification/hd-runtime-20261011.md)完成八條路徑98個HD狀態，與EGA狀態及中文紀錄相同；
修改前65張EGA畫格保持，實際住宿保存除saved_at相同，舊v3重載與原尺寸／縮小視窗點選通過。
首位姿勢／盾牌及門面現有尺寸納入此次核准；外框、圖示、戰場與傷勢遮罩尚未重繪。
[SPEC-039](docs/spec/039-hd-art-ui.md)首批有限CONFORMED，整體DRAFT。所有 PNG 限目前 private GitHub。
原版 C matching 維持暫停；既有 `v.1.0.0-20261011` 封包不覆寫。

2026-10-12 使用者再次要求三平台完整版、release 與推廣影片，採 `v.1.0.1-20261012`。
與最近已核准畫面切換版本的356份Go／module輸入相同。正式套件通過，vet保留兩則既有警示。
固定seed1的唯一正常抽樣1046狀態與13新程序重載完全相同，完整v3只差saved_at。
實際AppImage的十五個設定事件、三張核准畫面、縮放、鍵鼠及切換後住宿保存通過。
Windows實際ZIP的full／匯入後release繪圖及F10切換通過，十組中文EGA／HD畫格逐像素相同。
macOS雙架構、最低11.0、ARM64簽章及系統相依通過，尚未執行。
六包與76秒1080p影片完成，十幕目視、影音與來源檢查通過。
完整版與影片留本機；Release只提供三平台引擎包及其SHA-256，不含素材或完整手冊譯文。
本版權利審查已更新，使用者已明確核准private GitHub Release；正式發布與附件完整性核對完成。
前版[交付收據](docs/verification/release-v1.0.0-20261011.md)及產物保持，matching維持暫停。

2026-10-08使用者要求暫停目前對拍，改研究原版C的matching decompilation。
當時對拍暫停；本輪只依後續指令執行上述一次核查，原版 C matching 保持暫停。
matching 入口為[原始工具鏈與匹配工作](docs/re/matching-decompilation.md)。
六支EXE均有完整同位元組重建基準，完整來源已完成KOL、TANDYPIN、MAKEMODE與TITLE。
第三十四批新增圖像合成、人物選擇、事件與分頁四個完整函式，共 4203 bytes，分列 3701 C／502 ASM；1453 條原始指令全部匹配。
三個純 C 共 3212 bytes，一個混合共 991 bytes。六段硬體 ASM 在編譯前明列，C 的求值順序與未使用結果保留。
全部 254 個既有 compiler 函式保持；前十個 C／OBJ 與前批逐 byte 相同，新函式獨立編為 SUPPORT.C／SUPPORT.OBJ。
Town 現有 174 純 C 64776、84 混合 47159 bytes，混合分列 40268 C／6891 ASM。
十一個真實 CODE 物件各為 53366／13650／1707／3679／5221／6038／5554／6558／5611／6348／4203 bytes；111935-byte 合併只是分析視圖，各真實段均小於 64 KiB。
每個 PUBLIC／fixup 保留物件、段與原始偏移。重建使用[編譯單元清單](docs/verification/local/matching-c34-20261010/seed-compiler-units.json)與 full-build.py；seed-CODE.c 只供閱讀／分類。
三個字型候選與遠指標葉函式仍未匹配。
另 12 個 C／版本／前綴試驗保存於[試驗清冊](docs/verification/local/matching-c34-20261010/probe-summary.json)，未計入還原量；正式工具鏈保持。
Town 全部 6736 bytes 初始化 DATA 來源保持，原版 DATA 片段為 0。
246 個 compiler 初值符號供應 5600 bytes；31 個 LIB 範圍供應 991 實際 bytes 與 88 宣告零 bytes，另 57 原生零 bytes。
兩個近指標表與一個遠函式指標由 15 個真實 compiler fixup 重生；56 個 compiler 對齊零 bytes 另列。
52 個完整 LIB CODE 模組、75 個原始函式保持，涵蓋 9632 實際來源與 13 宣告零 bytes。
兩次乾淨重編的 36 項產物相同，116 反例拒絕；兩份獨立 reader 核全部 258 個 compiler 函式、真實 CODE 物件來源、CODE／DATA、fixup、回填與 2281 筆有序重定位。
LINK／EXEMOD4.02／EXEPACK4.06 原生重生 133679-byte 核心，再生成 81 個 1Ah。
兩次完整 133760-byte KOLTOWN.EXE 均與原版 SHA-256 相同；沒有手修指令、header 或複製原版 tail。
Town 尚有 5 函式、996 CODE bytes；Book 五函式、4003 CODE bytes 保持。
累計 228 純 C 74428、378 LIB 44396 實際來源 bytes、182 混合 C／ASM 75345 bytes，函式外另列 271 C／258 ASM。
目前[來源清冊](docs/verification/local/matching-c34-20261010/source-ledger.json)保持全部 799 個原始函式。
全六支仍有 10 函式、4999 CODE bytes 未還原，未還原 DATA 為 0；四支完整來源。使用者要求先 commit 與 push，matching 目標及遊戲對拍均暫停。
剩餘範圍另含 41 個函式外 CODE bytes；部分 CRT 函式有 235 bytes 已覆蓋、1 byte 未還原，均已納入總量。
接續 [C35 工作計畫](workplace/matching-c35-20261010/PLAN.json)的五個未匹配先行候選與兩份收據已核對保留，完整 C34 基準已恢復至 C35。另一次無號左值字型候選的三個函式仍未匹配；尚未整合或計入目前來源，工作區保留供恢復。
接續時保留 compiler_objects.py、seed-compiler-units.json、全部 CODE seed、CODE 分類與完整 DATA 鏈；恢復前一批不可只取 BULK.C／BULK.OBJ。
原始唯一 C 宣告／文字、未證實語意及 Go parity 的證明限制保持。
現有Go程式、已完成證據及原完成條件保留；原版C重建與Go移植的證明範圍分開處理。

## 目前狀態表

| 項目 | 現況與證據 | 限制 |
|---|---|---|
| 匹配原版來源 | [持續來源還原](docs/re/matching-decompilation.md#持續c來源還原)：228純C74428、378 LIB44396、182混合75345 bytes。圖像／人物選擇／事件／分頁新增四函式4203，分列3701 C／502 ASM；十一個真實CODE物件、36項乾淨產物、116反例及原生packed雜湊通過 | 尚10函式、4999 CODE bytes，含41函式外bytes；DATA來源完整，12個未匹配試驗另列。原始C宣告／語意不唯一，四支完整來源保持 |
| 正式來源 | 359份Go及2份module；四首配樂與獨立F9已接入正式Update／Draw；[來源收據](docs/verification/hd-music-runtime-20261013/receipt.json) | 開發版接入；現行v.1.0.1-20261012保持 |
| 最近測試 | cmd／internal與sceneverify通過；四首OGG、實際標題／旅店跨輪、F9／F10及缺件通過；正常住宿v3除saved_at相同，新程序重載7狀態與中文相同 | HD重載字緣最多1色階的GPU舍入另列；探索／戰鬥採準備存檔，原版整場parity及平台真機未新增 |
| 長任務與保存 | 前版 [RE-291](docs/re/20261006-visibility-current-regression.md) 雙語11976長任務狀態及26新程序重載；完整v3只差saved_at | 屬前版來源，本輪未重跑；固定seed1、明示遭遇骰100，不是自然全旅程 |
| 長任務畫面 | 40 張中 38 張保持；兩張逃離頁的差異只在戰場視圖，原圖保留 | 與移動快取修正的關係為強推論；沒有本頁原版 oracle |
| 跨平台 | [前版回歸](docs/verification/platform-quest-rng-20261007.md)：現行來源Linux逐byte重生、Windows AMD64及macOS兩架構／universal建置通過；Linux／Wine原正常鏈4620狀態、808PNG及Windows最新任務3888狀態、90PNG符合基準；本版六個封包另見[交付驗收](docs/verification/release-v1.0.0-20261011.md) | macOS未執行，Wine不是真機；主機FUSE與平台真機保持未驗 |
| 正常 AI 抽樣 | [反制掃描](docs/re/20261006-ai-counter-normal.md)、[成功側翼](docs/re/20261006-clockwise-flank-normal.md)、[成功撤退](docs/re/20261006-flee-normal.md) 已有有限正常案例 | 其他戰局、困住撤退及成功怪物拾取仍開放 |
| 最新原版試驗 | RE-295 真保存角色正常出城，seed1、169延伸鍵、170狀態；與現行首次紮營及111個同期世界標籤／日期框相符 | 本上限內沒有自然騎走問答；紮營後26個方向鍵原版忽略、現行換日，完整野外對拍未通過 |
| 怪物拾取證據 | [RE-292](docs/re/20261006-monster-pickup-low-strike.md)：883狀態、10回合、2264次自然亂數；完整16-byte掉物1筆 | 未取得站立空手怪物或成功拾取，不延長同場挑結果 |
| 資料格式 | 怪物 v2 已正式接線，四套必填、缺組拒絕載入；人物保存保持 kol-save-v3 | 世界時間已保存；戰鬥暫態依 SPEC-023 排除，中途保存政策未定 |
| 送禮 | [SPEC-031 第17節](docs/spec/031-item-detail-page.md) ENTER／SPACE 確認與完整物品搬移已正式驗收；六人確認畫面漏繪已修 | 選收件人版面、滿背包、內層取消及類型 0 明細仍有 DRAFT 分支 |
| 標題音訊 | 使用者已聆聽接受；[SPEC-004](docs/spec/004-ogg-audio-catalog.md) 與 [RE-228](docs/re/20261004-title-audio-runtime.md) 已完成有限正式 OGG 播放驗收 | 硬體時序為規格近似，不宣稱逐波形或實機時鐘一致 |
| 首版資產 | [EGA 重建收據](docs/verification/finish-audit-20261003/receipt.json)：30 組邏輯 archive、1022 payload 已分類；40 組 manifest、773 PNG 雜湊符合現行 | CGA／Hercules 是已決定的後續選項，不重新開啟 EGA 完成項 |
| 中文與版面 | [受控場景驗收](docs/verification/controlled-scenes-20261011.md)通過目前登記畫面，含商店、訓練、馬廄、法術、任務與七段結局；翻譯缺項0，6115字串缺字0 | 姓名、玩家輸入與精靈語法術名依現行契約保留；未聲明所有動態分支已驗 |
| 正常購馬 | [RE-293](docs/re/20261006-steed-trade-normal.md)：合法名冊單人、輕型戰馬400換馱馬50、住宿60與真正保存重載，兩版終值2490；坐騎16-byte、移動率10及中文名牌已核對 | 正式版仍合併韁繩訊息、先換馬並回清單；原版分段確認與回招呼頁的原型待選，DRAFT未接正式 |
| HD | [首批遊戲接入](docs/verification/hd-runtime-20261011.md)、[遊戲內切換](docs/verification/display-switch-runtime-20261012.md)及[操作介面與NPC頭像](docs/verification/hd-interface-runtime-20261012.md)已有限驗收；[室內樂試聽](docs/verification/hd-music-20261012/review.html)技術通過 | 介面最新批次為開發版；配樂人耳待確認。其他人物、裝備、地圖及傷勢遮罩未完成；資料PNG不重繪 |
| 交付 | `v.1.0.1-20261012`：三平台full／release、76秒1080p影片；private GitHub Release已發布，四附件完整性與tag核對通過，入口[本版收據](docs/verification/release-v1.0.1-20261012.md)及[收尾工具](tools/closeout_release.py) | 完整版及影片僅本機；release需匯入合法資料；無真機驗收 |

上一輪開發候選 Linux 執行檔 SHA-256：
`2a885b9a37e884d1830010f7ccb1051b030fe4b14c678fbb71a25955a745fd66`。
這是驗收時建置的開發候選雜湊。執行檔屬可重建產物，清理後不保留位元組；
Windows／macOS最新來源已重建；完整雜湊及限制見[平台回歸](docs/verification/platform-quest-rng-20261007.md)。
重建入口與原命令保留在壓縮證據中。

## 完成標準與證據限制

目前收尾以使用者 2026-10-11 指定的 [SPEC-040](docs/spec/040-controlled-scene-acceptance.md)為準。
已登記場景的中文與畫面驗收通過，即可進 HD。舊正常長流程與嚴格原版對拍不再阻擋。
本結論只涵蓋收據中的場景；不宣稱所有亂數分支、正常全旅程或嚴格原版一致。

2026-09-26 已決定驗收機制一致：場景、座標、遭遇、訊息記錄、勝負及保存欄位。
截圖作視覺核對，不要求整張畫面逐位元相同。不同旅程或亂數條件不能冒充同狀態。
這是歷史原版研究的判準。研究收據保留原等級，與目前受控場景完成條件分開。

| 分類 | 判定 |
|---|---|
| remake 已完成的有限功能 | 各CONFORMED子切片、任務入口亂數接續、受阻打暈持物與戰後保存、正常保存重載、標題音訊、送禮確認與首版EGA轉換；以各規格／收據範圍為限 |
| 原版仍未知或未完整對拍 | 整場戰鬥終態與保存鏈、成功怪物拾取、部分 UI／操作、原始角色記錄 +0x2C／+0x31 及其他未證實別名 |
| 已排後續或待取捨 | CGA／Hercules、美術 HD 化細則、三項加速細則、戰鬥中途保存及各 DRAFT 原型 |

2026-10-07使用者已確認受阻打暈保留持物；SPEC-008本子切片正式實作及有限驗收通過。
舊H名稱與跨場H仍開放，隔離候選保持原有閘門。
[RE-288 共用任務亂數原型](docs/re/20261006-quest-entry-rng-prototype.md)已依展示分支接正式路徑，
[正式驗收](docs/verification/quest-rng-20261007.md)限任務5前綴及remake回城保存，不擴大整場parity。
RE-290 組合候選的雙語 753 點、13554 角色記錄、50 語意欄及 seed 零差，
解除該有限樣本的九列 134 項差異；正式來源現已包含移動修正與共用亂數接續。
原始記錄未知欄位、完整終態、方向確認、末端隊伍與完整 UI 仍開放。

## 剩餘工作與下一閘門

<!-- worklist:stats:begin 由 tools/worklist.py render 產生，不要手改 -->
- **未完成項的權威是 `docs/worklist.json`**，不是 `WORKLIST.md` 的 checkbox。
  核實用 `tools/run_python_tool.sh tools/worklist.py verify`；`WORKLIST.md` 的
  那一節與這幾行數字都由 `render` 產生。
- 目前 **16 條未完成**（12 條沒有機器訊號、標 `manual`），
  另有 92 條已核實完成——後者跑的是回歸檢查，訊號不見了就開口。
  未完成項的 verify 種類：`manual` 12、`present` 4。
<!-- worklist:stats:end -->

下表列出仍開放的工作。本機封包、登記中文場景及GitHub Release已依各自收據結案。
前版逐項契約見 [逐項稽核](workplace/current-handoff-audit-20261006/gate-audit.json)，
本版交付與發布見[交付收據](docs/verification/release-v1.0.1-20261012.md)。

| 工作識別碼 | 下一閘門 |
|---|---|
| matching-koltown-binary | 接續Town5與Book五個CODE函式；依seed-compiler-units.json保留全部真實CODE物件與CODE／DATA獨立核對器，遊戲對拍暫停 |
| g07-battle-speed-options | SPEC-028 細則確認及 READY；三選項控制、保存與可見原型驗收 |
| v01-full-battle-parity | 受阻持物與任務亂數前綴已閉合；補正常整場結算與保存同狀態、舊H名稱及跨場H |
| v02-appimage-platform-acceptance | 主機 FUSE 與平台真機；本輪六包與推廣影音收據已完成 |
| dosgolem-gui-ai-acceptance | 取得成功怪物拾取的正常原版 GUI；有限反制／側翼／撤退不結案整項 |
| asset-inventory-format-spec | 三格式原條件保留開放；CGA／Hercules 後續再處理，EGA 首版已完成 |
| text-layout-remaining-screens | 補未訪城鎮、世界、戰鬥的 CJK 安全矩形 |
| cross-platform-build-smoke | 最新來源三平台已重建，Linux／Wine雙語正常鏈及任務保存通過；補macOS執行與平台真機 |
| longest-cjk-visual-regression | 補其他最長字串的正常路徑、畫面與輸入收據 |
| windows-macos-real-machine | Windows／macOS 真機啟動並走正常玩家路徑 |
| v01-remaining-battle-chain | 保持 v01-full-battle-parity 的完整裝備、指令、結算及戰後保存條件 |
| save-across-battle-and-clock | 決定戰鬥中途保存是否新增；世界時間 v3 保存已完成 |
| ui-details-unverified | 補五處UI未驗邊界及RE-295紮營確認；送禮選人版面、空白法術名、類型0明細與[馬廄確認原型](workplace/steed-trade-normal-20261006/review.html)仍待原問題回答 |
| mouse-original-semantics | 回答兩次點擊、右鍵、世界點擊三處原有取捨，依選擇實作與驗收 |
| character-creation-boundary-parity | 重名政策及滿槽／名稱邊界確認，SPEC-005／023 審查後正式修正 |
| hd-art-ui | 操作介面、NPC與四首配樂及獨立F9已接開發版並有限驗收；其餘美術與後續曲目依[SPEC-039](docs/spec/039-hd-art-ui.md)及[SPEC-004](docs/spec/004-ogg-audio-catalog.md)，新平台封包／真機仍開放 |

HD 採已核准的1920×1200原比例。首批及畫面設定已有限CONFORMED，其他批次維持DRAFT，
後續依[SPEC-039](docs/spec/039-hd-art-ui.md)確認造型與素材，不重新開啟已確認選項。
[送禮選人](workplace/gift-recipient-layout-20261005/review.html)仍依原DRAFT閘門處理；
[住宅告示](docs/re/20261006-residence-sign-prototype.md)的兩行「請勿／踐踏草坪」已核准並接入本版。

## 有效共同決策

| 日期 | 議題 | 已確認決定 |
|---|---|---|
| 2026-10-13 | HD配樂及控制 | 接受標題／旅店與探索／戰鬥四首室內樂；採A獨立開關、預設開啟，F9與下緣按鈕，EGA／HD切換持續播放；已接開發版並有限驗收 |
| 2026-10-12 | 新版交付及發布 | 三平台完整版、Release與推廣影片；v.1.0.1-20261012包含首批HD及切換，舊版保持；使用者明確核准wicanr2/kol-remake-cht的private Release，只上傳三個引擎包與SHA256SUMS.txt，完整版與影片留本機 |
| 2026-10-11 | 本機收尾交付 | 使用者授權判斷目前狀況、最多一次抽樣後打包v1.0.0；依規則採v.1.0.0-20261011，保留既有未驗項，不外部發布 |
| 2026-10-07 | 任務亂數 | 採比較頁已展示的原版共用世界來源；SPEC-008窄切片正式有限CONFORMED |
| 2026-10-07 | 受阻打暈 | 採原版保留完整持物；成功擊退仍掉物。SPEC-008窄契約已正式有限CONFORMED |
| 2026-10-05 | HD 美術 | 採用手繪重繪媒材；構圖、造型、解析度與介面另行確認 |
| 2026-10-04 | 怪物資料 | v2，四套必填、缺組拒絕載入；人物存檔 v3 保持 |
| 2026-10-04 | 送禮 | 採原版確認：選收件人後保留明細與訊息，ENTER／SPACE 才搬移完整物品記錄並回清單；未驗分支另核 |
| 2026-10-04 | 雷聲 | 已實際聆聽，可接受並接正式播放；候選 OGG 與已接受音量保持 |
| 2026-09-26 | 對拍判準 | 驗機制一致；不要求畫面逐位元相同 |
| 2026-09-25 | 收尾範圍 | 全部玩家文字繁中、手冊功能與機制逐項核對；README 新截圖待整體驗收 |
| 2026-09-07 | 輸入方式 | 可用滑鼠取代鍵盤，以操作結果等價為準；保留功能與保存要求 |
| 2026-09-04 | 授權 | RRSAL-1.0；原版素材不在程式授權範圍，見 [LICENSE](LICENSE) |
| 2026-09-03 | 戰鬥忠實度 | 忠實規則加三個可關閉加速選項，預設全開、原版對拍全關；細則待 SPEC-028 |
| 2026-09-03 | 保存 | remake 自有 JSON 加原版角色匯入；原版檔案唯讀、不寫回 |
| 2026-09-03 | 住宿 | 預設保留隊伍；faithful_inn_stay 切回原版，住宿對拍時開啟 |
| 2026-09-01 | 圖像 | 首版 EGA；CGA／Hercules 為後續選項 |
| 2026-09-01 | PNG 權利例外 | 已驗證、可重生且有來源 manifest 的 PNG 可進 private repo；不授權公開散布 |
| 2026-09-01 | 音訊來源 | 依已證實 PIT 序列與公開規格近似合成 OGG，不散布原版錄音；人耳閘門已由 2026-10-04 決定通過 |

法術名保留原文加繁中注解，依 [SPEC-034](docs/spec/034-cjk-spell-name-gloss.md)；
名牌與注解版面仍依未決原型核對。有效行為及證據等級以
[規格索引](docs/spec/README.md) 和各項收據為準，DRAFT 不進正式路徑。
原版 ZIP、EXE、archive、PDF、音訊及未授權衍生資料僅本機；
公開包與本機含素材包依 [權利審查](docs/RELEASE-RIGHTS-REVIEW.md) 分開。

## 工具、交接與歷史入口

- 最新平台回歸與工具恢復：[驗證入口](docs/verification/platform-quest-rng-20261007.md)。
- 程式與說明書：[README.md](README.md)、[繁中說明書](docs/manual/README.md)。
- 工作：[WORKLIST.md](WORKLIST.md)、[唯一工作資料](docs/worklist.json)。
- 規格與證據：[規格索引](docs/spec/README.md)、[逆向索引](docs/re/README.md)、[研究歷程](docs/re/RESEARCH-LOG.md)。
- 玩家驗收：[正常玩家路徑](docs/verification/player-path/README.md)、[手冊與繁中稽核](docs/verification/manual-fidelity-audit-20260923.md)。
- 最新任務規則：[任務亂數接續](docs/verification/quest-rng-20261007.md)；[受阻持物驗收](docs/verification/blocked-stun-20261007.md)，來源、正常控制器、收據及唯讀核對均由該入口索引。
- 前版平台：[驗證工具](tools/verify_current_platform.py)、[預登記](docs/verification/local/finish-platform-20261007/plan.json)及[收據](docs/verification/local/finish-platform-20261007/receipt.json)；同Go映像掛專案`/src:ro`、工作區`/out:ro`，執行`python3 -B /src/tools/verify_current_platform.py verify`，UID/GID1000:1000、無網路、60秒、1GiB、1CPU、96程序。Windows操作沿既有`coab-wine-smoke:ubuntu-noble-20260826`；macOS建置沿`kol-osxcross-go:20260907`，支援範圍見收據限制。
- 原版執行器：`/home/anr2/cht/dosgolem`；先讀該專案 README.md、CLAUDE.md 與能力矩陣，支援範圍以該矩陣為準。
- Docker 工具鏈：目前Go／Wine／macOS使用[恢復revision與命令](docs/verification/platform-quest-rng-20261007.md)；Go固定1.26.7、Ebitengine2.9.9。前版工具標籤已不可啟動，僅作歷史定位。所有分析、測試、建置與抓圖在有界容器內，原始研究輸入唯讀。
- 本輪文件稽核：[準備腳本](workplace/current-handoff-audit-20261006-prepare.py)、[唯讀核對](workplace/current-handoff-audit-20261006/verify.py)、[收據](workplace/current-handoff-audit-20261006/receipt.json)。核對用固定 Go 映像，專案掛 `/src:ro`、稽核目錄掛 `/out:ro`，執行 `python3 -B /out/verify.py`；UID/GID 1000:1000、無網路、180 秒外層逾時、1 GiB、1 CPU、96 程序。
- 整理前完整歷史：[不可變 RE-292 快照](workplace/monster-pickup-low-strike-20261006/documents/CONTEXT.md)，SHA-256 `aecb7fc8674592a2449a00d0522722089b5bd51eaf4f365f9b005b71d36364e4`。本輪保留同版原文，逐輪歷程歸 WORKLOG，不再把舊候選表當目前狀態。

遠端Issue最近於前輪2026-10-07以主機gh唯讀核對：private儲存庫，24 OPEN。
本機18未完成與遠端尚未完全同步，本機計數不代替遠端現況。
本次依使用者指示保存第六至三十四批 matching 異動與 C35 暫停交接。提交範圍限研究紀錄、工作清單與保存工具；本機原版輸入、衍生來源、工具與封存證據不納入 Git。
Docker、擁有權與暫存終檢見本輪收據及 WORKLOG 末次紀錄。

最新第三十四批來源證據與核對方法見[保存入口](docs/verification/retained-evidence.md)，
來源、原始位址、乾淨重編與反例均由本機收據索引。

## 比較與工作區清理

使用者要求比較完成後刪除該次 `workplace/` 工作目錄。
[任務亂數比較](docs/verification/local/quest-entry-rng-prototype-20261006/review.html)已重查RE-288／289／290封存結果；
RE-290組合候選的753點戰鬥語意及seed零差，262個圓環像素符合實際相位模型。
任務亂數接續已依該展示分支正式有限驗收；比較頁與歷史收據保持。
五個工作區已保存壓縮證據並刪除，約釋放738 MiB；核對方法見
[保存入口](docs/verification/retained-evidence.md)。歷史收據保持原SHA-256。
任務亂數正式驗收工作區也已封存並刪除，另釋放約221 MiB；正式來源與原收據保持。
現行平台回歸亦已封存並清理：1158份證據約6.1 MiB，七份執行檔保留雜湊；
兩個Wine前綴及本次工作區共釋放約1.46 GiB。三個恢復工具revision保留供後續建置。
