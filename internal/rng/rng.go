// Package rng 是**原版那一支亂數**（RE-132）。
//
// 原版用的是 C 函式庫的 `rand`，常數與取出的位元逐條照抄
// （KOLTOWN-unpacked 的 `sub_2D538`、`sub_2D526`、`sub_1B5A8`）：
//
//	srand(n): 種子 = n                                   ; 高 word 歸零
//	rand():   種子 = 種子 × 0x000343FD + 0x00269EC3      ; 32 位元環繞
//	          回 (種子 >> 16) & 0x7FFF
//	rand(n) = rand() % n + 1                             ; 回 1…n
//
// 兩邊用同一組常數，**同一個種子就會擲出同一串數**。這是拿原版當 oracle 對拍的
// 前提：不然「remake 走這一步時擲到什麼」與原版永遠對不起來，只能靠固定按鍵序列
// 凍結結果，而那個序列每次改動戰鬥數字就作廢。
//
// 原版自己是拿 BIOS tick 播種（`_main:0x10010` 的 `int 1Ah`，低 word 進
// `word_349AC` 再餵給 `srand`），落在 0…65535——與 remake 的 `-seed` 同一個空間。
package rng

// Mul／Add／Mask 是原版的三個常數（`0x2D538`…`0x2D55C`）。
const (
	Mul  = 0x000343FD
	Add  = 0x00269EC3
	Mask = 0x7FFF
)

// LCG 是那一支產生器。零值等於 `srand(0)`。
type LCG struct{ state uint32 }

// New 等於原版的 `srand(seed)`：種子直接放進去，高 word 歸零。
func New(seed int) *LCG { return &LCG{state: uint32(uint16(seed))} }

// NewState 從一個完整的 32 位元狀態接下去——對拍時要從原版某一刻的種子接手才用得到。
func NewState(state uint32) *LCG { return &LCG{state: state} }

// State 回目前的 32 位元種子，對得上原版的 `word_2F4D0`／`word_2F4D2`。
func (r *LCG) State() uint32 { return r.state }

// Next 是原版的 `rand()`：回 0…0x7FFF。
func (r *LCG) Next() int {
	r.state = r.state*Mul + Add
	return int(r.state>>16) & Mask
}

// Roll 是原版的 `sub_1B5A8(n)`：`rand() % n + 1`，回 1…n。
// n 不是正的就回 0 且**不動種子**——原版走不到那一條（呼叫端都是常數）。
func (r *LCG) Roll(n int) int {
	if n <= 0 {
		return 0
	}
	return r.Next()%n + 1
}

// Choose 是 `rand() % n`，回 0…n−1。標題序列的二選一用它。
func (r *LCG) Choose(n int) int {
	if n <= 0 {
		return 0
	}
	return r.Next() % n
}
