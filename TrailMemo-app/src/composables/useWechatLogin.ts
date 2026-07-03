/**
 * 微信登录 Hook
 */

import { useAuthStore } from '@/stores/useAuthStore'
import { useUserStore } from '@/stores/useUserStore'

export function useWechatLogin() {
  const authStore = useAuthStore()
  const userStore = useUserStore()

  // 执行微信登录流程
  async function wechatLogin(): Promise<boolean> {
    try {
      await authStore.loginByWechat()
      await userStore.fetchProfile()
      uni.showToast({ title: '登录成功', icon: 'success' })
      return true
    } catch (err: any) {
      console.error('WeChat login failed:', err)
      return false
    }
  }

  // 微信授权并获取用户信息 (如需头像昵称)
  async function getWechatUserInfo(): Promise<WechatMiniprogram.UserInfo | null> {
    return new Promise((resolve) => {
      uni.getUserInfo({
        provider: 'weixin',
        success: (res) => resolve(res.userInfo),
        fail: () => resolve(null),
      })
    })
  }

  return {
    wechatLogin,
    getWechatUserInfo,
  }
}
