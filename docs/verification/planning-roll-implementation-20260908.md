# 規劃檢視門檻實作驗證

日期：2026-09-08。SPEC-020 READY 範圍已實作，尚未 CONFORMED。

Round 保存目前規劃位置的 InspectionRoll 與一次性進入標記。BeginPlanning 重設標記，
每次 skipDown 在檢查可否行動前先消耗該位置的一次門檻；倒下／空位清除招式及防禦。
EndTurn 與 AdvanceToParty 共用此入口；重複進入同一位置及 BeginTurn 取消返回不重擲。
排序保留六個隊員槽與本場敵人槽，空槽按原版 EmptyKey 排最後，不再直接刪掉。
敵人 AI 之前已消耗該人物的門檻，沒有把所有門檻預先批次擲完。

既有排序測試移除「空位不進順序」的過時預期，改驗有效範圍與空槽排最後。
新增生命週期測試：倒下人物、空位、有效人物依序消耗三次；重複入口與 BeginTurn
三次不增加；輪末不增加；新回合重設後排序骰及位置骰依次消耗。

`kol-package:20260907` 內完整 `go test ./internal/... ./cmd/kol -count=1` 及建置通過。
後加的生命週期測試另以 `go test ./internal/battle -run TestPlanningInspectionEnteredOnce
-count=1 -v` 通過。產物 `workplace/planning-roll-fix-20260908/`，執行檔 SHA-256：
`92d00db74fe02018ffa24b3c14dcbfc9ae273d0d1a30a3ca1f7a3254a7a2abea`。

原版順序證據及 dosgolem 有界切片收據見 SPEC-020；本次尚未新增正式 remake 與原版
同一骰序的整合比較或正常 GUI。舊固定種子畫面不能直接沿用，須先驗證新序列再重生。
左下命令面板仍未繪製，完整 AppImage 及保存驗收未完成，不宣稱整體等價。
未提交／推送，本批容器自動移除。

## 後續正常 GUI 抽測

同一建置以 `verify_starting_inn_gui.sh` 執行，設定 KOL_ENTER_BATTLE=1、
KOL_WORLD_STEPS=100、KOL_KEY_DELAY=0.3、KOL_EQUIPMENT_PANEL=1，種子一。
從標題選 Theodoric、招募 Poontang、離店、出城步行遇敵，再切換肖像／裝備。
126 個正常輸入完成，人工確認兩種面板。肖像來回與裝備來回的整張畫面像素差異
各為零；執行檔與腳本雜湊前後相同，輸出擁有權 1000:1000。
收據 `workplace/planning-roll-fix-20260908/gui-verification.json` 逐檔保存雜湊。
這補足本次正常入口操作抽測，尚非完整戰鬥結算或原版同骰序對拍；SPEC-020 仍 READY。
