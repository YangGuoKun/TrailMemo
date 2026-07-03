<template>
  <view class="pd-page">
    <view class="navbar">
      <view class="nav-back" @tap="goBack"><u-icon name="arrow-left" size="40" color="#1C1C1E" /></view>
      <text class="nav-title">帖子详情</text>
    </view>

    <scroll-view scroll-y class="page-body" v-if="post">
      <!-- 封面 -->
      <image v-if="post.images" :src="coverImage" mode="aspectFill" class="pd-cover" />
      <view v-else class="pd-cover-placeholder">🗺️</view>

      <!-- 帖子信息 -->
      <view class="glass-card">
        <text class="pd-title">{{post.title}}</text>
        <view class="pd-author-row">
          <text class="pd-author">{{post.user?.nickname||post.user?.username||'匿名'}}</text>
          <text class="pd-time">{{post.created_at?.split('T')[0]}}</text>
        </view>
        <text class="pd-content">{{post.content}}</text>
      </view>

      <!-- 操作按钮 -->
      <view class="glass-card action-card">
        <text class="act-btn" :class="{liked:liked}" @tap="handleLike">{{liked?'❤️':'🤍'}} {{post.like_count||0}}</text>
        <text class="act-btn" @tap="scrollToComments">💬 {{post.comment_count||0}}</text>
        <text class="act-btn" @tap="handleCopy">📋 复用路线</text>
      </view>

      <!-- 评论区 -->
      <view class="glass-card" id="comment-section">
        <text class="section-title">评论 ({{comments.length}})</text>
        <view v-for="c in comments" :key="c.id" class="cmt-item">
          <view class="cmt-avatar">🧑</view>
          <view class="cmt-body">
            <text class="cmt-name">{{c.user?.nickname||'匿名'}}</text>
            <text class="cmt-text">{{c.content}}</text>
            <text class="cmt-time">{{c.created_at?.split('T')[0]}}</text>
          </view>
        </view>
        <view v-if="comments.length===0" class="empty-cmt"><text>暂无评论，来说点什么吧</text></view>
      </view>

      <!-- 评论输入 -->
      <view class="cmt-input-row">
        <input v-model="cmtText" placeholder="写下你的评论..." @confirm="handleAddComment" />
        <text class="cmt-send" @tap="handleAddComment">发送</text>
      </view>
      <view style="height:40rpx"/>
    </scroll-view>

    <view v-else class="loading-wrap"><u-loading-icon size="40"/><text>加载中...</text></view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useCommunityStore } from '@/stores/useCommunityStore'
import { useRouteStore } from '@/stores/useRouteStore'
import { getFullUrl } from '@/config'

const store = useCommunityStore()
const routeStore = useRouteStore()
const postId = ref(0)
const cmtText = ref('')
const liked = ref(false)

const post = computed(() => store.currentPost)
const comments = computed(() => store.comments)
const coverImage = computed(() => {
  if (!post.value?.images) return ''
  return getFullUrl(post.value.images.split(',')[0])
})

onLoad(async (options: any) => {
  const id = Number(options.id); if (!id) return
  postId.value = id
  try {
    await store.fetchPostDetail(id)
    liked.value = store.likeStatus['post_'+id] || false
  } catch (_) {}
})

async function handleLike() {
  try {
    liked.value = await store.togglePostLike(postId.value)
  } catch (_) {}
}
async function handleCopy() {
  if (!post.value) return
  try { await routeStore.copyRoute(post.value.route_id); uni.showToast({title:'已复制到我的路线',icon:'success'}) } catch (_) {}
}
async function handleAddComment() {
  const t = cmtText.value.trim(); if (!t) return
  try { await store.addComment(postId.value, t); cmtText.value = '' } catch (_) {}
}
function scrollToComments() {
  uni.createSelectorQuery().select('#comment-section').boundingClientRect((rect: any) => {
    if (rect) uni.pageScrollTo({ scrollTop: rect.top + 100 })
  }).exec()
}
function goBack() { uni.navigateBack() }
</script>

<style lang="scss" scoped>
.pd-page{ min-height:100vh; background:linear-gradient(170deg,#FDF8F4,#F7F4F0,#F3F1F5,#F5F3F8) }
.navbar{ position:sticky; top:0; z-index:50; display:flex; align-items:center; justify-content:center; height:88rpx; padding:0 32rpx;
  background:rgba(255,255,255,0.72); backdrop-filter:saturate(170%) blur(24px); -webkit-backdrop-filter:saturate(170%) blur(24px); border-bottom:1px solid rgba(255,255,255,0.4);
  .nav-back{ position:absolute; left:16rpx; width:64rpx; height:64rpx; border-radius:50%; background:rgba(0,0,0,0.04); display:flex; align-items:center; justify-content:center }
  .nav-title{ font-size:34rpx; font-weight:600; color:#1C1C1E } }
.page-body{ padding:16rpx 0 120rpx }
.pd-cover{ width:100%; height:400rpx; object-fit:cover }
.pd-cover-placeholder{ width:100%; height:400rpx; display:flex; align-items:center; justify-content:center; font-size:80rpx; background:linear-gradient(135deg,#FFD4C0,#FFB88C) }
.glass-card{ margin:14rpx 28rpx; padding:28rpx; background:rgba(255,255,255,0.55); border-radius:24rpx; box-shadow:0 8px 32px rgba(0,0,0,0.06),0 2px 8px rgba(0,0,0,0.03) }
.pd-title{ font-size:38rpx; font-weight:700; color:#1C1C1E; display:block; margin-bottom:12rpx }
.pd-author-row{ display:flex; align-items:center; gap:16rpx; margin-bottom:16rpx }
.pd-author{ font-size:26rpx; font-weight:600; color:#0DA5BF }
.pd-time{ font-size:22rpx; color:#AEAEB2 }
.pd-content{ font-size:30rpx; color:#636366; line-height:1.7 }
.action-card{ display:flex; justify-content:space-around; padding:20rpx }
.act-btn{ font-size:28rpx; color:#8E8E93; &.liked{ color:#FF3B30 } }
.section-title{ font-size:30rpx; font-weight:600; color:#1C1C1E; margin-bottom:16rpx; display:block }
.cmt-item{ display:flex; gap:14rpx; margin-bottom:16rpx }
.cmt-avatar{ width:56rpx; height:56rpx; border-radius:50%; background:linear-gradient(135deg,#E5E5EA,#D1D1D6); display:flex; align-items:center; justify-content:center; font-size:24rpx; flex-shrink:0 }
.cmt-body{ flex:1; .cmt-name{ font-size:26rpx; font-weight:600; color:#1C1C1E; display:block }
  .cmt-text{ font-size:28rpx; color:#636366; margin-top:4rpx; line-height:1.5 }
  .cmt-time{ font-size:20rpx; color:#AEAEB2; margin-top:4rpx } }
.empty-cmt{ text-align:center; padding:24rpx; font-size:26rpx; color:#AEAEB2 }
.cmt-input-row{ display:flex; gap:12rpx; padding:16rpx 28rpx; position:fixed; bottom:0; left:0; right:0; background:rgba(255,255,255,0.9); backdrop-filter:blur(20px); border-top:1px solid rgba(0,0,0,0.06);
  input{ flex:1; border:none; background:rgba(0,0,0,0.04); border-radius:9999rpx; padding:14rpx 20rpx; font-size:28rpx; outline:none }
  .cmt-send{ padding:14rpx 24rpx; border-radius:9999rpx; background:#0DA5BF; color:#fff; font-size:26rpx; font-weight:600 } }
.loading-wrap{ display:flex; align-items:center; justify-content:center; gap:12rpx; height:100vh; font-size:28rpx; color:#AEAEB2 }
</style>
