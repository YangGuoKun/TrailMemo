/**
 * 用户信息状态管理
 */

import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as userApi from '@/api/user'
import * as uploadApi from '@/api/upload'
import { getFullUrl } from '@/config'

export const useUserStore = defineStore('user', () => {
  const profile = ref<UserProfile | null>(null)
  const isLoading = ref(false)

  // 获取用户信息
  async function fetchProfile() {
    isLoading.value = true
    try {
      const p = await userApi.getUserInfo()
      if (p) {
        p.avatar = getFullUrl(p.avatar)
      }
      profile.value = p
    } finally {
      isLoading.value = false
    }
  }

  // 更新用户信息
  async function updateProfile(data: UpdateUserRequest) {
    await userApi.updateUserInfo(data)
    if (profile.value) {
      if (data.nickname !== undefined) profile.value.nickname = data.nickname
      if (data.avatar !== undefined) profile.value.avatar = data.avatar
    }
  }

  // 修改密码
  async function changePassword(oldPassword: string, newPassword: string) {
    await userApi.changePassword({ oldPassword, newPassword })
  }

  // 上传头像
  async function uploadAvatar(filePath: string): Promise<string> {
    const res = await uploadApi.uploadAvatar(filePath)
    const fullUrl = getFullUrl(res.url)
    if (profile.value) {
      profile.value.avatar = fullUrl
    }
    return fullUrl
  }

  // 清除用户信息
  function clearProfile() {
    profile.value = null
    isLoading.value = false
  }

  return {
    profile,
    isLoading,
    fetchProfile,
    updateProfile,
    changePassword,
    uploadAvatar,
    clearProfile,
  }
})
