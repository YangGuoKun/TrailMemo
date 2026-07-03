/**
 * 图片选择与上传 Hook
 */

import { config } from '@/config'

export function useImagePicker() {
  // 从相册选择图片
  function chooseImage(count: number = 1): Promise<string[]> {
    return new Promise((resolve, reject) => {
      uni.chooseImage({
        count,
        sizeType: ['compressed'],
        sourceType: ['album', 'camera'],
        success: (res) => resolve(res.tempFilePaths),
        fail: reject,
      })
    })
  }

  // 拍照
  function takePhoto(): Promise<string> {
    return new Promise((resolve, reject) => {
      uni.chooseImage({
        count: 1,
        sizeType: ['compressed'],
        sourceType: ['camera'],
        success: (res) => resolve(res.tempFilePaths[0]),
        fail: reject,
      })
    })
  }

  // 选择或拍照 (底部操作菜单)
  function pickImage(): Promise<string> {
    return new Promise((resolve, reject) => {
      uni.showActionSheet({
        itemList: ['拍照', '从相册选择'],
        success: (actionRes) => {
          const sourceType = actionRes.tapIndex === 0 ? ['camera'] : ['album']
          uni.chooseImage({
            count: 1,
            sizeType: ['compressed'],
            sourceType: sourceType as any,
            success: (imgRes) => resolve(imgRes.tempFilePaths[0]),
            fail: reject,
          })
        },
        fail: reject,
      })
    })
  }

  // 预览图片
  function previewImage(urls: string[], current: number = 0) {
    uni.previewImage({
      urls,
      current,
    })
  }

  return {
    chooseImage,
    takePhoto,
    pickImage,
    previewImage,
  }
}
