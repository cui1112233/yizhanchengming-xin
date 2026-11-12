import test from 'node:test';
import assert from 'node:assert/strict';
import { existsSync, mkdirSync, mkdtempSync, readFileSync, writeFileSync } from 'node:fs';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { spawnSync } from 'node:child_process';
import path from 'node:path';

test('collapsed timeline keeps a clickable reopen control inside the viewport with project tabs present', () => {
  const chrome = ['C:/Program Files/Google/Chrome/Application/chrome.exe', 'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(existsSync);
  assert.ok(chrome, 'browser required for layout regression');
  const archive = fileURLToPath(new URL('../../../../../_local-archive-20261028/', import.meta.url));
  mkdirSync(archive, { recursive: true });
  const dir = mkdtempSync(path.join(archive, 'aicut-timeline-visibility-'));
  const css = ['../aicut-page.css', './editor.css', './editor-redesign.css'].map(name => readFileSync(new URL(name, import.meta.url), 'utf8')).join('\n');
  const html = `<!doctype html><style>html,body{margin:0;height:100%;box-sizing:border-box}*{box-sizing:border-box}${css}</style>
  <div class="aicut-fullscreen-editor"><section class="aicut-editor-shell aicut-editor-v4 aicut-jianying-shell aicut-reference-editor timeline-collapsed" style="--aicut-resource-width:300px;--aicut-inspector-width:280px;--aicut-timeline-height:40px">
  <section class="aicut-reference-timeline aicut-timeline collapsed"><div class="aicut-timeline-tabs"><button>默认时间线</button></div><div class="aicut-timeline-toolbar aicut-timeline-toolbar-v4"><div class="aicut-timeline-toolbar-primary"><button>切分</button></div><div class="aicut-timeline-toolbar-secondary"><button>录音</button><button class="aicut-timeline-collapse" aria-label="展开时间线"><span>展开时间线</span></button></div></div><div class="aicut-timeline-ruler-row">00:00</div><div class="aicut-timeline-body">轨道内容</div></section></section></div><pre id="result">pending</pre><script>
  try {
    const shell=document.querySelector('.aicut-editor-shell'), timeline=document.querySelector('.aicut-timeline'), button=document.querySelector('.aicut-timeline-collapse');
    const rect=button.getBoundingClientRect(), bounds=timeline.getBoundingClientRect();
    const point=document.elementFromPoint(rect.x+rect.width/2,rect.y+rect.height/2);
    if(rect.y<bounds.y || rect.bottom>innerHeight || rect.bottom>bounds.bottom || !button.contains(point)) throw Error('reopen control is outside visible timeline: '+JSON.stringify({button:{y:rect.y,bottom:rect.bottom},bounds:{y:bounds.y,bottom:bounds.bottom},height:innerHeight}));
    button.addEventListener('click',()=>{shell.classList.remove('timeline-collapsed');timeline.classList.remove('collapsed');shell.style.setProperty('--aicut-timeline-height','280px');});
    button.click();
    const body=document.querySelector('.aicut-timeline-body').getBoundingClientRect();
    if(body.height<100 || body.bottom>innerHeight+1) throw Error('expanded tracks not visible: '+JSON.stringify({h:body.height,bottom:body.bottom,height:innerHeight}));
    document.querySelector('#result').textContent='PASS';
  }catch(error){document.querySelector('#result').textContent='FAIL '+error.message;}
  </script>`;
  const fixture=path.join(dir,'fixture.html'); writeFileSync(fixture,html);
  for (const size of ['1280,720','900,600']) {
    const result=spawnSync(chrome,['--headless=new','--no-sandbox','--disable-gpu','--no-first-run',`--user-data-dir=${path.join(dir,size.replace(',','-'))}`,`--window-size=${size}`,'--dump-dom',pathToFileURL(fixture).href],{encoding:'utf8',timeout:60000,maxBuffer:8*1024*1024});
    assert.equal(result.status,0,result.error?.message || result.stderr);
    assert.ok(/id="result">PASS<\/pre>/.test(result.stdout), result.stdout.match(/id="result">[^<]*/)?.[0] || 'missing result');
  }
});
