const routeDraftWords = [
  '规划路线',
  '生成路线',
  '创建路线',
  '设计一条',
  '设计路线',
  '安排行程',
  '做个攻略',
  '游玩路线',
  '旅行路线',
  '一日游',
  '两日游',
  '三日游',
  '四日游',
  '五日游',
  '攻略',
]

const addRouteWords = ['添加', '加入', '导入', '创建', '保存', '一键']
const generateRouteWords = ['生成', '设计', '规划', '做一份', '来一份', '安排']
const routeDestinationWords = [
  '我的路线',
  '个人路线',
  '打卡路线',
  '路线列表',
  '打卡列表',
  '个人打卡列表',
  'citywalk',
  'Citywalk',
]

type DraftRequest = {
  query: string
  days?: number
  travel_styles?: string[]
}

export function shouldUseRouteDraftWorkflow(message: string): boolean {
  const text = message.trim()
  if (!text) return false

  const hasRouteDraftWord = routeDraftWords.some((word) => text.includes(word))
  const hasAddIntent = addRouteWords.some((word) => text.includes(word))
  const hasGenerateIntent = generateRouteWords.some((word) => text.includes(word))
  const hasRouteDestination = routeDestinationWords.some((word) => text.includes(word))

  if (hasAddIntent && hasRouteDestination) return true
  if (hasGenerateIntent && hasRouteDestination) return true
  if (!hasRouteDraftWord) return false

  return hasAddIntent || hasRouteDestination || /[一二三四五六七八九十0-9]+日游/.test(text)
}

export function buildRouteDraftRequest(message: string): DraftRequest {
  const query = message.trim()
  const req: DraftRequest = { query }

  const dayMatch = query.match(/([一二三四五六七八九十0-9]+)日(?:游|行|citywalk|Citywalk)?/)
  if (dayMatch) {
    req.days = parseDays(dayMatch[1])
  }

  const styles: string[] = []
  if (query.includes('美食')) styles.push('美食')
  if (query.toLowerCase().includes('citywalk')) styles.push('Citywalk')
  if (query.includes('情侣')) styles.push('情侣')
  if (query.includes('亲子')) styles.push('亲子')
  if (query.includes('避暑')) styles.push('避暑')
  if (query.includes('历史') || query.includes('文化')) styles.push('历史文化')
  if (styles.length > 0) req.travel_styles = styles

  return req
}

function parseDays(value: string): number {
  const numeric = Number(value)
  if (Number.isFinite(numeric) && numeric > 0) return numeric

  const digits: Record<string, number> = {
    一: 1,
    二: 2,
    两: 2,
    三: 3,
    四: 4,
    五: 5,
    六: 6,
    七: 7,
    八: 8,
    九: 9,
    十: 10,
  }

  if (value === '十') return 10
  if (value.startsWith('十')) return 10 + (digits[value.slice(1)] || 0)
  if (value.endsWith('十')) return (digits[value[0]] || 1) * 10
  if (value.includes('十')) {
    const [tens, ones] = value.split('十')
    return (digits[tens] || 1) * 10 + (digits[ones] || 0)
  }

  return digits[value] || 1
}
