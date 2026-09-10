"""交付清單的版號與完整性回歸測試；僅在 Docker 執行。"""
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


class ManifestTest(unittest.TestCase):
    def test_version_and_nested_manifests(self):
        tool = Path(__file__).with_name("write_dist_manifest.py")
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            nested = root / "full/assets/manifest.json"
            nested.parent.mkdir(parents=True)
            nested.write_text('{"asset": true}\n')
            command = [sys.executable, str(tool), "--dist", str(root),
                       "--platform", "linux/amd64", "--version"]
            bad = subprocess.run(command + ["0.2.2"], capture_output=True)
            self.assertNotEqual(bad.returncode, 0)
            self.assertFalse((root / "full/manifest.json").exists())
            for _ in range(2):
                subprocess.run(command + ["v.0.2.2-20260908"],
                               check=True, capture_output=True)
                manifest = json.loads((root / "full/manifest.json").read_text())
                self.assertEqual(manifest["platform"], ["linux/amd64"])
                self.assertEqual(manifest["file_count"], 1)
                self.assertEqual(manifest["files"][0]["path"], "assets/manifest.json")


if __name__ == "__main__":
    unittest.main()
