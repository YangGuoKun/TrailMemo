package memory

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/trailmemo/internal/config"
	"github.com/trailmemo/internal/platform/logger"
	"go.uber.org/zap"
)

// PreferenceStore 管理用户 AI 画像的持久化存储。
// 支持显式配置 + 行为信号渐进学习的双来源更新。
type PreferenceStore struct{}

func NewPreferenceStore() *PreferenceStore { return &PreferenceStore{} }

// GetOrCreate 获取用户偏好；不存在时创建默认值。
func (s *PreferenceStore) GetOrCreate(ctx context.Context, userID uint64) (*AgentUserPreference, error) {
	db := config.GetDB().WithContext(ctx)
	var pref AgentUserPreference
	err := db.Where("user_id = ?", userID).First(&pref).Error
	if err == nil {
		return &pref, nil
	}
	// 创建默认画像
	pref = AgentUserPreference{
		UserID: userID, Source: "default", Confidence: 0,
		TravelStyles: "[]", Interests: "[]", PreferredCities: "[]",
		AvoidedCities: "[]", DislikedFactors: "[]", CompanionTypes: "[]",
	}
	if err := db.Create(&pref).Error; err != nil {
		return nil, fmt.Errorf("创建默认画像失败: %w", err)
	}
	return &pref, nil
}

// UpdateExplicit 用户手动设置的偏好（显式优先，置信度0.9+）。
func (s *PreferenceStore) UpdateExplicit(ctx context.Context, userID uint64, update *PreferenceUpdate) error {
	pref, err := s.GetOrCreate(ctx, userID)
	if err != nil {
		return err
	}

	if len(update.TravelStyles) > 0 {
		pref.TravelStyles = toJSON(update.TravelStyles)
	}
	if len(update.BudgetRange) == 2 {
		pref.BudgetLevel = fmt.Sprintf("%d-%d", update.BudgetRange[0], update.BudgetRange[1])
	}
	if len(update.PreferredDays) == 2 {
		pref.Pace = fmt.Sprintf("%d-%d天", update.PreferredDays[0], update.PreferredDays[1])
	}
	if update.Pace != "" {
		pref.Pace = update.Pace
	}
	if len(update.Interests) > 0 {
		pref.Interests = toJSON(update.Interests)
	}
	if len(update.CompanionTypes) > 0 {
		pref.CompanionTypes = toJSON(update.CompanionTypes)
	}
	if len(update.PreferredCities) > 0 {
		pref.PreferredCities = toJSON(update.PreferredCities)
	}
	if len(update.AvoidList) > 0 {
		pref.DislikedFactors = toJSON(update.AvoidList)
	}
	pref.Confidence = 0.9
	pref.Source = "explicit"

	db := config.GetDB().WithContext(ctx)
	return db.Save(pref).Error
}

// RecordSignal 记录一个行为信号，渐进更新偏好（置信度按权重累积）。
// 对应设计文档 §9.3：收藏+1、复用+2、完成打卡+1、高评分+1。
func (s *PreferenceStore) RecordSignal(ctx context.Context, userID uint64, signal BehaviorSignal) error {
	pref, err := s.GetOrCreate(ctx, userID)
	if err != nil {
		return err
	}

	var interests []string
	json.Unmarshal([]byte(pref.Interests), &interests)
	for _, tag := range signal.Tags {
		if !containsStr(interests, tag) {
			interests = append(interests, tag)
		}
	}
	pref.Interests = toJSON(interests)

	// 城市偏好
	if signal.City != "" {
		var cities []string
		json.Unmarshal([]byte(pref.PreferredCities), &cities)
		if !containsStr(cities, signal.City) {
			cities = append(cities, signal.City)
			pref.PreferredCities = toJSON(cities)
		}
	}

	// 置信度渐进增长（上限0.85，行为推断不超显式设定的0.9）
	newConf := pref.Confidence + float64(signal.Weight)*0.05
	if newConf > 0.85 {
		newConf = 0.85
	}
	pref.Confidence = newConf
	if pref.Source == "default" {
		pref.Source = "behavior"
	} else if pref.Source == "behavior" { /* keep */
	} else {
		pref.Source = "mixed"
	}

	db := config.GetDB().WithContext(ctx)
	logger.FromContext(ctx).Debug("preference_signal_recorded",
		zap.Uint64("user_id", userID),
		zap.String("action", signal.Action),
		zap.Int("weight", signal.Weight),
		zap.Float64("confidence", pref.Confidence))
	return db.Save(pref).Error
}

// GetSnapshot 返回用于注入 Workflow 的偏好快照。
func (s *PreferenceStore) GetSnapshot(ctx context.Context, userID uint64) *PreferenceSnapshot {
	pref, err := s.GetOrCreate(ctx, userID)
	if err != nil {
		return &PreferenceSnapshot{}
	}
	var interests, cities []string
	json.Unmarshal([]byte(pref.Interests), &interests)
	json.Unmarshal([]byte(pref.PreferredCities), &cities)
	return &PreferenceSnapshot{
		BudgetLevel: pref.BudgetLevel, Pace: pref.Pace,
		Interests: interests, PreferredCities: cities, Confidence: pref.Confidence,
	}
}

// DeleteMemory 清空用户 AI 记忆。
func (s *PreferenceStore) DeleteMemory(ctx context.Context, userID uint64) error {
	db := config.GetDB().WithContext(ctx)
	return db.Where("user_id = ?", userID).Delete(&AgentUserPreference{}).Error
}

// ── 辅助类型 ──────────────────────────────────────

// BehaviorSignal 是一次用户行为信号。
type BehaviorSignal struct {
	Action string   `json:"action"` // favorite/remix/checkin/high_rate
	Weight int      `json:"weight"` // 1-3
	Tags   []string `json:"tags"`   // 关联标签
	City   string   `json:"city"`   // 关联城市
}

// PreferenceUpdate 是用户显式设置的偏好更新。
type PreferenceUpdate struct {
	TravelStyles    []string `json:"travel_styles"`
	BudgetRange     []int    `json:"budget_range"`
	PreferredDays   []int    `json:"preferred_days"`
	Pace            string   `json:"pace"`
	Interests       []string `json:"interests"`
	CompanionTypes  []string `json:"companion_types"`
	PreferredCities []string `json:"preferred_cities"`
	AvoidList       []string `json:"avoid_list"`
}

// PreferenceSnapshot 是注入 Workflow 的轻量偏好快照。
type PreferenceSnapshot struct {
	BudgetLevel     string   `json:"budget_level"`
	Pace            string   `json:"pace"`
	Interests       []string `json:"interests"`
	PreferredCities []string `json:"preferred_cities"`
	Confidence      float64  `json:"confidence"`
}

func (s *PreferenceSnapshot) IsEmpty() bool {
	return s.Confidence == 0 && len(s.Interests) == 0
}

// Summary 返回可注入 LLM prompt 的中文偏好描述。
func (s *PreferenceSnapshot) Summary() string {
	if s.IsEmpty() {
		return "暂无偏好数据"
	}
	desc := ""
	if s.BudgetLevel != "" {
		desc += fmt.Sprintf("预算：%s；", s.BudgetLevel)
	}
	if s.Pace != "" {
		desc += fmt.Sprintf("节奏：%s；", s.Pace)
	}
	if len(s.Interests) > 0 {
		desc += fmt.Sprintf("兴趣：%v；", s.Interests)
	}
	if len(s.PreferredCities) > 0 {
		desc += fmt.Sprintf("偏好城市：%v；", s.PreferredCities)
	}
	return desc
}

func toJSON(v interface{}) string { b, _ := json.Marshal(v); return string(b) }
func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
