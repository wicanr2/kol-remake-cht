# 《武士傳說》重製目前狀態

更新日期：2026-10-24。

| 項目 | 目前狀態與入口 |
|---|---|
| 正式程式 | 開發版已接入使用者核准的[SPEC-043字型設定](docs/spec/043-rounded-game-font.md)，[正式GUI與保存驗收](docs/verification/rounded-font-20261024.md)及README截圖完成；既有SPEC-041／042修正與輔助保持 |
| 本機交付 | `v.1.0.7-20261024`六包及124秒新影片全部通過，分為`full/ega/`與`full/ega-hd/`；[驗收入口](docs/verification/repack-v1.0.7-20261024.md) |
| 平台 | 既有v1.0.7封包的Linux AppImage與Windows Wine9操作通過、macOS為結構驗證；本輪字型為Linux GUI及Windows編譯，未新增真機驗收 |
| 資源與存檔 | 開發版461份Go／module／內嵌字模輸入已核對，既有v1.0.7封包為453份；資源仍為純EGA912份／HD1744份。字型與工具列偏好獨立toolbar.json；角色值仍用kol-save-v3 |
| 範圍限制 | 無新增原版parity或自然全任務通關；Windows／macOS未真機驗收，Linux未驗主機FUSE |
| 遠端 | 字型功能提交`2a268057`已推送main並回讀相符；public儲存庫與已發布v.1.0.5引擎Release保持。六個完整版與影片僅本機，沒有新Release或移動tag |
| 工作區 | 字型原型與驗證共154項已完整封存並清空workplace，見[保存入口](docs/verification/retained-evidence.md)；既有C35候選保留，matching仍暫停 |

本輪影片為實際鍵鼠剪輯，準備資料的場景明示標記。原始素材、原始輸入及歷史收據保持。
新版交付根目錄為`dist-all/v.1.0.7-20261024/`；歷程只寫[WORKLOG](WORKLOG.md)。
使用者於2026-10-24明確要求只留最新版；舊版本機交付已刪除，歷史清冊及遠端Release／tag保留。

## 目前作業方向

新增字型設定需求：使用者要求參考`~/cht/zeliard`的圓潤字型。
已查證工具列本來使用粉圓2.1，遊戲中文使用現有Noto字模；Zeliard字級僅控制工具列及設定頁。
使用者已核准[字型比較與F10設定原型](docs/verification/rounded-font-20261024/review/index.html)，
現有字型／粉圓柔邊切換、保留字距與英數字、偏好獨立保存，依[SPEC-043](docs/spec/043-rounded-game-font.md)接入。
正式程式、完整Go回歸、EGA／HD人物頁與旅店、住宿及新程序偏好重載均已[驗收](docs/verification/rounded-font-20261024.md)。
字型往返、英文畫面及人物姓名／數字抽樣均為像素差0；既有六包與影片保持。
重生入口為[原型工具](tools/prototype_rounded_font.py)與[正式字模工具](tools/bake_rounded_game_font.py)。

使用者已要求重新錄製同時介紹v1.0.6修正與v1.0.7新增功能的影片，並重建三平台純EGA／EGA＋HD六包。
本輪交付版本為`v.1.0.7-20261024`，453份程式輸入及1744份素材已核對，六包及124秒新影片全部通過。
[本輪入口](docs/verification/repack-v1.0.7-20261024.md)記錄建置、重新錄影與實際封包驗收。
依使用者只留最新版的決定，v1.0.6本機交付已清除；381項本輪證據完整封存並核對，workspace已空。
完整版與影片仍只留本機，沒有新增Release或移動tag。

已將使用者確認的工具列與永久強化原型接入正式遊戲，依[SPEC-042](docs/spec/042-assist-toolbar.md)實作並驗收。
網友倉庫為空，使用者已同意依文字說明處理，沒有合併對方程式碼。
工具列、戰鬥規則測試、正常角色修改與住宿重載、README正式截圖已通過；新v.1.0.7六包已含本次功能。
先前要求的修正、commit／push、操作影片及三平台雙版本完整版已完成。
現行交付見本頁狀態表。公開引擎Release仍為v.1.0.5，沒有把本機完整版或影片上傳。
目前不重開matching、原版未知分支或真機驗收。清理後需要舊研究時，依保存入口還原指定目錄。

## 沿用證據與限制

下列研究收據沿用，日期與範圍以原入口為準。現行程式與交付見上方狀態表。

| 項目 | 現況與證據 | 限制 |
|---|---|---|
| 匹配原版來源 | [持續來源還原](docs/re/matching-decompilation.md#持續c來源還原)：228純C74428、378 LIB44396、182混合75345 bytes。圖像／人物選擇／事件／分頁新增四函式4203，分列3701 C／502 ASM；十一個真實CODE物件、36項乾淨產物、116反例及原生packed雜湊通過 | 尚10函式、4999 CODE bytes，含41函式外bytes；DATA來源完整，12個未匹配試驗另列。原始C宣告／語意不唯一，四支完整來源保持 |
| 長任務與保存 | 前版 [RE-291](docs/re/20261006-visibility-current-regression.md) 雙語11976長任務狀態及26新程序重載；完整v3只差saved_at | 屬前版來源，本輪未重跑；固定seed1、明示遭遇骰100，不是自然全旅程 |
| 長任務畫面 | 40 張中 38 張保持；兩張逃離頁的差異只在戰場視圖，原圖保留 | 與移動快取修正的關係為強推論；沒有本頁原版 oracle |
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
前版逐項契約見 [逐項稽核](docs/verification/local/workspace-cleanup-20261024/browse/workplace/current-handoff-audit-20261006/gate-audit.json)，
本機最新交付見[本輪收據](docs/verification/repack-v1.0.6-20261024.md)，公開引擎版見[v1.0.5發布收據](docs/verification/release-v1.0.5-20261018.md)。

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
| ui-details-unverified | 補五處UI未驗邊界及RE-295紮營確認；送禮選人版面、空白法術名、類型0明細與[馬廄確認原型](docs/verification/local/workspace-cleanup-20261024/browse/workplace/steed-trade-normal-20261006/review.html)仍待原問題回答 |
| mouse-original-semantics | 回答兩次點擊、右鍵、世界點擊三處原有取捨，依選擇實作與驗收 |
| character-creation-boundary-parity | 重名政策及滿槽／名稱邊界確認，SPEC-005／023 審查後正式修正 |
| hd-art-ui | 已完成13類HD、中文、F10、保存回歸及本版六包驗收；真機仍依獨立項目 |

HD採已核准的1920×1200原比例，各批完成範圍以目前狀態表及正式收據為準。
全量圖形與新版封包已依[SPEC-039](docs/spec/039-hd-art-ui.md)交付閘門通過，不重新開啟已確認選項。
[送禮選人](docs/verification/local/workspace-cleanup-20261024/browse/workplace/gift-recipient-layout-20261005/review.html)仍依原DRAFT閘門處理；
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
- 本輪文件稽核：[準備腳本](docs/verification/local/workspace-cleanup-20261024/browse/workplace/current-handoff-audit-20261006-prepare.py)、[唯讀核對](docs/verification/local/workspace-cleanup-20261024/browse/workplace/current-handoff-audit-20261006/verify.py)、[收據](docs/verification/local/workspace-cleanup-20261024/browse/workplace/current-handoff-audit-20261006/receipt.json)。核對用固定 Go 映像，專案掛 `/src:ro`、稽核目錄掛 `/out:ro`，執行 `python3 -B /out/verify.py`；UID/GID 1000:1000、無網路、180 秒外層逾時、1 GiB、1 CPU、96 程序。
- 整理前完整歷史：[不可變 RE-292 快照](docs/verification/local/workspace-cleanup-20261024/browse/workplace/monster-pickup-low-strike-20261006/documents/CONTEXT.md)，SHA-256 `aecb7fc8674592a2449a00d0522722089b5bd51eaf4f365f9b005b71d36364e4`。本輪保留同版原文，逐輪歷程歸 WORKLOG，不再把舊候選表當目前狀態。

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
