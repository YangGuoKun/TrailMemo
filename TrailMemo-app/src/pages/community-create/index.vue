<template>
  <view class="community-create-page">
    <view class="navbar glass-surface">
      <view class="nav-back" @tap="goBack"><text class="back-icon">←</text></view>
      <text class="nav-title">发布帖子</text>
      <text class="nav-action" @tap="handlePublish">发布</text>
    </view>

    <scroll-view scroll-y class="page-body">
      <view class="section">
        <text class="section-label">标题</text>
        <input v-model="form.title" placeholder="输入帖子标题" class="input-field" />
      </view>

      <view class="section">
        <text class="section-label">内容</text>
        <textarea
          v-model="form.content"
          placeholder="分享你的旅行故事..."
          class="textarea-field"
          :auto-height="true"
        />
      </view>

      <view class="section" v-if="form.route_id">
        <view class="route-tag">
          <text class="tag-icon">🗺️</text>
          <text class="tag-text">关联路线 ID: {{ form.route_id }}</text>
        </view>
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useCommunityStore } from '@/stores/useCommunityStore'

const communityStore = useCommunityStore()

const form = ref({
  title: '',
  content: '',
  route_id: 0,
})

onLoad((options: any) => {
  if (options.routeId) {
    form.value.route_id = Number(options.routeId)
  }
  if (options.content) {
    form.value.content = decodeURIComponent(options.content)
  }
})

async function handlePublish() {
  if (!form.value.title.trim()) {
    uni.showToast({ title: '请输入标题', icon: 'none' })
    return
  }
  if (!form.value.content.trim()) {
    uni.showToast({ title: '请输入内容', icon: 'none' })
    return
  }

  uni.showLoading({ title: '发布中...' })
  try {
    await communityStore.createPost({
      route_id: form.value.route_id,
      title: form.value.title,
      content: form.value.content,
      images: '',
    })
    uni.hideLoading()
    uni.showToast({ title: '发布成功', icon: 'success' })
    setTimeout(() => {
      uni.navigateBack()
    }, 1500)
  } catch (err) {
    uni.hideLoading()
    uni.showToast({ title: '发布失败', icon: 'none' })
  }
}

function goBack() {
  uni.navigateBack()
}
</script>

<style lang="scss" scoped>
.community-create-page { min-height: 100vh; background: linear-gradient(170deg, #FDF8F4, #F7F4F0, #F3F1F5, #F5F3F8); }

.navbar {
  position: sticky; top: 0; z-index: 50;
  display: flex; align-items: center; justify-content: center;
  height: 88rpx; padding: 0 32rpx;
  background: rgba(255,255,255,0.72);
  backdrop-filter: saturate(170%) blur(24px);
  -webkit-backdrop-filter: saturate(170%) blur(24px);
  border-bottom: 1px solid rgba(255,255,255,0.4);
  .nav-back { position: absolute; left: 16rpx; width: 64rpx; height: 64rpx; border-radius: 50%; background: rgba(0,0,0,0.04); display: flex; align-items: center; justify-content: center; }
  .back-icon { font-size: 32rpx; color: #1C1C1E; }
  .nav-title { font-size: 34rpx; font-weight: 600; color: #1C1C1E; }
  .nav-action { position: absolute; right: 32rpx; font-size: 28rpx; color: #0DA5BF; font-weight: 600; }
}

.page-body { padding: 24rpx 32rpx; }

.section { margin-bottom: 32rpx; }

.section-label { font-size: 28rpx; font-weight: 600; color: #1C1C1E; display: block; margin-bottom: 12rpx; }

.input-field {
  width: 100%;
  height: 88rpx;
  padding: 0 24rpx;
  background: rgba(255,255,255,0.7);
  border-radius: 16rpx;
  border: 1px solid rgba(0,0,0,0.08);
  font-size: 30rpx;
}

.textarea-field {
  width: 100%;
  min-height: 300rpx;
  padding: 20rpx 24rpx;
  background: rgba(255,255,255,0.7);
  border-radius: 16rpx;
  border: 1px solid rgba(0,0,0,0.08);
  font-size: 28rpx;
  line-height: 1.6;
}

.route-tag {
  display: flex;
  align-items: center;
  gap: 8rpx;
  padding: 16rpx 24rpx;
  background: rgba(137, 212, 207, 0.15);
  border-radius: 12rpx;
}

.tag-icon { font-size: 28rpx; }

.tag-text { font-size: 26rpx; color: #0DA5BF; }
</style>