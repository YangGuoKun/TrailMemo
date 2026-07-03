/**
 * 路线 & 打卡点模块类型定义
 */

/** 路线 (snake_case 来自 GORM) */
interface Route {
  id: number
  user_id: number
  title: string
  description: string
  cover_image: string
  start_city: string
  end_city: string
  total_distance: number
  estimated_hours: number
  publish_status: number
  view_count: number
  like_count: number
  favorite_count: number
  share_count: number
  reuse_count: number
  is_public: number
  user?: UserProfile
  created_at: string
  updated_at: string
}

/** 路线详情 (含打卡点列表) */
interface RouteDetail extends Route {
  checkpoints: Checkpoint[]
}

/** 打卡点 */
interface Checkpoint {
  id: number
  route_id: number
  name: string
  description: string
  latitude: number
  longitude: number
  address: string
  city: string
  sequence: number
  arrive_time: string
  stay_duration: number
  photo_url: string
}

/** 创建路线请求 (camelCase) */
interface CreateRouteRequest {
  title: string
  description?: string
  coverImage?: string
  startCity: string
  endCity: string
  totalDistance?: number
  estimatedHours?: number
  isPublic?: number
  checkpoints?: CheckpointInput[]
}

/** 打卡点输入 */
interface CheckpointInput {
  name: string
  description?: string
  latitude?: number
  longitude?: number
  address?: string
  city?: string
  sequence: number
  arriveTime?: string
  stayDuration?: number
  photoURL?: string
}

/** 更新路线请求 */
interface UpdateRouteRequest {
  title?: string
  description?: string
  coverImage?: string
  startCity?: string
  endCity?: string
  totalDistance?: number
  estimatedHours?: number
  isPublic?: number
}

/** 复制路线请求 */
interface CopyRouteRequest {
  isPublic?: number
}
