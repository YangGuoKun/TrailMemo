/**
 * 位置服务 Hook
 */

import { ref } from 'vue'

export function useLocation() {
  const currentLat = ref<number>(0)
  const currentLng = ref<number>(0)
  const locationReady = ref(false)

  // 获取当前位置 (gcj02 坐标系 — 微信/腾讯地图使用)
  async function getCurrentLocation(): Promise<{ lat: number; lng: number }> {
    return new Promise((resolve, reject) => {
      uni.getLocation({
        type: 'gcj02',
        success: (res) => {
          currentLat.value = res.latitude
          currentLng.value = res.longitude
          locationReady.value = true
          resolve({ lat: res.latitude, lng: res.longitude })
        },
        fail: (err) => {
          uni.showModal({
            title: '需要位置权限',
            content: '用于地图展示和打卡定位，请在设置中开启',
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

  // 打开地图选择位置 (用户拖动标记)
  function chooseLocation(): Promise<{
    name: string
    address: string
    lat: number
    lng: number
  }> {
    return new Promise((resolve, reject) => {
      uni.chooseLocation({
        success: (res) => {
          resolve({
            name: res.name || '',
            address: res.address || '',
            lat: res.latitude,
            lng: res.longitude,
          })
        },
        fail: reject,
      })
    })
  }

  return {
    currentLat,
    currentLng,
    locationReady,
    getCurrentLocation,
    chooseLocation,
  }
}
