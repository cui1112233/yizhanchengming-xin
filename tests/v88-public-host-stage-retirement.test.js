const assert = require('node:assert/strict');
const fs = require('node:fs');
const test = require('node:test');

function read(file) {
  return fs.readFileSync(file, 'utf8');
}

test('retired host-stage entry points cannot route public V88 traffic through a host port or container IP', () => {
  const stage = read('deploy/v88-direct/stage-node-host.sh');
  const cutover = read('deploy/v88-direct/cutover-node-host.sh');
  const stageWorkflowPath = '.github/workflows/v88-direct-deploy-node-stage.yml';
  const cutoverWorkflowPath = '.github/workflows/v88-direct-deploy-node-cutover.yml';
  // These historical Node workflows are no longer part of the checkout. The
  // direct Go release path is validated by deploy/v88-direct contract tests.
  if (!fs.existsSync(stageWorkflowPath) || !fs.existsSync(cutoverWorkflowPath)) {
    assert.match(stage, /retired/i);
    assert.match(cutover, /retired/i);
    return;
  }
  const stageWorkflow = read(stageWorkflowPath);
  const cutoverWorkflow = read(cutoverWorkflowPath);

  for (const source of [stage, cutover, stageWorkflow, cutoverWorkflow]) {
    assert.match(source, /retired/i);
    assert.doesNotMatch(source, /18081|NetworkSettings\.Networks|QIANTIE_GO_BASE_URL=http:\/\/\$\w+_ip|nginx\s+-s\s+reload/);
  }
  assert.doesNotMatch(stageWorkflow, /stage-node-host\.sh/);
  assert.doesNotMatch(cutoverWorkflow, /cutover-node-host\.sh/);
});
