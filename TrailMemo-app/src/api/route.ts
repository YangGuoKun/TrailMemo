/**
 * 路线模块 API
 */

import { get, post, put, del } from './request'

/** 获取路线列表 */
export function getRouteList(page: number = 1, size: number = 20): Promise<PaginatedData<Route>> {
  return get<PaginatedData<Route>>('/routes', { page, size })
}

/** 获取路线详情 (含打卡点) */
export function getRouteDetail(id: number): Promise<RouteDetail> {
  return get<RouteDetail>(`/routes/${id}`)
}

/** 创建路线 */
export function createRoute(data: CreateRouteRequest): Promise<Route> {
  return post<Route>('/routes', data as any)
}

/** 更新路线 */
export function updateRoute(id: number, data: UpdateRouteRequest): Promise<void> {
  return put<void>(`/routes/${id}`, data as any)
}

/** 删除路线 */
export function deleteRoute(id: number): Promise<void> {
  return del<void>(`/routes/${id}`)
}

/** 复制/复用路线 */
export function copyRoute(id: number, isPublic: number = 1): Promise<Route> {
  return post<Route>(`/routes/${id}/copy`, { isPublic })
}
