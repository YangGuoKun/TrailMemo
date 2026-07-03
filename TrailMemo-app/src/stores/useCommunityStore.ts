/** 社区状态管理 */

import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as communityApi from '@/api/community'
import { config } from '@/config'

export const useCommunityStore = defineStore('community', () => {
  // Feed
  const posts = ref<Post[]>([])
  const total = ref(0)
  const page = ref(1)
  const isListLoading = ref(false)

  // 详情
  const currentPost = ref<Post | null>(null)
  const comments = ref<Comment[]>([])
  const isDetailLoading = ref(false)

  // 点赞/收藏状态缓存
  const likeStatus = ref<Record<string, boolean>>({})
  const favStatus = ref<Record<number, boolean>>({})

  // 获取社区 Feed
  async function fetchPosts(p: number = 1) {
    isListLoading.value = true
    try {
      const res = await communityApi.getPostList({ page: p, size: config.defaultPageSize })
      if (p === 1) { posts.value = res.list } else { posts.value.push(...res.list) }
      total.value = res.total; page.value = p
    } finally { isListLoading.value = false }
  }

  // 获取帖子详情
  async function fetchPostDetail(id: number) {
    isDetailLoading.value = true
    try {
      currentPost.value = await communityApi.getPostDetail(id)
      const cmtRes = await communityApi.getCommentList(id, 1, 100)
      comments.value = cmtRes.list
      // 加载点赞状态
      const likeKey = `post_${id}`
      try { const s = await communityApi.getLikeStatus(id, 'post'); likeStatus.value[likeKey] = s.liked } catch (_) {}
    } finally { isDetailLoading.value = false }
  }

  // 创建帖子
  async function createPost(data: CreatePostRequest): Promise<Post> {
    const post = await communityApi.createPost(data)
    await fetchPosts(1)
    return post
  }

  // 删除帖子
  async function deletePost(id: number) {
    await communityApi.deletePost(id)
    posts.value = posts.value.filter(p => p.id !== id); total.value--
    if (currentPost.value?.id === id) currentPost.value = null
  }

  // 添加评论
  async function addComment(postId: number, content: string, parentId?: number) {
    await communityApi.createComment({ post_id: postId, content, parent_id: parentId || 0 })
    const cmtRes = await communityApi.getCommentList(postId, 1, 100)
    comments.value = cmtRes.list
    if (currentPost.value) currentPost.value.comment_count = cmtRes.total
  }

  // 删除评论
  async function removeComment(commentId: number) {
    await communityApi.deleteComment(commentId)
    comments.value = comments.value.filter(c => c.id !== commentId)
  }

  // 点赞切换
  async function togglePostLike(postId: number): Promise<boolean> {
    const res = await communityApi.toggleLike({ target_id: postId, target_type: 'post' })
    likeStatus.value[`post_${postId}`] = res.liked
    const p = posts.value.find(x => x.id === postId)
    if (p) p.like_count += res.liked ? 1 : -1
    if (currentPost.value?.id === postId) currentPost.value.like_count += res.liked ? 1 : -1
    return res.liked
  }

  // 收藏切换
  async function toggleRouteFav(routeId: number): Promise<boolean> {
    const res = await communityApi.toggleFavorite(routeId)
    favStatus.value[routeId] = res.favorited
    return res.favorited
  }

  // 获取图片数组
  function getPostImages(post: Post): string[] {
    if (!post.images) return []
    return post.images.split(',').filter(Boolean)
  }

  // 清除状态
  function clearState() {
    posts.value = []; total.value = 0; page.value = 1
    currentPost.value = null; comments.value = []
  }

  return {
    posts, total, page, isListLoading, currentPost, comments, isDetailLoading,
    likeStatus, favStatus,
    fetchPosts, fetchPostDetail, createPost, deletePost,
    addComment, removeComment, togglePostLike, toggleRouteFav,
    getPostImages, clearState,
  }
})
