/**
 * 文件上传 API
 */

import { upload } from './request'

/** 上传头像 */
export function uploadAvatar(filePath: string): Promise<AvatarUploadResponse> {
  return upload<AvatarUploadResponse>('/upload/avatar', filePath)
}
