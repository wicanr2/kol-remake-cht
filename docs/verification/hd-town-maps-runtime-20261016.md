# 第一城連續手繪地圖

狀態：第一城TOWN1地圖畫稿限定範圍CONFORMED。使用者選定A：按原地圖位置呈現連續手繪，契約見[SPEC-039](../spec/039-hd-art-ui.md#第一城位置地圖接入審查)。

42張1254×1254原生不透明PNG完整覆蓋64×64共4096格。原地圖、門位、通行、鏡頭與v3保持。畫稿先固定取樣整張，再按原座標裁格；核准北方原圖位元組保留。走路標記及其他城鎮、世界地圖仍待HD化。

[正式比較](hd-town-maps-runtime-20261016/review.html)、[收據](hd-town-maps-runtime-20261016/receipt.json)及[截圖權利與雜湊](hd-town-maps-runtime-20261016/manifest.json)保存穩定入口。

| 驗證 | 結果與範圍 |
|---|---|
| 正式來源 | 417份Go／module、436份HD資產；全套與sceneverify通過 |
| 正式畫布 | 23場景92畫格，含街區、城堡、兩橋、地圖角落、16門面與城門林地；完整Outdoor與中文相同，EGA零差、原生合成及相對候選整張HD最多一色階、同次序對照視窗外零差 |
| 完整覆蓋 | 隔離首輪101場景404畫格涵蓋4096格；兩橋修正版另覆核23場景92畫格、6鏡頭360格，其餘40張原圖保持 |
| 實際F10 | 正式程式在1920×1200鍵盤及960×600滑鼠共12事件通過，完整Outdoor保持，HD往返最多一色階 |
| 保存 | 415份隔離正常候選完成住宿11狀態、重載5狀態與前版v3，保存僅saved_at不同；正式全套另驗保存規則 |
| 錯誤來源 | 完整組、缺圖、缺格／重疊、越界、尺寸、透明、PNG／RGBA與來源錯誤皆拒絕；實際Grid與來源不符時不繪製半張地圖 |

首輪正式跨程序對照在既有字形出現一色階差異；原畫格及診斷保留。改以同一程序、相同繪製次序作地圖有／無的比較，視窗外零差，整張HD相對前次候選仍最多一色階。比較掛鉤只在隔離驗證副本，正式程式沒有該開關。

本批只驗重製版受控場景與保存相容，不代表原版parity、正常抵達第一城或全程通關。僅Linux Docker；現行`v.1.0.2-20261013`交付與tag保持，全畫面HD仍開放。

重生沿[保存入口](retained-evidence.md#第一城連續手繪地圖hd正式接入)的壓縮工具與原檔名。正式來源先核對[source-inputs.json](hd-town-maps-runtime-20261016/source-inputs.json)及[asset-inputs.json](hd-town-maps-runtime-20261016/asset-inputs.json)，使用`kol-package:1.26.7-ebiten2.9.9-20261011-r1`、Go1.26.7／Ebitengine2.9.9、UID/GID1000:1000、無網路、8GiB、3CPU、256程序與1200秒外層逾時，專案掛`/src`、解開的證據掛`/out`。正式建置與同次序比較分別執行`build-formal-city.sh`及`build-formal-city-comparison.sh`，再執行`check-formal-city.py`；歷史417份來源與436份素材必須相符。
