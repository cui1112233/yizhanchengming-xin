"""Read-only extraction of the shipped Nuitka constant-table catalog; no imports."""
import hashlib
import json
import re
from pathlib import Path

source = Path(r"F:/软件/洛水2026/洛水2026/modules/novels.pyd")
data = source.read_bytes()
start = data.index(b"uOOC\xe6\xa3\x80\xe6\x9f\xa5\0")
middle = data.index("u角色言行是否偏离".encode(), start)
end = data.index(b"\0S\x1d", middle)
pattern = re.compile(rb"u([^\0]+)\0u([^\0]+)\0")
names = list(pattern.finditer(data, start, middle))
descriptions = list(pattern.finditer(data, middle, end + 1))
assert len(names) == len(descriptions) == 37
# Nuitka constant-table set: S (set), 0x1d (29), l <one-byte int>.
raw_defaults = data[end + 3:end + 3 + 58]
assert all(raw_defaults[i] == ord('l') for i in range(0, 58, 2))
defaults = list(raw_defaults[1::2])
assert defaults == list(range(1, 28)) + [32, 33]
catalog = []
for number, (name, description) in enumerate(zip(names, descriptions), 1):
    catalog.append(dict(id=number, title=name[1].decode(), content=description[1].decode(), defaultEnabled=number in defaults))
repo = Path(__file__).resolve().parents[1]
out = repo / "docs/reverse-engineering/novel-review-dimensions-20261119"
out.mkdir(parents=True, exist_ok=True)
(out / "catalog.json").write_text(json.dumps(catalog, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
provenance = dict(source=str(source), sha256=hashlib.sha256(data).hexdigest(), titleTableOffset=start,
    descriptionTableOffset=middle, defaultSetOffset=end + 1, defaultSetHex=data[end+1:end+61].hex(),
    defaultIDs=defaults, count=37, entries=[dict(id=i+1,titleOffset=n.start()+1,descriptionOffset=d.start()+1) for i,(n,d) in enumerate(zip(names,descriptions))],
    limitation="Static constant-table recovery, not proof of native prompt assembly or paid output quality.")
(out / "provenance.json").write_text(json.dumps(provenance, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
print(json.dumps(dict(count=len(catalog), defaults=len(defaults), sourceSHA=provenance['sha256'])))
