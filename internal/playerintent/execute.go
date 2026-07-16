package playerintent

import (
	"fmt"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/workerstate"
	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

func Execute(state *workerstate.WorldState, payload *sdk.PlayerIntentInterpretation) (*ExecutionResult, error) {
	if state == nil || payload == nil || payload.Intent == nil {
		return nil, fmt.Errorf("player intent payload required")
	}
	view := workerstate.NewStateView(state)
	validation := Validate(view, payload)
	if !validation.OK {
		return nil, fmt.Errorf("player intent validation failed")
	}
	result := &ExecutionResult{ActorNodeID: payload.Intent.ActorNodeID, SceneNodeID: payload.Intent.SceneNodeID}
	steps := intentExecutionSteps(payload.Intent)
	for idx, step := range steps {
		outcome, err := executeStep(state, payload.Intent, step, idx)
		if err != nil {
			return result, err
		}
		result.Outcomes = append(result.Outcomes, outcome)
	}
	return result, nil
}

func executeStep(state *workerstate.WorldState, intent *sdk.PlayerIntent, step sdk.PlayerIntentStep, index int) (StepOutcome, error) {
	actorID := strings.TrimSpace(intent.ActorNodeID)
	targetID := firstNonEmpty(strings.TrimSpace(step.TargetNodeID), strings.TrimSpace(intent.TargetNodeID))
	sceneID := firstNonEmpty(strings.TrimSpace(step.SceneNodeID), strings.TrimSpace(intent.SceneNodeID))
	switch strings.TrimSpace(step.Type) {
	case sdk.PlayerIntentTypeGift:
		if err := transferInventoryItem(state, actorID, targetID, step.ItemID, 1); err != nil {
			return StepOutcome{}, err
		}
		return StepOutcome{StepIndex: index, Type: step.Type, Applied: true, Summary: fmt.Sprintf("gifted %s to %s", step.ItemID, targetID)}, nil
	case sdk.PlayerIntentTypeMove:
		if err := moveActorToScene(state, actorID, sceneID); err != nil {
			return StepOutcome{}, err
		}
		return StepOutcome{StepIndex: index, Type: step.Type, Applied: true, Summary: fmt.Sprintf("moved %s to %s", actorID, sceneID)}, nil
	case sdk.PlayerIntentTypeUseItem:
		return StepOutcome{StepIndex: index, Type: step.Type, Applied: true, Summary: fmt.Sprintf("validated use_item for %s", step.ItemID)}, nil
	case sdk.PlayerIntentTypeSpeech, sdk.PlayerIntentTypeShowItem, sdk.PlayerIntentTypeTradeRequest, sdk.PlayerIntentTypeThreaten, sdk.PlayerIntentTypeInspect:
		return StepOutcome{StepIndex: index, Type: step.Type, Applied: true, Summary: fmt.Sprintf("validated %s for target %s", step.Type, targetID)}, nil
	default:
		return StepOutcome{}, fmt.Errorf("unsupported player intent step type: %s", step.Type)
	}
}

func intentExecutionSteps(intent *sdk.PlayerIntent) []sdk.PlayerIntentStep {
	if intent == nil {
		return nil
	}
	if strings.TrimSpace(intent.Type) == sdk.PlayerIntentTypeComposite {
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

func transferInventoryItem(state *workerstate.WorldState, fromActorID, toActorID, itemID string, quantity int) error {
	if state == nil {
		return fmt.Errorf("state required")
	}
	from := state.Actors[strings.TrimSpace(fromActorID)]
	to := state.Actors[strings.TrimSpace(toActorID)]
	if from == nil || to == nil {
		return fmt.Errorf("source or target actor not found")
	}
	entryIndex := -1
	for i, entry := range from.Inventory {
		if strings.EqualFold(strings.TrimSpace(entry.ItemID), strings.TrimSpace(itemID)) && entry.Quantity >= quantity {
			entryIndex = i
			break
		}
	}
	if entryIndex < 0 {
		return fmt.Errorf("item %s not available on actor %s", itemID, fromActorID)
	}
	from.Inventory[entryIndex].Quantity -= quantity
	if from.Inventory[entryIndex].Quantity <= 0 {
		from.Inventory = append(from.Inventory[:entryIndex], from.Inventory[entryIndex+1:]...)
	}
	merged := false
	for i := range to.Inventory {
		if strings.EqualFold(strings.TrimSpace(to.Inventory[i].ItemID), strings.TrimSpace(itemID)) {
			to.Inventory[i].Quantity += quantity
			merged = true
			break
		}
	}
	if !merged {
		to.Inventory = append(to.Inventory, workerstate.InventoryEntry{ItemID: itemID, Quantity: quantity})
	}
	if item := state.Items[strings.TrimSpace(itemID)]; item != nil {
		item.OwnerID = to.ID
		item.SceneID = ""
	}
	return nil
}

func moveActorToScene(state *workerstate.WorldState, actorID, sceneID string) error {
	if state == nil {
		return fmt.Errorf("state required")
	}
	actor := state.Actors[strings.TrimSpace(actorID)]
	if actor == nil {
		return fmt.Errorf("actor %s not found", actorID)
	}
	if _, ok := state.Scenes[strings.TrimSpace(sceneID)]; !ok {
		return fmt.Errorf("scene %s not found", sceneID)
	}
	fromSceneID := strings.TrimSpace(actor.LocationID)
	if fromSceneID != "" {
		removeOccupant(state.Scenes[fromSceneID], actor.ID)
	}
	actor.LocationID = strings.TrimSpace(sceneID)
	appendOccupant(state.Scenes[sceneID], actor.ID)
	return nil
}

func removeOccupant(scene *workerstate.SceneState, actorID string) {
	if scene == nil {
		return
	}
	filtered := scene.Occupants[:0]
	for _, item := range scene.Occupants {
		if strings.TrimSpace(item) == strings.TrimSpace(actorID) {
			continue
		}
		filtered = append(filtered, item)
	}
	scene.Occupants = filtered
}

func appendOccupant(scene *workerstate.SceneState, actorID string) {
	if scene == nil {
		return
	}
	for _, item := range scene.Occupants {
		if strings.TrimSpace(item) == strings.TrimSpace(actorID) {
			return
		}
	}
	scene.Occupants = append(scene.Occupants, actorID)
}
