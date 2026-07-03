/**
 * 打卡状态管理
 */

import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as checkinApi from '@/api/checkin'
import { config } from '@/config'

export const useCheckinStore = defineStore('checkin', () => {
  const checkins = ref<Checkin[]>([])
  const total = ref(0)
  const currentPage = ref(1)
  const isListLoading = ref(false)
  const isSubmitting = ref(false)

  const currentDetail = ref<Checkin | null>(null)
  const progress = ref(0)

  // 获取打卡列表
  async function fetchCheckins(params: {
    page?: number
    route_id?: number
  } = {}) {
    isListLoading.value = true
    try {
      const page = params.page || 1
      const res = await checkinApi.getCheckinList({
        page,
        size: config.defaultPageSize,
        route_id: params.route_id,
      })
      if (page === 1) {
        checkins.value = res.list
      } else {
        checkins.value.push(...res.list)
      }
      total.value = res.total
      currentPage.value = page
    } finally {
      isListLoading.value = false
    }
  }

  // 获取打卡详情
  async function fetchCheckinDetail(id: number) {
    currentDetail.value = await checkinApi.getCheckinDetail(id)
  }

  // 获取路线打卡进度
  async function fetchRouteProgress(routeId: number): Promise<number> {
    const res = await checkinApi.getRouteProgress(routeId)
    progress.value = res.progress
    return res.progress
  }

  // 创建打卡
  async function createCheckin(data: CreateCheckinRequest): Promise<Checkin> {
    isSubmitting.value = true
    try {
      const checkin = await checkinApi.createCheckin(data)
      progress.value = 100 // 需要刷新
      return checkin
    } finally {
      isSubmitting.value = false
    }
  }

  // 更新打卡
  async function updateCheckin(id: number, data: UpdateCheckinRequest) {
    await checkinApi.updateCheckin(id, data)
    if (currentDetail.value?.id === id) {
      await fetchCheckinDetail(id)
    }
  }

  // 删除打卡
  async function deleteCheckin(id: number) {
    await checkinApi.deleteCheckin(id)
    checkins.value = checkins.value.filter((c) => c.id !== id)
    total.value--
    if (currentDetail.value?.id === id) {
      currentDetail.value = null
    }
  }

  // 清除状态
  function clearState() {
    checkins.value = []
    total.value = 0
    currentPage.value = 1
    currentDetail.value = null
    progress.value = 0
  }

  return {
    checkins,
    total,
    currentPage,
    isListLoading,
    isSubmitting,
    currentDetail,
    progress,
    fetchCheckins,
    fetchCheckinDetail,
    fetchRouteProgress,
    createCheckin,
    updateCheckin,
    deleteCheckin,
    clearState,
  }
})
