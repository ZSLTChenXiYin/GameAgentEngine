package sdk

import "strings"

const (
	PlayerIntentTypeSpeech       = "speech"
	PlayerIntentTypeShowItem     = "show_item"
	PlayerIntentTypeGift         = "gift"
	PlayerIntentTypeTradeRequest = "trade_request"
	PlayerIntentTypeThreaten     = "threaten"
	PlayerIntentTypeMove         = "move"
	PlayerIntentTypeInspect      = "inspect"
	PlayerIntentTypeUseItem      = "use_item"
	PlayerIntentTypeComposite    = "composite"
)

const (
	PlayerIntentRiskLow    = "low"
	PlayerIntentRiskMedium = "medium"
	PlayerIntentRiskHigh   = "high"
)

const (
	PlayerIntentPreconditionSameScene          = "same_scene"
	PlayerIntentPreconditionTargetPresent      = "target_present"
	PlayerIntentPreconditionItemPresent        = "item_present"
	PlayerIntentPreconditionMoneyAtLeast       = "money_at_least"
	PlayerIntentPreconditionTaskStatus         = "task_status"
	PlayerIntentPreconditionSceneFlag          = "scene_flag"
	PlayerIntentPreconditionLocationAccessible = "location_accessible"
)

const (
	MissingFactPlayerLocation = "player_location"
	MissingFactTargetLocation = "target_location"
	MissingFactItemPresence   = "item_presence"
	MissingFactSceneState     = "scene_state"
	MissingFactTaskState      = "task_state"
	MissingFactWalletState    = "wallet_state"
)

var playerIntentTypes = []string{
	PlayerIntentTypeSpeech,
	PlayerIntentTypeShowItem,
	PlayerIntentTypeGift,
	PlayerIntentTypeTradeRequest,
	PlayerIntentTypeThreaten,
	PlayerIntentTypeMove,
	PlayerIntentTypeInspect,
	PlayerIntentTypeUseItem,
	PlayerIntentTypeComposite,
}

var playerIntentStepTypes = []string{
	PlayerIntentTypeSpeech,
	PlayerIntentTypeShowItem,
	PlayerIntentTypeGift,
	PlayerIntentTypeTradeRequest,
	PlayerIntentTypeThreaten,
	PlayerIntentTypeMove,
	PlayerIntentTypeInspect,
	PlayerIntentTypeUseItem,
}

var playerIntentRiskLevels = []string{
	PlayerIntentRiskLow,
	PlayerIntentRiskMedium,
	PlayerIntentRiskHigh,
}

var playerIntentPreconditionTypes = []string{
	PlayerIntentPreconditionSameScene,
	PlayerIntentPreconditionTargetPresent,
	PlayerIntentPreconditionItemPresent,
	PlayerIntentPreconditionMoneyAtLeast,
	PlayerIntentPreconditionTaskStatus,
	PlayerIntentPreconditionSceneFlag,
	PlayerIntentPreconditionLocationAccessible,
}

var playerIntentMissingFactTypes = []string{
	MissingFactPlayerLocation,
	MissingFactTargetLocation,
	MissingFactItemPresence,
	MissingFactSceneState,
	MissingFactTaskState,
	MissingFactWalletState,
}

func ValidPlayerIntentTypes() []string {
	return append([]string(nil), playerIntentTypes...)
}

func ValidPlayerIntentStepTypes() []string {
	return append([]string(nil), playerIntentStepTypes...)
}

func ValidPlayerIntentRiskLevels() []string {
	return append([]string(nil), playerIntentRiskLevels...)
}

func ValidPlayerIntentPreconditionTypes() []string {
	return append([]string(nil), playerIntentPreconditionTypes...)
}

func ValidMissingFactTypes() []string {
	return append([]string(nil), playerIntentMissingFactTypes...)
}

func IsFollowupInteractionEventType(eventType string) bool {
	switch strings.TrimSpace(eventType) {
	case InteractionEventSpeech, InteractionEventGift, InteractionEventShowItem, InteractionEventTradeRequest, InteractionEventThreaten:
		return true
	default:
		return false
	}
}

func RequiresFollowupInteraction(stepType string, suggested *SuggestedInteraction) bool {
	if suggested != nil && strings.TrimSpace(suggested.EventType) != "" {
		return true
	}
	return IsFollowupInteractionEventType(stepType)
}
