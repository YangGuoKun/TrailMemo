<template>
  <view class="profile-page">
    <view class="navbar glass-surface">
      <text class="nav-title">我的</text>
      <text class="nav-action" @tap="goSettings">设置</text>
    </view>

    <scroll-view scroll-y class="page-body">
      <!-- 头像 + 昵称 + ID（原型居中布局） -->
      <view class="profile-header">
        <view class="ph-avatar" @tap="goEditProfile">
          <image v-if="userStore.profile?.avatar" :src="userStore.profile.avatar" mode="aspectFill" class="avatar-img" />
          <text v-else class="avatar-placeholder">🧑</text>
        </view>
        <text class="ph-name">{{ userStore.profile?.nickname || userStore.profile?.username || '未登录' }}</text>
        <text class="ph-id">@{{ userStore.profile?.username || '...' }}</text>
      </view>

      <!-- 统计（原型：纯数字 + 标签，不用卡片包裹） -->
      <view class="stat-row">
        <view class="stat-item" @tap="goRoutes">
          <text class="stat-num">{{ routeCount }}</text>
          <text class="stat-label">我的路线</text>
        </view>
        <view class="stat-item" @tap="goMyCheckins">
          <text class="stat-num">{{ checkinCount }}</text>
          <text class="stat-label">打卡记录</text>
        </view>
        <view class="stat-item">
          <text class="stat-num">0</text>
          <text class="stat-label">收藏</text>
        </view>
      </view>

      <!-- 菜单（外层阴影 + 内层圆角裁剪，避免阴影被裁） -->
      <view class="menu-group">
        <view class="menu-inner">
          <view class="menu-item" @tap="goRoutes">
            <view class="m-left"><text class="m-icon">🗺️</text><text class="m-text">我的路线</text></view>
            <text class="m-right">→</text>
          </view>
          <view class="menu-item" @tap="goMyCheckins">
            <view class="m-left"><text class="m-icon">✅</text><text class="m-text">我的打卡</text></view>
            <text class="m-right">→</text>
          </view>
          <view class="menu-item" @tap="goMyPosts">
            <view class="m-left"><text class="m-icon">📝</text><text class="m-text">我的游记</text></view>
            <text class="m-right">→</text>
          </view>
        </view>
      </view>
      <view class="menu-group">
        <view class="menu-inner">
          <view class="menu-item" @tap="goEditProfile">
            <view class="m-left"><text class="m-icon">✏️</text><text class="m-text">编辑资料</text></view>
            <text class="m-right">→</text>
          </view>
          <view class="menu-item" @tap="goChangePassword">
            <view class="m-left"><text class="m-icon">🔒</text><text class="m-text">修改密码</text></view>
            <text class="m-right">→</text>
          </view>
        </view>
      </view>
      <view class="menu-group">
        <view class="menu-inner">
          <view class="menu-item ai-menu" @tap="goPreferences">
            <view class="m-left"><text class="m-icon">🤖</text><text class="m-text">AI 偏好设置</text></view>
            <text class="m-right">→</text>
          </view>
        </view>
      </view>

      <!-- 退出登录（原型：淡红底色 + 红色文字，full-width） -->
      <button class="logout-btn" @tap="handleLogout">退出登录</button>

      <view style="height: 40rpx" />
    </scroll-view>

    <TmTabbar current="profile" @change="onTabChange" />

    <view class="agent-fab" @tap="goAgent">
      <text class="fab-icon">✨</text>
      <view class="fab-glow"></view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import TmTabbar from '@/components/common/TmTabbar.vue'
import { useUserStore } from '@/stores/useUserStore'
import { useAuthStore } from '@/stores/useAuthStore'
import { useRouteStore } from '@/stores/useRouteStore'
import { useCheckinStore } from '@/stores/useCheckinStore'

const userStore = useUserStore()
const authStore = useAuthStore()
const routeStore = useRouteStore()
const checkinStore = useCheckinStore()

const routeCount = ref(0)
const checkinCount = ref(0)

onShow(async () => {
  if (authStore.isLoggedIn) {
    await userStore.fetchProfile()
    await routeStore.fetchRoutes(1)
    routeCount.value = routeStore.total
    try { await checkinStore.fetchCheckins({ page: 1 }); checkinCount.value = checkinStore.total } catch (_) {}
  }
})

function goEditProfile() { uni.navigateTo({ url: '/pages/profile-edit/index' }) }
function goChangePassword() { uni.navigateTo({ url: '/pages/password-change/index' }) }
function goSettings() { uni.navigateTo({ url: '/pages/settings/index' }) }
function onTabChange(key: string) { uni.switchTab({ url: `/pages/${key === 'discover' ? 'index' : key}/index` }) }
function goRoutes() { uni.switchTab({ url: '/pages/routes/index' }) }
function goMyCheckins() { uni.switchTab({ url: '/pages/routes/index' }) }
function goMyPosts() { uni.navigateTo({ url: '/pages/my-posts/index' }) }

function goAgent() { uni.navigateTo({ url: '/pages/agent/index' }) }
function goPreferences() { uni.navigateTo({ url: '/pages/preferences/index' }) }

function handleLogout() {
  uni.showModal({
    title: '退出登录', content: '确定要退出登录吗？',
    success: (res) => {
      if (res.confirm) { authStore.logout(); userStore.clearProfile(); uni.reLaunch({ url: '/pages/login/index' }) }
    },
  })
}
</script>

<style lang="scss" scoped>
.profile-page { min-height: 100vh; background: linear-gradient(170deg, #FDF8F4, #F7F4F0, #F3F1F5, #F5F3F8); }

.navbar {
  position: sticky; top: 0; z-index: 50;
  display: flex; align-items: center; justify-content: center;
  height: 88rpx; padding: 0 32rpx;
  background: rgba(255,255,255,0.72);
  backdrop-filter: saturate(170%) blur(24px);
  -webkit-backdrop-filter: saturate(170%) blur(24px);
  border-bottom: 1px solid rgba(255,255,255,0.4);
  .nav-title { font-size: 34rpx; font-weight: 600; color: #1C1C1E; }
  .nav-action { position: absolute; right: 32rpx; font-size: 28rpx; color: #8E8E93; }
}

.page-body { padding: 16rpx 36rpx 120rpx; }

// 头像区 — 对齐原型 .profile-header
.profile-header { text-align: center; padding: 28rpx 0 8rpx;
  .ph-avatar { width: 128rpx; height: 128rpx; border-radius: 50%; overflow: hidden; margin: 0 auto 12rpx;
    background: linear-gradient(135deg,#E5E5EA,#D1D1D6); display: flex; align-items: center; justify-content: center; }
  .avatar-img { width: 100%; height: 100%; object-fit: cover; }
  .avatar-placeholder { font-size: 56rpx; }
  .ph-name { font-size: 40rpx; font-weight: 700; color: #1C1C1E; display: block; }
  .ph-id { font-size: 24rpx; color: #AEAEB2; margin-top: 4rpx; display: block; }
}

// 统计行 — 对齐原型 .stat-row
.stat-row { display: flex; justify-content: center; gap: 56rpx; margin: 24rpx 0;
  .stat-item { text-align: center; }
  .stat-num { font-size: 40rpx; font-weight: 700; color: #1C1C1E; display: block; }
  .stat-label { font-size: 22rpx; color: #AEAEB2; margin-top: 4rpx; }
}

// 菜单卡片（阴影在外层，圆角+溢出裁剪在内层）
.menu-group { margin-bottom: 20rpx;
  box-shadow: 0 8px 32px rgba(0,0,0,0.06), 0 2px 8px rgba(0,0,0,0.03);
  border-radius: 24rpx; }
.menu-inner { overflow: hidden; border-radius: 24rpx;
  background: rgba(255,255,255,0.55); }
.menu-item { display: flex; align-items: center; padding: 28rpx 32rpx;
  & + .menu-item { border-top: 1rpx solid rgba(0,0,0,0.05); }
  &:active { background: rgba(0,0,0,0.03); }
  .m-left { display: flex; align-items: center; gap: 16rpx; }
  .m-icon { font-size: 32rpx; }
  .m-text { font-size: 28rpx; color: #1C1C1E; }
  .m-right { margin-left: auto; font-size: 24rpx; color: #C7C7CC; }
}

// 退出 — 对齐原型 .logout-btn
.logout-btn { width: calc(100% - 72rpx); height: 88rpx; border-radius: 16rpx; margin: 28rpx 36rpx 0;
  background: rgba(255,59,48,0.1); border: 1px solid rgba(255,59,48,0.25);
  color: #FF3B30; font-size: 30rpx; font-weight: 600;
  display: flex; align-items: center; justify-content: center; line-height: 1;
  &:active { background: rgba(255,59,48,0.2); }
}

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
