# 原版同位元組反編譯

日期：2026-10-10。狀態：六支EXE均有同位元組重建基準；KOL、TANDYPIN、MAKEMODE與TITLE完整來源，KOLBOOK與KOLTOWN部分還原。
目前228個純C函式74428 bytes、378個函式庫函式44396實際bytes、182個混合C／ASM函式75345 bytes；另列271個C與258個ASM CODE bytes及TANDYPIN入口13 bytes。Town全部6736初始化DATA保持；剩餘10函式、4999 CODE bytes，未還原DATA為0，見「持續C來源還原」。
使用者要求暫停目前對拍，改以 matching decompilation 重建原版C程式，
希望降低重製過程中反覆行為比對的成本。本入口保存查證與首個試點範圍；
現有Go程式、規格、原版收據與未完成條件保留。

第三十四批已完成來源驗證並封存，現行入口見「持續C來源還原」。
使用者要求先 commit 與 push，matching 目標暫停。C35 已恢復完整 C34 基準並保留先行試驗；新增字型候選未匹配，來源完成量仍以 C34 為準，接續狀態見 [CONTEXT.md](../../CONTEXT.md)。

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

## 第一批C來源還原紀錄

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

第二批從46C／8library清冊接續；現況以後面的「持續C來源還原」及第二批來源清冊為準。

完整重編時，先依保存清冊解出輸入並按`fetch-tools.py`重取固定工具。
建立乾淨可寫輸出，複製`rebuild.py`、`dos-run.py`、`split-objects.py`與`pad-record.py`；
原版專案唯讀掛`/src`，解出的`baseline/`與`tools/`分別唯讀掛`/out/baseline`及`/out/tools`。
輸出掛`/out`，UID/GID1000:1000、無網路、1GiB、1CPU、96程序及240秒外層逾時，執行
`python3 -B /out/rebuild.py`。來源與所有掛載目錄先確認存在與擁有權。

## 第二批C與函式庫還原紀錄

2026-10-08第二批已完成來源與完整檔案核對；[來源清冊](../verification/local/matching-c2-20261008/source-ledger.json)
保存全部六支EXE。KOLTOWN與KOLBOOK為部分來源還原，其餘四支仍為`not-restored`。
完整EXE相同的證明包含原始未匹配片段，不能當作完整源碼還原。

| 程式 | 純C | 固定函式庫 | 原始未匹配CODE |
|---|---|---|---|
| KOLTOWN | 51函式／5057 bytes | 42函式／4214 bytes | 113305 bytes／245函式未還原 |
| KOLBOOK | 1函式／89 bytes | 本批未替換 | 37351 bytes／190函式未還原 |

合計52個C函式5146 bytes。KOLTOWN的6736-byte初始化DATA保持；原始CODE與DATA的來源明列，
不計為C或函式庫還原進度。原始名稱、IDA9.4線性位址、函式邊界與檔案偏移均保留。

本批新增KOLTOWN的`sub_22084`、`sub_2215A`、`sub_236AB`、`sub_14CCF`與`sub_14D3A`。
後兩者以C的32-bit運算重建三位元組讀寫，實際compiler helper由固定LIBH供應。
函式完整指令原位相同為已證實；C宣告與型別只證明能產生相同指令，不證明原作者文字唯一。
本批沒有改動正式Go來源或恢復遊戲對拍。

函式庫來源為同一固定commit的`MLIBC.LIB`、`LIBH.LIB`與`MLIBFP.LIB`。
候選比對只遮蔽地址欄位，計入進度前仍逐項解出外部符號、段基址與回填值，再經native LINK核對完整函式。
同一外部符號與同一物件段必須共享相容定位；不能各函式任意指定不同地址。
LIBH的部分LEDATA會以來源物件中後寫入的`retf 8`覆蓋較早的`retf`；這是來源記錄的順序，
沒有修改原始普通指令來湊匹配。三個LIB的Git blob、SHA-256與物件hash見
[來源輸入](../verification/local/matching-c2-20261008/library-inputs.json)及
[物件清冊](../verification/local/matching-c2-20261008/runtime-library-inventory.json)。
這組工具與來源相容為已證實，原版精確工具版本及原始符號名稱仍未知。

KOLBOOK使用`/AS`的near C函式`sub_1276F`，IDA線性位址`0x1276F`至`0x127C8`。
原版資料區的兩個重定位site以比資料段起點晚64 bytes的段框定位。
保留此段框後，native LINK產生的header、checksum、初始化映像與四筆有序重定位均相同，
再依原始55-byte `0x1A`尾端填補，完整40192 bytes的SHA-256為
`2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832`。
未分段的候選雖有相同物理site，重定位表及checksum不同，保留為負例；沒有手改checksum。

KOLTOWN的2281筆有序重定位與完整初始化映像相同；完整133760-byte packed檔案SHA-256仍為
`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。
兩支程式的乾淨C、compiler OBJ、原位OBJ、native MZ、MAP及KOLTOWN封裝共十三項逐byte相同。
六項反例涵蓋C、函式庫、最小記憶體、尾端填補、KOLBOOK checksum及C來源。

本機[收據](../verification/local/matching-c2-20261008/receipt.json)、
[乾淨重建](../verification/local/matching-c2-20261008/checkpoint-review.json)、
[KOLTOWN獨立來源核對](../verification/local/matching-c2-20261008/trial-51-source-proof.json)、
[KOLBOOK獨立來源核對](../verification/local/matching-c2-20261008/book-source-proof.json)、
[純C](../verification/local/matching-c2-20261008/bulk.c)、
[KOLBOOK純C](../verification/local/matching-c2-20261008/book.c)、
[封存核對器](../verification/local/matching-c2-20261008/verify.py)及
[保存清冊](../verification/local/matching-c2-20261008/retention.json)為本批入口。
核對器從固定ZIP重生衍生MZ，獨立解析compiler／LIB／原位OBJ，重算完整映像及重定位；
完整DOS重編的成果由收據與十三項雜湊另證，不宣稱封存核對器重新執行了DOS compiler。
全部匹配C、LIB、OBJ、DB及原始片段只留本機；完整matching目標保持進行中。

第三批從第二批壓縮證據接續，現況以後面的「持續C來源還原」及第三批清冊為準。

## 第三批KOLBOOK來源還原紀錄

2026-10-08第三批擴大KOLBOOK，原始191個函式均已列入
[目前來源清冊](../verification/local/matching-c3-20261008/source-ledger.json)。
KOLTOWN來源與完整封裝基準保持；本批新增KOLBOOK的八個純C函式505 bytes及36個函式庫函式3705 bytes。

| 程式 | 純C | 固定函式庫 | 原始未匹配CODE |
|---|---|---|---|
| KOLTOWN | 51函式／5057 bytes | 42函式／4214 bytes | 113305 bytes／245函式未還原 |
| KOLBOOK | 9函式／594 bytes | 36函式／3705 bytes | 33141 bytes／146函式未還原 |

合計60C／5651 bytes與78個函式庫函式／7919 bytes。其餘四支EXE仍為`not-restored`。
原始未匹配CODE與DATA持續明列來源，不計為還原進度。

新增C為`sub_14BE4`、`sub_14CE2`、`sub_151D5`、`sub_14FE7`、`sub_15702`、`sub_15481`、
`sub_15880`及`sub_14094`，原始名稱、IDA9.4線性位址與邊界保留。
完整函式機器碼相同為已證實；C變數型別、寫法及語意別名仍不代表唯一原作源碼。
`sub_13E25`的far pointer候選產生24 bytes，原版為26 bytes且指令不同，保留負例並列未還原；
沒有用ASM或原始指令補入C。等待迴圈只驗靜態編譯結果，不宣稱硬體時序或實機wall-clock一致。

KOLBOOK的near-call使用self-relative FIXUPP，原位工具保留compiler指令及addend，
交由native LINK計算相對目標。六項反例包含有效OMF checksum下的C指令、函式庫指令及near-call目標修改，
另拒絕ASM注入、錯誤MZ checksum與錯誤尾端填補。

搜尋固定`SLIBC.LIB`、`SLIBFP.LIB`與既有`LIBH.LIB`，沿同一commit及Git blob。
本批36個實際供應函式均來自SLIBC；先通過普通指令及共享binding，再實際native LINK核完整body，計入3705 bytes。
尚未支援的LIDATA物件明確拒絕來源計分；不能把可比對的部分當成整個函式庫支援完成。
來源、工具與物件記錄見[near函式庫輸入](../verification/local/matching-c3-20261008/book-library-inputs.json)、
[函式庫清冊](../verification/local/matching-c3-20261008/runtime-library-inventory.json)及
[完整原位差分](../verification/local/matching-c3-20261008/book-9rt36-review.json)。
相容來源已證實，精確歷史compiler／linker版本仍未知。

兩次乾淨BOOK.C、compiler OBJ、原位OBJ、native MZ、MAP及完整EXE共六項逐byte相同。
原始header、checksum、完整初始化映像、四筆有序重定位及55-byte填補均符合原版，完整hash保持
`2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832`。
KOLTOWN既有完整packed hash保持；封存核對器另重算兩支程式各兩份compiler／LIB／OMF來源。
正式342份程式輸入保持，全程只執行建置工具，遊戲對拍維持暫停。

本機[收據](../verification/local/matching-c3-20261008/receipt.json)、
[乾淨重編與反例](../verification/local/matching-c3-20261008/checkpoint-review.json)、
[獨立來源核對](../verification/local/matching-c3-20261008/book-9rt36-source-proof.json)、
[純C](../verification/local/matching-c3-20261008/book.c)、
[原位建置工具](../verification/local/matching-c3-20261008/build-book.py)、
[封存核對器](../verification/local/matching-c3-20261008/verify.py)及
[保存清冊](../verification/local/matching-c3-20261008/retention.json)為本批入口。
完成比較後封存並清理`workplace/matching-c3-20261008/`；下一批從本清冊接續，完整matching目標仍未完成。

第四批接續第三批清冊；現況以後面的「持續C來源還原」及第四批清冊為準。

## 第四批六支程式來源基準紀錄

2026-10-08第四批建立六支EXE的完整同位元組重建基準。
[來源清冊](../verification/local/matching-c4-20261008/source-ledger.json)保存全部799個IDA9.4原始函式及各自位址基準。
除TANDYPIN以外，五支程式仍使用原始未匹配CODE與DATA，完整matching目標尚未完成。

| 程式 | 純C | 固定函式庫 | 未還原CODE／函式 |
|---|---|---|---|
| KOLTOWN | 51函式／5057 bytes | 42函式／4214 bytes | 113305 bytes／245函式 |
| KOLBOOK | 9函式／594 bytes | 36函式／3705 bytes | 33141 bytes／146函式 |
| KOL | 本批未替換 | 35函式／3917 bytes | 6307 bytes／41函式 |
| TITLE | 本批未替換 | 36函式／3704 bytes | 13176 bytes／79函式 |
| MAKEMODE | 本批未替換 | 30函式／2459 bytes | 10357 bytes／48函式 |
| TANDYPIN | 不計為C | 組合語言入口13 bytes，連結填補3 bytes | 原始CODE片段0／未還原函式0 |

合計60C／5651 bytes、179個函式庫函式／17999 bytes及一個組合語言入口13 bytes。
尚有559個函式、176286 CODE bytes與13788初始化DATA bytes未還原，均不計為來源進度。
清冊的`source_root`明列前批來源所在目錄，歷史清冊、原始名稱、chunks及檔案偏移保留。

### TANDYPIN完整來源

IDA9.4原始`start`在載入線性位址`0x10000..0x1000D`，原始MZ檔案偏移`0x200..0x20D`。
五條指令由MASM5.10組譯，段selector由LINK4.07回填；後面三個零byte由段對齊產生。
96-byte資料由`16 dup ('stack ')`生成，沒有貼回原始CODE或DATA映像。
原始MZ的一筆重定位、header、checksum及全部624 bytes均相同，SHA-256保持
`036c39f330a4fe09bc2729756816b869f1154515bf4befba190e767d4e3cdce3`。

獨立reader解析真正MASM OBJ的CODE、巢狀LIDATA及segment fixup，重算完整112-byte初始化映像；
兩次乾淨ASM、OBJ、完整EXE及MAP共四項相同。
這支程式的完整來源重建已證實，原作者的唯一文字、原始段名及歷史工具版本仍未知。
組合語言`TANDY.ASM`保存於[完整來源封存](../verification/local/matching-c4-20261008/evidence.tar.xz)，
證據為[來源核對](../verification/local/matching-c4-20261008/tandy-repeat-source-proof.json)。
不把這13 bytes計為C，也不以它取代其餘遊戲源碼的還原。

### 三支程式的函式庫來源

KOL、TITLE、MAKEMODE分別匯出76、115、78個原始函式。
新增101個實際SLIBC供應函式10080 bytes，全部普通指令保持來源LIB，回填值另依共享binding核對。
native LINK的三支完整EXE、header、checksum、初始化映像與6／4／4筆有序重定位均相同；
MAKEMODE另外保留原始63-byte `0x1A`填補，KOL與TITLE沒有此填補。
各程式重新從真實LIB重生原位OBJ、MZ、MAP及完整EXE，共十二項相同。

### 特殊符號binding勘誤

`__aDBdoswp`的所有來源引用一致要求16-bit值`0xD6D6`。
先前candidate resolver把無已知PUBLIC類別的引用預設成DATA，所算位置在KOL與TITLE超出靜態段範圍；
這個預設不能證明原始DATA地址。先前普通指令及全檔相同的證據仍成立，該語意推論撤回。
目前在相容建置中明列ABS binding，保留原始符號、fixup、先前定位與勘誤索引。
完整body、native MAP的`0000:D6D6 Abs`與原版全檔相同為已證實；原版符號定義與用途仍未知。
不把相容ABS配置提升成原版歷史型別。

不供應此符號的native LINK反例會報L2029，仍可能留下EXE；驗證器明確拒絕有連結錯誤的產物。
另有初次EXTDEF索引遺漏的失敗，屬匯出工具問題，完整來源索引修正後另建乾淨候選。
沒有改普通指令、任意放寬段界線或刪除未解析引用。

本批十六項乾淨產物及九項反例均通過，包含三支LIB指令、ABS值、TANDYPIN指令及LIDATA修改，
另拒絕錯誤checksum、尾端填補與「連結出錯但EXE存在」。
完全唯讀封存核對器重算新舊共十二份來源證明，正式342份程式輸入保持；沒有遊戲執行或恢復對拍。

本機[收據](../verification/local/matching-c4-20261008/receipt.json)、
[乾淨重生與反例](../verification/local/matching-c4-20261008/checkpoint-review.json)、
[來源核對器](../verification/local/matching-c4-20261008/proof-program.py)、
[TANDYPIN核對器](../verification/local/matching-c4-20261008/proof-tandy.py)、
[封存核對器](../verification/local/matching-c4-20261008/verify.py)及
[保存清冊](../verification/local/matching-c4-20261008/retention.json)為本批入口。
已完成比較的`workplace/matching-c4-20261008/`封存後清理；下一批沿清冊還原五支剩餘來源。

第五批接續第四批清冊；現況以後面的「持續C來源還原」及第五批清冊為準。

## 第五批KOL完整來源還原紀錄

2026-10-08第五批已完成KOL.EXE的完整來源還原。
[來源清冊](../verification/local/matching-c5-20261008/source-ledger.json)保持全部六支、799個原始函式的範圍。
KOL與TANDYPIN為完整來源；KOLTOWN、KOLBOOK、TITLE及MAKEMODE仍部分來源，完整目標尚未完成。

KOL的原始`_main`為IDA線性`0x10010..0x10247`、MZ檔案偏移`0x210..0x447`。
控制流程以C重寫，直接BIOS／I/O、ROM掃描及畫面初始化以五個明列ASM區塊保留。
完整567 bytes經QuickC2.01與LINK4.07原位匹配，區分326個C產生bytes與241個硬體ASM bytes，
整個函式列為混合來源，不加到純C函式數。
`for (;;)`候選少一個三byte入口跳躍，實際compiler比較後改成原版的`while (1)`形狀；
沒有用ASM填補普通C控制流差異，也沒有使用`_emit`或原版CODE blob。

| KOL初始化映像來源 | 已驗範圍 |
|---|---|
| 混合C／硬體ASM `_main` | 567 bytes，一個完整函式 |
| 固定LIB CODE | 48個完整模組、75個IDA函式及全部原始chunks；9600個實際來源bytes，另13個宣告零byte |
| 其他CODE零填補 | 44 bytes，由宣告段與native LINK供應 |
| 初始化DATA | 全部1235 bytes；122個實際DATA.C compiler bytes，其餘固定LIB initializer與四個對齊零byte |
| 原版CODE／DATA片段 | 0 bytes |

解析器補LOCAL PUBLIC／EXTERNAL、COMDEF及LIDATA，所有external索引保持實際記錄順序。
LIDATA的fixup位置依編碼中的literal offset映射到每次展開後的位置；`cmiscdat.asm`的五個指標
由同一個來源word展開，不能把編碼偏移當展開後記憶體偏移。
局部名稱按模組隔離，完整模組供應共用tail及非函式內的資料，不改IDA原始chunk歸屬。
首次小函式候選的`_ldexp`與`B$FIST`在完整binding後不符，保留拒絕，沒有只以遮罩比對計入進度。

主程式DATA由可讀C字串、九個硬體register初值、旗標初值及兩個原始word產生。
`DS:+0x2F8`的`0x07E4／0x2464`初值已證實，語意仍未知；此限制不改寫為推測名稱。
固定LIB提供stdio、runtime訊息、copyright與其他initializer，普通DATA bytes另與實際物件核對。
第一個完整DATA候選把兩筆far pointer正規化成DGROUP表示，物理位置相同但四個bytes與checksum不同；
改為保留原始`0x0283`段框後，native LINK直接產生相同指標、checksum與六筆有序重定位。
沒有手改header、checksum或pointer bytes。

兩次乾淨MAIN.C／OBJ、DATA.C／OBJ、原位OBJ、native MZ、MAP及完整EXE共八項逐byte相同。
完整11971 bytes的SHA-256保持`4f798b078db889b6166b92c0f0bae1febe3efb3535ad67800b2e0aaed94d7b71`。
獨立reader重算兩份compiler、真實LIB CODE與initializer、byte／word BAKPAT、far pointer及段零填補，
完整11459-byte初始化映像相同，76個原始函式的全部chunks有來源，沒有重複LEDATA供應。
八項反例涵蓋C、硬體ASM、LIB、C資料、LIDATA、byte BAKPAT、header checksum與far pointer段框。

累計60個純C函式5651 bytes、219個函式庫函式23682 bytes、一個混合主程式567 bytes及TANDYPIN入口13 bytes。
KOL新增40個完整LIB函式，原始CODE／DATA未還原範圍均降為0；其餘四支尚有518函式、169979 CODE bytes及12553 DATA bytes。
原作者的唯一文字、部分初值語意及精確歷史工具版本仍未知；正式342份Go等程式輸入保持，對拍維持暫停。

本機[收據](../verification/local/matching-c5-20261008/receipt.json)、
[完整差分及反例](../verification/local/matching-c5-20261008/checkpoint-review.json)、
[MAIN.C與DATA.C完整來源封存](../verification/local/matching-c5-20261008/evidence.tar.xz)、
[獨立完整來源核對](../verification/local/matching-c5-20261008/proof-full-kol.py)、
[封存核對器](../verification/local/matching-c5-20261008/verify.py)及
[保存清冊](../verification/local/matching-c5-20261008/retention.json)為本批入口。
已完成比較的`workplace/matching-c5-20261008/`封存並清理，後續從本清冊接續其他四支來源。

第六批接續第五批清冊，現況以後面的「持續C來源還原」及第六批清冊為準。

## 第六批來源還原紀錄

2026-10-08第六批把已驗整模組方法擴展到TITLE、MAKEMODE及KOLBOOK。
[唯一來源清冊](../verification/local/matching-c6-20261008/source-ledger.json)保存全部799個原始函式。
KOL與TANDYPIN完整來源的狀態保持，其餘四支仍未完成。

| 程式 | 純C | 固定LIB來源 | 未還原函式／CODE |
|---|---|---|---|
| KOLTOWN | 51／5057 bytes | 42函式／4214 bytes | 245／113305 bytes |
| KOLBOOK | 9／594 bytes | 82函式／9400實際來源bytes，另13個宣告零byte | 100／27433 bytes |
| TITLE | 本批未替換 | 82函式／9459實際來源bytes，另13個宣告零byte | 33／7408 bytes |
| MAKEMODE | 4／1328 bytes，另18個compiler epilogue bytes | 64函式／6305 bytes | 10／5165 bytes |

整模組來源分別為TITLE 52模組、MAKEMODE 36模組、KOLBOOK 52模組；原始函式與模組數分列。
新的126個完整LIB函式包含所有IDA原始chunks及共享tail，不以primary span當完整函式。
固定LIB的near long-multiply後寫入記錄確實以`C2 0800`覆蓋`C3`，只加入此已觀察規則；
其餘不支援的overwrite仍拒絕，沒有從原版指令補LIB。
來源匯出子目錄的module搜尋路徑已明列；舊reader未涵蓋宣告零byte的失敗屬驗證工具問題，
按KOL已驗同一契約更新後，在相同image核對完整source及native結果。

MAKEMODE新增`_main`、`sub_11199`、`sub_1142C`及`sub_11497`四個純C函式。
原始IDA primary body分別為968、160、107、93 bytes，全部完整指令相同。
前兩個函式的IDA片段在不返回呼叫後停止，後面的兩段九byte compiler epilogue仍由source供應，
另列18 bytes，不修改原IDA邊界或把它算成新函式。

已定義函式間的QuickC call會預先編成同一CODE的相對位移，沒有可搬移的FIXUPP。
因此把函式分成獨立translation unit，真實compiler產生外部call FIXUPP，再交LINK回填。
沒有改普通call bytes來湊匹配。signed／unsigned prototype依實際MOV／PUSH差異核對，
隔離物件的PUBLIC、原始位置及每筆backpatch均由匯出器讀取。
兩個不符原版的`sub_103E1`算式候選保留為負例，不計還原；擷取器誤取prototype後的main也已修正，
正確exit來源另建物件，不沿用錯誤物件。

三支完整EXE、header、checksum、初始化映像與四筆有序重定位均符合原版，
KOLBOOK 55-byte及MAKEMODE 63-byte原始填補保持。所有新 ordinary CODE bytes與真實compiler／LIB核對。
22項乾淨源檔／物件／MZ／MAP／完整EXE逐byte相同，八項反例拒絕LIB、C與header修改；
六份獨立reader證明重算完整初始化映像，不只保存預期布林值。

累計64純C函式6979 bytes、345個LIB函式38978個實際來源bytes、KOL混合主程式567 bytes及TANDYPIN入口13 bytes。
宣告零byte與epilogue分列，剩餘388個函式、153311 CODE bytes及12553 DATA bytes未還原。
三支新全檔相同仍是混合來源基準，不宣稱這三支完整反編譯。
正式342份程式輸入保持，全程沒有執行遊戲或恢復對拍。

本機[收據](../verification/local/matching-c6-20261008/receipt.json)、
[三支來源及乾淨重建](../verification/local/matching-c6-20261008/checkpoint-review.json)、
[完整source核對器](../verification/local/matching-c6-20261008/proof-program.py)、
[純C與LIB封存](../verification/local/matching-c6-20261008/evidence.tar.xz)、
[封存核對器](../verification/local/matching-c6-20261008/verify.py)及
[保存清冊](../verification/local/matching-c6-20261008/retention.json)為本批入口。
完成比較的`workplace/matching-c6-20261008/`封存後清理，下一批沿此來源與原始位址接續。

## 第七批來源還原紀錄

2026-10-09第七批新增MAKEMODE九個完整函式，共1757 bytes。
現行[唯一來源清冊](../verification/local/matching-c7-20261009/source-ledger.json)
保持全部799個原始函式；KOL與TANDYPIN完整，其餘四支仍部分來源。

| 原始IDA名稱 | IDA linear | 原始EXE檔案偏移 | 完整函式bytes | 來源分類與分量 |
|---|---|---|---|---|
| `sub_103E1` | `0x103E1` | 1505 | 131 | 純C |
| `sub_11242` | `0x11242` | 5186 | 214 | C195／ASM19 |
| `sub_11318` | `0x11318` | 5400 | 67 | C15／ASM52 |
| `sub_1135B` | `0x1135B` | 5467 | 100 | C15／ASM85 |
| `sub_113BF` | `0x113BF` | 5567 | 109 | C74／ASM35 |
| `sub_114F4` | `0x114F4` | 5876 | 496 | C482／ASM14 |
| `sub_116E4` | `0x116E4` | 6372 | 80 | C15／ASM65 |
| `sub_11734` | `0x11734` | 6452 | 372 | C99／ASM273 |
| `sub_118A8` | `0x118A8` | 6824 | 188 | C36／ASM152 |

輸入為固定MAKEMODE.EXE，SHA-256
`3051cc981e2cbf62237c1cbb788f4d10b84c8dea4dcd6c2aad67171cd2196c56`。
原始定位來自IDA Pro 9.4匯出的函式、chunks與原始file bytes；表中兩種位址基準分列。
QuickC2.01沿`/AS /Od /Gs /Zl /G0`，LINK4.07原生回填全部符號與重定位。
上述全部函式的完整body與原始chunks逐byte相同，證據等級為已證實的匹配表達。

`sub_103E1`的舊整數算式候選撤回。原版DGROUP `+0x42`是六筆十byte磁片名，
`+0x7E`是兩筆十byte顯示模式名；`+0x7A0`的printf格式有兩個`%s`。
因此用兩張字串表加十byte步距的near pointer表達，compiler產生原始MOV CX、ADD及PUSH。
整數常數候選各少兩個bytes，兩個失敗候選及原DATA／consumer定位保留在封存中。
這項勘誤只修正匹配C來源，不修改Go資料或正式遊戲行為。

八個混合函式分列931個C產生bytes與695個明列ASM bytes。
ASM採暫存器、標籤與符號指令重寫，沒有`_emit`或原始指令blob。
只在已列區塊使用ASM，不以ASM補普通C差異；原作者的inline ASM形式仍未知。
QuickC參數名稱與ASM保留字衝突已改正，byte BAKPAT沿第五批已驗契約處理。
核對器改為檢查實際token，避免把註解中的`_emit`說明誤判成原始bytes指令。
上述失敗屬候選來源或驗證工具問題，同image修正重跑後通過。

十三份C來源於兩個空白目錄完整重編。32項源檔、OBJ、輸入清冊、原位OBJ、MAP、
native MZ及完整EXE逐byte相同；完整15744-byte MAKEMODE.EXE與原版相同。
獨立reader從真實compiler OBJ與固定LIB重算兩份完整初始化映像及四筆有序重定位，
並核對所有compiler fixup與byte／word backpatch，不只比對預期布林值。
十項反例拒絕實際改C再編譯、實際改ASM再編譯、C／ASM／LIB指令、byte回填、
near-call目標、ASM誤列純C及header checksum修改；修改OMF的六項反例保持合法checksum。

累計65純C函式7110 bytes、345個LIB函式38978實際來源bytes、
九個混合C／ASM函式2193 bytes與TANDYPIN入口13 bytes。
混合來源共1257個C產生bytes與936個ASM bytes，18個compiler epilogue bytes另列。
仍有379函式、151554 CODE bytes及12553初始化DATA bytes未還原。
MAKEMODE剩`sub_10464`的3381 bytes、函式外27 CODE bytes及2353 DATA bytes，
完整EXE相同仍屬混合基準，不宣稱MAKEMODE完整來源或六支反編譯完成。
正式342份程式輸入保持，遊戲對拍維持暫停。

本機[收據](../verification/local/matching-c7-20261009/receipt.json)、
[乾淨重編與反例](../verification/local/matching-c7-20261009/checkpoint-review.json)、
[完整source核對器](../verification/local/matching-c7-20261009/proof-program.py)、
[來源封存](../verification/local/matching-c7-20261009/evidence.tar.xz)、
[唯讀核對器](../verification/local/matching-c7-20261009/verify.py)及
[保存清冊](../verification/local/matching-c7-20261009/retention.json)為本批入口。
完成比較的`workplace/matching-c7-20261009/`核對封存後清理，下一批沿本清冊接續。

## 第八批來源還原紀錄

2026-10-09第八批完成MAKEMODE全部78個原始函式、12816 CODE bytes與2353初始化DATA bytes的來源還原。
完整15744-byte EXE由C、明列ASM、固定LIB與宣告填補供應，原始CODE／DATA片段均為0。
[唯一來源清冊](../verification/local/matching-c8-20261009/source-ledger.json)保持六支、799個原始函式的完整範圍。
KOL、TANDYPIN與MAKEMODE完整來源；TITLE、KOLBOOK與KOLTOWN仍部分來源，完整目標未完成。

| MAKEMODE範圍 | bytes | 已核對來源 |
|---|---|---|
| 五個純C函式 | 1459 | QuickC2.01原生編譯 |
| 九個混合C／ASM函式 | 5007 | 4224個C產生bytes與783個明列ASM bytes |
| 不返回呼叫後epilogue | 18 | 真實compiler OBJ，原始函式邊界保持 |
| 64個原始LIB函式 | 6305 | 36個固定LIB完整CODE模組，包含所有chunks與shared tail |
| CODE填補 | 27 | 宣告段長度的原生零填補，沒有LEDATA原始片段 |
| C初始化DATA | 1495 | 67個符號，固定寬度名稱表、byte lookup、3／6／2-byte記錄表、字串及明列初值 |
| LIB初始化DATA | 854 | 766個真實初值bytes與88個LIB宣告零bytes，91個DATA來源範圍中的24個 |
| DATA填補 | 4 | 原生零填補 |
| 外層填補 | 63 | 宣告的`0x1A`，與原版完整檔案再比一次 |

本批新增`sub_10464`，IDA Pro 9.4 linear `0x10464..0x11199`、
原始MZ檔案偏移1636、3381 bytes，分列3293個C產生bytes與88個ASM bytes。
23個switch分支、結構步距、文件讀寫、尺寸與溢位條件均由C表達。
五個直接搬移或交換區塊明列ASM；沒有以ASM補普通C差異。
原位完整body與原始chunks逐byte相同，證據等級為已證實的匹配表達，原作者的唯一寫法仍未知。

首個header高／低byte算式的運算順序與原指令不同，調整C表達後匹配。
兩個word採有號C表達才產生原始轉換與暫存器順序；`switch`未列default才符合原始分派。
最後四個PUSH bytes由呼叫端未定型prototype保留原ABI而匹配。
上述候選與失敗差分均保存，不遮罩普通指令，也不宣稱追回原作者的實際型別宣告。

DATA.C的最後字串為奇數長度，compiler實際產生1495 bytes，最後一個對齊byte由LINK供應。
初始化表共同邊界DS`+0x4E2`的兩個bytes為`00 1F`。
固定LIB資料段未找到這兩個初值；`__cinit`於IDA`0x11ADB`、`0x11ADE`、`0x11AE4`、`0x11AE7`
分別令SI、DI同為`0x4E2`，空表範圍不讀取這個值。
因此明列`0x1F00`初值以重生觀察bytes，值為已證實，用途未知，原始C宣告為假說。
這項未知語意保留，不因完整EXE相同就稱為已解出。

實際Maker DATA沒有far-pointer型fixup，反例腳本的舊段框假設已由原始OMF核對撤回。
改測實際segment fixup的外部目標，保留全部source記錄、有序重定位與原生LINK結果。
沿固定QuickC2.01、`/AS /Od /Gs /Zl /G0`及LINK4.07，十五份來源於兩個空白目錄重編。
36項C源檔、OBJ、輸入清冊、原位OBJ、MAP、native MZ與完整EXE逐byte相同。
完整EXE兩次均與原版SHA-256
`3051cc981e2cbf62237c1cbb788f4d10b84c8dea4dcd6c2aad67171cd2196c56`相同。

兩份獨立reader從真實compiler CODE／DATA OBJ與固定LIB重算完整15169-byte初始化映像，
逐byte確認唯一來源、全部compiler fixup／byte與word backpatch、所有DATA source fixup與四筆有序重定位。
十三項反例拒絕實際C／ASM／未知初值再編譯、C與ASM指令、CODE／DATA函式庫、
compiler DATA、byte回填、segment目標、額外零區LEDATA、來源清冊漏項及header checksum。
八項OMF修改仍有合法checksum，來源漏項即使保持原始全檔也不通過。

累計65純C函式7110 bytes、345個LIB函式38978實際來源bytes、十個混合C／ASM函式5574 bytes，
混合來源共4550個C產生bytes與1024個ASM bytes，TANDYPIN入口13 bytes另列。
剩餘378個函式、148146 CODE bytes與10200初始化DATA bytes；完整matching目標維持active。
正式342份Go及模組輸入雜湊保持，遊戲對拍仍暫停，未把原版來源匹配延伸成Go parity。

本機[收據](../verification/local/matching-c8-20261009/receipt.json)、
[來源清冊](../verification/local/matching-c8-20261009/source-ledger.json)、
[乾淨重編與反例](../verification/local/matching-c8-20261009/checkpoint-review.json)、
[完整核對器](../verification/local/matching-c8-20261009/proof-program.py)、
[DATA來源產生器](../verification/local/matching-c8-20261009/materialize-maker-data.py)、
[來源封存](../verification/local/matching-c8-20261009/evidence.tar.xz)、
[唯讀核對器](../verification/local/matching-c8-20261009/verify.py)及
[保存清冊](../verification/local/matching-c8-20261009/retention.json)為本批入口。
完成比較的`workplace/matching-c8-20261009/`核對封存後清理，下一批接續TITLE的剩餘來源。

## 第九批來源還原紀錄

2026-10-09第九批完成TITLE全部115個原始函式、16880 CODE bytes與1279初始化DATA bytes的來源還原。
完整18671-byte EXE由C、明列ASM、固定LIB與原生填補供應，原始CODE／DATA片段均為0。
[唯一來源清冊](../verification/local/matching-c9-20261009/source-ledger.json)保持全部六支、799個原始函式的範圍。
四支完整來源，KOLBOOK與KOLTOWN仍部分來源，完整目標未完成。

| TITLE範圍 | bytes | 已核對來源 |
|---|---|---|
| 十個純C原始函式 | 1625 | QuickC2.01原生編譯 |
| 23個混合C／ASM原始函式 | 5629 | 4269個C產生bytes與1360個明列ASM bytes |
| 未列為原始函式的CODE | 118 | 原始IDA linear `0x110BD..0x11133`另列C來源，不增加函式數 |
| 82個原始LIB函式 | 9472範圍 | 52個完整CODE模組，9459個實際來源bytes與13個宣告零bytes |
| CODE填補 | 36 | 原生零填補，沒有原始LEDATA片段 |
| C初始化DATA | 218 | 二十個符號，模式名、Hercules word記錄、字串及明列初值 |
| LIB初始化DATA | 1057 | 969個真實初值bytes與88個宣告零bytes |
| DATA填補 | 4 | 原生零填補 |

本批新增33個原始函式，包含完整2784-byte `_main`。
普通C控制流程、解碼、讀檔、尺寸運算與亂數呼叫保留，硬體／暫存器區塊另列ASM。
原始IDA名稱、位址、chunks、file bytes與caller保持；所有完整body逐byte相同，
證據等級為已證實的匹配表達，原作者實際型別、唯一寫法與精確歷史工具版本仍未知。
少量差分由漏列或多列呼叫、運算次序、局部變數配置及prototype轉換修正，沒有以ASM補普通C差異。
不為播放時序或busy-wait做逐週期硬體研究，沒有新增原版／remake遊戲執行或RND結果對拍。

IDA Pro 9.4從正式DB的一次性副本核對`0x110BD..0x11133`的49項原始CODE指令與xref。
這段118 bytes沒有原始函式歸屬；正式DB雜湊與115個原始函式不變，沒有建立或改名函式。
另列C範圍後，真實QuickC完整產生相同118 bytes。別名只作導覽，不當成原始名稱或語意證據。
普通函式與額外CODE分開計數，沒有用庫模組、共享tail或未命名範圍膨脹799個原始函式。

DATA包含五byte模式名稱記錄、九筆Hercules word記錄、十二個字串及初值。
DS`+0x2D6`的`0x227A`與`0x3E9E`值已證實，用途未知，原始C宣告為假說，限制隨來源保持。
compiler與LIB initializer實際供應全部1279 bytes，不用原版DATA片段或未宣告fixture。

沿固定QuickC2.01、`/AS /Od /Gs /Zl /G0`及LINK4.07，35份來源加共同header於兩個空白目錄重編。
77項源檔、OBJ、header、清冊、原位OBJ、MAP、native MZ與完整EXE逐byte相同，
兩次均與原版SHA-256 `8a2879d924d9a9cb01e000bd548c348022c0f922288565015f7560ab8fc6fad6`相同。
兩份獨立reader重算18159-byte初始化映像、全部compiler／LIB／DATA來源、fixup、backpatch與四筆有序重定位。
十五項反例拒絕實際C／ASM／未知初值／額外C再編譯、CODE／DATA指令、byte回填、segment目標、
額外零區LEDATA、DATA來源及額外C來源漏項、header checksum；八項OMF修改仍有合法checksum。

累計75純C函式8735 bytes、345個LIB函式38978實際來源bytes、33個混合C／ASM函式11203 bytes，
混合來源含8819個C產生bytes與2384個ASM bytes；額外C CODE118 bytes與TANDYPIN入口13 bytes分列。
剩餘345個函式、140738 CODE bytes與8921初始化DATA bytes，完整matching目標維持active。
正式342份Go及模組輸入雜湊保持，未把原版來源匹配延伸成Go parity。

本機[收據](../verification/local/matching-c9-20261009/receipt.json)、
[來源清冊](../verification/local/matching-c9-20261009/source-ledger.json)、
[乾淨重編與反例](../verification/local/matching-c9-20261009/checkpoint-review.json)、
[完整核對器](../verification/local/matching-c9-20261009/proof-program.py)、
[來源封存](../verification/local/matching-c9-20261009/evidence.tar.xz)、
[唯讀核對器](../verification/local/matching-c9-20261009/verify.py)及
[保存清冊](../verification/local/matching-c9-20261009/retention.json)為本批入口。
完成比較的`workplace/matching-c9-20261009/`核對封存後清理，下一批接續KOLBOOK剩餘來源。

## 第十批來源還原紀錄

2026-10-09第十批KOLBOOK新增75個完整原始函式，共8485 bytes，
分列二十個純C2475 bytes與55個混合C／ASM6010 bytes。
原有九個純C594 bytes與82個完整LIB原始函式保持，沒有退回原版片段。
[唯一來源清冊](../verification/local/matching-c10-20261009/source-ledger.json)仍保存全部六支、799個原始函式。
KOLBOOK仍有25個原始函式、18948 CODE bytes及2185初始化DATA bytes未還原，未宣稱整支來源完成。

| KOLBOOK已核對範圍 | 函式／bytes | 來源 |
|---|---|---|
| 純C | 29／3069 | 原有九函式加本批二十函式，真實QuickC編譯 |
| 混合C／ASM | 55／6010 | 1939個C產生bytes與4071個明列ASM bytes |
| 固定LIB | 82／9400實際bytes | 52個完整CODE模組，9413範圍包含13個宣告零bytes |
| 剩餘CODE | 18948 | 包含25原始函式與函式外範圍，仍列未還原 |
| 初始化DATA | 2185 | 本批仍用原版增量基準，不計來源完成 |

本批涵蓋圖像分派、共用解碼、檔案讀寫、字串查找、幾何交疊、色盤與多種繪圖機制。
寄存器區塊由原始IDA Pro 9.4指令重寫為ASM，普通C設定、條件、呼叫及數值運算分開。
KOLBOOK的DS基址、資料欄位與實際BODY逐項核對，沒有只按TITLE函式大小推定相同。
原始名稱、位址、chunks、bytes及caller保留，全部新增完整body與原位外部引用逐byte相同。
證據等級為已證實的匹配表達，原作者的實際型別、唯一寫法與精確工具版本仍未知。

少量差分包括失敗分支形狀、解碼暫存區DS`+0x1130`、普通C前置條件的stack offset，
以及header把有陣列consumer的`g_1234`、`g_13D4`當成scalar宣告。
`offset unk_1924B`這個原始IDA運算元另以DS偏移別名綁定，原運算元留在註解，沒有猜語意改名。
修改來源後以同一QuickC／LINK重跑，沒有以ASM補普通C差異。

原有九個C函式共用BOOK.OBJ，獨立reader按真實PUBLIC offset與完整非重疊區間驗證全部594-byte OBJ，
沒有把「同OBJ還有其他函式」誤當成函式BODY不完整，也不略過已有C來源。
84個原始C／ASM來源函式、82個LIB原始函式均核對完整原始chunks。
兩個空白目錄從七十六個C源檔加共同BOOK.H乾淨重編，159項源檔、OBJ、header、清冊、
原位OBJ、MAP、native MZ與完整40192-byte EXE逐byte相同。
兩次完整EXE均與原版SHA-256 `2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832`相同，
55個外層`0x1A`、header、checksum及四筆有序重定位保持。
兩份獨立reader從真實compiler／LIB重算完整39625-byte初始化映像、全部fixup與byte／word backpatch。

九項反例拒絕實際C／ASM再編譯、新C／ASM／既有共享OBJ與LIB指令、byte回填、
共享OBJ來源漏項及header checksum；五項OMF修改仍有合法checksum。
新增來源8485 bytes之外的CODE／DATA仍明列原版基準，完整EXE相同不代替Book來源完成。
封存核對器逐檔重算SHA-256後只快取本批核對需要的檔案，避免反覆xz定位，
證據內容、雜湊與完成閘門不變；既有不可變收據不改寫。

累計95純C函式11210 bytes、345個LIB函式38978實際來源bytes、88個混合C／ASM函式17213 bytes，
混合來源含10758個C產生bytes與6455個ASM bytes；額外C CODE118 bytes與TANDYPIN入口13 bytes另列。
全六支仍有270個函式、132253 CODE bytes與8921初始化DATA bytes未還原，完整goal維持active。
正式342份Go及模組輸入保持，沒有遊戲執行、恢復對拍、RND結果對拍或硬體逐週期研究。

本機[收據](../verification/local/matching-c10-20261009/receipt.json)、
[來源清冊](../verification/local/matching-c10-20261009/source-ledger.json)、
[乾淨重編與反例](../verification/local/matching-c10-20261009/checkpoint-review.json)、
[完整來源核對器](../verification/local/matching-c10-20261009/proof-program.py)、
[來源封存](../verification/local/matching-c10-20261009/evidence.tar.xz)、
[唯讀核對器](../verification/local/matching-c10-20261009/verify.py)及
[保存清冊](../verification/local/matching-c10-20261009/retention.json)為本批入口。
完成比較的`workplace/matching-c10-20261009/`核對封存後清理，下一批接續Book剩餘25原始函式與DATA。



## 第十一批來源還原紀錄

2026-10-09第十一批KOLBOOK新增18個完整原始函式，共6012 bytes。
十個純C3499 bytes，八個混合C／ASM2513 bytes；混合來源含2116個C與397個ASM bytes。
原有84個C／ASM函式與82個完整LIB原始函式保持。
[唯一來源清冊](../verification/local/matching-c11-20261009/source-ledger.json)保存全部六支、799個原始函式。
KOLBOOK仍有七個原始函式、12936 CODE與2185初始化DATA bytes未還原。

| KOLBOOK已核對範圍 | 函式／bytes | 來源 |
|---|---|---|
| 純C | 39／6568 | 真實QuickC編譯；本批新增十個 |
| 混合C／ASM | 63／8523 | 4055個C產生bytes與4468個明列ASM bytes |
| 固定LIB | 82／9400實際bytes | 52個完整CODE模組；9413範圍含13個宣告零bytes |
| 剩餘CODE | 12936 | 七個原始函式及函式外範圍，仍列未還原 |
| 初始化DATA | 2185 | 仍列原版增量基準，不計來源完成 |

新增來源包括磁碟提示、文字選單、輸入、資料檔載入、字串組裝、動畫與幾何設定。
sub_146CE保留原始1302-byte完整函式，45 bytes為三段明列的滑鼠／鍵盤硬體ASM，
其餘條件、鍵碼、迴圈、數值運算、呼叫與返回均由C編譯。
sub_134C2的continue控制流、sub_142B9的參數型別，以及輸入函式的有號座標與運算順序，
均由原始指令與實際編譯差分核對；沒有以ASM補普通C差異。
原始名稱、IDA linear、file offset、chunks、bytes及caller保持。
已證實的是這組來源能產生相同完整BODY與外部引用，原作者的唯一型別／寫法仍未知。

四個候選未計完成：sub_14101尚有置中算式的暫存器與型別差異；
sub_15906、sub_16042、sub_1663F各有減法指令差異。
這些差異不以原始bytes、遮罩或新增ASM修補，原始範圍維持未還原。
另有_main、sub_11820與sub_1293E尚待還原。
IDA 9.4的一次性Book資料庫副本核對SHA與191個原始函式，未改名或增刪函式；
反編譯器未提供四個大函式的偽碼，指令與原生編譯結果仍為依據。
28段固定LIB初始化資料已有來源候選與重定位核對，尚未建立完整typed DATA建置，
2185-byte DATA進度不變，見封存中的data-resolved-candidates.json。

九十四個C源檔加BOOK.H在兩個空白目錄乾淨重編，195項源檔、OBJ、header、清冊、
原位OBJ、MAP、native MZ與完整EXE逐byte相同。
兩次完整40192-byte EXE的SHA-256均為
`2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832`。
55個外層1Ah、header、checksum、四筆有序重定位及完整39625-byte初始化映像保持。
兩份獨立reader核對102個完整compiler函式、52個固定LIB模組、共享BOOK.OBJ的完整區間，
並重算全部fixup與byte／word backpatch。九項反例拒絕實際C／ASM再編譯、
新C／ASM／舊共享OBJ／LIB指令、byte回填、共享來源漏項及header checksum。

累計105純C14709 bytes、345個LIB函式38978實際來源bytes、96個混合C／ASM19726 bytes。
混合來源分列12874個C與6852個ASM bytes，額外C CODE118 bytes、TANDYPIN入口13 bytes及18個epilogue bytes保持。
四支完整來源維持；Book與Town仍有252函式、126241 CODE與8921 DATA bytes，完整goal保持active。
342份Go及模組輸入保持，遊戲對拍維持暫停。

本機[收據](../verification/local/matching-c11-20261009/receipt.json)、
[來源清冊](../verification/local/matching-c11-20261009/source-ledger.json)、
[來源選擇與未匹配候選](../verification/local/matching-c11-20261009/accepted-review.json)、
[乾淨重編與反例](../verification/local/matching-c11-20261009/checkpoint-review.json)、
[原位來源核對器](../verification/local/matching-c11-20261009/proof-program.py)、
[來源封存](../verification/local/matching-c11-20261009/evidence.tar.xz)、
[唯讀核對器](../verification/local/matching-c11-20261009/verify.py)及
[保存清冊](../verification/local/matching-c11-20261009/retention.json)為本批入口。
來源與候選工具只留本機；核對封存後清理本批工作區，下一批接續Book七個函式與DATA。

## 第十二批來源還原紀錄

2026-10-09第十二批還原KOLBOOK完整主函式及全部初始化DATA。
原始_main的4789 bytes分列4562個C與227個硬體ASM bytes。
IDA linear 0x112C5..0x112D5的十六個函式外bytes也由同一C物件供應，包含清理呼叫十bytes與尾聲六bytes。
原始191個Book函式與全六支799個原始函式數保持。
[唯一來源清冊](../verification/local/matching-c12-20261009/source-ledger.json)仍把六個原始函式及8085 CODE bytes列為未還原。

| KOLBOOK已核對範圍 | 函式／bytes | 來源 |
|---|---|---|
| 純C | 39／6568 | 原有來源保持 |
| 混合C／ASM | 64／13312 | 8617個C與4695個ASM bytes |
| 固定LIB CODE | 82／9400實際bytes | 52個完整模組；9413範圍含13個宣告零bytes |
| 函式外C尾碼 | 16 | 主函式物件完整供應；清理呼叫10與尾聲6 |
| 原生CODE零 | 46 | LINK的宣告零填補，非原始片段 |
| 初始化DATA | 2185，全部有來源 | 1123個compiler C、1057個LIB與五個原生零bytes |
| 剩餘CODE | 8085 | 六原始函式與函式外區段，仍列原版來源 |

主函式包含啟動、選單、磁碟與檔案處理、角色記錄組裝、圖像編輯入口及程式切換。
普通C控制與運算不補ASM，十一段硬體寄存器／記憶體複製區塊按原始指令另列。
原始名稱、IDA linear、file offset、chunks與bytes保持，沒有改名覆蓋定位資訊。
最初把BP-2猜成使用者變數，導致初始化差異；實際for逗號初始化式會自行產生這個暫存值。
移除猜出的變數後完整4805-byte compiler物件與原位BODY及尾碼相符。
診斷工具也改為按下一個真實PUBLIC或OBJ結尾計算完整尺寸，避免較長的物件被宣告大小截短。
已證實的是可重現的匹配表達，原作者唯一型別／寫法及精確工具版本仍未知。

初始化DATA拆成78個C符號，包含word初值、二／三／五byte記錄、39個平行word、
固定寬度檔名、色盤、游標word遮罩與近字串。consumer、步幅、值與原位位置分開核對。
28個固定LIB初始化區段及PAD／EPAD共1057 bytes，含969實際初值與88宣告零bytes。
C物件1123 bytes的最後對齊由LINK供應，其他空隙合計五個原生零bytes；原始DATA blob為0。
DS+57A的0x72FA與0x8F1E觀察值已核對，用途未知，原始C宣告仍為假說。
原版DB副本讀不到輸入SHA時未越過閘門；改用前輪已核對副本，SHA與191函式通過。
原始DB保持唯讀，資料shape與xref保存於本機封存。

原位建置與獨立reader新增資料完成範圍，完整資料與未還原程式碼分開計數。
每個DATA byte均有唯一compiler／LIB／原生零來源；8085個CODE bytes仍明列原版片段。
假報CODE已完成、移除DATA來源或關閉資料完成旗標均被拒絕，整支來源完成旗標保持false。
圖像編輯器候選及置中／三個繪字函式仍未匹配，不計還原進度，也不以ASM補差異。
尚待還原的原始函式為sub_11820、sub_1293E、sub_14101、sub_15906、sub_16042、sub_1663F。

95個CODE源檔加DATA.C與BOOK.H在兩個空白目錄重編，201項C、OBJ、header、清冊、
原位OBJ、DATA來源、MAP、native MZ與完整EXE逐byte相同。
兩次40192-byte EXE的SHA-256均為
`2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832`。
55個外層1Ah、header、checksum、四筆有序重定位與39625-byte初始化映像保持。
兩份獨立reader核對103個完整C／ASM來源函式、52個LIB模組、全部DATA與fixup／byte及word回填。
十五項反例涵蓋實際C／ASM／DATA再編譯、CODE／DATA指令與值、舊共享OBJ、LIB、
byte回填、額外零區LEDATA、來源漏項、資料旗標、假報程式碼完成及header checksum。

累計105純C14709 bytes、345個LIB函式38978實際來源bytes、97個混合C／ASM24515 bytes。
混合來源分列17436個C與7079個ASM bytes，額外C118與TANDYPIN入口13 bytes保持。
Book六與Town245共251函式、121390 CODE與6736 DATA bytes未還原，四支完整來源與完整goal active保持。
342份Go及模組輸入保持，遊戲對拍仍暫停。

本機[收據](../verification/local/matching-c12-20261009/receipt.json)、
[來源清冊](../verification/local/matching-c12-20261009/source-ledger.json)、
[C初始化形狀與未知初值](../verification/local/matching-c12-20261009/C-initializer-plan.json)、
[乾淨重編與反例](../verification/local/matching-c12-20261009/checkpoint-review.json)、
[完整資料與原位來源核對器](../verification/local/matching-c12-20261009/proof-program.py)、
[來源封存](../verification/local/matching-c12-20261009/evidence.tar.xz)、
[唯讀核對器](../verification/local/matching-c12-20261009/verify.py)及
[保存清冊](../verification/local/matching-c12-20261009/retention.json)為本批入口。
完成比較的本批工作區在封存唯讀核對後清理，下一批接續Book六函式與函式外CODE。

## 第十三批來源還原紀錄

2026-10-09第十三批還原角色建立完整3671-byte函式與411個函式外CODE bytes。
sub_11820分列3646個C與25個複製ASM bytes，包含輸入、取消返回、選項、重擲、肖像、記錄組裝與保存。
for逗號初始化的暫存值及三個算式暫存由compiler自行產生，未以猜出的變數或ASM補普通C差異。
原始名稱、IDA linear、file offset、chunks、bytes與191／799原始函式數保持。
證據等級為已證實的匹配表達；原始唯一型別／寫法與精確歷史工具版本仍未知。
[唯一來源清冊](../verification/local/matching-c13-20261009/source-ledger.json)保存全六支目前來源。

| KOLBOOK已核對範圍 | 函式／bytes | 來源 |
|---|---|---|
| 純C | 39／6568 | 已有來源保持 |
| 混合C／ASM | 65／16983 | 12263個C與4720個ASM bytes |
| 固定LIB CODE | 82／9400實際bytes | 52個完整模組；9413範圍含13個宣告零bytes |
| 函式外來源 | 411 | 153個C與258個ASM bytes，另列而不增加原始函式數 |
| 主函式外尾碼 | 16 | 前輪來源保持；清理呼叫10與真正尾聲6 |
| 原生CODE零 | 46 | LINK宣告零填補保持 |
| 初始化DATA | 2185，全部有來源 | compiler C1123、固定LIB1057及原生零5保持 |
| 剩餘CODE | 4003 | 正好五個原始函式，函式外未還原區段為0 |

三個額外區段以IDA 9.4原始指令、prologue／epilogue、全段raw file bytes核對，未新增IDA函式。
導覽別名sub_14F98對應IDA linear 0x14F98..0x14FE7的79-byte數字字串輸出；
sub_15029對應0x15029..0x1513A的273-byte位元解碼，其中258 bytes為明列的寄存器ASM，15 bytes由C wrapper產生；
sub_1513A對應0x1513A..0x15175的59-byte限長文字輸出，為純C。
三段沒有直接caller，用途描述只支持本段資料流，不因此宣稱沒有間接呼叫；來源足以重建後停止深挖。
原名未知，別名只供導航；原始位址、bytes、輸入SHA、工具版本及位址空間保留。
新增source declaration為四個，新增原始函式只有sub_11820一個，原始總數不變。

99個CODE源檔加DATA.C／BOOK.H在兩個空白目錄重編，209項產物逐byte相同。
兩份獨立reader核對104個原始compiler函式、三個額外區段、52個LIB模組、全部DATA與fixup／byte及word回填。
兩次完整40192-byte EXE的SHA-256均為
`2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832`。
55個外層1Ah、header、checksum、39625-byte映像與四筆有序重定位保持。
十七反例涵蓋實際C／ASM／DATA再編譯、額外C指令、額外區段漏項、舊共享OBJ、LIB、
DATA來源與旗標、byte回填、額外零LEDATA、假報CODE完成及header checksum；九項OMF改動保持合法checksum。

QuickC2.01仍為本批匹配工具。多種寫法與lvalue探針未解出其餘差異，不計還原進度。
固定同一[候選工具來源](https://github.com/davidly/dos_compilers/tree/da3f15ce4912893a0dd9f8349dab29116a53ca59)，
以commit／tree／Git blob／SHA-256核對QuickC1與2.51修訂：QuickC1的純C prologue與配置不同且不接受本批_asm；
2.51的置中與繪字仍不匹配，圖像編輯器出現未支援的thread FIXUPP，只列未驗，不擅稱匹配。
工具存在、相似BODY或相同物件大小都不證明原作工具版本，候選資料只留本機。

仍未還原sub_1293E、sub_14101、sub_15906、sub_16042、sub_1663F，共4003 CODE bytes。
普通C差異仍不補ASM。全六支尚250函式、117308 CODE與6736 DATA bytes，完整goal active保持。
合計105純C14709、345個LIB38978、98混合28186 bytes；混合分列21082 C／7104 ASM。
函式外C由既有118加本批153成271 bytes，額外ASM258 bytes另列，TANDYPIN入口13 bytes保持。
四支完整來源、正式342份Go／模組輸入與遊戲對拍暫停保持。

本機[收據](../verification/local/matching-c13-20261009/receipt.json)、
[來源清冊](../verification/local/matching-c13-20261009/source-ledger.json)、
[額外區段導航聯集](../verification/local/matching-c13-20261009/ida-selected-extra-functions.json)、
[乾淨重編與反例](../verification/local/matching-c13-20261009/checkpoint-review.json)、
[原位來源核對器](../verification/local/matching-c13-20261009/proof-program.py)、
[來源封存](../verification/local/matching-c13-20261009/evidence.tar.xz)、
[唯讀核對器](../verification/local/matching-c13-20261009/verify.py)及
[保存清冊](../verification/local/matching-c13-20261009/retention.json)為本批入口。
已完成比較的工作區在封存唯讀核對後清理，下一批接續Book五個原始函式。

## 第十四批來源還原紀錄

2026-10-09第十四批回到KOLTOWN完整來源目標，保留Book全部已核對來源與五個未匹配函式。
Book的編譯旗標探針未解出差異，不以ASM、遮罩或原始bytes補普通C。
Town由既有51個純C5057 bytes及42個LIB原始函式接續，新增33個完整原始來源函式。
[唯一來源清冊](../verification/local/matching-c14-20261009/source-ledger.json)保存799個原始函式，沒有增加合成函式數。

| KOLTOWN來源 | 函式／bytes | 證明範圍 |
|---|---|---|
| 純C | 51／5057 | 原有來源全部重新編譯保持 |
| 真實LIB | 75原始函式／9632實際bytes | 52個完整CODE模組；範圍9645含13宣告零bytes |
| 新增來源 | 33原始函式／5431範圍bytes | 5418實際LIB bytes及13宣告零bytes |
| 未還原CODE | 212函式／107874 bytes | 玩家流程與其他未匹配CODE仍明列原版來源 |
| 初始化DATA | 6736 | 本批未還原，不因完整packed相同計完成 |

早期Town清冊只列global EXTDEF，output.c的LEXTDEF／本機符號未列，導致完整模組fixup越界。
沿已核對的清冊parser補齊local PUBDEF／EXTDEF，固定LIB檔名、SHA與bytes保持。
本機符號使用module SHA加原始名稱分開约束，不把不同模組的同名local當成同一符號。
完整模組候選按CODE segment界線、每個已供應指令byte、fixup、宣告零區與共享位置一致性核對。
局部相似只作候選；closeall.c雖有同型BODY，__iob位置與已知約束相差40，被拒絕。
包含啟動、初始化／終止、DOS檔案與記憶體、格式輸出、字串、配置、exec及整數helper的52個模組全部原位核對。
原始名稱、位址空間、原始chunks與bytes保留；有共同尾碼的函式核全部chunks，不只核第一段。
75個原始LIB函式包括原有42個，51個C沒有重疊或被退回原版片段。

兩個空白目錄重新編譯BULK.C、原位OBJ、LINK及封裝，十項C、compiler OBJ、原位OBJ、
來源清冊、native MZ、MAP、壓縮核心與完整EXE逐byte相同。
兩份獨立reader直接解析compiler及固定LIB，核唯一範圍、全部來源fixup與word回填，
重生129312-byte初始化映像與2281筆原始有序重定位，source缺漏或local目標改動均拒絕。
原生LINK4.07後以EXEMOD4.02 /MIN C34及EXEPACK4.06重生133679-byte壓縮核心。
MZ宣告EOF外的81個1Ah由常數生成，不複製原版tail；兩次133760-byte全檔SHA-256均為
`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。
沒有手改指令、header、checksum、重定位或壓縮payload。

十一反例涵蓋實際C再編譯、C與新舊LIB指令、word回填、output模組與共享C漏項、local fixup目標，
以及packed header、壓縮核心與外層填補；四項OMF修改保留合法checksum。
早期反例腳本選到多個相似word讀取式，改用完整表達式定位後在新目錄重跑，屬研究工具問題。
新增實際LIB來源5418 bytes後，LIB合計378原始函式44396實際bytes。
105純C14709、98混合28186、額外271 C／258 ASM與Book完整DATA均保持。
全六支尚217函式、111877 CODE及6736 DATA bytes，Town212／Book五分列，完整goal active。
四支完整來源不變；342份Go／模組輸入與遊戲對拍暫停保持，完整原版C證明不延伸為Go parity。

本機[收據](../verification/local/matching-c14-20261009/receipt.json)、
[來源清冊](../verification/local/matching-c14-20261009/source-ledger.json)、
[乾淨重編與反例](../verification/local/matching-c14-20261009/checkpoint-review.json)、
[完整模組掃描](../verification/local/matching-c14-20261009/town-all-module-candidates.py)、
[原生LINK與封裝](../verification/local/matching-c14-20261009/town-link.py)、
[來源封存](../verification/local/matching-c14-20261009/evidence.tar.xz)、
[唯讀獨立核對器](../verification/local/matching-c14-20261009/verify.py)及
[保存清冊](../verification/local/matching-c14-20261009/retention.json)為本批入口。
完成比較的工作區在封存唯讀重驗後清理，下一批接續Town玩家CODE與初始化資料，Book五函式仍列未還原。

## 第十五批來源還原紀錄

2026-10-09第十五批接續Town中型模型，新增30個純C原始函式3203 bytes。
原有51個C與52個完整LIB模組保持，沒有把普通C差異改用ASM補齊。
[唯一來源清冊](../verification/local/matching-c15-20261009/source-ledger.json)仍列799個原始函式。

| 範圍 | 函式／bytes | 證明範圍 |
|---|---|---|
| Town純C | 81／8260 | 新增30／3203；原有51／5057重編保持 |
| Town真實LIB | 75原始函式／9632實際bytes | 原有52完整模組、13宣告零bytes保持 |
| Town未還原 | 182函式／104671 CODE、6736 DATA bytes | 原版片段仍明列來源，不計C完成 |
| Book未還原 | 五函式／4003 CODE bytes | 完整DATA與先前來源保持 |
| 六支合計 | 135純C／17912、378 LIB／44396、98混合／28186 bytes | 尚187函式、108674 CODE及6736 DATA bytes |

新增函式涵蓋記錄迴圈、資料頁、訊息組裝、顯示模式分派、數值門檻與下限處理。
[逐函式定位](../verification/local/matching-c15-20261009/new-functions.json)保存IDA Pro9.4原名、
IDA linear、固定解壓MZ file offset、完整bytes雜湊、原始chunks與來源位置。
匹配表達為已證實；原作者的唯一C文字、實際型別與變數宣告仍未知。
本批沒有據此修改Go規則或擴大遊戲對拍聲明。

`sub_252CB`原始`[bp+var_6]`的寫入可由QuickC對有號參數及無號運算的轉換暫存重生。
手動多宣告該暫存的候選多六bytes；移除重複暫存後，完整81-byte函式原生匹配。
其他初始差異由條件方向、連鎖賦值、整數／指標表示與運算次序核清。
EXTDEF超過單筆OMF上限後按順序拆成多筆，沒有改指令、符號目標或回填值。
早期封裝報告沿用51個C的固定計數，已改由實際清冊取得，再從兩個空白目錄重編。
這些失敗屬研究來源或驗證工具，沒有寫成Go產品缺陷。

十項C、OBJ、來源清冊、MAP、native MZ及packed全檔逐byte相同；
兩份獨立reader核全部來源、fixup、word回填、129312-byte映像及2281筆有序重定位。
十三反例涵蓋新C實際再編譯、新舊C與LIB指令、word回填、新C呼叫目標、
新舊C及LIB來源漏項、local目標，以及packed header、壓縮核心與外層填補。
五項OMF修改保留合法checksum，來源漏項也會拒絕。
LINK4.07、EXEMOD4.02與EXEPACK4.06兩次重生133679-byte核心，再生成81個1Ah。
完整133760-byte `KOLTOWN.EXE`兩次SHA-256均為
`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

本機[完整C](../verification/local/matching-c15-20261009/town-source.c)、
[新增C](../verification/local/matching-c15-20261009/new-source.c)、
[收據](../verification/local/matching-c15-20261009/receipt.json)、
[乾淨重編與反例](../verification/local/matching-c15-20261009/checkpoint-review.json)、
[來源清冊](../verification/local/matching-c15-20261009/source-ledger.json)、
[唯讀獨立核對器](../verification/local/matching-c15-20261009/verify.py)及
[保存清冊](../verification/local/matching-c15-20261009/retention.json)為接續入口。
四支完整來源、Book五函式、342份正式輸入及對拍暫停保持。
完成比較的工作區依保存契約核對後清理；下一批仍接Town玩家CODE與初始化DATA。

## 第十六批來源還原紀錄

2026-10-09第十六批新增Town25個完整純C原始函式5595 bytes。
原有81個C、75個LIB原始函式及52個完整LIB模組保持，全部799個原始函式仍逐項登記。
四支完整來源保持，Book五個未匹配函式與6736個Town初始化DATA bytes沒有計入完成。

| 範圍 | 函式／bytes | 證明範圍 |
|---|---|---|
| Town純C | 106／13855 | 新增25／5595，既有81／8260重編保持 |
| Town真實LIB | 75／9632實際bytes | 52完整模組與13宣告零bytes保持 |
| Town未還原 | 157函式／99076 CODE、6736 DATA bytes | 原始片段與來源限制保持 |
| Book未還原 | 五函式／4003 CODE bytes | 全部既有來源保持 |
| 六支合計 | 160純C／23507、378 LIB／44396、98混合／28186 bytes | 尚162函式、103079 CODE及6736 DATA bytes |

新增來源涵蓋選單、文字組裝、物品／資料頁、角色狀態搬移、排序、位置判斷與戰鬥數值。
[逐函式定位](../verification/local/matching-c16-20261009/new-functions.json)保留原始名稱、
IDA Pro9.4 linear、固定解壓MZ偏移、chunks、原始bytes雜湊與compiler位置。
已證實範圍為完整原位native byte equality；原作者的唯一C文字、型別與變數名未知。

`__aFlmul`新增綁定原始`sub_2DE8C`，由`sub_2198A`／`sub_2590F`的IDA caller、
52-byte完整LIBH `lmul.asm`與實際QuickC生成呼叫共同核對。
該LIB模組已在前批來源內，不重複增加LIB函式或bytes。
compiler helper的來源狀態改按實際LIB覆蓋判斷，原始位址保持。
普通C差異由求值次序、賦值順序、參數轉換及16-byte記錄步距核清，沒有補ASM。

`sub_166CC`的近指標相加差兩bytes。七個明列C表達式探針維持相同compiler與旗標；
無號word位移能生成原始`MOV AX`、`MOV CX`、`ADD CX,AX`序列，完整269-byte函式匹配。
此結果證明可用表達，不將其中一種局部宣告稱為原作者唯一型別。
生成核對器時集合欄位的文字比對漏了右括號，修正後以相同輸入重跑；失敗屬研究工具。

兩個空白目錄重編C、compiler OBJ、原位OBJ、清冊、MAP、native MZ與packed EXE，十項逐byte相同。
兩份獨立reader直接解析compiler／LIB來源、fixup與word回填，
重生129312-byte初始化映像及2281筆有序重定位。
十四反例包含新C實際再編譯、新舊C與LIB指令、word回填、新C遠呼叫、
長整數乘法helper綁定、新舊C及LIB漏項、local目標與packed三區；六項OMF修改仍有合法checksum。
兩次LINK4.07、EXEMOD4.02與EXEPACK4.06原生重生133679-byte核心，再生成81個1Ah。
完整133760-byte `KOLTOWN.EXE`兩次SHA-256均為
`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

本機[完整C](../verification/local/matching-c16-20261009/town-source.c)、
[新增C](../verification/local/matching-c16-20261009/new-source.c)、
[來源清冊](../verification/local/matching-c16-20261009/source-ledger.json)、
[收據](../verification/local/matching-c16-20261009/receipt.json)、
[乾淨重編與反例](../verification/local/matching-c16-20261009/checkpoint-review.json)、
[指標探針](../verification/local/matching-c16-20261009/pointer-probe.json)、
[唯讀核對器](../verification/local/matching-c16-20261009/verify.py)與
[保存清冊](../verification/local/matching-c16-20261009/retention.json)為本批入口。
342份正式輸入與遊戲對拍暫停保持；完整六支matching目標仍未完成。

## 第十七批來源還原紀錄

2026-10-09第十七批新增Town43個完整原始來源函式4170 bytes。
12個純C1905 bytes涵蓋圖形分派、邊界控制、行索引與文字換行；
31個明列C／ASM2265 bytes涵蓋圖形暫存器、segment、port、字串指令與位元編解碼區塊。
31個混合函式分列465個compiler C框架與1800個ASM bytes，沒有計入純C數。
原作者的唯一C文字、實際型別與原始inline ASM形式保持未知。

| 範圍 | 函式／bytes | 證明範圍 |
|---|---|---|
| Town純C | 118／15760 | 原有106／13855保持，新增12／1905 |
| Town混合C／ASM | 31／2265 | 每函式9-byte C序言與6-byte C尾碼，ASM範圍逐項登記 |
| Town真實LIB | 75／9632實際bytes | 52完整模組、13宣告零bytes保持 |
| Town未還原 | 114函式／94906 CODE、6736 DATA bytes | CODE與DATA仍分開，不以全檔雜湊代替來源完成 |
| Book未還原 | 五函式／4003 CODE bytes | 既有來源與DATA保持 |
| 六支合計 | 172純C／25412、378 LIB／44396、129混合／30451 bytes | 尚119函式、98909 CODE與6736 DATA bytes |

[逐函式索引](../verification/local/matching-c17-20261009/new-functions.json)保留IDA Pro9.4原名、
IDA linear、固定解壓MZ偏移、完整chunks、bytes雜湊、compiler位置與來源分類。
ASM逐條沿用原始指令與分支，語法調整及全域別名另有原始定位；沒有修改IDA DB或函式邊界。
編譯器產生C序言／尾碼，所有普通C控制與換行程式保持純C；未匹配的普通far-pointer候選仍開放。
BIOS區塊只作來源重建，沒有執行遊戲或研究硬體wall-clock。

`g_6710`與`g_670A`的無號word表達及局部有號值，可重生原始比較與減法順序。
兩個除法包裝函式在普通C宣告順序下保留未定型遠呼叫，除法用無號參數，完整原位匹配。
全部既有106個C的機器碼保持；這些是有效C表達，不宣稱追回唯一歷史型別或translation unit。
候選函式擷取器誤排除內文分號、ASM反例誤取extern宣告，均修正後用新目錄重跑，失敗保留。

十項乾淨C、compiler OBJ、原位OBJ、清冊、MAP、native MZ與packed全檔逐byte相同。
兩份獨立reader核全部149個compiler函式、52個LIB模組、fixup與word回填，
重生129312-byte映像與2281筆有序重定位。
來源分類器驗證106個既有C、12個新C及31個明列ASM範圍的集合、數量、邊界與原始hash；
普通C內出現ASM、混合來源誤列純C、來源漏項或範圍錯誤均拒絕。
21個反例涵蓋實際C與ASM再編譯、新舊C、ASM及C框架指令、LIB、回填、
遠呼叫、helper綁定、來源漏項、local目標、錯誤分類／範圍與packed三區；八項OMF修改保持合法checksum。
LINK4.07、EXEMOD4.02與EXEPACK4.06兩次重生133679-byte核心，再生成81個1Ah。
完整133760-byte `KOLTOWN.EXE`兩次SHA-256均為
`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

本機[完整C／ASM](../verification/local/matching-c17-20261009/town-source.c)、
[新增來源](../verification/local/matching-c17-20261009/new-source.c)、
[來源清冊](../verification/local/matching-c17-20261009/source-ledger.json)、
[收據](../verification/local/matching-c17-20261009/receipt.json)、
[乾淨重編與反例](../verification/local/matching-c17-20261009/checkpoint-review.json)、
[ASM範圍](../verification/local/matching-c17-20261009/assembly-inputs.json)、
[純C清單](../verification/local/matching-c17-20261009/pure-inputs.json)、
[來源分類器](../verification/local/matching-c17-20261009/source_categories.py)、
[唯讀核對器](../verification/local/matching-c17-20261009/verify.py)與
[保存清冊](../verification/local/matching-c17-20261009/retention.json)為接續入口。
342份正式輸入、四支完整來源、Book五函式與遊戲對拍暫停保持，完整goal未完成。

## 第十八批來源還原紀錄

2026-10-09第十八批完成Town全部6736 bytes初始化DATA來源，原版DATA片段為0。
全部149個compiler函式、75個LIB原始函式與CODE來源保持，沒有增加或縮小799個原始函式。
本批只重建初值與原位連結；原作者唯一C宣告、部分資料語意與原始symbol定義仍未知。

| 初始化DATA來源 | bytes | 證明範圍 |
|---|---|---|
| compiler C初值 | 5600 | 246個符號，包含數值／位元表、逐個ASCII字串、兩個近指標表與遠函式指標 |
| 真實LIB初值 | 991 | 固定MLIBC／MLIBFP／LIBH的實際資料與fixup，31個範圍 |
| LIB宣告零區 | 88 | 原始段宣告、未寫入區與原版零值一致 |
| 原生零填補 | 57 | 字串／段對齊與尾端零區，拒絕額外LEDATA冒充來源 |
| compiler對齊 | 56，另列 | 兩份DATA物件中的額外零bytes不拷入原位，不重複計數 |
| 未還原 | 0 DATA／98909 CODE | Town114函式94906 CODE與Book五函式4003 CODE保持開放 |

[初值與引用清單](../verification/local/matching-c18-20261009/C-initializer-plan.json)保存每個DS offset、
大小、形狀、觀察值hash及原始consumer；原始C宣告標為hypothesis，語意未證實處保持unknown。
246個具名C初值由真實QuickC編譯，數值表與字串分開，不把6736 bytes整段包成C blob。
兩個七筆近指標表與遠指標使用獨立translation unit，讓compiler產生15筆真正external FIXUPP；
LINK依原始DS與CODE位置解析，沒有手改指標值或資料bytes。
遠指標保持原始`sub_2DB42`／匹配LIB `_execve` public定位，該資料記錄的用途仍未知。

原有LIB CODE binding與完整DATA-only模組共同定位初值及XP／MSG／PAD／EPAD。
CPP reader沿既有QuickC thread preamble契約處理，真實指標fixup只接受已觀察的明列external形式。
獨立reader從真實LIB bytes重算資料、LIDATA、fixup與宣告零值，再核compiler PUBLIC、初值及56個對齊零bytes。
來源缺漏或多加一個零值LEDATA，即使全檔值相同也會拒絕。
`__aDBdoswp`相關欄位只保留已觀察輸出值，不由定位別名推定其原始symbol定義或用途。

沿既有IDA Pro9.4 image及UID/GID1000:1000，從固定解壓MZ建立一次性DB。
正確CLI載入paragraph為`1000`，預設DS依已證實DGROUP設為`2DED`；原始輸入bytes不變。
338個函式與八個段的幾何保持，DATA匯出271項與1359筆直接引用，raw file bytes與IDA loaded bytes各列位址基準。
初次CLI單位與下一次API命名空間錯誤分類為工具問題，使用新副本修正後重跑。
沒有改正式DB、重命名原始函式、執行遊戲或恢復Go對拍。

兩次空白目錄重編全部CODE及兩份DATA來源，15項C、OBJ、來源清單、MAP、native MZ與packed全檔逐byte相同。
31反例包括原有21項CODE／C／ASM／LIB／分類控制，以及實際DATA C與近指標再編譯、
C／LIB資料bytes、近／遠指標目標、C／LIB初值漏項、額外零LEDATA與compiler對齊修改；14項OMF修改保持合法checksum。
兩份獨立reader核完整129312-byte映像、2281筆有序重定位及九個DATA relocation。
LINK4.07、EXEMOD4.02、EXEPACK4.06兩次原生重生133679-byte核心，再生成81個1Ah。
完整133760-byte `KOLTOWN.EXE`兩次SHA-256均為
`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

本機[C初值](../verification/local/matching-c18-20261009/DATA-source.c)、
[C指標](../verification/local/matching-c18-20261009/PTRS-source.c)、
[既有C／ASM](../verification/local/matching-c18-20261009/town-source.c)、
[LIB來源定位](../verification/local/matching-c18-20261009/data-resolved-candidates.json)、
[IDA資料引用](../verification/local/matching-c18-20261009/IDA-DATA-receipt.json)、
[來源清冊](../verification/local/matching-c18-20261009/source-ledger.json)、
[重編與反例](../verification/local/matching-c18-20261009/checkpoint-review.json)、
[收據](../verification/local/matching-c18-20261009/receipt.json)、
[DATA獨立核對器](../verification/local/matching-c18-20261009/data_proof.py)、
[全唯讀核對器](../verification/local/matching-c18-20261009/verify.py)及
[保存清冊](../verification/local/matching-c18-20261009/retention.json)為接續入口。
CODE計數維持172純C25412、378 LIB44396、129混合30451 bytes；DATA不混入CODE進度。
四支完整來源、Book五函式、342正式輸入與遊戲對拍暫停保持，完整goal未完成。

## 第十九批來源還原紀錄

2026-10-10第十九批新增Town24個完整原始CODE函式4337 bytes，包含3純C466與21明列C／ASM3871 bytes。
既有149個compiler函式及全部DATA來源保持；工作目錄沿用開始日20261009，不另建跨日副本。
Book已驗來源只作重用線索，Town原始指令、常數、位址與呼叫均另行核對，再以完整native函式驗收。
普通far-pointer未匹配候選`sub_1A4E4`仍保持開放，沒有藉重用ASM補普通C差異。

| 範圍 | 函式／bytes | 證明範圍 |
|---|---|---|
| 新增純C | 3／466 | 文字行處理及讀寫檔控制 |
| 新增混合C／ASM | 21／3871 | 1438個C設定／控制bytes與2433個明列ASM bytes |
| Town compiler來源 | 173函式 | 121純C16226、52混合6136 bytes，混合分列1903 C／4233 ASM |
| Town LIB | 75原始函式／9632實際bytes | 52個完整CODE模組及13宣告零bytes保持 |
| Town未還原 | 90函式／90569 CODE bytes | 全部6736初始化DATA保持，原版DATA片段為0 |
| 六支未還原 | 95函式／94572 CODE bytes | Book五函式4003 CODE保持，四支完整來源保持 |

新增來源包括兩個解碼器的C控制、圖形拷貝／遮罩／游標暫存器區塊，以及DOS讀寫檔重試。
字典far pointer、byte陣列與word陣列依已驗header及Town原始consumer核對；回傳值與近指標參數型別按實際PUSH序列修正。
`sub_1A26B`原先導航用void／整數參數改為可供consumer使用的C表達，全部既有機器碼保持。
型別與原始source spelling仍不唯一，沒有改IDA原名、函式邊界或把推測別名當語意證據。

來源轉換器曾將`0Fh`／`0A000h`等常數移除前導0，造成C2094未定義標籤。
錯誤碼依[Microsoft文件](https://learn.microsoft.com/en-us/cpp/error-messages/compiler-errors-1/compiler-error-c2094?view=msvc-170)核對；
保留合法十六進位語法後，同compiler與旗標重跑，未加ASM或放寬byte比較。
驗證器局部變數遮蔽案例路徑、反例原先把函式本體延伸到檔尾，均在新目錄修正重跑，失敗保留。

兩個空白目錄重編173個compiler函式與兩份DATA來源，15項C、OBJ、清冊、MAP、native MZ及packed全檔逐byte相同。
35反例保留原有31項CODE／DATA／LIB／分類控制，新增解碼C再編譯、新kernel ASM、純C來源漏項與新ASM範圍修改；15項OMF修改保持合法checksum。
兩份獨立reader核CODE／DATA、全部fixup／回填、129312-byte映像與2281筆有序重定位，九個DATA relocation保持。
LINK4.07、EXEMOD4.02、EXEPACK4.06兩次原生重生133679-byte核心，再生成81個1Ah。
完整133760-byte `KOLTOWN.EXE`兩次SHA-256均為
`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

本機[新函式與範圍](../verification/local/matching-c19-20261009/new-functions.json)、
[新增來源](../verification/local/matching-c19-20261009/new-source.c)、
[完整C／ASM](../verification/local/matching-c19-20261009/town-source.c)、
[來源清冊](../verification/local/matching-c19-20261009/source-ledger.json)、
[重編與反例](../verification/local/matching-c19-20261009/checkpoint-review.json)、
[收據](../verification/local/matching-c19-20261009/receipt.json)、
[來源分類器](../verification/local/matching-c19-20261009/source_categories.py)、
[唯讀核對器](../verification/local/matching-c19-20261009/verify.py)及
[保存清冊](../verification/local/matching-c19-20261009/retention.json)為本批入口。
總計175純C25878、378 LIB44396、150混合34322 bytes，函式外271 C／258 ASM另列。
342正式輸入、全部DATA、Book五函式、四支完整來源及遊戲對拍暫停保持，完整goal未完成。

## 第二十批來源還原紀錄

2026-10-10 第二十批新增 Town 16 個完整原始 CODE 函式，共 3823 bytes。
全部 173 個既有 compiler 函式及初始化 DATA 來源保持。原始函式總數仍為 799。

| 範圍 | 函式／bytes | 證明範圍 |
|---|---|---|
| 新增純 C | 1／276 | `sub_1B5C8` 的輸入字元篩選、長度限制與退格控制 |
| 新增混合 C／ASM | 15／3547 | 3086 個 C 運算／控制／呼叫 bytes，461 個明列暫存器 bytes |
| Town compiler 來源 | 189 函式 | 122 純 C 16502、67 混合 9683 bytes，混合分列 4989 C／4694 ASM |
| Town LIB | 75 原始函式／9632 實際 bytes | 52 個完整 CODE 模組及 13 宣告零 bytes 保持 |
| Town 未還原 | 74 函式／86746 CODE bytes | 全部 6736 初始化 DATA 保持，原版 DATA 片段為 0 |
| 六支未還原 | 79 函式／90749 CODE bytes | Book 五函式 4003 CODE 保持，四支完整來源保持 |

新增來源涵蓋圖像載入與搬移、地圖樣本讀取、隊員文字呈現、鍵盤清除及輸入編輯。
C 使用已核對的近指標、50-byte 記錄與 DATA 宣告；運算順序依原始指令及實際 compiler 結果確認。
直接操作 ES／DS、BIOS 資料區與解碼暫存器的區段明列原始 IDA linear 範圍，不計為純 C。
已證實的是固定工具鏈的完整函式 byte equality；唯一原始宣告、文字與原始 ASM 寫法仍未知。

三個字型候選 `sub_1CE04`、`sub_1D755`、`sub_1DE84` 的 C 減法及游標增量指令仍不匹配，
本批未納入來源覆蓋，也未擴大 ASM 範圍補差異。普通 far-pointer 候選 `sub_1A4E4` 同樣保持開放。
反例腳本曾誤選函式前置宣告，另有整數索引遮蔽來源清冊變數；修正腳本後於新案例目錄完整重跑。
這些失敗屬驗證腳本，未修改原版結果或降低比較條件。

兩個空白目錄重編 189 個 compiler 函式與兩份 DATA 來源；15 項 C、OBJ、清冊、MAP、native MZ 及 packed 全檔逐 byte 相同。
39 反例全部拒絕，包含前批 35 項與本批輸入 C、BIOS ASM 的實際再編譯，以及漏函式和錯誤 ASM 範圍。
15 項 OMF 修改保持合法 checksum。兩份獨立 reader 核 CODE／DATA、fixup／回填、129312-byte 映像及 2281 筆有序重定位。
九個 DATA relocation 與 15 個 compiler 指標 fixup 保持。
LINK4.07、EXEMOD4.02、EXEPACK4.06 兩次原生重生 133679-byte 核心，再生成 81 個 1Ah。
完整 133760-byte `KOLTOWN.EXE` 的 SHA-256 兩次均為
`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

本機[新函式與範圍](../verification/local/matching-c20-20261010/new-functions.json)、
[新增來源](../verification/local/matching-c20-20261010/new-source.c)、
[完整 C／ASM](../verification/local/matching-c20-20261010/town-source.c)、
[來源清冊](../verification/local/matching-c20-20261010/source-ledger.json)、
[重編與反例](../verification/local/matching-c20-20261010/checkpoint-review.json)、
[收據](../verification/local/matching-c20-20261010/receipt.json)、
[來源分類器](../verification/local/matching-c20-20261010/source_categories.py)、
[唯讀核對器](../verification/local/matching-c20-20261010/verify.py)及
[保存清冊](../verification/local/matching-c20-20261010/retention.json)為本批入口。
累計 176 純 C 26154、378 LIB 44396、165 混合 37869 bytes，函式外 271 C／258 ASM 另列。
342 份正式輸入、全部 DATA、Book 五函式及遊戲對拍暫停保持；完整 matching 尚未完成。

## 第二十一批來源還原紀錄

2026-10-10 第二十一批新增 Town 14 個純 C 原始函式，共 5268 bytes，沒有新增 ASM。
全部 189 個既有 compiler 函式、52 個 LIB CODE 模組與初始化 DATA 來源保持；原始函式總數仍為 799。

| 範圍 | 函式／bytes | 證明範圍 |
|---|---|---|
| 新增純 C | 14／5268 | 文字控制、畫面控制、命令選擇、地圖取樣、方向及日期計算 |
| Town compiler 來源 | 203 函式 | 136 純 C 21770、67 混合 9683 bytes，混合分列 4989 C／4694 ASM |
| Town LIB | 75 原始函式／9632 實際 bytes | 52 個完整 CODE 模組及 13 宣告零 bytes 保持 |
| Town 未還原 | 60 函式／81478 CODE bytes | 全部 6736 初始化 DATA 保持，原版 DATA 片段為 0 |
| 六支未還原 | 65 函式／85481 CODE bytes | Book 五函式 4003 CODE 保持，四支完整來源保持 |

`sub_1A8FD` 的控制字元、十六進位參數、置中與逐行掃描，以及 `sub_1103F`、`sub_11390`、`sub_11722` 的畫面及輸入控制已有完整 C。
其他來源包含記錄交換、顯示提示、解碼呼叫包裝、地圖範圍讀取、圖像位置、狀態百分比、方向與日期計算。
功能名稱只作導航；已證實的範圍為固定 QuickC 工具鏈的完整原始函式 byte equality，不宣稱原始唯一宣告或全部遊戲語意。

C 表達依原始指令確認 signed／unsigned 比較、近指標相加順序及低 16-bit 乘除。
`g_005C`、`g_006C` 使用 unsigned word 表達，原始 signed 分支明示轉型；`sub_1C94E` 回傳型別供原始 AX consumer 使用。
全部 189 個既有函式機器碼保持。遠函式呼叫使用合法 C 函式指標轉型，保留五個實際 stack words。
迴圈用 C 標籤表達原始入口與回跳，不以 ASM 補普通 C 指令差異。
候選期間的 far 函式指標語法錯誤、迴圈及比較差異均保留在封存；最後 compiler 無警告且全部 203 函式匹配。
前批三個字型 C 候選與 `sub_1A4E4` 仍未匹配，不列入本批來源覆蓋。

兩個空白目錄重編 203 個 compiler 函式與兩份 DATA 來源；15 項 C、OBJ、清冊、MAP、native MZ 及 packed 全檔逐 byte 相同。
43 反例全部拒絕，保留前批 39 項，新增共用文字控制與命令回傳值的實際 C 再編譯、來源漏項及純 C 插入 ASM。
15 項 OMF 修改保持合法 checksum。兩份獨立 reader 核 CODE／DATA、fixup／回填、129312-byte 映像與 2281 筆有序重定位。
九個 DATA relocation、15 個 compiler 指標 fixup、56 個獨立 compiler 對齊零 bytes 均保持。
LINK4.07、EXEMOD4.02、EXEPACK4.06 兩次原生重生 133679-byte 核心，再生成 81 個 1Ah。
完整 133760-byte `KOLTOWN.EXE` 的 SHA-256 兩次均為
`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

本機[新函式與範圍](../verification/local/matching-c21-20261010/new-functions.json)、
[新增純 C](../verification/local/matching-c21-20261010/new-source.c)、
[完整 C／ASM](../verification/local/matching-c21-20261010/town-source.c)、
[來源清冊](../verification/local/matching-c21-20261010/source-ledger.json)、
[重編與反例](../verification/local/matching-c21-20261010/checkpoint-review.json)、
[收據](../verification/local/matching-c21-20261010/receipt.json)、
[來源分類器](../verification/local/matching-c21-20261010/source_categories.py)、
[唯讀核對器](../verification/local/matching-c21-20261010/verify.py)及
[保存清冊](../verification/local/matching-c21-20261010/retention.json)為本批入口。
累計 190 純 C 31422、378 LIB 44396、165 混合 37869 bytes，函式外 271 C／258 ASM 另列。
342 份正式輸入、全部 DATA、Book 五函式及遊戲對拍暫停保持；完整 matching 尚未完成。

## 第二十二批來源還原紀錄

2026-10-10 第二十二批新增 Town 14 個完整原始來源函式，共 6851 bytes。
全部 203 個既有 compiler 函式、52 個 LIB CODE 模組與完整 DATA 鏈保持；原始函式總數仍為 799。

| 範圍 | 函式／bytes | 證明範圍 |
|---|---|---|
| 新增純 C | 12／5944 | 角色選擇、記錄旗標、路徑步進、移動判定、文字替換及顯示控制 |
| 新增混合 C／ASM | 2／907 | 493 C／414 ASM bytes，含兩個 ES 讀取與三個顯示暫存器區段 |
| Town compiler 來源 | 217 函式 | 148 純 C 27714、69 混合 10590 bytes，混合分列 5482 C／5108 ASM |
| Town LIB | 75 原始函式／9632 實際 bytes | 52 個完整 CODE 模組及 13 宣告零 bytes 保持 |
| Town 未還原 | 46 函式／74627 CODE bytes | 全部 6736 初始化 DATA 保持，原版 DATA 片段為 0 |
| 六支未還原 | 51 函式／78630 CODE bytes | Book 五函式 4003 CODE 保持，四支完整來源保持 |

`sub_24C35`、`sub_24FD3`、`sub_2578F`、`sub_25D7A`、`sub_26384`、`sub_265F3` 已有純 C。
其他純 C 包含 `sub_14564`、`sub_15920`、`sub_18D61`、`sub_27FC1`、`sub_29987`、`sub_29DDE`。
混合來源 `sub_25F8A` 只有兩個 ES 地圖讀取區段列為 ASM，其餘 452 bytes 由 C 產生；
`sub_2A18E` 的三個顯示暫存器區段共 384 ASM bytes，41 個 C 框架／分支 bytes 另列。
功能名稱只作導航；已證實的是固定工具鏈的完整原始函式 byte equality，原始唯一宣告與原始 ASM 寫法仍未知。

C 表達按實際乘除呼叫順序、signed／unsigned 分支、近指標運算與 stack words 核對。
長整數式 `sub_2578F` 最初把除數與被除數顛倒；依原版參數順序修正後完整匹配，不以同長度當成功。
共用出口與 switch 的空 default 依原始控制流表達，未自行補固定回傳值。
局部陣列、word 比較與參數宣告由完整 compiler 差分確認；`g_008A` 的導航型別調整後，全部 203 個既有函式機器碼保持。
前置宣告與定義衝突造成一次編譯失敗，修正 C 宣告後同 compiler／旗標重跑；最後 compiler 無警告。
普通 far-pointer 與前批三個字型 C 候選保持未匹配，沒有擴大 ASM 範圍補普通 C 差異。

兩個空白案例重編 217 個 compiler 函式與兩份 DATA 來源，15 項 C、OBJ、清冊、MAP、native MZ 及 packed 全檔逐 byte 相同。
47 反例全部拒絕，保留前批 43 項，新增路徑 C 與地圖 ASM 的實際再編譯，以及漏純 C 函式和錯誤顯示 ASM 範圍。
15 項 OMF 修改保持合法 checksum。兩份獨立 reader 核 CODE／DATA、fixup／回填、129312-byte 映像與 2281 筆有序重定位。
九個 DATA relocation、15 個 compiler 指標 fixup、56 個獨立 compiler 對齊零 bytes 均保持。
LINK4.07、EXEMOD4.02、EXEPACK4.06 兩次原生重生 133679-byte 核心，再生成 81 個 1Ah。
完整 133760-byte `KOLTOWN.EXE` 的 SHA-256 兩次均為
`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

本機[新函式與範圍](../verification/local/matching-c22-20261010/new-functions.json)、
[新增來源](../verification/local/matching-c22-20261010/new-source.c)、
[完整 C／ASM](../verification/local/matching-c22-20261010/town-source.c)、
[來源清冊](../verification/local/matching-c22-20261010/source-ledger.json)、
[重編與反例](../verification/local/matching-c22-20261010/checkpoint-review.json)、
[收據](../verification/local/matching-c22-20261010/receipt.json)、
[來源分類器](../verification/local/matching-c22-20261010/source_categories.py)、
[唯讀核對器](../verification/local/matching-c22-20261010/verify.py)及
[保存清冊](../verification/local/matching-c22-20261010/retention.json)為本批入口。
累計 202 純 C 37366、378 LIB 44396、167 混合 38776 bytes，函式外 271 C／258 ASM 另列。
342 份正式輸入、全部 DATA、Book 五函式及遊戲對拍暫停保持；完整 matching 尚未完成。

## 第二十三批來源還原紀錄

2026-10-10 第二十三批新增 Town 主函式、選單與輸入三個完整來源函式，共 6095 bytes。
全部 217 個既有 compiler 函式、52 個 LIB CODE 模組與完整 DATA 鏈保持；原始函式總數仍為 799。

| 原始函式 | 來源範圍 | 證明範圍 |
|---|---|---|
| `sub_10000` | 3548 bytes，3388 C／160 ASM | 完整初始化、載入與主輸入控制流；六個 BIOS／DOS／顯示區段另列 |
| `sub_1AAE8` | 純 C 1355 bytes | 完整選單文字掃描、選擇、捲動、滑鼠與命令控制 |
| `sub_1B100` | 1192 bytes，1145 C／47 ASM | 完整輸入控制；兩個滑鼠暫存器與一個 BIOS 鍵盤區段另列 |
| Town compiler 來源 | 220 函式 | 149 純 C 29069、71 混合 15330 bytes，混合分列 10015 C／5315 ASM |
| Town LIB | 75 原始函式／9632 實際 bytes | 52 個完整 CODE 模組及 13 宣告零 bytes 保持 |
| Town 未還原 | 43 函式／68532 CODE bytes | 全部 6736 初始化 DATA 保持，原版 DATA 片段為 0 |
| 六支未還原 | 48 函式／72535 CODE bytes | Book 五函式 4003 CODE 保持，四支完整來源保持 |

主函式的 1095 條原始指令以 C 控制流與六個明列硬體區段重生。資源載入、邊界處理、世界／選單切換、角色選擇與退出分支按原始定位保存。
輸入與選單來源保留原版分支、word 運算、按鍵／滑鼠回傳及字串索引；沒有自行猜補未訪玩家行為。
已證實的範圍為固定 QuickC 工具鏈的完整原始函式 byte equality；原始唯一宣告、原始 ASM 寫法與未證實語意仍未知。

主函式以明示 near-word 讀寫保留 `g_2528` 的 far-pointer 形狀，選單的兩個相等比較以 unsigned word 視圖保留原始指令。
其他 signed 運算、乘除、迴圈與比較維持。九個 ASM 區段限 BIOS／DOS／顯示／滑鼠暫存器；普通 C 未匹配候選沒有加入 ASM。
候選期間把 `g_668A` 陣列當 scalar 指派造成一次編譯失敗，改用第一元素後重跑；其餘迴圈／比較差異以 C 修正。
最後 compiler 無警告，全部 220 個完整函式匹配，全部既有來源機器碼保持。

兩個空白案例重編全部來源與 DATA，15 項 C、OBJ、清冊、MAP、native MZ 及 packed 全檔逐 byte 相同。
51 反例全部拒絕，保留前批 47 項，新增主函式 C／ASM 的實際再編譯、漏選單純 C 與錯誤輸入 ASM 範圍。
15 項 OMF 修改保持合法 checksum。兩份獨立 reader 核 CODE／DATA、fixup／回填、129312-byte 映像與 2281 筆有序重定位。
九個 DATA relocation、15 個 compiler 指標 fixup、56 個獨立 compiler 對齊零 bytes 均保持。
LINK4.07、EXEMOD4.02、EXEPACK4.06 兩次原生重生 133679-byte 核心，再生成 81 個 1Ah。
完整 133760-byte `KOLTOWN.EXE` 的 SHA-256 兩次均為
`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

本機[新函式與範圍](../verification/local/matching-c23-20261010/new-functions.json)、
[新增來源](../verification/local/matching-c23-20261010/new-source.c)、
[完整 C／ASM](../verification/local/matching-c23-20261010/town-source.c)、
[來源清冊](../verification/local/matching-c23-20261010/source-ledger.json)、
[重編與反例](../verification/local/matching-c23-20261010/checkpoint-review.json)、
[收據](../verification/local/matching-c23-20261010/receipt.json)、
[來源分類器](../verification/local/matching-c23-20261010/source_categories.py)、
[唯讀核對器](../verification/local/matching-c23-20261010/verify.py)及
[保存清冊](../verification/local/matching-c23-20261010/retention.json)為本批入口。
累計 203 純 C 38721、378 LIB 44396、169 混合 43516 bytes，函式外 271 C／258 ASM 另列。
342 份正式輸入、全部 DATA、Book 五函式及遊戲對拍暫停保持；完整 matching 尚未完成。

## 第二十四批來源還原紀錄

2026-10-10 第二十四批新增完整城鎮互動控制函式 `sub_11A93`，共 8967 bytes。
全部 220 個既有 compiler 函式、52 個 LIB CODE 模組與完整 DATA 鏈保持；原始函式總數仍為 799。

| 範圍 | 函式／bytes | 證明範圍 |
|---|---|---|
| 新增 `sub_11A93` | 8967 bytes，8939 C／28 ASM | 2874 條原始指令的完整控制流、資料操作及呼叫；單一 ES 圖像搬移區段另列 |
| Town compiler 來源 | 221 函式 | 149 純 C 29069、72 混合 24297 bytes，混合分列 18954 C／5343 ASM |
| Town LIB | 75 原始函式／9632 實際 bytes | 52 個完整 CODE 模組及 13 宣告零 bytes 保持 |
| Town 未還原 | 42 函式／59565 CODE bytes | 全部 6736 初始化 DATA 保持，原版 DATA 片段為 0 |
| 六支未還原 | 47 函式／63568 CODE bytes | Book 五函式 4003 CODE 保持，四支完整來源保持 |

來源保留角色檢查、對話與服務選單、交易、物品及訓練控制的完整原始分支，包含取消、重試、跳轉與共用出口。
功能名稱只作導航；已證實的範圍為固定工具鏈的完整原始函式 byte equality，原始唯一宣告、原始 ASM 寫法與未證實語意仍未知。

長整數金額運算、byte 記錄、字串／表格索引、signed／unsigned 比較及迴圈均由 C 重生。
唯一 ASM 範圍為 IDA linear 半開範圍 `[0x12CF2,0x12D0E)` 的 28-byte ES 圖像搬移，其他 8939 bytes 由 C 產生。
原始函式位址為 IDA linear `0x11A93`；原始解壓檔案偏移為 `16019`，兩種位址基準分開記錄。
沒有以 ASM 補普通 C 差異，也沒有改原始 IDA 名稱或函式邊界。

候選差異包括多出的提示呼叫、兩處冗餘回跳、指標／長整數求值順序及 word 比較。
逐項按原始指令與真實 compiler 差分修正；最後無警告且全部 221 個完整函式匹配。
原始 compiler OMF 的未回填反組譯只作交叉檢查；其局部跳轉尚有 word backpatch，不當作最終控制流或 native 證明。
前批三個字型 C 候選與普通 far-pointer 候選保持未匹配。

兩個空白案例重編全部來源與 DATA，15 項 C、OBJ、清冊、MAP、native MZ 及 packed 全檔逐 byte 相同。
55 反例全部拒絕，保留前批 51 項，新增交易金額 C 與圖像搬移 ASM 的實際再編譯、漏城鎮控制函式及錯誤 ASM 範圍。
15 項 OMF 修改保持合法 checksum。兩份獨立 reader 核 CODE／DATA、fixup／回填、129312-byte 映像與 2281 筆有序重定位。
九個 DATA relocation、15 個 compiler 指標 fixup、56 個獨立 compiler 對齊零 bytes 均保持。
LINK4.07、EXEMOD4.02、EXEPACK4.06 兩次原生重生 133679-byte 核心，再生成 81 個 1Ah。
完整 133760-byte `KOLTOWN.EXE` 的 SHA-256 兩次均為
`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

本機[新函式與範圍](../verification/local/matching-c24-20261010/new-functions.json)、
[新增來源](../verification/local/matching-c24-20261010/new-source.c)、
[完整 C／ASM](../verification/local/matching-c24-20261010/town-source.c)、
[來源清冊](../verification/local/matching-c24-20261010/source-ledger.json)、
[重編與反例](../verification/local/matching-c24-20261010/checkpoint-review.json)、
[收據](../verification/local/matching-c24-20261010/receipt.json)、
[來源分類器](../verification/local/matching-c24-20261010/source_categories.py)、
[唯讀核對器](../verification/local/matching-c24-20261010/verify.py)及
[保存清冊](../verification/local/matching-c24-20261010/retention.json)為本批入口。
累計 203 純 C 38721、378 LIB 44396、170 混合 52483 bytes，函式外 271 C／258 ASM 另列。
342 份正式輸入、全部 DATA、Book 五函式及遊戲對拍暫停保持；完整 matching 尚未完成。

## 持續C來源還原

2026-10-10 第三十四批新增四個完整函式，共 4203 bytes，分列 3701 C／502 ASM；1453 條原始指令全部匹配。
三個純 C 共 3212 bytes，一個混合共 991 bytes。全部 254 個既有 compiler 函式、52 個 LIB CODE 模組與初始化 DATA 保持。

| 原始 IDA 名稱 | IDA linear | 解壓檔案偏移 | bytes | C／ASM | 已核對範圍 |
|---|---|---|---|---|---|
| `sub_1BD17` | `0x1BD17` | 57623 | 991 | 489／502 | 裁切、分模式複製／合成、暫存與回呼；六段硬體 ASM 另列 |
| `sub_2AAA8` | `0x2AAA8` | 118440 | 1623 | 1623／0 | 人物選擇、32-bit 數值比較、確認、戰鬥準備與返回 |
| `sub_2B0FF` | `0x2B0FF` | 120063 | 722 | 722／0 | 事件資料、訊息、條件掃描、戰鬥準備與返回 |
| `sub_2B4A7` | `0x2B4A7` | 120999 | 867 | 867／0 | 多頁文字、畫面參數、翻頁與恢復 |

功能名稱只作導航。已證實的是固定工具鏈產生完整相同函式；原始唯一 C 宣告、文字、型別與未證實語意仍未知。
圖像常式的六段硬體 ASM 共 502 bytes，編譯前已明列；其餘 C 控制共 489 bytes。普通 C 差異沒有改用 ASM。
IDA linear 半開範圍為 `[0x1BE19,0x1BE55)`、`[0x1BE62,0x1BEB2)`、`[0x1BEBF,0x1BF05)`、`[0x1BF1F,0x1BF6F)`、`[0x1BF95,0x1BFF8)` 與 `[0x1C019,0x1C08A)`。
有號常數、布林求值順序、近指標運算、32-bit 比較與未使用的運算式結果按原始指令保留。三個字型候選與 `sub_1A4E4` 仍未匹配。

| 範圍 | 函式／bytes | 證明範圍 |
|---|---|---|
| 真實 compiler CODE 物件 | 53366／13650／1707／3679／5221／6038／5554／6558／5611／6348／4203 bytes | 前十個 C／OBJ 與前批逐 byte 相同；SUPPORT.C／SUPPORT.OBJ 供應新四函式 |
| Town compiler 來源 | 258 函式 | 174 純 C 64776、84 混合 47159 bytes，混合分列 40268 C／6891 ASM |
| Town LIB | 75 原始函式／9632 實際 bytes | 52 個完整 CODE 模組及 13 宣告零 bytes 保持 |
| Town 未還原 | 5 函式／996 CODE bytes | 全部 6736 初始化 DATA 保持，原版 DATA 片段為 0 |
| 六支未還原 | 10 函式／4999 CODE bytes | Book 五函式 4003 CODE 與四支完整來源保持 |

十一個真實 CODE 段各自小於 64 KiB。111935-byte 合併只作分析座標，每筆 PUBLIC／fixup／回填保留實際物件、段與偏移。
兩個空白案例的 36 項來源、OBJ、清冊、MAP、native MZ 與 packed 全檔逐 byte 相同。
116 反例全部拒絕，包含前批全部 110 項與新裁切 C、人物選擇 C、圖像 ASM 再編譯、漏分頁 C、錯硬體 ASM 範圍與實際 SUPPORT PUBLIC 修改；25 項 OMF 修改保持合法 checksum。
前批 801／800 長度反例與兩批失敗紀錄保持獨立核對。兩份 reader 核 CODE／DATA、真實物件來源、fixup／回填、129312-byte 映像與 2281 有序重定位。
九個 DATA relocation、15 個 compiler 指標 fixup 與 56 個獨立 compiler 對齊零 bytes 保持。
LINK4.07／EXEMOD4.02／EXEPACK4.06 兩次原生重生 133679-byte 核心，再生成 81 個 1Ah。
完整 133760-byte KOLTOWN.EXE 的 SHA-256 兩次均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

本機[新函式與範圍](../verification/local/matching-c34-20261010/new-functions.json)、
[實際新編譯來源](../verification/local/matching-c34-20261010/new-source.c)、
[完整來源閱讀視圖](../verification/local/matching-c34-20261010/town-source.c)、
[來源清冊](../verification/local/matching-c34-20261010/source-ledger.json)、
[重編與反例](../verification/local/matching-c34-20261010/checkpoint-review.json)、
[收據](../verification/local/matching-c34-20261010/receipt.json)、
[來源分類器](../verification/local/matching-c34-20261010/source_categories.py)、
[真實物件讀取器](../verification/local/matching-c34-20261010/compiler_objects.py)、
[編譯單元清單](../verification/local/matching-c34-20261010/seed-compiler-units.json)、
[唯讀核對器](../verification/local/matching-c34-20261010/verify.py)及
[保存清冊](../verification/local/matching-c34-20261010/retention.json)為本批入口。
累計 228 純 C 74428、378 LIB 44396、182 混合 75345 bytes，函式外 271 C／258 ASM 另列。
[剩餘範圍核對](../verification/local/matching-c34-20261010/remaining-coverage-review.json)逐 byte 合併實際來源：
`sub_2B9BE` 的 236 bytes 中，235 已由兩個 CRT 模組覆蓋，1 byte 仍未還原；另有 41 未還原 CODE bytes 位於原始函式外。
本批四個目標與既有來源無重疊。這些範圍已納入 996／4999 CODE bytes，不把完整函式長度重複加進剩餘量。
[剩餘 C 試驗](../verification/local/matching-c34-20261010/probe-summary.json)另保存 12 個受控源碼／版本／前綴試驗，沒有增加匹配函式數。
近指標差、長整數、無號常數、參數型別與運算式求值的字型候選仍不符。QuickC1 候選拒絕測過的 `_asm` 源碼；其純 C 前綴也不符。
QuickC251 可接受本次 `_segment`／`_based` 宣告，但測過的完整字型與遠指標葉函式仍未匹配。只限這些源碼／參數，不推定其他寫法不可能。
正式重建仍使用原九項工具／錯誤資源；替代工具另有固定 commit、Git blob 與 SHA-256，只作研究。
另有[接續 C35 試驗](../../docs/verification/local/workspace-cleanup-20261024/browse/workplace/matching-c35-20261010/PLAN.json)的五個先行候選，尚未取得完整匹配或整合；目前還原量以 C34 為準。
342 份正式輸入、全部 DATA、Book 五函式與遊戲對拍暫停保持；完整 matching 尚未完成。

## 第三十三批來源還原紀錄

2026-10-10 第三十三批新增四個完整純 C 函式，共 6348 bytes；2048 條原始指令全部匹配。
全部 250 個既有 compiler 函式、52 個 LIB CODE 模組與初始化 DATA 保持。

| 原始 IDA 名稱 | IDA linear | 解壓檔案偏移 | bytes | C／ASM | 已核對範圍 |
|---|---|---|---|---|---|
| `sub_2818D` | `0x2818D` | 107917 | 1759 | 1759／0 | 完整戰鬥效果、部位／狀態、數值回填與訊息 |
| `sub_289EE` | `0x289EE` | 110062 | 1391 | 1391／0 | 掉物產生、遠指標資料讀取、選人、交換與重試 |
| `sub_29004` | `0x29004` | 111620 | 2030 | 2030／0 | 世界輸入、移動、格點切換、事件與選單 |
| `sub_2A4E5` | `0x2A4E5` | 116965 | 1168 | 1168／0 | 遭遇篩選、權重、分支、介面準備與返回 |

功能名稱只作導航。已證實的是固定工具鏈產生完整相同函式；原始唯一 C 宣告、文字、型別與未證實語意仍未知。
本批沒有 ASM。遠指標讀取、三處 DS 轉換、有號／無號比較、求值順序與隨機呼叫均由 C 重生。
原版陣列基址比較、被跳過的按鍵分支與清理／重試跳轉保持，沒有按推測修正玩法。三個字型候選與 `sub_1A4E4` 仍未匹配。

| 範圍 | 函式／bytes | 證明範圍 |
|---|---|---|
| 真實 compiler CODE 物件 | 53366／13650／1707／3679／5221／6038／5554／6558／5611／6348 bytes | 前九個 C／OBJ 與前批逐 byte 相同；WORLD.C／WORLD.OBJ 供應新四函式 |
| Town compiler 來源 | 254 函式 | 171 純 C 61564、83 混合 46168 bytes，混合分列 39779 C／6389 ASM |
| Town LIB | 75 原始函式／9632 實際 bytes | 52 個完整 CODE 模組及 13 宣告零 bytes 保持 |
| Town 未還原 | 9 函式／5199 CODE bytes | 全部 6736 初始化 DATA 保持，原版 DATA 片段為 0 |
| 六支未還原 | 14 函式／9202 CODE bytes | Book 五函式 4003 CODE 與四支完整來源保持 |

十個真實 CODE 段各自小於 64 KiB。107732-byte 合併只作分析座標，每筆 PUBLIC／fixup／回填保留實際物件、段與偏移。
兩個空白案例的 34 項來源、OBJ、清冊、MAP、native MZ 與 packed 全檔逐 byte 相同。
110 反例全部拒絕，包含前批全部 104 項與新效果 C、世界 C 再編譯、漏掉物 C、偽造 ASM 分類、錯誤物件來源與實際 WORLD PUBLIC 修改；24 項 OMF 修改保持合法 checksum。
前批 801／800 長度反例與兩批失敗紀錄保持獨立核對。兩份 reader 核 CODE／DATA、真實物件來源、fixup／回填、129312-byte 映像與 2281 有序重定位。
九個 DATA relocation、15 個 compiler 指標 fixup 與 56 個獨立 compiler 對齊零 bytes 保持。
LINK4.07／EXEMOD4.02／EXEPACK4.06 兩次原生重生 133679-byte 核心，再生成 81 個 1Ah。
完整 133760-byte KOLTOWN.EXE 的 SHA-256 兩次均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

本機[新函式與範圍](../verification/local/matching-c33-20261010/new-functions.json)、
[實際新編譯來源](../verification/local/matching-c33-20261010/new-source.c)、
[完整來源閱讀視圖](../verification/local/matching-c33-20261010/town-source.c)、
[來源清冊](../verification/local/matching-c33-20261010/source-ledger.json)、
[重編與反例](../verification/local/matching-c33-20261010/checkpoint-review.json)、
[收據](../verification/local/matching-c33-20261010/receipt.json)、
[來源分類器](../verification/local/matching-c33-20261010/source_categories.py)、
[真實物件讀取器](../verification/local/matching-c33-20261010/compiler_objects.py)、
[編譯單元清單](../verification/local/matching-c33-20261010/seed-compiler-units.json)、
[唯讀核對器](../verification/local/matching-c33-20261010/verify.py)及
[保存清冊](../verification/local/matching-c33-20261010/retention.json)為本批入口。
累計 225 純 C 71216、378 LIB 44396、181 混合 74354 bytes，函式外 271 C／258 ASM 另列。
[剩餘範圍核對](../verification/local/matching-c33-20261010/remaining-coverage-review.json)逐 byte 合併實際來源：
`sub_2B9BE` 的 236 bytes 中，235 已由兩個 CRT 模組覆蓋，1 byte 仍未還原；另有 41 未還原 CODE bytes 位於原始函式外。
本批四個目標與既有來源無重疊。這些範圍已納入 5199／9202 CODE bytes，不把完整函式長度重複加進剩餘量。
342 份正式輸入、全部 DATA、Book 五函式與遊戲對拍暫停保持；完整 matching 尚未完成。

## 第三十二批來源還原紀錄

2026-10-10 第三十二批新增四個完整純 C 函式，共 5611 bytes；1891 條原始指令全部匹配。
全部 246 個既有 compiler 函式、52 個 LIB CODE 模組與初始化 DATA 保持。

| 原始 IDA 名稱 | IDA linear | 解壓檔案偏移 | bytes | C／ASM | 已核對範圍 |
|---|---|---|---|---|---|
| `sub_253E7` | `0x253E7` | 96231 | 620 | 620／0 | 完整目標掃描、方向與候選權重 |
| `sub_25A3A` | `0x25A3A` | 97850 | 725 | 725／0 | 32-bit 權重、正規化、隨機選擇與 fallback |
| `sub_26D76` | `0x26D76` | 102774 | 2686 | 2686／0 | 完整攻擊分支、命中／防禦、損傷與訊息控制 |
| `sub_277F4` | `0x277F4` | 105460 | 1580 | 1580／0 | 傷害回填、部位／狀態、位移與返回控制 |

功能名稱只作導航。已證實的是固定工具鏈產生完整相同函式；原始唯一 C 宣告、文字、型別與未證實語意仍未知。
本批沒有 ASM。16／32-bit 整數運算、signed／unsigned 比較、運算元求值順序、隨機呼叫與分支均由 C 重生。
原版覆寫權重、未用到的機率運算、重複運算與跳轉保持，沒有按推測修正玩法。三個字型候選與 `sub_1A4E4` 仍未匹配。

| 範圍 | 函式／bytes | 證明範圍 |
|---|---|---|
| 真實 compiler CODE 物件 | 53366／13650／1707／3679／5221／6038／5554／6558／5611 bytes | 前八個 C／OBJ 與前批逐 byte 相同；ATTACK.C／ATTACK.OBJ 供應新四函式 |
| Town compiler 來源 | 250 函式 | 167 純 C 55216、83 混合 46168 bytes，混合分列 39779 C／6389 ASM |
| Town LIB | 75 原始函式／9632 實際 bytes | 52 個完整 CODE 模組及 13 宣告零 bytes 保持 |
| Town 未還原 | 13 函式／11547 CODE bytes | 全部 6736 初始化 DATA 保持，原版 DATA 片段為 0 |
| 六支未還原 | 18 函式／15550 CODE bytes | Book 五函式 4003 CODE 與四支完整來源保持 |

九個真實 CODE 段各自小於 64 KiB。101384-byte 合併只作分析座標，每筆 PUBLIC／fixup／回填保留實際物件、段與偏移。
兩個空白案例的 32 項來源、OBJ、清冊、MAP、native MZ 與 packed 全檔逐 byte 相同。
104 反例全部拒絕，包含前批全部 98 項與新權重 C、攻擊 C 再編譯、漏傷害 C、偽造 ASM 分類、錯誤物件來源與實際 ATTACK PUBLIC 修改；23 項 OMF 修改保持合法 checksum。
前批 801／800 長度反例與兩批失敗紀錄保持獨立核對。兩份 reader 核 CODE／DATA、真實物件來源、fixup／回填、129312-byte 映像與 2281 有序重定位。
九個 DATA relocation、15 個 compiler 指標 fixup 與 56 個獨立 compiler 對齊零 bytes 保持。
LINK4.07／EXEMOD4.02／EXEPACK4.06 兩次原生重生 133679-byte 核心，再生成 81 個 1Ah。
完整 133760-byte KOLTOWN.EXE 的 SHA-256 兩次均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

本機[新函式與範圍](../verification/local/matching-c32-20261010/new-functions.json)、
[實際新編譯來源](../verification/local/matching-c32-20261010/new-source.c)、
[完整來源閱讀視圖](../verification/local/matching-c32-20261010/town-source.c)、
[來源清冊](../verification/local/matching-c32-20261010/source-ledger.json)、
[重編與反例](../verification/local/matching-c32-20261010/checkpoint-review.json)、
[收據](../verification/local/matching-c32-20261010/receipt.json)、
[來源分類器](../verification/local/matching-c32-20261010/source_categories.py)、
[真實物件讀取器](../verification/local/matching-c32-20261010/compiler_objects.py)、
[編譯單元清單](../verification/local/matching-c32-20261010/seed-compiler-units.json)、
[唯讀核對器](../verification/local/matching-c32-20261010/verify.py)及
[保存清冊](../verification/local/matching-c32-20261010/retention.json)為本批入口。
累計 221 純 C 64868、378 LIB 44396、181 混合 74354 bytes，函式外 271 C／258 ASM 另列。
342 份正式輸入、全部 DATA、Book 五函式與遊戲對拍暫停保持；完整 matching 尚未完成。

## 第三十一批來源還原紀錄

2026-10-10 第三十一批新增四個完整函式，共 6558 bytes，分列 6553 C／5 ASM；2182 條原始指令全部匹配。
三個純 C 共 5707 bytes，一個混合共 851 bytes。全部 242 個既有 compiler 函式、52 個 LIB CODE 模組與初始化 DATA 保持。

| 原始 IDA 名稱 | IDA linear | 解壓檔案偏移 | bytes | C／ASM | 已核對範圍 |
|---|---|---|---|---|---|
| `sub_22EFC` | `0x22EFC` | 86780 | 851 | 846／5 | 完整戰鬥文字／圖像顯示、角色與清單；DS 讀取另列 |
| `sub_2324F` | `0x2324F` | 87631 | 1001 | 1001／0 | 狀態數值、32-bit 比例運算、夾限與命令 glyph |
| `sub_2374C` | `0x2374C` | 88908 | 2533 | 2533／0 | 完整戰鬥選擇、權重、方向搜尋與重試 |
| `sub_24131` | `0x24131` | 91441 | 2173 | 2173／0 | 完整鍵盤／滑鼠目標輸入、夾限、狀態更新與返回 |

功能名稱只作導航。已證實的是固定工具鏈產生完整相同函式；原始唯一 C 宣告、文字、型別與未證實語意仍未知。
唯一 ASM 為候選編譯前已明列的 IDA linear 半開範圍 `[0x230D2,0x230D7)`；沒有用 ASM 補普通 C 差異。
128-byte 視圖、unsigned offset 表、post-condition 迴圈與 32-bit 整數運算均由 C 重生。原版重複判定、跳過的程式區段與重試分支保持。
方向重試回查完整 IDA 匯出及 raw bytes：`0x2406A` 的 `8b46fa8946fc` 先恢復基準方向，`0x24070` 的 `8346fc01` 再加一。先前截斷輸出的推測未進核准來源。
候選的缺項宣告與 QuickC 近指標陣列宣告語法已修正，同工具從空白案例重跑；最後四函式無警告。三個字型候選與 `sub_1A4E4` 仍未匹配。

| 範圍 | 函式／bytes | 證明範圍 |
|---|---|---|
| 真實 compiler CODE 物件 | 53366／13650／1707／3679／5221／6038／5554／6558 bytes | 前七個 C／OBJ 與前批逐 byte 相同；DECIDE.C／DECIDE.OBJ 供應新四函式 |
| Town compiler 來源 | 246 函式 | 163 純 C 49605、83 混合 46168 bytes，混合分列 39779 C／6389 ASM |
| Town LIB | 75 原始函式／9632 實際 bytes | 52 個完整 CODE 模組及 13 宣告零 bytes 保持 |
| Town 未還原 | 17 函式／17158 CODE bytes | 全部 6736 初始化 DATA 保持，原版 DATA 片段為 0 |
| 六支未還原 | 22 函式／21161 CODE bytes | Book 五函式 4003 CODE 與四支完整來源保持 |

八個真實 CODE 段各自小於 64 KiB。95773-byte 合併只作分析座標，每筆 PUBLIC／fixup／回填保留實際物件、段與偏移。
兩個空白案例的 30 項來源、OBJ、清冊、MAP、native MZ 與 packed 全檔逐 byte 相同。
98 反例全部拒絕，包含前批全部 92 項及新顯示 C、戰鬥選擇 C、DS ASM 再編譯、漏目標輸入 C、錯 DS 範圍與實際 DECIDE PUBLIC 修改；22 項 OMF 修改保持合法 checksum。
前批 801／800 長度反例與兩批失敗紀錄保持獨立核對。兩份 reader 核 CODE／DATA、真實物件來源、fixup／回填、129312-byte 映像與 2281 有序重定位。
九個 DATA relocation、15 個 compiler 指標 fixup 與 56 個獨立 compiler 對齊零 bytes 保持。
LINK4.07／EXEMOD4.02／EXEPACK4.06 兩次原生重生 133679-byte 核心，再生成 81 個 1Ah。
完整 133760-byte KOLTOWN.EXE 的 SHA-256 兩次均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

本機[新函式與範圍](../verification/local/matching-c31-20261010/new-functions.json)、
[實際新編譯來源](../verification/local/matching-c31-20261010/new-source.c)、
[完整來源閱讀視圖](../verification/local/matching-c31-20261010/town-source.c)、
[來源清冊](../verification/local/matching-c31-20261010/source-ledger.json)、
[重編與反例](../verification/local/matching-c31-20261010/checkpoint-review.json)、
[收據](../verification/local/matching-c31-20261010/receipt.json)、
[來源分類器](../verification/local/matching-c31-20261010/source_categories.py)、
[真實物件讀取器](../verification/local/matching-c31-20261010/compiler_objects.py)、
[編譯單元清單](../verification/local/matching-c31-20261010/seed-compiler-units.json)、
[唯讀核對器](../verification/local/matching-c31-20261010/verify.py)及
[保存清冊](../verification/local/matching-c31-20261010/retention.json)為本批入口。
累計 217 純 C 59257、378 LIB 44396、181 混合 74354 bytes，函式外 271 C／258 ASM 另列。
342 份正式輸入、全部 DATA、Book 五函式與遊戲對拍暫停保持；完整 matching 尚未完成。

## 第三十批來源還原紀錄

2026-10-10 第三十批新增四個完整函式，共 5554 bytes，分列 5539 C／15 ASM；1864 條原始指令全部匹配。
三個純 C 共 4530 bytes，一個混合共 1024 bytes。全部 238 個既有 compiler 函式、52 個 LIB CODE 模組與初始化 DATA 保持。

| 原始 IDA 名稱 | IDA linear | 解壓檔案偏移 | bytes | C／ASM | 已核對範圍 |
|---|---|---|---|---|---|
| `sub_18350` | `0x18350` | 42832 | 1022 | 1022／0 | 完整欄位文字、數值與價格彙總 |
| `sub_18F3A` | `0x18F3A` | 45882 | 1903 | 1903／0 | 完整分層圖像、表格座標與繪製調度 |
| `sub_223C0` | `0x223C0` | 83904 | 1024 | 1009／15 | 九乘九格點、角色／物件狀態與呈現；ES／SI 讀取另列 |
| `sub_228B7` | `0x228B7` | 85175 | 1605 | 1605／0 | 完整戰鬥命令選單、輸入、目標與返回控制 |

功能名稱只作導航。已證實的是固定工具鏈產生完整相同函式；原始唯一 C 宣告、文字、型別與未證實語意仍未知。
唯一 ASM 為候選編譯前已明列的 IDA linear 半開範圍 `[0x2254E,0x2255D)`；沒有用 ASM 補普通 C 差異。
欄位／圖像函式的二維、三維 byte 視圖、word 視圖、unsigned 運算與加法求值順序均由 C 重生。候選的 `break` 分支已按原始跳轉修正。
候選工具最初的前批路徑多一層與兩個缺項宣告已修正，同 image／同命令從空白案例重跑；最後四函式無警告。
原版夾限、重複運算、狀態回填與取消分支保持。三個字型候選與 `sub_1A4E4` 仍未匹配。

| 範圍 | 函式／bytes | 證明範圍 |
|---|---|---|
| 真實 compiler CODE 物件 | 53366／13650／1707／3679／5221／6038／5554 bytes | 前六個 C／OBJ 與前批逐 byte 相同；TACTIC.C／TACTIC.OBJ 供應新四函式 |
| Town compiler 來源 | 242 函式 | 160 純 C 43898、82 混合 45317 bytes，混合分列 38933 C／6384 ASM |
| Town LIB | 75 原始函式／9632 實際 bytes | 52 個完整 CODE 模組及 13 宣告零 bytes 保持 |
| Town 未還原 | 21 函式／23716 CODE bytes | 全部 6736 初始化 DATA 保持，原版 DATA 片段為 0 |
| 六支未還原 | 26 函式／27719 CODE bytes | Book 五函式 4003 CODE 與四支完整來源保持 |

七個真實 CODE 段各自小於 64 KiB。89215-byte 合併只作分析座標，每筆 PUBLIC／fixup／回填保留實際物件、段與偏移。
兩個空白案例的 28 項來源、OBJ、清冊、MAP、native MZ 與 packed 全檔逐 byte 相同。
92 反例全部拒絕，包含前批全部 86 項及新欄位 C、命令 C、格點 ASM 再編譯、漏圖像 C、錯 ES／SI 範圍與實際 TACTIC PUBLIC 修改；21 項 OMF 修改保持合法 checksum。
前批 801／800 長度反例與兩批失敗紀錄保持獨立核對。兩份 reader 核 CODE／DATA、真實物件來源、fixup／回填、129312-byte 映像與 2281 有序重定位。
九個 DATA relocation、15 個 compiler 指標 fixup 與 56 個獨立 compiler 對齊零 bytes 保持。
LINK4.07／EXEMOD4.02／EXEPACK4.06 兩次原生重生 133679-byte 核心，再生成 81 個 1Ah。
完整 133760-byte KOLTOWN.EXE 的 SHA-256 兩次均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

本機[新函式與範圍](../verification/local/matching-c30-20261010/new-functions.json)、
[實際新編譯來源](../verification/local/matching-c30-20261010/new-source.c)、
[完整來源閱讀視圖](../verification/local/matching-c30-20261010/town-source.c)、
[來源清冊](../verification/local/matching-c30-20261010/source-ledger.json)、
[重編與反例](../verification/local/matching-c30-20261010/checkpoint-review.json)、
[收據](../verification/local/matching-c30-20261010/receipt.json)、
[來源分類器](../verification/local/matching-c30-20261010/source_categories.py)、
[真實物件讀取器](../verification/local/matching-c30-20261010/compiler_objects.py)、
[編譯單元清單](../verification/local/matching-c30-20261010/seed-compiler-units.json)、
[唯讀核對器](../verification/local/matching-c30-20261010/verify.py)及
[保存清冊](../verification/local/matching-c30-20261010/retention.json)為本批入口。
累計 214 純 C 53550、378 LIB 44396、180 混合 73503 bytes，函式外 271 C／258 ASM 另列。
342 份正式輸入、全部 DATA、Book 五函式與遊戲對拍暫停保持；完整 matching 尚未完成。

## 第二十九批來源還原紀錄

2026-10-10 第二十九批新增四個完整函式，共 6038 bytes，分列 5911 C／127 ASM；2023 條原始指令全部匹配。
兩個純 C 共 3966 bytes，兩個混合共 2072 bytes。全部 234 個既有 compiler 函式、52 個 LIB CODE 模組及初始化 DATA 保持。

| 原始 IDA 名稱 | IDA linear | 解壓檔案偏移 | bytes | C／ASM | 已核對範圍 |
|---|---|---|---|---|---|
| `sub_16239` | `0x16239` | 34361 | 1171 | 1166／5 | 完整角色摘要、文字置中、表格與數值公式；DS 讀取另列 |
| `sub_16986` | `0x16986` | 36230 | 901 | 779／122 | 完整物品選單、重啟、明細與返回；三個 ES 區段另列 |
| `sub_17318` | `0x17318` | 38680 | 1480 | 1480／0 | 既有項目與可購目錄、價格、確認及記錄搬移 |
| `sub_178E0` | `0x178E0` | 40160 | 2486 | 2486／0 | 完整欄位編輯、鍵盤／滑鼠選擇、費用與回填 |

功能名稱只作導航。已證實的是固定工具鏈產生完整相同函式；原始唯一 C 宣告、文字、型別及未證實語意仍未知。
四個 DS／ES ASM 範圍在編譯前已明列，沒有用 ASM 補普通 C 差異。word 視圖修正 signed 傳參，混合 signed／unsigned 索引亦逐 byte 核對。
未初始化的多餘實參與原始壓棧／清理保留。`sub_16986` 的 IDA linear `0x16B7F`／`0x16A95` 分別傳七個／五個實參，清理 14／10 bytes；不據此推定唯一 prototype。
候選最初 C2086 為預設 int 宣告與新 void 定義衝突，補明示宣告後同工具重跑，最後無警告。三個字型候選與 `sub_1A4E4` 仍未匹配。

| 範圍 | 函式／bytes | 證明範圍 |
|---|---|---|
| 真實 compiler CODE 物件 | 53366／13650／1707／3679／5221／6038 bytes | 前五個 C／OBJ 與前批逐 byte 相同；CHAR.C／CHAR.OBJ 供應新四函式 |
| Town compiler 來源 | 238 函式 | 157 純 C 39368、81 混合 44293 bytes，混合分列 37924 C／6369 ASM |
| Town LIB | 75 原始函式／9632 實際 bytes | 52 個完整 CODE 模組及 13 宣告零 bytes 保持 |
| Town 未還原 | 25 函式／29270 CODE bytes | 全部 6736 初始化 DATA 保持，原版 DATA 片段為 0 |
| 六支未還原 | 30 函式／33273 CODE bytes | Book 五函式 4003 CODE 及四支完整來源保持 |

六個真實 CODE 段各自小於 64 KiB。83661-byte 合併只作分析座標，每筆 PUBLIC／fixup／回填保留實際物件、段與偏移。
兩個空白案例的 26 項來源、OBJ、清冊、MAP、native MZ 與 packed 全檔逐 byte 相同。
86 反例全部拒絕，包含前批全部 80 項及新摘要 C、欄位編輯 C、選單 ASM 再編譯、漏純 C、錯 DS 範圍與實際 CHAR PUBLIC 修改；20 項 OMF 修改保持合法 checksum。
前批 801／800 長度反例與兩批失敗紀錄保持獨立核對。兩份 reader 核 CODE／DATA、真實物件來源、fixup／回填、129312-byte 映像及 2281 有序重定位。
九個 DATA relocation、15 個 compiler 指標 fixup 與 56 個獨立 compiler 對齊零 bytes 保持。
LINK4.07／EXEMOD4.02／EXEPACK4.06 兩次原生重生 133679-byte 核心，再生成 81 個 1Ah。
完整 133760-byte KOLTOWN.EXE 的 SHA-256 兩次均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

本機[新函式與範圍](../verification/local/matching-c29-20261010/new-functions.json)、
[實際新編譯來源](../verification/local/matching-c29-20261010/new-source.c)、
[完整來源閱讀視圖](../verification/local/matching-c29-20261010/town-source.c)、
[來源清冊](../verification/local/matching-c29-20261010/source-ledger.json)、
[重編與反例](../verification/local/matching-c29-20261010/checkpoint-review.json)、
[收據](../verification/local/matching-c29-20261010/receipt.json)、
[來源分類器](../verification/local/matching-c29-20261010/source_categories.py)、
[真實物件讀取器](../verification/local/matching-c29-20261010/compiler_objects.py)、
[編譯單元清單](../verification/local/matching-c29-20261010/seed-compiler-units.json)、
[唯讀核對器](../verification/local/matching-c29-20261010/verify.py)及
[保存清冊](../verification/local/matching-c29-20261010/retention.json)為本批入口。
累計 211 純 C 49020、378 LIB 44396、179 混合 72479 bytes，函式外 271 C／258 ASM 另列。
342 份正式輸入、全部 DATA、Book 五函式與遊戲對拍暫停保持；完整 matching 尚未完成。

## 第二十八批來源還原紀錄

2026-10-10 第二十八批新增四個完整函式，共 5221 bytes，分列 5033 C／188 ASM；1756 條原始指令全部匹配。
包含兩個純 C 2654 bytes 與兩個混合來源 2567 bytes。全部 230 個既有 compiler 函式、52 個 LIB CODE 模組及初始化 DATA 保持。

| 原始 IDA 名稱 | IDA linear | 解壓檔案偏移 | bytes | C／ASM | 已核對範圍 |
|---|---|---|---|---|---|
| `sub_147DD` | `0x147DD` | 27613 | 1266 | 1202／64 | 名冊比對、選擇、記錄載入與狀態控制；四個 DOS／ES 區段另列 |
| `sub_14DB1` | `0x14DB1` | 29105 | 1301 | 1177／124 | 金額檢查、保存槽搜尋、記錄回填與多檔寫入；九個 DOS／ES 區段另列 |
| `sub_153FF` | `0x153FF` | 30719 | 801 | 801／0 | 完整物品明細文字、旗標、數值與分支 |
| `sub_15ABF` | `0x15ABF` | 32447 | 1853 | 1853／0 | 完整角色資料頁、比例條、清單、命令與返回控制 |

上述功能名稱只作導航。已證實的是固定工具鏈產生完整相同函式；原始唯一 C 宣告、文字、型別及未證實語意仍未知。
13 個 ASM 範圍在候選編譯前已按原始 DOS 中斷與 ES 記憶體搬移明列，沒有以 ASM 補普通 C 差異。
比對、金額與狀態控制由 C 產生。值轉型與 word 視圖分開，保留原始求值順序、unsigned 除法／右移與 signed 傳參。
最初 ASM 別名未宣告的 C2094 已改用既有 pool 綁定；候選各自攜帶固定錯誤資源，保留可讀診斷。最後四函式無警告。
原版重複讀寫、未用區域變數、磁碟還原與取消分支保持。先前三個字型候選與 `sub_1A4E4` 仍未匹配。

| 範圍 | 函式／bytes | 證明範圍 |
|---|---|---|
| 真實 compiler CODE 物件 | 53366／13650／1707／3679／5221 bytes | BULK、BATTLE、CONTROL、MENU 的 C／OBJ 與前批逐 byte 相同；ROSTER.C／ROSTER.OBJ 供應新四函式 |
| Town compiler 來源 | 234 函式 | 155 純 C 35402、79 混合 42221 bytes，混合分列 35979 C／6242 ASM |
| Town LIB | 75 原始函式／9632 實際 bytes | 52 個完整 CODE 模組及 13 宣告零 bytes 保持 |
| Town 未還原 | 29 函式／35308 CODE bytes | 全部 6736 初始化 DATA 保持，原版 DATA 片段為 0 |
| 六支未還原 | 34 函式／39311 CODE bytes | Book 五函式 4003 CODE 保持，四支完整來源保持 |

五個真實 CODE 段各自小於 64 KiB。77623-byte 合併只作分析座標，每筆 PUBLIC／fixup／回填保留實際物件、段與偏移。
兩個空白案例的 24 項來源、OBJ、清冊、MAP、native MZ 與 packed 全檔逐 byte 相同。
80 反例全部拒絕，包含全部前批 74 項，以及新名冊 C、物品明細 C、記錄搬移 ASM 的實際再編譯、漏資料頁、錯誤 DOS 範圍與實際 ROSTER PUBLIC 偏移修改。
物品明細的 128 改為 127，真實 compiler 函式由 801 縮為 800 bytes，已在原生轉接長度閘門拒絕。腳本明確登記這個預期階段，再從空白案例重跑；原生正例未改，失敗腳本與 73 個案例保存。封存核對器另重讀兩個真實物件確認 801／800。
19 項 OMF 修改保持合法 checksum。兩份獨立 reader 核 CODE／DATA、真實物件來源、fixup／回填、129312-byte 映像及 2281 筆有序重定位，前批腳本失敗記錄也保留獨立核對。
九個 DATA relocation、15 個 compiler 指標 fixup、56 個獨立 compiler 對齊零 bytes 均保持。
LINK4.07、EXEMOD4.02、EXEPACK4.06 兩次原生重生 133679-byte 核心，再生成 81 個 1Ah。
完整 133760-byte `KOLTOWN.EXE` 的 SHA-256 兩次均為
`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

本機[新函式與範圍](../verification/local/matching-c28-20261010/new-functions.json)、
[實際新編譯來源](../verification/local/matching-c28-20261010/new-source.c)、
[完整來源閱讀視圖](../verification/local/matching-c28-20261010/town-source.c)、
[來源清冊](../verification/local/matching-c28-20261010/source-ledger.json)、
[重編與反例](../verification/local/matching-c28-20261010/checkpoint-review.json)、
[收據](../verification/local/matching-c28-20261010/receipt.json)、
[來源分類器](../verification/local/matching-c28-20261010/source_categories.py)、
[真實物件讀取器](../verification/local/matching-c28-20261010/compiler_objects.py)、
[編譯單元清單](../verification/local/matching-c28-20261010/seed-compiler-units.json)、
[驗證階段勘誤](../verification/local/matching-c28-20261010/checkpoint-failure.json)、
[唯讀核對器](../verification/local/matching-c28-20261010/verify.py)及
[保存清冊](../verification/local/matching-c28-20261010/retention.json)為本批入口。
累計 209 純 C 45054、378 LIB 44396、177 混合 70407 bytes，函式外 271 C／258 ASM 另列。
342 份正式輸入、全部 DATA、Book 五函式及遊戲對拍暫停保持；完整 matching 尚未完成。

## 第二十七批來源還原紀錄

2026-10-10 第二十七批新增四個完整純 C 函式，共 3679 bytes，1236 條原始指令全部匹配，沒有新增 ASM。
全部 226 個既有 compiler 函式、52 個 LIB CODE 模組及初始化 DATA 保持。

| 原始 IDA 名稱 | IDA linear | 解壓檔案偏移 | bytes | 原始指令 | 已核對範圍 |
|---|---|---|---|---|---|
| `sub_16D0B` | `0x16D0B` | 37131 | 667 | 231 | 選項重繪、17-byte 表格、offset 表及記錄欄位讀取 |
| `sub_196A9` | `0x196A9` | 47785 | 990 | 331 | 文字分頁、鍵盤／滑鼠選擇、角色與模式切換、返回值 |
| `sub_18857` | `0x18857` | 44119 | 1150 | 374 | 完整物品命令 switch、轉交／搬移／清除及取消分支 |
| `sub_1A4FE` | `0x1A4FE` | 51454 | 872 | 300 | 預載來源選擇、檔案開啟／讀取／重試、遠指標與圖像調度 |

上述功能名稱只作導航。已證實的是固定工具鏈產生完整相同函式；原始唯一 C 宣告、文字、型別及未證實語意仍未知。
候選差異由二維 17-byte 表格、unsigned offset 表、stack word 視圖與求值順序閉合。
遠指標以明示兩個 word 組成，保留原始近／遠傳參；沒有貼回指令、使用 `_emit` 或以 ASM 補普通 C 差異。
原版比較式、未用區域變數、重複顯示與重試分支保持，沒有順便修正遊戲行為。
先前未匹配的三個字型候選與 `sub_1A4E4` 遠指標葉函式仍開放。

| 範圍 | 函式／bytes | 證明範圍 |
|---|---|---|
| 真實 compiler CODE 物件 | 53366／13650／1707／3679 bytes | BULK、BATTLE、CONTROL 的 C／OBJ 與前批逐 byte 相同；MENU.C／MENU.OBJ 供應新四函式 |
| Town compiler 來源 | 230 函式 | 153 純 C 32748、77 混合 39654 bytes，混合分列 33600 C／6054 ASM |
| Town LIB | 75 原始函式／9632 實際 bytes | 52 個完整 CODE 模組及 13 宣告零 bytes 保持 |
| Town 未還原 | 33 函式／40529 CODE bytes | 全部 6736 初始化 DATA 保持，原版 DATA 片段為 0 |
| 六支未還原 | 38 函式／44532 CODE bytes | Book 五函式 4003 CODE 保持，四支完整來源保持 |

四個真實 CODE 段各自小於 64 KiB。72402-byte 合併只作分析座標，每筆 PUBLIC／fixup／回填保留實際物件、段與偏移。
兩個空白案例的 22 項來源、OBJ、清冊、MAP、native MZ 與 packed 全檔逐 byte 相同。
74 反例全部拒絕，包含全部前批 68 項，以及新選擇 C、圖像 C 的實際再編譯、漏新純 C、偽造 ASM 分類、錯誤來源物件與實際 MENU PUBLIC 偏移修改。
首次反例腳本因純 C 清冊省略可選欄位而報 KeyError。以函式 size 建立偽造欄位後，同 image／同命令從空白案例重跑；原生重建未改，失敗腳本與 69 個案例另保存。
18 項 OMF 修改保持合法 checksum。兩份獨立 reader 核 CODE／DATA、真實物件來源、fixup／回填、129312-byte 映像及 2281 筆有序重定位。
九個 DATA relocation、15 個 compiler 指標 fixup、56 個獨立 compiler 對齊零 bytes 均保持。
LINK4.07、EXEMOD4.02、EXEPACK4.06 兩次原生重生 133679-byte 核心，再生成 81 個 1Ah。
完整 133760-byte `KOLTOWN.EXE` 的 SHA-256 兩次均為
`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

本機[新函式與範圍](../verification/local/matching-c27-20261010/new-functions.json)、
[實際新編譯來源](../verification/local/matching-c27-20261010/new-source.c)、
[完整來源閱讀視圖](../verification/local/matching-c27-20261010/town-source.c)、
[來源清冊](../verification/local/matching-c27-20261010/source-ledger.json)、
[重編與反例](../verification/local/matching-c27-20261010/checkpoint-review.json)、
[收據](../verification/local/matching-c27-20261010/receipt.json)、
[來源分類器](../verification/local/matching-c27-20261010/source_categories.py)、
[真實物件讀取器](../verification/local/matching-c27-20261010/compiler_objects.py)、
[編譯單元清單](../verification/local/matching-c27-20261010/seed-compiler-units.json)、
[驗證腳本勘誤](../verification/local/matching-c27-20261010/checkpoint-failure.json)、
[唯讀核對器](../verification/local/matching-c27-20261010/verify.py)及
[保存清冊](../verification/local/matching-c27-20261010/retention.json)為本批入口。
累計 207 純 C 42400、378 LIB 44396、175 混合 67840 bytes，函式外 271 C／258 ASM 另列。
342 份正式輸入、全部 DATA、Book 五函式及遊戲對拍暫停保持；完整 matching 尚未完成。

## 第二十六批來源還原紀錄

2026-10-10 第二十六批新增四個完整函式，共 1707 bytes，分列 1012 C／695 ASM。
全部 222 個既有 compiler 函式、52 個 LIB CODE 模組及初始化 DATA 保持。

| 原始 IDA 名稱 | IDA linear | 解壓檔案偏移 | bytes | C／ASM | 已核對範圍 |
|---|---|---|---|---|---|
| `sub_19B29` | `0x19B29` | 48937 | 454 | 209／245 | 圖像平面轉換；普通 C 遠記憶體讀寫迴圈與兩個暫存器 codec 區段分列 |
| `sub_19E9D` | `0x19E9D` | 49821 | 433 | 253／180 | 顯示模式與矩形搬移調度；三個硬體區段分列 |
| `sub_1B94F` | `0x1B94F` | 56655 | 273 | 15／258 | 完整五位元文字 codec 暫存器區段與 C 框架分列 |
| `sub_1CBA9` | `0x1CBA9` | 61353 | 547 | 535／12 | 磁碟開啟、重試提示、狀態保存／還原；切換磁碟的 DOS 中斷區段另列 |

上述功能名稱只作導航。已證實的是完整原始函式、固定來源與原生建置的 byte equality；原始唯一 C 宣告、ASM 來源形式及未證實語意仍未知。
`sub_19B29` 的 unsigned 除法、signed 迴圈比較與明示 far pointer 視圖由 C 產生。
`sub_1CBA9` 的 13 個區域 word、重試跳轉、提示與狀態還原由 C 產生；沒有用 ASM 補普通 C 差異。
三個字型候選仍因普通 C 減法的指令選擇不符而保持未還原。`sub_1A4E4` 的 `_segment／_based` 候選報 C2061，語法支援未證實，該函式也保持未還原。

| 範圍 | 函式／bytes | 證明範圍 |
|---|---|---|
| 真實 compiler CODE 物件 | 53366／13650／1707 bytes | BULK.C／BULK.OBJ、BATTLE.C／BATTLE.OBJ 與前批逐 byte 相同；CONTROL.C／CONTROL.OBJ 供應新四函式 |
| Town compiler 來源 | 226 函式 | 149 純 C 29069、77 混合 39654 bytes，混合分列 33600 C／6054 ASM |
| Town LIB | 75 原始函式／9632 實際 bytes | 52 個完整 CODE 模組及 13 宣告零 bytes 保持 |
| Town 未還原 | 37 函式／44208 CODE bytes | 全部 6736 初始化 DATA 保持，原版 DATA 片段為 0 |
| 六支未還原 | 42 函式／48211 CODE bytes | Book 五函式 4003 CODE 保持，四支完整來源保持 |

三個真實 CODE 段各自小於 64 KiB。68723-byte 合併只作分析座標，每筆 PUBLIC／fixup／回填保留實際物件、段與偏移。
兩個空白案例的 20 項來源、OBJ、清冊、MAP、native MZ 與 packed 全檔逐 byte 相同。
68 反例全部拒絕，包含全部前批 62 項，以及新轉換 C、磁碟 C、codec ASM 的實際再編譯、漏磁碟控制、錯誤硬體區段邊界與實際 CONTROL PUBLIC 偏移修改。
17 項 OMF 修改保持合法 checksum。兩份獨立 reader 核 CODE／DATA、真實物件來源、fixup／回填、129312-byte 映像及 2281 筆有序重定位。
九個 DATA relocation、15 個 compiler 指標 fixup、56 個獨立 compiler 對齊零 bytes 均保持。
LINK4.07、EXEMOD4.02、EXEPACK4.06 兩次原生重生 133679-byte 核心，再生成 81 個 1Ah。
完整 133760-byte `KOLTOWN.EXE` 的 SHA-256 兩次均為
`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

本機[新函式與範圍](../verification/local/matching-c26-20261010/new-functions.json)、
[實際新編譯來源](../verification/local/matching-c26-20261010/new-source.c)、
[完整來源閱讀視圖](../verification/local/matching-c26-20261010/town-source.c)、
[來源清冊](../verification/local/matching-c26-20261010/source-ledger.json)、
[重編與反例](../verification/local/matching-c26-20261010/checkpoint-review.json)、
[收據](../verification/local/matching-c26-20261010/receipt.json)、
[來源分類器](../verification/local/matching-c26-20261010/source_categories.py)、
[真實物件讀取器](../verification/local/matching-c26-20261010/compiler_objects.py)、
[編譯單元清單](../verification/local/matching-c26-20261010/seed-compiler-units.json)、
[唯讀核對器](../verification/local/matching-c26-20261010/verify.py)及
[保存清冊](../verification/local/matching-c26-20261010/retention.json)為本批入口。
累計 203 純 C 38721、378 LIB 44396、175 混合 67840 bytes，函式外 271 C／258 ASM 另列。
342 份正式輸入、全部 DATA、Book 五函式及遊戲對拍暫停保持；完整 matching 尚未完成。

## 第二十五批來源還原紀錄

2026-10-10 第二十五批新增完整戰鬥主控制 `sub_1E3A8`，共 13650 bytes，分列 13634 C／16 ASM。
4389 條原始指令全部匹配；全部 221 個既有 compiler 函式、52 個 LIB CODE 模組及初始化 DATA 保持。

| 範圍 | 函式／bytes | 證明範圍 |
|---|---|---|
| 新增 `sub_1E3A8` | 13650 bytes，13634 C／16 ASM | 完整初始化、角色配置、行動規劃、結算及退出分支；單一 ES／SI 地圖讀取另列 |
| 真實 compiler CODE 物件 | 53366／13650 bytes | BULK.OBJ 保持前批，BATTLE.OBJ 由新輸入重編；各段小於 64 KiB |
| Town compiler 來源 | 222 函式 | 149 純 C 29069、73 混合 37947 bytes，混合分列 32588 C／5359 ASM |
| Town LIB | 75 原始函式／9632 實際 bytes | 52 個完整 CODE 模組及 13 宣告零 bytes 保持 |
| Town 未還原 | 41 函式／45915 CODE bytes | 全部 6736 初始化 DATA 保持，原版 DATA 片段為 0 |
| 六支未還原 | 46 函式／49918 CODE bytes | Book 五函式 4003 CODE 保持，四支完整來源保持 |

原始函式為 IDA linear `0x1E3A8`，解壓檔案偏移為 `67496`，兩種位址基準分開保存。
唯一 ASM 為 IDA linear 半開範圍 `[0x1EB0D,0x1EB1D)` 的 16-byte ES／SI 地圖讀取。
其他遠指標陣列讀寫、50-byte 角色及 128-byte 記錄視圖、控制與算式皆由 C 重生，沒有用 ASM 補普通 C 差異。
原版重複判定、未保存的計算結果、暫存及取消／重試分支保持。
功能名稱只作導航；已證實的是固定工具鏈的完整函式 byte equality，原始唯一宣告、原始 ASM 寫法與未證實語意仍未知。

加入新函式後，單一 CODE 輸入的預計長度超過 64 KiB。實際 QuickC 編譯拒絕 C1028，
[固定版本的廠商錯誤資源](https://raw.githubusercontent.com/davidly/dos_compilers/da3f15ce4912893a0dd9f8349dab29116a53ca59/Microsoft%20QuickC%20v2/C1.ERR)
載明該上限；資源 SHA-256 為 `96d46d83a797b372f605abab019e72c9a50f85ac84003401a01f3fc266dca019`。
沿同 compiler、`/AM /Od /Gs /Zl /G0` 與 image，把 CODE 分為兩個真實輸入單元。BULK.C 與 BULK.OBJ 逐 byte 保持前批。
沒有合成超過 64 KiB 的 compiler OMF 段；67016-byte 合併只作分析座標，每筆 PUBLIC／fixup／回填保留實際物件、段與偏移。
適配器及獨立來源核對器用真實物件重讀，空間資訊清冊升為版本 2，另記實體原位與分析偏移。

兩個空白案例的 18 項來源、OBJ、清冊、MAP、native MZ 與 packed 全檔逐 byte 相同。
62 反例全部拒絕，包含前批 55 項、新戰鬥 C／ASM 實際再編譯、漏 CODE 單元、錯誤物件歸屬、過期單元順序、真實 PUBLIC 偏移修改及單段 64 KiB 邊界。
16 項 OMF 修改保持合法 checksum。兩份獨立 reader 核 CODE／DATA、真實物件來源、fixup／回填、129312-byte 映像及 2281 筆有序重定位。
九個 DATA relocation、15 個 compiler 指標 fixup、56 個獨立 compiler 對齊零 bytes 均保持。
LINK4.07、EXEMOD4.02、EXEPACK4.06 兩次原生重生 133679-byte 核心，再生成 81 個 1Ah。
完整 133760-byte `KOLTOWN.EXE` 的 SHA-256 兩次均為
`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。

本機[新函式與範圍](../verification/local/matching-c25-20261010/new-functions.json)、
[實際新編譯來源](../verification/local/matching-c25-20261010/new-source.c)、
[完整來源閱讀視圖](../verification/local/matching-c25-20261010/town-source.c)、
[來源清冊](../verification/local/matching-c25-20261010/source-ledger.json)、
[重編與反例](../verification/local/matching-c25-20261010/checkpoint-review.json)、
[收據](../verification/local/matching-c25-20261010/receipt.json)、
[來源分類器](../verification/local/matching-c25-20261010/source_categories.py)、
[真實物件讀取器](../verification/local/matching-c25-20261010/compiler_objects.py)、
[編譯單元清單](../verification/local/matching-c25-20261010/seed-compiler-units.json)、
[唯讀核對器](../verification/local/matching-c25-20261010/verify.py)及
[保存清冊](../verification/local/matching-c25-20261010/retention.json)為本批入口。
累計 203 純 C 38721、378 LIB 44396、171 混合 66133 bytes，函式外 271 C／258 ASM 另列。
342 份正式輸入、全部 DATA、Book 五函式及遊戲對拍暫停保持；完整 matching 尚未完成。

## 工具與重生入口

第三十一批的 `full-build.py` 支援 246 個 compiler 函式、八個真實 CODE 物件、52 個 LIB CODE 模組與全部 DATA。
從[保存清冊](../verification/local/matching-c34-20261010/retention.json)恢復來源／工具，以合法本機 ZIP 重生原版；依 `fetch-tools.py` 重取固定工具，另恢復 C1.ERR／CL.ERR 與完整工具清冊。
來源唯讀、案例輸出可寫掛到 `/out`，執行 `python3 -B /out/full-build.py <新的空白案例名稱>`。
`compile-DATA.py` 依 `seed-compiler-units.json` 分別編譯 BULK／BATTLE／CONTROL／MENU／ROSTER／CHAR／TACTIC／DECIDE 八個 CODE 物件與 DATA／PTRS。
`compiler-objects.json` 記錄實際次序，`compiler_objects.py` 逐物件核界限與原位。`seed-CODE.c`／`town-source.c` 只供閱讀及分類，不作單一編譯輸入。
接續時保留全部 seed、57 個後續純 C、83 個混合分類、106 純 C 基準與完整 DATA 鏈；`bootstrap.py` 逐一恢復前批單元，`verify.py` 獨立唯讀核對。

第三十批的 `full-build.py` 支援 242 個 compiler 函式、七個真實 CODE 物件、52 個 LIB CODE 模組與全部 DATA。
從[保存清冊](../verification/local/matching-c30-20261010/retention.json)恢復來源／工具，以合法本機 ZIP 重生原版；依 `fetch-tools.py` 重取固定工具，另恢復 C1.ERR／CL.ERR 與完整工具清冊。
來源唯讀、案例輸出可寫掛到 `/out`，執行 `python3 -B /out/full-build.py <新的空白案例名稱>`。
`compile-DATA.py` 依 `seed-compiler-units.json` 分別編譯 BULK／BATTLE／CONTROL／MENU／ROSTER／CHAR／TACTIC 七個 CODE 物件與 DATA／PTRS。
`compiler-objects.json` 記錄實際次序，`compiler_objects.py` 逐物件核界限與原位。`seed-CODE.c`／`town-source.c` 只供閱讀及分類，不作單一編譯輸入。
接續時保留全部 seed、54 個後續純 C、82 個混合分類、106 純 C 基準與完整 DATA 鏈；`bootstrap.py` 逐一恢復前批單元，`verify.py` 獨立唯讀核對。

第二十九批的 `full-build.py` 支援 238 個 compiler 函式、六個真實 CODE 物件、52 個 LIB CODE 模組及全部 DATA。
從[保存清冊](../verification/local/matching-c29-20261010/retention.json)恢復來源／工具，以合法本機 ZIP 重生原版；依 `fetch-tools.py` 重取固定工具，另恢復 C1.ERR／CL.ERR 與完整工具清冊。
來源唯讀、案例輸出可寫掛到 `/out`，執行 `python3 -B /out/full-build.py <新的空白案例名稱>`。
`compile-DATA.py` 依 `seed-compiler-units.json` 分別編譯 BULK／BATTLE／CONTROL／MENU／ROSTER／CHAR 六個 CODE 物件與 DATA／PTRS。
`compiler-objects.json` 記錄實際次序，`compiler_objects.py` 逐物件核界限與原位。`seed-CODE.c`／`town-source.c` 只供閱讀及分類，不作單一編譯輸入。
接續時保留全部 seed、51 個後續純 C、81 個混合分類、106 純 C 基準與完整 DATA 鏈；`bootstrap.py` 逐一恢復前批單元，`verify.py` 獨立唯讀核對。

第二十八批的 `full-build.py` 支援 234 個 compiler 函式、五個真實 CODE 物件、52 個 LIB CODE 模組及全部 DATA。
從[保存清冊](../verification/local/matching-c28-20261010/retention.json)恢復來源／工具，以合法本機 ZIP 重生原版；固定工具依 `fetch-tools.py` 重取，錯誤資源另用 `fetch-error-resources.py`。
將來源唯讀、案例輸出可寫掛到 `/out`，執行 `python3 -B /out/full-build.py <新的空白案例名稱>`。
`compile-DATA.py` 依 `seed-compiler-units.json` 複製全部五個 CODE seed，分別編譯為 BULK.OBJ／BATTLE.OBJ／CONTROL.OBJ／MENU.OBJ／ROSTER.OBJ，另編 DATA／PTRS。
`compiler-objects.json` 記錄實際次序；`compiler_objects.py` 逐物件核界限與原位。`seed-CODE.c` 與 `town-source.c` 只供閱讀及分類，不作單一編譯輸入。
接續時保留全部 seed 單元、49 個後續純 C、79 個混合函式分類、106 純 C 基準與完整 DATA 鏈。`bootstrap.py` 逐一恢復前批單元，`verify.py` 提供獨立唯讀核對。

第二十七批的 `full-build.py` 支援 230 個 compiler 函式、四個真實 CODE 物件、52 個 LIB CODE 模組及全部 DATA。
從[保存清冊](../verification/local/matching-c27-20261010/retention.json)恢復來源／工具，以合法本機 ZIP 重生原版；固定工具依 `fetch-tools.py` 重取，錯誤資源另用 `fetch-error-resources.py`。
將來源唯讀、案例輸出可寫掛到 `/out`，執行 `python3 -B /out/full-build.py <新的空白案例名稱>`。
`compile-DATA.py` 依 `seed-compiler-units.json` 複製 CODE-BASE.c／CODE-BATTLE.c／CODE-CONTROL.c／CODE-MENU.c，分別編譯為 BULK.OBJ／BATTLE.OBJ／CONTROL.OBJ／MENU.OBJ，另編 DATA／PTRS。
`compiler-objects.json` 記錄實際次序；`compiler_objects.py` 逐物件核界限與原位。`seed-CODE.c` 與 `town-source.c` 只供閱讀及分類，不作單一編譯輸入。
接續時保留全部 seed 單元、47 個後續純 C、77 個混合函式分類、106 純 C 基準與完整 DATA 鏈。`bootstrap.py` 依前批 seed 清單恢復全部單元，`verify.py` 提供獨立唯讀核對。

第二十六批的 `full-build.py` 支援 226 個 compiler 函式、三個真實 CODE 物件、52 個 LIB CODE 模組及全部 DATA。
從[保存清冊](../verification/local/matching-c26-20261010/retention.json)恢復來源／工具，以合法本機 ZIP 重生原版；固定工具依 `fetch-tools.py` 重取，錯誤資源另用 `fetch-error-resources.py`。
將來源唯讀、案例輸出可寫掛到 `/out`，執行 `python3 -B /out/full-build.py <新的空白案例名稱>`。
`compile-DATA.py` 依 `seed-compiler-units.json` 複製 CODE-BASE.c／CODE-BATTLE.c／CODE-CONTROL.c，分別編譯為 BULK.OBJ／BATTLE.OBJ／CONTROL.OBJ，另編 DATA／PTRS。
`compiler-objects.json` 記錄實際輸入次序；`compiler_objects.py` 逐物件核界限與原位。`seed-CODE.c` 與 `town-source.c` 只供閱讀及分類，不作單一編譯輸入。
接續時保留全部 seed 單元、43 個後續純 C、77 個混合函式分類、106 純 C 基準與完整 DATA 鏈。`verify.py` 提供獨立唯讀核對。

第二十五批的 `full-build.py` 支援 222 個 compiler 函式、兩個真實 CODE 物件、52 個 LIB CODE 模組及全部 DATA。
從保存清冊恢復來源／工具，以合法本機 ZIP 重生原版；固定工具可依 `fetch-tools.py` 重取，錯誤資源另用 `fetch-error-resources.py`。
將來源唯讀、案例輸出可寫掛到 `/out`，執行 `python3 -B /out/full-build.py <新的空白案例名稱>`。
`compile-DATA.py` 依 `seed-compiler-units.json` 複製 CODE-BASE.c／CODE-BATTLE.c，分別編譯為 BULK.OBJ／BATTLE.OBJ，另編 DATA／PTRS。
`compiler-objects.json` 記錄案例的實際輸入次序；`compiler_objects.py` 逐物件核界限與原位，分析合併不形成 OMF 段。
`seed-CODE.c` 與 `town-source.c` 是合併閱讀／分類視圖，不作單一編譯輸入。接續來源時保留全部 seed 單元，`prepare.py` 已支援恢復每個真實 CODE 物件。
分類清單含 43 個後續純 C 與 73 個混合函式，106 純 C 分類基準保持。完整 DATA 鏈與 `verify.py` 保持獨立核對。


第二十四批重編入口為上述保存目錄的 `full-build.py`，支援 221 個 compiler 函式、52 個 LIB CODE 模組與全部 DATA。
從封存恢復來源與固定工具，原版 EXE 由合法本機 ZIP 重生；輸入唯讀、案例輸出可寫掛到 `/out`，
執行 `python3 -B /out/full-build.py <新的空白案例名稱>`。
`seed-CODE.c` 為完整 C／ASM 輸入；分類清單含 43 個後續純 C 與 72 個混合函式，106 純 C 分類基準保持。
`compile-DATA.py`、`materialize-DATA.py`、`data_proof.py` 保留完整 DATA 鏈；唯讀核對入口為 `verify.py`。


第二十三批重編入口為上述保存目錄的 `full-build.py`，支援 220 個 compiler 函式、52 個 LIB CODE 模組與全部 DATA。
從封存恢復來源與固定工具，原版 EXE 由合法本機 ZIP 重生；輸入唯讀、案例輸出可寫掛到 `/out`，
執行 `python3 -B /out/full-build.py <新的空白案例名稱>`。
`seed-CODE.c` 為完整 C／ASM 輸入；分類清單含 43 個後續純 C 與 71 個混合函式，106 純 C 分類基準保持。
`compile-DATA.py`、`materialize-DATA.py`、`data_proof.py` 保留完整 DATA 鏈；唯讀核對入口為 `verify.py`。


第二十二批重編入口為上述保存目錄的 `full-build.py`，支援 217 個 compiler 函式、52 個 LIB CODE 模組與全部 DATA。
從封存恢復來源與固定工具，原版 EXE 由合法本機 ZIP 重生；輸入唯讀、案例輸出可寫掛到 `/out`，
執行 `python3 -B /out/full-build.py <新的空白案例名稱>`。
`seed-CODE.c` 為完整 C／ASM 輸入；分類清單含 42 個後續純 C 與 69 個混合函式，106 純 C 分類基準保持。
`compile-DATA.py`、`materialize-DATA.py`、`data_proof.py` 保留完整 DATA 鏈；唯讀核對入口為 `verify.py`。


第二十一批重編入口為上述保存目錄的 `full-build.py`，支援 203 個 compiler 函式、52 個 LIB CODE 模組與全部 DATA。
從封存恢復來源與固定工具，原版 EXE 由合法本機 ZIP 重生；輸入唯讀、案例輸出可寫掛到 `/out`，
執行 `python3 -B /out/full-build.py <新的空白案例名稱>`。
`seed-CODE.c` 為完整 C／ASM 輸入；分類清單含 30 個後續純 C 與 67 個混合函式，106 純 C 分類基準保持。
`compile-DATA.py`、`materialize-DATA.py`、`data_proof.py` 保留完整 DATA 鏈；唯讀核對入口為 `verify.py`。


第二十批重編入口為上述保存目錄的 `full-build.py`，支援 189 個 compiler 函式、52 個 LIB CODE 模組與全部 DATA。
從封存恢復來源與固定工具，原版 EXE 由合法本機 ZIP 重生；輸入唯讀、案例輸出可寫掛到 `/out`，
執行 `python3 -B /out/full-build.py <新的空白案例名稱>`。
`seed-CODE.c` 是完整 C／ASM 輸入；分類清單含 16 個後續純 C 與 67 個混合函式，106 純 C 分類基準保持。
`compile-DATA.py`、`materialize-DATA.py`、`data_proof.py` 保留完整 DATA 鏈；唯讀核對入口為 `verify.py`。


第十九批完整原生重編入口為上述保存目錄的`full-build.py`，支援173個compiler函式、52個LIB CODE模組與全部DATA。
`seed-CODE.c`為已核對的完整C／ASM輸入，`compile-DATA.py`與DATA來源鏈保持；
恢復封存與固定工具、由合法ZIP重生原版後，以輸入唯讀、案例輸出可寫掛到`/out`，
執行`python3 -B /out/full-build.py <新的空白案例名稱>`。
來源分類清單增為15個後續純C及52個混合函式，106純C分類基準保持，不因增加ASM區塊而吞入普通C。
全唯讀核對入口為`verify.py`；新增CODE後同時核既有CODE、DATA與原始799函式數。

第十八批完整原生重編入口為上述保存目錄的`full-build.py`，支援149個compiler函式、52個LIB CODE模組及全部6736 DATA。
恢復封存、固定工具與合法原版ZIP後，輸入唯讀、案例輸出可寫掛到`/out`，
執行`python3 -B /out/full-build.py <新的空白案例名稱>`。
`data_objects.py`讀真實compiler DATA物件，`materialize-DATA.py`建立來源，`data_proof.py`獨立核全部初值／fixup／零區。
來源分類器使用`units/KOLTOWN/classification-baseline.json`的106純C基準，與現行149函式seed分開保存。
新增CODE來源時保留此DATA鏈與全部既有C／ASM／LIB；全唯讀收據核對入口為`verify.py`。

第十七批完整來源核對需保留`source_categories.py`、`units/KOLTOWN/assembly-inputs.json`、
`pure-inputs.json`與`source-base/adaptation.json`的106純C分類基準，均在上述保存清冊內。
`verify.py`會在容器暫存重建必要檔案並呼叫分類器；輸入唯讀，適用118純C、31混合及52個LIB模組。
需原生重編時，從封存恢復`KERNELS.C`、`WRAPPERS.C`、`trial.py`與固定工具，
原版EXE由合法ZIP重生，再執行`python3 -B /out/trial.py <新的空白案例名稱>`。
前批還原的函式不得退回原始片段；接續新增來源時一併擴充明列分類清單並核全部既有函式。

第十六批的`prepare.py`、`plan.json`、`trial.py`、`checkpoint.py`及`town-link.py`
保存在上述本機保存入口。重編時解壓到容器暫存，依`fetch-tools.py`與固定manifest重取工具，
由合法本機ZIP重生原版EXE，再將輸入唯讀、輸出可寫掛到`/out`。
執行`python3 -B /out/trial.py <新的空白案例名稱>`，支援Town106個純C與52個LIB模組的原生重建。

第十五批的`prepare.py`、`plan.json`、`trial.py`、`checkpoint.py`及`town-link.py`
保存在上述本機保存入口。需重編時解壓到容器暫存，固定原版與工具唯讀、輸出可寫，
執行`python3 -B /out/trial.py <新的空白案例名稱>`，支援81個Town純C與52個LIB模組的完整原生重建。
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
python3 -B /out/matching-c2-20261008/verify.py
python3 -B /out/matching-c3-20261008/verify.py
python3 -B /out/matching-c4-20261008/verify.py
python3 -B /out/matching-c5-20261008/verify.py
python3 -B /out/matching-c6-20261008/verify.py
python3 -B /out/matching-c7-20261009/verify.py
python3 -B /out/matching-c8-20261009/verify.py
python3 -B /out/matching-c9-20261009/verify.py
python3 -B /out/matching-c10-20261009/verify.py
python3 -B /out/matching-c11-20261009/verify.py
python3 -B /out/matching-c12-20261009/verify.py
python3 -B /out/matching-c13-20261009/verify.py
python3 -B /out/matching-c14-20261009/verify.py
python3 -B /out/matching-c15-20261009/verify.py
python3 -B /out/matching-c16-20261009/verify.py
python3 -B /out/matching-c17-20261009/verify.py
python3 -B /out/matching-c18-20261009/verify.py
python3 -B /out/matching-c19-20261009/verify.py
python3 -B /out/matching-c20-20261010/verify.py
python3 -B /out/matching-c21-20261010/verify.py
python3 -B /out/matching-c22-20261010/verify.py
python3 -B /out/matching-c23-20261010/verify.py
python3 -B /out/matching-c24-20261010/verify.py
python3 -B /out/matching-c25-20261010/verify.py
```

需乾淨重編時按保存清冊抽出源檔、固定候選工具及命令到Docker暫存，來源唯讀、輸出可寫。
可重取的工具及可重建EXE只保留hash，原版EXE由合法本機ZIP重生，不冒稱已匹配整支遊戲。
