// Package oracle 是拿 dosgolem 當原版 oracle 的那一層（RE-132）。
//
// dosgolem 是無頭、決定性、可以 import 的 DOS 執行器，用 `go.work` 從本機副本接
// 進來（`tools/go_docker.sh` 偵測到旁邊有一份就唯讀掛在 `/dosgolem`）。這一包底下
// 的對拍測試都在 `-tags oracle` 之後，沒有那份掛載的人跑 `go test ./...` 會整包
// 跳過，不會假裝通過。
//
// 這一層要回答的是「原版跑到這一步時，那個值是多少」——判準因此是原版自己的
// 變數與呼叫參數，不是從畫面反推的像素。
package oracle
