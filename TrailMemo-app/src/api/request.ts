/**
 * TrailMemo API 请求封装
 * 统一处理 {code, message, data} 响应信封
 */

import { config } from '@/config'
import { useAuthStore } from '@/stores/useAuthStore'

const BASE_URL = config.apiBaseUrl
const TIMEOUT = 15000

/** 请求选项 */
interface RequestOptions {
  /** 是否显示 loading */
  loading?: boolean
  /** 是否显示错误 toast */
  showError?: boolean
}

/**
 * 请求拦截 — 解包响应信封
 * code=200 → 返回 data
 * code=401 → 清除 token → 跳转 login
 * 其他 → 抛出错误
 */
function unwrapResponse<T>(res: ApiResponse<T>, showError: boolean): T {
  if (res.code === 200) {
    return res.data
  }

  if (res.code === 401) {
    const authStore = useAuthStore()
    authStore.logout()
    uni.reLaunch({ url: '/pages/login/index' })
    throw new Error(res.message || '登录已过期，请重新登录')
  }

  if (showError) {
    uni.showToast({
      title: res.message || '请求失败',
      icon: 'none',
      duration: 2000,
    })
  }

  throw new Error(res.message || '请求失败')
}

/**
 * 基础请求方法
 */
function request<T>(
  method: 'GET' | 'POST' | 'PUT' | 'DELETE',
  url: string,
  data?: any,
  options: RequestOptions = {},
): Promise<T> {
  const { loading = false, showError = true } = options
  const authStore = useAuthStore()

  if (loading) {
    uni.showLoading({ title: '加载中...', mask: true })
  }

  return new Promise((resolve, reject) => {
    uni.request({
      url: BASE_URL + url,
      method,
      data,
      timeout: TIMEOUT,
      header: {
        'Content-Type': 'application/json',
        ...(authStore.token ? { Authorization: authStore.authorizationHeader } : {}),
      },
      success: (res) => {
        try {
          const result = unwrapResponse<T>(res.data as ApiResponse<T>, showError)
          resolve(result)
        } catch (err) {
          reject(err)
        }
      },
      fail: (err) => {
        if (showError) {
          uni.showToast({
            title: '网络开小差了，请稍后重试',
            icon: 'none',
            duration: 2000,
          })
        }
        reject(err)
      },
      complete: () => {
        if (loading) {
          uni.hideLoading()
        }
      },
    })
  })
}

/** 文件上传 */
function uploadFile<T>(
  url: string,
  filePath: string,
  formData?: Record<string, any>,
  options: RequestOptions = {},
): Promise<T> {
  const { loading = false, showError = true } = options
  const authStore = useAuthStore()

  if (loading) {
    uni.showLoading({ title: '上传中...', mask: true })
  }

  return new Promise((resolve, reject) => {
    uni.uploadFile({
      url: BASE_URL + url,
      filePath,
      name: 'file',
      formData,
      timeout: 30000,
      header: {
        ...(authStore.token ? { Authorization: authStore.authorizationHeader } : {}),
      },
      success: (res) => {
        try {
          const parsed = JSON.parse(res.data) as ApiResponse<T>
          const result = unwrapResponse<T>(parsed, showError)
          resolve(result)
        } catch (err) {
          reject(err)
        }
      },
      fail: (err) => {
        if (showError) {
          uni.showToast({
            title: '上传失败，请稍后重试',
            icon: 'none',
            duration: 2000,
          })
        }
        reject(err)
      },
      complete: () => {
        if (loading) {
          uni.hideLoading()
        }
      },
    })
  })
}

// 导出便捷方法
export function get<T>(url: string, params?: Record<string, any>, options?: RequestOptions): Promise<T> {
  return request<T>('GET', url, params, options)
}

export function post<T>(url: string, data?: Record<string, any>, options?: RequestOptions): Promise<T> {
  return request<T>('POST', url, data, options)
}

export function put<T>(url: string, data?: Record<string, any>, options?: RequestOptions): Promise<T> {
  return request<T>('PUT', url, data, options)
}

export function del<T>(url: string, data?: Record<string, any>, options?: RequestOptions): Promise<T> {
  return request<T>('DELETE', url, data, options)
}

export function upload<T>(url: string, filePath: string, formData?: Record<string, any>, options?: RequestOptions): Promise<T> {
  return uploadFile<T>(url, filePath, formData, options)
}
