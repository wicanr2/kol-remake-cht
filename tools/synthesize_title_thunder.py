#!/usr/bin/env python3
"""依 TITLE.EXE 的 `sub_10F75` 合成標題雷聲，輸出 16-bit mono WAV。

這是**規格近似合成**，不是原版錄音：波形由公開的 PC Speaker／PIT channel 2 規格與
`sub_10F75` 已證實的演算法產生，檔案裡沒有任何原版音訊資料。使用者於 2026-09-01
決定 TITLE 音訊採此路線。

已證實的演算法（RE-047 與 `docs/re/ida-fingerprints/TITLE-speaker-summary.json`）：

- 兩段音爆。每段先開喇叭（`0x10f33`：`in al,61h` / `or al,3` / `out 61h,al`），
  更新 100 次 PIT channel 2 除數（`0x10f5d`：低位再高位寫入埠 `0x42`），最後關喇叭
  （`0x10f48`：`and al,0FCh`）。
- 每一步的除數是 `500 × 段次 + rand()%10000 + 1 + 1000`，因此第一段落在 1001–11000、
  第二段落在 1501–11500；方波頻率為 `1193182 / 除數`。
- 兩段之間有 50% 機率多等一個單位（`sub_11904`）。

**時間長度不能靜態讀出**：步進間隔是用開機期校正值（`sub_118D1` 數的是「一個 BIOS tick
內跑得完幾圈」）算出來的忙等，而校正迴圈與延遲迴圈的指令不同，換一台機器就換一個長度。
因此總長度取自 `docs/re/title-audio-oracle.json` 的 DOSBox-X 擷取量測值，並標為
`hardware-spec approximation`。

  python3 tools/synthesize_title_thunder.py --out title-thunder.wav
"""

from __future__ import annotations

import argparse
import json
import math
import pathlib
import struct
import sys
import wave

PIT_HZ = 1193182.0
STEPS_PER_BURST = 100
BURSTS = 2
DIVISOR_BASE = 1000
DIVISOR_SPAN = 10000
DIVISOR_BURST_STEP = 500
TICK_SECONDS = 65536.0 / 1193182.0  # 一個 BIOS tick


class LCG:
    """可重播的亂數來源。原版用 C 的 rand()，種子是開機時的 BIOS tick；

    remake 不可能重現那個種子，也不宣稱頻率序列與原版相同——只宣稱分布與結構相同。
    """

    def __init__(self, seed: int) -> None:
        self.state = seed & 0xFFFFFFFF

    def next(self) -> int:
        self.state = (self.state * 1664525 + 1013904223) & 0xFFFFFFFF
        return self.state >> 16

    def below(self, n: int) -> int:
        return self.next() % n


def synthesize(seed: int, rate: int, cue_seconds: float, amplitude: float
               ) -> tuple[bytes, list[dict]]:
    rng = LCG(seed)
    step_seconds = cue_seconds / (BURSTS * STEPS_PER_BURST)
    samples: list[int] = []
    steps: list[dict] = []
    phase = 0.0
    peak = int(round(amplitude * 32767))
    for burst in range(BURSTS):
        for _ in range(STEPS_PER_BURST):
            divisor = DIVISOR_BURST_STEP * burst + rng.below(DIVISOR_SPAN) + 1 + DIVISOR_BASE
            frequency = PIT_HZ / divisor
            count = int(round(step_seconds * rate))
            increment = frequency / rate
            for _ in range(count):
                samples.append(peak if phase < 0.5 else -peak)
                phase = (phase + increment) % 1.0
            steps.append({"burst": burst, "divisor": divisor,
                          "frequency_hz": round(frequency, 3), "samples": count})
        if burst + 1 < BURSTS:
            # 兩段之間喇叭關閉；原版有 50% 機率多等一個單位。
            pause = TICK_SECONDS if rng.below(2) == 1 else 0.0
            silence = int(round(pause * rate))
            samples.extend([0] * silence)
            steps.append({"burst": burst, "gap_samples": silence})
    return struct.pack("<%dh" % len(samples), *samples), steps


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--out", required=True)
    p.add_argument("--steps-out", help="把逐步除數與頻率寫成 JSON")
    p.add_argument("--seed", type=int, default=1)
    p.add_argument("--sample-rate", type=int, default=44100)
    p.add_argument("--cue-seconds", type=float, default=0.961,
                   help="一次 sub_10F75 的總長度；預設是 DOSBox-X 擷取四段的平均值")
    p.add_argument("--peak-db", type=float, default=-10.317424,
                   help="方波峰值，預設對齊 DOSBox-X 擷取的量測峰值")
    args = p.parse_args()

    amplitude = 10.0 ** (args.peak_db / 20.0)
    if not 0.0 < amplitude <= 1.0:
        raise SystemExit("peak-db 必須小於等於 0")
    frames, steps = synthesize(args.seed, args.sample_rate, args.cue_seconds, amplitude)

    out = pathlib.Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    with wave.open(str(out), "wb") as handle:
        handle.setnchannels(1)
        handle.setsampwidth(2)
        handle.setframerate(args.sample_rate)
        handle.writeframes(frames)
    if args.steps_out:
        pathlib.Path(args.steps_out).write_text(json.dumps({
            "schema": "kol-title-thunder-steps-v1",
            "source_function": "sub_10F75", "source_ea": "0x10f75",
            "seed": args.seed, "sample_rate_hz": args.sample_rate,
            "cue_seconds": args.cue_seconds, "peak_db": args.peak_db,
            "steps": steps}, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps({
        "out": str(out), "frames": len(frames) // 2,
        "seconds": round(len(frames) / 2 / args.sample_rate, 6),
        "distinct_divisors": len({s["divisor"] for s in steps if "divisor" in s}),
        "frequency_hz_range": [min(s["frequency_hz"] for s in steps if "frequency_hz" in s),
                               max(s["frequency_hz"] for s in steps if "frequency_hz" in s)],
    }, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
