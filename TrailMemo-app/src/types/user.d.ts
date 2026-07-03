/**
 * 用户模块类型定义
 */

/** 用户信息 (GET /users/me 返回，camelCase) */
interface UserProfile {
  id: number
  username: string
  nickname: string
  avatar: string
  phone: string
  email: string
  createdAt: string
}

/** 注册请求 */
interface RegisterRequest {
  username: string
  password: string
  phone?: string
  email?: string
}

/** 注册响应 */
interface RegisterResponse {
  id: number
  username: string
  nickname: string
}

/** 密码登录请求 */
interface LoginRequest {
  username: string
  password: string
}

/** 登录响应 */
interface LoginResponse {
  token: string
}

/** 微信登录请求 */
interface WechatLoginRequest {
  code: string
}

/** 更新用户信息请求 */
interface UpdateUserRequest {
  nickname?: string
  avatar?: string
}

/** 修改密码请求 */
interface ChangePasswordRequest {
  oldPassword: string
  newPassword: string
}

/** 头像上传响应 */
interface AvatarUploadResponse {
  url: string
}
