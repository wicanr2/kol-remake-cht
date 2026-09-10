# 彈藥資訊框接線驗證

2026-09-08，依 SPEC-022 的 READY 增補，戰鬥裝備模式於 (24,60) 顯示當前隊員的
ARROWS:／BOLTS: 與即時剩餘發數。武器旗標與種類沿用現行 Member，發數讀 Actor；
肖像、空手、近戰武器或敵人不顯示。不新增持久庫存或資料格式。

完整 internal／cmd 測試與建置通過；消耗後顯示更新、肖像隱藏與丟棄後清除皆通過。
128 組正式標頭／數字與 dosgolem 原版比較通過，六組真實原版文字流程通過。
整批面板測試 0.393 秒，包含原有 2,550 組消耗及其他面板案例。
第一次第六槽比較失敗是測試漏設 Actor.Member，修正後保留失敗收據並乾淨重跑。

正常 GUI 由標題選 Olear、招募、離開旅店、探索及進入戰鬥，128 個輸入完成。
人工確認 battle-equipment.png 顯示 Olear、Elf Bow 與 ARROWS:20，位於裝備列下方且未裁切。
後續延伸一發射擊，第一批結算後先輪到 Theodoric，該畫格不能驗證 Olear 的剩餘數字。
正確人物的後續驗證另存 shot-current-actor；結果另追加，不把錯人物畫格當成通過。

執行檔 `workplace/ammunition-panel-20260908/kol`，SHA-256：
`aa40c20fefbee39e555a55e4110d35893ee2cdd8349ed15d3049c59193132aa7`。
完整測試、來源雜湊、GUI 原始畫格及 oracle 收據位於同目錄。
初始 GUI 腳本在延伸前未另存快照，後續依唯一新增區塊還原並核對執行前雜湊一致，
保存 initial-gui-script.sh；第一次射擊保存 verified-script.sh。

這些 GUI 為 remake 正常路徑，不是 dosgolem 正常原版戰鬥同狀態收據。
原版仍為有界常式及文字出口證據。AppImage 未重打包，SPEC 保持 READY。

## 正常射擊回讀通過

shot-current-actor 共 150 個正常輸入，從標題選 Olear、招募 Theodoric、離店、
探索、進戰場，使用射擊圖示及目標游標選右側空格，確認規劃並完成一輪。
下一輪先到 Theodoric，選休息後回到 Olear；切換裝備模式，人工確認
ammunition-current-actor.png 顯示 ARROWS:19。初始 battle-equipment.png 為 ARROWS:20。
未使用座標注入、直接進入戰場或彈藥測試旗標；GUI 選單與正常結算完成發數更新。

gui-verification.json 保存 150 個輸入計數、畫格與輸出雜湊；執行前後執行檔與
腳本相同，全部產物為 1000:1000。容器使用 --rm、資源上限、逾時及 Xvfb trap，
本批工作結束後專案相關執行中／已停止容器清單均空。
本次正常射擊為 Elf Bow 單發抽測；弩裝填與耗盡目前仍是原版常式／規則層證據。
未提交或推送。
