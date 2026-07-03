/** 社区模块类型定义 */

/** 分享帖子 (snake_case from GORM) */
interface Post {
  id: number
  user_id: number
  route_id: number
  title: string
  content: string
  images: string           // 逗号分隔的图片URL
  view_count: number
  like_count: number
  comment_count: number
  share_count: number
  reuse_count: number
  status: number
  user?: { id: number; username: string; nickname: string; avatar: string }
  created_at: string
  updated_at: string
}

/** 评论 */
interface Comment {
  id: number
  user_id: number
  post_id: number
  parent_id: number
  content: string
  like_count: number
  user?: { id: number; username: string; nickname: string; avatar: string }
  created_at: string
  updated_at: string
}

/** 创建帖子 */
interface CreatePostRequest {
  route_id: number
  title: string
  content: string
  images?: string
}

/** 创建评论 */
interface CreateCommentRequest {
  post_id: number
  parent_id?: number
  content: string
}

/** 点赞切换 */
interface ToggleLikeRequest {
  target_id: number
  target_type: 'post' | 'comment'
}

/** 收藏切换 */
interface ToggleFavoriteRequest {
  route_id: number
}
