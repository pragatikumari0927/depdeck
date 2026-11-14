const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const test = require('node:test');

const templatePath = path.join(__dirname, '..', 'pkg', 'render', 'templates', 'deck.html');
const template = fs.readFileSync(templatePath, 'utf8');

function createDeck(tags = ['dependencies', 'devDependencies', 'optionalDependencies']) {
  const script = template.match(/<script>([\s\S]*?)<\/script>/)?.[1];
  assert.ok(script, 'template has an inline filter script');
  const values = ['', 'dependencies', 'devDependencies', 'optionalDependencies'];
  const buttons = values.map((value) => ({
    value,
    pressed: value === '' ? 'true' : 'false',
    getAttribute(name) {
      if (name === 'data-filter') return this.value;
      if (name === 'aria-pressed') return this.pressed;
      return null;
    },
    setAttribute(name, next) { if (name === 'aria-pressed') this.pressed = next; },
  }));
  const cards = tags.map((tag) => ({
    tag,
    hidden: false,
    getAttribute(name) { return name === 'data-tag' ? this.tag : null; },
  }));
  const empty = { hidden: true };
  let text = tags.length + ' of ' + tags.length + ' in view';
  let updates = 0;
  const count = {
    get textContent() { return text; },
    set textContent(value) { text = value; updates++; },
    get updates() { return updates; },
  };
  let clickHandler;
  const nav = {
    querySelector(selector) { return selector === '[aria-pressed="true"]' ? buttons[0] : null; },
    addEventListener(type, handler) { if (type === 'click') clickHandler = handler; },
    contains(button) { return buttons.includes(button); },
  };
  const document = {
    querySelector(selector) { return selector === '.filters' ? nav : null; },
    getElementById(id) { return { 'view-count': count, 'filter-empty': empty }[id] ?? null; },
    querySelectorAll(selector) { return selector === '.card' ? cards : []; },
  };
  vm.runInNewContext(script, { document });
  return {
    buttons,
    cards,
    count,
    empty,
    click(value) {
      const button = buttons.find((item) => item.value === value);
      assert.ok(button, `filter button exists for ${value}`);
      const child = { closest: (selector) => selector === '[data-filter]' ? button : null };
      clickHandler({ target: child });
    },
  };
}

test('tag filter shows only matching cards and restores all cards', () => {
  const deck = createDeck();
  for (const tag of ['dependencies', 'devDependencies', 'optionalDependencies']) {
    deck.click(tag);
    assert.equal(deck.count.textContent, '1 of 3 in view');
    assert.equal(deck.empty.hidden, true);
    assert.deepEqual(deck.cards.map((card) => card.hidden), [
      tag !== 'dependencies',
      tag !== 'devDependencies',
      tag !== 'optionalDependencies',
    ]);
    assert.deepEqual(deck.buttons.map((button) => button.pressed), [
      'false',
      tag === 'dependencies' ? 'true' : 'false',
      tag === 'devDependencies' ? 'true' : 'false',
      tag === 'optionalDependencies' ? 'true' : 'false',
    ]);
  }
  deck.click('');
  assert.equal(deck.count.textContent, '3 of 3 in view');
  assert.deepEqual(deck.cards.map((card) => card.hidden), [false, false, false]);
});

test('filter feedback distinguishes no matches from an empty collection', () => {
  const deck = createDeck(['dependencies']);
  deck.click('devDependencies');
  assert.equal(deck.count.textContent, '0 of 1 in view');
  assert.equal(deck.cards[0].hidden, true);
  assert.equal(deck.empty.hidden, false);

  deck.click('');
  assert.equal(deck.empty.hidden, true);

  const emptyDeck = createDeck([]);
  emptyDeck.click('optionalDependencies');
  assert.equal(emptyDeck.count.textContent, '0 of 0 in view');
  assert.equal(emptyDeck.empty.hidden, true);
});

test('the live count updates only when the number of visible Cards changes', () => {
  const deck = createDeck();
  deck.click('');
  assert.equal(deck.count.updates, 0);
  deck.click('dependencies');
  assert.equal(deck.count.updates, 1);
  deck.click('dependencies');
  deck.click('devDependencies');
  assert.equal(deck.count.updates, 1);
  assert.deepEqual(deck.cards.map(card => card.hidden), [true, false, true]);
  deck.click('');
  assert.equal(deck.count.updates, 2);
});
