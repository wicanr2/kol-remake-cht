# 《武士傳說》重製目前狀態

更新日期：2026-10-08。正式來源含 SPEC-008 受阻打暈與任務入口亂數接續，
正常購馬保存為 RE-293，HD 工作批次為 RE-294。完整重製尚未完成。
本檔保存目前狀態、有效決策與下一閘門；歷程見 [WORKLOG.md](WORKLOG.md)，
可執行項目及完成條件以 [docs/worklist.json](docs/worklist.json) 為準。

## 目前作業方向

2026-10-08使用者要求暫停目前對拍，改研究原版C的matching decompilation。
對拍維持暫停，沒有執行中的專案容器。新方向先做
[原始工具鏈與匹配工作](docs/re/matching-decompilation.md)，KOLTOWN的51個純C函式共5057 bytes，
42個固定函式庫函式4214 bytes，經QuickC2.01／LINK4.07原位匹配；EXEMOD4.02與EXEPACK4.06完成原版packed EXE重建，
133760 bytes與原版SHA-256 c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221完全相同。
KOLBOOK另有九個C函式594 bytes、36個函式庫函式3705 bytes，完整40192-byte EXE與原版相同。
KOL、TITLE、MAKEMODE另有101個固定函式庫函式10080 bytes，三支完整EXE均與原版相同。
TANDYPIN由13-byte組合語言入口、3-byte連結填補與96-byte重複字串完整還原來源，原生重建624 bytes相同。
合計60個C函式5651 bytes、179個函式庫函式17999 bytes及一個組合語言入口13 bytes。
本批16項乾淨產物、九項反例及十二份新舊來源核對通過。六支EXE都有同位元組重建基準，五支仍為部分來源。
目前[來源清冊](docs/verification/local/matching-c4-20261008/source-ledger.json)保存全部799個原始函式，
尚有559個函式、176286 CODE bytes與13788初始化DATA bytes未還原；完整反編譯尚未完成。
現有Go程式、已完成證據及原完成條件保留；原版C重建與Go移植的證明範圍分開處理。

## 目前狀態表

| 項目 | 現況與證據 | 限制 |
|---|---|---|
| 匹配原版來源 | [持續來源還原](docs/re/matching-decompilation.md#持續c來源還原)：60C／5651 bytes、179個函式庫函式／17999 bytes、TANDYPIN完整組合語言來源13 bytes；六支完整EXE均有同位元組重建基準。本批16項乾淨產物、九項反例及十二份來源核對通過 | 五支部分來源仍有559函式、176286 CODE bytes與13788 DATA bytes；混合來源全檔相同不代表完整還原。特殊符號原始定義與精確工具版本仍未知 |
| 正式來源 | 340份Go、342份程式輸入；[任務亂數接續](docs/verification/quest-rng-20261007.md)與[受阻打暈持物](docs/verification/blocked-stun-20261007.md)有限CONFORMED；RE-295出城移動率及RE-290移動呈現保持 | 最新只改main.go既有任務入口；舊H名稱、原始記錄未知欄位及整場末端仍未閉合 |
| 最近測試 | 十九套件通過；雙語正式及觀察前綴3540狀態／108PNG，各753個原版戰鬥點零差；正常回城保存及新程序重載4236狀態／72PNG，完整v3核對通過 | 重播原前綴，不增加原版覆蓋；回城末端只驗remake |
| 長任務與保存 | 前版 [RE-291](docs/re/20261006-visibility-current-regression.md) 雙語11976長任務狀態及26新程序重載；完整v3只差saved_at | 屬前版來源，本輪未重跑；固定seed1、明示遭遇骰100，不是自然全旅程 |
| 長任務畫面 | 40 張中 38 張保持；兩張逃離頁的差異只在戰場視圖，原圖保留 | 與移動快取修正的關係為強推論；沒有本頁原版 oracle |
| 跨平台 | [最新回歸](docs/verification/platform-quest-rng-20261007.md)：現行來源Linux逐byte重生、Windows AMD64及macOS兩架構／universal建置通過；Linux／Wine原正常鏈4620狀態、808PNG及Windows最新任務3888狀態、90PNG符合基準；完整v3與新程序重載通過 | macOS未執行，Wine不是真機；真機、FUSE、正式封包與原完成條件保持 |
| 正常 AI 抽樣 | [反制掃描](docs/re/20261006-ai-counter-normal.md)、[成功側翼](docs/re/20261006-clockwise-flank-normal.md)、[成功撤退](docs/re/20261006-flee-normal.md) 已有有限正常案例 | 其他戰局、困住撤退及成功怪物拾取仍開放 |
| 最新原版試驗 | RE-295 真保存角色正常出城，seed1、169延伸鍵、170狀態；與現行首次紮營及111個同期世界標籤／日期框相符 | 本上限內沒有自然騎走問答；紮營後26個方向鍵原版忽略、現行換日，完整野外對拍未通過 |
| 怪物拾取證據 | [RE-292](docs/re/20261006-monster-pickup-low-strike.md)：883狀態、10回合、2264次自然亂數；完整16-byte掉物1筆 | 未取得站立空手怪物或成功拾取，不延長同場挑結果 |
| 資料格式 | 怪物 v2 已正式接線，四套必填、缺組拒絕載入；人物保存保持 kol-save-v3 | 世界時間已保存；戰鬥暫態依 SPEC-023 排除，中途保存政策未定 |
| 送禮 | [SPEC-031 第17節](docs/spec/031-item-detail-page.md) ENTER／SPACE 確認與完整物品搬移已正式驗收；六人確認畫面漏繪已修 | 選收件人版面、滿背包、內層取消及類型 0 明細仍有 DRAFT 分支 |
| 標題音訊 | 使用者已聆聽接受；[SPEC-004](docs/spec/004-ogg-audio-catalog.md) 與 [RE-228](docs/re/20261004-title-audio-runtime.md) 已完成有限正式 OGG 播放驗收 | 硬體時序為規格近似，不宣稱逐波形或實機時鐘一致 |
| 首版資產 | [EGA 重建收據](docs/verification/finish-audit-20261003/receipt.json)：30 組邏輯 archive、1022 payload 已分類；40 組 manifest、773 PNG 雜湊符合現行 | CGA／Hercules 是已決定的後續選項，不重新開啟 EGA 完成項 |
| 中文與版面 | [RE-261](docs/re/20261006-cjk-v2-name-population.md) 名稱字寬核對與 [RE-286](docs/re/20261006-world-prompt-longest.md) 正常最長世界提示證據保留；[RE-285](docs/re/20261006-sheet-subpages.md) 法術／任務子頁已有限抽驗 | 未訪畫面、圖片內英文、其他最長字串、職業／專名及完整名牌仍需核對 |
| 正常購馬 | [RE-293](docs/re/20261006-steed-trade-normal.md)：合法名冊單人、輕型戰馬400換馱馬50、住宿60與真正保存重載，兩版終值2490；坐騎16-byte、移動率10及中文名牌已核對 | 正式版仍合併韁繩訊息、先換馬並回清單；原版分段確認與回招呼頁的原型待選，DRAFT未接正式 |
| HD | 使用者選定手繪重繪；[RE-229](docs/re/20261004-hd-art-ui-prototype.md) 完成方向樣張；[RE-294](docs/re/20261006-hd-render-dependencies.md) 核對現行804張PNG、43圖組與八批繪圖依賴 | 只核准媒材；解析度、版面、造型、構圖、素材契約與正式整合尚未核准；資料PNG與結構遮罩另列 |
| 交付 | 打包防護與資料組裝已驗，入口 [tools/build_dist.sh](tools/build_dist.sh) | 正式版號與外部發行未定；本輪沒有封包、影片或 Release |

現行 Linux 執行檔 SHA-256：
`2a885b9a37e884d1830010f7ccb1051b030fe4b14c678fbb71a25955a745fd66`。
這是驗收時建置的開發候選雜湊。執行檔屬可重建產物，清理後不保留位元組；
Windows／macOS最新來源已重建；完整雜湊及限制見[平台回歸](docs/verification/platform-quest-rng-20261007.md)。
重建入口與原命令保留在壓縮證據中。

## 完成標準與證據限制

使用者要求全部玩家可見文字有繁體中文，手冊功能與機制須逐項核對原版。
尚未通過全部閘門，既有候選不能稱為完整重製或全中文完成。

2026-09-26 已決定驗收機制一致：場景、座標、遭遇、訊息記錄、勝負及保存欄位。
截圖作視覺核對，不要求整張畫面逐位元相同。不同旅程或亂數條件不能冒充同狀態。
測試通過只支持該收據範圍；局部呈現、常式與觀察副本不能取代正常玩家路徑。

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
- 目前 **19 條未完成**（15 條沒有機器訊號、標 `manual`），
  另有 89 條已核實完成——後者跑的是回歸檢查，訊號不見了就開口。
  未完成項的 verify 種類：`manual` 15、`present` 4。
<!-- worklist:stats:end -->

下表與十八個原條目逐項對應；不修改完成條件、verify、狀態或 Issue。
精確契約及證據雜湊另見本輪 [逐項稽核](workplace/current-handoff-audit-20261006/gate-audit.json)。

| 工作識別碼 | 下一閘門 |
|---|---|
| g07-battle-speed-options | SPEC-028 細則確認及 READY；三選項控制、保存與可見原型驗收 |
| v01-full-battle-parity | 受阻持物與任務亂數前綴已閉合；補正常整場結算與保存同狀態、舊H名稱及跨場H |
| v02-appimage-platform-acceptance | 主機 FUSE、三平台與推廣影音收據 |
| dosgolem-gui-ai-acceptance | 取得成功怪物拾取的正常原版 GUI；有限反制／側翼／撤退不結案整項 |
| asset-inventory-format-spec | 三格式原條件保留開放；CGA／Hercules 後續再處理，EGA 首版已完成 |
| text-layout-remaining-screens | 補未訪城鎮、世界、戰鬥的 CJK 安全矩形 |
| cjk-remaining-screens | 補全部玩家中文路徑，核對職業／專名與動態名稱 |
| cross-platform-build-smoke | 最新來源三平台已重建，Linux／Wine雙語正常鏈及任務保存通過；補macOS執行與平台真機 |
| longest-cjk-visual-regression | 補其他最長字串的正常路徑、畫面與輸入收據 |
| windows-macos-real-machine | Windows／macOS 真機啟動並走正常玩家路徑 |
| github-release | 公開方式及外部發布決定，再依原條件發布 |
| v01-remaining-battle-chain | 保持 v01-full-battle-parity 的完整裝備、指令、結算及戰後保存條件 |
| save-across-battle-and-clock | 決定戰鬥中途保存是否新增；世界時間 v3 保存已完成 |
| dist-rebuild-version-number | 正式版號定案後乾淨重建與驗收 dist-all |
| ui-details-unverified | 補五處UI未驗邊界及RE-295紮營確認；送禮選人版面、空白法術名、類型0明細與[馬廄確認原型](workplace/steed-trade-normal-20261006/review.html)仍待原問題回答 |
| mouse-original-semantics | 回答兩次點擊、右鍵、世界點擊三處原有取捨，依選擇實作與驗收 |
| character-creation-boundary-parity | 重名政策及滿槽／名稱邊界確認，SPEC-005／023 審查後正式修正 |
| hd-art-ui | 比較既有 [解析度與版面原型](workplace/hd-art-ui-20261004/layout-review.html)；再核對造型、構圖、素材契約、中文、命中區、動畫與各平台封包 |

已備的 HD 版面比較為 1920×1200 保留 16:10 與 1920×1080 重新配置。
[手繪樣張](workplace/hd-art-ui-20261004/review.html) 的媒材已核准，整張生成圖未定版。
[HD工作批次](workplace/hd-art-dependencies-20261006/review.html) 已列圖片與操作畫面的現行依賴；版面原問題保持。
原有待回答問題保持，不因「GO」新增未展示的產品決定。
其他版面原型包括 [送禮選人](workplace/gift-recipient-layout-20261005/review.html)
與 [住宅告示](docs/re/20261006-residence-sign-prototype.md)，仍按各自 DRAFT 閘門處理。

## 有效共同決策

| 日期 | 議題 | 已確認決定 |
|---|---|---|
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
本輪已依授權提交與推送matching研究紀錄；沒有Issue寫入、版號、封包或發行。
Docker、擁有權與暫存終檢見本輪收據及 WORKLOG 末次紀錄。

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
