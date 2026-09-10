# 護甲減傷接線驗證（2026-09-08）

已依 SPEC-017 READY 接入近戰與射擊，並完成受控原版後段比較；**尚未完成正常 GUI
戰鬥、完整存檔或 AppImage 對拍**，SPEC-017 不升為 CONFORMED。

## 修正

- `Actor.Armour` 保存三個護甲槽的獨立物品副本；隊員與每名敵人分別建立，避免磨損污染來源。
- 兩條攻擊路徑在選部位後、傷勢前擲護甲骰，使用各自的隊員／敵方映射。全零空槽也擲骰。
- 原版的負加值、完全阻擋、耐久歸零保留與負數清槽均保留。受創值 Hurt 使用護甲前傷害。
- 耐久同步 Condition 與完整 Record；清槽移除物品並更新重量／徒手裝備快取供後續使用。
  缺少有效 16-byte Record 的非空護甲明確失敗，不猜造旗標。
- 勝利與撤退從 Actor 回寫三個穿戴槽並同步 Carried，後續不再沿用破損前物品。
- 草稿中的 Percent 假設已訂正：它來自物品 +0x0E，是盾牌格擋率，不能按耐久重算。

## 驗證

使用既有 `kol-package:20260907`、Go 1.26.7、斷網及資源受限的非 root Docker。
`workplace/spec-fixes-20260907/run.sh` 從唯讀來源複製至容器臨時目錄：
`go test ./internal/... ./cmd/kol -count=1` 與 `go build ./cmd/kol` 通過。
測試與本機執行檔保存於 `workplace/armour-implementation-20260908/`。

新增回歸涵蓋近戰／射擊完全阻擋仍保留 Hurt、隊員及兩名敵人的物品隔離、破損減重、
缺少原始記錄拒絕，以及勝利／撤退的 Condition、Record、Carried 與來源不污染。
原 `TestResolve` 假設 Hurt 等於護甲後傷勢；依原版證據改為明確六次骰序、Hurt=16、傷勢=14。
第一輪失敗保存在 `first-tests/`，沒有透過跳過護甲骰讓舊斷言過關。

`KOL_CHECK=durability` 以新版 dosgolem 比較 `TestArmourReductionOriginal` 的 84 組：
相同三槽記錄、部位、輸入傷害、骰值，逐項核對剩餘傷害、完整記錄、Condition 與未變的 Percent。
另三項原版裝備測試一併通過，組合用時 0.046 秒。收據位於 `oracle/manifest.json`，
在複製來源前建立，`source_changed_during_run=false`。

原版只執行護甲後段，最後扣傷子常式替換成參數觀測；這不是完整攻擊或正常玩家路徑。
武器傷害上限造成的磨損、其他四個磨損呼叫及完整存檔仍須繼續處理。
本輪未更新 AppImage、未提交、未推送。建置後只更正一行骰序註解；建置清冊若於事後建立，
必須標示為執行後盤點，不能假稱為執行前來源收據。
