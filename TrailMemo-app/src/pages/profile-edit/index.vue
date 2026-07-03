<template>
  <view class="edit-page">
    <view class="navbar">
      <view class="navbar-back" @tap="goBack">
        <u-icon name="arrow-left" size="40" color="#1C1C1E" />
      </view>
      <text class="navbar-title">编辑资料</text>
      <view class="navbar-save" @tap="handleSave">
        <text class="save-text">保存</text>
      </view>
    </view>

    <view class="page-body">
      <!-- 头像 -->
      <view class="avatar-section card" @tap="handleChangeAvatar">
        <text class="section-label">头像</text>
        <view class="avatar-right">
          <view class="avatar-wrap">
            <image
              v-if="avatarUrl"
              :src="avatarUrl"
              mode="aspectFill"
              class="avatar-img"
            />
            <view v-else class="avatar-placeholder">
              <u-icon name="account" size="64" color="#C7C7CC" />
            </view>
          </view>
          <u-icon name="arrow-right" size="28" color="#C7C7CC" />
        </view>
      </view>

      <!-- 昵称 -->
      <view class="info-section card">
        <u-input
          v-model="nickname"
          placeholder="设置昵称"
          border="bottom"
          clearable
        />
      </view>

      <!-- 账号信息 (只读) -->
      <view class="readonly-section card">
        <view class="info-row">
          <text class="info-label">用户名</text>
          <text class="info-value">{{ profile?.username || '-' }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">手机号</text>
          <text class="info-value">{{ profile?.phone || '未绑定' }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">邮箱</text>
          <text class="info-value">{{ profile?.email || '未绑定' }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">注册时间</text>
          <text class="info-value">{{ formatDate(profile?.createdAt) }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useUserStore } from '@/stores/useUserStore'
import { useImagePicker } from '@/composables/useImagePicker'

const userStore = useUserStore()
const { pickImage } = useImagePicker()

const profile = ref<UserProfile | null>(null)
const nickname = ref('')
const avatarUrl = ref('')

onShow(() => {
  profile.value = userStore.profile
  nickname.value = profile.value?.nickname || ''
  avatarUrl.value = profile.value?.avatar || ''
})

// 更换头像
async function handleChangeAvatar() {
  try {
    const path = await pickImage()
    const url = await userStore.uploadAvatar(path)
    avatarUrl.value = url
    uni.showToast({ title: '头像已更新', icon: 'success' })
  } catch (_) {
    // cancelled or failed
  }
}

// 保存
async function handleSave() {
  try {
    await userStore.updateProfile({
      nickname: nickname.value,
      avatar: avatarUrl.value,
    })
    uni.showToast({ title: '保存成功', icon: 'success' })
    setTimeout(() => uni.navigateBack(), 500)
  } catch (_) {
    // handled
  }
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return '-'
  return dateStr.split('T')[0] || dateStr
}

function goBack() {
  uni.navigateBack()
}
</script>

<style lang="scss" scoped>
.edit-page {
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

  .navbar-back, .navbar-save {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    padding-top: constant(safe-area-inset-top);
    padding-top: env(safe-area-inset-top);
  }
  .navbar-back { left: $page-inset; }
  .navbar-save { right: $page-inset; }

  .navbar-title {
    font-size: $font-size-headline;
    font-weight: $font-weight-semibold;
    color: $color-gray-900;
  }

  .save-text {
    font-size: $font-size-callout;
    color: $color-primary-500;
    font-weight: $font-weight-semibold;
  }
}

.page-body {
  padding-top: calc($navbar-height + constant(safe-area-inset-top) + $space-4);
  padding-top: calc($navbar-height + env(safe-area-inset-top) + $space-4);
  padding: $space-4 $page-inset;
  padding-top: calc($navbar-height + constant(safe-area-inset-top) + $space-4);
  padding-top: calc($navbar-height + env(safe-area-inset-top) + $space-4);
}

.card {
  padding: $space-5;
  margin-bottom: $space-4;
}

.section-label {
  font-size: $font-size-callout;
  color: $color-gray-600;
}

.avatar-section {
  display: flex;
  align-items: center;
  justify-content: space-between;

  .avatar-right {
    display: flex;
    align-items: center;
    gap: $space-2;

    .avatar-wrap {
      width: $avatar-lg;
      height: $avatar-lg;
      border-radius: 50%;
      overflow: hidden;

      .avatar-img {
        width: 100%;
        height: 100%;
        object-fit: cover;
      }

      .avatar-placeholder {
        width: 100%;
        height: 100%;
        background: $color-gray-100;
        display: flex;
        align-items: center;
        justify-content: center;
      }
    }
  }
}

.readonly-section {
  .info-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: $space-3 0;

    & + .info-row {
      border-top: 1rpx solid $color-gray-100;
    }

    .info-label {
      font-size: $font-size-callout;
      color: $color-gray-600;
    }
    .info-value {
      font-size: $font-size-subhead;
      color: $color-gray-800;
    }
  }
}
</style>
