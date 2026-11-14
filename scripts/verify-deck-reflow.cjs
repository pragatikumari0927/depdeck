const assert = require('node:assert/strict');
const { spawn } = require('node:child_process');
const { mkdtemp, readFile, rm } = require('node:fs/promises');
const { tmpdir } = require('node:os');
const path = require('node:path');
const { setTimeout: delay } = require('node:timers/promises');
const { pathToFileURL } = require('node:url');

const [chrome, deck] = process.argv.slice(2);
assert.ok(chrome && deck, 'Usage: node scripts/verify-deck-reflow.cjs <Chromium executable> <fixture Deck.html>');

async function checkReflow() {
  const profile = await mkdtemp(path.join(tmpdir(), 'depdeck-reflow-'));
  const browser = spawn(chrome, [
    '--headless', '--remote-debugging-port=0', '--disable-background-networking',
    '--no-first-run', '--disable-extensions', '--user-data-dir=' + profile, 'about:blank',
  ], { windowsHide: true, stdio: 'ignore' });
  const stopped = new Promise(resolve => {
    browser.once('error', resolve);
    browser.once('exit', resolve);
  });
  let socket;
  try {
    let page;
    for (let attempt = 0; attempt < 100 && !page; attempt++) {
      try {
        const portFile = await readFile(path.join(profile, 'DevToolsActivePort'), 'utf8');
        const port = portFile.split('\n')[0];
        const response = await fetch('http://127.0.0.1:' + port + '/json/list', { signal: AbortSignal.timeout(1000) });
        page = (await response.json()).find(item => item.type === 'page');
      } catch {
        await delay(100);
      }
    }
    assert.ok(page, 'Chromium DevTools did not start');
    socket = new WebSocket(page.webSocketDebuggerUrl);
    await new Promise((resolve, reject) => {
      socket.addEventListener('open', resolve, { once: true });
      socket.addEventListener('error', reject, { once: true });
    });
    const pending = new Map();
    const requests = [];
    let id = 0;
    socket.addEventListener('message', ({ data }) => {
      const message = JSON.parse(data);
      if (message.method === 'Network.requestWillBeSent') requests.push(message.params.request.url);
      pending.get(message.id)?.(message);
    });
    async function send(method, params = {}) {
      const request = { id: ++id, method, params };
      let timer;
      try {
        const message = await new Promise((resolve, reject) => {
          timer = setTimeout(() => reject(new Error(method + ' timed out')), 5000);
          pending.set(request.id, resolve);
          socket.send(JSON.stringify(request));
        });
        assert.ok(!message.error, JSON.stringify(message.error));
        return message.result;
      } finally {
        clearTimeout(timer);
        pending.delete(request.id);
      }
    }
    async function evaluate(expression) {
      const result = await send('Runtime.evaluate', { expression, returnByValue: true });
      assert.ok(!result.exceptionDetails, JSON.stringify(result.exceptionDetails));
      return result.result.value;
    }

    await send('Network.enable');
    await send('Network.setBlockedURLs', { urls: ['http://*', 'https://*'] });
    const navigation = await send('Page.navigate', { url: pathToFileURL(path.resolve(deck)).href });
    assert.ok(!navigation.errorText, navigation.errorText);
    for (let attempt = 0; attempt < 100; attempt++) {
      if (await evaluate('document.readyState === "complete" && Boolean(document.querySelector(".masthead"))')) break;
      await delay(50);
    }
    assert.ok(await evaluate('document.querySelectorAll(".card").length > 0 && Boolean(document.querySelector(".notice"))'),
      'Use a rendered Deck with assessed and degraded Cards');

    for (const width of [320, 390, 768, 1024, 1440]) {
      await send('Emulation.setDeviceMetricsOverride', { width, height: 1050, deviceScaleFactor: 1, mobile: false });
      for (const scale of [100, 200]) {
        const result = await evaluate('(() => {' +
          'document.documentElement.style.fontSize = ' + JSON.stringify(scale + '%') + ';' +
          'const root = document.documentElement;' +
          'const targets = document.querySelectorAll(".masthead, .masthead h1, .summary dt, .atlas-row, .stats dd, .notice");' +
          'return { width: root.clientWidth, scrollWidth: root.scrollWidth,' +
          'overflow: [...targets].filter(e => e.scrollWidth > e.clientWidth + 1).map(e => e.textContent.trim()) };' +
          '})()');
        const scenario = width + ' CSS px at ' + scale + '% text';
        assert.equal(result.scrollWidth, result.width, 'Document overflow at ' + scenario);
        assert.deepEqual(result.overflow, [], 'Content clipped at ' + scenario);
        console.log(scenario + ': passed');
      }
    }
    assert.deepEqual(requests.filter(url => !url.startsWith('file:')), [], 'Unexpected external requests');
  } finally {
    socket?.close();
    browser.kill();
    await stopped;
    // Chromium child processes can briefly hold the profile open on Windows.
    await rm(profile, { recursive: true, force: true, maxRetries: 6, retryDelay: 200 });
  }
}

checkReflow().catch(error => {
  console.error(error);
  process.exitCode = 1;
});
