package main

import "testing"

type fixedRand int

func (f fixedRand) Roll(int) int { return int(f) }

func TestScriptedRollsThenFallback(t *testing.T) {
	r, err := parseScriptedRolls("1, 7")
	if err != nil {
		t.Fatal(err)
	}
	r.next = fixedRand(42)
	for i, want := range []int{1, 7, 42, 42} {
		if got := r.Roll(100); got != want {
			t.Fatalf("第 %d 次：期望 %d，實際 %d", i, want, got)
		}
	}
}

func TestScriptedRollsRejectBadInput(t *testing.T) {
	for _, text := range []string{"0", "x", "1,,2"} {
		if _, err := parseScriptedRolls(text); err == nil {
			t.Fatalf("%q 應該被拒絕", text)
		}
	}
	if r, err := parseScriptedRolls(""); r != nil || err != nil {
		t.Fatal("空字串應該是不啟用")
	}
	r, _ := parseScriptedRolls("5")
	defer func() {
		if recover() == nil {
			t.Fatal("骰值超出骰面應該停下")
		}
	}()
	r.Roll(3)
}
