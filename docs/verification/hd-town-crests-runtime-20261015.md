# 七城徽章原生 HD 驗收

狀態：限定範圍 CONFORMED。六張新徽章沿已核准的原生 Canvas 紋章風格，霍布城核准 PNG 保持位元組相同。七城均已接入正式開發程式；現行發行包仍為 `v.1.0.2-20261013`，尚未納入本批。

[七組同狀態比較](hd-town-crests-runtime-20261015/review.html)、[驗收收據](hd-town-crests-runtime-20261015/receipt.json)及[素材清冊](../../assets/hd/first-slice/town-crests/manifest.json)記錄來源與範圍。本批完成於2026-10-16，入口沿用開工日期20261015。

| 項目 | 結果 |
|---|---|
| 來源與原生圖 | TWN1…7PICS0，來源56×57、二值透明遮罩；七張原生320×320 PNG／RGBA及來源PNG／RGB／RGBA均驗證 |
| 核准圖保持 | 霍布城PNG SHA-256 `3e9fa444a362271b2978c508f7a40c04ba9a14cf212d55cd2ff5fcf9683d30f4`；完整HD相對前版最多一色階 |
| 受控七城畫面 | 28張畫格，完整可序列化Outdoor與觀測中文相同；EGA RGBA零差；獨立透明合成最多一色階；舊徽章框外零差 |
| 中文 | 168個實際觀測漢字、6種字，缺字0；只涵蓋收據中的日期文字，不擴大到所有動態分支 |
| 遊戲內切換 | 第一城與霍布城、1920×1200及960×600、鍵盤及滑鼠，共24個實際F10事件；完整Outdoor保持、HD往返最多一色階 |
| 正常操作與保存 | 兩條既有操作共每側41狀態，文字與控制器相同，EGA零差、HD最多一色階；住宿11狀態、重載5狀態與前版v3通過 |
| 正式來源 | 412份Go／module、390份HD資產；全套與sceneverify通過，正式執行檔SHA-256 `7a1dfaa7eae9dbccab12ade81065402a2756f12eeacc537eac2e43b7f52080a8` |

圖形直接由 Chromium 154.0.8037.92 原生 Canvas 匯出，六張圖保留原來源的主色與構圖。銅框、色盤與320×320位置沿已核准 `review.js` 的 `crest()`；[繪圖程式](hd-town-crests-runtime-20261015/crests.js)與生成清冊保存可重生來源。日輪、樹木、帆船、拱橋、城門及雙色紋章依實際來源核對，不把圖案名稱當成原作設定證據。KOLMPICS28仍是結構AND遮罩，沒有改成可見美術。

載入器先驗完整七城順序、實際archive數量、圖號、來源與原生雜湊。全部成功後才交換HD畫布；缺件、來源變動與最後一張錯誤保持既有畫布。原生徽章位於130,502至450,822。第一至六城不再保留舊EGA徽章大框，空白區回到同一HD背景，差異範圍限120,490至550,940。

本批使用鎖版Linux Docker工具鏈。場景由真實typed門位準備，沒有新增原版oracle或自然通關聲明。實際GUI只固定驗收標記時鐘，正式規則、動畫、存檔格式與生命值保持。最後一城GUI首次停在擷取途中，原輸出保留；僅補跑該城，同命令通過，未重跑已通過項目。受控場景完成依[SPEC-040](../spec/040-controlled-scene-acceptance.md)，原版parity與平台真機另列限制。

完整畫格、模型、v3及中止輸出僅留本機，保存與清理沿[證據入口](retained-evidence.md#七城徽章hd正式接入)。重生時先核對412份來源與390份素材索引，將固定來源掛`/src`；從封存取出`baseline-source.tar`、`gallery.go.txt`、`normal.py`、`save.py`、`gui.py`、`verify.sh`及`check.py`至空白輸出目錄，再掛`/out`，原收據保持唯讀。使用`kol-package:1.26.7-ebiten2.9.9-20261011-r1`、UID/GID1000:1000、無網路、8GiB、4CPU及512程序，外層1500秒逾時，執行`bash /out/verify.sh`及`python3 /out/check.py`。原生Canvas匯出另用`logh3-hd-preview:20261011-r1`將專案掛`/repo`，恢復本批工作區後執行`export.py`，只支援已列七張徽章。城鎮／世界地圖、戰場、天候、傷勢及其餘事件仍開放。
