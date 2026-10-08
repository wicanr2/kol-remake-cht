#!/usr/bin/env python3
"""逐步生成一條打完戰鬥的按鍵腳本。

做法：跑一次腳本 → 讀 `-trace` 的最後一筆狀態 → 依規劃段停在哪一層決定下一批
token → 再跑。**產出的是一串固定按鍵**，之後用一般的 `-flow-script` 重播就好，
遊戲本身不需要任何自動化旗標——收據仍然是「按鍵走出來的」。

改了戰鬥數值之後長路徑收據的戰鬥那一段就作廢了（打幾個回合、每一手的訊息都不同），
這一支是用來重生那一段的。走路與城裡那幾段照舊沿用原腳本的前後綴。

  KOL_DRIVE_DIR=workplace/tmp/drive-a \
  tools/drive_battle_script.py <seed> <存檔> <起始 token 檔> <輸出> [輪數] \
      [世界目標 X,Y] [城鎮目標 X,Y] [城鎮邊界 loX,hiX,loY,hiY]

`KOL_QUEST_SITES` 指定換過的任務地點資料檔；`KOL_DRIVE_DIR` 是這一次的暫存目錄
（同時跑好幾條時要各給一個，否則會互相覆寫）。
`KOL_STOP_AFTER_BATTLE=1` 讓腳本在第一場戰鬥離場時停止；勝負仍須讀
`VerdictLost=1`／`VerdictWon=2`／`VerdictFled=3`，不能把離場當成勝利。
存檔給 `-` 表示不讀檔，從標題走（`-scene flow`）。
`KOL_DRIVE_QUEST=1` 是任務地點的「撿物逃走」策略（SPEC-014 §5）：離任務物品最近的
隊員走去撿，撿到後走到最近的戰場邊，逃走問句答 y；其餘隊員護送持物者。找路讀軌跡的
`walls`（要 `-trace-detail`，每一場記一次）。

游標移到目標格的序列是「先歸位再走」：`U U L L` 之後一定停在 (3,3)，
接著 `D`×(dx+1)、`R`×(dy+1) 就到相對位移 (dx, dy) 那一格——先垂直再水平不會
撞到中央那一格的跳過規則。
"""
import collections
import json
import os
import pathlib
import subprocess
import sys

REPO = pathlib.Path('.').resolve()
SPEC = pathlib.Path(os.environ.get('KOL_DRIVE_DIR', 'workplace/tmp/drive'))
SITES = os.environ.get('KOL_QUEST_SITES', '')
SPEC.mkdir(parents=True, exist_ok=True)

CMD_YES, CMD_WEAPON, CMD_UNARMED, CMD_MOVE, CMD_REST = 0x07, 0x21, 0x29, 0x27, 0x05
CMD_MISSILE = 0x23
CMD_PICKUP = 0x26
# KOL_DRIVE_QUEST=1：任務地點那一場改走「搶任務物品、走到戰場邊、逃走」（SPEC-014 §5）。
# 離任務物品最近的那一名去撿，撿到之後往最近的戰場邊走；其他人照舊迎敵。
QUEST_MODE = os.environ.get('KOL_DRIVE_QUEST') == '1'
EDGE_LO, EDGE_HI = 4, 0x3B
ICON_MOVE_WALK = 0x3C


# KOL_BIN 有值就直接呼叫那個執行檔（`tools/run_driver.sh` 會在容器裡設好它與
# DISPLAY）。沒有的話照舊每一輪開一個容器——那是主要成本：一輪要花三秒開容器，
# 而一場戰鬥要跑上百輪。
KOL_BIN = os.environ.get('KOL_BIN', '')
KOL_SCENE = os.environ.get('KOL_SCENE', 'flow')
KOL_SCRIPT_FLAG = os.environ.get('KOL_SCRIPT_FLAG', '-flow-script')
KOL_MONSTERS = os.environ.get('KOL_MONSTERS', '')
# 受控亂數（plan 的 `ambush_rolls`）：重生時要與重播時帶同一組，否則遇到的不是同一場。
KOL_AMBUSH_ROLLS = os.environ.get('KOL_AMBUSH_ROLLS', '')


def run(tokens, seed, save):
    args = ['-scene', KOL_SCENE, '-seed', str(seed), '-trace-detail']
    # 存檔給 `-` 表示不讀檔：從標題／書本一路走的正常流程（`-scene flow`）。
    if save != '-':
        args += ['-load', save]
    if SITES:
        args += ['-quest-sites', SITES]
    if KOL_MONSTERS:
        args += ['-monsters', KOL_MONSTERS]
    if KOL_AMBUSH_ROLLS:
        args += ['-ambush-rolls', KOL_AMBUSH_ROLLS]
    args += [KOL_SCRIPT_FLAG, ' '.join(tokens),
             '-shot', str(SPEC / 'last.png'), '-trace', str(SPEC / 'last.json')]
    if KOL_BIN:
        subprocess.run([KOL_BIN] + args, check=True,
                       stdout=subprocess.DEVNULL, stderr=subprocess.PIPE, text=True)
    else:
        spec = SPEC / 'spec.txt'
        spec.write_text('|'.join([str(SPEC / 'last.png')] + args) + '\n',
                        encoding='utf-8')
        subprocess.run(['tools/run_kol_shots_fast.sh', str(spec)], check=True,
                       stdout=subprocess.DEVNULL)
    steps = json.loads((SPEC / 'last.json').read_text(encoding='utf-8'))['steps']
    row = steps[-1]
    # 地形每一場只記一次（`walls`），這裡把最近那一份接到最後一步上。
    if row.get('battle') is not None and 'walls' not in row['battle']:
        for prev in reversed(steps):
            walls = (prev.get('battle') or {}).get('walls')
            if walls:
                row['battle']['walls'] = walls
                break
    return row


def parse_roster(lines):
    """軌跡把參戰者壓成一行一名，這裡解回來。名字可能有空白（Fighting Bull）。"""
    out = []
    for line in lines:
        p = line.split()
        down = p[-1] == 'down'
        if down:
            p = p[:-1]
        chest = int(p[-1].split('=')[1])
        head = int(p[-2].split('=')[1])
        x, y = (int(v) for v in p[-3].split(','))
        # 競技場那一場沒上場的人標成 watch（記錄 +0x01 = 2），既不是隊員
        # 也不是敵人——側別給 2，下面挑目標與分派時自然排除。
        side = {'party': 0, 'foe': 1}.get(p[-4], 2)
        out.append({'i': int(p[0]), 'name': ' '.join(p[1:-4]), 'side': side,
                    'x': x, 'y': y, 'head': head, 'chest': chest, 'down': down})
    return out


def to_cell(dx, dy):
    return ['U', 'U', 'L', 'L'] + ['D'] * (dx + 1) + ['R'] * (dy + 1)


def to_ranged_cell(dx, dy):
    """射擊游標歸位後沿外側前往目標，避開中央 3×3 禁區（SPEC-010）。"""
    row, col = dx + 4, dy + 4
    if not (0 <= row <= 8 and 0 <= col <= 8) or (3 <= row <= 5 and 3 <= col <= 5):
        return None
    # 原版允許記憶目標的初始座標在視窗外；第一鍵先夾回 0…8，再走八格歸零。
    reset = ['U'] * 9 + ['L'] * 9
    if 3 <= row <= 5:
        return reset + ['R'] * col + ['D'] * row
    return reset + ['D'] * row + ['R'] * col


def pick(bar, code):
    return ['>'] * bar.index(code) + ['RET']


DIRS = {(-1, 0): 'L', (1, 0): 'R', (0, -1): 'U', (0, 1): 'D',
        (-1, -1): 'A', (1, -1): 'S', (-1, 1): 'Z', (1, 1): 'X'}


# 一次補幾個 SPACE。從 1 開始，成功就加倍，補過頭再減半——上限給 8，
# 因為執行段一句一個鍵，一次跳太多會把後面的階段也按過去。
PRESS_MAX = 8


def press(memo):
    n = memo.get('batch', 1)
    memo['last_batch'] = n
    memo['batch'] = min(n * 2, PRESS_MAX)
    return ['SPACE'] * n


def plan_step(here, target, blocked, bounds):
    """在「已知擋住的邊」之外找一條最短路，回傳第一步的方向。

    城鎮地圖沒有從遊戲裡倒出來，所以這裡邊走邊學：沒撞過的邊先當可以走，
    撞到了就記下來重算。走一次就夠了，不需要完整地圖。
    """
    from collections import deque
    lo_x, hi_x, lo_y, hi_y = bounds
    prev = {here: None}
    q = deque([here])
    while q:
        cur = q.popleft()
        if cur == target:
            break
        for d in DIRS:
            if (cur, d) in blocked:
                continue
            nxt = (cur[0] + d[0], cur[1] + d[1])
            if not (lo_x <= nxt[0] <= hi_x and lo_y <= nxt[1] <= hi_y):
                continue
            if nxt in prev:
                continue
            prev[nxt] = (cur, d)
            q.append(nxt)
    if target not in prev:
        return None
    node, path = target, []
    while prev[node] is not None:
        cur, d = prev[node]
        path.append(d)
        node = cur
    return path[-1]


def town_step(tokens, row, target, memo):
    """城裡的一步：朝目標格走。撞到建築就把那條邊記下來重算路線。"""
    t = row['town']
    here = (t['x'], t['y'])
    # 走到地圖邊界就是出城口：選 DOOR 離開。
    if t['mode'] == 3:
        tokens.append('RET')
        return True
    if target is None or here == tuple(target):
        print('走到城裡的目標格了', t)
        return False
    # 踩到別人家門口時方向鍵沒有作用（原版的門口迴圈只吃圖示列），
    # 先按 U-TURN 回地圖再繼續走。
    if t['mode'] == 1:
        tokens += ['>', 'RET']
        return True
    blocked = memo.setdefault('blocked', set())
    if memo.get('pos') == here and memo.get('dir') is not None:
        blocked.add((here, memo['dir']))
    d = plan_step(here, tuple(target), blocked, memo.get('bounds', (48, 63, 0, 15)))
    if d is None:
        print('城裡找不到路了', t, len(blocked), '條邊擋住')
        return False
    memo['pos'], memo['dir'] = here, d
    tokens.append(DIRS[d])
    return True


def world_step(tokens, row, target):
    """戰鬥以外的一步：野外朝目標格走，換日與任務提示各按一個鍵。"""
    w = row.get('world')
    if w is None:
        print('不在野外', row.get('scene'), row.get('town'))
        return False
    if w.get('camping'):
        tokens.append('L')          # CAMPING! 停在原地等一個鍵
        return True
    if w.get('heal_prompt'):
        tokens.append('RET')        # 自動治療提示先收掉，方向鍵此時不會移動
        return True
    if w.get('quest_prompt'):
        tokens.append('RET')        # Will you partake?
        return True
    if target is None:
        print('到野外了', w)
        return False
    tx, ty = target
    if (w['x'], w['y']) == (tx, ty):
        print('走到目標格了', w)
        return False
    dx = (tx > w['x']) - (tx < w['x'])
    dy = (ty > w['y']) - (ty < w['y'])
    tokens.append(DIRS[(dx, dy)])
    return True


def first_step(walls, taken, start, is_goal):
    """在地形（`walls[y][x]`）與站著的人之間找最短路，回第一步的 (dx, dy)。
    站著的人只擋第一步：再遠的人下一回合多半已經走開了。"""
    def free(x, y):
        if not (EDGE_LO <= x <= EDGE_HI and EDGE_LO <= y <= EDGE_HI):
            return False
        return not walls or walls[y][x] != '#'
    prev = {start: None}
    queue = collections.deque([start])
    while queue:
        cur = queue.popleft()
        if cur != start and is_goal(*cur):
            while prev[cur] != start:
                cur = prev[cur]
            return cur[0] - start[0], cur[1] - start[1]
        for dx in (-1, 0, 1):
            for dy in (-1, 0, 1):
                nxt = (cur[0] + dx, cur[1] + dy)
                if (dx or dy) and nxt not in prev and free(*nxt) and \
                        not (cur == start and nxt in taken):
                    prev[nxt] = cur
                    queue.append(nxt)
    return None


def quest_step(b, memo):
    """任務物品那一名這一手要按的鍵；不是那一名就回 None 交給一般邏輯。"""
    roster = {a['i']: a for a in b['roster']}
    me = roster.get(b['cursor'])
    if me is None or me['side'] != 0:
        return None
    holders = b.get('quest_holders') or []
    item = None
    for g in b.get('ground') or []:
        name, pos = g.rsplit(' ', 1)
        if name == memo.get('quest_item_name', name):
            x, y = (int(v) for v in pos.split(','))
            item = (x, y)
            memo.setdefault('quest_item_name', name)
            break
    party = [a for a in b['roster'] if a['side'] == 0 and not a['down']]
    if holders:
        carrier = holders[0]
    elif item is not None and party:
        carrier = memo.get('carrier')
        if carrier is None or roster.get(carrier, {}).get('down', True):
            carrier = min(party, key=lambda a: max(abs(a['x'] - item[0]), abs(a['y'] - item[1])))['i']
            memo['carrier'] = carrier
    else:
        return None
    bar = b['bar']
    taken = {(a['x'], a['y']) for a in b['roster'] if a['i'] != me['i'] and not a['down']}
    if me['i'] != carrier:
        # 護送：身邊有敵人就交給一般邏輯去打，否則跟著持物者走，擋在他與追兵之間。
        foes = [a for a in b['roster'] if a['side'] == 1 and not a['down']]
        if any(max(abs(f['x'] - me['x']), abs(f['y'] - me['y'])) == 1 for f in foes):
            return None
        c = roster[carrier]
        if max(abs(c['x'] - me['x']), abs(c['y'] - me['y'])) <= 1 or CMD_MOVE not in bar:
            return pick(bar, CMD_REST) if CMD_REST in bar else None
        near = lambda x, y: max(abs(c['x'] - x), abs(c['y'] - y)) == 1
        first = first_step(b.get('walls'), taken, (me['x'], me['y']), near)
        if first is None:
            return None
        return pick(bar, CMD_MOVE) + to_cell(*first) + ['RET', '>', 'RET']
    on_edge = lambda x, y: x in (EDGE_LO, EDGE_HI) or y in (EDGE_LO, EDGE_HI)
    if me['i'] in holders:
        if on_edge(me['x'], me['y']):
            return pick(bar, CMD_REST) if CMD_REST in bar else None
        is_goal = on_edge
    else:
        if (me['x'], me['y']) == item and CMD_PICKUP in bar:
            return pick(bar, CMD_PICKUP)
        is_goal = lambda x, y: (x, y) == item
    if CMD_MOVE not in bar:
        return None
    first = first_step(b.get('walls'), taken, (me['x'], me['y']), is_goal)
    if first is None:
        return pick(bar, CMD_REST) if CMD_REST in bar else None
    dx, dy = first
    # 子選單從 U-TURN 起，`>` 一次選步行；一次只走一格（SPEC-008 §5）。
    return pick(bar, CMD_MOVE) + to_cell(dx, dy) + ['RET', '>', 'RET']


def main():
    seed = sys.argv[1]
    save = sys.argv[2]
    tokens = pathlib.Path(sys.argv[3]).read_text(encoding='utf-8').split()
    out = pathlib.Path(sys.argv[4])
    limit = int(sys.argv[5]) if len(sys.argv) > 5 else 60
    target = None
    if len(sys.argv) > 6 and sys.argv[6]:
        target = tuple(int(v) for v in sys.argv[6].split(','))
    town_target = None
    if len(sys.argv) > 7 and sys.argv[7]:
        town_target = tuple(int(v) for v in sys.argv[7].split(','))
    memo = {}
    stop_after_battle = os.environ.get('KOL_STOP_AFTER_BATTLE') == '1'
    seen_battle = False
    if len(sys.argv) > 8 and sys.argv[8]:
        memo['bounds'] = tuple(int(v) for v in sys.argv[8].split(','))

    for step in range(limit):
        out.write_text(' '.join(tokens) + '\n', encoding='utf-8')
        try:
            row = run(tokens, seed, save)
        except subprocess.CalledProcessError as error:
            if error.stderr:
                print(error.stderr, file=sys.stderr)
            # 「一句一個鍵」那幾段是一次補一批的。補過頭時多出來的 SPACE 會漏到
            # 戰鬥之外，而野外的 token 表沒有 SPACE，整條腳本就失敗——退回那一批
            # 再減半重試。減到 1 還是不行才真的停。
            n = memo.get('last_batch', 0)
            if n > 1:
                tokens = tokens[:-n]
                memo['batch'] = max(1, n // 2)
                memo['last_batch'] = 0
                print(f'那一批 {n} 個鍵補過頭了，改用 {memo["batch"]} 個重試')
                continue
            print('這一串跑不動，退回上一批 token')
            break
        memo['last_batch'] = 0
        b = row.get('battle')
        if b is None:
            if stop_after_battle and seen_battle:
                print('首場戰鬥已結束並返回遊戲路徑', row.get('world'), row.get('town'))
                break
            if row.get('world') is None:
                if not town_step(tokens, row, town_target, memo):
                    break
            elif not world_step(tokens, row, target):
                break
            continue
        seen_battle = True
        # roster 只在與上一步不同時才寫進軌跡，沒寫就沿用上一份。
        if 'roster' in b:
            memo['roster'] = parse_roster(b['roster'])
        b['roster'] = memo.get('roster', [])
        stage, prompt = b['stage'], b['prompt']
        print(f'[{step}] {len(tokens)} tokens stage={stage} prompt={prompt} '
              f'round={b["round"]} turn={b["turn"]} cursor={b["cursor"]} bar={b["bar"]}')
        if stage == 5:
            # 勝負那兩句、獎勵與戰利品清單都要按鍵推完，才真的離開戰場。
            verdict = {1: '敗北', 2: '勝利', 3: '逃走'}.get(b['verdict'], '未知')
            print(f'這一場的去向 {verdict}（verdict={b["verdict"]}）')
            tokens += press(memo)
            continue
        if stage == 1:
            tokens.append('SPACE')
            continue
        if stage == 4:
            # 戰利品清單：白框移到 U-TURN（第 1 格）再確認，不領東西就離開。只按 RET
            # 會一直領、背包滿了就停在這一頁出不去（full-chain-reload 重生時卡過）。
            tokens += ['>', 'RET']
            continue
        if stage == 3:
            # 逃走那一問（StageFlee）。任務模式有人拿著任務物品時答 y（SPEC-014 §5）。
            holders = b.get('quest_holders') or []
            tokens.append('Y' if QUEST_MODE and holders else 'N')
            continue
        if prompt == 10:
            # 執行段一句一個鍵。一次補一批再看——一場戰鬥的執行段有上百句，
            # 一次補一個的話光是這一段就要跑上百輪。
            tokens += press(memo)
            continue
        if prompt == 8:
            # Begin next round?
            tokens.append('RET')
            continue
        if prompt == 9:
            # Do you want to flee? 一般路徑要打完，不逃；任務模式在有人拿著任務物品
            # 站在邊上時答 y（帶著東西逃掉也算完成）。
            holders = b.get('quest_holders') or []
            tokens.append('Y' if QUEST_MODE and holders else 'N')
            continue
        if prompt == 11:
            # Do you want to pick it up? 任務模式一律撿。
            tokens.append('Y' if QUEST_MODE else 'N')
            continue
        # 攻擊式那一列的第一個選項（0x2A）換算出來的攻擊選項是 **0**，
        # 防禦式那一列的第一個（0x2A）換算出來的防禦選項也是 0——兩個都是 0
        # 的時候 `ResolveAttack` 會把整手改成休息。要真的打出去得挑第二個。
        if prompt in (2, 3):
            tokens += ['>', '>', 'RET']
            continue
        if prompt == 5:
            tokens += ['>', '>', '>', 'RET']
            continue
        if prompt == 4:
            tokens += ['>', 'RET']
            continue
        if prompt == 1:
            # 停在選目標格那一層：上一批 token 在挑目標的中途用完了（指令列在兩批
            # 之間換過內容就會這樣——有人倒下，格數就變了）。`to_cell` 開頭那四下
            # 是把游標推到左上角，所以這一段是**絕對**的，重下一次就收斂。
            roster = {a['i']: a for a in b['roster']}
            me = roster.get(b['cursor'])
            foes = [a for a in b['roster'] if a['side'] == 1 and not a['down']]
            if me is None or not foes:
                print('選目標格那一層收不回來', b['cursor'], b['bar'])
                break
            foes.sort(key=lambda f: max(abs(f['x'] - me['x']), abs(f['y'] - me['y'])))
            foe = foes[0]
            if b['cursor_cell'][2] == 2:
                keys = to_ranged_cell(foe['x'] - me['x'], foe['y'] - me['y'])
                if keys is None:
                    tokens += ['>', 'RET']  # U-TURN 回主指令列，重新挑選指令
                    memo.pop('decided', None)
                    continue
            else:
                keys = to_cell((foe['x'] > me['x']) - (foe['x'] < me['x']),
                               (foe['y'] > me['y']) - (foe['y'] < me['y']))
            tokens += keys + ['RET']
            continue
        if prompt != 0:
            print('停在沒預期的一層', prompt, b['bar'])
            break
        # 主指令列。這一名在這一回合已經下過指令了：寫回之後原版停在同一名、白框在
        # YES（RE-158），按 RET 確認換人。
        key = (b['round'], b['turn'])
        if memo.get('decided') == key:
            tokens.append('RET')
            continue
        memo['decided'] = key
        if QUEST_MODE:
            step = quest_step(b, memo)
            if step is not None:
                tokens += step
                continue
        # 相鄰就打，不然往最近的敵人走一格。
        roster = {a['i']: a for a in b['roster']}
        me = roster.get(b['cursor'])
        if me is None:
            print('輪到的參戰者不在名單裡', b['cursor'])
            break
        foes = [a for a in b['roster'] if a['side'] == 1 and not a['down']]
        if not foes:
            print('沒有站著的敵人了')
            break
        # 六個人全部朝同一隻走會擠成一列，後面幾個一直「blocked from moving」。
        # 依隊伍裡的次序把人分給兩隻，再挑一格沒有人站的相鄰格前進。
        mine = [a['i'] for a in b['roster'] if a['side'] == 0 and not a['down']]
        # 挑**離自己最近**的那兩隻之一。按索引分給全部敵人的話，站得開的那幾場
        # （任務地點有十二隻）會讓每個人各追一隻跑到天邊，收斂不了。
        foes.sort(key=lambda f: max(abs(f['x'] - me['x']), abs(f['y'] - me['y'])))
        # 射擊是一項正式指令；弓手不必為了生成腳本而一直追近戰目標。
        if CMD_MISSILE in b['bar']:
            for candidate in foes:
                keys = to_ranged_cell(candidate['x'] - me['x'], candidate['y'] - me['y'])
                if keys is not None:
                    tokens += pick(b['bar'], CMD_MISSILE) + keys + ['RET']
                    break
            else:
                keys = None
            if keys is not None:
                continue
        foe = foes[mine.index(me['i']) % min(len(foes), 2)]
        dist = lambda x, y: max(abs(foe['x'] - x), abs(foe['y'] - y))
        bar = b['bar']
        if dist(me['x'], me['y']) == 1:
            dx = (foe['x'] > me['x']) - (foe['x'] < me['x'])
            dy = (foe['y'] > me['y']) - (foe['y'] < me['y'])
            if CMD_WEAPON in bar:
                tokens += pick(bar, CMD_WEAPON) + to_cell(dx, dy) + ['RET']
                continue
            if CMD_UNARMED in bar:
                tokens += pick(bar, CMD_UNARMED) + to_cell(dx, dy) + ['RET']
                continue
        taken = {(a['x'], a['y']) for a in b['roster'] if a['i'] != me['i']}
        moves = [(dx, dy) for dx in (-1, 0, 1) for dy in (-1, 0, 1)
                 if (dx or dy) and (me['x'] + dx, me['y'] + dy) not in taken]
        if CMD_MOVE in bar and moves:
            dx, dy = min(moves, key=lambda d: dist(me['x'] + d[0], me['y'] + d[1]))
            tokens += pick(bar, CMD_MOVE) + to_cell(dx, dy) + ['RET', '>', 'RET']
            continue
        if CMD_REST in bar:
            tokens += pick(bar, CMD_REST)
            continue
        print('指令列上沒有可用的一格', bar)
        break
    out.write_text(' '.join(tokens) + '\n', encoding='utf-8')
    print('寫出', out, len(tokens), 'tokens')


main()
