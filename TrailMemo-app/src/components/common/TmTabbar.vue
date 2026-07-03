<template>
  <view class="custom-tabbar">
    <view
      v-for="tab in tabs"
      :key="tab.key"
      class="tab-item"
      :class="{ active: current === tab.key }"
      @tap="switchTab(tab.key)"
    >
      <text class="tab-icon">{{ tab.icon }}</text>
      <text class="tab-text">{{ tab.text }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{ current: string }>()
const emit = defineEmits<{ (e: 'change', key: string): void }>()

const tabs = [
  { key: 'discover', icon: '🧭', text: '发现' },
  { key: 'routes', icon: '📍', text: '路线' },
  { key: 'profile', icon: '🧑🏽', text: '我的' },
]

function switchTab(key: string) {
  if (key !== props.current) emit('change', key)
}
</script>

<style lang="scss" scoped>
.custom-tabbar {
  position: fixed; bottom: 0; left: 0; right: 0; z-index: 999;
  display: flex; height: calc(100rpx + env(safe-area-inset-bottom, 0));
  padding: 4rpx 0 env(safe-area-inset-bottom, 0);
  background: rgba(255,255,255,0.82);
  backdrop-filter: saturate(170%) blur(24px);
  -webkit-backdrop-filter: saturate(170%) blur(24px);
  border-top: 1px solid rgba(0,0,0,0.06);
}

.tab-item {
  flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 2rpx;
  font-size: 20rpx; color: #8E8E93; transition: all 0.2s cubic-bezier(0.34,1.56,0.64,1);
  &.active { color: #0DA5BF; }
  .tab-icon { font-size: 40rpx; transition: transform 0.25s cubic-bezier(0.34,1.56,0.64,1); }
  &:active .tab-icon { transform: scale(0.82); }
}
</style>
