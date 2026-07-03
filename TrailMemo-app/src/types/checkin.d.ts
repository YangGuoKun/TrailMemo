/**
 * 打卡模块类型定义
 */

/** 打卡记录 */
interface Checkin {
  id: number
  user_id: number
  route_id: number
  checkpoint_id: number
  checkin_time: string
  latitude: number
  longitude: number
  photo_url: string
  content: string
  rating: number // 1-5
  user?: UserProfile
  created_at: string
  updated_at: string
}

/** 创建打卡请求 */
interface CreateCheckinRequest {
  route_id: number
  checkpoint_id: number
  latitude?: number
  longitude?: number
  photo_url?: string
  content?: string
  rating?: number // 1-5
}

/** 更新打卡请求 */
interface UpdateCheckinRequest {
  photo_url?: string
  content?: string
  rating?: number
}

/** 路线打卡进度 */
interface RouteProgressData {
  progress: number // 0-100
}
