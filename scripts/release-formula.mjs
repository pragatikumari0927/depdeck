import { createHash } from 'node:crypto';
import { createReadStream } from 'node:fs';
import { mkdir, readFile, writeFile } from 'node:fs/promises';
import { dirname, resolve } from 'node:path';

const [tag, sourceArchive, formulaPath, stagedURL, ...extra] = process.argv.slice(2);
if (!/^v\d+\.\d+\.\d+$/.test(tag ?? '') || !sourceArchive || !formulaPath || extra.length) {
  throw new Error('Usage: node scripts/release-formula.mjs v1.0.0 source.tar.gz Formula/depdeck.rb [loopback-source-url]');
}
const url = stagedURL
  ? new URL(stagedURL)
  : new URL(`https://github.com/pragatikumari0927/depdeck/archive/refs/tags/${tag}.tar.gz`);
if (stagedURL && (url.protocol !== 'http:' || url.hostname !== '127.0.0.1' ||
    url.username || url.password || url.search || url.hash || !/^\/[\w./-]+\.tar\.gz$/.test(url.pathname))) {
  throw new Error('The staged source URL must be a loopback HTTP tar.gz URL');
}

const hash = createHash('sha256');
for await (const chunk of createReadStream(sourceArchive)) hash.update(chunk);
const template = await readFile(new URL('../release/Formula/depdeck.rb.in', import.meta.url), 'utf8');
const formula = template
  .replace('@SOURCE_URL@', url.href)
  .replace('@VERSION@', tag.slice(1))
  .replace('@SHA256@', hash.digest('hex'));
const destination = resolve(formulaPath);
await mkdir(dirname(destination), { recursive: true });
await writeFile(destination, formula);
console.log(destination);