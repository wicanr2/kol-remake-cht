# 《武士傳說》重製目前狀態

更新日期：2026-10-24。

使用者收集的 v1.0.5 試玩回報已修正於開發版，契約見
[SPEC-041](docs/spec/041-player-feedback-fixes.md)，結果見
[本輪驗證](docs/verification/player-feedback-20261024.md)。自動續玩、人物／裝備數字切換、
名稱大小寫、清單點選、販賣、攻擊部位、方向長按及滾輪等待頁防穿透已接線。
正式套件測試、Linux 正常建角／住宿／無參數重啟、雙語受控服務與戰鬥選單通過。
發行版仍為下列 v1.0.5，本輪尚未打包或發布；沒有新原版 parity 或平台真機結論。

三平台完整版現為`v.1.0.5-20261018`，各有純EGA及EGA＋HD，共六包，見[完整版驗收](docs/verification/repack-v1.0.5-20261018.md)。
435份正式Go／module與[Windows記憶體修正](docs/verification/windows-memory-20261017/receipt.json)完全相同，1744份資源與v3相同，僅重新編譯、注入新版號及分裝。
三平台均按需讀圖、限制快取及回EGA釋放HD；缺完整HD素材不能切換，壞圖回到EGA。純EGA保留912份資源，移除832份HD圖像與配套；兩版均保留四首配樂及獨立F9。

六包manifest、逐檔SHA、內容、平台與版本通過。Linux實際AppImage解包後經AppRun、Wine實際Windows ZIP正常開場、建角選單、組隊、旅店、F9／F10及缺HD停用通過；四個流程共24張圖，EGA零差、HD最多一色階，中文與模型保持。
兩平台4 GiB無交換空間，OOM皆0；macOS雙架構、arm64簽章結構、最低11.0及系統動態庫通過，尚無真機執行。
規則、圖稿、配樂與kol-save-v3保持，單元與素材深驗沿前輪435份同來源收據。本輪未重跑完整套件、原版oracle或自然全任務通關。

六個完整版只留本機`dist-all/v.1.0.5-20261018/full/`。同版三平台純EGA／EGA＋HD六個引擎包及SHA256SUMS.txt已公開發布，Release ID415941791，七個匿名下載HTTP200且大小／SHA-256全符，見[遠端回讀](docs/verification/release-v1.0.5-20261018/publication.json)。既有tag固定bc087d31，完整版與影片未上傳，舊tag／Release保持；76秒影片仍v.1.0.3。
儲存庫保持public，832份HD圖像配套與6份配樂授權全量收錄，見[公開化審查](docs/RELEASE-RIGHTS-REVIEW.md#儲存庫公開化與hd全量收錄)。
README保留故事背景、原版DOS／中文EGA／HD差異、事件對照及12格總覽；截圖沿既有驗收來源，素材位元組保持。
本檔保存目前真相；歷程見[WORKLOG.md](WORKLOG.md)，可執行項目以[工作清單](docs/worklist.json)為準。

## 目前作業方向

使用者要求的README更新及v.1.0.5正式Release均完成，已設為GitHub最新版。
[引擎包驗收](docs/verification/release-v1.0.5-20261018.md)涵蓋六包內容、四包Linux／Wine實際匯入與正常操作、兩包macOS匯入及24圖與完整版零差。
七個公開附件與GitHub digest、匿名下載SHA全符。準備提交10936e68已push；本次遠端結果另作文件提交，不移動tag。
71份驗收檔案及3份可重建程式雜湊已封存核對，本輪工作區已清理；沒有新增工具映像檔。
採已核准的手繪、1920×1200、遊戲內F10及獨立F9；新圖最長邊768，既有832份HD素材保持。
原生紋理192 MiB及縮放96 MiB是快取預算，不含驅動與畫布；首次HD仍需載入。
Windows原先開場閃退未取得同一失敗收據，Windows／macOS真機及Linux主機FUSE沿原限制。
完整HD基準見[全量驗收](docs/verification/hd-full-runtime-20261017.md)、[全地圖](docs/verification/hd-all-maps-runtime-20261017.md)及[事件](docs/verification/hd-scene-events-runtime-20261017.md)。
收尾依SPEC-040受控場景，原版C matching保持暫停，不重新開啟已完成美術或另改玩法。

## 目前狀態表

| 項目 | 現況與證據 | 限制 |
|---|---|---|
| 本輪開發修正 | [試玩回報驗證](docs/verification/player-feedback-20261024.md)：正式套件測試、兩人建角與各扣60後自動重載、數字換人、鍵鼠、雙語販賣／護甲／法術、受控攻擊部位通過 | SPEC-041 保持 READY；未打包，未重驗原版／HD／Windows／macOS 真機 |
| 匹配原版來源 | [持續來源還原](docs/re/matching-decompilation.md#持續c來源還原)：228純C74428、378 LIB44396、182混合75345 bytes。圖像／人物選擇／事件／分頁新增四函式4203，分列3701 C／502 ASM；十一個真實CODE物件、36項乾淨產物、116反例及原生packed雜湊通過 | 尚10函式、4999 CODE bytes，含41函式外bytes；DATA來源完整，12個未匹配試驗另列。原始C宣告／語意不唯一，四支完整來源保持 |
| 發行版來源 | 435份Go／module、1744份資源；[引擎發布收據](docs/verification/release-v1.0.5-20261018/receipt.json)鎖定，三個執行檔與同版完整版相同 | 程式沿已驗記憶體修正，832份HD圖像配套及規則／v3保持；不含本輪開發修正 |
| 發行版最近測試 | 六引擎包內容通過；四個Linux／Wine實際匯入及正常開場、建角選單、組隊、F9／F10與24圖通過，與完整版零差、文字／模型相同，4 GiB無OOM；七附件匿名下載及SHA相符 | 此列為 v1.0.5 發行時驗收；本輪開發測試另見上列 |
| 長任務與保存 | 前版 [RE-291](docs/re/20261006-visibility-current-regression.md) 雙語11976長任務狀態及26新程序重載；完整v3只差saved_at | 屬前版來源，本輪未重跑；固定seed1、明示遭遇骰100，不是自然全旅程 |
| 長任務畫面 | 40 張中 38 張保持；兩張逃離頁的差異只在戰場視圖，原圖保留 | 與移動快取修正的關係為強推論；沒有本頁原版 oracle |
| 跨平台 | 三平台v.1.0.5各兩個完整版與兩個公開引擎包；Linux／Wine實際操作，macOS雙架構、結構及實際匯入通過 | Windows／macOS真機及主機FUSE未驗，Wine不等於Windows真機 |
| 正常 AI 抽樣 | [反制掃描](docs/re/20261006-ai-counter-normal.md)、[成功側翼](docs/re/20261006-clockwise-flank-normal.md)、[成功撤退](docs/re/20261006-flee-normal.md) 已有有限正常案例 | 其他戰局、困住撤退及成功怪物拾取仍開放 |
| 最新原版試驗 | RE-295 真保存角色正常出城，seed1、169延伸鍵、170狀態；與現行首次紮營及111個同期世界標籤／日期框相符 | 本上限內沒有自然騎走問答；紮營後26個方向鍵原版忽略、現行換日，完整野外對拍未通過 |
| 怪物拾取證據 | [RE-292](docs/re/20261006-monster-pickup-low-strike.md)：883狀態、10回合、2264次自然亂數；完整16-byte掉物1筆 | 未取得站立空手怪物或成功拾取，不延長同場挑結果 |
| 資料格式 | 怪物 v2 已正式接線，四套必填、缺組拒絕載入；人物保存保持 kol-save-v3 | 世界時間已保存；戰鬥暫態依 SPEC-023 排除，中途保存政策未定 |
| 送禮 | [SPEC-031 第17節](docs/spec/031-item-detail-page.md) ENTER／SPACE 確認與完整物品搬移已正式驗收；六人確認畫面漏繪已修 | 選收件人版面、滿背包、內層取消及類型 0 明細仍有 DRAFT 分支 |
| 標題音訊 | 使用者已聆聽接受；[SPEC-004](docs/spec/004-ogg-audio-catalog.md) 與 [RE-228](docs/re/20261004-title-audio-runtime.md) 已完成有限正式 OGG 播放驗收 | 硬體時序為規格近似，不宣稱逐波形或實機時鐘一致 |
| 首版資產 | [EGA 重建收據](docs/verification/finish-audit-20261003/receipt.json)：30 組邏輯 archive、1022 payload 已分類；40 組 manifest、773 PNG 雜湊符合現行 | CGA／Hercules 是已決定的後續選項，不重新開啟 EGA 完成項 |
| 中文與版面 | [受控場景驗收](docs/verification/controlled-scenes-20261011.md)通過目前登記畫面，含商店、訓練、馬廄、法術、任務與七段結局；翻譯缺項0，6115字串缺字0 | 姓名、玩家輸入與精靈語法術名依現行契約保留；未聲明所有動態分支已驗 |
| 正常購馬 | [RE-293](docs/re/20261006-steed-trade-normal.md)：合法名冊單人、輕型戰馬400換馱馬50、住宿60與真正保存重載，兩版終值2490；坐騎16-byte、移動率10及中文名牌已核對 | 正式版仍合併韁繩訊息、先換馬並回清單；原版分段確認與回招呼頁的原型待選，DRAFT未接正式 |
| HD | 目前玩家路徑13類圖形及操作介面已接入並限定CONFORMED，含地圖、天候、戰場、傷勢、標記、片頭、翻書、事件及所有共用操作頁；[13類覆蓋](docs/verification/hd-completion-20261014.md#目前覆蓋)、[完整驗收](docs/verification/hd-full-runtime-20261017.md) | 本版六包通過；原版oracle未知及未抽測玩法邊界另列，不阻塞已驗HD圖形；自訂點陣與資料PNG保持 |
| 交付 | `v.1.0.3-20261017`：三平台full／release、76秒1080p影片；Release四附件大小及SHA相符，目前隨public儲存庫公開；[交付收據](docs/verification/release-v1.0.3-20261017.md)及[公開化清冊](docs/verification/public-repository-20261017.json) | 完整版及影片本機；無真機驗收 |
| GitHub Issue | 2026-10-14關閉十一個已有完成證據的Issue，回讀皆completed；[收尾收據](docs/verification/release-v1.0.2-20261013/issue-closeout-20261014.json) | 遠端十三項保持開放；本機另含未建Issue項，本輪HD本機項結案；遠端Issue不變 |

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
| 已排後續或待取捨 | CGA／Hercules、三項加速細則、戰鬥中途保存及各 DRAFT 原型 |

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
- 目前 **15 條未完成**（11 條沒有機器訊號、標 `manual`），
  另有 94 條已核實完成——後者跑的是回歸檢查，訊號不見了就開口。
  未完成項的 verify 種類：`manual` 11、`present` 4。
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
| hd-art-ui | 已完成13類HD、中文、F10、保存回歸及本版六包驗收；真機仍依獨立項目 |

HD採已核准的1920×1200原比例，各批完成範圍以目前狀態表及正式收據為準。
全量圖形與新版封包已依[SPEC-039](docs/spec/039-hd-art-ui.md)交付閘門通過，不重新開啟已確認選項。
[送禮選人](workplace/gift-recipient-layout-20261005/review.html)仍依原DRAFT閘門處理；
[住宅告示](docs/re/20261006-residence-sign-prototype.md)的兩行「請勿／踐踏草坪」已核准並接入本版。

## 有效共同決策

| 日期 | 議題 | 已確認決定 |
|---|---|---|
| 2026-10-17 | 儲存庫公開與HD全量收錄 | 使用者明確要求現有repo轉public；838份正式HD圖像／配套／配樂已全在Git，既有歷史一併公開。原作散布授權未確認，RRSAL與第三方權利保持；舊tag、Release及素材雜湊不改 |
| 2026-10-16至17 | 全量HD及縮圖 | 核准同一手繪風格，連續完成並commit＋push至private repo；遊戲素材也縮小，後續新圖最長邊768；畫布1920×1200及F10保持 |
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

GitHub現況沿本頁目前狀態表與最新主機gh核對；本機計數不代替遠端Issue。matching第六至三十四批與C35屬歷史交接，本輪續作全畫面HD。原始輸入、完整模型、存檔與封存證據僅本機。
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
