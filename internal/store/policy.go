package store

import (
	"encoding/json"
	"log"
	"time"

	"gorm.io/gorm"
)

// GetWorldPolicy 获取世界的动作策略。
func GetWorldPolicy(worldUUID string) (*WorldPolicyModel, error) {
	return GetWorldPolicyTx(DB, worldUUID)
}

func GetWorldPolicyTx(tx *gorm.DB, worldUUID string) (*WorldPolicyModel, error) {
	var p WorldPolicyModel
	if err := tx.Where("world_uuid = ?", worldUUID).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// UpsertWorldPolicy 创建或更新世界的动作策略。
// blocked 和 safe 以 JSON 字符串形式存储。
func UpsertWorldPolicy(worldUUID string, blocked, safe []string) (*WorldPolicyModel, error) {
	blockedJSON, _ := json.Marshal(blocked)
	safeJSON, _ := json.Marshal(safe)

	worldID := ResolveWorldUUID(worldUUID)
	var existing WorldPolicyModel
	err := DB.Where("world_uuid = ?", worldUUID).First(&existing).Error
	if err == nil {
		existing.BlockedActions = string(blockedJSON)
		existing.SafeActions = string(safeJSON)
		existing.UpdatedAt = time.Now()
		if err := Writer().Save(&existing).Error; err != nil {
			return nil, err
		}
		return &existing, nil
	}

	p := &WorldPolicyModel{
		WorldID:        worldID,
		WorldUUID:      worldUUID,
		BlockedActions: string(blockedJSON),
		SafeActions:    string(safeJSON),
	}
	if err := Writer().Create(p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

// GetOrCreateWorldPolicy 获取或创建默认世界策略。
func GetOrCreateWorldPolicy(worldUUID string) (*WorldPolicyModel, error) {
	p, err := GetWorldPolicy(worldUUID)
	if err == nil {
		return p, nil
	}
	worldID := ResolveWorldUUID(worldUUID)
	p = &WorldPolicyModel{
		WorldID:        worldID,
		WorldUUID:      worldUUID,
		BlockedActions: "[]",
		SafeActions:    "[]",
	}
	if err := Writer().Create(p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

// ParseBlockedActions 将策略记录的 JSON 字符串解析为字符串切片。
func (p *WorldPolicyModel) ParseBlockedActions() []string {
	var actions []string
	if err := json.Unmarshal([]byte(p.BlockedActions), &actions); err != nil {
		log.Printf("[policy] parse blocked_actions: %v", err)
	}
	return actions
}

// ParseSafeActions 将策略记录的 JSON 字符串解析为字符串切片。
func (p *WorldPolicyModel) ParseSafeActions() []string {
	var actions []string
	if err := json.Unmarshal([]byte(p.SafeActions), &actions); err != nil {
		log.Printf("[policy] parse safe_actions: %v", err)
	}
	return actions
}
