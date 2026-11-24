const assert = require('node:assert/strict');
const { spawn, spawnSync } = require('node:child_process');
const { mkdir, mkdtemp, readFile, rm, writeFile } = require('node:fs/promises');
const { tmpdir } = require('node:os');
const path = require('node:path');
const { setTimeout: delay } = require('node:timers/promises');
const { pathToFileURL } = require('node:url');

const [chrome] = process.argv.slice(2);
assert.ok(chrome, 'Usage: node scripts/verify-deck-reflow.cjs <Chromium executable>');
const repo = path.resolve(__dirname, '..');
const expected = {
  '': ['runtime-legendary', 'runtime-common', "<script>alert('fixture')</script>", 'development-epic', 'development-unknown', 'optional-rare', 'optional-unknown'],
  dependencies: ['runtime-legendary', 'runtime-common', "<script>alert('fixture')</script>"],
  devDependencies: ['development-epic', 'development-unknown'],
  optionalDependencies: ['optional-rare', 'optional-unknown'],
};
const expectedAtlas = { Legendary: '1 Card', Epic: '1 Card', Rare: '1 Card', Common: '1 Card', Unknown: '3 Cards' };

function makeFixture() {
  const result = spawnSync('go', ['run', './scripts/browser-fixture'], {
    cwd: repo,
    encoding: 'buffer',
    maxBuffer: 4 * 1024 * 1024,
    windowsHide: true,
  });
  assert.equal(result.status, 0, `Go fixture renderer failed:\n${result.stderr?.toString() ?? ''}`);
  return result.stdout;
}

async function checkBrowser() {
  const profile = await mkdtemp(path.join(tmpdir(), 'depdeck-browser-'));
  const firstDeck = path.join(profile, 'deck.html');
  const first = makeFixture();
  const second = makeFixture();
  assert.deepEqual(first, second, 'identical fixture assessments must produce identical HTML bytes');
  await writeFile(firstDeck, first);

  const browserProfile = path.join(profile, 'chromium-profile');
  await mkdir(browserProfile);
  const browserArgs = [
    '--headless', '--remote-debugging-port=0', '--disable-background-networking',
    '--disable-crash-reporter', '--disable-breakpad', '--no-first-run',
    '--disable-extensions', '--disable-default-apps',
    '--user-data-dir=' + browserProfile, 'about:blank',
  ];
  if (process.env.DEPDECK_CHROMIUM_NO_SANDBOX === '1') browserArgs.push('--no-sandbox');
  const browser = spawn(chrome, browserArgs, { windowsHide: true, stdio: ['ignore', 'ignore', 'pipe'] });
  let browserDiagnostics = '';
  browser.stderr.setEncoding('utf8');
  browser.stderr.on('data', chunk => { browserDiagnostics += chunk; });
  const stopped = new Promise(resolve => {
    browser.once('error', resolve);
    browser.once('exit', resolve);
  });
  let socket;
  try {
    let page;
    for (let attempt = 0; attempt < 120 && !page; attempt++) {
      try {
        const portFile = await readFile(path.join(browserProfile, 'DevToolsActivePort'), 'utf8');
        const port = portFile.split('\n')[0];
        const response = await fetch('http://127.0.0.1:' + port + '/json/list', { signal: AbortSignal.timeout(1000) });
        page = (await response.json()).find(item => item.type === 'page');
      } catch {
        await delay(100);
      }
    }
    assert.ok(page, 'Chromium DevTools did not start. ' + browserDiagnostics);
    socket = new WebSocket(page.webSocketDebuggerUrl);
    await new Promise((resolve, reject) => {
      socket.addEventListener('open', resolve, { once: true });
      socket.addEventListener('error', reject, { once: true });
    });
    const pending = new Map();
    const requests = [];
    const failedRequests = [];
    const requestURLs = new Map();
    let id = 0;
    socket.addEventListener('message', ({ data }) => {
      const message = JSON.parse(data);
      if (message.method === 'Network.requestWillBeSent') {
        requests.push(message.params.request.url);
        requestURLs.set(message.params.requestId, message.params.request.url);
      }
      if (message.method === 'Network.loadingFailed') failedRequests.push(requestURLs.get(message.params.requestId));
      pending.get(message.id)?.(message);
    });
    async function send(method, params = {}) {
      const request = { id: ++id, method, params };
      let timer;
      try {
        const message = await new Promise((resolve, reject) => {
          timer = setTimeout(() => reject(new Error(method + ' timed out')), 10000);
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
    async function clickFilter(tag) {
      const encoded = JSON.stringify(tag);
      await evaluate('[...document.querySelectorAll(".filters button")].find(button => button.getAttribute("data-filter") === ' + encoded + ').click()');
    }
    async function activeAtlas() {
      return evaluate('Object.fromEntries([...document.querySelectorAll(".atlas-row")].map(row => [row.dataset.rarity, row.querySelector(".atlas-count").textContent.trim()]))');
    }
    async function visibleNames() {
      return evaluate('[...document.querySelectorAll(".card:not([hidden]) .name")].map(node => node.textContent)');
    }

    await send('Network.enable');
    await send('Network.setBlockedURLs', { urls: ['http://*', 'https://*'] });
    const navigation = await send('Page.navigate', { url: pathToFileURL(firstDeck).href });
    assert.ok(!navigation.errorText, navigation.errorText);
    for (let attempt = 0; attempt < 120; attempt++) {
      if (await evaluate('document.readyState === "complete" && Boolean(document.querySelector(".masthead"))')) break;
      await delay(50);
    }
    assert.equal(await evaluate('document.querySelectorAll(".card").length'), 7, 'fixture should contain seven Cards');
    assert.deepEqual(await activeAtlas(), expectedAtlas, 'Atlas counts must match the complete assessed Deck');

    const semantics = await evaluate('(() => ({' +
      'main: document.querySelectorAll("main").length,' +
      'navigation: document.querySelector("[role=group][aria-label=\\"Filter Cards by dependency group\\"]") !== null,' +
      'buttons: [...document.querySelectorAll(".filters button")].map(button => ({name: button.textContent.trim(), type: button.type, pressed: button.getAttribute("aria-pressed")})),' +
      'headings: [...document.querySelectorAll("h1,h2,h3")].map(heading => heading.tagName)' +
    '}))()');
    assert.equal(semantics.main, 1, 'Deck has one main landmark');
    assert.equal(semantics.navigation, true, 'filters have an accessible group label');
    assert.deepEqual(semantics.buttons.map(button => button.name), ['All Cards', 'Runtime', 'Development', 'Optional']);
    assert.ok(semantics.buttons.every(button => button.type === 'button'));
    assert.ok(semantics.headings.includes('H1') && semantics.headings.includes('H2') && semantics.headings.includes('H3'));

    const hostile = await evaluate('({' +
      'name: [...document.querySelectorAll(".name")].find(node => node.textContent.includes("<script>"))?.textContent,' +
      'version: [...document.querySelectorAll(".version")].find(node => node.textContent.includes("<img"))?.textContent,' +
      'scripts: document.scripts.length,' +
      'injectedImages: [...document.images].filter(image => image.getAttribute("src") === "x").length' +
    '})');
    assert.equal(hostile.name, "<script>alert('fixture')</script>", 'hostile package name remains text');
    assert.ok(hostile.version.includes('<img src=x onerror=alert(1)>'), 'hostile declared version remains text');
    assert.equal(hostile.scripts, 1, 'only the static filter script exists');
    assert.equal(hostile.injectedImages, 0, 'hostile version did not create an image element');
    assert.ok(first.toString('utf8').includes('&lt;script&gt;alert(&#39;fixture&#39;)&lt;/script&gt;'), 'source HTML contains escaped hostile name');

    for (const [tag, names] of Object.entries(expected)) {
      await clickFilter(tag);
      const actual = await visibleNames();
      assert.deepEqual([...actual].sort(), [...names].sort(), `visible Cards for filter ${tag || 'all'}`);
      assert.deepEqual(await activeAtlas(), expectedAtlas, `Atlas totals changed under filter ${tag || 'all'}`);
      assert.equal(await evaluate('document.querySelector(".filters [aria-pressed=\\"true\\"]").getAttribute("data-filter")'), tag);
      const visibleCount = names.length;
      assert.equal(await evaluate('document.getElementById("view-count").textContent'), `${visibleCount} of 7 in view`);
      assert.equal(await evaluate('([...document.querySelectorAll(".card")].every(card => card.hidden ? card.getBoundingClientRect().height === 0 : card.getBoundingClientRect().height > 0))'), true,
        `visible and hidden Cards have correct layout for filter ${tag || 'all'}`);
    }

    await clickFilter('');
    let focusedFilter = false;
    for (let attempt = 0; attempt < 24; attempt++) {
      await send('Input.dispatchKeyEvent', { type: 'keyDown', key: 'Tab', code: 'Tab', windowsVirtualKeyCode: 9 });
      await send('Input.dispatchKeyEvent', { type: 'keyUp', key: 'Tab', code: 'Tab', windowsVirtualKeyCode: 9 });
      focusedFilter = await evaluate('(() => { const el = document.activeElement; return Boolean(el?.matches(".filters button") && getComputedStyle(el).outlineStyle !== "none" && getComputedStyle(el).outlineWidth !== "0px"); })()');
      if (focusedFilter) break;
    }
    assert.equal(focusedFilter, true, 'keyboard navigation reaches a filter with a visible focus outline');

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
    const external = requests.filter(url => /^https?:/i.test(url));
    const failedExternal = failedRequests.filter(url => url && /^https?:/i.test(url));
    assert.deepEqual(external, [], 'Deck attempted external HTTP requests');
    assert.deepEqual(failedExternal, [], 'Deck attempted blocked external requests');
    console.log('No external network requests: passed');
    console.log('Filters, Atlas totals, keyboard focus, semantics, hostile text, and deterministic output: passed');
  } finally {
    socket?.close();
    browser.kill();
    await stopped;
    await rm(profile, { recursive: true, force: true, maxRetries: 6, retryDelay: 200 });
  }
}

checkBrowser().catch(error => {
  console.error(error);
  process.exitCode = 1;
});
