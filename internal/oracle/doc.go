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
// ⚠ **要 dosgolem 的 `kol-force-call` 分支**（`ce78896`…`1e204ce`）。三件事都還沒
// 進 dosgolem 主線：`Call`（接主線會編不過）、線性版的 `At()`（**編得過但永遠不
// 成立**，只會安靜地跑滿指令上限）、FAT 的 8.3 截斷（原版開 `steedpics` 會失敗，
// 而它不檢查開檔結果，會一路跑進沒有映射的記憶體）。
//
// # 這一層驗到什麼、沒驗到什麼
//
// `Call` 是壓好種子直接呼叫原版的常式，走的是同一段機器碼、同一份全域狀態，
// 所以驗的是**常式的實作**。**呼叫時機沒有驗過**——誰在什麼時候抽、抽幾次，
// 要能驅動原版走到戰鬥才問得到，那是 dosgolem 第四層（`apps/kol/`）的事。
package oracle
