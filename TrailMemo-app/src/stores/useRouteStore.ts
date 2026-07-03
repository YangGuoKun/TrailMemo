/**
 * 路线状态管理
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as routeApi from '@/api/route'
import { config } from '@/config'

export const useRouteStore = defineStore('route', () => {
  // 列表
  const routes = ref<Route[]>([])
  const total = ref(0)
  const currentPage = ref(1)
  const isListLoading = ref(false)

  // 详情
  const currentDetail = ref<RouteDetail | null>(null)
  const isDetailLoading = ref(false)

  // 筛选
  const statusFilter = ref<string>('all')

  // 获取路线列表
  async function fetchRoutes(page: number = 1) {
    isListLoading.value = true
    try {
      const res = await routeApi.getRouteList(page, config.defaultPageSize)
      if (page === 1) {
        routes.value = res.list
      } else {
        routes.value.push(...res.list)
      }
      total.value = res.total
      currentPage.value = page
    } finally {
      isListLoading.value = false
    }
  }

  // 加载更多
  async function fetchMoreRoutes() {
    if (routes.value.length >= total.value) return
    await fetchRoutes(currentPage.value + 1)
  }

  // 刷新列表
  async function refreshRoutes() {
    await fetchRoutes(1)
  }

  // 获取路线详情
  async function fetchRouteDetail(id: number): Promise<RouteDetail> {
    isDetailLoading.value = true
    try {
      const detail = await routeApi.getRouteDetail(id)
      currentDetail.value = detail
      return detail
    } finally {
      isDetailLoading.value = false
    }
  }

  // 创建路线
  async function createRoute(data: CreateRouteRequest): Promise<Route> {
    const route = await routeApi.createRoute(data)
    // 刷新列表
    await fetchRoutes(1)
    return route
  }

  // 更新路线
  async function updateRoute(id: number, data: UpdateRouteRequest) {
    await routeApi.updateRoute(id, data)
    // 刷新详情
    if (currentDetail.value?.id === id) {
      await fetchRouteDetail(id)
    }
    // 刷新列表
    await fetchRoutes(1)
  }

  // 删除路线
  async function deleteRoute(id: number) {
    await routeApi.deleteRoute(id)
    // 从列表中移除
    routes.value = routes.value.filter((r) => r.id !== id)
    total.value--
    if (currentDetail.value?.id === id) {
      currentDetail.value = null
    }
  }

  // 复制路线
  async function copyRoute(id: number, isPublic: number = 1): Promise<Route> {
    const route = await routeApi.copyRoute(id, isPublic)
    await fetchRoutes(1)
    return route
  }

  // 清除详情
  function clearDetail() {
    currentDetail.value = null
    isDetailLoading.value = false
  }

  // 根据状态筛选 (本地)
  const filteredRoutes = computed(() => {
    if (statusFilter.value === 'all') return routes.value
    let publishStatus: number
    switch (statusFilter.value) {
      case 'in_progress': publishStatus = 1; break
      case 'completed': publishStatus = 2; break
      case 'draft': publishStatus = 0; break
      default: return routes.value
    }
    return routes.value.filter((r) => r.publish_status === publishStatus)
  })

  return {
    routes,
    total,
    currentPage,
    isListLoading,
    currentDetail,
    isDetailLoading,
    statusFilter,
    filteredRoutes,
    fetchRoutes,
    fetchMoreRoutes,
    refreshRoutes,
    fetchRouteDetail,
    createRoute,
    updateRoute,
    deleteRoute,
    copyRoute,
    clearDetail,
  }
})
