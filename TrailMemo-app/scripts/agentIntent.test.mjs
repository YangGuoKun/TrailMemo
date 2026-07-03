import assert from 'node:assert/strict'
import { execFileSync } from 'node:child_process'
import { mkdirSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { dirname, resolve } from 'node:path'
import { createRequire } from 'node:module'
import { fileURLToPath } from 'node:url'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const outDir = resolve(tmpdir(), `trailmemo-agent-intent-${Date.now()}`)

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
    resolve(root, 'src', 'utils', 'agentIntent.ts'),
  ],
  { stdio: 'inherit' },
)

const require = createRequire(import.meta.url)
const mod = require(resolve(outDir, 'agentIntent.js'))

assert.equal(
  mod.shouldUseRouteDraftWorkflow('帮我设计一条广州的游玩路线，并添加到我的打卡路线中'),
  true,
)
assert.equal(
  mod.shouldUseRouteDraftWorkflow('美食体验，一天，情侣出行，优先地铁步行，请直接生成路线并添加到我的打卡列表'),
  true,
)
assert.equal(mod.shouldUseRouteDraftWorkflow('生成一份 广州一日citywalk的打卡路线给我'), true)
assert.equal(mod.shouldUseRouteDraftWorkflow('做一份打卡路线'), true)
assert.equal(mod.shouldUseRouteDraftWorkflow('添加至我的个人打卡列表'), true)
assert.equal(mod.shouldUseRouteDraftWorkflow('杭州三日游攻略'), true)
assert.equal(mod.shouldUseRouteDraftWorkflow('推荐夏天避暑的旅行路线'), false)
assert.equal(mod.shouldUseRouteDraftWorkflow('你好，今天心情不错'), false)

assert.deepEqual(mod.buildRouteDraftRequest('杭州三日游攻略'), {
  query: '杭州三日游攻略',
  days: 3,
})
assert.deepEqual(mod.buildRouteDraftRequest('广州一日游，美食体验'), {
  query: '广州一日游，美食体验',
  days: 1,
  travel_styles: ['美食'],
})
assert.deepEqual(mod.buildRouteDraftRequest('生成一份 广州一日citywalk的打卡路线给我'), {
  query: '生成一份 广州一日citywalk的打卡路线给我',
  days: 1,
  travel_styles: ['Citywalk'],
})

assert.deepEqual(mod.buildRouteDraftRequest('做一份打卡路线'), {
  query: '做一份打卡路线',
})

console.log('agent intent tests passed')
