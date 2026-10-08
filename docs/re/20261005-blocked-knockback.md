# RE-253：受阻打暈與掉落呼叫

日期：2026-10-05。狀態：DRAFT，原版受阻隊員已證實，隔離原型通過；返回暫存與舊H仍未驗。

沿逆向技能、dosgolem正常玩家路徑與文件職責入口。RE-252已發現原版
IDA linear 0x27DD7至0x27DF4印打暈及掉落句，該分支未見sub_269BA。
現行Knockback兩分支共同掉落，是否應修正需完整控制流與正常樣本查證。

輸入KOLTOWN-unpacked.EXE SHA-256
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
IDA Pro9.4 linear載入基底0x10000、DGROUP0x2DED0；沿RE-252的867道已核原始指令。
dosgolem固定revision968ca774705853276ae55c8d1163c369481ac71f，以封存來源重建。
預先選seed2與3，旅店招呼後招募前各設定一次；保留兩個完整結果，不挑通過樣本。
正常片頭、書本、六人招募、自然旅行與REST首場，最多20回合或2500個戰鬥鍵。
只觀察0x1A8FD記錄158／159／15B／15C及sub_269BA的參數、呼叫點、原始手持與地上槽。
沒有座標／人物／遭遇注入、direct-entry、重播種或等待替代。
未出現受阻打暈也保留結果，不宣稱該分支通過。正式338來源與Linuxc8764592保持。

本機入口workplace/blocked-knockback-20261005/prepare.py、probe.go、build-original.sh與run-original.py。
原版資料與衍生畫格僅本機；未改玩法、資料契約、HD、發行或公開方式。

V1反例：兩批467狀態均停在0xE6治療問句，完整畫面證實YES／NO圖示，驗證腳本未回答。雖原版程序退出0，仍拒絕作戰鬥覆蓋。V2保留同seed2／3，見治療問句便以正常圖示選NO；另預先加四次32格方向方圈，總旅行690鍵，最多20回合或2500戰鬥鍵。未挑新種子，未改原版或正式規則。

## 正常原版與反例

已證實：seed2正常第317格，Olear保持33／32，完整手持Elf Bow記錄
`72b0c013ee000000225828ff190be200`在158、15B及同鍵結果均相同。
沒有兩句之間的sub_269BA呼叫，完整256-byte地上槽保持，+25由5清0。
H索引仍0，畫面卻顯示Short Sword落地；地上槽本身全零，不能把訊息當實際掉落證據。
本輪未研究空記錄的名稱解碼原因，不推定其他索引、滿槽或空手訊息。

seed2共433狀態，第14回合自然返回。seed3共578狀態，完成20回合後在第21回合
規劃入口有界停止，戰鬥仍在進行；沒有打暈樣本，不能稱第二次分支通過。
兩組共1011正常狀態，只有一個已驗受阻持物隊員樣本。怪物裝備未由這個觀察器採集。

V2首次seed2在3GiB遭SIGKILL，cgroup明示oom_kill=1。原容器提高到8GiB後
seed3仍SIGKILL；第二次沒有留下終態cgroup，原因不另稱已證實OOM。
固定dosgolem來源bios.go:int33在每次AX=3持續append Polls；這是已查出的無界診斷記錄。
seed2同一程式／種子／plan在8GiB完成。seed3只加GOGC=20、GOMEMLIMIT=6GiB
與gctrace，維持8GiB硬上限、900秒test及950秒子程序逾時，593.684秒完成。
實際抽核peak達6103363584 bytes，抽核oom_kill一直0；這不是完整最終峰值。
原版兩份失敗前綴共584PNG逐檔與完成重跑相同，沒有重播種、挑結果或修改模擬器。

## 隔離候選與停止線

正式Linuxc8764592的正常第五回合會搬走精靈弓。候選同seed2、321等效鍵的雙語
前版／候選與各自唯讀副本共2576狀態。第320與321步完整trace只差地上清單；
56完整PNG的前後及觀察副本相同。原版317格對應重製320步。
完整手持記錄與原版相同。候選整份Party、Ground保持，Actors只清Olear的Command，
其他觀察欄保持；兩版結算seed2126174334與原版相同。未注入狀態或遭遇。

反例單測涵蓋隊員及怪物，舊正式兩者均失敗，隔離candidate測試通過。
另有既有seed1雙語1136保存／26真實新程序重載，68PNG與完整v3保持；
此回歸沒有受阻打暈，不能取代新增分支的戰後保存驗收。

只讀事件另揭露ReturnScratch：前值6，舊版掉落後0，候選保持6。
原版快照未讀word_39A54，因此返回成本／暫存值仍未知。它與H的word_30D02不同，
不能拿同一個欄位代用。下一閘門是只讀補採原版暫存及正常戰後保存鏈，
再審查有限READY；此原型不接正式。人物v3、怪物v2、送禮及雷聲核准保持。

重生入口：build-original.sh、run-original.py、run-retry.py、run-retry-3.py、
test-candidate.sh、run-candidate-regression.sh、build-observers.sh、run-observers.sh、
run-blocked-normal.py、review-native.py、review-blocked-normal.py與seal.py／verify.py。
使用固定eob-remake-go:1.26.7-ebiten2.9.9-audio，UID/GID1000:1000、network none；
專案/src唯讀，只有本機研究輸出/out可寫。有Xvfb的流程均trap結束背景程序。
完整重生應用新的研究輸出，腳本拒絕覆寫已存在的案例。所有原版payload僅本機。
本輪沒有新的正式原版規則、HD版面決定、平台驗收、版號、封包或發布。

封存入口：[有限收據](../../workplace/blocked-knockback-20261005/receipt.json)、
[完整重驗](../../workplace/blocked-knockback-20261005/verify.py)。

## RE-254 有限勘誤

本檔上文與收據保存RE-253當輪狀態。後續[RE-254](20261005-blocked-knockback-return.md)
已只讀採集同一KOLTOWN-unpacked.EXE的IDA linear word_39A54 0x39A54／DS:BB84，
第317格暫存6已證實，H的word_30D02仍0。原版760正常／41新程序重載與候選雙語
正常回城、六人各17原始carried槽及持久欄通過，解除本分支暫存與戰後保存的當輪未知。
世界額外SPACE暫存差分、等效Right確認、舊H、複數呈現與完整終態另列限制。
候選尚未接正式，規則DRAFT待確認與READY審查；不回寫本輪封存歷史。

## RE-262 有限勘誤

上文「畫面卻顯示Short Sword落地」有誤。[RE-262](20261006-stale-ground-name.md)
重驗原始第317格PNG，物品名稱實際空白。H=0的全零地面記錄由原版解碼器
產生24個空白。142道原始指令與該行六列字模核對通過，錯填SHORT SWORD差143像素。
原始收據及歷史敘述保留；受阻保留精靈弓與RE-254暫存6的結論保持。
