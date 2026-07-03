/**
 * 用户模块 API
 */

import { get, post, put } from './request'

/** 注册 */
export function register(data: RegisterRequest): Promise<RegisterResponse> {
  return post<RegisterResponse>('/users/register', data as any)
}

/** 密码登录 */
export function login(data: LoginRequest): Promise<LoginResponse> {
  return post<LoginResponse>('/users/login', data as any)
}

/** 微信登录 */
export function loginByWechat(code: string): Promise<LoginResponse> {
  return post<LoginResponse>('/users/login/wechat', { code })
}

/** 获取当前用户信息 */
export function getUserInfo(): Promise<UserProfile> {
  return get<UserProfile>('/users/me')
}

/** 更新用户信息 */
export function updateUserInfo(data: UpdateUserRequest): Promise<void> {
  return put<void>('/users/me', data as any)
}

/** 修改密码 */
export function changePassword(data: ChangePasswordRequest): Promise<void> {
  return put<void>('/users/me/password', data as any)
}
