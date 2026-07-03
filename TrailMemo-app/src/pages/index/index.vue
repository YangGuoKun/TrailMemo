<template>
  <view class="discover-page">
    <view class="navbar">
      <text class="nav-title">发现</text>
    </view>

    <scroll-view scroll-y class="page-body">
      <!-- 搜索 -->
      <view class="search-bar"><text>🔍</text><input v-model="searchQuery" placeholder="搜索路线、目的地..." @input="onSearch" /></view>

      <!-- 横幅 -->
      <view class="hero-banner"><text class="hero-title">探索新的旅程</text><text class="hero-desc">发现他人分享的精彩路线</text></view>

      <!-- 空状态 -->
      <view v-if="!store.isListLoading && feedPosts.length === 0" class="empty-state">
        <text class="empty-icon">📭</text>
        <text class="empty-title">还没有人分享</text>
        <text class="empty-desc">完成一条路线后，来社区分享你的旅行故事吧</text>
      </view>

      <!-- Feed 列表 -->
      <view v-for="p in feedPosts" :key="p.id" class="feed-card" @tap="goPostDetail(p.id)">
        <view class="feed-cover">
          <image v-if="p.images" :src="getFirstImage(p)" mode="aspectFill" class="cover-img" />
          <view v-else class="cover-placeholder" :style="{background:coverGradients[p.id%3]}"><text>🗺️</text></view>
          <view class="cover-info"><text class="fr-title">{{p.title}}</text></view>
        </view>
        <view class="feed-body">
          <view class="feed-author">
            <view class="fa-avatar">🧑</view>
            <text class="fa-name">{{p.user?.nickname||p.user?.username||'匿名'}}</text>
            <text class="fa-time">{{formatTime(p.created_at)}}</text>
          </view>
          <text class="feed-caption">{{p.content}}</text>
          <view class="feed-actions">
            <text class="fa-btn" :class="{liked:store.likeStatus['post_'+p.id]}" @tap.stop="handleLike(p)">{{store.likeStatus['post_'+p.id]?'❤️':'🤍'}} {{p.like_count||0}}</text>
            <text class="fa-btn" @tap.stop="goPostDetail(p.id)">💬 {{p.comment_count||0}}</text>
            <text class="fa-btn" @tap.stop="handleCopyRoute(p)">📋 复用</text>
          </view>
        </view>
      </view>

      <view v-if="store.isListLoading" class="loading-row"><text>加载中...</text></view>
      <view style="height:40rpx"/>
    </scroll-view>

    <TmTabbar current="discover" @change="onTabChange" />

    <view class="agent-fab" @tap="goAgent">
      <text class="fab-icon">✨</text>
      <view class="fab-glow"></view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useCommunityStore } from '@/stores/useCommunityStore'
import { useRouteStore } from '@/stores/useRouteStore'
import { getFullUrl } from '@/config'
import TmTabbar from '@/components/common/TmTabbar.vue'

const store = useCommunityStore()
const routeStore = useRouteStore()
const searchQuery = ref('')
const coverGradients = ['linear-gradient(135deg,#FFD4C0,#FFB88C)','linear-gradient(135deg,#89D4CF,#4DCADC)','linear-gradient(135deg,#C4A6E8,#FBC2EB)']

const feedPosts = computed(() => {
  if (!searchQuery.value) return store.posts
  const q = searchQuery.value.toLowerCase()
  return store.posts.filter(p => p.title.toLowerCase().includes(q) || p.content.toLowerCase().includes(q))
})

onShow(() => { store.fetchPosts(1) })

function onSearch() {}
function getFirstImage(p: Post): string {
  const imgs = store.getPostImages(p)
  return imgs[0] ? getFullUrl(imgs[0]) : ''
}
function formatTime(t: string): string {
  if (!t) return ''; return t.split('T')[0] || t.slice(0, 10)
}
function goPostDetail(id: number) { uni.navigateTo({ url: `/pages/post-detail/index?id=${id}` }) }

async function handleLike(p: Post) {
  try { await store.togglePostLike(p.id) } catch (_) {}
}
async function handleCopyRoute(p: Post) {
  try { await routeStore.copyRoute(p.route_id) } catch (_) {}
}
function onTabChange(key: string) { uni.switchTab({ url: `/pages/${key==='discover'?'index':key}/index` }) }
function goAgent() { uni.navigateTo({ url: '/pages/agent/index' }) }
</script>

<style lang="scss" scoped>
.discover-page{ min-height:100vh; background:linear-gradient(170deg,#FDF8F4,#F7F4F0,#F3F1F5,#F5F3F8) }
.navbar{ position:sticky; top:0; z-index:50; display:flex; align-items:center; justify-content:center; height:88rpx;
  background:rgba(255,255,255,0.72); backdrop-filter:saturate(170%) blur(24px); -webkit-backdrop-filter:saturate(170%) blur(24px); border-bottom:1px solid rgba(255,255,255,0.4);
  .nav-title{ font-size:34rpx; font-weight:600; color:#1C1C1E } }
.page-body{ padding:16rpx 28rpx 120rpx }
.search-bar{ display:flex; align-items:center; gap:8rpx; padding:16rpx 20rpx; margin-bottom:16rpx;
  background:rgba(255,255,255,0.5); border-radius:9999rpx; box-shadow:0 8px 32px rgba(0,0,0,0.04);
  input{ flex:1; border:none; background:transparent; font-size:28rpx; color:#1C1C1E; outline:none } }
.hero-banner{ height:160rpx; border-radius:24rpx; overflow:hidden; margin-bottom:20rpx;
  background:linear-gradient(135deg,#FFB88C,#FF9A76); display:flex; flex-direction:column; justify-content:center; padding:28rpx;
  .hero-title{ font-size:36rpx; font-weight:700; color:#fff }
  .hero-desc{ font-size:24rpx; color:rgba(255,255,255,0.85); margin-top:6rpx } }
.empty-state{ text-align:center; padding:80rpx 40rpx;
  .empty-icon{ font-size:80rpx; opacity:0.35 }
  .empty-title{ font-size:32rpx; font-weight:600; color:#636366; margin:16rpx 0 8rpx }
  .empty-desc{ font-size:26rpx; color:#AEAEB2; line-height:1.5 } }
.feed-card{ margin-bottom:20rpx; border-radius:24rpx; overflow:hidden;
  background:rgba(255,255,255,0.55); box-shadow:0 8px 32px rgba(0,0,0,0.06),0 2px 8px rgba(0,0,0,0.03) }
.feed-cover{ position:relative; height:300rpx; background:#F0F0F2 }
.cover-img{ width:100%; height:100%; object-fit:cover }
.cover-placeholder{ width:100%; height:100%; display:flex; align-items:center; justify-content:center; font-size:56rpx }
.cover-info{ position:absolute; bottom:0; left:0; right:0; padding:40rpx 20rpx 16rpx; background:linear-gradient(transparent,rgba(0,0,0,0.5)) }
.fr-title{ font-size:34rpx; font-weight:700; color:#fff; text-shadow:0 1px 3px rgba(0,0,0,0.3) }
.feed-body{ padding:20rpx }
.feed-author{ display:flex; align-items:center; gap:10rpx; margin-bottom:10rpx }
.fa-avatar{ width:48rpx; height:48rpx; border-radius:50%; background:linear-gradient(135deg,#E5E5EA,#D1D1D6); display:flex; align-items:center; justify-content:center; font-size:24rpx }
.fa-name{ font-size:26rpx; font-weight:600; color:#1C1C1E }
.fa-time{ margin-left:auto; font-size:22rpx; color:#AEAEB2 }
.feed-caption{ font-size:28rpx; color:#636366; line-height:1.5; display:-webkit-box; -webkit-box-orient:vertical; -webkit-line-clamp:3; overflow:hidden }
.feed-actions{ display:flex; gap:28rpx; padding-top:14rpx; margin-top:14rpx; border-top:1rpx solid rgba(0,0,0,0.04) }
.fa-btn{ font-size:24rpx; color:#8E8E93; &.liked{ color:#FF3B30 } }
.loading-row{ text-align:center; padding:32rpx; font-size:24rpx; color:#AEAEB2 }

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
