<template>
  <view class="pwd-page">
    <view class="navbar">
      <view class="navbar-back" @tap="goBack">
        <u-icon name="arrow-left" size="40" color="#1C1C1E" />
      </view>
      <text class="navbar-title">修改密码</text>
    </view>

    <view class="page-body">
      <view class="form-section card">
        <u-input
          v-model="oldPassword"
          type="password"
          placeholder="请输入旧密码"
          border="bottom"
          prefixIcon="lock"
        />
        <u-input
          v-model="newPassword"
          type="password"
          placeholder="请输入新密码 (至少6位)"
          border="bottom"
          prefixIcon="lock"
        />
        <u-input
          v-model="confirmPassword"
          type="password"
          placeholder="请确认新密码"
          border="bottom"
          prefixIcon="lock"
        />
      </view>

      <button class="submit-btn" @tap="handleSubmit">确认修改</button>

      <view class="tips">
        <text class="tip-title">安全提示</text>
        <text class="tip-text">· 密码长度至少 6 位</text>
        <text class="tip-text">· 建议使用字母、数字和符号的组合</text>
        <text class="tip-text">· 定期更换密码以保证账户安全</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useUserStore } from '@/stores/useUserStore'

const userStore = useUserStore()

const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')

async function handleSubmit() {
  if (!oldPassword.value) {
    uni.showToast({ title: '请输入旧密码', icon: 'none' })
    return
  }
  if (newPassword.value.length < 6) {
    uni.showToast({ title: '新密码至少 6 位', icon: 'none' })
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    uni.showToast({ title: '两次密码不一致', icon: 'none' })
    return
  }

  try {
    await userStore.changePassword(oldPassword.value, newPassword.value)
    uni.showToast({ title: '密码修改成功', icon: 'success' })
    setTimeout(() => uni.navigateBack(), 800)
  } catch (_) {
    // handled
  }
}

function goBack() {
  uni.navigateBack()
}
</script>

<style lang="scss" scoped>
.pwd-page {
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

.form-section {
  padding: 0 $space-5;
  margin-bottom: $space-6;
}

.submit-btn {
  width: 100%;
  height: $button-height;
  background: $color-primary-500;
  color: #FFFFFF;
  border-radius: $radius-md;
  font-size: $font-size-callout;
  font-weight: $font-weight-semibold;
  margin-bottom: $space-8;

  &:active {
    background: $color-primary-600;
  }
}

.tips {
  padding: $space-5;
  background: rgba($color-warning, 0.08);
  border-radius: $radius-md;

  .tip-title {
    display: block;
    font-size: $font-size-subhead;
    font-weight: $font-weight-semibold;
    color: $color-warning;
    margin-bottom: $space-2;
  }

  .tip-text {
    display: block;
    font-size: $font-size-footnote;
    color: $color-gray-600;
    line-height: 2;
  }
}
</style>
