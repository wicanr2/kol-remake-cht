#!/usr/bin/env python3
"""純標準函式庫的最小 X11 客戶端，用 XTEST 對 Xvfb 送出鍵盤事件。

容器映像沒有 xdotool／xdo，而 DOSBox-X 的 `AUTOTYPE` 只能在啟動遊戲前一次排定
固定時序的按鍵；原版選單的載入時間會隨磁碟解碼變動，盲打時序不可靠。此模組直接
連上 Xvfb 的 unix socket，查出 keysym→keycode 對照後以 XTEST FakeInput 送鍵，
讓擷取器可以「等畫面靜止 → 送下一鍵」。

只實作需要的請求：連線握手、QueryExtension、GetKeyboardMapping、QueryTree、
SetInputFocus、XTestFakeInput。
"""

from __future__ import annotations

import socket
import struct
import time

KEY_PRESS = 2
KEY_RELEASE = 3

# 需要用到的 keysym（X11 keysymdef 常數）。
KEYSYMS = {
    "Return": 0xFF0D,
    "Escape": 0xFF1B,
    "BackSpace": 0xFF08,
    "Up": 0xFF52,
    "Down": 0xFF54,
    "Left": 0xFF51,
    "Right": 0xFF53,
    "space": 0x0020,
}
for _c in "abcdefghijklmnopqrstuvwxyz":
    KEYSYMS[_c] = ord(_c)
for _c in "ABCDEFGHIJKLMNOPQRSTUVWXYZ":
    KEYSYMS[_c] = ord(_c)
for _c in "0123456789":
    KEYSYMS[_c] = ord(_c)


def _pad(n: int) -> int:
    return (4 - n % 4) % 4


class X11Client:
    def __init__(self, display: str = ":77"):
        num = display.lstrip(":").split(".")[0]
        self.sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        self.sock.connect(f"/tmp/.X11-unix/X{num}")
        self.sock.settimeout(20)     # 寧可明確失敗，也不要無限期 block
        self.seq = 0
        self._handshake()
        self.xtest_opcode = self._query_extension("XTEST")
        if self.xtest_opcode is None:
            raise RuntimeError("X server 沒有 XTEST 擴充")
        self.keycodes = self._keyboard_mapping()

    # -- 底層 -------------------------------------------------------------
    def _recv_exact(self, n: int) -> bytes:
        buf = b""
        while len(buf) < n:
            chunk = self.sock.recv(n - len(buf))
            if not chunk:
                raise RuntimeError("X 連線中斷")
            buf += chunk
        return buf

    def _request(self, data: bytes, expect_reply: bool) -> bytes | None:
        self.seq = (self.seq + 1) & 0xFFFF
        seq = self.seq
        self.sock.sendall(data)
        if not expect_reply:
            return None
        # X server 會主動送出 32-byte event／error 封包（例如 MappingNotify）。
        # 這些封包沒有 reply 的長度欄位，若當成 reply 解析就會要求讀取任意長度而
        # 永久 block；因此逐包檢查第一個 byte，只有 1 才是 reply。
        while True:
            head = self._recv_exact(32)
            kind = head[0]
            if kind == 0:
                err_seq = struct.unpack_from("<H", head, 2)[0]
                if err_seq == seq:
                    raise RuntimeError(f"X 錯誤：code={head[1]} req={head[10]}")
                continue
            if kind == 1:
                extra = struct.unpack_from("<I", head, 4)[0] * 4
                return head + (self._recv_exact(extra) if extra else b"")
            # kind >= 2：事件，丟棄後繼續等 reply。

    def _handshake(self) -> None:
        req = struct.pack("<BBHHHHH", 0x6C, 0, 11, 0, 0, 0, 0)
        self.sock.sendall(req)
        head = self._recv_exact(8)
        if head[0] != 1:
            reason_len = head[1]
            raise RuntimeError(f"X 握手失敗：{self._recv_exact(reason_len)!r}")
        length = struct.unpack_from("<H", head, 6)[0] * 4
        body = self._recv_exact(length)
        (release, self.id_base, self.id_mask, motion_size, vendor_len,
         max_req, num_screens, num_formats) = struct.unpack_from("<IIIIHHBB", body, 0)
        self.min_keycode = body[26]
        self.max_keycode = body[27]
        off = 32 + vendor_len + _pad(vendor_len) + 8 * num_formats
        self.root = struct.unpack_from("<I", body, off)[0]

    def _query_extension(self, name: str) -> int | None:
        raw = name.encode()
        n = len(raw)
        req = struct.pack("<BBHHH", 98, 0, 2 + (n + _pad(n)) // 4, n, 0) + raw + b"\x00" * _pad(n)
        reply = self._request(req, True)
        present, major = reply[8], reply[9]
        return major if present else None

    def _keyboard_mapping(self) -> dict[int, int]:
        count = self.max_keycode - self.min_keycode + 1
        req = struct.pack("<BBHBBH", 101, 0, 2, self.min_keycode, count, 0)
        reply = self._request(req, True)
        per = reply[1]
        body = reply[32:]
        table: dict[int, int] = {}
        for i in range(count):
            for j in range(per):
                sym = struct.unpack_from("<I", body, (i * per + j) * 4)[0]
                if sym and sym not in table:
                    table[sym] = self.min_keycode + i
        return table

    # -- 對外 -------------------------------------------------------------
    def query_children(self) -> list[int]:
        req = struct.pack("<BBHI", 15, 0, 2, self.root)
        reply = self._request(req, True)
        n = struct.unpack_from("<H", reply, 16)[0]
        return list(struct.unpack_from(f"<{n}I", reply, 32))

    def get_geometry(self, window: int) -> dict:
        reply = self._request(struct.pack("<BBHI", 14, 0, 2, window), True)
        x, y, w, h = struct.unpack_from("<hhHH", reply, 12)
        return {"x": x, "y": y, "width": w, "height": h}

    def set_input_focus(self, window: int) -> None:
        self._request(struct.pack("<BBHII", 42, 1, 3, window, 0), False)

    def fake_key(self, keysym_name: str, hold_ms: int = 60) -> None:
        if keysym_name not in KEYSYMS:
            raise KeyError(f"未定義的 keysym 名稱：{keysym_name}")
        sym = KEYSYMS[keysym_name]
        keycode = self.keycodes.get(sym)
        shift = False
        if keycode is None and 0x41 <= sym <= 0x5A:      # 大寫字母走 Shift + 小寫
            keycode = self.keycodes.get(sym + 0x20)
            shift = True
        if keycode is None:
            raise KeyError(f"keysym {keysym_name} 沒有對應 keycode")
        shift_code = self.keycodes.get(0xFFE1)
        if shift and shift_code:
            self._fake(KEY_PRESS, shift_code)
        self._fake(KEY_PRESS, keycode)
        time.sleep(hold_ms / 1000.0)
        self._fake(KEY_RELEASE, keycode)
        if shift and shift_code:
            self._fake(KEY_RELEASE, shift_code)

    def _fake(self, event_type: int, detail: int) -> None:
        req = struct.pack("<BBHBBHIIQhhQ", self.xtest_opcode, 2, 9,
                          event_type, detail, 0, 0, 0, 0, 0, 0, 0)
        assert len(req) == 36
        self._request(req, False)

    def close(self) -> None:
        try:
            self.sock.close()
        except OSError:
            pass


if __name__ == "__main__":
    import sys
    c = X11Client(sys.argv[1] if len(sys.argv) > 1 else ":77")
    print(f"root=0x{c.root:x} xtest_opcode={c.xtest_opcode} keycodes={len(c.keycodes)}")
    for name in sys.argv[2:]:
        c.fake_key(name)
        time.sleep(0.2)
