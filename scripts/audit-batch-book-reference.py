"""Read the supplied source without running it; emit traceable review inventories."""
import ast
import hashlib
import json
import pathlib
import re
from html.parser import HTMLParser

ROOT = pathlib.Path(__file__).resolve().parents[1]
SOURCE = ROOT / '_local-archive-20261028/rea-batch-book-20261112/批量原文改文系统_源码'
OUT = ROOT / 'docs/requirements/batch-book-reference'
OUT.mkdir(parents=True, exist_ok=True)

def save(name, value):
    (OUT / name).write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')

files = []
for path in sorted(SOURCE.rglob('*')):
    if path.is_file():
        data = path.read_bytes()
        files.append({'file': path.relative_to(SOURCE).as_posix(), 'bytes': len(data), 'sha256': hashlib.sha256(data).hexdigest()})
save('source-manifest.json', {'source_directory': str(SOURCE), 'method': '只读；未执行原程序；源码包无需反编译', 'files': files})

text = (SOURCE / 'main.py').read_text(encoding='utf-8-sig')
tree = ast.parse(text)
functions = [node for node in tree.body if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef))]
known = {node.name: node.lineno for node in functions}
records = []
for node in functions:
    calls = sorted({ast.unparse(call.func) for call in ast.walk(node) if isinstance(call, ast.Call)})
    states = []
    for assign in ast.walk(node):
        if isinstance(assign, ast.Assign) and isinstance(assign.value, ast.Constant) and isinstance(assign.value.value, str):
            for target in assign.targets:
                if isinstance(target, ast.Subscript) and isinstance(target.slice, ast.Constant) and target.slice.value in ('status', 'queue_status', 'original_status', 'ai_status', 'site_submit_status'):
                    states.append({'line': assign.lineno, 'field': ast.unparse(target), 'value': assign.value.value})
    records.append({'function': node.name, 'file': 'main.py', 'line': node.lineno, 'end_line': node.end_lineno,
                    'calls': calls, 'local_callees': [{'function': call, 'line': known[call]} for call in calls if call in known],
                    'state_assignments': states, 'verification': '静态提取；动态分派与分支结果待核对'})
save('original-function-chain.json', records)

class Controls(HTMLParser):
    def __init__(self):
        super().__init__()
        self.items = []
        self.current = []
    def handle_starttag(self, tag, attrs):
        if tag in ('button', 'input', 'select', 'textarea'):
            item = {'tag': tag, 'line': self.getpos()[0], 'attributes': dict(attrs), 'text': '', 'acceptance': '未运行验收'}
            self.items.append(item)
            if tag != 'input':
                self.current.append((tag, item))
    def handle_data(self, data):
        for _, item in self.current:
            item['text'] += data
    def handle_endtag(self, tag):
        if self.current and self.current[-1][0] == tag:
            self.current.pop()

html = Controls()
html.feed((SOURCE / 'static/index.html').read_text(encoding='utf-8-sig'))
for item in html.items:
    item['text'] = ' '.join(item['text'].split())
save('original-ui-controls.json', html.items)
js = (SOURCE / 'static/app.js').read_text(encoding='utf-8-sig').splitlines()
sites = []
for number, line in enumerate(js, 1):
    if re.search(r'\bapi\s*\(|addEventListener\s*\(|\bonclick\b|fetch\s*\(', line):
        sites.append({'file': 'static/app.js', 'line': number, 'evidence': line.strip()[:1600], 'verification': '定位线索；需展开回调、请求参数与返回处理'})
save('original-ui-request-sites.json', sites)
inventory = json.loads((ROOT / 'docs/requirements/2026-11-14-batch-book-full-parity-inventory.json').read_text(encoding='utf-8-sig'))
by_name = {node.name: node for node in functions}
by_record = {item['function']: item for item in records}
chains = []
for route in inventory['routes']:
    pending = [route['function']]
    reached = set()
    while pending:
        name = pending.pop()
        if name in reached or name not in by_record:
            continue
        reached.add(name)
        pending.extend(x['function'] for x in by_record[name]['local_callees'])
    boundaries = []
    for name in sorted(reached):
        for call in ast.walk(by_name[name]):
            if not isinstance(call, ast.Call):
                continue
            target = ast.unparse(call.func)
            if re.search(r'(^requests\.|\.request$|\.read_text$|\.write_text$|\.unlink$|^write_json$|^read_json$|^append_log$|^append_realtime_event$|^write_meta$|^read_meta$|^write_schedule$|^write_web_submit_result$)', target):
                boundaries.append({'function': name, 'line': call.lineno, 'call': target})
    chains.append({'method': route['method'], 'route': route['route'], 'entry_line': route['source_line'],
                   'reachable_functions': [{'function': name, 'line': known[name]} for name in sorted(reached)],
                   'data_and_network_boundaries': sorted(boundaries, key=lambda x: x['line']),
                   'limit': '仅静态直接调用可达性，不代表每次执行所有分支；动态分派仍需核实'})
save('original-api-call-chains.json', chains)
go_routes = []
for path in sorted((ROOT / 'backend/internal/httpapi').glob('*.go')):
    if path.name.endswith('_test.go'):
        continue
    for number, line in enumerate(path.read_text(encoding='utf-8-sig').splitlines(), 1):
        match = re.search(r'(?:HandleFunc|Handle)\("(GET|POST|PUT|PATCH|DELETE) ([^" ]+)"', line)
        if match and match[2].startswith(('/api/novel-fetch-workshop/', '/api/batch-rewrite/', '/api/process/', '/api/web-submit/')):
            go_routes.append({'method': match[1], 'route': match[2], 'file': path.relative_to(ROOT).as_posix(), 'line': number})
def normalized(route):
    return re.sub(r'\{[^}]+\}', '{}', route)
mapping = []
for route in inventory['routes']:
    expected = route['route'].replace('/api/', '/api/novel-fetch-workshop/', 1)
    candidates = [entry for entry in go_routes if entry['method'] == route['method'] and normalized(entry['route']) == normalized(expected)]
    alternatives = [entry for entry in go_routes if entry['method'] == route['method'] and
                    normalized(entry['route']) in {normalized(route['route']), normalized(route['route'].replace('/api/', '/api/batch-rewrite/', 1))}]
    mapping.append({'method': route['method'], 'original_route': route['route'], 'original_line': route['source_line'],
                    'exact_route_candidates': candidates, 'compatibility_route_candidates': alternatives,
                    'conclusion': '仅入口同名或兼容入口存在，行为待逐项验收' if candidates or alternatives else '无同名入口，需核实替代入口或缺失；不直接判定功能缺失'})
save('v88-route-mapping.json', mapping)
for item in html.items:
    control_id = item['attributes'].get('id')
    item['js_references'] = [] if not control_id else [{'line': number, 'evidence': line.strip()[:1000]}
        for number, line in enumerate(js, 1) if control_id in line]
save('original-ui-controls.json', html.items)
lines = ['# 原系统接口规格核对表', '', '这份表保留原包证据。候选接口名称相同不代表行为一致。每条仍须补齐正常、失败、恢复和界面验收。', '', '| 方法 / 接口 | 原包入口 | 输入证据 | 输出字段证据 | 明确失败 | 验收 |', '|---|---|---|---|---|---|']
for route in inventory['routes']:
    def cell(value):
        return json.dumps(value, ensure_ascii=False).replace('|', '\\|').replace('\n', ' ')[:1100]
    lines.append(f"| {route['method']} `{route['route']}` | main.py:{route['source_line']} `{route['function']}` | {cell(route['inputs'])} | {cell(route['direct_output_keys'])} | {cell(route['explicit_failures'])} | 未逐项验收 |")
(OUT / 'api-contract-review.md').write_text('\n'.join(lines) + '\n', encoding='utf-8')
print(json.dumps({'files': len(files), 'functions': len(records), 'controls': len(html.items), 'request_and_event_sites': len(sites), 'routes': len(inventory['routes']), 'output': str(OUT)}, ensure_ascii=False))
