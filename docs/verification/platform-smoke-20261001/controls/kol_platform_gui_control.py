"""在隔離容器透過 X11 真實按鍵驗證互動存檔，不修改遊戲狀態。"""
from pathlib import Path
import sys, os, json, struct, time, re
sys.path.insert(0, '/work/tools')
from x11_key_inject import X11Client, KEY_PRESS, KEY_RELEASE
from png_io import write_png_rgb

root = Path(sys.argv[1])
script = Path(sys.argv[2])
log = root / 'engine.log'
assert (root.stat().st_uid, root.stat().st_gid) == (os.getuid(), os.getgid())
keys = {'SPACE':'space', 'RET':'Return', 'ESC':'Escape', 'BS':'BackSpace',
        'L':'Left','R':'Right','U':'Up','D':'Down','<':'comma','>':'period'}
steps = {0,2,103,104,111,116,117,120,126,137,138,198}
if 'reload' in root.name:
    steps = {0,1,3,14,15}
deadline = time.monotonic() + 25
client = None
while time.monotonic() < deadline:
    try:
        client = X11Client(os.environ['DISPLAY'])
        break
    except (FileNotFoundError, ConnectionRefusedError):
        time.sleep(.1)
assert client is not None, 'X11 未就緒'
window = None
while time.monotonic() < deadline:
    for candidate in client.query_children():
        g = client.get_geometry(candidate)
        if g['width'] == 960 and g['height'] == 600:
            window = candidate
            break
    if window is not None:
        break
    time.sleep(.1)
assert window is not None, '960×600 互動視窗未出現'
client.set_input_focus(window)
time.sleep(1)

def capture(step):
    req = struct.pack('<BBHIhhHHI', 73, 2, 5, window, 0, 0, 960, 600, 0xffffffff)
    reply = client._request(req, True)
    assert reply[1] == 24, f'不支援的 X11 深度 {reply[1]}'
    data = reply[32:]
    assert len(data) == 960*600*4
    rgb = bytearray(960*600*3)
    rgb[0::3] = data[2::4]
    rgb[1::3] = data[1::4]
    rgb[2::3] = data[0::4]
    return write_png_rgb(root/f'step-{step:03}.png', 960, 600, bytes(rgb))

def inputs():
    return [line for line in log.read_text(errors='replace').splitlines() if line.startswith('input scene=')]

tokens = script.read_text().split()
records = []
try:
    if 0 in steps:
        capture(0)
    for step, token in enumerate(tokens, 1):
        before = len(inputs())
        client.fake_key(keys.get(token, token.lower()), hold_ms=120)
        until = time.monotonic() + 6
        while time.monotonic() < until and len(inputs()) == before:
            time.sleep(.05)
        observed = inputs()
        assert len(observed) == before+1, (step, token, before, len(observed))
        assert f'token=[{token}]' in observed[-1], (step, token, observed[-1])
        time.sleep(.18)
        rec = {'step':step, 'token':token, 'input_log':observed[-1]}
        if step in steps:
            rec['png_sha256'] = capture(step)
        records.append(rec)
        if step % 50 == 0:
            print(f'互動按鍵 {step}/{len(tokens)} 已由程式確認', flush=True)
    (root/'inputs.json').write_text(json.dumps({'script':str(script),'tokens':tokens,'records':records,'window':window,'geometry':client.get_geometry(window)},ensure_ascii=False,indent=2)+'\n')
    (root/'controller-done').write_text('成功\n')
    ctrl = client.keycodes[0xffe3]
    client._fake(KEY_PRESS, ctrl)
    try:
        client.fake_key('q', 120)
        client._fake(KEY_RELEASE, ctrl)
    except BrokenPipeError:
        # 程式收到 Ctrl+Q 後會退出；擁有容器將以 wait 的實際退出碼驗收。
        (root/'quit-disconnected.txt').write_text('Ctrl+Q 按下後 X11 擁有程序已關閉；另核實遊戲退出碼。\n')
    print(f'{len(tokens)} 個互動按鍵均確認，已送出 Ctrl+Q', flush=True)
finally:
    client.close()
