#!/usr/bin/env python3
"""在 KOLTOWN 解包後入口停下，擷取 DOS 執行期載入映像。

只使用 DOSBox-X 原生除錯橋接器的唯讀 API；輸出原始記憶體只留在被
`.gitignore` 排除的 `workplace/`。位址換算依 `docs/re/address-spaces.md`。
"""

import hashlib
import json
import socket
import sys
import time


HOST = "127.0.0.1"
PORT = 9876
PACKED_CS = 0x1F51
UNPACKED_CS = 0x1B8F
UNPACKED_IP = 0x001C
LOAD_IMAGE_SIZE = 138528 - 0x240
PACKED_MARKER = b"Packed file is corrupt"
PACKED_MARKER_LOAD_OFFSET = 128551


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
    bridge.request("execution.continue")
    time.sleep(3)
    stopped = bridge.request("execution.pause")

    conventional = bytearray()
    for physical in range(0, 0xA0000, 4096):
        result = bridge.request(
            "memory.read", address=f"{physical // 16:04X}:0000", length=4096
        )
        conventional.extend(int(value, 16) for value in result["bytes"])
    marker_physical = conventional.find(PACKED_MARKER)
    if marker_physical < 0:
        raise RuntimeError("DOS 記憶體中找不到 EXEPACK marker；程式可能尚未載入或已退出")
    load_physical = marker_physical - PACKED_MARKER_LOAD_OFFSET
    if load_physical < 0 or load_physical % 16:
        raise RuntimeError(f"由 marker 推得的載入基底無效：{load_physical:#x}")
    load_segment = load_physical // 16
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
        "packed_marker_physical": f"{marker_physical:05X}",
        "packed_marker_load_offset": PACKED_MARKER_LOAD_OFFSET,
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
