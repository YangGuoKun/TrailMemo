/**
 * 通用类型定义
 */

/** API 原始响应信封 */
interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

/** 分页响应结构 */
interface PaginatedData<T> {
  list: T[]
  total: number
  page: number
  size: number
}

/** 分页参数 */
interface PaginationParams {
  page?: number
  size?: number
}

/** 心情标签 */
type MoodType = 'happy' | 'excited' | 'peaceful' | 'tired' | 'moved'

/** 路线状态 */
type RouteStatus = 'draft' | 'in_progress' | 'completed' | 'paused'

/** 点赞目标类型 */
type LikeTargetType = 'route' | 'checkin' | 'post' | 'comment'

/** 性别 */
type Gender = 0 | 1 | 2 // 0=未知 1=男 2=女
