export type AgentFailureStep = {
  index: number
  type: string
  name: string
  status: string
  latency_ms?: number
}

export type AgentFailureRun = {
  run_id?: string
  status?: string
  error_code?: string
  error_message?: string
  steps?: AgentFailureStep[]
  warnings?: string[]
}

export type AgentFailureInfo = {
  visible: boolean
  title: string
  summary: string
  reason: string
  failedSteps: AgentFailureStep[]
  suggestion: string
}

export function buildAgentFailureInfo(input: AgentFailureRun | string | Error | null | undefined): AgentFailureInfo {
  if (!input) return emptyFailureInfo()

  if (typeof input === 'string') {
    return fromMessage(input)
  }

  if (input instanceof Error) {
    return fromMessage(input.message)
  }

  const failedSteps = (input.steps || []).filter((step) => step.status === 'failed')
  const warnings = input.warnings || []
  const reason = input.error_message || warnings[0] || inferReasonFromSteps(failedSteps) || 'Agent 运行未完成'
  const title = input.status === 'failed' ? '生成失败' : failedSteps.length > 0 ? '部分步骤失败' : '需要确认'

  return {
    visible: Boolean(input.error_message || warnings.length > 0 || failedSteps.length > 0 || input.status === 'failed'),
    title,
    summary: summarizeReason(reason),
    reason,
    failedSteps,
    suggestion: buildSuggestion(reason, failedSteps),
  }
}

function emptyFailureInfo(): AgentFailureInfo {
  return {
    visible: false,
    title: '',
    summary: '',
    reason: '',
    failedSteps: [],
    suggestion: '',
  }
}

function fromMessage(message: string): AgentFailureInfo {
  const reason = message.trim() || '请求失败'
  return {
    visible: true,
    title: '请求失败',
    summary: summarizeReason(reason),
    reason,
    failedSteps: [],
    suggestion: buildSuggestion(reason, []),
  }
}

function inferReasonFromSteps(steps: AgentFailureStep[]): string {
  if (steps.length === 0) return ''
  const first = steps[0]
  if (first.type === 'llm') return '模型生成步骤失败'
  if (first.type === 'tool') return `${first.name || '工具'} 调用失败`
  if (first.type === 'validation') return '输出格式校验失败'
  if (first.type === 'approval') return '确认流程未完成'
  return `${first.name || first.type} 执行失败`
}

function summarizeReason(reason: string): string {
  if (reason.includes('JSON') || reason.includes('格式') || reason.includes('schema') || reason.includes('校验')) {
    return '返回格式未通过校验'
  }
  if (reason.includes('LLM') || reason.includes('AI') || reason.includes('模型')) {
    return '模型服务暂时没有返回可用结果'
  }
  if (reason.includes('POI') || reason.includes('高德') || reason.includes('地图') || reason.includes('tool') || reason.includes('工具')) {
    return '外部工具或地图数据查询失败'
  }
  if (reason.includes('登录') || reason.includes('401') || reason.includes('unauthorized')) {
    return '登录状态失效'
  }
  return reason.length > 34 ? `${reason.slice(0, 34)}...` : reason
}

function buildSuggestion(reason: string, steps: AgentFailureStep[]): string {
  const text = reason.toLowerCase()
  if (reason.includes('登录') || text.includes('unauthorized') || reason.includes('401')) {
    return '请重新登录后再试。'
  }
  if (reason.includes('POI') || reason.includes('高德') || reason.includes('地图')) {
    return '可以稍后重试，或换一个城市/关键词让系统重新查找点位。'
  }
  if (reason.includes('JSON') || reason.includes('格式') || reason.includes('schema') || reason.includes('校验')) {
    return '可以直接重试一次，系统会重新生成结构化路线。'
  }
  if (steps.some((step) => step.type === 'tool')) {
    return '工具查询失败，可以重试或查看运行详情定位具体工具。'
  }
  return '请稍后重试；如果持续失败，打开运行详情查看失败步骤。'
}
