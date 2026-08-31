#!/usr/bin/env python3
"""在 KOLTOWN 解包後入口停下，擷取 DOS 執行期載入映像。

只使用 DOSBox-X 原生除錯橋接器的唯讀 API；輸出原始記憶體只留在被
`.gitignore` 排除的 `workplace/`。位址換算依 `docs/re/address-spaces.md`。
"""

import hashlib
import json
import os
import socket
import sys
import time


HOST = "127.0.0.1"
PORT = 9876
UNPACKED_CS = 0x1B8F
UNPACKED_IP = 0x001C
LOAD_IMAGE_SIZE = 138528 - 0x2400
RUNTIME_MARKER = b"MS Run-Time Library - Copyright (c) 1989, Microsoft Corp"
RUNTIME_MARKER_LOAD_OFFSET = 0x1DED8


class Bridge:
    def __init__(self):
        self.sock = socket.create_connection((HOST, PORT), timeout=10)
        self.buf = b""
        self.seq = 0

    def request(self, method, **params):
        self.seq += 1
        request = {"id": self.seq, "method": method, "params": params}
        self.sock.sendall((json.dumps(request) + "\n").encode())
        while b"\n" not in self.buf:
            chunk = self.sock.recv(65536)
            if not chunk:
                raise RuntimeError("DOSBox-X bridge 關閉連線")
            self.buf += chunk
        line, self.buf = self.buf.split(b"\n", 1)
        response = json.loads(line)
        if not response.get("ok"):
            error = response.get("error", {})
            raise RuntimeError(f"{error.get('code')}: {error.get('message')}")
        return response.get("result")


def hex_word(value):
    if isinstance(value, int):
        return value
    return int(value, 16)


def main():
    if len(sys.argv) != 3:
        raise SystemExit("用法：koltown_runtime_oracle.py OUTPUT.bin RECEIPT.json")
    output_path, receipt_path = sys.argv[1:]
    bridge = Bridge()

    initial = bridge.request("debug.status")
    if not initial.get("stopped"):
        raise RuntimeError("啟動時除錯器未停止，不能建立載入段正對照")
    expected_load_text = os.environ.get("KOL_ORACLE_EXPECT_LOAD_SEGMENT")
    breakpoint_result = None
    if expected_load_text:
        expected_load = int(expected_load_text, 16)
        target = f"{expected_load + UNPACKED_CS:04X}:{UNPACKED_IP:04X}"
        breakpoint_result = bridge.request("breakpoint.set", address=target)
        bridge.request("execution.continue")
        deadline = time.time() + 60
        stopped = None
        while time.time() < deadline:
            state = bridge.request("debug.status")
            if state.get("stopped"):
                stopped = state
                break
            time.sleep(0.1)
        if stopped is None:
            raise RuntimeError(f"60 秒內未命中解包後入口 {target}")
        if stopped["location"] != {"cs": target[:4], "eip": target[5:]}:
            raise RuntimeError(f"停止位置不是解包後入口：{stopped['location']} != {target}")
        delay = None
    else:
        bridge.request("execution.continue")
        delay = float(os.environ.get("KOL_ORACLE_DELAY", "25"))
        time.sleep(delay)
        stopped = bridge.request("execution.pause")

    conventional = bytearray()
    for physical in range(0, 0xA0000, 4096):
        result = bridge.request(
            "memory.read", address=f"{physical // 16:04X}:0000", length=4096
        )
        conventional.extend(int(value, 16) for value in result["bytes"])
    marker_physical = conventional.find(RUNTIME_MARKER)
    if marker_physical < 0:
        raise RuntimeError("DOS 記憶體中找不到 Microsoft runtime marker；程式可能已退出")
    load_physical = marker_physical - RUNTIME_MARKER_LOAD_OFFSET
    if load_physical < 0 or load_physical % 16:
        raise RuntimeError(f"由 marker 推得的載入基底無效：{load_physical:#x}")
    load_segment = load_physical // 16
    if expected_load_text and load_segment != expected_load:
        raise RuntimeError(
            f"runtime marker 推得 load segment {load_segment:04X}，不等於斷點預期 {expected_load:04X}"
        )
    target_segment = load_segment + UNPACKED_CS
    target = f"{target_segment:04X}:{UNPACKED_IP:04X}"

    image = bytearray()
    for linear_offset in range(0, LOAD_IMAGE_SIZE, 4096):
        length = min(4096, LOAD_IMAGE_SIZE - linear_offset)
        segment = load_segment + linear_offset // 16
        offset = linear_offset % 16
        result = bridge.request(
            "memory.read", address=f"{segment:04X}:{offset:04X}", length=length
        )
        image.extend(int(value, 16) for value in result["bytes"])

    with open(output_path, "wb") as output:
        output.write(image)
    receipt = {
        "schema": 1,
        "initial_stop": initial["location"],
        "runtime_stop": stopped["location"],
        "continue_delay_seconds": delay,
        "entry_breakpoint_result": breakpoint_result,
        "runtime_marker_physical": f"{marker_physical:05X}",
        "runtime_marker_load_offset": RUNTIME_MARKER_LOAD_OFFSET,
        "load_segment": f"{load_segment:04X}",
        "unpacked_entry": target,
        "load_image_size": len(image),
        "load_image_sha256": hashlib.sha256(image).hexdigest(),
    }
    with open(receipt_path, "w", encoding="utf-8") as output:
        json.dump(receipt, output, ensure_ascii=False, indent=2)
        output.write("\n")
    print(json.dumps(receipt, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
