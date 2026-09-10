"""將 dosgolem 正常入口角色記錄交給既有抽取器；只產生本機測試輸入。"""
import hashlib
import json
from pathlib import Path
from extract_character_roster import character_sheet, equipment, worn_slot, cstring

root = Path("/src/workplace/battle-state-records-20260908")
records = json.loads((root / "battle-state-records.json").read_text())
alignment = json.loads((root / "battle-alignment.json").read_text())
assert records["source"] == alignment["source"] == "dosgolem-normal-player-path"
assert records["alignment_seed"] == alignment["alignment_seed"] == 1
party = []
for row in records["characters"]:
    raw = bytes.fromhex(row["record_hex"])
    assert len(raw) == 0x300
    slot = row["slot"]
    name = cstring(raw[:16])
    assert name == alignment["actors"][slot]["name"]
    party.append(dict(slot=slot, name=name, wounds=raw[0x1b],
                      sheet=character_sheet(raw), equipment=equipment(raw),
                      in_hand=worn_slot(raw, 0, 0x30)))
assert len(records["monsters"]) == 1 and records["monsters"][0]["slot"] == 6
monster_raw = bytes.fromhex(records["monsters"][0]["record_hex"])
assert len(monster_raw) == 0x80 and 0 < monster_raw[0] <= 12
monster_name = monster_raw[1:1 + monster_raw[0]].decode("ascii")
monster_path = Path("/src/assets/original-derived-data/monsters.json")
monster = next(m for m in json.loads(monster_path.read_text())["monsters"] if m["name"] == monster_name)
for offset, field in enumerate(("strength", "quickness", "foresight", "size", "health", "intellect"), 0x11):
    assert monster[field] == monster_raw[offset], field
assert monster["weapon"]["record"] == monster_raw[0x20:0x30].hex()
assert monster["ranged_chance"] == monster_raw[0x1e]
assert monster["ranged_code"] == monster_raw[0x1f]
fixture = dict(party=party, actors=alignment["actors"], counters=records["counters"],
               monster=monster, monster_level=records["monster_level"],
               seed=records["alignment_seed"], source="dosgolem-normal-player-path",
               sources_sha256={p.name: hashlib.sha256(p.read_bytes()).hexdigest()
                               for p in (root / "battle-state-records.json", root / "battle-alignment.json", monster_path)})
target = Path("/out/party-fixture.json")
assert not target.exists()
target.write_text(json.dumps(fixture, ensure_ascii=False, indent=2)+"\n")
