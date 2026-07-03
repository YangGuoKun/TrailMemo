/**
 * 认证状态管理
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { config } from '@/config'
import * as userApi from '@/api/user'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(null)
  const loginMethod = ref<'wechat' | 'password' | null>(null)

  // 是否已登录
  const isLoggedIn = computed(() => !!token.value)

  // 请求头 Authorization
  const authorizationHeader = computed(() => `Bearer ${token.value}`)

  // 恢复本地 token
  function restoreToken(): string | null {
    try {
      const saved = uni.getStorageSync(config.tokenKey)
      if (saved) {
        token.value = saved
        return saved
      }
    } catch (_) {
      // ignore
    }
    return null
  }

  // 保存 token
  function saveToken(t: string) {
    token.value = t
    uni.setStorageSync(config.tokenKey, t)
  }

  // 清除 token
  function clearToken() {
    token.value = null
    loginMethod.value = null
    uni.removeStorageSync(config.tokenKey)
  }

  // 密码登录
  async function loginByPassword(username: string, password: string) {
    const res = await userApi.login({ username, password })
    saveToken(res.token)
    loginMethod.value = 'password'
  }

  // 微信登录
  async function loginByWechat() {
    return new Promise<void>((resolve, reject) => {
      uni.login({
        provider: 'weixin',
        success: async (loginRes) => {
          try {
            const res = await userApi.loginByWechat(loginRes.code!)
            saveToken(res.token)
            loginMethod.value = 'wechat'
            resolve()
          } catch (err) {
            reject(err)
          }
        },
        fail: (err) => {
          uni.showToast({ title: '微信登录失败', icon: 'none' })
          reject(err)
        },
      })
    })
  }

  // 注册
  async function register(data: RegisterRequest) {
    await userApi.register(data)
  }

  // 退出登录
  function logout() {
    clearToken()
  }

  return {
    token,
    loginMethod,
    isLoggedIn,
    authorizationHeader,
    restoreToken,
    saveToken,
    clearToken,
    loginByPassword,
    loginByWechat,
    register,
    logout,
  }
})
