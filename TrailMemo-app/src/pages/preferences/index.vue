<template>
  <view class="preferences-page">
    <view class="navbar glass-surface">
      <view class="nav-back" @tap="goBack"><text class="back-icon">←</text></view>
      <text class="nav-title">AI 偏好设置</text>
      <text class="nav-action" @tap="handleSave">保存</text>
    </view>

    <scroll-view scroll-y class="page-body">
      <view class="section">
        <text class="section-title">旅行风格</text>
        <text class="section-desc">告诉 AI 你喜欢的旅行方式</text>
        <view class="chip-grid">
          <view
            v-for="style in travelStyles"
            :key="style"
            class="style-chip"
            :class="{ active: selectedStyles.includes(style) }"
            @tap="toggleStyle(style)"
          >
            <text>{{ style }}</text>
          </view>
        </view>
      </view>

      <view class="section">
        <text class="section-title">预算范围</text>
        <text class="section-desc">设置你的旅行预算偏好</text>
        <view class="range-card">
          <view class="range-row">
            <text class="range-label">最低预算</text>
            <text class="range-value">¥{{ budgetRange[0] }}</text>
          </view>
          <slider
            :value="budgetRange[0]"
            :min="0"
            :max="50000"
            :step="1000"
            activeColor="#A18CD1"
            backgroundColor="#E5E5EA"
            block-size="24"
            @change="(e:any) => budgetRange[0] = e.detail.value"
          />
          <view class="range-row">
            <text class="range-label">最高预算</text>
            <text class="range-value">¥{{ budgetRange[1] }}</text>
          </view>
          <slider
            :value="budgetRange[1]"
            :min="0"
            :max="100000"
            :step="1000"
            activeColor="#FBC2EB"
            backgroundColor="#E5E5EA"
            block-size="24"
            @change="(e:any) => budgetRange[1] = e.detail.value"
          />
        </view>
      </view>

      <view class="section">
        <text class="section-title">旅行天数</text>
        <text class="section-desc">你通常喜欢的旅行时长</text>
        <view class="range-card">
          <view class="range-row">
            <text class="range-label">最短天数</text>
            <text class="range-value">{{ daysRange[0] }}天</text>
          </view>
          <slider
            :value="daysRange[0]"
            :min="1"
            :max="30"
            :step="1"
            activeColor="#89D4CF"
            backgroundColor="#E5E5EA"
            block-size="24"
            @change="(e:any) => daysRange[0] = e.detail.value"
          />
          <view class="range-row">
            <text class="range-label">最长天数</text>
            <text class="range-value">{{ daysRange[1] }}天</text>
          </view>
          <slider
            :value="daysRange[1]"
            :min="1"
            :max="60"
            :step="1"
            activeColor="#0DA5BF"
            backgroundColor="#E5E5EA"
            block-size="24"
            @change="(e:any) => daysRange[1] = e.detail.value"
          />
        </view>
      </view>

      <view class="section">
        <text class="section-title">感兴趣的城市</text>
        <text class="section-desc">输入你想去的城市</text>
        <view class="input-group">
          <input
            v-model="newCity"
            placeholder="输入城市名"
            class="city-input"
            @confirm="addCity"
          />
          <button class="add-btn" @tap="addCity">添加</button>
        </view>
        <view class="city-chips">
          <view
            v-for="city in preferredCities"
            :key="city"
            class="city-chip"
            @tap="removeCity(city)"
          >
            <text>{{ city }}</text>
            <text class="remove-icon">×</text>
          </view>
        </view>
      </view>

      <view class="section">
        <text class="section-title">兴趣爱好</text>
        <text class="section-desc">选择你的旅行兴趣</text>
        <view class="chip-grid">
          <view
            v-for="interest in interestsList"
            :key="interest"
            class="style-chip"
            :class="{ active: selectedInterests.includes(interest) }"
            @tap="toggleInterest(interest)"
          >
            <text>{{ interest }}</text>
          </view>
        </view>
      </view>

      <view class="section">
        <text class="section-title">避坑清单</text>
        <text class="section-desc">不想要的旅行体验</text>
        <view class="chip-grid">
          <view
            v-for="item in avoidList"
            :key="item"
            class="style-chip avoid"
            :class="{ active: selectedAvoid.includes(item) }"
            @tap="toggleAvoid(item)"
          >
            <text>{{ item }}</text>
          </view>
        </view>
      </view>

      <view style="height: 40rpx" />
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as agentApi from '@/api/agent'

const travelStyles = ['休闲放松', '美食探店', '自然风光', '文化古迹', '冒险挑战', '亲子游', '情侣浪漫', '摄影打卡']
const interestsList = ['美食', '购物', '户外', '历史', '艺术', '音乐', '运动', '夜生活']
const avoidList = ['人太多', '价格贵', '交通不便', '环境脏乱', '商业化严重']

const selectedStyles = ref<string[]>([])
const budgetRange = ref<[number, number]>([1000, 10000])
const daysRange = ref<[number, number]>([3, 7])
const preferredCities = ref<string[]>([])
const selectedInterests = ref<string[]>([])
const selectedAvoid = ref<string[]>([])
const newCity = ref('')

onMounted(async () => {
  try {
    const pref = await agentApi.getPreferences()
    selectedStyles.value = pref.travel_styles || []
    budgetRange.value = pref.budget_range || [1000, 10000]
    daysRange.value = pref.preferred_days || [3, 7]
    preferredCities.value = pref.preferred_cities || []
    selectedInterests.value = pref.interests || []
    selectedAvoid.value = pref.avoid_list || []
  } catch (_) {}
})

function toggleStyle(style: string) {
  const idx = selectedStyles.value.indexOf(style)
  if (idx > -1) {
    selectedStyles.value.splice(idx, 1)
  } else {
    selectedStyles.value.push(style)
  }
}

function toggleInterest(interest: string) {
  const idx = selectedInterests.value.indexOf(interest)
  if (idx > -1) {
    selectedInterests.value.splice(idx, 1)
  } else {
    selectedInterests.value.push(interest)
  }
}

function toggleAvoid(item: string) {
  const idx = selectedAvoid.value.indexOf(item)
  if (idx > -1) {
    selectedAvoid.value.splice(idx, 1)
  } else {
    selectedAvoid.value.push(item)
  }
}

function addCity() {
  const city = newCity.value.trim()
  if (city && !preferredCities.value.includes(city)) {
    preferredCities.value.push(city)
    newCity.value = ''
  }
}

function removeCity(city: string) {
  const idx = preferredCities.value.indexOf(city)
  if (idx > -1) {
    preferredCities.value.splice(idx, 1)
  }
}

async function handleSave() {
  uni.showLoading({ title: '保存中...' })
  try {
    await agentApi.updatePreferences({
      travel_styles: selectedStyles.value,
      budget_range: budgetRange.value,
      preferred_days: daysRange.value,
      preferred_cities: preferredCities.value,
      interests: selectedInterests.value,
      avoid_list: selectedAvoid.value,
    })
    uni.hideLoading()
    uni.showToast({ title: '保存成功', icon: 'success' })
  } catch (err) {
    uni.hideLoading()
    uni.showToast({ title: '保存失败', icon: 'none' })
  }
}

function goBack() {
  uni.navigateBack()
}
</script>

<style lang="scss" scoped>
.preferences-page { min-height: 100vh; background: linear-gradient(170deg, #FDF8F4, #F7F4F0, #F3F1F5, #F5F3F8); }

.navbar {
  position: sticky; top: 0; z-index: 50;
  display: flex; align-items: center; justify-content: center;
  height: 88rpx; padding: 0 32rpx;
  background: rgba(255,255,255,0.72);
  backdrop-filter: saturate(170%) blur(24px);
  -webkit-backdrop-filter: saturate(170%) blur(24px);
  border-bottom: 1px solid rgba(255,255,255,0.4);
  .nav-back { position: absolute; left: 16rpx; width: 64rpx; height: 64rpx; border-radius: 50%; background: rgba(0,0,0,0.04); display: flex; align-items: center; justify-content: center; }
  .back-icon { font-size: 32rpx; color: #1C1C1E; }
  .nav-title { font-size: 34rpx; font-weight: 600; color: #1C1C1E; }
  .nav-action { position: absolute; right: 32rpx; font-size: 28rpx; color: #0DA5BF; font-weight: 600; }
}

.page-body { padding: 24rpx 32rpx; }

.section { margin-bottom: 36rpx; }

.section-title { font-size: 32rpx; font-weight: 700; color: #1C1C1E; display: block; }

.section-desc { font-size: 24rpx; color: #8E8E93; margin-top: 8rpx; display: block; }

.chip-grid { display: flex; flex-wrap: wrap; gap: 16rpx; margin-top: 20rpx; }

.style-chip {
  padding: 16rpx 28rpx;
  border-radius: 9999rpx;
  background: rgba(255,255,255,0.6);
  border: 1px solid rgba(0,0,0,0.08);
  font-size: 26rpx;
  color: #636366;
  transition: all 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
  &.active {
    background: linear-gradient(135deg, rgba(161,140,209,0.2), rgba(251,194,235,0.2));
    border-color: #A18CD1;
    color: #A18CD1;
    font-weight: 600;
  }
  &.avoid.active {
    background: rgba(255,59,48,0.1);
    border-color: rgba(255,59,48,0.3);
    color: #FF3B30;
  }
  &:active { transform: scale(0.95); }
}

.range-card {
  margin-top: 20rpx;
  padding: 24rpx;
  background: rgba(255,255,255,0.6);
  border-radius: 20rpx;
}

.range-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
  &:last-child { margin-bottom: 0; }
}

.range-label { font-size: 26rpx; color: #636366; }

.range-value { font-size: 28rpx; font-weight: 600; color: #1C1C1E; }

.input-group {
  display: flex;
  gap: 16rpx;
  margin-top: 20rpx;
}

.city-input {
  flex: 1;
  height: 80rpx;
  padding: 0 24rpx;
  background: rgba(255,255,255,0.6);
  border-radius: 16rpx;
  border: 1px solid rgba(0,0,0,0.08);
  font-size: 28rpx;
}

.add-btn {
  height: 80rpx;
  padding: 0 32rpx;
  background: linear-gradient(135deg, #A18CD1, #FBC2EB);
  border-radius: 16rpx;
  color: #fff;
  font-size: 28rpx;
  font-weight: 600;
  border: none;
  &:active { opacity: 0.8; }
}

.city-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  margin-top: 16rpx;
}

.city-chip {
  display: flex;
  align-items: center;
  gap: 8rpx;
  padding: 12rpx 20rpx;
  background: rgba(137, 212, 207, 0.15);
  border-radius: 9999rpx;
  font-size: 24rpx;
  color: #0DA5BF;
  &:active { opacity: 0.7; }
  .remove-icon { font-size: 28rpx; color: #8E8E93; }
}
</style>