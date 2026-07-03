<template>
  <view class="routes-page">
    <!-- 导航栏（毛玻璃） -->
    <view class="navbar glass-surface">
      <text class="navbar-title">我的路线</text>
      <text class="navbar-action" @tap="goCreateRoute">+ 创建</text>
    </view>

    <view class="page-body">
      <!-- 筛选标签 -->
      <scroll-view scroll-x class="filter-bar" :show-scrollbar="false">
        <view
          v-for="tab in filterTabs"
          :key="tab.key"
          class="filter-tab"
          :class="{ active: activeFilter === tab.key }"
          @tap="switchFilter(tab.key)"
        >{{ tab.label }}</view>
      </scroll-view>

      <!-- 内联创建卡片 -->
      <view class="create-inline-card" @tap="goCreateRoute">
        <text class="cic-icon">＋</text>
        <text class="cic-text">创建新路线</text>
      </view>

      <!-- 空状态 -->
      <view v-if="!routeStore.isListLoading && displayRoutes.length === 0" class="empty-state">
        <text class="empty-icon">🗺️</text>
        <text class="empty-title">还没有路线</text>
        <text class="empty-desc">开始规划你的第一条旅行路线吧</text>
        <button class="create-btn" @tap="goCreateRoute">创建路线</button>
      </view>

      <!-- 路线列表 -->
      <view v-for="route in displayRoutes" :key="route.id" class="route-card" @tap="goRouteDetail(route.id)">
        <!-- 封面 -->
        <view class="card-cover" :style="{ background: getCoverBg(route.id) }">
          <image v-if="route.cover_image" :src="route.cover_image" mode="aspectFill" class="cover-img" />
          <view class="status-badge" :style="{ background: getStatusColor(route.publish_status) }">
            {{ getStatusLabel(route.publish_status) }}
          </view>
          <view class="cover-label">
            <text class="rtitle">{{ route.title }}</text>
            <text class="rcities">{{ route.start_city }} → {{ route.end_city }}</text>
          </view>
        </view>
        <view class="card-body">
          <text class="route-desc text-ellipsis-2" v-if="route.description">{{ route.description }}</text>
          <view class="route-stats">
            <text v-if="route.total_distance">🛣️ {{ route.total_distance }}km</text>
            <text v-if="route.estimated_hours">🕐 {{ route.estimated_hours }}h</text>
            <text>👁 {{ route.view_count || 0 }}</text>
          </view>
        </view>
      </view>

      <!-- 加载更多 -->
      <view v-if="routeStore.isListLoading" class="loading-row">
        <u-loading-icon size="32" />
        <text>加载中...</text>
      </view>
      <view v-else-if="routeStore.routes.length >= routeStore.total && routeStore.routes.length > 0" class="no-more">
        <text>— 没有更多了 —</text>
      </view>
    </view>

    <TmTabbar current="routes" @change="onTabChange" />

    <view class="agent-fab" @tap="goAgent">
      <text class="fab-icon">✨</text>
      <view class="fab-glow"></view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useRouteStore } from '@/stores/useRouteStore'
import TmTabbar from '@/components/common/TmTabbar.vue'

const routeStore = useRouteStore()
const activeFilter = ref('all')

const filterTabs = [
  { key: 'all', label: '全部' },
  { key: 'in_progress', label: '进行中' },
  { key: 'completed', label: '已完成' },
  { key: 'draft', label: '草稿' },
]

const displayRoutes = computed(() => {
  if (activeFilter.value === 'all') return routeStore.routes
  const statusMap: Record<string, number> = { in_progress: 1, completed: 2, draft: 0 }
  return routeStore.routes.filter(r => r.publish_status === statusMap[activeFilter.value])
})

onShow(() => { routeStore.fetchRoutes(1) })

function switchFilter(key: string) { activeFilter.value = key }
function goRouteDetail(id: number) { uni.navigateTo({ url: `/pages/route-detail/index?id=${id}` }) }
function goCreateRoute() { uni.navigateTo({ url: '/pages/route-create/index' }) }
function onTabChange(key: string) { uni.switchTab({ url: `/pages/${key === 'discover' ? 'index' : key}/index` }) }
function goAgent() { uni.navigateTo({ url: '/pages/agent/index' }) }

const covers = [
  'linear-gradient(135deg,#FFD4C0,#FFB88C,#FECF9E)',
  'linear-gradient(135deg,#89D4CF,#4DCADC,#0DA5BF)',
  'linear-gradient(135deg,#C4A6E8,#A18CD1,#FBC2EB)',
]
function getCoverBg(id: number) { return covers[id % 3] }
function getStatusLabel(s: number) { const m: Record<number,string> = {0:'草稿',1:'进行中',2:'已完成'}; return m[s]||'草稿' }
function getStatusColor(s: number) { const m: Record<number,string> = {0:'#8E8E93',1:'#0DA5BF',2:'#34C759'}; return m[s]||'#8E8E93' }
</script>

<style lang="scss" scoped>
.routes-page { min-height: 100vh; background: linear-gradient(170deg, #FDF8F4, #F7F4F0, #F3F1F5, #F5F3F8); }

.navbar {
  position: sticky; top: 0; z-index: 50;
  display: flex; align-items: center; justify-content: center;
  height: 88rpx; padding: 0 32rpx;
  background: rgba(255,255,255,0.72);
  backdrop-filter: saturate(170%) blur(24px);
  -webkit-backdrop-filter: saturate(170%) blur(24px);
  border-bottom: 1px solid rgba(255,255,255,0.4);
  .navbar-title { font-size: 34rpx; font-weight: 600; color: #1C1C1E; }
  .navbar-action { position: absolute; right: 32rpx; font-size: 28rpx; font-weight: 600; color: #0DA5BF; }
}

.page-body { padding: 16rpx 28rpx 120rpx; }

.filter-bar { white-space: nowrap; margin-bottom: 16rpx; display: flex; gap: 12rpx;
  .filter-tab { display: inline-flex; align-items: center; height: 56rpx; padding: 0 28rpx;
    border-radius: 9999rpx; font-size: 26rpx; color: #636366; background: rgba(255,255,255,0.6);
    transition: all 0.2s;
    &.active { color: #fff; background: #0DA5BF; font-weight: 500; }
  }
}

// 内联创建卡片
.create-inline-card { margin: 8rpx 0 20rpx; padding: 32rpx; text-align: center;
  border: 2px dashed rgba(13,165,191,0.3); border-radius: 24rpx;
  background: rgba(13,165,191,0.04);
  &:active { background: rgba(13,165,191,0.1); transform: scale(0.98); }
  .cic-icon { font-size: 56rpx; color: #0DA5BF; }
  .cic-text { font-size: 28rpx; font-weight: 600; color: #0DA5BF; margin-top: 4rpx; }
}

.route-card { margin-bottom: 20rpx; border-radius: 24rpx; overflow: hidden;
  background: rgba(255,255,255,0.55);
  box-shadow: 0 8px 32px rgba(0,0,0,0.06), 0 2px 8px rgba(0,0,0,0.03);
  &:active { transform: scale(0.98); }
  .card-cover { height: 260rpx; position: relative; display: flex; align-items: flex-end; }
  .cover-img { position: absolute; inset: 0; width: 100%; height: 100%; object-fit: cover; }
  .status-badge { position: absolute; top: 16rpx; right: 16rpx; padding: 4rpx 20rpx;
    border-radius: 9999rpx; font-size: 20rpx; font-weight: 600; color: #fff; }
  .cover-label { position: relative; padding: 20rpx 24rpx; color: #fff; width: 100%;
    background: linear-gradient(transparent, rgba(0,0,0,0.35)); }
  .rtitle { font-size: 34rpx; font-weight: 700; display: block; text-shadow: 0 1px 3px rgba(0,0,0,0.3); }
  .rcities { font-size: 24rpx; opacity: 0.85; margin-top: 4rpx; }
  .card-body { padding: 20rpx 24rpx; }
  .route-desc { font-size: 24rpx; color: #636366; margin-bottom: 12rpx; -webkit-line-clamp: 2; }
  .route-stats { display: flex; gap: 28rpx; font-size: 22rpx; color: #8E8E93; }
}

.empty-state { text-align: center; padding: 80rpx 40rpx;
  .empty-icon { font-size: 100rpx; opacity: 0.35; }
  .empty-title { font-size: 32rpx; font-weight: 600; color: #636366; margin: 16rpx 0 8rpx; }
  .empty-desc { font-size: 26rpx; color: #AEAEB2; margin-bottom: 32rpx; }
  .create-btn { padding: 16rpx 48rpx; background: #0DA5BF; color: #fff;
    border-radius: 9999rpx; font-size: 28rpx; font-weight: 600; }
}

.loading-row, .no-more { text-align: center; padding: 32rpx; font-size: 24rpx; color: #AEAEB2; }

.agent-fab{ position:fixed; bottom:160rpx; right:32rpx; z-index:200;
  width:96rpx; height:96rpx; border-radius:50%;
  background:linear-gradient(135deg,#A18CD1,#FBC2EB);
  display:flex; align-items:center; justify-content:center;
  box-shadow:0 8px 28px rgba(161,140,209,0.4); animation:fab-glow 2s ease-in-out infinite;
  &:active{ transform:scale(0.9) }
  .fab-icon{ font-size:44rpx }
  .fab-glow{ position:absolute; inset:0; border-radius:50%;
    background:linear-gradient(135deg,#A18CD1,#FBC2EB); opacity:0.4;
    animation:fab-pulse 2s ease-in-out infinite; } }
@keyframes fab-glow{ 0%,100%{box-shadow:0 8px 28px rgba(161,140,209,0.4)} 50%{box-shadow:0 8px 40px rgba(161,140,209,0.7)} }
@keyframes fab-pulse{ 0%,100%{transform:scale(1); opacity:0.4} 50%{transform:scale(1.3); opacity:0} }
</style>
