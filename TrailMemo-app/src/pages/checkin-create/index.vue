<template>
  <view class="checkin-page">
    <view class="navbar">
      <view class="nav-back" @tap="goBack"><u-icon name="arrow-left" size="40" color="#1C1C1E" /></view>
      <text class="nav-title">发布打卡</text>
      <text class="nav-save" :class="{ disabled: submitting }" @tap="handleSubmit">发布</text>
    </view>

    <scroll-view scroll-y class="page-body">
      <!-- 打卡点信息 -->
      <view class="glass-card cp-info-card">
        <u-icon name="map-pin-fill" size="36" color="#0DA5BF" />
        <view>
          <text class="cp-name">{{ checkpointName }}</text>
          <text class="cp-hint">记录此刻的美好</text>
        </view>
      </view>

      <!-- 照片上传 -->
      <view class="glass-card">
        <text class="sec-label">打卡照片</text>
        <view class="photo-area" @tap="pickPhoto" v-if="!photoUrl">
          <u-icon name="camera-fill" size="64" color="#C7C7CC" />
          <text class="photo-hint">点击拍照或选择照片</text>
        </view>
        <view v-else class="photo-preview">
          <image :src="photoUrl" mode="aspectFill" class="preview-img" @tap="previewPhoto" />
          <view class="photo-actions">
            <text class="act-link" @tap="pickPhoto">重新选择</text>
            <text class="act-link del" @tap="removePhoto">删除</text>
          </view>
        </view>
      </view>

      <!-- 感受记录 -->
      <view class="glass-card">
        <text class="sec-label">感受记录</text>
        <textarea v-model="content" class="content-input" placeholder="记录此刻的感受..." :maxlength="500" :auto-height="true" />
        <text class="char-count">{{ content.length }}/500</text>
      </view>

      <!-- 心情（多选） -->
      <view class="glass-card">
        <text class="sec-label">心情（可多选）</text>
        <view class="mood-row">
          <view v-for="m in moodOptions" :key="m.value" class="mood-tag" :class="[m.cls, { sel: selectedMoods.includes(m.value) }]" @tap="toggleMood(m.value)">
            <text>{{ m.emoji }} {{ m.label }}</text>
          </view>
        </view>
      </view>

      <!-- 评分 -->
      <view class="glass-card rating-card">
        <text class="sec-label">评分</text>
        <view class="stars-row">
          <text v-for="i in 5" :key="i" class="star" :class="{ on: i <= rating, off: i > rating }" @tap="setRating(i)">★</text>
        </view>
        <text class="rating-label" v-if="rating > 0">{{ ratingLabels[rating] }}</text>
      </view>

      <!-- 位置 -->
      <view class="glass-card">
        <view class="loc-row">
          <text class="sec-label">打卡位置</text>
          <text class="refresh-loc" @tap="getLocation">重新定位</text>
        </view>
        <view class="loc-display">
          <u-icon name="map-pin" size="32" color="#0DA5BF" />
          <text v-if="currentLat">{{ currentLat.toFixed(6) }}, {{ currentLng.toFixed(6) }}</text>
          <text v-else class="loc-pending">正在获取位置...</text>
        </view>
      </view>

      <view style="height: 200rpx" />
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useCheckinStore } from '@/stores/useCheckinStore'
import { useImagePicker } from '@/composables/useImagePicker'
import { useLocation } from '@/composables/useLocation'

const checkinStore = useCheckinStore()
const { pickImage, previewImage } = useImagePicker()
const { currentLat, currentLng, getCurrentLocation } = useLocation()

const routeId = ref(0)
const checkpointId = ref(0)
const checkpointName = ref('')
const photoUrl = ref('')
const content = ref('')
const rating = ref(0)
const selectedMoods = ref<string[]>([])
const submitting = ref(false)

const moodOptions = [
  { label: '开心', value: 'happy', emoji: '😊', cls: 'm-happy' },
  { label: '兴奋', value: 'excited', emoji: '🤩', cls: 'm-excited' },
  { label: '平静', value: 'peace', emoji: '😌', cls: 'm-peace' },
  { label: '感动', value: 'moved', emoji: '🥹', cls: 'm-moved' },
]
const ratingLabels = ['', '很差', '一般', '好', '很好', '超棒']

onLoad((options: any) => {
  routeId.value = Number(options.routeId) || 0
  checkpointId.value = Number(options.checkpointId) || 0
  checkpointName.value = decodeURIComponent(options.name || '打卡点')
})
onMounted(() => { getCurrentLocation().catch(() => {}) })

async function pickPhoto() {
  try { photoUrl.value = await pickImage() } catch (_) {}
}
function previewPhoto() { previewImage([photoUrl.value], 0) }
function removePhoto() { photoUrl.value = '' }
function setRating(v: number) { rating.value = v }
function toggleMood(m: string) {
  const i = selectedMoods.value.indexOf(m)
  if (i >= 0) selectedMoods.value.splice(i, 1)
  else selectedMoods.value.push(m)
}
function getLocation() { getCurrentLocation().catch(() => uni.showToast({ title: '获取位置失败', icon: 'none' })) }

async function handleSubmit() {
  if (submitting.value) return
  if (!photoUrl.value) { uni.showToast({ title: '请上传打卡照片', icon: 'none' }); return }
  submitting.value = true
  try {
    await checkinStore.createCheckin({
      route_id: routeId.value,
      checkpoint_id: checkpointId.value,
      latitude: currentLat.value || undefined,
      longitude: currentLng.value || undefined,
      photo_url: photoUrl.value,
      content: content.value,
      rating: rating.value || undefined,
    })
    uni.showToast({ title: '打卡成功！🎉', icon: 'success' })
    setTimeout(() => uni.navigateBack(), 800)
  } catch (_) {} finally { submitting.value = false }
}
function goBack() { uni.navigateBack() }
</script>

<style lang="scss" scoped>
.checkin-page { min-height: 100vh; background: linear-gradient(170deg, #FDF8F4, #F7F4F0, #F3F1F5, #F5F3F8); }

.navbar {
  position: sticky; top: 0; z-index: 50;
  display: flex; align-items: center; justify-content: center;
  height: 88rpx; padding: 0 32rpx;
  background: rgba(255,255,255,0.72);
  backdrop-filter: saturate(170%) blur(24px);
  -webkit-backdrop-filter: saturate(170%) blur(24px);
  border-bottom: 1px solid rgba(255,255,255,0.4);
  .nav-back { position: absolute; left: 16rpx; width: 64rpx; height: 64rpx;
    display: flex; align-items: center; justify-content: center;
    border-radius: 50%; background: rgba(0,0,0,0.04); }
  .nav-title { font-size: 34rpx; font-weight: 600; color: #1C1C1E; }
  .nav-save { position: absolute; right: 32rpx; font-size: 28rpx; font-weight: 600; color: #0DA5BF;
    &.disabled { opacity: 0.4; } }
}

.page-body { padding: 16rpx 28rpx 40rpx; }

.glass-card { margin-bottom: 20rpx; padding: 28rpx;
  background: rgba(255,255,255,0.55);
  border-radius: 24rpx; box-shadow: 0 8px 32px rgba(0,0,0,0.06), 0 2px 8px rgba(0,0,0,0.03); }

.sec-label { font-size: 30rpx; font-weight: 600; color: #1C1C1E; display: block; margin-bottom: 16rpx; }

.cp-info-card { display: flex; align-items: center; gap: 16rpx;
  .cp-name { font-size: 32rpx; font-weight: 600; color: #1C1C1E; display: block; }
  .cp-hint { font-size: 24rpx; color: #AEAEB2; }
}

.photo-area { display: flex; flex-direction: column; align-items: center; justify-content: center;
  height: 320rpx; background: rgba(0,0,0,0.02); border: 2px dashed rgba(0,0,0,0.08);
  border-radius: 16rpx; gap: 12rpx;
  .photo-hint { font-size: 24rpx; color: #AEAEB2; }
}
.photo-preview { .preview-img { width: 100%; height: 360rpx; border-radius: 16rpx; object-fit: cover; }
  .photo-actions { display: flex; gap: 40rpx; justify-content: center; padding-top: 16rpx;
    .act-link { font-size: 26rpx; color: #007AFF; &.del { color: #FF3B30; } } }
}

.content-input { width: 100%; min-height: 140rpx; font-size: 28rpx; color: #1C1C1E; line-height: 1.5; }
.char-count { text-align: right; font-size: 22rpx; color: #AEAEB2; margin-top: 8rpx; }

.mood-row { display: flex; gap: 12rpx; flex-wrap: wrap; }
.mood-tag { padding: 12rpx 24rpx; border-radius: 9999rpx; font-size: 24rpx; font-weight: 600;
  border: 1px solid rgba(255,255,255,0.4); transition: all 0.2s;
  &.m-happy { background: rgba(255,159,10,0.12); color: #B86800; }
  &.m-excited { background: rgba(255,59,48,0.1); color: #C92016; }
  &.m-peace { background: rgba(13,165,191,0.1); color: #076F7E; }
  &.m-moved { background: rgba(255,107,138,0.1); color: #C93A58; }
  &.sel { transform: scale(1.08); box-shadow: 0 4px 16px rgba(0,0,0,0.1); }
}

.rating-card { text-align: center; }
.stars-row { display: flex; justify-content: center; gap: 8rpx; }
.star { font-size: 52rpx; cursor: pointer; transition: all 0.15s;
  &.on { color: #FF9F0A; } &.off { color: #E5E5EA; }
  &:active { transform: scale(1.2); }
}
.rating-label { font-size: 24rpx; color: #636366; margin-top: 8rpx; display: block; }

.loc-row { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8rpx;
  .sec-label { margin-bottom: 0; }
  .refresh-loc { font-size: 24rpx; color: #007AFF; }
}
.loc-display { display: flex; align-items: center; gap: 8rpx; padding: 16rpx;
  background: rgba(0,0,0,0.02); border-radius: 12rpx; font-size: 24rpx; color: #636366; }
.loc-pending { color: #AEAEB2; }
</style>
