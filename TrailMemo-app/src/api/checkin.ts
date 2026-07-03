/**
 * 打卡模块 API
 */

import { get, post, put, del } from './request'

interface CheckinListParams extends PaginationParams {
  route_id?: number
}

/** 获取打卡列表 (可按路线筛选) */
export function getCheckinList(params: CheckinListParams): Promise<PaginatedData<Checkin>> {
  const query: Record<string, number> = {}
  if (typeof params.page === 'number') {
    query.page = params.page
  }
  if (typeof params.size === 'number') {
    query.size = params.size
  }
  if (typeof params.route_id === 'number') {
    query.route_id = params.route_id
  }
  return get<PaginatedData<Checkin>>('/checkins', query)
}

/** 获取打卡详情 */
export function getCheckinDetail(id: number): Promise<Checkin> {
  return get<Checkin>(`/checkins/${id}`)
}

/** 创建打卡 */
export function createCheckin(data: CreateCheckinRequest): Promise<Checkin> {
  return post<Checkin>('/checkins', data as any)
}

/** 更新打卡 */
export function updateCheckin(id: number, data: UpdateCheckinRequest): Promise<void> {
  return put<void>(`/checkins/${id}`, data as any)
}

/** 删除打卡 */
export function deleteCheckin(id: number): Promise<void> {
  return del<void>(`/checkins/${id}`)
}

/** 获取路线打卡进度 */
export function getRouteProgress(routeId: number): Promise<RouteProgressData> {
  return get<RouteProgressData>(`/checkins/progress/${routeId}`)
}
