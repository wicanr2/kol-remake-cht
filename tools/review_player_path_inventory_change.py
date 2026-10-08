#!/usr/bin/env python3
"""審查 RE-170 引起的玩家路徑基準差分；不改比較器或正式收據。

固定輸入的原始空格證據已在 verify_inventory_empty_slots.py 核實。本工具只接受
持物摘要與原空物品移除／戰後同步；其他軌跡或保存差異皆拒絕。輸出只是審查證據，
嚴格比較仍須另由 check_player_paths.py 全量重跑。
"""
import argparse
import copy
import hashlib
import json
from pathlib import Path
import re


def sha(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def stripped(lines):
    return [re.sub(r'carried=\d+', 'carried=<已獨立審查>', line) for line in lines]


def review_trace(old, new, removed):
    assert len(old) == len(new), '不可接受路徑長度變動'
    cached = [None, None]
    changes = []
    for i, pair in enumerate(zip(old, new)):
        normalized = []
        for side, row in enumerate(pair):
            row = copy.deepcopy(row)
            if 'town' in row:
                if 'party' in row['town']:
                    cached[side] = row['town']['party']
                row['town'].pop('party', None)
            normalized.append(row)
        assert normalized[0] == normalized[1], (i, '持物摘要以外狀態變動')
        if cached[0] is not None or cached[1] is not None:
            assert cached[0] is not None and cached[1] is not None
            assert stripped(cached[0]) == stripped(cached[1]), (i, '隊伍其他數值變動')
            for a, b in zip(cached[0], cached[1], strict=True):
                name = a.split(' gold=')[0]
                before = int(re.search(r'carried=(\d+)', a)[1])
                after = int(re.search(r'carried=(\d+)', b)[1])
                assert 0 <= before - after <= removed.get(name, 0), (i, name, before, after)
        if pair[0] != pair[1]:
            changes.append({'step': i, 'old_party_event': pair[0].get('town', {}).get('party'),
                            'new_party_event': pair[1].get('town', {}).get('party')})
    return changes


def review_save(old, new):
    if old['schema'] == 'kol-save-v1' and new['schema'] == 'kol-save-v3':
        # 舊住宿夾具早於 SPEC-023；只能比較它當時實際保存的欄位。
        for key in ('origin', 'settings', 'world', 'location'):
            assert old[key] == new[key], key
        assert len(old['party']) == len(new['party'])
        checked = []
        for i, member in enumerate(old['party']):
            current = new['characters'][new['party'][i]]
            for key, value in member.items():
                if key in ('rank', 'adventure_points'):
                    got = current['sheet'][key]
                elif key == 'speaker':
                    got = new['speaker'] == i
                else:
                    got = current[key]
                assert value == got, (i, key, 'v1 歷史欄位不同')
                checked.append(f'party[{i}].{key}')
        return [{'scope': 'v1 歷史保存逐欄投影；不聲稱 v3 新增欄位曾在舊檔保存',
                 'checked_legacy_fields': checked, 'current_schema': 'kol-save-v3'}]
    adjusted = copy.deepcopy(old)
    changes = []
    if adjusted['schema'] == 'kol-save-v3':
        assert new['schema'] == adjusted['schema']
        for i, (a, b) in enumerate(zip(adjusted['characters'], new['characters'], strict=True)):
            if a is None:
                assert b is None
                continue
            fields = [key for key in a if a[key] != b[key]]
            if fields:
                changes.append({'slot': i, 'name': a['name'], 'fields': fields})
            a['backpack'] = [g for g in (a['backpack'] or [])
                             if bytes.fromhex(g['record'])[0] != 0] or None
            if a['in_hand'] and bytes.fromhex(a['in_hand']['record'])[0] == 0:
                a['in_hand'] = None
            if a['carried'] != b['carried']:
                worn = a.get('worn', a.get('equipment'))
                assert worn is not None and len(worn) == 10
                pack = a['backpack'] or []
                items = [a['in_hand']] + pack + [None] * (6-len(pack)) + worn
                a['carried'] = [g['record'] if g else '00'*16 for g in items]
            assert a == b, (i, '空物品與既有同步以外的角色保存差異')
    adjusted.pop('saved_at', None)
    new = {k: v for k, v in new.items() if k != 'saved_at'}
    assert adjusted == new, '其他保存欄位不同'
    return changes


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--before', required=True, help='完整舊 player-path 備份目錄')
    parser.add_argument('--replay', required=True, help='check_player_paths.py 的輸出目錄')
    parser.add_argument('--before-roster', required=True)
    parser.add_argument('--before-saves', help='原保存檔備份；更新正式保存入口後重生審查時必填')
    parser.add_argument('--out', required=True)
    args = parser.parse_args()
    before_root, replay_root = Path(args.before), Path(args.replay)
    roster = json.loads(Path(args.before_roster).read_text())
    assert sha(args.before_roster) == 'f8f05ff26b1c278ab457616eb951fa1517c2e4ab8e7e8401a5c3151209fa972e'
    removed = {}
    for slot in roster['slots']:
        if not slot['occupied']:
            continue
        items = (slot['backpack'] or []) + (slot['equipment'] or []) + [slot['in_hand']]
        removed[slot['name']] = sum(g is not None and bytes.fromhex(g['record'])[0] == 0 for g in items)
    assert sum(removed.values()) == 22
    reports = []
    for plan_path in sorted(before_root.glob('*.json')):
        if plan_path.name.endswith('-receipt.json'):
            continue
        plan = json.loads(plan_path.read_text())
        name = plan_path.stem
        old_trace = before_root / name / 'trace.json'
        new_trace = replay_root / name / 'trace.json'
        old = json.loads(old_trace.read_text())['steps']
        new = json.loads(new_trace.read_text())['steps']
        changes = review_trace(old, new, removed)
        row = {'name': name, 'states': len(new), 'changed_steps': changes,
               'old_trace_sha256': sha(old_trace), 'new_trace_sha256': sha(new_trace),
               'other_trace_fields_identical': True, 'party_other_values_identical': True}
        if plan.get('save'):
            old_save = (Path(args.before_saves) / Path(plan['save']['file']).name
                        if args.before_saves else Path(plan['save']['file']))
            new_save = replay_root / name / 'save.json'
            row['save'] = {'changed_character_fields': review_save(
                json.loads(old_save.read_text()), json.loads(new_save.read_text())),
                'old_sha256': sha(old_save), 'new_sha256': sha(new_save),
                'other_fields_identical': True}
        reports.append(row)
    assert len(reports) == 29
    result = {'schema': 'kol-player-path-inventory-change-review-v1', 'paths': reports,
              'path_count': len(reports), 'changed_path_count': sum(bool(r['changed_steps']) for r in reports),
              'removed_empty_items': removed,
              'scope': '固定 RE-170 初始資料變更的逐欄審查；不修改嚴格比較、不宣稱原版 parity',
              'tool_sha256': sha(__file__)}
    Path(args.out).write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps({'paths': len(reports), 'changed_paths': result['changed_path_count']}, ensure_ascii=False))


if __name__ == '__main__':
    main()
