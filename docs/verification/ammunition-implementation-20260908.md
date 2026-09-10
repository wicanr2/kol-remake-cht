# 隊員彈藥消耗修復驗證

2026-09-08，依 [SPEC-022](../spec/022-battle-ammunition.md) READY 修正
`State.shoot`：真正射出後共同扣除，裝填不扣；耗盡清除指令，需裝填武器回到裝填。
不改敵人的怪物索引，不新增跨戰鬥庫存。原版定位及限制見
[彈藥補證](../re/20260908-ammunition.md)。

- 完整 `go test ./internal/... ./cmd/... -count=1 -timeout 120s` 及建置通過。
- 射擊函式回歸涵蓋裝填、射空、撞牆與最後一發；敵人索引不被扣除。
- dosgolem 有界消耗 2,550 組直接比較正式函式通過；128 組顯示分支亦通過。
  整批面板測試 0.147 秒。這不代表命中彈道或正常整場戰鬥對拍完成。
- 原版比較來源未變、輸出雜湊與全部產物擁有權 1000:1000 核對通過。

本機產物 `workplace/ammunition-implementation-20260908/kol`，SHA-256：
`235e42438f6c8affde7f5441663224db1b6b48478f7a0d77b488964d5b671280`。
測試、來源雜湊及原版收據放於同目錄；此為中間產物，未重打包 AppImage。

使用既有 `kol-package:20260907`，斷網、4 GiB、2 CPU、256 程序、非 root，
原始資料唯讀。首次完整測試因 go.work 指向未掛載 `/dosgolem` 而未開始；
保存 first-workspace-error.log，明確設定 GOWORK=off 後同映像乾淨重跑通過。
建置有唯讀模組快取寫入警告，但退出碼 0，執行檔雜湊已核對。
一次性容器與 Xvfb 已清理，專案容器清單為空。

彈藥資訊框、正常玩家射擊、完整存讀檔及 AppImage 仍未驗收；SPEC-022 保持 READY。
未提交或推送。
