/**
 * 全局应用状态管理
 */

import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAppStore = defineStore('app', () => {
  const isOnline = ref(true)
  const isFirstLaunch = ref(true)
  const currentLocation = ref<{ lat: number; lng: number } | null>(null)

  // 检查网络状态
  function checkNetwork() {
    uni.getNetworkType({
      success: (res) => {
        isOnline.value = res.networkType !== 'none'
      },
    })
    uni.onNetworkStatusChange((res) => {
      isOnline.value = res.isConnected
    })
  }

  // 请求位置权限并获取当前位置
  function requestLocation(): Promise<{ lat: number; lng: number }> {
    return new Promise((resolve, reject) => {
      uni.getLocation({
        type: 'gcj02',
        success: (res) => {
          const loc = { lat: res.latitude, lng: res.longitude }
          currentLocation.value = loc
          resolve(loc)
        },
        fail: (err) => {
          uni.showModal({
            title: '需要获取您的位置',
            content: '用于在地图上显示您的当前位置和记录打卡坐标',
            success: (modalRes) => {
              if (modalRes.confirm) {
                uni.openSetting({})
              }
            },
          })
          reject(err)
        },
      })
    })
  }

  // 标记首次启动完成
  function setFirstLaunchDone() {
    isFirstLaunch.value = false
  }

  return {
    isOnline,
    isFirstLaunch,
    currentLocation,
    checkNetwork,
    requestLocation,
    setFirstLaunchDone,
  }
})
