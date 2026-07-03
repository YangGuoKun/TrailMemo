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
];
const addRouteWords = ['添加', '加入', '导入', '创建', '保存', '一键'];
const routeDestinationWords = ['我的路线', '打卡路线', '路线列表', '打卡列表'];
export function shouldUseRouteDraftWorkflow(message) {
    const text = message.trim();
    if (!text)
        return false;
    const hasRouteDraftWord = routeDraftWords.some((word) => text.includes(word));
    if (!hasRouteDraftWord)
        return false;
    const hasAddIntent = addRouteWords.some((word) => text.includes(word));
    const hasRouteDestination = routeDestinationWords.some((word) => text.includes(word));
    return hasAddIntent || hasRouteDestination || /[一二三四五六七八九十0-9]+日游/.test(text);
}
export function buildRouteDraftRequest(message) {
    const query = message.trim();
    const req = { query };
    const dayMatch = query.match(/([一二三四五六七八九十0-9]+)日游/);
    if (dayMatch) {
        req.days = parseDays(dayMatch[1]);
    }
    const styles = [];
    if (query.includes('美食'))
        styles.push('美食');
    if (query.includes('情侣'))
        styles.push('情侣');
    if (query.includes('亲子'))
        styles.push('亲子');
    if (query.includes('避暑'))
        styles.push('避暑');
    if (query.includes('历史') || query.includes('文化'))
        styles.push('历史文化');
    if (styles.length > 0)
        req.travel_styles = styles;
    return req;
}
function parseDays(value) {
    const numeric = Number(value);
    if (Number.isFinite(numeric) && numeric > 0)
        return numeric;
    const digits = {
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
    };
    if (value === '十')
        return 10;
    if (value.startsWith('十'))
        return 10 + (digits[value.slice(1)] || 0);
    if (value.endsWith('十'))
        return (digits[value[0]] || 1) * 10;
    if (value.includes('十')) {
        const [tens, ones] = value.split('十');
        return (digits[tens] || 1) * 10 + (digits[ones] || 0);
    }
    return digits[value] || 1;
}
