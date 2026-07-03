/** 社区模块 API */

import { get, post, put, del } from './request'

// ── 帖子 ────────────────────────────────────

export function createPost(data: CreatePostRequest): Promise<Post> {
  return post<Post>('/posts', data as any)
}

export function getPostList(params?: { page?: number; size?: number; user_id?: number; route_id?: number }): Promise<PaginatedData<Post>> {
  return get<PaginatedData<Post>>('/posts', params as any)
}

export function getPostDetail(id: number): Promise<Post> {
  return get<Post>(`/posts/${id}`)
}

export function updatePost(id: number, data: { title?: string; content?: string; images?: string }): Promise<void> {
  return put<void>(`/posts/${id}`, data as any)
}

export function deletePost(id: number): Promise<void> {
  return del<void>(`/posts/${id}`)
}

// ── 评论 ────────────────────────────────────

export function createComment(data: CreateCommentRequest): Promise<Comment> {
  return post<Comment>('/comments', data as any)
}

export function getCommentList(postId: number, page?: number, size?: number): Promise<PaginatedData<Comment>> {
  return get<PaginatedData<Comment>>('/comments', { post_id: postId, page, size })
}

export function deleteComment(id: number): Promise<void> {
  return del<void>(`/comments/${id}`)
}

// ── 点赞 ────────────────────────────────────

export function toggleLike(data: ToggleLikeRequest): Promise<{ liked: boolean }> {
  return post<{ liked: boolean }>('/likes/toggle', data as any)
}

export function getLikeStatus(targetId: number, targetType: string): Promise<{ liked: boolean }> {
  return get<{ liked: boolean }>('/likes/status', { target_id: targetId, target_type: targetType })
}

export function getLikeCount(targetId: number, targetType: string): Promise<{ count: number }> {
  return get<{ count: number }>('/likes/count', { target_id: targetId, target_type: targetType })
}

// ── 收藏 ────────────────────────────────────

export function toggleFavorite(routeId: number): Promise<{ favorited: boolean }> {
  return post<{ favorited: boolean }>('/favorites/toggle', { route_id: routeId })
}

export function getFavoriteStatus(routeId: number): Promise<{ favorited: boolean }> {
  return get<{ favorited: boolean }>('/favorites/status', { route_id: routeId })
}

export function getUserFavorites(page?: number, size?: number): Promise<PaginatedData<any>> {
  return get<PaginatedData<any>>('/favorites/list', { page, size })
}
