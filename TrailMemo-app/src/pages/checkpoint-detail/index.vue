<template>
  <view class="cp-detail-page">
    <!-- 导航栏 -->
    <view class="navbar">
      <view class="navbar-back" @tap="goBack">
        <u-icon name="arrow-left" size="40" color="#1C1C1E" />
      </view>
      <text class="navbar-title">打卡点详情</text>
    </view>

    <scroll-view class="page-body" scroll-y>
      <!-- 缩略地图 -->
      <view class="mini-map">
        <map
          id="cp-map"
          class="cp-map"
          :latitude="cp?.latitude || 39.9"
          :longitude="cp?.longitude || 116.4"
          :markers="mapMarkers"
          :scale="15"
        />
      </view>

      <!-- 打卡点信息 -->
      <view class="info-section card">
        <text class="cp-name">{{ cp?.name || '加载中...' }}</text>
        <text class="cp-addr" v-if="cp?.address">
          <u-icon name="map-pin" size="28" color="#8E8E93" />
          {{ cp.address }}
        </text>
        <text class="cp-desc" v-if="cp?.description">{{ cp.description }}</text>

        <view class="cp-meta" v-if="cp">
          <view class="meta-item" v-if="cp.arrive_time">
            <u-icon name="clock" size="28" color="#AEAEB2" />
            <text>预计到达: {{ cp.arrive_time }}</text>
          </view>
          <view class="meta-item" v-if="cp.stay_duration">
            <u-icon name="hourglass" size="28" color="#AEAEB2" />
            <text>建议停留: {{ cp.stay_duration }} 分钟</text>
          </view>
        </view>
      </view>

      <!-- 坐标信息 -->
      <view class="coord-section card" v-if="cp">
        <text class="section-label">坐标</text>
        <view class="coord-row">
          <text class="coord-text">经度: {{ cp.longitude }}</text>
          <text class="coord-text">纬度: {{ cp.latitude }}</text>
        </view>
      </view>

      <!-- 打卡按钮 -->
      <view class="action-section">
        <button
          v-if="!isChecked"
          class="checkin-btn"
          @tap="goCheckin"
        >
          <u-icon name="checkmark-circle" size="40" color="#FFFFFF" />
          <text class="btn-text">立即打卡</text>
        </button>
        <view v-else class="checked-badge">
          <u-icon name="checkmark-circle-fill" size="56" color="#34C759" />
          <text class="checked-text">已打卡</text>
        </view>
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useRouteStore } from '@/stores/useRouteStore'
import { useCheckinStore } from '@/stores/useCheckinStore'

const routeStore = useRouteStore()
const checkinStore = useCheckinStore()

const cpId = ref(0)
const routeId = ref(0)
const cp = ref<Checkpoint | null>(null)
const isChecked = ref(false)

const mapMarkers = ref<any[]>([])

onLoad(async (options: any) => {
  cpId.value = Number(options.id) || 0
  routeId.value = Number(options.routeId) || 0

  try {
    const detail = await routeStore.fetchRouteDetail(routeId.value)
    const checkpoint = detail.checkpoints?.find((c) => c.id === cpId.value)
    if (checkpoint) {
      cp.value = checkpoint

      // 设置地图标记
      mapMarkers.value = [{
        id: checkpoint.id,
        latitude: checkpoint.latitude,
        longitude: checkpoint.longitude,
        iconPath: '/static/icons/marker-pending.png',
        width: 40,
        height: 40,
        callout: {
          content: checkpoint.name,
          fontSize: 13,
          padding: 8,
          borderRadius: 6,
          display: 'ALWAYS',
        },
      }]

      // 检查是否已打卡
      await checkinStore.fetchCheckins({ route_id: routeId.value, page: 1 })
      isChecked.value = checkinStore.checkins.some((c) => c.checkpoint_id === cpId.value)
    }
  } catch (_) {
    // handled
  }
})

function goCheckin() {
  if (!cp.value) return
  uni.navigateTo({
    url: `/pages/checkin-create/index?routeId=${routeId.value}&checkpointId=${cp.value.id}&name=${encodeURIComponent(cp.value.name)}`,
  })
}

function goBack() {
  uni.navigateBack()
}
</script>

<style lang="scss" scoped>
.cp-detail-page {
  min-height: 100vh;
  background: $color-bg-secondary;
}

.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: $z-sticky;
  height: $navbar-height;
  display: flex;
  align-items: center;
  justify-content: center;
  background: $glass-bg;
  backdrop-filter: blur($glass-blur);
  padding-top: constant(safe-area-inset-top);
  padding-top: env(safe-area-inset-top);

  .navbar-back {
    position: absolute;
    left: $page-inset;
    top: 50%;
    transform: translateY(-50%);
    padding-top: constant(safe-area-inset-top);
    padding-top: env(safe-area-inset-top);
  }

  .navbar-title {
    font-size: $font-size-headline;
    font-weight: $font-weight-semibold;
    color: $color-gray-900;
  }
}

.page-body {
  padding-top: calc($navbar-height + constant(safe-area-inset-top));
  padding-top: calc($navbar-height + env(safe-area-inset-top));
}

.mini-map {
  width: 100%;
  height: 320rpx;

  .cp-map {
    width: 100%;
    height: 100%;
  }
}

.card {
  margin: $space-4 $page-inset;
  padding: $space-5;
}

.cp-name {
  display: block;
  font-size: $font-size-title2;
  font-weight: $font-weight-bold;
  color: $color-gray-900;
  margin-bottom: $space-3;
}

.cp-addr {
  display: flex;
  align-items: center;
  gap: $space-1;
  font-size: $font-size-subhead;
  color: $color-gray-600;
  margin-bottom: $space-2;
}

.cp-desc {
  font-size: $font-size-callout;
  color: $color-gray-700;
  line-height: $line-height-normal;
  margin-bottom: $space-3;
}

.cp-meta {
  display: flex;
  flex-direction: column;
  gap: $space-2;

  .meta-item {
    display: flex;
    align-items: center;
    gap: $space-1;
    font-size: $font-size-footnote;
    color: $color-gray-500;
  }
}

.section-label {
  display: block;
  font-size: $font-size-subhead;
  font-weight: $font-weight-medium;
  color: $color-gray-600;
  margin-bottom: $space-2;
}

.coord-row {
  display: flex;
  gap: $space-6;

  .coord-text {
    font-size: $font-size-footnote;
    color: $color-gray-500;
  }
}

.action-section {
  text-align: center;
  padding: $space-8 0;

  .checkin-btn {
    display: inline-flex;
    align-items: center;
    gap: $space-2;
    padding: $space-4 $space-10;
    background: $color-primary-500;
    color: #FFFFFF;
    border-radius: $radius-full;
    box-shadow: 0 8rpx 24rpx rgba($color-primary-500, 0.3);

    .btn-text {
      font-size: $font-size-body;
      font-weight: $font-weight-semibold;
    }

    &:active {
      transform: scale(0.96);
      opacity: 0.9;
    }
  }

  .checked-badge {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: $space-2;

    .checked-text {
      font-size: $font-size-headline;
      font-weight: $font-weight-semibold;
      color: $color-success;
    }
  }
}
</style>
