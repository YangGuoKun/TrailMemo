<template>
  <view class="login-page">
    <!-- 品牌区域 -->
    <view class="brand-section">
      <view class="logo-container">
        <image class="logo" src="/static/images/logo.png" mode="aspectFit" />
      </view>
      <text class="app-name">迹忆旅图</text>
      <text class="app-slogan">记录每一段旅途的足迹与回忆</text>
    </view>

    <!-- 登录表单 -->
    <view class="form-section">
      <!-- 微信一键登录 -->
      <button class="wechat-btn" @tap="handleWechatLogin" :loading="wechatLoading">
        <u-icon name="weixin-fill" size="40" color="#FFFFFF" />
        <text class="wechat-btn-text">微信一键登录</text>
      </button>

      <!-- 分割线 -->
      <view class="divider">
        <view class="divider-line" />
        <text class="divider-text">或</text>
        <view class="divider-line" />
      </view>

      <!-- 账号密码登录 -->
      <view class="password-form" v-if="showPasswordForm">
        <u-form :model="loginForm" :rules="rules" ref="formRef">
          <u-form-item prop="username" border-bottom>
            <u-input
              v-model="loginForm.username"
              placeholder="请输入用户名"
              prefixIcon="account"
              clearable
            />
          </u-form-item>
          <u-form-item prop="password" border-bottom>
            <u-input
              v-model="loginForm.password"
              type="password"
              placeholder="请输入密码"
              prefixIcon="lock"
            />
          </u-form-item>
        </u-form>

        <button class="login-btn" @tap="handlePasswordLogin" :loading="loginLoading">
          登录
        </button>

        <view class="form-footer">
          <text class="link-text" @tap="showRegister = true; showPasswordForm = false">
            还没有账号？立即注册
          </text>
        </view>
      </view>

      <!-- 注册表单 -->
      <view class="register-form" v-if="showRegister">
        <u-form :model="registerForm" :rules="registerRules" ref="registerFormRef">
          <u-form-item prop="username" border-bottom>
            <u-input
              v-model="registerForm.username"
              placeholder="请输入用户名 (3-64位)"
              prefixIcon="account"
              clearable
            />
          </u-form-item>
          <u-form-item prop="password" border-bottom>
            <u-input
              v-model="registerForm.password"
              type="password"
              placeholder="请输入密码 (至少6位)"
              prefixIcon="lock"
            />
          </u-form-item>
          <u-form-item prop="confirmPassword" border-bottom>
            <u-input
              v-model="registerForm.confirmPassword"
              type="password"
              placeholder="请确认密码"
              prefixIcon="lock"
            />
          </u-form-item>
          <u-form-item prop="phone" border-bottom>
            <u-input
              v-model="registerForm.phone"
              placeholder="手机号 (选填)"
              prefixIcon="phone"
            />
          </u-form-item>
        </u-form>

        <button class="login-btn" @tap="handleRegister" :loading="registerLoading">
          注册
        </button>

        <view class="form-footer">
          <text class="link-text" @tap="showRegister = false; showPasswordForm = true">
            已有账号？返回登录
          </text>
        </view>
      </view>

      <!-- 切换按钮 -->
      <view class="toggle-btn" v-if="!showPasswordForm && !showRegister">
        <text class="link-text" @tap="showPasswordForm = true">
          使用账号密码登录
        </text>
      </view>
    </view>

    <!-- 底部协议 -->
    <view class="agreement-text">
      <text class="agreement-desc">登录即表示同意</text>
      <text class="agreement-link">《用户协议》</text>
      <text class="agreement-desc">和</text>
      <text class="agreement-link">《隐私政策》</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useAuthStore } from '@/stores/useAuthStore'
import { useUserStore } from '@/stores/useUserStore'
import { useWechatLogin } from '@/composables/useWechatLogin'

const authStore = useAuthStore()
const userStore = useUserStore()
const { wechatLogin } = useWechatLogin()

// UI 状态
const showPasswordForm = ref(false)
const showRegister = ref(false)
const wechatLoading = ref(false)
const loginLoading = ref(false)
const registerLoading = ref(false)

// 登录表单
const loginForm = reactive({
  username: '',
  password: '',
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: ['blur', 'change'] }],
  password: [{ required: true, message: '请输入密码', trigger: ['blur', 'change'] }],
}

// 注册表单
const registerForm = reactive({
  username: '',
  password: '',
  confirmPassword: '',
  phone: '',
})

const registerRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: ['blur', 'change'] },
    { min: 3, max: 64, message: '用户名长度 3-64 位', trigger: ['blur', 'change'] },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: ['blur', 'change'] },
    { min: 6, message: '密码至少 6 位', trigger: ['blur', 'change'] },
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: ['blur', 'change'] },
    {
      validator: (_: any, value: string) => value === registerForm.password,
      message: '两次密码不一致',
      trigger: ['blur', 'change'],
    },
  ],
}

// 微信登录
async function handleWechatLogin() {
  wechatLoading.value = true
  try {
    const success = await wechatLogin()
    if (success) {
      navigateToHome()
    }
  } finally {
    wechatLoading.value = false
  }
}

// 密码登录
async function handlePasswordLogin() {
  if (!loginForm.username || !loginForm.password) {
    uni.showToast({ title: '请填写用户名和密码', icon: 'none' })
    return
  }

  loginLoading.value = true
  try {
    await authStore.loginByPassword(loginForm.username, loginForm.password)
    await userStore.fetchProfile()
    uni.showToast({ title: '登录成功', icon: 'success' })
    setTimeout(navigateToHome, 500)
  } catch (_) {
    // 错误已由 request 拦截器处理
  } finally {
    loginLoading.value = false
  }
}

// 注册
async function handleRegister() {
  if (!registerForm.username || !registerForm.password) {
    uni.showToast({ title: '请填写用户名和密码', icon: 'none' })
    return
  }
  if (registerForm.password !== registerForm.confirmPassword) {
    uni.showToast({ title: '两次密码不一致', icon: 'none' })
    return
  }

  registerLoading.value = true
  try {
    // 手机号自动补 +86 前缀（后端走 E.164 校验）
    let phone = registerForm.phone || undefined
    if (phone && /^\d{11}$/.test(phone)) {
      phone = '+86' + phone
    }
    await authStore.register({
      username: registerForm.username,
      password: registerForm.password,
      phone,
    })
    uni.showToast({ title: '注册成功，正在登录...', icon: 'success' })
    await authStore.loginByPassword(registerForm.username, registerForm.password)
    await userStore.fetchProfile()
    setTimeout(navigateToHome, 500)
  } catch (_) {
    // 错误已处理
  } finally {
    registerLoading.value = false
  }
}

// 跳转到首页
function navigateToHome() {
  uni.switchTab({ url: '/pages/index/index' })
}
</script>

<style lang="scss" scoped>
.login-page {
  min-height: 100vh;
  background: $color-bg-primary;
  display: flex;
  flex-direction: column;
  padding: 0 $page-inset;
}

// 品牌区域
.brand-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding-top: 120rpx;

  .logo-container {
    width: 160rpx;
    height: 160rpx;
    border-radius: $radius-xl;
    background: linear-gradient(135deg, $color-primary-400, $color-primary-600);
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: $space-6;
    box-shadow: 0 8rpx 32rpx rgba($color-primary-500, 0.3);

    .logo {
      width: 100rpx;
      height: 100rpx;
    }
  }

  .app-name {
    font-size: $font-size-title1;
    font-weight: $font-weight-bold;
    color: $color-gray-900;
    margin-bottom: $space-3;
  }

  .app-slogan {
    font-size: $font-size-callout;
    color: $color-gray-500;
  }
}

// 表单区域
.form-section {
  padding: $space-12 0 $space-8;
}

// 微信登录按钮
.wechat-btn {
  width: 100%;
  height: $button-height;
  background: $color-wechat;
  border-radius: $radius-md;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: $space-2;
  box-shadow: 0 4rpx 16rpx rgba($color-wechat, 0.3);
  transition: all $duration-fast $spring-default;

  &:active {
    transform: scale(0.97);
    opacity: 0.9;
  }

  .wechat-btn-text {
    color: #FFFFFF;
    font-size: $font-size-body;
    font-weight: $font-weight-semibold;
  }
}

// 分割线
.divider {
  display: flex;
  align-items: center;
  margin: $space-8 0;
  gap: $space-4;

  .divider-line {
    flex: 1;
    height: 1rpx;
    background: $color-gray-200;
  }

  .divider-text {
    font-size: $font-size-subhead;
    color: $color-gray-400;
  }
}

// 登录按钮
.login-btn {
  width: 100%;
  height: $button-height;
  background: $color-primary-500;
  color: #FFFFFF;
  border-radius: $radius-md;
  font-size: $font-size-body;
  font-weight: $font-weight-semibold;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: $space-6;
  box-shadow: 0 4rpx 16rpx rgba($color-primary-500, 0.3);
  transition: all $duration-fast $spring-default;

  &:active {
    background: $color-primary-600;
    transform: scale(0.97);
  }
}

// 表单底部链接
.form-footer {
  text-align: center;
  margin-top: $space-6;
}

.link-text {
  font-size: $font-size-subhead;
  color: $color-info;
}

.toggle-btn {
  text-align: center;
  margin-top: $space-8;
}

// 协议
.agreement-text {
  text-align: center;
  padding: $space-6 0;
  padding-bottom: constant(safe-area-inset-bottom);
  padding-bottom: env(safe-area-inset-bottom);

  .agreement-desc {
    font-size: $font-size-caption2;
    color: $color-gray-400;
  }

  .agreement-link {
    font-size: $font-size-caption2;
    color: $color-info;
  }
}
</style>
