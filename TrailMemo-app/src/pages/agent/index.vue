<template>
  <view class="agent-page">
    <view class="sidebar-mask" v-if="showSidebar" @tap="toggleSidebar" />
    
    <view class="sidebar" :class="{ show: showSidebar }">
      <view class="sidebar-header">
        <text class="sidebar-title">🤖 AI 助手</text>
        <text class="sidebar-close" @tap="toggleSidebar">✕</text>
      </view>
      
      <view class="sidebar-action">
        <view class="new-chat-btn" @tap="handleNewChat">
          <text class="new-chat-icon">+</text>
          <text class="new-chat-text">新对话</text>
        </view>
      </view>
      
      <scroll-view scroll-y class="sessions-list">
        <view v-if="store.loadingSessions" class="loading-sessions">
          <text class="loading-text">加载中...</text>
        </view>
        
        <view v-else-if="store.sessions.length === 0" class="empty-sessions">
          <text class="empty-icon">📭</text>
          <text class="empty-text">暂无会话</text>
        </view>
        
        <view
          v-for="session in store.sessions"
          :key="session.session_id"
          class="session-item"
          :class="{ active: store.sessionId === session.session_id }"
          @tap="handleSwitchSession(session)"
        >
          <view class="session-content">
            <text class="session-title">{{ session.title }}</text>
            <view class="session-meta">
              <text class="session-time">{{ formatSessionTime(session.last_message_at) }}</text>
              <text class="session-count">{{ session.message_count }}条</text>
            </view>
          </view>
          <view class="session-delete" @tap.stop="handleDeleteSession(session)">
            <text class="delete-icon">×</text>
          </view>
        </view>
      </scroll-view>
    </view>

    <view class="main-content">
      <view class="navbar">
        <view class="nav-left">
          <view class="back-btn" @tap="goBack">
            <text class="back-icon">←</text>
          </view>
          <view class="menu-btn" @tap="toggleSidebar">
            <text class="menu-icon">☰</text>
          </view>
        </view>
        <view class="nav-center">
          <text class="nav-title">{{ store.currentSession?.title || 'AI 旅行助手' }}</text>
        </view>
        <view class="nav-right">
          <text class="clear-btn" @tap="clearChat" v-if="store.hasMessages">清空</text>
        </view>
      </view>

      <scroll-view
        scroll-y
        class="chat-container"
        :scroll-into-view="scrollToId"
        scroll-with-animation
      >
        <view class="quick-chips">
          <view
            v-for="q in store.quickQuestions"
            :key="q.label"
            class="chip"
            @tap="sendQuickQuestion(q)"
          >
            <text class="chip-icon">{{ q.icon }}</text>
            <text class="chip-text">{{ q.label }}</text>
          </view>
        </view>

        <view v-if="store.loading" class="loading-chat">
          <text class="loading-text">加载中...</text>
        </view>

        <view v-else-if="!store.hasMessages && !store.isStreaming" class="empty-chat">
          <view class="empty-icon">✨</view>
          <text class="empty-title">AI 旅行助手</text>
          <text class="empty-desc">告诉我你的旅行需求，我来帮你规划完美行程</text>
          <view class="empty-features">
            <text>📍 智能路线推荐</text>
            <text>📝 个性化行程攻略</text>
            <text>🍜 美食景点推荐</text>
          </view>
        </view>

        <view
          v-for="msg in store.messages"
          :key="msg.id"
          :id="'msg-' + msg.id"
          class="message-item"
          :class="{ 'is-user': msg.role === 'user' }"
        >
          <view class="msg-avatar">
            <text>{{ msg.role === 'user' ? '🧑' : '🤖' }}</text>
          </view>
          <view class="msg-content">
            <view class="msg-bubble" :class="{ 'user-bubble': msg.role === 'user' }">
              <text v-if="msg.role === 'user' || msg.routeDraftData" class="msg-text">{{ msg.content }}</text>
              <rich-text
                v-else
                class="msg-text md-content"
                :nodes="renderMarkdown(msg.content)"
              />
              <view v-if="msg.role === 'assistant' && (store.isStreaming || (store.loading && !msg.content))" class="typing-indicator">
                <view class="typing-dot"></view>
                <view class="typing-dot"></view>
                <view class="typing-dot"></view>
              </view>

              <view v-if="msg.failureInfo?.visible" class="failure-card">
                <view class="failure-header">
                  <text class="failure-title">{{ msg.failureInfo.title }}</text>
                  <text class="failure-summary">{{ msg.failureInfo.summary }}</text>
                </view>
                <text class="failure-reason">{{ msg.failureInfo.reason }}</text>
                <view v-if="msg.failureInfo.failedSteps.length" class="failure-steps">
                  <view
                    v-for="step in msg.failureInfo.failedSteps"
                    :key="step.index"
                    class="failure-step"
                  >
                    <text class="failure-step-index">{{ step.index }}</text>
                    <text class="failure-step-name">{{ step.name || step.type }}</text>
                    <text class="failure-step-type">{{ step.type }}</text>
                  </view>
                </view>
                <view class="failure-footer">
                  <text class="failure-suggestion">{{ msg.failureInfo.suggestion }}</text>
                  <text
                    v-if="msg.routeDraftData?.run_id || msg.runId"
                    class="failure-detail"
                    @tap="showRunDetail(msg.routeDraftData?.run_id || msg.runId || '')"
                  >查看详情</text>
                </view>
              </view>
              
              <view v-if="msg.recommendData" class="recommend-cards">
                <view
                  v-for="(item, idx) in msg.recommendData.items"
                  :key="idx"
                  class="recommend-card"
                  @tap="importRecommendedRoute(msg.recommendData!.artifact_id, item)"
                >
                  <view class="rc-header">
                    <text class="rc-title">{{ item.title }}</text>
                    <text class="rc-city">{{ item.city }}</text>
                  </view>
                  <view class="rc-info">
                    <text class="rc-days">📅 {{ item.days }}天</text>
                    <text class="rc-budget">💰 ¥{{ item.estimated_budget }}</text>
                  </view>
                  <text class="rc-reason">{{ item.reason }}</text>
                  <view class="rc-tags">
                    <text v-for="tag in item.tags" :key="tag" class="rc-tag">{{ tag }}</text>
                  </view>
                  <view class="rc-action">
                    <text class="rc-btn">📍 添加到我的路线</text>
                  </view>
                </view>
              </view>

              <view v-if="msg.routeDraftData" class="draft-card">
                <view class="dc-status" :class="msg.routeDraftStatus || 'pending'">
                  <text>{{ routeDraftStatusText(msg.routeDraftStatus) }}</text>
                </view>
                <view class="dc-header">
                  <text class="dc-title">{{ msg.routeDraftData.route_draft.title }}</text>
                  <text class="dc-budget">💰 预估 ¥{{ msg.routeDraftData.route_draft.estimated_budget }}</text>
                </view>
                <text class="dc-summary">{{ msg.routeDraftData.route_draft.summary }}</text>
                <view class="dc-checkpoints">
                  <view v-for="cp in msg.routeDraftData.route_draft.checkpoints" :key="cp.sequence" class="dc-cp">
                    <text class="dc-cp-num">{{ cp.sequence }}</text>
                    <view class="dc-cp-info">
                      <text class="dc-cp-name">{{ cp.name }}</text>
                      <text class="dc-cp-time">{{ cp.arrive_time }} · 停留{{ cp.stay_duration }}分钟</text>
                    </view>
                  </view>
                </view>
                <view class="dc-action">
                  <text
                    class="dc-btn ghost"
                    @tap="showRunDetail(msg.routeDraftData.run_id)"
                  >运行详情</text>
                  <text
                    v-if="(msg.routeDraftStatus || 'pending') === 'pending'"
                    class="dc-btn secondary"
                    @tap="cancelRouteDraft(msg.routeDraftData.artifact_id)"
                  >取消</text>
                  <text
                    v-if="(msg.routeDraftStatus || 'pending') === 'pending'"
                    class="dc-btn primary"
                    @tap="confirmRouteDraft(msg.routeDraftData.artifact_id)"
                  >确认加入</text>
                  <text
                    v-else
                    class="dc-btn disabled"
                  >{{ routeDraftStatusText(msg.routeDraftStatus) }}</text>
                </view>
              </view>
            </view>
            <text class="msg-time">{{ formatTime(msg.timestamp) }}</text>
          </view>
        </view>

        <view v-if="store.error" id="error-anchor" class="error-message">
          <text class="error-icon">⚠️</text>
          <text class="error-text">{{ store.error }}</text>
        </view>

        <view class="scroll-spacer" />
      </scroll-view>

      <view v-if="runDetailVisible" class="run-panel-mask" @tap="closeRunDetail" />
      <view v-if="runDetailVisible" class="run-panel">
        <view class="run-panel-header">
          <view>
            <text class="run-panel-title">运行详情</text>
            <text class="run-panel-subtitle">{{ runDetail?.intent || 'agent' }} · {{ runDetail?.status || 'loading' }}</text>
          </view>
          <text class="run-panel-close" @tap="closeRunDetail">×</text>
        </view>
        <view v-if="runDetailLoading" class="run-panel-loading">加载中...</view>
        <view v-else-if="runDetail" class="run-panel-body">
          <view class="run-metrics">
            <view class="run-metric">
              <text class="run-metric-value">{{ runDetail.total_tokens }}</text>
              <text class="run-metric-label">tokens</text>
            </view>
            <view class="run-metric">
              <text class="run-metric-value">{{ runDetail.latency_ms }}ms</text>
              <text class="run-metric-label">耗时</text>
            </view>
            <view class="run-metric">
              <text class="run-metric-value">{{ runDetail.steps.length }}</text>
              <text class="run-metric-label">步骤</text>
            </view>
          </view>
          <view class="run-section">
            <text class="run-section-title">Steps</text>
            <view v-for="step in runDetail.steps" :key="step.index" class="run-step">
              <text class="run-step-index">{{ step.index }}</text>
              <view class="run-step-content">
                <text class="run-step-name">{{ step.name }}</text>
                <text class="run-step-meta">{{ step.type }} · {{ step.status }} · {{ step.latency_ms }}ms</text>
              </view>
            </view>
          </view>
          <view class="run-section">
            <text class="run-section-title">Artifacts</text>
            <view v-for="artifact in runDetail.artifacts" :key="artifact.artifact_id" class="run-artifact">
              <text class="run-artifact-type">{{ artifact.type }}</text>
              <text class="run-artifact-status">{{ artifact.status }}</text>
            </view>
          </view>
        </view>
        <view v-else class="run-panel-loading">暂无运行详情</view>
      </view>

      <view class="input-bar">
        <input
          v-model="inputText"
          class="msg-input"
          placeholder="输入你的旅行需求..."
          confirm-type="send"
          :disabled="store.isStreaming"
          @confirm="sendMessage"
        />
        <view
          class="send-btn"
          :class="{ disabled: !inputText.trim() || store.isStreaming }"
          @tap="sendMessage"
        >
          <text class="send-icon">➤</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useAgentStore } from '@/stores/useAgentStore'
import { renderMarkdownToRichText } from '@/utils/markdown'

const store = useAgentStore()
const inputText = ref('')
const scrollToId = ref('')
const showSidebar = ref(false)
const runDetailVisible = ref(false)
const runDetailLoading = ref(false)
const runDetail = ref<RunDetailResponse | null>(null)

onMounted(async () => {
  await store.loadSessions()
  
  if (store.sessions.length > 0 && !store.sessionId) {
    await store.switchSession(store.sessions[0])
  } else if (!store.hasMessages) {
    store.initSession()
  }
})

onShow(async () => {
  await store.loadSessions()
  
  if (store.sessions.length > 0 && !store.sessionId) {
    await store.switchSession(store.sessions[0])
  } else if (!store.hasMessages) {
    store.initSession()
  }
})

watch(
  () => store.messages.length,
  () => {
    setTimeout(() => {
      if (store.error) {
        scrollToId.value = 'error-anchor'
      } else if (store.messages.length > 0) {
        scrollToId.value = 'msg-' + store.messages[store.messages.length - 1].id
      }
    }, 100)
  }
)

watch(
  () => store.error,
  () => {
    setTimeout(() => {
      if (store.error) scrollToId.value = 'error-anchor'
    }, 100)
  }
)

function toggleSidebar() {
  showSidebar.value = !showSidebar.value
}

async function handleNewChat() {
  await store.createNewSession()
  showSidebar.value = false
}

async function handleSwitchSession(session: SessionInfo) {
  await store.switchSession(session)
  showSidebar.value = false
}

async function handleDeleteSession(session: SessionInfo) {
  uni.showModal({
    title: '删除会话',
    content: `确定要删除 "${session.title}" 吗？`,
    success: async (res) => {
      if (res.confirm) {
        await store.deleteSessionById(session.session_id)
      }
    },
  })
}

function goBack() {
  uni.navigateBack()
}

function sendMessage() {
  if (!inputText.value.trim() || store.isStreaming) return
  const text = inputText.value.trim()
  inputText.value = ''
  store.sendMessage(text)
}

function sendQuickQuestion(q: QuickQuestion) {
  inputText.value = q.query
  store.handleQuickAction(q)
}

async function importRecommendedRoute(artifactId: string, item: RecommendItem) {
  uni.showLoading({ title: '导入中...' })
  try {
    await store.commitArtifact(artifactId, {
      commit_type: 'create_route',
      idempotency_key: 'key_' + Date.now(),
    })
    uni.hideLoading()
    uni.showToast({ title: '已添加到我的路线', icon: 'success' })
  } catch (err) {
    uni.hideLoading()
    uni.showToast({ title: '导入失败', icon: 'none' })
  }
}

async function confirmRouteDraft(artifactId: string) {
  uni.showLoading({ title: '加入中...' })
  try {
    await store.approveArtifact(artifactId)
    const result = await store.commitArtifact(artifactId, {
      commit_type: 'create_route',
      idempotency_key: 'key_' + Date.now(),
    })
    store.markRouteDraftStatus(artifactId, 'committed')
    uni.hideLoading()
    uni.showToast({ title: '已加入个人打卡路线', icon: 'success' })
    uni.navigateTo({ url: `/pages/route-detail/index?id=${result.entity_id}` })
  } catch (err) {
    uni.hideLoading()
    uni.showToast({ title: '加入失败', icon: 'none' })
  }
}

async function showRunDetail(runId: string) {
  if (!runId) return
  runDetailVisible.value = true
  runDetailLoading.value = true
  runDetail.value = null
  try {
    runDetail.value = await store.getRunDetail(runId)
  } catch (err) {
    uni.showToast({ title: '运行详情加载失败', icon: 'none' })
  } finally {
    runDetailLoading.value = false
  }
}

function closeRunDetail() {
  runDetailVisible.value = false
}

function cancelRouteDraft(artifactId: string) {
  store.markRouteDraftStatus(artifactId, 'cancelled')
  uni.showToast({ title: '已取消', icon: 'none' })
}

function routeDraftStatusText(status?: 'pending' | 'committed' | 'cancelled') {
  if (status === 'committed') return '已加入'
  if (status === 'cancelled') return '已取消'
  return '待确认加入'
}

function renderMarkdown(content: string) {
  return renderMarkdownToRichText(content)
}

function clearChat() {
  uni.showModal({
    title: '确认清空',
    content: '确定要清空所有聊天记录吗？',
    success: (res) => {
      if (res.confirm) {
        store.clearMessages()
        store.initSession()
      }
    },
  })
}

function formatTime(timestamp: number): string {
  const date = new Date(timestamp)
  const hours = date.getHours().toString().padStart(2, '0')
  const minutes = date.getMinutes().toString().padStart(2, '0')
  return `${hours}:${minutes}`
}

function formatSessionTime(timeStr?: string): string {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  
  if (days === 0) {
    const hours = date.getHours().toString().padStart(2, '0')
    const minutes = date.getMinutes().toString().padStart(2, '0')
    return `${hours}:${minutes}`
  } else if (days === 1) {
    return '昨天'
  } else if (days < 7) {
    return `${days}天前`
  } else {
    const month = (date.getMonth() + 1).toString().padStart(2, '0')
    const day = date.getDate().toString().padStart(2, '0')
    return `${month}-${day}`
  }
}
</script>

<style lang="scss" scoped>
.agent-page {
  min-height: 100vh;
  background: linear-gradient(170deg, #FDF8F4, #F7F4F0, #F3F1F5, #F5F3F8);
  display: flex;
}

.sidebar-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  z-index: 100;
  animation: fade-in 0.2s ease;
}

@keyframes fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

.sidebar {
  position: fixed;
  top: 0;
  left: 0;
  width: 680rpx;
  height: 100vh;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: saturate(170%) blur(24px);
  -webkit-backdrop-filter: saturate(170%) blur(24px);
  z-index: 101;
  transform: translateX(-100%);
  transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  display: flex;
  flex-direction: column;
  box-shadow: 8px 0 32px rgba(0, 0, 0, 0.08);
}

.sidebar.show {
  transform: translateX(0);
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 40rpx 32rpx 24rpx;
  border-bottom: 1px solid rgba(0, 0, 0, 0.04);
}

.sidebar-title {
  font-size: 36rpx;
  font-weight: 700;
  color: #1C1C1E;
}

.sidebar-close {
  font-size: 36rpx;
  color: #8E8E93;
  padding: 8rpx;
}

.sidebar-action {
  padding: 20rpx 24rpx;
}

.new-chat-btn {
  display: flex;
  align-items: center;
  gap: 12rpx;
  padding: 20rpx 24rpx;
  background: linear-gradient(135deg, #A18CD1, #FBC2EB);
  border-radius: 16rpx;
  box-shadow: 0 4px 16px rgba(161, 140, 209, 0.3);
}

.new-chat-icon {
  font-size: 36rpx;
  color: #fff;
  font-weight: 300;
}

.new-chat-text {
  font-size: 28rpx;
  color: #fff;
  font-weight: 600;
}

.sessions-list {
  flex: 1;
  padding: 0 16rpx;
}

.loading-sessions,
.empty-sessions {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80rpx 0;
}

.empty-icon {
  font-size: 64rpx;
  margin-bottom: 16rpx;
}

.loading-text,
.empty-text {
  font-size: 26rpx;
  color: #AEAEB2;
}

.session-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20rpx 20rpx 20rpx 24rpx;
  margin-bottom: 8rpx;
  background: rgba(0, 0, 0, 0.02);
  border-radius: 16rpx;
  transition: all 0.2s ease;
}

.session-item:active {
  background: rgba(161, 140, 209, 0.1);
}

.session-item.active {
  background: linear-gradient(135deg, rgba(161, 140, 209, 0.15), rgba(251, 194, 235, 0.15));
  border: 1px solid rgba(161, 140, 209, 0.3);
}

.session-content {
  flex: 1;
  overflow: hidden;
}

.session-title {
  font-size: 28rpx;
  font-weight: 600;
  color: #1C1C1E;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.session-meta {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-top: 8rpx;
}

.session-time {
  font-size: 22rpx;
  color: #AEAEB2;
}

.session-count {
  font-size: 22rpx;
  color: #C7C7CC;
  background: rgba(0, 0, 0, 0.04);
  padding: 2rpx 10rpx;
  border-radius: 8rpx;
}

.session-delete {
  padding: 8rpx;
  margin-left: 12rpx;
}

.delete-icon {
  font-size: 32rpx;
  color: #C7C7CC;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.navbar {
  position: sticky;
  top: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 88rpx;
  padding: 0 20rpx;
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: saturate(170%) blur(24px);
  -webkit-backdrop-filter: saturate(170%) blur(24px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.4);
}

.nav-left {
  width: 140rpx;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 8rpx;
}

.back-btn {
  width: 64rpx;
  height: 64rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.04);
}

.back-icon {
  font-size: 36rpx;
  color: #1C1C1E;
  font-weight: 300;
}

.menu-btn {
  width: 64rpx;
  height: 64rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.04);
}

.menu-icon {
  font-size: 32rpx;
  color: #1C1C1E;
}

.nav-center {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  max-width: 500rpx;
}

.nav-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1C1C1E;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.nav-right {
  width: 80rpx;
  display: flex;
  align-items: center;
  justify-content: flex-end;
}

.clear-btn {
  font-size: 26rpx;
  color: #8E8E93;
}

.chat-container {
  flex: 1;
  padding: 20rpx 28rpx;
  padding-bottom: 220rpx;
}

.quick-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-bottom: 24rpx;
}

.chip {
  display: flex;
  align-items: center;
  gap: 6rpx;
  padding: 12rpx 20rpx;
  background: rgba(255, 255, 255, 0.6);
  border-radius: 9999rpx;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  transition: all 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.chip:active {
  transform: scale(0.95);
  background: rgba(161, 140, 209, 0.15);
}

.chip-icon {
  font-size: 28rpx;
}

.chip-text {
  font-size: 24rpx;
  color: #636366;
}

.loading-chat {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 80rpx 0;
}

.empty-chat {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 80rpx 40rpx;
}

.empty-chat .empty-icon {
  width: 160rpx;
  height: 160rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 80rpx;
  background: linear-gradient(135deg, rgba(161, 140, 209, 0.2), rgba(251, 194, 235, 0.2));
  border-radius: 50%;
  animation: pulse-glow 2s ease-in-out infinite;
}

@keyframes pulse-glow {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(161, 140, 209, 0.4);
  }
  50% {
    box-shadow: 0 0 0 20rpx rgba(161, 140, 209, 0);
  }
}

.empty-title {
  font-size: 36rpx;
  font-weight: 700;
  color: #1C1C1E;
  margin-top: 24rpx;
}

.empty-desc {
  font-size: 26rpx;
  color: #AEAEB2;
  margin-top: 12rpx;
  text-align: center;
}

.empty-features {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
  margin-top: 32rpx;
}

.empty-features text {
  font-size: 24rpx;
  color: #8E8E93;
}

.message-item {
  display: flex;
  gap: 16rpx;
  margin-bottom: 24rpx;
  animation: fade-in-up 0.3s ease-out;
}

@keyframes fade-in-up {
  from {
    opacity: 0;
    transform: translateY(10rpx);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.message-item.is-user {
  flex-direction: row-reverse;
}

.msg-avatar {
  width: 72rpx;
  height: 72rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  flex-shrink: 0;
}

.message-item:not(.is-user) .msg-avatar {
  background: linear-gradient(135deg, #A18CD1, #FBC2EB);
}

.message-item.is-user .msg-avatar {
  background: linear-gradient(135deg, #89D4CF, #0DA5BF);
}

.msg-content {
  max-width: 75%;
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}

.message-item.is-user .msg-content {
  align-items: flex-end;
}

.msg-bubble {
  padding: 20rpx 24rpx;
  border-radius: 24rpx;
  background: rgba(255, 255, 255, 0.85);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.06);
}

.msg-bubble.user-bubble {
  background: linear-gradient(135deg, #0DA5BF, #26BBD2);
  border-radius: 24rpx 24rpx 6rpx 24rpx;
}

.message-item:not(.is-user) .msg-bubble {
  border-radius: 24rpx 24rpx 24rpx 6rpx;
}

.msg-text {
  font-size: 28rpx;
  line-height: 1.6;
  color: #1C1C1E;
  word-break: break-all;
}

.md-content {
  display: block;
}

.md-content :deep(.md-h1),
.md-content :deep(.md-h2),
.md-content :deep(.md-h3) {
  display: block;
  margin: 0 0 14rpx;
  color: #1C1C1E;
  font-weight: 700;
  line-height: 1.35;
}

.md-content :deep(.md-h1) {
  font-size: 34rpx;
}

.md-content :deep(.md-h2) {
  font-size: 31rpx;
}

.md-content :deep(.md-h3) {
  font-size: 28rpx;
}

.md-content :deep(.md-p) {
  display: block;
  margin: 0 0 14rpx;
  color: #1C1C1E;
  font-size: 28rpx;
  line-height: 1.65;
}

.md-content :deep(.md-p:last-child),
.md-content :deep(.md-ul:last-child),
.md-content :deep(.md-ol:last-child),
.md-content :deep(.md-blockquote:last-child),
.md-content :deep(.md-pre:last-child) {
  margin-bottom: 0;
}

.md-content :deep(.md-ul),
.md-content :deep(.md-ol) {
  display: block;
  margin: 0 0 14rpx 26rpx;
  padding-left: 20rpx;
}

.md-content :deep(.md-li) {
  display: list-item;
  margin-bottom: 8rpx;
  color: #1C1C1E;
  font-size: 28rpx;
  line-height: 1.6;
}

.md-content :deep(.md-strong) {
  font-weight: 700;
  color: #111214;
}

.md-content :deep(.md-em) {
  font-style: italic;
}

.md-content :deep(.md-code) {
  padding: 2rpx 8rpx;
  border-radius: 8rpx;
  color: #0B6F7F;
  background: rgba(13, 165, 191, 0.1);
  font-size: 25rpx;
}

.md-content :deep(.md-pre) {
  display: block;
  margin: 0 0 14rpx;
  padding: 16rpx;
  border-radius: 14rpx;
  background: rgba(28, 28, 30, 0.06);
  white-space: pre-wrap;
  word-break: break-all;
}

.md-content :deep(.md-pre .md-code) {
  padding: 0;
  background: transparent;
  color: #333437;
}

.md-content :deep(.md-blockquote) {
  display: block;
  margin: 0 0 14rpx;
  padding: 10rpx 16rpx;
  border-left: 6rpx solid rgba(13, 165, 191, 0.35);
  color: #636366;
  background: rgba(13, 165, 191, 0.06);
}

.msg-bubble.user-bubble .msg-text {
  color: #FFFFFF;
}

.typing-indicator {
  display: flex;
  gap: 8rpx;
  margin-top: 8rpx;
  padding-left: 4rpx;
}

.typing-dot {
  width: 12rpx;
  height: 12rpx;
  border-radius: 50%;
  background: #AEAEB2;
  animation: typing-bounce 1.4s ease-in-out infinite both;
}

.typing-dot:nth-child(1) { animation-delay: -0.32s; }
.typing-dot:nth-child(2) { animation-delay: -0.16s; }

@keyframes typing-bounce {
  0%, 80%, 100% {
    transform: scale(0);
  }
  40% {
    transform: scale(1);
  }
}

.failure-card {
  margin-top: 16rpx;
  padding: 18rpx;
  border-radius: 18rpx;
  background: rgba(255, 244, 230, 0.9);
  border: 1rpx solid rgba(255, 149, 0, 0.22);
}

.failure-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 10rpx;
}

.failure-title {
  color: #9A4F00;
  font-size: 26rpx;
  font-weight: 700;
}

.failure-summary {
  flex: 1;
  color: #C77700;
  font-size: 22rpx;
  text-align: right;
  word-break: break-all;
}

.failure-reason {
  display: block;
  color: #5C3A12;
  font-size: 23rpx;
  line-height: 1.5;
  word-break: break-all;
}

.failure-steps {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
  margin-top: 14rpx;
}

.failure-step {
  display: flex;
  align-items: center;
  gap: 10rpx;
  padding: 10rpx 12rpx;
  border-radius: 12rpx;
  background: rgba(255, 255, 255, 0.62);
}

.failure-step-index {
  width: 34rpx;
  height: 34rpx;
  border-radius: 50%;
  color: #FFFFFF;
  background: #C77700;
  font-size: 20rpx;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.failure-step-name {
  flex: 1;
  min-width: 0;
  color: #1C1C1E;
  font-size: 23rpx;
  font-weight: 600;
  word-break: break-all;
}

.failure-step-type {
  color: #8E6B37;
  font-size: 21rpx;
}

.failure-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  margin-top: 14rpx;
}

.failure-suggestion {
  flex: 1;
  color: #7A5318;
  font-size: 22rpx;
  line-height: 1.45;
}

.failure-detail {
  flex-shrink: 0;
  padding: 8rpx 14rpx;
  border-radius: 999rpx;
  color: #0B6F7F;
  background: rgba(13, 165, 191, 0.1);
  font-size: 22rpx;
  font-weight: 700;
}

.msg-time {
  font-size: 20rpx;
  color: #C7C7CC;
}

.error-message {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 8rpx;
  padding: 18rpx 20rpx;
  margin: 20rpx 0 24rpx 88rpx;
  background: rgba(255, 59, 48, 0.1);
  border: 1rpx solid rgba(255, 59, 48, 0.12);
  border-radius: 18rpx;
}

.error-icon {
  font-size: 24rpx;
}

.error-text {
  font-size: 24rpx;
  color: #FF3B30;
  line-height: 1.4;
  flex: 1;
}

.scroll-spacer {
  height: 220rpx;
}

.input-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 16rpx 28rpx;
  padding-bottom: calc(16rpx + env(safe-area-inset-bottom));
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: saturate(170%) blur(24px);
  -webkit-backdrop-filter: saturate(170%) blur(24px);
  border-top: 1px solid rgba(0, 0, 0, 0.06);
}

.msg-input {
  flex: 1;
  height: 72rpx;
  padding: 0 24rpx;
  background: rgba(0, 0, 0, 0.04);
  border-radius: 9999rpx;
  font-size: 28rpx;
  color: #1C1C1E;
}

.send-btn {
  width: 72rpx;
  height: 72rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #A18CD1, #FBC2EB);
  border-radius: 50%;
  box-shadow: 0 4px 16px rgba(161, 140, 209, 0.4);
  transition: all 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.send-btn:active {
  transform: scale(0.9);
}

.send-btn.disabled {
  background: #D1D1D6;
  box-shadow: none;
}

.send-icon {
  font-size: 32rpx;
  color: #FFFFFF;
}

.send-btn.disabled .send-icon {
  color: #8E8E93;
}

.recommend-cards {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  margin-top: 16rpx;
}

.recommend-card {
  padding: 20rpx;
  background: rgba(255, 255, 255, 0.9);
  border-radius: 20rpx;
  border: 1px solid rgba(161, 140, 209, 0.2);
  transition: all 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.recommend-card:active {
  transform: scale(0.98);
  background: rgba(161, 140, 209, 0.1);
}

.rc-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
}

.rc-title {
  font-size: 28rpx;
  font-weight: 600;
  color: #1C1C1E;
}

.rc-city {
  font-size: 24rpx;
  color: #8E8E93;
  background: rgba(0, 0, 0, 0.04);
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
}

.rc-info {
  display: flex;
  gap: 20rpx;
  margin-bottom: 8rpx;
}

.rc-days,
.rc-budget {
  font-size: 24rpx;
  color: #636366;
}

.rc-reason {
  font-size: 24rpx;
  color: #8E8E93;
  line-height: 1.5;
  margin-bottom: 12rpx;
}

.rc-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8rpx;
  margin-bottom: 16rpx;
}

.rc-tag {
  font-size: 22rpx;
  color: #A18CD1;
  background: rgba(161, 140, 209, 0.1);
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
}

.rc-action {
  display: flex;
  justify-content: flex-end;
}

.rc-btn {
  font-size: 24rpx;
  color: #0DA5BF;
  font-weight: 600;
}

.draft-card {
  padding: 20rpx;
  background: rgba(255, 255, 255, 0.94);
  border-radius: 18rpx;
  border: 1px solid rgba(13, 165, 191, 0.22);
  margin-top: 16rpx;
  box-shadow: 0 8rpx 24rpx rgba(13, 165, 191, 0.08);
}

.dc-status {
  display: inline-flex;
  align-items: center;
  padding: 6rpx 14rpx;
  border-radius: 999rpx;
  margin-bottom: 14rpx;
}

.dc-status text {
  font-size: 22rpx;
  font-weight: 600;
}

.dc-status.pending {
  background: rgba(255, 159, 10, 0.12);
}

.dc-status.pending text {
  color: #C77700;
}

.dc-status.committed {
  background: rgba(52, 199, 89, 0.12);
}

.dc-status.committed text {
  color: #248A3D;
}

.dc-status.cancelled {
  background: rgba(142, 142, 147, 0.12);
}

.dc-status.cancelled text {
  color: #636366;
}

.dc-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
}

.dc-title {
  font-size: 28rpx;
  font-weight: 600;
  color: #1C1C1E;
}

.dc-budget {
  font-size: 24rpx;
  color: #8E8E93;
}

.dc-summary {
  font-size: 24rpx;
  color: #636366;
  line-height: 1.5;
  margin-bottom: 16rpx;
}

.dc-checkpoints {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  margin-bottom: 20rpx;
}

.dc-cp {
  display: flex;
  gap: 12rpx;
  align-items: flex-start;
}

.dc-cp-num {
  width: 40rpx;
  height: 40rpx;
  border-radius: 50%;
  background: linear-gradient(135deg, #89D4CF, #0DA5BF);
  color: #fff;
  font-size: 22rpx;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.dc-cp-info {
  display: flex;
  flex-direction: column;
  gap: 4rpx;
}

.dc-cp-name {
  font-size: 26rpx;
  color: #1C1C1E;
  font-weight: 500;
}

.dc-cp-time {
  font-size: 22rpx;
  color: #AEAEB2;
}

.dc-action {
  display: flex;
  justify-content: flex-end;
  gap: 16rpx;
}

.dc-btn {
  min-width: 150rpx;
  height: 60rpx;
  padding: 0 22rpx;
  border-radius: 14rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 25rpx;
  font-weight: 600;
}

.dc-btn.primary {
  color: #FFFFFF;
  background: linear-gradient(135deg, #0DA5BF, #26BBD2);
  box-shadow: 0 6rpx 16rpx rgba(13, 165, 191, 0.24);
}

.dc-btn.secondary {
  color: #636366;
  background: rgba(0, 0, 0, 0.05);
}

.dc-btn.ghost {
  min-width: 132rpx;
  color: #0B6F7F;
  background: rgba(13, 165, 191, 0.08);
}

.dc-btn.disabled {
  color: #8E8E93;
  background: rgba(0, 0, 0, 0.04);
}

.run-panel-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 80;
  background: rgba(0, 0, 0, 0.26);
}

.run-panel {
  position: fixed;
  left: 24rpx;
  right: 24rpx;
  bottom: calc(112rpx + env(safe-area-inset-bottom));
  z-index: 81;
  max-height: 68vh;
  padding: 24rpx;
  border-radius: 24rpx;
  background: rgba(255, 255, 255, 0.98);
  box-shadow: 0 20rpx 60rpx rgba(28, 28, 30, 0.18);
}

.run-panel-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 20rpx;
}

.run-panel-title {
  display: block;
  color: #1C1C1E;
  font-size: 32rpx;
  font-weight: 700;
}

.run-panel-subtitle {
  display: block;
  margin-top: 6rpx;
  color: #8E8E93;
  font-size: 23rpx;
}

.run-panel-close {
  width: 56rpx;
  height: 56rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  color: #636366;
  background: rgba(0, 0, 0, 0.05);
  font-size: 36rpx;
}

.run-panel-loading {
  padding: 48rpx 0;
  text-align: center;
  color: #8E8E93;
  font-size: 26rpx;
}

.run-panel-body {
  max-height: calc(68vh - 130rpx);
  overflow: auto;
}

.run-metrics {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12rpx;
  margin-bottom: 24rpx;
}

.run-metric {
  padding: 16rpx;
  border-radius: 16rpx;
  background: rgba(13, 165, 191, 0.08);
}

.run-metric-value {
  display: block;
  color: #0B6F7F;
  font-size: 26rpx;
  font-weight: 700;
}

.run-metric-label {
  display: block;
  margin-top: 4rpx;
  color: #636366;
  font-size: 21rpx;
}

.run-section {
  margin-top: 20rpx;
}

.run-section-title {
  display: block;
  margin-bottom: 12rpx;
  color: #1C1C1E;
  font-size: 25rpx;
  font-weight: 700;
}

.run-step {
  display: flex;
  gap: 12rpx;
  padding: 14rpx 0;
  border-bottom: 1rpx solid rgba(0, 0, 0, 0.06);
}

.run-step-index {
  width: 40rpx;
  height: 40rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  color: #FFFFFF;
  background: #0DA5BF;
  font-size: 22rpx;
  font-weight: 700;
  flex-shrink: 0;
}

.run-step-content {
  flex: 1;
  min-width: 0;
}

.run-step-name {
  display: block;
  color: #1C1C1E;
  font-size: 25rpx;
  font-weight: 600;
  word-break: break-all;
}

.run-step-meta {
  display: block;
  margin-top: 4rpx;
  color: #8E8E93;
  font-size: 21rpx;
}

.run-artifact {
  display: flex;
  justify-content: space-between;
  padding: 14rpx 0;
  border-bottom: 1rpx solid rgba(0, 0, 0, 0.06);
}

.run-artifact-type,
.run-artifact-status {
  color: #636366;
  font-size: 24rpx;
}
</style>
