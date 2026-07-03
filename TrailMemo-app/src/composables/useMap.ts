/**
 * 腾讯地图操作 Hook
 * 封装微信小程序 map 组件的常用操作
 */

import { ref, type Ref } from 'vue'

interface MapMarker {
  id: number
  latitude: number
  longitude: number
  iconPath: string
  width: number
  height: number
  callout?: {
    content: string
    fontSize: number
    padding: number
    borderRadius: number
    display: string
  }
  label?: {
    content: string
    fontSize: number
    color: string
    anchorX: number
    anchorY: number
  }
}

interface MapPolyline {
  points: Array<{ latitude: number; longitude: number }>
  color: string
  width: number
  dottedLine?: boolean
  arrowLine?: boolean
  borderColor?: string
  borderWidth?: number
}

export function useMap(mapId: string = 'route-map') {
  const mapContext = ref<any>(null)
  const markers = ref<MapMarker[]>([])
  const polylines = ref<MapPolyline[]>([])
  const mapCenter = ref({ lat: 39.9042, lng: 116.4074 })
  const mapScale = ref(14)

  // 初始化地图上下文
  function initMap() {
    mapContext.value = uni.createMapContext(mapId)
  }

  // 设置打卡点标记
  function setCheckpointMarkers(
    checkpoints: Checkpoint[],
    checkedInIds: Set<number> = new Set(),
  ) {
    markers.value = checkpoints.map((cp, index) => ({
      id: cp.id,
      latitude: cp.latitude,
      longitude: cp.longitude,
      iconPath: checkedInIds.has(cp.id)
        ? '/static/icons/marker-checked.png'
        : '/static/icons/marker-pending.png',
      width: 36,
      height: 36,
      callout: {
        content: cp.name,
        fontSize: 12,
        padding: 8,
        borderRadius: 6,
        display: 'BYCLICK',
      },
      label: {
        content: String(index + 1),
        fontSize: 12,
        color: '#FFFFFF',
        anchorX: 18,
        anchorY: 13,
      },
    }))
  }

  // 设置用户位置标记
  function setUserLocationMarker(lat: number, lng: number) {
    markers.value.push({
      id: -1,
      latitude: lat,
      longitude: lng,
      iconPath: '/static/icons/marker-user.png',
      width: 24,
      height: 24,
      callout: {
        content: '我的位置',
        fontSize: 11,
        padding: 6,
        borderRadius: 4,
        display: 'ALWAYS',
      },
    })
  }

  // 绘制路线连线
  function drawRoutePolyline(checkpoints: Checkpoint[]) {
    if (checkpoints.length < 2) return

    const sorted = [...checkpoints].sort((a, b) => a.sequence - b.sequence)
    polylines.value = [{
      points: sorted.map((cp) => ({
        latitude: cp.latitude,
        longitude: cp.longitude,
      })),
      color: '#0DA5BF',
      width: 4,
      arrowLine: true,
      borderColor: '#FFFFFF',
      borderWidth: 1,
    }]
  }

  // 视角自动适配所有标记
  function fitAllMarkers(checkpoints: Checkpoint[]) {
    if (!mapContext.value || checkpoints.length === 0) return

    const points = checkpoints.map((cp) => ({
      latitude: cp.latitude,
      longitude: cp.longitude,
    }))
    mapContext.value.includePoints({
      points,
      padding: [50, 40, 50, 40],
    })
  }

  // 聚焦到某个打卡点
  function focusOnCheckpoint(checkpoint: Checkpoint) {
    if (!mapContext.value) return

    mapCenter.value = {
      lat: checkpoint.latitude,
      lng: checkpoint.longitude,
    }
    mapScale.value = 16
  }

  // 移动到用户位置
  function moveToUserLocation() {
    if (!mapContext.value) return
    mapContext.value.moveToLocation()
  }

  // 处理标记点击
  function onMarkerTap(e: any, callback: (checkpointId: number) => void) {
    const markerId = e.detail?.markerId ?? e.markerId
    if (markerId && markerId > 0) {
      callback(markerId)
    }
  }

  // 在地图上选点 (长按)
  function onMapTap(e: any): { lat: number; lng: number } | null {
    if (e.detail) {
      return {
        lat: e.detail.latitude,
        lng: e.detail.longitude,
      }
    }
    return null
  }

  return {
    mapContext,
    markers,
    polylines,
    mapCenter,
    mapScale,
    initMap,
    setCheckpointMarkers,
    setUserLocationMarker,
    drawRoutePolyline,
    fitAllMarkers,
    focusOnCheckpoint,
    moveToUserLocation,
    onMarkerTap,
    onMapTap,
  }
}
