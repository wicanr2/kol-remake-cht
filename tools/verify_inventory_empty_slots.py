#!/usr/bin/env python3
"""核對固定 CHARDATA 的物品空格與本輪抽取差分；在 Docker 執行。

用法：python3 tools/verify_inventory_empty_slots.py --before BEFORE.json
    --after assets/original-derived-data/character-roster.json --zip Knights-of-legend.zip
"""
import argparse
import hashlib
import json
from pathlib import Path
import zipfile

from extract_character_roster import (
    CHARDATA_NAME, EXPECTED_CHARDATA_SHA256, EXPECTED_ZIP_SHA256, worn_slot,
)


def verify(before, after, raw):
    checked = 0
    corrections = []
    for old, new in zip(before['slots'], after['slots'], strict=True):
        if not old['occupied']:
            assert old == new
            continue
        record = raw[old['slot'] * 768:(old['slot'] + 1) * 768]
        for field, offsets in (
            ('equipment', list(range(0x100, 0x1A0, 16))),
            ('backpack', list(range(0xA0, 0x100, 16))),
            ('in_hand', [0x30]),
        ):
            old_items = [old[field]] if field == 'in_hand' else old[field]
            new_items = [new[field]] if field == 'in_hand' else new[field]
            assert len(old_items) == len(new_items) == len(offsets)
            for i, (offset, a, b) in enumerate(zip(offsets, old_items, new_items, strict=True)):
                checked += 1
                block = record[offset:offset + 16]
                assert (b is None) == (block[0] == 0), (old['slot'], field, i)
                if block[0]:
                    assert a == b and bytes.fromhex(b['record']) == block
                elif a is not None:
                    assert any(block[1:]) and bytes.fromhex(a['record']) == block
                    corrections.append({'slot': old['slot'], 'name': old['name'],
                                        'field': field, 'index': i, 'offset': offset})
        assert {k: v for k, v in old.items() if k not in ('equipment', 'backpack', 'in_hand')} == {
            k: v for k, v in new.items() if k not in ('equipment', 'backpack', 'in_hand')}
    # 唯一允許的 metadata 差分是新登記的空格判準。
    old_meta = {k: v for k, v in before.items() if k != 'slots'}
    new_meta = {k: v for k, v in after.items() if k != 'slots'}
    new_meta['layout'] = dict(new_meta['layout'])
    assert new_meta['layout'].pop('item_empty_rule')
    assert old_meta == new_meta
    assert checked == 238 and len(corrections) == 22
    assert sum(c['field'] == 'backpack' for c in corrections) == 21
    assert all(g is None for g in after['slots'][0]['backpack'])
    assert after['slots'][12]['in_hand'] is None
    # 原版邊界反例：全部尾端非零仍是空格；有效首 byte 不可被略過。
    assert worn_slot(bytes([0]) + bytes([255]) * 15, 0, 0) is None
    assert worn_slot(bytes(16), 0, 0) is None
    assert worn_slot(bytes([0x10]) + bytes(15), 0, 0) is not None
    return {'schema': 'kol-inventory-empty-slot-verification-v1', 'checked_slots': checked,
            'corrected_slots': corrections, 'other_character_fields_unchanged': True,
            'carried_raw_bytes_unchanged': True, 'boundary_checks': 3,
            'original_basis': 'IDA 靜態指令與固定原版資料；非原版動態對拍'}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--before', required=True)
    parser.add_argument('--after', required=True)
    parser.add_argument('--zip', required=True)
    args = parser.parse_args()
    zipped = Path(args.zip).read_bytes()
    assert hashlib.sha256(zipped).hexdigest() == EXPECTED_ZIP_SHA256
    with zipfile.ZipFile(args.zip) as source:
        raw = source.read(CHARDATA_NAME)
    assert hashlib.sha256(raw).hexdigest() == EXPECTED_CHARDATA_SHA256
    result = verify(json.loads(Path(args.before).read_text()), json.loads(Path(args.after).read_text()), raw)
    result['inputs_sha256'] = {
        key: hashlib.sha256(Path(value).read_bytes()).hexdigest()
        for key, value in (('before', args.before), ('after', args.after), ('tool', __file__))}
    print(json.dumps(result, ensure_ascii=False, indent=2))


if __name__ == '__main__':
    main()
