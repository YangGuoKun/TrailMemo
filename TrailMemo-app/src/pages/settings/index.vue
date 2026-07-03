<template>
  <view class="settings-page">
    <view class="navbar">
      <view class="navbar-back" @tap="goBack">
        <u-icon name="arrow-left" size="40" color="#1C1C1E" />
      </view>
      <text class="navbar-title">设置</text>
    </view>

    <view class="page-body">
      <!-- 通用设置 -->
      <view class="menu-section card">
        <view class="menu-item" @tap="clearCache">
          <text class="menu-text">清除缓存</text>
          <text class="menu-extra">{{ cacheSize }}</text>
        </view>
        <view class="menu-item">
          <text class="menu-text">当前版本</text>
          <text class="menu-extra">v1.0.0</text>
        </view>
      </view>

      <!-- 关于 -->
      <view class="menu-section card">
        <view class="menu-item">
          <text class="menu-text">关于迹忆旅图</text>
          <u-icon name="arrow-right" size="28" color="#C7C7CC" />
        </view>
        <view class="menu-item">
          <text class="menu-text">用户协议</text>
          <u-icon name="arrow-right" size="28" color="#C7C7CC" />
        </view>
        <view class="menu-item">
          <text class="menu-text">隐私政策</text>
          <u-icon name="arrow-right" size="28" color="#C7C7CC" />
        </view>
      </view>

      <!-- 退出登录 -->
      <button class="logout-btn" @tap="handleLogout">退出登录</button>

      <!-- 版权信息 -->
      <view class="copyright">
        <text class="copyright-text">迹忆旅图 © 2024</text>
        <text class="copyright-text">记录每一段旅途的足迹与回忆</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/useAuthStore'
import { useUserStore } from '@/stores/useUserStore'
import { useRouteStore } from '@/stores/useRouteStore'

const authStore = useAuthStore()
const userStore = useUserStore()
const routeStore = useRouteStore()

const cacheSize = ref('0 KB')

// 清除缓存
function clearCache() {
  uni.showModal({
    title: '清除缓存',
    content: '将清除本地缓存的图片和数据，不会影响您的账号信息',
    success: (res) => {
      if (res.confirm) {
        // 清除图片缓存
        uni.getSavedFileList({
          success: (fileList) => {
            fileList.fileList.forEach((file) => {
              uni.removeSavedFile({ filePath: file.filePath })
            })
          },
        })
        uni.showToast({ title: '缓存已清除', icon: 'success' })
        cacheSize.value = '0 KB'
      }
    },
  })
}

// 退出登录
function handleLogout() {
  uni.showModal({
    title: '退出登录',
    content: '确定要退出登录吗？',
    success: (res) => {
      if (res.confirm) {
        authStore.logout()
        userStore.clearProfile()
        routeStore.clearDetail()
        uni.reLaunch({ url: '/pages/login/index' })
      }
    },
  })
}

function goBack() {
  uni.navigateBack()
}
</script>

<style lang="scss" scoped>
.settings-page {
  min-height: 100vh;
  background: $color-bg-secondary;
}

.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: $z-sticky;
  height: $navbar-height;
  display: flex;
  align-items: center;
  justify-content: center;
  background: $glass-bg;
  backdrop-filter: blur($glass-blur);
  padding-top: constant(safe-area-inset-top);
  padding-top: env(safe-area-inset-top);

  .navbar-back {
    position: absolute;
    left: $page-inset;
    top: 50%;
    transform: translateY(-50%);
    padding-top: constant(safe-area-inset-top);
    padding-top: env(safe-area-inset-top);
  }

  .navbar-title {
    font-size: $font-size-headline;
    font-weight: $font-weight-semibold;
    color: $color-gray-900;
  }
}

.page-body {
  padding-top: calc($navbar-height + constant(safe-area-inset-top) + $space-4);
  padding-top: calc($navbar-height + env(safe-area-inset-top) + $space-4);
  padding: $space-4 $page-inset;
  padding-top: calc($navbar-height + constant(safe-area-inset-top) + $space-4);
  padding-top: calc($navbar-height + env(safe-area-inset-top) + $space-4);
}

.menu-section {
  margin-bottom: $space-4;
  overflow: hidden;

  .menu-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: $space-4 $space-5;

    & + .menu-item {
      border-top: 1rpx solid $color-gray-100;
    }

    .menu-text {
      font-size: $font-size-callout;
      color: $color-gray-800;
    }

    .menu-extra {
      font-size: $font-size-footnote;
      color: $color-gray-400;
    }
  }
}

.logout-btn {
  width: 100%;
  height: $button-height;
  background: $color-bg-primary;
  color: $color-error;
  border-radius: $radius-md;
  font-size: $font-size-callout;
  font-weight: $font-weight-medium;
  margin-top: $space-8;

  &:active {
    background: $color-gray-100;
  }
}

.copyright {
  text-align: center;
  padding: $space-10 0;

  .copyright-text {
    display: block;
    font-size: $font-size-caption2;
    color: $color-gray-400;
    line-height: 2;
  }
}
</style>
