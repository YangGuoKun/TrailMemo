<template>
  <view class="detail-page">
    <view class="navbar">
      <view class="nav-back" @tap="goBack"><u-icon name="arrow-left" size="40" color="#1C1C1E" /></view>
      <text class="nav-title">{{ detail?.title || '路线详情' }}</text>
    </view>

    <scroll-view scroll-y class="page-body">
      <!-- 半屏地图 -->
      <view class="map-area">
        <map
          id="route-map" class="route-map"
          :latitude="mapLat" :longitude="mapLng"
          :markers="markers" :polyline="polylines"
          :scale="14" :min-scale="10" :max-scale="18"
          :enable-3D="true" :enable-overlooking="true"
          :enable-zoom="true" :enable-scroll="true" :enable-rotate="false"
          :show-location="true" :show-scale="false" :show-compass="true"
          :enable-traffic="false" :enable-poi="false" :enable-building="false"
          @markertap="onMarkerTap"
        />
        <view class="map-hint">{{ checkpoints.length }}个打卡点 · 点击可打卡</view>
      </view>

      <view class="glass-card">
        <text class="detail-title">{{ detail?.title }}</text>
        <text class="detail-desc" v-if="detail?.description">{{ detail.description }}</text>
        <view class="info-row">
          <text>📍 {{ detail?.start_city || '-' }} → {{ detail?.end_city || '-' }}</text>
          <text v-if="detail?.total_distance">🛣️ {{ detail.total_distance }}km</text>
          <text v-if="detail?.estimated_hours">🕐 {{ detail.estimated_hours }}h</text>
        </view>
      </view>

      <view class="date-pill">
        <text class="dp-day">{{ today }}</text>
        <view class="dp-month-wrap"><text class="dp-month">{{ year }} / {{ month }}</text><text class="dp-weekday">{{ weekday }}</text></view>
        <text class="dp-dot">·</text><text>{{ detail?.start_city || '出发' }}</text>
      </view>

      <view class="mood-row" v-if="checkedMoods.length">
        <text v-for="m in checkedMoods" :key="m" class="mood-badge" :class="'m-'+m">{{ moodMap[m] }}</text>
      </view>

      <scroll-view scroll-x class="photo-wall" v-if="checkedPhotos.length">
        <view v-for="(p,i) in checkedPhotos" :key="i" class="photo-tile" :class="i%3===0?'big':'small'" @tap="previewCheckinPhoto(p.url)">
          <image v-if="p.url" :src="p.url" mode="aspectFill" class="pt-img" />
          <view v-else class="pt-placeholder" :style="{background:photoGradients[i%4]}"><text class="pt-label">{{p.name}}</text></view>
        </view>
      </scroll-view>

      <view class="glass-card progress-card">
        <view class="progress-ring" :style="{background:'conic-gradient(#34C759 0deg '+progressDeg+'deg, rgba(0,0,0,0.04) '+progressDeg+'deg 360deg)'}">
          <view class="ring-inner"><text class="ring-text">{{checkedCount}}/{{checkpoints.length}}</text></view>
        </view>
        <view class="progress-info">
          <text class="pi-title">打卡进度 {{progressPercent}}%</text>
          <text class="pi-sub">已完成{{checkedCount}}个，还剩{{checkpoints.length-checkedCount}}个 ✨</text>
        </view>
      </view>

      <view class="glass-card">
        <text class="section-title">打卡点</text>
        <view v-for="(cp,i) in checkpoints" :key="cp.id" class="cp-item">
          <view class="cp-left">
            <view class="cp-dot" :class="{done:isChecked(cp.id), now:!isChecked(cp.id)&&i===nextIdx}">
              <text v-if="isChecked(cp.id)">✓</text><text v-else>{{cp.sequence||i+1}}</text>
            </view>
            <view v-if="i<checkpoints.length-1" class="cp-line" :class="{done:isChecked(cp.id)&&isChecked(checkpoints[i+1]?.id||0)}"/>
          </view>
          <view class="cp-right" @tap="onCheckpointTap(cp)">
            <view class="cp-card-inner" :class="{done:isChecked(cp.id),now:!isChecked(cp.id)&&i===nextIdx}">
              <text class="cp-name">{{cp.name}}</text>
              <text class="cp-addr" v-if="cp.address">{{cp.address}}</text>
              <text class="cp-time" v-if="cp.arrive_time||cp.stay_duration">🕐 {{cp.arrive_time||''}} · {{cp.stay_duration||0}}分钟</text>
              <view v-if="isChecked(cp.id)&&cpCheckinMap[cp.id]" class="cp-checkin-preview">
                <image v-if="cpCheckinMap[cp.id].photo_url" :src="getFullUrl(cpCheckinMap[cp.id].photo_url)" mode="aspectFill" class="cp-photo"/>
                <text class="cp-rating" v-if="cpCheckinMap[cp.id].rating">{{'★'.repeat(cpCheckinMap[cp.id].rating)}}{{'☆'.repeat(5-cpCheckinMap[cp.id].rating)}}</text>
              </view>
              <view v-if="!isChecked(cp.id)&&i===nextIdx" class="checkin-btn" @tap.stop="goCheckin(cp)"><text>📍 立即打卡</text></view>
            </view>
          </view>
        </view>
      </view>

      <view class="action-row">
        <button class="act-btn outline" @tap="goEditRoute">编辑路线</button>
        <button v-if="isCompleted||allChecked" class="act-btn share" @tap="handleShareRoute">📤 分享</button>
        <button class="act-btn danger" @tap="handleDelete">删除</button>
      </view>

      <view class="action-row ai-row">
        <button class="act-btn ai" @tap="handleGenerateNote">✍️ AI 生成游记</button>
        <button class="act-btn remix" @tap="handleRemixRoute">✨ AI 改造路线</button>
      </view>
      <view style="height:160rpx"/>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useRouteStore } from '@/stores/useRouteStore'
import { useCheckinStore } from '@/stores/useCheckinStore'
import { useCommunityStore } from '@/stores/useCommunityStore'
import { useAgentStore } from '@/stores/useAgentStore'
import { getFullUrl } from '@/config'

const routeStore = useRouteStore()
const checkinStore = useCheckinStore()
const agentStore = useAgentStore()
const detailId = ref(0)
const checkpoints = ref<any[]>([])
const cpCheckinMap = ref<Record<number,any>>({})
const moodMap: Record<string,string> = { happy:'😊 开心', excited:'🤩 兴奋', peace:'😌 平静', moved:'🥹 感动' }
const photoGradients = ['linear-gradient(135deg,#FFA585,#FFB88C)','linear-gradient(180deg,#89D4CF,#4DCADC)','linear-gradient(180deg,#E8C5A0,#D4A574)','linear-gradient(135deg,#C4A6E8,#FBC2EB)']
const now = new Date()
const today = ref(now.getDate())
const year = ref(now.getFullYear())
const month = ref(String(now.getMonth()+1).padStart(2,'0'))
const weekday = ref(['日','一','二','三','四','五','六'][now.getDay()])

const checkedMoods = computed(() => {
  const s = new Set<string>()
  Object.values(cpCheckinMap.value).forEach((ci:any) => { (ci.moods||[]).forEach((m:string) => s.add(m)) })
  return [...s]
})
const checkedPhotos = computed(() => {
  const list: {name:string;url:string}[] = []
  checkpoints.value.forEach((cp:any) => {
    const ci = cpCheckinMap.value[cp.id]
    if (ci?.photo_url) list.push({name:cp.name, url:getFullUrl(ci.photo_url)})
  })
  return list
})

const mapLat = ref(30.65)
const mapLng = ref(104.07)
const markers = ref<any[]>([])
const polylines = ref<any[]>([])

const detail = computed(() => routeStore.currentDetail)
const checkedCount = computed(() => checkpoints.value.filter((c:any) => cpCheckinMap.value[c.id]).length)
const nextIdx = computed(() => checkpoints.value.findIndex((c:any) => !cpCheckinMap.value[c.id]))
const progressPercent = computed(() => checkpoints.value.length ? Math.round(checkedCount.value/checkpoints.value.length*100) : 0)
const progressDeg = computed(() => progressPercent.value/100*360)
const isCompleted = computed(() => detail.value?.publish_status===2)
const allChecked = computed(() => checkedCount.value===checkpoints.value.length)
function isChecked(id:number){ return !!cpCheckinMap.value[id] }

onLoad(async (options:any) => {
  const id = Number(options.id)
  if(!id){ uni.showToast({title:'路线不存在',icon:'none'}); return }
  detailId.value = id
  try {
    const d = await routeStore.fetchRouteDetail(id)
    const cps = d.checkpoints || []
    checkpoints.value = cps
    const valid = cps.filter((c:any) => c.latitude && c.longitude)
    if(valid.length>0){
      mapLat.value = valid.reduce((s:number,c:any) => s+c.latitude,0)/valid.length
      mapLng.value = valid.reduce((s:number,c:any) => s+c.longitude,0)/valid.length
    } else { mapLat.value = 30.5728; mapLng.value = 104.0668 }
    const offsets = cps.map((_:any,i:number) => ({
      lat: mapLat.value + Math.cos(i*1.8)*0.008*(i+1)/Math.max(cps.length,1),
      lng: mapLng.value + Math.sin(i*1.8)*0.012*(i+1)/Math.max(cps.length,1),
    }))
    await checkinStore.fetchCheckins({route_id:id,page:1})
    const ck: Record<number,any> = {}
    checkinStore.checkins.forEach((ci:any) => { ck[ci.checkpoint_id] = ci })
    cpCheckinMap.value = ck
    markers.value = cps.map((cp:any,i:number) => {
      const lat = cp.latitude || offsets[i].lat
      const lng = cp.longitude || offsets[i].lng
      const done = !!ck[cp.id]
      return {
        id: cp.id, latitude: lat, longitude: lng, width: 30, height: 30,
        callout: { content: (i+1)+'. '+cp.name, fontSize: 12, padding: 8, borderRadius: 4, display: 'BYCLICK', bgColor: done?'#34C759':'#0DA5BF', color: '#FFFFFF' },
        label: { content: done?'✓':String(i+1), fontSize: 14, color: '#FFFFFF', bgColor: done?'#34C759':'#0DA5BF', borderRadius: 14, padding: 4, anchorX: -6, anchorY: -45 },
      }
    })
    const sorted = [...cps].sort((a:any,b:any) => (a.sequence||0)-(b.sequence||0))
    if(sorted.length>1){
      polylines.value = [{
        points: sorted.map((cp:any,i:number) => ({latitude:cp.latitude||offsets[i].lat, longitude:cp.longitude||offsets[i].lng})),
        color: '#0DA5BF', width: 4, arrowLine: true, borderColor: '#FFFFFF', borderWidth: 1,
      }]
    }
  } catch(_){ uni.showToast({title:'加载失败',icon:'none'}) }
})

function onMarkerTap(e:any){
  const id = e.detail?.markerId || e.markerId
  const cp = checkpoints.value.find((c:any) => c.id===id)
  if(!cp) return
  const ci = cpCheckinMap.value[id]
  if(ci){
    const stars = '★'.repeat(ci.rating||0)+'☆'.repeat(5-(ci.rating||0))
    uni.showModal({title:'📍 '+cp.name, content: stars+'\n\n"'+(ci.content||'未填写感受')+'"', confirmText:'知道了', showCancel:false})
  } else { goCheckin(cp) }
}
function onCheckpointTap(cp:any){ if(!cpCheckinMap.value[cp.id]) goCheckin(cp) }
function previewCheckinPhoto(url:string){ uni.previewImage({urls:[url],current:0}) }
function goCheckin(cp:any){ uni.navigateTo({url:`/pages/checkin-create/index?routeId=${detailId.value}&checkpointId=${cp.id}&name=${encodeURIComponent(cp.name)}`}) }
function goEditRoute(){ uni.navigateTo({url:`/pages/route-create/index?id=${detailId.value}`}) }
function goBack(){ uni.navigateBack() }
async function handleShareRoute(){
  if(!detail.value) return
  try {
    const cs = useCommunityStore()
    await cs.createPost({ route_id: detailId.value, title: detail.value.title+' · 完结分享 🎉', content: detail.value.description||'分享我的旅行路线~', images: detail.value.cover_image||'' })
    uni.showToast({title:'分享成功！',icon:'success'})
  } catch(_){ uni.showToast({title:'分享失败',icon:'none'}) }
}
function handleDelete(){
  uni.showModal({title:'删除路线',content:'确定删除？不可恢复',confirmColor:'#FF3B30',
    success:async(res)=>{ if(res.confirm){ await routeStore.deleteRoute(detailId.value); uni.showToast({title:'已删除',icon:'success'}); setTimeout(()=>uni.navigateBack(),500) }}})
}
async function handleGenerateNote(){
  uni.showLoading({ title: 'AI 正在生成...' })
  try {
    const note = await agentStore.generateTravelNote({
      route_id: detailId.value,
      mood: checkedMoods.value.join(', ') || '愉快',
      style: 'story',
    })
    uni.hideLoading()
    uni.showModal({
      title: '✍️ 游记生成成功',
      content: note.content.slice(0, 500) + (note.content.length > 500 ? '...' : ''),
      showCancel: false,
      confirmText: '去发布',
      success: (res) => {
        if (res.confirm) {
          uni.navigateTo({ url: `/pages/community-create/index?routeId=${detailId.value}&content=${encodeURIComponent(note.content)}` })
        }
      },
    })
  } catch (err) {
    uni.hideLoading()
    uni.showToast({ title: '生成失败', icon: 'none' })
  }
}
async function handleRemixRoute(){
  uni.showLoading({ title: 'AI 正在改造...' })
  try {
    const result = await agentStore.remixRoute(String(detailId.value), {
      query: '优化这条路线，增加更多美食体验',
    })
    uni.hideLoading()
    uni.showModal({
      title: '✨ 路线改造完成',
      content: result.route_draft.summary,
      showCancel: true,
      confirmText: '创建路线',
      success: async (res) => {
        if (res.confirm) {
          uni.showLoading({ title: '创建中...' })
          try {
            const commitResult = await agentStore.commitArtifact(result.artifact_id, {
              commit_type: 'create_route',
              idempotency_key: 'key_' + Date.now(),
            })
            uni.hideLoading()
            uni.navigateTo({ url: `/pages/route-detail/index?id=${commitResult.entity_id}` })
          } catch (_) {
            uni.hideLoading()
            uni.showToast({ title: '创建失败', icon: 'none' })
          }
        }
      },
    })
  } catch (err) {
    uni.hideLoading()
    uni.showToast({ title: '改造失败', icon: 'none' })
  }
}
</script>

<style lang="scss" scoped>
.detail-page{ min-height:100vh; background:linear-gradient(170deg,#FDF8F4,#F7F4F0,#F3F1F5,#F5F3F8) }
.navbar{ position:sticky; top:0; z-index:50; display:flex; align-items:center; justify-content:center; height:88rpx; padding:0 32rpx;
  background:rgba(255,255,255,0.72); backdrop-filter:saturate(170%) blur(24px); -webkit-backdrop-filter:saturate(170%) blur(24px); border-bottom:1px solid rgba(255,255,255,0.4);
  .nav-back{ position:absolute; left:16rpx; width:64rpx; height:64rpx; border-radius:50%; background:rgba(0,0,0,0.04); display:flex; align-items:center; justify-content:center }
  .nav-title{ font-size:34rpx; font-weight:600; color:#1C1C1E } }
.page-body{ padding-bottom:40rpx }
.map-area{ position:relative; width:100%; height:50vh; margin-bottom:20rpx; border-radius:16rpx; overflow:hidden;
  .route-map{ width:100%; height:100% }
  .map-hint{ position:absolute; bottom:12rpx; right:12rpx; background:rgba(0,0,0,0.45); color:#fff; font-size:20rpx; padding:4rpx 14rpx; border-radius:16rpx } }
.glass-card{ margin:14rpx 28rpx; padding:28rpx; background:rgba(255,255,255,0.55); border-radius:24rpx; box-shadow:0 8px 32px rgba(0,0,0,0.06),0 2px 8px rgba(0,0,0,0.03) }
.detail-title{ font-size:38rpx; font-weight:700; color:#1C1C1E; display:block }
.detail-desc{ font-size:26rpx; color:#636366; margin-top:8rpx; display:block; line-height:1.5 }
.info-row{ display:flex; gap:24rpx; margin-top:16rpx; font-size:24rpx; color:#636366 }
.progress-card{ display:flex; align-items:center; gap:24rpx }
.progress-ring{ width:100rpx; height:100rpx; border-radius:50%; display:flex; align-items:center; justify-content:center; flex-shrink:0 }
.ring-inner{ width:76rpx; height:76rpx; border-radius:50%; background:#fff; display:flex; align-items:center; justify-content:center }
.ring-text{ font-size:28rpx; font-weight:700; color:#1C1C1E }
.pi-title{ font-size:28rpx; font-weight:600; display:block }
.pi-sub{ font-size:22rpx; color:#AEAEB2; margin-top:4rpx }
.section-title{ font-size:30rpx; font-weight:600; color:#1C1C1E; margin-bottom:20rpx; display:block }
.date-pill{ display:inline-flex; align-items:center; gap:12rpx; margin:16rpx 28rpx 8rpx; padding:14rpx 22rpx; background:rgba(255,255,255,0.5); border-radius:9999rpx; box-shadow:0 8px 32px rgba(0,0,0,0.06),0 2px 8px rgba(0,0,0,0.03) }
.dp-day{ font-size:44rpx; font-weight:700; color:#1C1C1E; line-height:1 }
.dp-month-wrap{ display:flex; flex-direction:column }
.dp-month{ font-size:16rpx; color:#AEAEB2 }
.dp-weekday{ font-size:22rpx; color:#636366; font-weight:500 }
.dp-dot{ color:#AEAEB2 }
.mood-row{ display:flex; gap:10rpx; padding:0 28rpx; flex-wrap:wrap; margin-bottom:12rpx }
.mood-badge{ padding:10rpx 20rpx; border-radius:9999rpx; font-size:22rpx; font-weight:600; border:1px solid rgba(255,255,255,0.35) }
.m-happy{ background:rgba(255,159,10,0.14); color:#B86800 }
.m-excited{ background:rgba(255,59,48,0.1); color:#C92016 }
.m-peace{ background:rgba(13,165,191,0.1); color:#076F7E }
.m-moved{ background:rgba(255,107,138,0.1); color:#C93A58 }
.photo-wall{ display:flex; gap:10rpx; padding:0 28rpx; white-space:nowrap; margin-bottom:16rpx }
.photo-tile{ flex-shrink:0; border-radius:16rpx; overflow:hidden; background:rgba(255,255,255,0.55); box-shadow:0 8px 32px rgba(0,0,0,0.06) }
.photo-tile.big{ width:260rpx; height:300rpx }
.photo-tile.small{ width:180rpx; height:300rpx }
.pt-img{ width:100%; height:100%; object-fit:cover }
.pt-placeholder{ width:100%; height:100%; display:flex; align-items:flex-end; padding:16rpx }
.pt-label{ color:#fff; font-size:22rpx; font-weight:600; text-shadow:0 1px 3px rgba(0,0,0,0.3) }
.cp-item{ display:flex; gap:16rpx }
.cp-left{ display:flex; flex-direction:column; align-items:center; width:48rpx; flex-shrink:0 }
.cp-dot{ width:44rpx; height:44rpx; border-radius:50%; background:rgba(0,0,0,0.08); display:flex; align-items:center; justify-content:center; font-size:22rpx; font-weight:700; color:#AEAEB2;
  &.done{ background:#34C759; color:#fff; box-shadow:0 4px 12px rgba(52,199,89,0.3) }
  &.now{ background:#FF9F0A; color:#fff; box-shadow:0 4px 16px rgba(255,159,10,0.35) } }
.cp-line{ width:3rpx; flex:1; min-height:24rpx; background:rgba(0,0,0,0.06); margin:4rpx 0; &.done{ background:rgba(52,199,89,0.3) } }
.cp-right{ flex:1; padding-bottom:20rpx }
.cp-card-inner{ padding:20rpx; border-radius:16rpx; background:rgba(0,0,0,0.02);
  &.done{ background:rgba(52,199,89,0.05); border:1px solid rgba(52,199,89,0.12) }
  &.now{ border:1px solid rgba(255,159,10,0.2); background:rgba(255,159,10,0.03) }
  .cp-name{ font-size:30rpx; font-weight:600; color:#1C1C1E; display:block }
  .cp-addr{ font-size:22rpx; color:#AEAEB2; margin-top:4rpx; display:block }
  .cp-time{ font-size:22rpx; color:#636366; margin-top:4rpx; display:block }
  .cp-checkin-preview{ margin-top:12rpx }
  .cp-photo{ width:100%; height:160rpx; border-radius:12rpx; object-fit:cover }
  .cp-rating{ font-size:24rpx; color:#FF9F0A; margin-top:4rpx; display:block }
  .checkin-btn{ margin-top:12rpx; padding:12rpx 24rpx; background:#0DA5BF; border-radius:9999rpx; display:inline-flex;
    text{ color:#fff; font-size:24rpx; font-weight:600 } } }
.action-row{ display:flex; gap:12rpx; padding:20rpx 28rpx;
  .act-btn{ flex:1; height:80rpx; border-radius:16rpx; font-size:28rpx; font-weight:500;
    &.outline{ background:#fff; color:#0DA5BF; border:1px solid #0DA5BF }
    &.share{ background:linear-gradient(135deg,#FF6B6B,#FF8E53); color:#fff; border:none }
    &.danger{ background:#fff; color:#FF3B30; border:1px solid rgba(255,59,48,0.2) }
    &.ai{ background:linear-gradient(135deg,#A18CD1,#FBC2EB); color:#fff; border:none }
    &.remix{ background:linear-gradient(135deg,#89D4CF,#0DA5BF); color:#fff; border:none }
    &:active{ opacity:0.7; transform:scale(0.97) } }
}
.ai-row{ padding-top:0 }
</style>
