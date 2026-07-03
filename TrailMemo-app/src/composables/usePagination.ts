/**
 * 通用分页加载 Hook
 */

import { ref } from 'vue'

export function usePagination<T>(
  fetchFn: (page: number) => Promise<{ list: T[]; total: number }>,
  pageSize: number = 20,
) {
  const list = ref<T[]>([]) as Ref<T[]>
  const total = ref(0)
  const page = ref(1)
  const isLoading = ref(false)
  const isFinished = ref(false)

  // 加载第一页
  async function loadFirst() {
    page.value = 1
    isFinished.value = false
    isLoading.value = true
    try {
      const res = await fetchFn(1)
      list.value = res.list
      total.value = res.total
      if (res.list.length >= res.total) {
        isFinished.value = true
      }
    } finally {
      isLoading.value = false
    }
  }

  // 加载下一页
  async function loadMore() {
    if (isLoading.value || isFinished.value) return
    isLoading.value = true
    try {
      const nextPage = page.value + 1
      const res = await fetchFn(nextPage)
      list.value.push(...res.list)
      total.value = res.total
      page.value = nextPage
      if (list.value.length >= res.total) {
        isFinished.value = true
      }
    } finally {
      isLoading.value = false
    }
  }

  // 重置
  function reset() {
    list.value = []
    total.value = 0
    page.value = 1
    isFinished.value = false
    isLoading.value = false
  }

  return {
    list,
    total,
    page,
    isLoading,
    isFinished,
    loadFirst,
    loadMore,
    reset,
  }
}
