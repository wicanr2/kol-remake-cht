# 完整物品與隊伍HD正式接入

狀態：本切片有限CONFORMED。沿[SPEC-039](../spec/039-hd-art-ui.md#完整裝備頁正式接入)接入六型、162款裝備、獨立翼部、28武器、四盾與城鎮／野外背板。完整物品17槽及隊伍頁使用原生HD；其他NPC、地圖、戰場、天候與事件仍續作。
[正式比較頁](hd-equipment-runtime-20261015/review.html)提供正常物品與六型隊伍EGA／HD，以及受控盾面、前臂與腰帶畫格；[截圖清冊](hd-equipment-runtime-20261015/manifest.json)分開標示正常路徑與夾具。

| 驗證 | 結果 |
|---|---|
| 完整素材 | 46張PNG、169身體／裝備部件及32武器盾牌，每張與每個裁取RGBA雜湊核對；169部件原生尺寸涵蓋呈現矩形 |
| 實際來源 | 216張PNG、原始RGB與呈現RGB分別核對；七個archive數量及typed版面SHA符合固定輸入，GPU再次核對實際指紋與尺寸 |
| 嚴格負例 | 23項缺件／來源／幾何／原錨點負例、nil／錯母體／11手圖拒絕；五項UI原子替換負例保持原組，錯typed來源保留EGA且沒有半套快取 |
| 受控合成 | 38張完整HD與V10盾面、V12護腿／前臂及V13腰帶原型最多一色階，alpha零差；646筆中文槽位紀錄 |
| 個人盾牌 | 六型預設／複製預設匹配原生HD；十二張修改盾面的240×228顯示區與已逐40×38核對的V10內容完全相同 |
| 正常操作 | 從標題選Kenneth、Shyla的17槽及招募六型隊伍，三組每側106狀態與中文字紀錄相同，106張EGA RGB／alpha完全保持 |
| 實際F10 | Shyla物品／隊伍頁、1920×1200與960×600、鍵鼠共24設定事件，完整Inn／隊伍不變，HD往返零差，Ctrl+Q退出0 |
| 保存重載 | 正常住宿11狀態、另程序重載5狀態相同，完整v3只差saved_at；修改前程式產生的舊v3可讀且原檔SHA保持 |
| 最後來源綁定 | 男性體型殘留的固定原點改讀同一已驗typed綁定。UI／入口窄驗通過，正常31張HD與前一正式建置最多一色階，狀態及中文相同 |

[畫格](hd-equipment-runtime-20261015/gallery-check.json)、[正常操作](hd-equipment-runtime-20261015/normal-check.json)、[GUI](hd-equipment-runtime-20261015/gui-check.json)、[保存](hd-equipment-runtime-20261015/save-check.json)及[最後窄驗](hd-equipment-runtime-20261015/origin-bind-check.json)列完整範圍。
正式391份Go／module及185份HD資產有[輸入雜湊](hd-equipment-runtime-20261015/formal-inputs.json)與[素材雜湊](hd-equipment-runtime-20261015/asset-inputs.json)。全套及sceneverify通過，最後兩處來源綁定改動另跑受影響UI／入口；兩個階段由[前一階段輸入](hd-equipment-runtime-20261015/full-check-inputs-before-origin-bind.json)及[收據](hd-equipment-runtime-20261015/receipt.json)分開記錄。

原生資料組位於[equipment/manifest.json](../../assets/hd/first-slice/equipment/manifest.json)，[generation.json](../../assets/hd/first-slice/equipment/generation.json)保存46張圖片的生成工具、來源entry／SHA與精確提示。兩款護腿原生高度346／342不足354／348，使用內建image_gen補成680×814及682×812，[圖集](../../assets/hd/first-slice/equipment/legspics-build0-01-native-painted.png)與[最終提示](../../assets/hd/first-slice/equipment/prompts/leg-native-correction.txt)已保存；其餘167部件與所有World矩形保持。

V11上身卸裝畫格暴露前臂缺段，V12按原生身體分離alpha段補回前臂，限既有手部區域並快取。V13依RE-128將腰帶維持固定頁面層，不隨握點移動。原缺口與失敗輸出全部保留；沒有改玩法、typed表、原圖索引或kol-save-v3。

首次正式GPU驗證發現HEADPICS27的兩個透明像素保留白色RGB。擴查另有TORPICS4、LEGSPICS36及49；四張原PNG與原始RGB全部保持，增加獨立呈現RGB，按預乘alpha契約核對GPU。修正後相同38畫格、三條正常路徑及保存命令乾淨通過，未略過來源或放寬色差。詳見[指紋勘誤](hd-equipment-runtime-20261015/source-fingerprint-correction.json)。

`internal/hdart/equipment.go`只定義純呈現型別，分層由[SPEC-038](../spec/038-package-boundaries.md)與既有守門測試維持。原始身體／手部錨點由已驗typed版面及PNG著色框於記憶體推導，Binding排除JSON；清冊不含原始錨點或完整模型。正式UI不讀原型`/out`路徑，沒有測試旗標；GUI僅在隔離主程式加入原有20行觀察，不改正式入口。

重跑腳本：[建置](hd-equipment-runtime-20261015/build.sh)、[GPU](hd-equipment-runtime-20261015/gallery.sh)與[夾具](hd-equipment-runtime-20261015/gallery.go.txt)、[畫格核對](hd-equipment-runtime-20261015/gallery-check.py)、[正常操作](hd-equipment-runtime-20261015/normal.py)與[核對](hd-equipment-runtime-20261015/normal-check.py)、[GUI](hd-equipment-runtime-20261015/gui.py)、[保存](hd-equipment-runtime-20261015/save.py)、[最後窄驗](hd-equipment-runtime-20261015/origin-bind.sh)。在鎖定Docker映像中將本批恢復工作區掛至`/out`，專案掛至`/src`；依[保存入口](retained-evidence.md#完整物品與隊伍hd正式接入)重建對應來源階段及觀察覆蓋檔。

本批只驗Linux容器，既有Release保持。完整型別模型、原錨點、v3與失敗日誌只留本機封存；工作區逐檔核對後清理。其他HD類別及自然完整通關不作新完成聲明。
