package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
)

// scriptedRolls 是測試用的受控亂數（CLAUDE.md「RND() 固定種子對拍鐵則」）：依序吐出
// 指定的骰值，用完就交回種子亂數。只接在野外遭遇的擲點與騎走判定上，用來把「這一次時段變換
// 擲到治療」這種要驗的規則隔離出來；正式遊戲不帶 -ambush-rolls 就不會經過這裡。
type scriptedRolls struct {
	values []int
	next   battle.Rand
}

func parseScriptedRolls(text string) (*scriptedRolls, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}
	out := &scriptedRolls{}
	for _, field := range strings.Split(text, ",") {
		v, err := strconv.Atoi(strings.TrimSpace(field))
		if err != nil || v < 1 {
			return nil, fmt.Errorf("-ambush-rolls 的骰值要是正整數：%q", field)
		}
		out.values = append(out.values, v)
	}
	return out, nil
}

// Roll 回傳 1…n。指定的骰值超出這一次的骰面就是腳本寫錯，直接停下，不偷偷夾限。
func (r *scriptedRolls) Roll(n int) int {
	if len(r.values) == 0 {
		return r.next.Roll(n)
	}
	v := r.values[0]
	r.values = r.values[1:]
	if v > n {
		panic(fmt.Sprintf("-ambush-rolls 的骰值 %d 超出這一次的骰面 %d", v, n))
	}
	return v
}
