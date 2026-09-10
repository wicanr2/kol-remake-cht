# 競技場繁中對話修正與進場抽驗

日期：2026-09-08。G03 局部驗收，SPEC-025 維持 READY。

## 缺陷與修正

競技場沒有已證實的店主肖像，傳入的 `Portrait.Archive` 為空。
英文 `Inn` 已略過空肖像；繁中 `InnCJK` 卻呼叫素材讀取，導致
「繪製失敗：manifest 沒有 archive」，中斷後續對話與圖示繪製。
繁中介面現採相同的空肖像處理，並在 SPEC-025 補明契約。
本修正沒有猜補原版特殊肖像；該圖像仍是證據缺口。

`tools/verify_arena_gui.sh` 新增執行權限前置檢查及繪圖錯誤閘門，
避免只有按鍵成功就誤判畫面完成。

## 驗證與限制

- Docker image：`kol-package:20260907`，SHA-256
  `3d32441befe02810c3220a2e9f2db5b84ee65c703e5b65ee563712a3519ad193`。
- `go test ./internal/... ./cmd/... -count=1 -timeout 120s` 與正式程式建置通過。
- 固定測試 seed=1，使用已有的受控 v1 資格存檔；來源及位置調整見
  `workplace/arena-gui-fixed-20260908/input-receipt.json`。
- 經匯入提示、世界移動、競技場門口、選 Moraine、確認、隧道、拒絕下注、
  開場提示，進入戰場規劃。九張截圖與無繪圖錯誤日誌已保存；人工檢視
  選人清單與戰場畫面，確認繁中標題、清單、圖示及固定圍牆可見。
- 新執行檔 SHA-256：
  `ce177bb544f9984d46a4b9b23b1fc32fa4c613baf018050c2b14aeea61d27cc9`。
  完整雜湊索引：`workplace/arena-gui-fixed-20260908/receipt.json`。
- 前次失敗仍留在 `workplace/arena-gui-20260908/`，未覆蓋；首次權限錯誤
  屬環境問題，後續空肖像錯誤屬產品問題，兩者分別處理。

本輪未執行新的 dosgolem 收據，不宣稱原版畫面或戰鬥結算對拍。
固定底圖的 dosgolem 4096 格比較沿用 SPEC-025 已登錄證據；受控資格
不能代替正常取得資格，亦不涵蓋勝敗、下注支付、升階及完整保存鏈。
AppImage 未重建，自評維持 74／100。

所有工作容器使用 `--rm`，本批結束無該 image 的殘留容器；
輸出擁有權均為 1000:1000，`git diff --check` 通過。
