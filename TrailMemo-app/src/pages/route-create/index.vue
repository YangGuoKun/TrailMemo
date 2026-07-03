<template>
  <view class="create-page">
    <!-- 导航栏 -->
    <view class="navbar">
      <view class="navbar-back" @tap="goBack">
        <u-icon name="arrow-left" size="40" color="#1C1C1E" />
      </view>
      <text class="navbar-title">{{ isEdit ? '编辑路线' : '创建路线' }}</text>
      <view class="navbar-save" @tap="handleSave">
        <text class="save-text">保存</text>
      </view>
    </view>

    <scroll-view class="form-body" scroll-y>
      <!-- 基本信息 -->
      <view class="section card">
        <text class="section-title">基本信息</text>

        <!-- 封面图 -->
        <view class="cover-picker" @tap="pickCover">
          <image v-if="coverImage" :src="coverImage" mode="aspectFill" class="cover-preview" />
          <view v-else class="cover-placeholder">
            <u-icon name="camera" size="48" color="#C7C7CC" />
            <text class="placeholder-text">添加封面图片</text>
          </view>
        </view>

        <u-input
          v-model="form.title"
          placeholder="给路线起个名字吧"
          border="bottom"
          clearable
        />

        <u-input
          v-model="form.description"
          placeholder="描述这条路线的亮点..."
          border="bottom"
          type="textarea"
          :auto-height="true"
        />

        <view class="city-row">
          <u-input
            v-model="form.startCity"
            placeholder="出发城市"
            border="bottom"
          />
          <u-icon name="arrow-right" size="32" color="#C7C7CC" />
          <u-input
            v-model="form.endCity"
            placeholder="目的城市"
            border="bottom"
          />
        </view>
      </view>

      <!-- 行程参数 -->
      <view class="section card">
        <view class="section-header flex-between" @tap="showParams = !showParams">
          <text class="section-title">行程参数</text>
          <u-icon :name="showParams ? 'arrow-up' : 'arrow-down'" size="32" color="#C7C7CC" />
        </view>

        <view v-if="showParams">
          <u-input
            v-model="form.totalDistance"
            placeholder="总距离 (km)"
            border="bottom"
            type="number"
          />
          <u-input
            v-model="form.estimatedHours"
            placeholder="预估时长 (小时)"
            border="bottom"
            type="number"
          />
          <view class="switch-row flex-between">
            <text class="switch-label">公开路线</text>
            <u-switch v-model="form.isPublic" activeColor="#34C759" />
          </view>
        </view>
      </view>

      <!-- 打卡点 -->
      <view class="section">
        <view class="section-header flex-between">
          <text class="section-title">打卡点 ({{ checkpoints.length }})</text>
          <view class="add-btn" @tap="addCheckpoint">
            <u-icon name="plus" size="32" color="#0DA5BF" />
            <text class="add-text">添加</text>
          </view>
        </view>

        <view
          v-for="(cp, index) in checkpoints"
          :key="index"
          class="checkpoint-card card"
        >
          <view class="cp-header flex-between">
            <view class="cp-seq">
              <text class="seq-num">{{ index + 1 }}</text>
              <text class="seq-label">打卡点</text>
            </view>
            <view class="cp-actions">
              <view class="action-icon" @tap="moveUp(index)" v-if="index > 0">
                <u-icon name="arrow-up" size="32" color="#8E8E93" />
              </view>
              <view class="action-icon" @tap="moveDown(index)" v-if="index < checkpoints.length - 1">
                <u-icon name="arrow-down" size="32" color="#8E8E93" />
              </view>
              <view class="action-icon delete" @tap="removeCheckpoint(index)">
                <u-icon name="trash" size="32" color="#FF3B30" />
              </view>
            </view>
          </view>

          <u-input
            v-model="cp.name"
            placeholder="景点/地点名称 *"
            border="bottom"
          />
          <u-input
            v-model="cp.address"
            placeholder="详细地址"
            border="bottom"
          />

          <!-- 坐标选择 -->
          <view class="location-picker" @tap="pickLocation(index)">
            <u-icon name="map" size="32" color="#0DA5BF" />
            <text class="location-text">
              {{ cp.latitude ? `${cp.latitude.toFixed(6)}, ${cp.longitude.toFixed(6)}` : '在地图上选择位置' }}
            </text>
            <u-icon name="arrow-right" size="28" color="#C7C7CC" />
          </view>

          <view class="cp-extra">
            <u-input
              v-model="cp.arriveTime"
              placeholder="预计到达时间"
              border="bottom"
              class="half-input"
            />
            <u-input
              v-model="cp.stayDuration"
              placeholder="停留时长(分钟)"
              border="bottom"
              type="number"
              class="half-input"
            />
          </view>
        </view>
      </view>

      <view style="height: 200rpx" />
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useRouteStore } from '@/stores/useRouteStore'
import { useImagePicker } from '@/composables/useImagePicker'

const routeStore = useRouteStore()
const { pickImage } = useImagePicker()

const isEdit = ref(false)
const editId = ref<number>(0)
const coverImage = ref('')
const showParams = ref(false)

const form = reactive({
  title: '',
  description: '',
  startCity: '',
  endCity: '',
  totalDistance: '',
  estimatedHours: '',
  isPublic: true,
})

interface CheckpointForm {
  name: string
  address: string
  latitude: number | null
  longitude: number | null
  arriveTime: string
  stayDuration: string
}

const checkpoints = ref<CheckpointForm[]>([])

onLoad(async (options: any) => {
  if (options.id) {
    isEdit.value = true
    editId.value = Number(options.id)
    const detail = await routeStore.fetchRouteDetail(editId.value)
    form.title = detail.title
    form.description = detail.description
    form.startCity = detail.start_city
    form.endCity = detail.end_city
    form.totalDistance = String(detail.total_distance || '')
    form.estimatedHours = String(detail.estimated_hours || '')
    coverImage.value = detail.cover_image
    checkpoints.value = (detail.checkpoints || []).map((cp) => ({
      name: cp.name,
      address: cp.address || '',
      latitude: cp.latitude,
      longitude: cp.longitude,
      arriveTime: cp.arrive_time || '',
      stayDuration: String(cp.stay_duration || ''),
    }))
  }
})

// 封面选择
async function pickCover() {
  try {
    const path = await pickImage()
    coverImage.value = path
  } catch (_) {
    // cancelled
  }
}

// 地图选点
function pickLocation(index: number) {
  uni.chooseLocation({
    success: (res) => {
      checkpoints.value[index].latitude = res.latitude
      checkpoints.value[index].longitude = res.longitude
      if (res.name) checkpoints.value[index].name = res.name
      if (res.address) checkpoints.value[index].address = res.address
    },
  })
}

// 打卡点增删排序
function addCheckpoint() {
  checkpoints.value.push({
    name: '',
    address: '',
    latitude: null,
    longitude: null,
    arriveTime: '',
    stayDuration: '',
  })
}

function removeCheckpoint(index: number) {
  checkpoints.value.splice(index, 1)
}

function moveUp(index: number) {
  if (index <= 0) return
  const item = checkpoints.value[index]
  checkpoints.value.splice(index, 1)
  checkpoints.value.splice(index - 1, 0, item)
}

function moveDown(index: number) {
  if (index >= checkpoints.value.length - 1) return
  const item = checkpoints.value[index]
  checkpoints.value.splice(index, 1)
  checkpoints.value.splice(index + 1, 0, item)
}

// 保存
async function handleSave() {
  if (!form.title.trim()) {
    uni.showToast({ title: '请输入路线名称', icon: 'none' })
    return
  }
  if (!form.startCity.trim() || !form.endCity.trim()) {
    uni.showToast({ title: '请填写出发和目的城市', icon: 'none' })
    return
  }
  if (checkpoints.value.length === 0) {
    uni.showToast({ title: '请至少添加一个打卡点', icon: 'none' })
    return
  }

  const invalidCp = checkpoints.value.find((cp) => !cp.name.trim())
  if (invalidCp) {
    uni.showToast({ title: '请填写所有打卡点的名称', icon: 'none' })
    return
  }

  const data: CreateRouteRequest = {
    title: form.title,
    description: form.description,
    coverImage: coverImage.value,
    startCity: form.startCity,
    endCity: form.endCity,
    totalDistance: form.totalDistance ? Number(form.totalDistance) : undefined,
    estimatedHours: form.estimatedHours ? Number(form.estimatedHours) : undefined,
    isPublic: form.isPublic ? 1 : 0,
    checkpoints: checkpoints.value.map((cp, idx) => ({
      name: cp.name,
      latitude: cp.latitude || 0,
      longitude: cp.longitude || 0,
      address: cp.address,
      sequence: idx + 1,
      arriveTime: cp.arriveTime,
      stayDuration: cp.stayDuration ? Number(cp.stayDuration) : undefined,
    })),
  }

  try {
    if (isEdit.value) {
      await routeStore.updateRoute(editId.value, data)
      uni.showToast({ title: '更新成功', icon: 'success' })
    } else {
      await routeStore.createRoute(data)
      uni.showToast({ title: '创建成功', icon: 'success' })
    }
    setTimeout(() => uni.navigateBack(), 500)
  } catch (_) {
    // handled
  }
}

function goBack() {
  uni.navigateBack()
}
</script>

<style lang="scss" scoped>
.create-page {
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
  -webkit-backdrop-filter: blur($glass-blur);
  padding-top: constant(safe-area-inset-top);
  padding-top: env(safe-area-inset-top);

  .navbar-back, .navbar-save {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    padding-top: constant(safe-area-inset-top);
    padding-top: env(safe-area-inset-top);
  }

  .navbar-back { left: $page-inset; }
  .navbar-save { right: $page-inset; }

  .navbar-title {
    font-size: $font-size-headline;
    font-weight: $font-weight-semibold;
    color: $color-gray-900;
  }

  .save-text {
    font-size: $font-size-callout;
    color: $color-primary-500;
    font-weight: $font-weight-semibold;
  }
}

.form-body {
  padding-top: calc($navbar-height + constant(safe-area-inset-top) + $space-4);
  padding-top: calc($navbar-height + env(safe-area-inset-top) + $space-4);
  padding-left: $page-inset;
  padding-right: $page-inset;
}

.section {
  padding: $space-5;
  margin-bottom: $space-4;
}

.section-title {
  font-size: $font-size-headline;
  font-weight: $font-weight-semibold;
  color: $color-gray-900;
  display: block;
  margin-bottom: $space-3;
}

.section-header {
  margin-bottom: $space-3;
}

.cover-picker {
  width: 100%;
  height: 320rpx;
  border-radius: $radius-md;
  overflow: hidden;
  margin-bottom: $space-4;
  background: $color-gray-100;

  .cover-preview {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .cover-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: $space-2;

    .placeholder-text {
      font-size: $font-size-subhead;
      color: $color-gray-400;
    }
  }
}

.city-row {
  display: flex;
  align-items: center;
  gap: $space-2;
  margin-top: $space-2;
}

.switch-row {
  padding: $space-3 0;

  .switch-label {
    font-size: $font-size-callout;
    color: $color-gray-800;
  }
}

.add-btn {
  display: flex;
  align-items: center;
  gap: 4rpx;
  padding: $space-1 $space-3;

  .add-text {
    font-size: $font-size-subhead;
    color: $color-primary-500;
    font-weight: $font-weight-medium;
  }
}

.checkpoint-card {
  padding: $space-4;
  margin-bottom: $space-3;

  .cp-header {
    margin-bottom: $space-3;

    .cp-seq {
      .seq-num {
        font-size: $font-size-title3;
        font-weight: $font-weight-bold;
        color: $color-primary-500;
        margin-right: 4rpx;
      }
      .seq-label {
        font-size: $font-size-caption2;
        color: $color-gray-400;
      }
    }

    .cp-actions {
      display: flex;
      gap: $space-2;

      .action-icon {
        width: 48rpx;
        height: 48rpx;
        display: flex;
        align-items: center;
        justify-content: center;
      }
    }
  }

  .location-picker {
    display: flex;
    align-items: center;
    gap: $space-2;
    padding: $space-3 0;
    border-bottom: 1rpx solid $color-gray-200;

    .location-text {
      flex: 1;
      font-size: $font-size-subhead;
      color: $color-gray-500;
    }
  }

  .cp-extra {
    display: flex;
    gap: $space-3;

    .half-input {
      flex: 1;
    }
  }
}
</style>
