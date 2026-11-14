'use strict';

// PostToolUse adapter: check only Go files named by this edit, never the dirty tree.
const fs = require('node:fs');
const path = require('node:path');
const { spawnSync } = require('node:child_process');
const root = path.resolve(__dirname, '..');

function editedFiles(event, repo = root) {
  if (event.hook_event_name !== 'PostToolUse') return [];
  const input = event.tool_input || {};
  const names = [];
  if (event.tool_name === 'apply_patch') {
    for (const match of String(input.command || input.patch || '').matchAll(/^\*\*\* (?:Add File|Update File|Delete File|Move to): (.+)\r?$/gm)) {
      names.push(match[1].trim());
    }
  } else if (['Edit', 'Write'].includes(event.tool_name) && input.file_path) {
    names.push(input.file_path);
  }
  const base = fs.realpathSync(repo);
  return [...new Set(names.map(name => path.resolve(event.cwd || repo, name)))].filter(file => {
    if (path.extname(file) !== '.go' || !fs.existsSync(file)) return false;
    const relative = path.relative(base, fs.realpathSync(file));
    return relative && !relative.startsWith('..' + path.sep) && relative !== '..' && !path.isAbsolute(relative);
  });
}

function run(event) {
  const files = editedFiles(event);
  if (!files.length) return 0;
  const packages = [...new Set(files.map(file => './' + path.relative(root, path.dirname(file)).replaceAll('\\', '/')))].map(p => p === './' ? '.' : p);
  let failed = false;
  function command(binary, args, options = {}) {
    process.stderr.write(`[depdeck] ${binary} ${args.join(' ')}\n`);
    const result = spawnSync(binary, args, { cwd: root, encoding: 'utf8', timeout: 180000, ...options });
    if (result.error?.code === 'ENOENT') {
      process.stderr.write(`[depdeck] Missing ${binary}; see docs/DEVELOPMENT.md. Skipping this step.\n`);
      return;
    }
    if (result.stdout) process.stderr.write(result.stdout);
    if (result.stderr) process.stderr.write(result.stderr);
    if (result.error || result.status !== 0) {
      failed = true;
      if (result.error) process.stderr.write(result.error.message + '\n');
    }
  }
  command('gofmt', ['-w', ...files]);
  command('goimports', ['-w', ...files]);
  command('golangci-lint', ['run', ...packages]);
  const race = process.platform !== 'win32' || spawnSync('gcc', ['-dumpmachine'], { encoding: 'utf8' }).stdout?.trim() === 'x86_64-w64-mingw32';
  if (!race) process.stderr.write('Skipping -race on Windows (no 64-bit gcc). CI runs it on Linux.\n');
  command('go', ['test', ...(race ? ['-race'] : []), ...packages]);
  return failed ? 2 : 0;
}

module.exports = { editedFiles, run };
if (require.main === module) {
  try {
    process.exitCode = run(JSON.parse(fs.readFileSync(0, 'utf8')));
  } catch (error) {
    process.stderr.write(`[depdeck] Go hook failed: ${error.message}\n`);
    process.exitCode = 2;
  }
}
