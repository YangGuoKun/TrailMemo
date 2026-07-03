import { config } from '@/config'
import { useAuthStore } from '@/stores/useAuthStore'

const BASE_URL = config.apiBaseUrl
const TIMEOUT = 60000

function getAuthHeader(): Record<string, string> {
  const authStore = useAuthStore()
  if (!authStore.token) return {}
  return { Authorization: authStore.authorizationHeader }
}

export async function chat(message: string, sessionId?: string): Promise<ChatLoopResponse> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}/agent/chat`,
      method: 'POST',
      data: { message, session_id: sessionId },
      timeout: TIMEOUT,
      header: {
        'Content-Type': 'application/json',
        ...getAuthHeader(),
      },
      success: (res) => {
        const result = res.data as ApiResponse<ChatLoopResponse>
        if (result.code === 200) {
          resolve(result.data)
        } else {
          reject(new Error(result.message || '请求失败'))
        }
      },
      fail: (err) => reject(err),
    })
  })
}

type StreamCallback = (chunk: StreamChunk) => void

export async function chatStream(
  message: string,
  sessionId: string,
  onChunk: StreamCallback,
  onComplete?: (error?: string) => void
): Promise<() => void> {
  let aborted = false

  return new Promise((resolve) => {
    const authStore = useAuthStore()
    const token = authStore.token

    const requestTask = uni.request({
      url: `${BASE_URL}/agent/chat/stream`,
      method: 'POST',
      data: { message, session_id: sessionId },
      timeout: 120000, // LLM 流式超时 2 分钟
      enableChunked: true,
      responseType: 'text',
      header: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: authStore.authorizationHeader } : {}),
      },
      success: () => {
        // 流式完成——success 回调在 all chunks received 后触发
        if (!aborted && onComplete) onComplete()
      },
      fail: (err) => {
        if (aborted) return
        const errMsg = err.errMsg || '网络错误'
        onChunk({ error: errMsg })
        if (onComplete) onComplete(errMsg)
      },
      complete: () => {
        resolve(() => { aborted = true })
      },
    })

    // 真正的流式接收：每行一个 JSON 事件
    let buffer = ''
    if (requestTask && typeof (requestTask as any).onChunkReceived === 'function') {
      (requestTask as any).onChunkReceived((res: any) => {
        if (aborted) return
        try {
          const text = typeof res.data === 'string' ? res.data : new TextDecoder('utf-8').decode(new Uint8Array(res.data))
          buffer += text
          // 按行分割，处理完整的 JSON 事件
          const lines = buffer.split('\n')
          buffer = lines.pop() || '' // 最后一行可能不完整，保留到下次
          for (const line of lines) {
            const trimmed = line.trim()
            if (!trimmed) continue
            try {
              const parsed = JSON.parse(trimmed)
              if (parsed.content) onChunk({ content: parsed.content })
              else if (parsed.progress) onChunk({ content: parsed.progress }) // 进度事件也推送
              else if (parsed.error) onChunk({ error: parsed.error })
              else if (parsed.done) onChunk({ content: `[完成，共${parsed.steps}步]` })
            } catch (_) { /* skip */ }
          }
        } catch (_) { /* skip */ }
      })
    }

    resolve(() => {
      aborted = true
      if (requestTask?.abort) { requestTask.abort() }
    })
  })
}

export async function createRouteDraft(data: RouteDraftRequest): Promise<RouteDraftResponse> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}/agent/routes/draft`,
      method: 'POST',
      data,
      timeout: TIMEOUT,
      header: {
        'Content-Type': 'application/json',
        ...getAuthHeader(),
      },
      success: (res) => {
        const result = res.data as ApiResponse<RouteDraftResponse>
        if (result.code === 200) {
          resolve(result.data)
        } else {
          reject(new Error(result.message || '路线生成失败'))
        }
      },
      fail: (err) => reject(err),
    })
  })
}

export async function recommend(data: RecommendRequest): Promise<RecommendResponse> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}/agent/recommend`,
      method: 'POST',
      data,
      timeout: TIMEOUT,
      header: {
        'Content-Type': 'application/json',
        ...getAuthHeader(),
      },
      success: (res) => {
        const result = res.data as ApiResponse<RecommendResponse>
        if (result.code === 200) {
          resolve(result.data)
        } else {
          reject(new Error(result.message || '推荐失败'))
        }
      },
      fail: (err) => reject(err),
    })
  })
}

export async function generateTravelNote(data: TravelNoteRequest): Promise<TravelNoteResponse> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}/agent/notes/generate`,
      method: 'POST',
      data,
      timeout: TIMEOUT,
      header: {
        'Content-Type': 'application/json',
        ...getAuthHeader(),
      },
      success: (res) => {
        const result = res.data as ApiResponse<TravelNoteResponse>
        if (result.code === 200) {
          resolve(result.data)
        } else {
          reject(new Error(result.message || '游记生成失败'))
        }
      },
      fail: (err) => reject(err),
    })
  })
}

export async function remixRoute(routeId: string, data: RemixRequest): Promise<RemixResponse> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}/agent/routes/${routeId}/remix`,
      method: 'POST',
      data,
      timeout: TIMEOUT,
      header: {
        'Content-Type': 'application/json',
        ...getAuthHeader(),
      },
      success: (res) => {
        const result = res.data as ApiResponse<RemixResponse>
        if (result.code === 200) {
          resolve(result.data)
        } else {
          reject(new Error(result.message || '路线改造失败'))
        }
      },
      fail: (err) => reject(err),
    })
  })
}

export async function commitArtifact(artifactId: string, data: ArtifactCommitRequest): Promise<ArtifactCommitResponse> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}/agent/artifacts/${artifactId}/commit`,
      method: 'POST',
      data,
      timeout: TIMEOUT,
      header: {
        'Content-Type': 'application/json',
        ...getAuthHeader(),
      },
      success: (res) => {
        const result = res.data as ApiResponse<ArtifactCommitResponse>
        if (result.code === 200) {
          resolve(result.data)
        } else {
          reject(new Error(result.message || '提交失败'))
        }
      },
      fail: (err) => reject(err),
    })
  })
}

export async function approveArtifact(artifactId: string): Promise<ArtifactApprovalResponse> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}/agent/artifacts/${artifactId}/approve`,
      method: 'POST',
      timeout: TIMEOUT,
      header: {
        'Content-Type': 'application/json',
        ...getAuthHeader(),
      },
      success: (res) => {
        const result = res.data as ApiResponse<ArtifactApprovalResponse>
        if (result.code === 200) {
          resolve(result.data)
        } else {
          reject(new Error(result.message || '确认失败'))
        }
      },
      fail: (err) => reject(err),
    })
  })
}

export async function getRunDetail(runId: string): Promise<RunDetailResponse> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}/agent/runs/${runId}`,
      method: 'GET',
      timeout: TIMEOUT,
      header: {
        'Content-Type': 'application/json',
        ...getAuthHeader(),
      },
      success: (res) => {
        const result = res.data as ApiResponse<RunDetailResponse>
        if (result.code === 200) {
          resolve(result.data)
        } else {
          reject(new Error(result.message || '获取运行详情失败'))
        }
      },
      fail: (err) => reject(err),
    })
  })
}

export async function getPreferences(): Promise<PreferencesResponse> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}/agent/preferences`,
      method: 'GET',
      timeout: TIMEOUT,
      header: {
        'Content-Type': 'application/json',
        ...getAuthHeader(),
      },
      success: (res) => {
        const result = res.data as ApiResponse<PreferencesResponse>
        if (result.code === 200) {
          resolve(result.data)
        } else {
          reject(new Error(result.message || '获取偏好失败'))
        }
      },
      fail: (err) => reject(err),
    })
  })
}

export async function updatePreferences(data: PreferencesUpdateRequest): Promise<PreferencesResponse> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}/agent/preferences`,
      method: 'PUT',
      data,
      timeout: TIMEOUT,
      header: {
        'Content-Type': 'application/json',
        ...getAuthHeader(),
      },
      success: (res) => {
        const result = res.data as ApiResponse<PreferencesResponse>
        if (result.code === 200) {
          resolve(result.data)
        } else {
          reject(new Error(result.message || '更新偏好失败'))
        }
      },
      fail: (err) => reject(err),
    })
  })
}

export async function deleteMemory(): Promise<void> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}/agent/preferences/memory`,
      method: 'DELETE',
      timeout: TIMEOUT,
      header: {
        'Content-Type': 'application/json',
        ...getAuthHeader(),
      },
      success: (res) => {
        const result = res.data as ApiResponse<null>
        if (result.code === 200) {
          resolve()
        } else {
          reject(new Error(result.message || '删除记忆失败'))
        }
      },
      fail: (err) => reject(err),
    })
  })
}

export async function getCapabilities(): Promise<CapabilityResponse> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}/agent/capabilities`,
      method: 'GET',
      timeout: TIMEOUT,
      header: { 'Content-Type': 'application/json' },
      success: (res) => {
        const result = res.data as ApiResponse<CapabilityResponse>
        if (result.code === 200) {
          resolve(result.data)
        } else {
          reject(new Error(result.message || '获取能力列表失败'))
        }
      },
      fail: (err) => reject(err),
    })
  })
}

export async function getHealth(): Promise<HealthResponse> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}/agent/health`,
      method: 'GET',
      timeout: TIMEOUT,
      header: { 'Content-Type': 'application/json' },
      success: (res) => {
        const result = res.data as ApiResponse<HealthResponse>
        if (result.code === 200) {
          resolve(result.data)
        } else {
          reject(new Error(result.message || '健康检查失败'))
        }
      },
      fail: (err) => reject(err),
    })
  })
}

export async function listSessions(): Promise<SessionListResponse> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}/agent/sessions`,
      method: 'GET',
      timeout: TIMEOUT,
      header: {
        'Content-Type': 'application/json',
        ...getAuthHeader(),
      },
      success: (res) => {
        const result = res.data as ApiResponse<SessionListResponse>
        if (result.code === 200) {
          resolve(result.data)
        } else {
          reject(new Error(result.message || '获取会话列表失败'))
        }
      },
      fail: (err) => reject(err),
    })
  })
}

export async function getSession(sessionId: string): Promise<SessionDetailResponse> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}/agent/sessions/${sessionId}`,
      method: 'GET',
      timeout: TIMEOUT,
      header: {
        'Content-Type': 'application/json',
        ...getAuthHeader(),
      },
      success: (res) => {
        const result = res.data as ApiResponse<SessionDetailResponse>
        if (result.code === 200) {
          resolve(result.data)
        } else {
          reject(new Error(result.message || '获取会话详情失败'))
        }
      },
      fail: (err) => reject(err),
    })
  })
}

export async function deleteSession(sessionId: string): Promise<void> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}/agent/sessions/${sessionId}`,
      method: 'DELETE',
      timeout: TIMEOUT,
      header: {
        'Content-Type': 'application/json',
        ...getAuthHeader(),
      },
      success: (res) => {
        const result = res.data as ApiResponse<null>
        if (result.code === 200) {
          resolve()
        } else {
          reject(new Error(result.message || '删除会话失败'))
        }
      },
      fail: (err) => reject(err),
    })
  })
}

export async function renameSession(sessionId: string, title: string): Promise<void> {
  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}/agent/sessions/${sessionId}/title`,
      method: 'PUT',
      data: { title },
      timeout: TIMEOUT,
      header: {
        'Content-Type': 'application/json',
        ...getAuthHeader(),
      },
      success: (res) => {
        const result = res.data as ApiResponse<null>
        if (result.code === 200) {
          resolve()
        } else {
          reject(new Error(result.message || '重命名会话失败'))
        }
      },
      fail: (err) => reject(err),
    })
  })
}
