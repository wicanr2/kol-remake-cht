"""沿用前輪附件與擁有權稽核，只更新本輪收據位置與schema。"""
from pathlib import Path

text = Path('/src/docs/verification/world-clock-20261002/audit_time.py').read_text()
assert text.count('world-clock-20261002') == 2
text = text.replace('world-clock-20261002', 'return-cost-20261002')
text = text.replace('kol-world-time-audit-v1', 'kol-return-cost-audit-v1')
text = text.replace("HERE / 'receipt.json'", "HERE / 'receipt2.json'")
exec(compile(text, 'audit_time.py', 'exec'), {'__name__': '__main__'})
