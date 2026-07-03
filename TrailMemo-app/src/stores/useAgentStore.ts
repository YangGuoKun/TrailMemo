import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as agentApi from '@/api/agent'
import { buildRouteDraftRequest, shouldUseRouteDraftWorkflow } from '@/utils/agentIntent'
import { buildAgentFailureInfo, type AgentFailureInfo } from '@/utils/agentFailure'

interface MessageWithData extends ChatMessage {
  recommendData?: RecommendResponse
  routeDraftData?: RouteDraftResponse
  routeDraftStatus?: 'pending' | 'committed' | 'cancelled'
  failureInfo?: AgentFailureInfo
  runId?: string
}

export const useAgentStore = defineStore('agent', () => {
  const messages = ref<MessageWithData[]>([])
  const sessionId = ref('')
  const isStreaming = ref(false)
  const loading = ref(false)
  const error = ref('')
  
  const sessions = ref<SessionInfo[]>([])
  const loadingSessions = ref(false)

  const quickQuestions: QuickQuestion[] = [
    { label: '夏天避暑', icon: '🏖️', query: '推荐夏天避暑的旅行路线', type: 'recommend' },
    { label: '成都美食', icon: '🍜', query: '推荐成都美食路线', type: 'recommend' },
    { label: '情侣旅行', icon: '💕', query: '推荐适合情侣的浪漫旅行路线', type: 'recommend' },
    { label: '杭州攻略', icon: '📝', query: '杭州三日游攻略', type: 'route' },
  ]

  const hasMessages = computed(() => messages.value.length > 0)

  const currentSession = computed(() => {
    return sessions.value.find(s => s.session_id === sessionId.value)
  })

  function generateSessionId(): string {
    return 'session_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9)
  }

  function isAuthError(errMsg: string): boolean {
    return errMsg.includes('unauthorized') || errMsg.includes('未登录') || errMsg.includes('401')
  }

  function getAuthErrorContent(errMsg: string): string {
    return isAuthError(errMsg) ? '请先登录后再使用 AI 助手功能。' : '抱歉，AI 服务暂时不可用，请稍后再试。'
  }

  function createUserMessage(content: string): MessageWithData {
    return {
      id: 'user_' + Date.now(),
      role: 'user',
      content: content.trim(),
      timestamp: Date.now(),
    }
  }

  async function loadSessions() {
    loadingSessions.value = true
    try {
      const result = await agentApi.listSessions()
      sessions.value = result.sessions
    } catch (_) {
      sessions.value = []
    } finally {
      loadingSessions.value = false
    }
  }

  function initSession() {
    messages.value = []
    sessionId.value = generateSessionId()
    error.value = ''
    
    messages.value.push({
      id: 'welcome',
      role: 'assistant',
      content: '你好！我是 AI 旅行助手，可以帮你推荐路线、生成攻略、规划行程。有什么我可以帮你的吗？',
      timestamp: Date.now(),
    })
  }

  async function createNewSession() {
    messages.value = []
    sessionId.value = ''
    error.value = ''
    
    messages.value.push({
      id: 'welcome',
      role: 'assistant',
      content: '你好！我是 AI 旅行助手，可以帮你推荐路线、生成攻略、规划行程。有什么我可以帮你的吗？',
      timestamp: Date.now(),
    })
    
    await loadSessions()
  }

  async function switchSession(session: SessionInfo) {
    loading.value = true
    error.value = ''
    
    try {
      const result = await agentApi.getSession(session.session_id)
      
      sessionId.value = session.session_id
      
      messages.value = result.messages.map((msg, idx) => ({
        id: `${msg.role}_${idx}`,
        role: msg.role as 'user' | 'assistant',
        content: msg.content,
        timestamp: Date.now(),
      }))
      
      if (result.expired) {
        messages.value.push({
          id: 'expired',
          role: 'assistant',
          content: '会话已过期，请重新开始对话。',
          timestamp: Date.now(),
        })
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : '加载会话失败'
      sessionId.value = ''
      messages.value = []
    } finally {
      loading.value = false
    }
  }

  async function deleteSessionById(sessionIdToDelete: string) {
    try {
      await agentApi.deleteSession(sessionIdToDelete)
      
      const idx = sessions.value.findIndex(s => s.session_id === sessionIdToDelete)
      if (idx > -1) {
        sessions.value.splice(idx, 1)
      }
      
      if (sessionId.value === sessionIdToDelete) {
        await createNewSession()
      }
      
      return true
    } catch (_) {
      return false
    }
  }

  async function renameSessionById(sessionIdToRename: string, newTitle: string) {
    try {
      await agentApi.renameSession(sessionIdToRename, newTitle)
      
      const session = sessions.value.find(s => s.session_id === sessionIdToRename)
      if (session) {
        session.title = newTitle
      }
      
      return true
    } catch (_) {
      return false
    }
  }

  async function sendMessage(message: string): Promise<void> {
    if (!message.trim() || isStreaming.value) return

    error.value = ''
    const text = message.trim()
    
    messages.value.push(createUserMessage(text))

    if (shouldUseRouteDraftWorkflow(text)) {
      loading.value = true
      
      if (!sessionId.value) {
        sessionId.value = generateSessionId()
      }
      
      const assistantMessage: MessageWithData = {
        id: 'assistant_' + Date.now(),
        role: 'assistant',
        content: '',
        timestamp: Date.now(),
      }
      messages.value.push(assistantMessage)
      
      try {
        const result = await agentApi.createRouteDraft({
          ...(buildRouteDraftRequest(text) as RouteDraftRequest),
          session_id: sessionId.value
        })
        assistantMessage.content = '我整理成了一条可加入个人打卡路线的方案，请确认是否加入。'
        assistantMessage.routeDraftData = result
        assistantMessage.routeDraftStatus = 'pending'
        await loadSessions()
      } catch (err) {
        const errMsg = err instanceof Error ? err.message : '路线生成失败'
        error.value = errMsg
        assistantMessage.content = '抱歉，路线生成失败，请稍后再试。'
        assistantMessage.failureInfo = buildAgentFailureInfo(errMsg)
      } finally {
        loading.value = false
      }
      return
    }

    if (!sessionId.value) {
      sessionId.value = generateSessionId()
    }

    isStreaming.value = true
    
    const assistantMessage: ChatMessage = {
      id: 'assistant_' + Date.now(),
      role: 'assistant',
      content: '',
      timestamp: Date.now(),
    }
    messages.value.push(assistantMessage)

    try {
      await agentApi.chatStream(
        text,
        sessionId.value,
        (chunk) => {
          if (chunk.error) {
            error.value = chunk.error
            assistantMessage.content = getAuthErrorContent(chunk.error)
            ;(assistantMessage as MessageWithData).failureInfo = buildAgentFailureInfo(chunk.error)
            isStreaming.value = false
            return
          }
          if (chunk.content) {
            assistantMessage.content += chunk.content
          }
        },
        () => {
          isStreaming.value = false
          loadSessions()
        }
      )
    } catch (err) {
      const errMsg = err instanceof Error ? err.message : '发送失败'
      error.value = errMsg
      assistantMessage.content = isAuthError(errMsg)
        ? '请先登录后再使用 AI 助手功能。'
        : '抱歉，发送失败，请检查网络后重试。'
      ;(assistantMessage as MessageWithData).failureInfo = buildAgentFailureInfo(errMsg)
      isStreaming.value = false
    }
  }

  async function createRouteDraft(data: RouteDraftRequest): Promise<RouteDraftResponse> {
    loading.value = true
    try {
      return await agentApi.createRouteDraft(data)
    } finally {
      loading.value = false
    }
  }

  async function recommend(data: RecommendRequest): Promise<RecommendResponse> {
    loading.value = true
    try {
      return await agentApi.recommend(data)
    } finally {
      loading.value = false
    }
  }

  async function generateTravelNote(data: TravelNoteRequest): Promise<TravelNoteResponse> {
    loading.value = true
    try {
      return await agentApi.generateTravelNote(data)
    } finally {
      loading.value = false
    }
  }

  async function remixRoute(routeId: string, data: RemixRequest): Promise<RemixResponse> {
    loading.value = true
    try {
      return await agentApi.remixRoute(routeId, data)
    } finally {
      loading.value = false
    }
  }

  async function commitArtifact(artifactId: string, data: ArtifactCommitRequest): Promise<ArtifactCommitResponse> {
    loading.value = true
    try {
      return await agentApi.commitArtifact(artifactId, data)
    } finally {
      loading.value = false
    }
  }

  async function approveArtifact(artifactId: string): Promise<ArtifactApprovalResponse> {
    loading.value = true
    try {
      return await agentApi.approveArtifact(artifactId)
    } finally {
      loading.value = false
    }
  }

  async function getRunDetail(runId: string): Promise<RunDetailResponse> {
    const detail = await agentApi.getRunDetail(runId)
    const msg = messages.value.find(item => item.routeDraftData?.run_id === runId || item.runId === runId)
    if (msg) {
      const failureInfo = buildAgentFailureInfo({
        status: detail.status,
        error_code: detail.error_code,
        error_message: detail.error_message,
        steps: detail.steps,
      })
      if (failureInfo.visible) msg.failureInfo = failureInfo
    }
    return detail
  }

  function markRouteDraftStatus(artifactId: string, status: 'committed' | 'cancelled') {
    const msg = messages.value.find(item => item.routeDraftData?.artifact_id === artifactId)
    if (msg) {
      msg.routeDraftStatus = status
      msg.content = status === 'committed'
        ? '已加入我的路线。'
        : '已取消加入这条路线。'
    }
  }

  async function handleQuickAction(q: QuickQuestion): Promise<void> {
    if (!q.query || loading.value) return

    error.value = ''
    messages.value.push(createUserMessage(q.query))
    loading.value = true

    try {
      if (q.type === 'recommend') {
        const result = await agentApi.recommend({ query: q.query })
        
        const assistantMessage: MessageWithData = {
          id: 'assistant_' + Date.now(),
          role: 'assistant',
          content: result.items.map(item => `${item.title} · ${item.city} · ${item.days}天 · ¥${item.estimated_budget}`).join('\n'),
          timestamp: Date.now(),
          recommendData: result,
        }
        messages.value.push(assistantMessage)
      } else if (q.type === 'route') {
        const result = await agentApi.createRouteDraft({ query: q.query })
        
        const assistantMessage: MessageWithData = {
          id: 'assistant_' + Date.now(),
          role: 'assistant',
          content: result.route_draft.summary,
          timestamp: Date.now(),
          routeDraftData: result,
        }
        messages.value.push(assistantMessage)
      } else {
        await sendMessage(q.query)
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : '请求失败'
      const errorMessage: MessageWithData = {
        id: 'assistant_' + Date.now(),
        role: 'assistant',
        content: '抱歉，请求失败，请稍后再试。',
        timestamp: Date.now(),
        failureInfo: buildAgentFailureInfo(error.value),
      }
      messages.value.push(errorMessage)
    } finally {
      loading.value = false
    }
  }

  function clearMessages() {
    messages.value = []
    sessionId.value = ''
    error.value = ''
  }

  return {
    messages,
    sessionId,
    isStreaming,
    loading,
    error,
    sessions,
    loadingSessions,
    quickQuestions,
    hasMessages,
    currentSession,
    initSession,
    createNewSession,
    loadSessions,
    switchSession,
    deleteSessionById,
    renameSessionById,
    sendMessage,
    createRouteDraft,
    recommend,
    generateTravelNote,
    remixRoute,
    approveArtifact,
    commitArtifact,
    getRunDetail,
    markRouteDraftStatus,
    handleQuickAction,
    clearMessages,
  }
})
