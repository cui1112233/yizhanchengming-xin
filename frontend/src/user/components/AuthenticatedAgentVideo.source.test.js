import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(dirname, 'AuthenticatedAgentVideo.jsx'), 'utf8');

test('video preview retries a protected source after a temporary read failure', () => {
  assert.match(source, /const \[retryAttempt, setRetryAttempt\] = useState\(0\)/);
  assert.match(source, /nextAuthenticatedMediaRetryAttempt\(source, retryAttempt\)/);
  assert.match(source, /setRetryAttempt\(nextAttempt\)/);
  assert.match(source, /onClick=\{\(\) => setRetryAttempt\(value => value \+ 1\)\}/);
});
