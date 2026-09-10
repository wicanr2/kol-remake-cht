# 戰場地形索引修正驗證

日期：2026-09-08。依 SPEC-008 亂數地形索引增補，修正 `Generate` 與 `Scatter`：
原版線性位置 `Y*64+X` 轉為正式資料的 `X*64+Y`，不改擲骰順序或機率。
證據入口：`docs/re/20260908-battle-field-index.md`。

dosgolem 執行原版生成／灑點區段，以三組共同固定種子 0、1、65535，
比較一般遭遇生成及固定非對稱測試地圖灑點。六組共 **24,576 格全部一致**，
六組亂數終態亦一致。固定圖包含不可灑點的道路及牆。

- 修正前：`workplace/terrain-oracle-before-20260908/terrain.log`，六組皆有地形差異。
- 修正後：`workplace/terrain-oracle-fixed-20260908/terrain.log`，全部通過。
- 每批 `manifest.json` 保存原版、dosgolem 版本及受測來源雜湊。

首次完整回歸僅 `TestGenerate` 失敗：舊測試將擲骰順序誤當成先走 Y 軸。
依原版契約改為檢查 `(0,0)`、`(1,0)`、`(0,1)`，乾淨重跑完整 Go 測試與建置通過。
失敗紀錄保留在 `workplace/terrain-build-20260908/`，成功結果在
`workplace/terrain-build-final-20260908/`。

二進位 SHA-256：`c3730080dd9808f9705a6700161a9fb613ce50b3f2d6205985156cc76efa6c47`。
輸出擁有權均為 1000:1000。本批 KOL 容器均已移除；同映像下其他專案容器不處理。

這是受控區段的正式實作對拍，不代表正常戰鬥 GUI、完整地形快照匯入、
指令／攻擊結算或戰後保存已完成。AppImage 尚未重建，自評維持 75/100。
