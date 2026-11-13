const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const test = require('node:test');

const templatePath = path.join(__dirname, '..', 'pkg', 'render', 'templates', 'deck.html');
const template = fs.readFileSync(templatePath, 'utf8');

function createDeck() {
  const script = template.match(/<script>([\s\S]*?)<\/script>/)?.[1];
  assert.ok(script, 'template has an inline filter script');
  const values = ['', 'dependencies', 'devDependencies', 'optionalDependencies'];
  const buttons = values.map((value) => ({
    value,
    pressed: value === '' ? 'true' : 'false',
    getAttribute(name) { return name === 'data-filter' ? this.value : null; },
    setAttribute(name, next) { if (name === 'aria-pressed') this.pressed = next; },
  }));
  const cards = ['dependencies', 'devDependencies', 'optionalDependencies'].map((tag) => ({
    tag,
    hidden: false,
    getAttribute(name) { return name === 'data-tag' ? this.tag : null; },
  }));
  let clickHandler;
  const nav = {
    querySelectorAll(selector) { return selector === '[data-filter]' ? buttons : []; },
    addEventListener(type, handler) { if (type === 'click') clickHandler = handler; },
    contains(button) { return buttons.includes(button); },
  };
  const document = {
    querySelector(selector) { return selector === '.filters' ? nav : null; },
    querySelectorAll(selector) { return selector === '.card' ? cards : []; },
  };
  vm.runInNewContext(script, { document });
  return {
    buttons,
    cards,
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
  assert.deepEqual(deck.cards.map((card) => card.hidden), [false, false, false]);
});
