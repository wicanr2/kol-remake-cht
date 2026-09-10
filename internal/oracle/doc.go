// Package oracle 是拿 dosgolem 當原版 oracle 的那一層（RE-132）。
//
// dosgolem 是無頭、決定性、可以 import 的 DOS 執行器。這一包底下的對拍測試都在
// `-tags oracle` 之後，沒有原版檔案的人跑 `go test ./...` 會整包跳過，不會假裝通過。
//
// 這一層要回答的是「原版跑到這一步時，那個值是多少」——判準因此是原版自己的
// 變數與呼叫參數，不是從畫面反推的像素。
//
// # 怎麼跑起來
//
// dosgolem 用 `go.work` 從本機副本接進來，**不要在 `go.mod` 加 `require`**：
// 建置容器是 `--network none`，抓不到就整個編不過。`go.work` 是 gitignore 的
// （只在容器裡成立），內容是：
//
//	go 1.24.0
//	use .
//	use /dosgolem
//
// `tools/go_docker.sh` 偵測到 `$DOSGOLEM`（預設 `~/cht/dosgolem-kolrand`）底下有
// `go.mod` 就把它唯讀掛在 `/dosgolem`。
//
// 對拍用的是 dosgolem 的 `kol-oracle` 分支（base 為 `main`）。`Call`、`Stub`、
// 線性版的 `At()` 與 FAT 的 8.3 截斷這四件事**都已經在 main 裡**；`kol-oracle`
// 另外帶著 KOL 專用的三份規格：194（EGA BIOS 預設色盤）、195（DOS 直接讀鍵）、
// 196（BIOS 讀鍵橋接）。
//
// # 對拍一條規則的三步
//
//	寫記憶體設好輸入 → 釘死骰值 → 呼叫常式讀結果
//
//	o := prepare(t)                              // 跑到 _main，對好 DGROUP
//	o.WriteU16(o.DS(actorFatigue), 15)           // 輸入
//	o.StubValue(o.IDA(randEnter), 7)             // 骰值釘死（不擲骰的規則免）
//	o.Call(o.IDA(fatigueRatioFn), 0)             // 呼叫原版的規則
//	got := o.Word(o.IDA(fatigueRatioOut))        // 讀結果，比對 remake
//
// **不必等遊戲走到那一步，也不必讓兩邊的亂數序列一樣。** 種子、抽取順序、抽幾次
// 兩邊都不同，強求同步是在解錯的題目——釘死骰值之後，剩下的差異就只剩規則本身。
//
// 每一支對拍都要有**負對照**：至少一組輸入，常式沒跑就會紅。不然綠燈可能只是
// 「記憶體沒被改，而 remake 剛好也算出同一個數」。
//
// # 這一層驗到什麼、沒驗到什麼
//
// 驗的是**常式的實作**：同一組輸入、同一個骰值，兩邊算出同一個數。
//
// **呼叫時機沒有驗過**——誰在什麼時候呼叫、呼叫幾次、抽幾次骰，要能驅動原版走到
// 戰鬥才問得到。那是 dosgolem 第四層（`apps/kol/`）的事，也是收據基準重生的前置。
package oracle
