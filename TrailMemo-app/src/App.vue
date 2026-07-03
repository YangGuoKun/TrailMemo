<script setup lang="ts">
import { onLaunch, onShow } from '@dcloudio/uni-app'
import { useAuthStore } from '@/stores/useAuthStore'
import { useAppStore } from '@/stores/useAppStore'

const appStore = useAppStore()
const authStore = useAuthStore()

onLaunch(() => {
  // 恢复登录状态
  authStore.restoreToken()
  // 检查网络状态
  appStore.checkNetwork()
})

onShow(() => {
  // App 进入前台时刷新网络状态
  appStore.checkNetwork()
})

// 全局路由拦截 — 检查需要认证的页面
const authRequiredPages = [
  'pages/routes/index',
  'pages/profile/index',
  'pages/route-detail/index',
  'pages/route-create/index',
  'pages/checkin-create/index',
  'pages/checkpoint-detail/index',
  'pages/profile-edit/index',
  'pages/password-change/index',
  'pages/settings/index',
  'pages/agent/index',
  'pages/preferences/index',
]

// 使用 uni.addInterceptor 做路由守卫
uni.addInterceptor('navigateTo', {
  invoke(args) {
    const url = (args as any).url || ''
    const pagePath = url.split('?')[0]
    if (authRequiredPages.includes(pagePath) && !authStore.isLoggedIn) {
      uni.reLaunch({ url: '/pages/login/index' })
      return false
    }
  },
})

uni.addInterceptor('switchTab', {
  invoke(args) {
    const url = (args as any).url || ''
    const pagePath = url.split('?')[0]
    if (authRequiredPages.includes(pagePath) && !authStore.isLoggedIn) {
      uni.reLaunch({ url: '/pages/login/index' })
      return false
    }
  },
})
</script>

<style lang="scss">
@import 'uview-plus/index.scss';
@import '@/styles/reset.scss';
@import '@/styles/common.scss';
</style>
