<template>
  <view class="my-posts-page">
    <view class="navbar glass-surface">
      <view class="nav-back" @tap="goBack"><text class="back-icon">←</text></view>
      <text class="nav-title">我的游记</text>
    </view>

    <scroll-view scroll-y class="page-body">
      <!-- 统计 -->
      <view class="stat-banner">
        <text class="stat-num">{{ total }}</text>
        <text class="stat-label">篇游记</text>
      </view>

      <!-- 空状态 -->
      <view v-if="!loading && posts.length === 0" class="empty-state">
        <text class="empty-icon">📝</text>
        <text class="empty-title">还没有发布游记</text>
        <text class="empty-desc">完成一条路线后，分享你的旅行故事吧</text>
        <button class="empty-btn" @tap="goCommunityCreate">发布游记</button>
      </view>

      <!-- 帖子列表 -->
      <view v-for="p in posts" :key="p.id" class="post-card" @tap="goPostDetail(p.id)">
        <view class="post-cover">
          <image v-if="getFirstImage(p)" :src="getFirstImage(p)" mode="aspectFill" class="cover-img" />
          <view v-else class="cover-placeholder" :style="{background: coverGradients[p.id % 3]}"><text>🗺️</text></view>
        </view>
        <view class="post-body">
          <text class="post-title">{{ p.title }}</text>
          <text class="post-time">{{ formatTime(p.created_at) }}</text>
          <text class="post-content">{{ p.content }}</text>
          <view class="post-actions">
            <text class="action-btn">❤️ {{ p.like_count || 0 }}</text>
            <text class="action-btn">💬 {{ p.comment_count || 0 }}</text>
            <text class="action-btn delete" @tap.stop="handleDelete(p.id)">🗑️ 删除</text>
          </view>
        </view>
      </view>

      <view v-if="loading" class="loading-row"><text>加载中...</text></view>
      <view style="height: 40rpx" />
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as communityApi from '@/api/community'
import { useUserStore } from '@/stores/useUserStore'
import { getFullUrl } from '@/config'

const userStore = useUserStore()
const posts = ref<Post[]>([])
const total = ref(0)
const loading = ref(false)

const coverGradients = ['linear-gradient(135deg,#FFD4C0,#FFB88C)', 'linear-gradient(135deg,#89D4CF,#4DCADC)', 'linear-gradient(135deg,#C4A6E8,#FBC2EB)']

onMounted(async () => {
  await loadPosts()
})

async function loadPosts() {
  loading.value = true
  try {
    const userId = userStore.profile?.id
    if (!userId) return
    const result = await communityApi.getPostList({ user_id: userId, page: 1, size: 20 })
    posts.value = result.list || []
    total.value = result.total || 0
  } catch (_) {
  } finally {
    loading.value = false
  }
}

function getFirstImage(p: Post): string {
  if (!p.images) return ''
  const imgs = p.images.split(',').filter(Boolean)
  return imgs[0] ? getFullUrl(imgs[0]) : ''
}

function formatTime(t: string): string {
  if (!t) return ''
  const d = new Date(t)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return Math.floor(diff / 60000) + '分钟前'
  if (diff < 86400000) return Math.floor(diff / 3600000) + '小时前'
  return d.toLocaleDateString('zh-CN')
}

async function handleDelete(postId: number) {
  uni.showModal({
    title: '确认删除',
    content: '确定要删除这篇游记吗？',
    success: async (res) => {
      if (res.confirm) {
        try {
          await communityApi.deletePost(postId)
          posts.value = posts.value.filter((p) => p.id !== postId)
          total.value -= 1
          uni.showToast({ title: '删除成功', icon: 'success' })
        } catch (_) {
          uni.showToast({ title: '删除失败', icon: 'none' })
        }
      }
    },
  })
}

function goBack() {
  uni.navigateBack()
}

function goPostDetail(id: number) {
  uni.navigateTo({ url: `/pages/post-detail/index?id=${id}` })
}

function goCommunityCreate() {
  uni.navigateTo({ url: '/pages/community-create/index' })
}
</script>

<style lang="scss" scoped>
.my-posts-page {
  min-height: 100vh;
  background: linear-gradient(170deg, #fdf8f4, #f7f4f0, #f3f1f5, #f5f3f8);
}

.navbar {
  position: sticky;
  top: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 88rpx;
  padding: 0 32rpx;
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: saturate(170%) blur(24px);
  -webkit-backdrop-filter: saturate(170%) blur(24px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.4);

  .nav-back {
    position: absolute;
    left: 16rpx;
    width: 64rpx;
    height: 64rpx;
    border-radius: 50%;
    background: rgba(0, 0, 0, 0.04);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .back-icon {
    font-size: 32rpx;
    color: #1c1c1e;
  }

  .nav-title {
    font-size: 34rpx;
    font-weight: 600;
    color: #1c1c1e;
  }
}

.page-body {
  padding: 24rpx 32rpx;
}

.stat-banner {
  text-align: center;
  padding: 32rpx 0;
  background: rgba(255, 255, 255, 0.6);
  border-radius: 20rpx;
  margin-bottom: 24rpx;

  .stat-num {
    font-size: 48rpx;
    font-weight: 700;
    color: #1c1c1e;
  }

  .stat-label {
    font-size: 26rpx;
    color: #8e8e93;
    margin-left: 8rpx;
  }
}

.empty-state {
  text-align: center;
  padding: 80rpx 32rpx;

  .empty-icon {
    font-size: 80rpx;
    display: block;
    margin-bottom: 16rpx;
  }

  .empty-title {
    font-size: 32rpx;
    font-weight: 600;
    color: #1c1c1e;
    display: block;
  }

  .empty-desc {
    font-size: 26rpx;
    color: #8e8e93;
    margin-top: 8rpx;
    display: block;
  }

  .empty-btn {
    margin-top: 32rpx;
    padding: 16rpx 48rpx;
    background: linear-gradient(135deg, #a18cd1, #fbc2eb);
    border-radius: 9999rpx;
    color: #fff;
    font-size: 28rpx;
    font-weight: 600;
    border: none;
  }
}

.post-card {
  background: rgba(255, 255, 255, 0.6);
  border-radius: 20rpx;
  margin-bottom: 20rpx;
  overflow: hidden;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.04);
}

.post-cover {
  height: 200rpx;

  .cover-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .cover-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;

    text {
      font-size: 56rpx;
    }
  }
}

.post-body {
  padding: 20rpx 24rpx;

  .post-title {
    font-size: 30rpx;
    font-weight: 600;
    color: #1c1c1e;
    display: block;
  }

  .post-time {
    font-size: 22rpx;
    color: #8e8e93;
    margin-top: 4rpx;
    display: block;
  }

  .post-content {
    font-size: 26rpx;
    color: #636366;
    margin-top: 12rpx;
    display: block;
    line-height: 1.5;
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
  }
}

.post-actions {
  display: flex;
  gap: 16rpx;
  margin-top: 16rpx;

  .action-btn {
    font-size: 24rpx;
    color: #8e8e93;
    padding: 8rpx 16rpx;
    background: rgba(0, 0, 0, 0.04);
    border-radius: 9999rpx;
  }

  .delete {
    color: #ff3b30;
    background: rgba(255, 59, 48, 0.1);
  }
}

.loading-row {
  text-align: center;
  padding: 32rpx;

  text {
    font-size: 26rpx;
    color: #8e8e93;
  }
}
</style>