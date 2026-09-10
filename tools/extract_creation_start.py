#!/usr/bin/env python3
"""依 SPEC-005 writer 契約產生建角起始 typed data；只在容器離線執行。"""
import hashlib
import json
from pathlib import Path
import zipfile
from extract_town_buildings import good
from extract_character_roster import HEIGHT_BASE, WEIGHT_BASE, WEIGHT_PER_FOOT

root = Path(__file__).resolve().parents[1]
archive = root / "Knights-of-legend.zip"
assert hashlib.sha256(archive.read_bytes()).hexdigest() == "2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46"
with zipfile.ZipFile(archive) as z:
    region = z.read("Knights-of-legend/REGION0.DAT")
assert len(region) == 48
assert hashlib.sha256(region).hexdigest() == "21d1291a41792688268af02bbeabcf56a5402c614d3c0d86da2adbe55a9e6ca4"
contract_path = root / "docs/re/ida-fingerprints/KOLBOOK-character-writer-summary.json"
weapon_path = root / "assets/original-derived-data/weapon-templates.json"
contracts = json.loads(contract_path.read_text())["profession_starting_contracts"]
templates = json.loads(weapon_path.read_text())
profiles = []
for contract, selection in zip(contracts, templates["professions"], strict=True):
    assert contract["record_index"] == selection["record_index"]
    skills = []
    hand = None
    for row, key in enumerate(("primary", "secondary")):
        index = selection[key + "_record"]
        if index is None:
            continue
        rec = bytes.fromhex(templates["records"][index])
        assert len(rec) == 16
        item = good(rec, index)
        if row == 0:
            hand = item
        attack, defence = contract[key + "_skill_pair_internal"]
        skills.append(dict(row=row, name=item["name"], weapon=rec[13],
                           offence=attack // 10, defence=defence // 10,
                           offence_raw=attack, defence_raw=defence))
    profiles.append(dict(record_index=contract["record_index"], gold=contract["gold_crowns"],
                         weapons=skills, in_hand=hand,
                         rank=sum(sum(contract[k + "_skill_pair_internal"]) for k in ("primary", "secondary")) // 200))
out = dict(schema="kol-creation-start-v1", place=region[3:13].decode("ascii"),
           town=region[13:16].decode("ascii"), building_id=int.from_bytes(region[1:3], "big"),
           height_base=HEIGHT_BASE, weight_base=WEIGHT_BASE, weight_per_foot=WEIGHT_PER_FOOT,
           profiles=profiles, evidence="SPEC-005、RE-050、RE-074；IDA 9.4 KOLBOOK writer REGION copy map",
           sources={str(p.relative_to(root)): hashlib.sha256(p.read_bytes()).hexdigest()
                    for p in (archive, contract_path, weapon_path)},
           region_sha256=hashlib.sha256(region).hexdigest(), rights="僅本機；未取得公開散布權")
target = root / "assets/original-derived-data/creation-start.json"
assert target.parent.stat().st_uid == 1000
if target.exists():
    assert target.stat().st_uid == 1000
target.write_text(json.dumps(out, ensure_ascii=False, indent=2) + "\n")
print(target, hashlib.sha256(target.read_bytes()).hexdigest())
