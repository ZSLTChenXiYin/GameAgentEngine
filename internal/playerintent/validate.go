package playerintent

import (
	"fmt"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/workerstate"
	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

func Validate(view *workerstate.StateView, payload *sdk.PlayerIntentInterpretation) ValidationResult {
	if view == nil || payload == nil || payload.Intent == nil {
		return ValidationResult{OK: false, Issues: []ValidationIssue{{Code: "invalid_payload", Message: "player intent payload required"}}}
	}
	steps := intentSteps(payload.Intent)
	issues := make([]ValidationIssue, 0)
	for idx, step := range steps {
		issues = append(issues, validateStep(view, payload.Intent, step, idx)...)
	}
	for _, item := range payload.MissingFacts {
		issues = append(issues, ValidationIssue{
			Code:    "missing_fact_declared",
			Message: firstNonEmpty(item.Reason, fmt.Sprintf("missing fact: %s", item.Type)),
			MissingFact: &sdk.MissingFact{
				Type:   item.Type,
				NodeID: item.NodeID,
				ItemID: item.ItemID,
				TaskID: item.TaskID,
				Reason: item.Reason,
			},
		})
	}
	return ValidationResult{OK: len(issues) == 0, Issues: issues}
}

func validateStep(view *workerstate.StateView, intent *sdk.PlayerIntent, step sdk.PlayerIntentStep, index int) []ValidationIssue {
	issues := make([]ValidationIssue, 0)
	for _, pre := range step.Preconditions {
		if issue, ok := validatePrecondition(view, intent, step, pre, index); ok {
			issues = append(issues, issue)
		}
	}
	stepType := strings.TrimSpace(step.Type)
	actorID := strings.TrimSpace(intent.ActorNodeID)
	targetID := firstNonEmpty(strings.TrimSpace(step.TargetNodeID), strings.TrimSpace(intent.TargetNodeID))
	sceneID := firstNonEmpty(strings.TrimSpace(step.SceneNodeID), strings.TrimSpace(intent.SceneNodeID))
	switch stepType {
	case "show_item", "gift", "use_item":
		if strings.TrimSpace(step.ItemID) == "" {
			issues = append(issues, ValidationIssue{StepIndex: index, Code: "item_required", Message: "item_id required"})
		} else if !view.ItemPresentOnActor(actorID, step.ItemID) {
			issues = append(issues, ValidationIssue{StepIndex: index, Code: "item_missing", Message: fmt.Sprintf("actor %s does not possess item %s", actorID, step.ItemID), MissingFact: &sdk.MissingFact{Type: "item_presence", NodeID: actorID, ItemID: step.ItemID, Reason: "actor inventory does not contain required item"}})
		}
	case "speech", "trade_request", "threaten":
		if strings.TrimSpace(targetID) == "" {
			issues = append(issues, ValidationIssue{StepIndex: index, Code: "target_required", Message: "target_node_id required"})
		}
	}
	if sceneID != "" && targetID != "" {
		if targetScene, ok := view.ActorLocation(targetID); !ok || strings.TrimSpace(targetScene) != sceneID {
			issues = append(issues, ValidationIssue{StepIndex: index, Code: "target_not_in_scene", Message: fmt.Sprintf("target %s is not in scene %s", targetID, sceneID), MissingFact: &sdk.MissingFact{Type: "target_location", NodeID: targetID, Reason: "target not present in required scene"}})
		}
	}
	return issues
}

func validatePrecondition(view *workerstate.StateView, intent *sdk.PlayerIntent, step sdk.PlayerIntentStep, pre sdk.PlayerIntentPrecondition, index int) (ValidationIssue, bool) {
	actorID := firstNonEmpty(strings.TrimSpace(pre.ActorNodeID), strings.TrimSpace(intent.ActorNodeID))
	targetID := firstNonEmpty(strings.TrimSpace(pre.TargetNodeID), strings.TrimSpace(step.TargetNodeID), strings.TrimSpace(intent.TargetNodeID))
	sceneID := firstNonEmpty(strings.TrimSpace(pre.SceneNodeID), strings.TrimSpace(step.SceneNodeID), strings.TrimSpace(intent.SceneNodeID))
	switch strings.TrimSpace(pre.Type) {
	case "same_scene":
		actorScene, actorOK := view.ActorLocation(actorID)
		targetScene, targetOK := view.ActorLocation(targetID)
		if !actorOK || !targetOK || actorScene == "" || targetScene == "" || actorScene != targetScene {
			return ValidationIssue{StepIndex: index, Code: "same_scene_failed", Message: fmt.Sprintf("actor %s and target %s are not in the same scene", actorID, targetID), MissingFact: &sdk.MissingFact{Type: "target_location", NodeID: targetID, Reason: "same_scene precondition failed"}}, true
		}
	case "target_present":
		if sceneID == "" {
			if currentScene, ok := view.ActorLocation(actorID); ok {
				sceneID = currentScene
			}
		}
		if targetScene, ok := view.ActorLocation(targetID); !ok || strings.TrimSpace(targetScene) != strings.TrimSpace(sceneID) {
			return ValidationIssue{StepIndex: index, Code: "target_present_failed", Message: fmt.Sprintf("target %s is not present in scene %s", targetID, sceneID), MissingFact: &sdk.MissingFact{Type: "target_location", NodeID: targetID, Reason: "target_present precondition failed"}}, true
		}
	case "item_present":
		if !view.ItemPresentOnActor(actorID, pre.ItemID) {
			return ValidationIssue{StepIndex: index, Code: "item_present_failed", Message: fmt.Sprintf("actor %s does not have item %s", actorID, pre.ItemID), MissingFact: &sdk.MissingFact{Type: "item_presence", NodeID: actorID, ItemID: pre.ItemID, Reason: "item_present precondition failed"}}, true
		}
	case "money_at_least":
		money, ok := view.ActorMoney(actorID)
		minimum := intArg(pre.Args, "amount")
		if !ok || money < minimum {
			return ValidationIssue{StepIndex: index, Code: "money_at_least_failed", Message: fmt.Sprintf("actor %s money %d below required %d", actorID, money, minimum), MissingFact: &sdk.MissingFact{Type: "wallet_state", NodeID: actorID, Reason: "money_at_least precondition failed"}}, true
		}
	case "task_status":
		status, _, ok := view.QuestStatus(pre.TaskID)
		if !ok || !strings.EqualFold(strings.TrimSpace(status), strings.TrimSpace(pre.Expected)) {
			return ValidationIssue{StepIndex: index, Code: "task_status_failed", Message: fmt.Sprintf("task %s status mismatch", pre.TaskID), MissingFact: &sdk.MissingFact{Type: "task_state", TaskID: pre.TaskID, Reason: "task_status precondition failed"}}, true
		}
	case "scene_flag":
		scene, ok := view.Scene(sceneID)
		flagName := firstString(pre.Args, "flag")
		expected := firstString(pre.Args, "value")
		if !ok || scene == nil || !matchesFlag(scene.Flags, flagName, expected) {
			return ValidationIssue{StepIndex: index, Code: "scene_flag_failed", Message: fmt.Sprintf("scene %s flag %s mismatch", sceneID, flagName), MissingFact: &sdk.MissingFact{Type: "scene_state", NodeID: sceneID, Reason: "scene_flag precondition failed"}}, true
		}
	case "location_accessible":
		if sceneID == "" {
			return ValidationIssue{StepIndex: index, Code: "location_accessible_failed", Message: "scene_node_id required for location_accessible", MissingFact: &sdk.MissingFact{Type: "player_location", Reason: "scene id required"}}, true
		}
		if _, ok := view.Scene(sceneID); !ok {
			return ValidationIssue{StepIndex: index, Code: "location_accessible_failed", Message: fmt.Sprintf("scene %s not found", sceneID), MissingFact: &sdk.MissingFact{Type: "scene_state", NodeID: sceneID, Reason: "scene not found"}}, true
		}
	}
	return ValidationIssue{}, false
}

func intentSteps(intent *sdk.PlayerIntent) []sdk.PlayerIntentStep {
	if intent == nil {
		return nil
	}
	if strings.TrimSpace(intent.Type) == "composite" {
		return append([]sdk.PlayerIntentStep(nil), intent.Steps...)
	}
	return []sdk.PlayerIntentStep{{
		Type:          intent.Type,
		TargetNodeID:  intent.TargetNodeID,
		SceneNodeID:   intent.SceneNodeID,
		ItemID:        firstIntentItemID(intent),
		Content:       firstIntentContent(intent),
		Args:          firstIntentArgs(intent),
		Preconditions: firstIntentPreconditions(intent),
	}}
}

func firstIntentItemID(intent *sdk.PlayerIntent) string {
	if intent == nil || len(intent.Steps) == 0 {
		return ""
	}
	return intent.Steps[0].ItemID
}

func firstIntentContent(intent *sdk.PlayerIntent) string {
	if intent == nil || len(intent.Steps) == 0 {
		return ""
	}
	return intent.Steps[0].Content
}

func firstIntentArgs(intent *sdk.PlayerIntent) map[string]any {
	if intent == nil || len(intent.Steps) == 0 {
		return nil
	}
	return intent.Steps[0].Args
}

func firstIntentPreconditions(intent *sdk.PlayerIntent) []sdk.PlayerIntentPrecondition {
	if intent == nil || len(intent.Steps) == 0 {
		return nil
	}
	return append([]sdk.PlayerIntentPrecondition(nil), intent.Steps[0].Preconditions...)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstString(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	if value, ok := values[key].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func intArg(values map[string]any, key string) int {
	if values == nil {
		return 0
	}
	switch value := values[key].(type) {
	case int:
		return value
	case float64:
		return int(value)
	default:
		return 0
	}
}

func matchesFlag(flags map[string]any, key, expected string) bool {
	if flags == nil || strings.TrimSpace(key) == "" {
		return false
	}
	value, ok := flags[key]
	if !ok {
		return false
	}
	if strings.TrimSpace(expected) == "" {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(fmt.Sprint(value)), strings.TrimSpace(expected))
}
