// Package runtime 实现 Agent 意图路由分类器。
// Phase 4 使用加权关键词匹配 + 置信度评分；
// Phase 5+ 可替换为 LLM-based 分类或 embedding 相似度分类。
package runtime

import "strings"

// IntentRule 定义一条意图匹配规则。
type IntentRule struct {
	Keywords []string // 触发关键词（OR 关系）
	Weight   int      // 命中后的基础权重（短关键词权重低，长关键词权重高）
}

// RouteResult 是意图路由的输出，包含意图、模式和置信度。
type RouteResult struct {
	Intent       Intent        // 识别出的意图
	Mode         ExecutionMode // 推荐执行模式
	Confidence   float64       // 置信度 0-1
	MatchedWords []string      // 命中了哪些关键词（用于调试和解释）
}

// IntentRouter 基于加权关键词的意图分类器。
// 解决关键词重叠：按 Weight 和命中数综合评分，取最高分。
type IntentRouter struct {
	rules map[Intent]IntentRule
}

// NewIntentRouter 创建默认关键词规则的意图路由器。
func NewIntentRouter() *IntentRouter {
	return &IntentRouter{rules: map[Intent]IntentRule{
		IntentRouteDraft: {
			Keywords: []string{
				"规划路线", "生成路线", "创建路线", "做个攻略", "安排行程",
				"几天游", "一日游", "两日游", "三日游", "四日游", "五日游",
				"行程怎么安排", "帮我规划", "怎么玩比较好", "路线推荐",
				"去几天", "玩几天", "行程建议", "自由行路线",
			},
			Weight: 3,
		},
		IntentRouteRemix: {
			Keywords: []string{
				"改造路线", "改一下", "改造一下", "调整路线", "优化路线",
				"改成亲子", "改成情侣", "改成美食版", "改成轻松版", "改成低强度",
				"亲子版", "情侣版", "美食版", "轻松版", "雨天版",
				"少走路", "多加", "替换", "重新安排",
			},
			Weight: 3,
		},
		IntentRecommend: {
			Keywords: []string{
				"推荐", "建议", "哪个好", "有什么选择", "去哪里",
				"去哪玩", "哪个适合", "避暑", "度假", "适合",
				"有什么", "帮我选", "有哪些", "求推荐", "安利",
			},
			Weight: 2,
		},
		IntentTravelNote: {
			Keywords: []string{
				"写游记", "生成游记", "总结旅程", "记录旅行", "打卡记录",
				"帮我写", "来一篇", "分享到社区", "发布游记",
				"写一篇", "写个游记", "旅行总结",
			},
			Weight: 3,
		},
	}}
}

// SetRules 允许外部覆盖或扩充关键词规则（配置化入口）。
func (r *IntentRouter) SetRules(rules map[Intent]IntentRule) {
	r.rules = rules
}

// Route 根据用户输入计算最佳意图路由。
// 返回置信度最高的 RouteResult。
func (r *IntentRouter) Route(userMessage string) RouteResult {
	msg := strings.ToLower(userMessage)
	var best RouteResult

	for intent, rule := range r.rules {
		score := 0
		matched := make([]string, 0)

		for _, kw := range rule.Keywords {
			if strings.Contains(msg, strings.ToLower(kw)) {
				score += rule.Weight
				// 长关键词（4字以上）额外加分，避免"推荐"误匹配"路线推荐"
				if len([]rune(kw)) >= 5 {
					score += 2
				}
				matched = append(matched, kw)
			}
		}

		if score > 0 {
			// 归一化置信度：maxScore 按所有规则总分估算
			maxPossible := 12 // 假设最多命中4个5字关键词，每个weight+2
			confidence := float64(score) / float64(maxPossible)
			if confidence > 1.0 {
				confidence = 1.0
			}

			if confidence > best.Confidence {
				best = RouteResult{
					Intent:       intent,
					Mode:         intentToMode(intent),
					Confidence:   confidence,
					MatchedWords: matched,
				}
			}
		}
	}

	// 无命中 → 开放对话
	if best.Confidence == 0 {
		best = RouteResult{
			Intent:     IntentChat,
			Mode:       ExecutionModeAgentLoop,
			Confidence: 0.0,
		}
	}

	return best
}

// intentToMode 将意图映射到推荐执行模式。
func intentToMode(intent Intent) ExecutionMode {
	switch intent {
	case IntentChat:
		return ExecutionModeAgentLoop
	case IntentModeration:
		return ExecutionModeOneShot
	default:
		return ExecutionModeWorkflow
	}
}
