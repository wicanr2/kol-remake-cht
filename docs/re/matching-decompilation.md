# 原版同位元組反編譯

日期：2026-10-08。狀態：完整packed KOLTOWN.EXE持續逐byte匹配原版。
目前46個純C函式4309 bytes、八個固定library函式372 bytes；117895個CODE bytes仍未還原。
使用者要求暫停目前對拍，改以 matching decompilation 重建原版C程式，
希望降低重製過程中反覆行為比對的成本。本入口保存查證與首個試點範圍；
現有Go程式、規格、原版收據與未完成條件保留。

## 目前已知

| 項目 | 證據與等級 |
|---|---|
| KOLTOWN工具鏈 | Microsoft C runtime family已證實；精確compiler／linker未知，見[工具鏈證據](toolchain.md) |
| 原版封裝 | Microsoft EXEPACK已證實；EXEPACK4.06能重生全部壓縮bytes與stub，精確歷史版本未知 |
| 其他主程式 | KOL、KOLBOOK、MAKEMODE、TITLE的Quick C1.0／2.01、MSC5.1為FLIRT候選，強推論，不作精確版本 |
| 本機工具 | MSVC1.0／1.52仍為未驗候選；本輪新增釘commit及hash的MSC5.10負例與QuickC2.01匹配候選 |
| 對拍狀態 | 已按使用者指令暫停；現有專案容器清單為空 |

唯讀容器重查原版ZIP，六支EXE如下。原始輸入未修改，沒有執行原版或新對拍。

| ZIP entry | 大小 | SHA-256 |
|---|---|---|
| `Knights-of-legend/KOL.EXE` | 11971 | `4f798b078db889b6166b92c0f0bae1febe3efb3535ad67800b2e0aaed94d7b71` |
| `Knights-of-legend/KOLBOOK.EXE` | 40192 | `2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832` |
| `Knights-of-legend/KOLTOWN.EXE` | 133760 | `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221` |
| `Knights-of-legend/MAKEMODE.EXE` | 15744 | `3051cc981e2cbf62237c1cbb788f4d10b84c8dea4dcd6c2aad67171cd2196c56` |
| `Knights-of-legend/TANDYPIN.EXE` | 624 | `036c39f330a4fe09bc2729756816b869f1154515bf4befba190e767d4e3cdce3` |
| `Knights-of-legend/TITLE.EXE` | 18671 | `8a2879d924d9a9cb01e000bd548c348022c0f922288565015f7560ab8fc6fad6` |

## 完成標準的區別

1. 函式匹配：C在固定編譯器、參數及記憶體模型下產生相同機器碼，逐項登記fixup及外部引用。
2. 解壓映像匹配：程式碼、資料、段配置及relocation符合鎖定的衍生映像。
3. 原始EXE匹配：header、relocation、padding、EXEPACK結果及所有位元組符合原版SHA-256。

三層分開登記。不把相同解壓結果或正規化反組譯稱為完整EXE匹配。
若仍使用原版片段、組合語言或既有library物件，分別列出其範圍；
把原版bytes重新包裝而hash相同，不能當作C反編譯完成。

同位元組重建證明特定建置結果與原版相同，無法追回唯一的變數名、註解、巨集與原始寫法。
取得匹配C後，移植至Go或更換整數寬度、記憶體模型及硬體介面，仍屬另一個實作。
其驗證範圍需另定，不能直接把原版C的匹配結果延伸成Go重製一致。
方法依[decomp.me說明](https://www.decomp.me/faq)及
[decomp-toolkit的逐物件替換流程](https://github.com/encounter/decomp-toolkit)。
後者專用GameCube／Wii，流程可參考，沒有據此認定其工具支援DOS MZ。

## 下一個窄試點

使用者於2026-10-08確認開始此試點。工作目錄為`workplace/matching-pilot-20261008/`，
保存`input-plan.json`、一次性IDA資料庫、`ida-probe.py`、C源檔、編譯命令、objects及差分收據。
Microsoft官方MS-DOS公開repo的`v4.0/src/TOOLS`作候選來源，下載釘commit及Git blob／SHA-256；
沒有據此認定本作編譯器。DOS執行工具見[Dockerfile.matching-dos](../../docker/Dockerfile.matching-dos)。

先以KOLTOWN辨識compiler／linker、memory model、最佳化及library版本，
選代表性的運算、分支與結構／far pointer函式，實測C編譯結果。
已有固定IDA原始定位、bytes及xref可作輸入，原名與位址維持。
首批匹配成立後再建立分段、物件與增量link流程，最後處理原始EXEPACK EXE。

目前不能判定全量matching會比既有路線更快。試點須記錄實際匹配結果與耗時，
再評估擴展。研究、舊工具及所有原版衍生輸出只留本機，沿Docker-only與既有權利界線。

## 首批實際匹配

固定QuickC2.01，參數`/AM /Od /Gs /Zl /G0`，linker4.07。
純C經編譯及實際link後，三個完整函式共268 bytes符合原版，差分為零：

| 原始IDA名稱 | IDA linear | 解壓EXE檔案偏移 | 長度 | 有效C表達與等級 |
|---|---|---|---|---|
| `sub_14427` | `0x14427` | `26663` | 62 | 取目前記錄的高／低byte組成word；已證實的匹配表達 |
| `sub_18CD5` | `0x18CD5` | `45269` | 85 | 交換兩組16-byte記錄；已證實的匹配表達 |
| `sub_1A866` | `0x1A866` | `52326` | 121 | 比較兩個近指標字串，完全相等回1；已證實的匹配表達 |

保留原名、位址及bytes，沒有以語意改名覆蓋原定位。表格中的位址基準各列明，
不把IDA linear、segment:offset及file offset混成同一數值。
這些C寫法能產生相同機器碼，不據此宣稱它們就是原作者的唯一寫法。

原版ZIP經既有CC0解包器原樣重生：來源SHA-256
`0e1da11a1d5a4aa5722dc83dd16c683b4c2c25cec31d3741ee71039840a6b03c`，
輸出138528 bytes，SHA-256
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
三段raw file bytes與IDA輸出相同，再與實際linked MZ比較，沒有指令正規化或差分遮罩。
乾淨重編的C、OBJ、BIND、MZ及MAP皆逐byte相同。

link fixture只綁`_pool`的DGROUP offset`0xBB88`、`_active`的`0xBB82`及CRT anchor`__acrtused=0`。
沒有重建原版資料、CRT、startup、stack或entry。780-byte片段只作連結證據，不能作可玩的原版。
完整原版header、段配置、2281筆relocation及EXEPACK仍未閉合；工作目標保持完整EXE匹配。

MSC5.10先產生的五組OBJ皆不匹配，特別是SI／DI保存順序與指令選擇不同，作負例保留。
QuickC2.01的`sub_1A4E4`純C far-pointer樣本仍不匹配，不能擅稱為原版asm或偷偷補asm。
QuickC2.01／linker4.07因此是本樣本已證實可用的工具組；原版精確版本及全程記憶體模型仍未知。

本機[收據](../verification/local/matching-pilot-20261008/receipt.json)、
[匹配C源檔](../verification/local/matching-pilot-20261008/linked-pilot.c)、
[完整函式差分](../verification/local/matching-pilot-20261008/linked-review-v2.json)、
[乾淨重編](../verification/local/matching-pilot-20261008/repeat-review.json)及
[唯讀核對](../verification/local/matching-pilot-20261008/verify.py)保存全部定位、命令、工具及輸入hash。
原始code、DB、工具binary與C衍生資料只留本機；這次沒有重新啟動原版、重製或遊戲對拍。

已完成比較的工作區依[保存契約](../verification/retained-evidence.md)封存、核對後清理。
下一步是分段與增量link基準，再逐函式用C替換原版物件，分類C、library、asm及原始未匹配bytes。
流程參考[decomp-toolkit維護者文件](https://github.com/encounter/decomp-toolkit)，DOS MZ實作另行建立。

使用者已確認繼續。增量連結研究工作區`workplace/matching-link-20261008/`已完成比較，
核對封存後清理；固定源檔、物件、腳本與收據保存在下節的本機入口。

## 增量連結與第四個函式

沿原始IDA9.4資料庫匯出八個段與338個函式的定位。六個CODE段的起點、長度，
DGROUP、入口`1B8F:001C`與堆疊`2B48:07D0`均由LINK的MAP與原始MZ核對。
這裡的segment:offset是MZ相對載入基址，不能直接當作IDA linear。
原始IDA DATA段涵蓋初始化資料與未初始化空間；按固定解包器輸出的檔案邊界分開，
以免LINK把47968 bytes的未初始化空間寫進檔案。這是可用的重建物件邊界，
原作者的OBJ名稱與切分仍未知。

原始未匹配片段放在`ORIG.OBJ`。純C編譯後的LEDATA、FIXUPP與BAKPAT另作段定位適配，
放在`CSLOT.OBJ`。適配保留編譯器輸出的指令bytes、回填值與外部目標，只轉換位置。
真正的LINK4.07完成所有解析；沒有把已連結機器碼貼回原始EXE。
物件覆蓋檢查拒絕缺洞、重疊及把原版片段列入C範圍。
格式依[TIS1.1 OMF規格](https://openwatcom.org/ftp/devel/docs/omf.pdf)的
COMMON段、LEDATA、FIXUPP與QuickC的BAKPAT規則。

新增的跨段呼叫函式完整匹配：

| 原始IDA名稱 | IDA linear | 解壓EXE檔案偏移 | 長度 | 有效C表達與等級 |
|---|---|---|---|---|
| `sub_1B5A8` | `0x1B5A8` | `55720` | 32 | `return sub_2D538() % n + 1;`，已證實的匹配表達 |

`sub_2D538`仍保留原版bytes。LINK以原位PUBLIC解析遠呼叫，目標為MZ相對
`1B8F:1C48`。檔案中的segment word為`0x1B8F`；IDA載入基址為`0x1000`，
因此IDA原始匯出顯示`0x2B8F`。這四個fixup bytes現已完整核對，沒有遮罩或忽略。
本函式原始檔bytes SHA-256為
`5cfb827ff3a09bcc3fbb26a9ad9bbd9e4674233b606a2875810734d51d0ad343`。
首批另外三函式的定位與hash維持。

增量LINK的初始化映像129312 bytes與原版逐byte相同；2281筆重定位的
實體位置與目標、入口與堆疊也相同。範圍清冊明列：

| 來源分類 | bytes | 證明範圍 |
|---|---|---|
| 純C編譯 | 300 | 四個完整函式與原位外部引用均匹配 |
| 原始未匹配片段 | 129012 | 只作增量連結基準，沒有C還原進度 |
| 已匹配library／重寫asm | 0 | 本輪沒有新增此類匹配 |

`_active`與`_pool`已由原版DGROUP的實際BSS位置解析，取代首批的絕對位址fixture。
`__acrtused=0`仍作無程式引用的編譯器link anchor，原版CRT仍列未匹配。
這不證明原作者曾使用相同PUBLIC名稱。

完整native MZ仍與固定解包器輸出不同：

| 標頭欄位 | 固定解包器輸出 | LINK4.07輸出 |
|---|---|---|
| `e_minalloc` | 2308 | 3123 |
| `e_csum` | 0 | 57581 |
| `e_lfarlc` | 28 | 30 |

重定位項目的順序與segment:offset編碼也不同，雖然全部實體位置和目標相同。
沒有覆寫標頭、換用較寬鬆的完成標準或稱為完整解壓MZ匹配。
增量native MZ SHA-256為
`3bb05eecbaff312af6d76ba79abbb019d8cf942e90ef82a133b99ca92a4a51d6`，
與原始packed EXE的hash不同。EXEPACK尚未重建。

乾淨第二工作區從唯讀原版ZIP靜態重生，重新編譯與連結。
C、compiler OBJ、兩個增量OBJ、native MZ、MAP與重定位清冊七項逐byte相同。
獨立核對器另直接解析兩次OMF，套用有明示目標的fixup與backpatch，
重生全部初始化映像及2281筆重定位，與原版再比一次。
修改C指令、移除重定位、改重定位目標及改入口CS四個負例均被拒絕。
正式342份程式輸入保持，沒有執行原版、重製或遊戲對拍。

本機[收據](../verification/local/matching-link-20261008/receipt.json)、
[四函式C](../verification/local/matching-link-20261008/four-functions.c)、
[完整映像與負例](../verification/local/matching-link-20261008/image-review.json)、
[乾淨重編](../verification/local/matching-link-20261008/repeat-review.json)、
[唯讀獨立核對器](../verification/local/matching-link-20261008/verify.py)及
[保存清冊](../verification/local/matching-link-20261008/retention.json)為本輪入口。
下一閘門是原始EXEPACK／MZ重建基準，再沿增量物件擴展C匹配；完整原版C仍未完成。

使用者於2026-10-08要求先commit及push，再繼續。檢查點`ff25f7a`已推送並核對遠端main。
本輪封裝工作區`workplace/matching-pack-20261008/`完成比較後封存、核對並清理，收據與重建入口如下。

## 完整packed EXE的增量重建基準

QuickC2.01的四個C函式、129012 bytes原始未匹配片段、已知段配置與重定位metadata，
經LINK4.07、EXEMOD4.02、EXEPACK4.06與明列外層填補，重生133760-byte KOLTOWN.EXE。
全部bytes及SHA-256符合原版：

`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

這是已證實的完整檔案重建基準。C範圍維持四函式300 bytes，沒有增加library或asm匹配；
129012 bytes仍由原始片段供應，完整C反編譯尚未完成。
後續替換必須同時保存來源覆蓋清冊與完整packed檔案hash。

### 標頭來源勘誤

先前表格的`e_minalloc=2308`來自固定CC0解包器`unexepack.c:384`的
`fullSize / 60`占位計算。該工具同時重建header、清零checksum及重定位表。
其完整MZ是工具衍生物，不能當作原版未封裝標頭的oracle；初始化映像及原始定位證據維持。
先前native MZ與工具輸出不同的數值表保留，用此來源註記修正解讀。

本輪改由原版packed MZ的已知loader契約推導配置：

| 項目 | 已證實的值與來源 |
|---|---|
| 原版實際檔案／MZ宣告長度 | 133760／133679 bytes，原始ZIP entry |
| packed header | 512 bytes，MZ `e_cparhdr=32` |
| 最小載入記憶體 | `ceil((133679-512)/16)+2883=11206` paragraphs，原版MZ `e_minalloc=2883` |
| 未封裝初始化映像 | 129312 bytes，8082 paragraphs |
| 重建前置配置 | `11206-8082=3124` paragraphs，EXEMOD參數以hex寫為`/MIN C34` |

EXEMOD4.02設定build metadata後由EXEPACK封裝，最終packed MZ與原版完全相同。
這是重建loader契約，原始未封裝EXE與歷史建置命令仍未知。
EXEMOD接受十六進位數值，依[Microsoft C使用者手冊第188頁](https://archive.computerhistory.org/resources/access/text/2024/05/102734492-05-0002-acc.pdf#page=104)。
最初誤傳十進位字串`3124`的診斷保留，工具實際寫入`0x3124`；改為同一值的`C34`後核對欄位。

### 重定位與封裝流程

直接LINK的原始片段放在C物件之前時，C遠呼叫的relocation排到末端。
原版metadata另把部分DATA relocation插在CRT CODE之間，因此單純按檔案偏移排序也不足。
每個原始relocation word改由獨立LEDATA／FIXUPP記錄供應，按固定原版序列排列。
普通原始bytes各供應一次，prefix、C、suffix物件交給native LINK後，
完整初始化映像及2281筆實體relocation的順序與目標均符合原版。
物件切分用於重建，不推定原作者的OBJ名稱與邊界。

| 階段 | 實際工具與產物 |
|---|---|
| C與來源分類 | `rebuild.py`由乾淨`PILOT.C`生成compiler OBJ，再由既有OMF適配器生成C與原始物件 |
| 原始序列 | `split-objects.py`生成`PRE.OBJ + CSLOT.OBJ + POST.OBJ`，來源bytes、fixup值與C指令保持 |
| 連結 | QuickC LINK4.07 `/MAP`，原位解析全部引用 |
| loader配置 | 官方EXEMOD4.02 `/MIN C34` |
| 封裝 | 官方EXEPACK4.06，命令依[Microsoft C手冊第147頁](https://archive.computerhistory.org/resources/access/text/2024/05/102734492-05-0002-acc.pdf#page=82) |
| 外層序列化 | `pad-record.py`以128-byte記錄與`0x1A`填補，產生81 bytes尾端 |

原版KOLBOOK與MAKEMODE也分別含55／63 bytes的`0x1A`填補至128-byte邊界。
來源與填補bytes已證實，原作者採用此序列化的原因與工具未知。
LINK4.07及LINK3.65內建`/EXEPACK`的候選均未匹配，保留完整負例；
官方standalone EXEPACK4.06的壓縮資料、18-byte EXEPACK header、261-byte stub及壓縮relocation全部對上。
相容工具組已證實，原版精確compiler／linker／packer版本仍未知。

### 重現與證明範圍

乾淨第二工作區從唯讀ZIP、四函式C、固定工具與明列原始片段完整重建。
11項native、排序物件、MZ、MAP、loader配置、packed core及最終EXE逐byte相同。
最終檔案的header、壓縮資料、EXEPACK header、stub、錯誤字串、重定位與外層填補七區全部核對。
改最小記憶體欄、壓縮byte、重定位byte或尾端填補四個負例均被拒絕。
兩次最終EXE經固定解包器再次輸出相同衍生MZ；正式342份程式輸入保持。
全程只執行compiler／linker／封裝工具，沒有執行原版或重製遊戲，也沒有恢復對拍。

本機[收據](../verification/local/matching-pack-20261008/receipt.json)、
[完整分區差分與重編](../verification/local/matching-pack-20261008/pack-review.json)、
[loader配置契約](../verification/local/matching-pack-20261008/memory-contract.json)、
[完整重建腳本](../verification/local/matching-pack-20261008/rebuild.py)、
[封存核對器](../verification/local/matching-pack-20261008/verify.py)及
[保存清冊](../verification/local/matching-pack-20261008/retention.json)保存證據。
封存核對器解析OMF、檢查原位C與原始覆蓋，並核對原生工具完整重建收據；
完整重編由`rebuild.py`與釘選DOS工具完成，兩者證明範圍分開。
下一閘門是擴大C匹配，每次替換都重生並核對完整原版EXE，129012 bytes仍列未匹配。
目前入口固定四函式與已審查的OMF記錄類型。新增函式須同步擴充來源清冊、外部binding與排序，
再用相同完整packed hash核對；原版code、工具與匹配C仍只留本機。

持續目標為完成《武士傳說》的matching decompilation。C擴展工作區為
`workplace/matching-c-20261008/`，先取代KOLTOWN剩餘函式，再核對原版其他EXE的來源範圍。
完整packed基準沿用；原始未匹配CODE不計為C、library或重寫asm的還原進度。

## 持續C來源還原

目前[來源清冊](../verification/local/matching-c-20261008/source-ledger.json)包含固定ZIP內六支EXE，
以及IDA9.4匯出的全部338個KOLTOWN函式、原始名稱、線性位址、chunks與檔案偏移。
KOLTOWN以外五支EXE的來源狀態仍為`not-restored`。歷史四函式基準保留，不能取代本節現況。

| 目前初始化映像來源 | 數量／bytes | 已證實範圍 |
|---|---|---|
| 純C | 46函式／4309 bytes | compiler原始指令與回填值原位轉換，完整函式與原版零差 |
| 固定library | 8函式／372 bytes | 從固定MLIBC.LIB物件取得，不從原版遊戲貼回；引用與seed初值另核 |
| 原始未匹配CODE | 117895 bytes／284函式仍未還原 | 不計為源碼還原進度 |
| 原始初始化DATA | 6736 bytes | 保留輸入資料來源；不計為C函式進度 |

八個library供應函式為`_strcat`、`_strcpy`、`_strlen`、`_strchr`、`_strrchr`、`_sprintf`、`_srand`與`_rand`，
原始IDA名稱與位址維持。來源`Microsoft QuickC v2/LIB/MLIBC.LIB`釘選commit
`da3f15ce4912893a0dd9f8349dab29116a53ca59`，Git blob
`a8c67bc9a9a8dbb7b7663c16bfb11f2bafa4e45d`，306個物件全部OMF checksum核對。
這證明相容來源，原版精確工具版本與原始符號名稱仍未知。
另一MSC候選的`crt0.asm` COMENT checksum不符，保持診斷，不放寬檢查或把它當匹配來源。

本輪新增的C包含角色條件、骰值欄位、字串與繪圖調度；`sub_21AC4`的720-byte戰鬥記錄初始化、
`sub_21E0B`的450-byte裝備種類掃描與`sub_21FCD`的183-byte switch均完整匹配。
50-byte記錄的欄位依原始offset命名，欄位語意未分級時只作導航，不進正式Go資料格式。
無原版CODE blob、`_asm`或`_emit`塞入C源；沒有修改非fixup指令來湊匹配。

原位工具由四函式專用改為讀取實際compiler PUBLIC／EXTERNAL的批次適配。
每個CODE byte只供應一次；地址欄位由native LINK回填，backpatch另按原始函式位置定位。
一般C與library指令bytes分別與真實compiler／LIB物件核對，獨立reader重生完整初始化映像，
2281筆relocation的順序與目標均符合原版。
目前已驗最大C函式720 bytes；較長函式與其他OMF記錄類型仍須逐批實測，不推定全量已支持。

兩次乾淨C、compiler OBJ、原位OBJ、native MZ、MAP、loader配置、packed core與最終EXE共八項逐byte相同。
原版全檔hash維持`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。
本機[完整函式差分](../verification/local/matching-c-20261008/trial-46c-review.json)、
[獨立來源核對](../verification/local/matching-c-20261008/trial-46c-source-proof.json)、
[乾淨重建](../verification/local/matching-c-20261008/checkpoint-review.json)、
[純C](../verification/local/matching-c-20261008/bulk.c)、
[封存核對器](../verification/local/matching-c-20261008/verify.py)與
[保存清冊](../verification/local/matching-c-20261008/retention.json)為本輪入口。
已完成批次封存、核對後清理工作區；下一批從此源碼與清冊接續，不回退至四函式基準。
正式342份程式輸入保持，遊戲對拍維持暫停。完整matching目標仍進行中，完成條件包含全部來源範圍，
不能只以目前重建EXE的hash相同結案。

完整重編時，先依保存清冊解出輸入並按`fetch-tools.py`重取固定工具。
建立乾淨可寫輸出，複製`rebuild.py`、`dos-run.py`、`split-objects.py`與`pad-record.py`；
原版專案唯讀掛`/src`，解出的`baseline/`與`tools/`分別唯讀掛`/out/baseline`及`/out/tools`。
輸出掛`/out`，UID/GID1000:1000、無網路、1GiB、1CPU、96程序及240秒外層逾時，執行
`python3 -B /out/rebuild.py`。來源與所有掛載目錄先確認存在與擁有權。

## 工具與重生入口

- IDA9.4使用既有`ida-pro-9.4-idapython:py312-v1`，UID1000:1000、正式DB唯讀、一次性副本可寫；
  入口`ida-probe.py`及`ida-export-small.py`保存輸入SHA、函式邊界及原始指令。
- 編譯器來源：[Microsoft公開工具](https://github.com/microsoft/MS-DOS)，commit
  `2d04cacc5322951f187bb17e017c12920ac8ebe2`；[QuickC候選清冊](https://github.com/davidly/dos_compilers)，commit
  `da3f15ce4912893a0dd9f8349dab29116a53ca59`。每檔核Git blob與SHA-256，下載成果留本機。
- DOS工具`kol-matching-dos:bookworm-20261008-r1`沿[Dockerfile](../../docker/Dockerfile.matching-dos)，
  支援無頭DOSBox0.74-3、候選compiler執行及Python核對。只編C，不跑遊戲。
  `compile-linked.py`從乾淨輸出產生OBJ／MZ／MAP，`review-linked.py`按PUBLIC及CODE段長度核完整函式。
- OMF依[TIS1.1規格](https://openwatcom.org/ftp/devel/docs/omf.pdf)接受計算checksum或合法零checksum；
  record長度、PUBLIC、LEDATA及真實link結果另核，沒有把零checksum當損壞或當完整性證明。

獨立核對使用目前Go驗證revision，專案唯讀掛`/src`、本機證據父目錄唯讀掛`/out`，
UID/GID1000:1000、無網路、1GiB、1CPU、96程序與120秒逾時，執行：

```sh
python3 -B /out/matching-pilot-20261008/verify.py
python3 -B /out/matching-link-20261008/verify.py
python3 -B /out/matching-pack-20261008/verify.py
python3 -B /out/matching-c-20261008/verify.py
```

需乾淨重編時按保存清冊抽出源檔、固定候選工具及命令到Docker暫存，來源唯讀、輸出可寫。
可重取的工具及可重建EXE只保留hash，原版EXE由合法本機ZIP重生，不冒稱已匹配整支遊戲。
