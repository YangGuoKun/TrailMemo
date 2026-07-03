import assert from 'node:assert/strict'
import { execFileSync } from 'node:child_process'
import { mkdirSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { dirname, resolve } from 'node:path'
import { createRequire } from 'node:module'
import { fileURLToPath } from 'node:url'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const outDir = resolve(tmpdir(), `trailmemo-agent-failure-${Date.now()}`)

mkdirSync(outDir, { recursive: true })

execFileSync(
  process.execPath,
  [
    resolve(root, 'node_modules', 'typescript', 'bin', 'tsc'),
    '--target',
    'ES2020',
    '--module',
    'CommonJS',
    '--moduleResolution',
    'node',
    '--outDir',
    outDir,
    '--skipLibCheck',
    resolve(root, 'src', 'utils', 'agentFailure.ts'),
  ],
  { stdio: 'inherit' },
)

const require = createRequire(import.meta.url)
const mod = require(resolve(outDir, 'agentFailure.js'))

const mapFailure = mod.buildAgentFailureInfo({
  status: 'failed',
  error_message: '高德POI搜索失败: INVALID_USER_KEY',
  steps: [
    { index: 1, type: 'validation', name: 'input_guardrail', status: 'success' },
    { index: 2, type: 'tool', name: 'map.poi_search', status: 'failed', latency_ms: 120 },
  ],
})

assert.equal(mapFailure.visible, true)
assert.equal(mapFailure.title, '生成失败')
assert.equal(mapFailure.summary, '外部工具或地图数据查询失败')
assert.equal(mapFailure.failedSteps.length, 1)
assert.match(mapFailure.suggestion, /重试|关键词/)

const schemaFailure = mod.buildAgentFailureInfo({
  status: 'completed',
  warnings: ['LLM输出JSON格式校验失败，尝试修复'],
  steps: [{ index: 6, type: 'validation', name: 'output_schema', status: 'failed' }],
})

assert.equal(schemaFailure.visible, true)
assert.equal(schemaFailure.title, '部分步骤失败')
assert.equal(schemaFailure.summary, '返回格式未通过校验')
assert.match(schemaFailure.suggestion, /重试/)

const authFailure = mod.buildAgentFailureInfo('unauthorized 401')
assert.equal(authFailure.summary, '登录状态失效')
assert.match(authFailure.suggestion, /重新登录/)

const empty = mod.buildAgentFailureInfo(null)
assert.equal(empty.visible, false)

console.log('agent failure tests passed')
