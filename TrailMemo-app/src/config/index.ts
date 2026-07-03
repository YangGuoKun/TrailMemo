/**
 * TrailMemo 应用配置
 */

// 环境判断 — 小程序编译时由 Vite 替换，import.meta.env.MODE 比 process.env.NODE_ENV 更可靠
const isDev = import.meta.env.MODE === 'development'

/** 将后端返回的相对路径转为完整 URL */
export function getFullUrl(path: string | undefined | null): string {
  if (!path) return ''
  if (path.startsWith('http')) return path
  // 处理 Windows 反斜杠
  const normalized = path.replace(/\\/g, '/')
  const apiBase = isDev ? 'http://localhost:8087' : 'https://api.trailmemo.com'
  return apiBase + '/uploads/' + normalized
}

export const config = {
  // API 基础地址
  apiBaseUrl: isDev ? 'http://localhost:8087/api/v1' : 'https://api.trailmemo.com/api/v1',

  // 腾讯地图 Key
  mapKey: '', // TODO: 填入腾讯地图 Key

  // Token 存储 key
  tokenKey: 'TRAILMEMO_TOKEN',

  // 用户信息缓存 key
  userInfoKey: 'TRAILMEMO_USER',

  // 分页默认值
  defaultPageSize: 20,

  // 图片上传限制
  uploadMaxSize: 10 * 1024 * 1024, // 10MB

  // 心情标签选项
  moodOptions: [
    { label: '开心', value: 'happy', color: '#FF9F0A' },
    { label: '兴奋', value: 'excited', color: '#FF3B30' },
    { label: '平静', value: 'peaceful', color: '#0DA5BF' },
    { label: '疲惫', value: 'tired', color: '#8E8E93' },
    { label: '感动', value: 'moved', color: '#FF6B8A' },
  ],

  // 路线状态
  routeStatusMap: {
    draft: { label: '草稿', color: '#8E8E93' },
    in_progress: { label: '进行中', color: '#0DA5BF' },
    completed: { label: '已完成', color: '#34C759' },
    paused: { label: '已暂停', color: '#FF9F0A' },
  } as Record<string, { label: string; color: string }>,

  // 评分等级 (1-5)
  ratingLabels: ['', '一般', '不错', '很好', '很棒', '超赞'],
}
