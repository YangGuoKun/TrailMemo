import assert from 'node:assert/strict'
import { execFileSync } from 'node:child_process'
import { mkdirSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { dirname, resolve } from 'node:path'
import { createRequire } from 'node:module'
import { fileURLToPath } from 'node:url'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const outDir = resolve(tmpdir(), `trailmemo-markdown-${Date.now()}`)

mkdirSync(outDir, { recursive: true })

execFileSync(
  process.execPath,
  [
    resolve(root, 'node_modules', 'typescript', 'bin', 'tsc'),
    '--target', 'ES2020',
    '--module', 'CommonJS',
    '--moduleResolution', 'node',
    '--outDir', outDir,
    '--skipLibCheck',
    resolve(root, 'src', 'utils', 'markdown.ts'),
  ],
  { stdio: 'inherit' },
)

const require = createRequire(import.meta.url)
const mod = require(resolve(outDir, 'markdown.js'))

const rendered = mod.renderMarkdownToRichText('# 广州路线\n\n- **沙面**\n- `永庆坊`')
assert.equal(rendered.includes('<h1 class="md-h1">广州路线</h1>'), true)
assert.equal(
  rendered.includes(
    '<ul class="md-ul"><li class="md-li"><strong class="md-strong">沙面</strong></li><li class="md-li"><code class="md-code">永庆坊</code></li></ul>',
  ),
  true,
)

const escaped = mod.renderMarkdownToRichText('<script>alert(1)</script>\n\n**bold**')
assert.equal(escaped.includes('<script>'), false)
assert.equal(escaped.includes('&lt;script&gt;alert(1)&lt;/script&gt;'), true)
assert.equal(escaped.includes('<strong class="md-strong">bold</strong>'), true)

console.log('markdown tests passed')
