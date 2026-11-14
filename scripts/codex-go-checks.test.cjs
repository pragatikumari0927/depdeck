'use strict';
const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { editedFiles } = require('./codex-go-checks.cjs');

test('selects only edited existing Go files inside the repository', t => {
  const temp = fs.mkdtempSync(path.join(os.tmpdir(), 'depdeck-hook-'));
  t.after(() => fs.rmSync(temp, { recursive: true, force: true }));
  const repo = path.join(temp, 'repo');
  fs.mkdirSync(repo);
  for (const name of ['changed.go', 'untouched.go', 'notes.md']) fs.writeFileSync(path.join(repo, name), '');
  fs.writeFileSync(path.join(temp, 'outside.go'), '');
  const event = { hook_event_name: 'PostToolUse', tool_name: 'apply_patch', cwd: repo, tool_input: { command: '*** Update File: changed.go\n*** Update File: changed.go\n*** Delete File: deleted.go\n*** Update File: notes.md\n*** Update File: ../outside.go\n' } };
  assert.deepEqual(editedFiles(event, repo), [path.join(repo, 'changed.go')]);
  assert.deepEqual(editedFiles({ ...event, hook_event_name: 'PreToolUse' }, repo), []);
  assert.deepEqual(editedFiles({ ...event, tool_name: 'Bash' }, repo), []);
  assert.deepEqual(editedFiles({ ...event, tool_name: 'Edit', tool_input: { file_path: path.join(repo, 'changed.go') } }, repo), [path.join(repo, 'changed.go')]);
});
