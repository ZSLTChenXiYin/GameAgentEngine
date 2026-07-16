package workercli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/playerintent"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/workerstate"
	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

type playSession struct {
	client          *sdk.Client
	view            *workerstate.StateView
	worldID         string
	playerNodeID    string
	currentSceneID  string
	currentTargetID string
	sessionID       string
	turnIndex       int
}

type playCommand struct {
	Name string
	Args string
	Raw  string
}

type playInteractionSpec struct {
	Mode          string
	AudienceScope string
	EventType     string
	ItemID        string
	Input         string
	TargetNodeID  string
	Participants  []string
	EventArgs     map[string]any
}

func (a *app) runPlay() error {
	statePath := strings.TrimSpace(a.cfg.StateFile)
	if statePath == "" {
		return errors.New("play mode requires --state-file")
	}
	state, err := workerstate.LoadFile(statePath)
	if err != nil {
		return err
	}
	a.setAuthorityState(state)
	view := a.authorityView()
	if view == nil {
		return errors.New("failed to initialize authority state")
	}
	worldID := strings.TrimSpace(a.cfg.PlayWorldID)
	if worldID == "" {
		worldID = strings.TrimSpace(view.WorldID())
	}
	if worldID == "" {
		return errors.New("play mode requires --world-id or state file world_id")
	}
	playerID, err := a.resolvePlayPlayerNodeID(view)
	if err != nil {
		return err
	}
	sceneID, ok := view.ActorLocation(playerID)
	if !ok {
		return fmt.Errorf("player %s has no location_id in state file", playerID)
	}
	sessionID := strings.TrimSpace(a.cfg.PlaySessionID)
	if sessionID == "" {
		sessionID = fmt.Sprintf("play-%s", strings.TrimSpace(playerID))
	}
	s := &playSession{
		client:         sdk.NewClient(a.cfg.EngineBaseURL, a.cfg.EngineAPIKey),
		view:           view,
		worldID:        worldID,
		playerNodeID:   playerID,
		currentSceneID: sceneID,
		sessionID:      sessionID,
	}

	var cancel context.CancelFunc
	if a.cfg.PlayAutoWorker {
		var ctx context.Context
		ctx, cancel = context.WithCancel(context.Background())
		defer cancel()
		go a.runPlayWorkerLoop(ctx)
	}

	fmt.Printf("进入 play 模式。世界=%s 玩家=%s 场景=%s\n", s.worldID, s.playerNodeID, s.currentSceneID)
	fmt.Println("输入 /+help 查看命令；直接输入文本会发给当前对话目标。")
	fmt.Println(s.renderSceneSummary())

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("play> ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return err
			}
			fmt.Println()
			return nil
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "/") {
			cmd := parsePlayCommand(line)
			exit, err := a.executePlayCommand(s, cmd)
			if err != nil {
				fmt.Printf("错误: %v\n", err)
				continue
			}
			if exit {
				return nil
			}
			continue
		}
		if err := a.runPlayDialogue(s, line); err != nil {
			fmt.Printf("错误: %v\n", err)
		}
	}
}

func (a *app) runPlayWorkerLoop(ctx context.Context) {
	ticker := time.NewTicker(a.cfg.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		_, _, err := a.processOnePendingTask()
		if err != nil {
			a.logJSON("play_worker_error", map[string]any{"error": err.Error()})
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (a *app) resolvePlayPlayerNodeID(view *workerstate.StateView) (string, error) {
	if id := strings.TrimSpace(a.cfg.PlayPlayerNodeID); id != "" {
		if actorID, ok := view.FindActorIDByName(id); ok {
			return actorID, nil
		}
		return "", fmt.Errorf("player node %q not found in state file", id)
	}
	for _, actor := range view.Actors() {
		if actor != nil && strings.EqualFold(strings.TrimSpace(actor.Kind), "player") {
			return actor.ID, nil
		}
	}
	actors := view.Actors()
	if len(actors) == 1 {
		return actors[0].ID, nil
	}
	return "", errors.New("play mode requires --player-node-id when state file has multiple actors and none is marked kind=player")
}

func parsePlayCommand(line string) playCommand {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return playCommand{}
	}
	body := strings.TrimPrefix(trimmed, "/")
	body = strings.TrimSpace(strings.TrimPrefix(body, "+"))
	parts := strings.Fields(body)
	if len(parts) == 0 {
		return playCommand{Raw: trimmed}
	}
	name := strings.ToLower(strings.TrimSpace(parts[0]))
	args := ""
	if len(parts) > 1 {
		args = strings.TrimSpace(strings.Join(parts[1:], " "))
	}
	return playCommand{Name: name, Args: args, Raw: trimmed}
}

func (a *app) executePlayCommand(s *playSession, cmd playCommand) (bool, error) {
	switch cmd.Name {
	case "quit", "exit":
		fmt.Println("退出 play 模式。")
		return true, nil
	case "help":
		fmt.Println(playHelpText())
		return false, nil
	case "look":
		s.refreshView(a)
		fmt.Println(s.renderSceneSummary())
		return false, nil
	case "who":
		s.refreshView(a)
		fmt.Println(s.renderOccupants())
		return false, nil
	case "state":
		s.refreshView(a)
		fmt.Println(s.renderPlayerState())
		return false, nil
	case "talk":
		return false, a.setPlayTarget(s, cmd.Args)
	case "target":
		fmt.Println(s.renderTargetStatus())
		return false, nil
	case "clear_target", "untalk":
		s.currentTargetID = ""
		fmt.Println("已清除当前对话目标。")
		return false, nil
	case "room":
		s.refreshView(a)
		fmt.Println(s.renderRoomState())
		return false, nil
	case "say":
		return false, a.runPlaySay(s, cmd.Args)
	case "ask":
		return false, a.runPlayAsk(s, cmd.Args)
	case "act":
		return false, a.runPlayAct(s, cmd.Args)
	case "gift":
		return false, a.runPlayGift(s, cmd.Args)
	case "show_item", "show":
		return false, a.runPlayShowItem(s, cmd.Args)
	case "trade":
		return false, a.runPlayTrade(s, cmd.Args)
	case "threaten":
		return false, a.runPlayThreaten(s, cmd.Args)
	default:
		return false, fmt.Errorf("unknown command %q", cmd.Name)
	}
}

func (a *app) setPlayTarget(s *playSession, arg string) error {
	s.refreshView(a)
	target, err := s.resolveSceneActor(arg)
	if err != nil {
		return err
	}
	if target.ID == s.playerNodeID {
		return errors.New("cannot set player as talk target")
	}
	s.currentTargetID = target.ID
	label := target.ID
	if strings.TrimSpace(target.Name) != "" {
		label = fmt.Sprintf("%s (%s)", target.Name, target.ID)
	}
	fmt.Printf("当前对话目标: %s\n", label)
	return nil
}

func (a *app) runPlayDialogue(s *playSession, input string) error {
	targetID, err := s.resolveDirectDialogueTarget()
	if err != nil {
		return err
	}
	s.currentTargetID = targetID
	return a.invokePlayInteraction(s, playInteractionSpec{
		Mode:          sdk.InteractionModeDirectDialogue,
		AudienceScope: sdk.InteractionAudiencePrivate,
		EventType:     sdk.InteractionEventSpeech,
		Input:         input,
		TargetNodeID:  s.currentTargetID,
		Participants:  []string{s.playerNodeID, s.currentTargetID},
	})
}

func (a *app) runPlayAct(s *playSession, args string) error {
	text := strings.TrimSpace(args)
	if text == "" {
		return errors.New("/act requires action text")
	}
	s.refreshView(a)
	participants := s.sceneParticipantIDs()
	targetID := strings.TrimSpace(s.currentTargetID)
	if targetID == "" {
		if autoTarget, _, err := s.resolveGroupChatTarget(""); err == nil {
			targetID = autoTarget
		}
	}
	resp, err := s.client.InterpretPlayerInput(&sdk.PlayerInputInterpretRequest{
		WorldID:            s.worldID,
		PlayerNodeID:       s.playerNodeID,
		SceneNodeID:        s.currentSceneID,
		TargetNodeID:       targetID,
		SessionID:          s.sessionID,
		Message:            text,
		ParticipantNodeIDs: participants,
		Context: &sdk.InvokeContext{
			IncludeRelatedNodes: a.cfg.PlayIncludeRelated,
			PipelineMode:        a.cfg.PlayPipelineMode,
			DynamicInterfaces: []sdk.DynamicInterface{
				sdk.NewAuthorityDataRequest("play_authority", 8),
			},
		},
	})
	if err != nil {
		return err
	}
	resp, err = a.resolvePlayResponse(resp)
	if err != nil {
		return err
	}
	if resp == nil || resp.PlayerIntent == nil || resp.PlayerIntent.Intent == nil {
		if resp != nil && strings.TrimSpace(resp.Reply) != "" {
			fmt.Printf("[system] %s\n", strings.TrimSpace(resp.Reply))
		}
		return errors.New("player intent interpretation missing")
	}
	validation := playerintent.Validate(s.view, resp.PlayerIntent)
	if !validation.OK {
		for _, issue := range validation.Issues {
			fmt.Printf("[system] validation: %s\n", issue.Message)
		}
		return errors.New("player intent validation failed")
	}
	a.authorityMu.Lock()
	result, execErr := playerintent.Execute(a.authority, resp.PlayerIntent)
	a.authorityMu.Unlock()
	if execErr != nil {
		return execErr
	}
	s.refreshView(a)
	a.printPlayExecutionResult(s, result)
	bridge, err := playerintent.BuildInteractionSpec(resp.PlayerIntent, s.playerNodeID, s.currentSceneID)
	if err != nil {
		if result != nil {
			fmt.Printf("[system] executed %d step(s) without follow-up interaction\n", len(result.Outcomes))
			return nil
		}
		return err
	}
	if len(participants) > 0 {
		bridge.Participants = uniqueParticipantIDs(participants, bridge.Participants...)
	}
	s.currentTargetID = bridge.TargetNodeID
	return a.invokePlayInteraction(s, playInteractionSpec{
		Mode:          bridge.Mode,
		AudienceScope: bridge.AudienceScope,
		EventType:     bridge.EventType,
		ItemID:        bridge.ItemID,
		Input:         bridge.Input,
		TargetNodeID:  bridge.TargetNodeID,
		Participants:  bridge.Participants,
	})
}

func (a *app) runPlaySay(s *playSession, args string) error {
	text := strings.TrimSpace(args)
	if text == "" {
		return errors.New("/say requires message text")
	}
	s.refreshView(a)
	targetID, participants, err := s.resolveGroupChatTarget("")
	if err != nil {
		return err
	}
	return a.invokePlayInteraction(s, playInteractionSpec{
		Mode:          sdk.InteractionModeGroupChat,
		AudienceScope: sdk.InteractionAudiencePublic,
		EventType:     sdk.InteractionEventSpeech,
		Input:         text,
		TargetNodeID:  targetID,
		Participants:  participants,
	})
}

func (a *app) runPlayAsk(s *playSession, args string) error {
	parts := strings.Fields(args)
	if len(parts) < 2 {
		return errors.New("/ask requires: /ask <npc> <message>")
	}
	targetArg := parts[0]
	text := strings.TrimSpace(strings.Join(parts[1:], " "))
	if text == "" {
		return errors.New("/ask requires message text")
	}
	s.refreshView(a)
	targetID, participants, err := s.resolveGroupChatTarget(targetArg)
	if err != nil {
		return err
	}
	s.currentTargetID = targetID
	return a.invokePlayInteraction(s, playInteractionSpec{
		Mode:          sdk.InteractionModeGroupChat,
		AudienceScope: sdk.InteractionAudiencePublic,
		EventType:     sdk.InteractionEventSpeech,
		Input:         text,
		TargetNodeID:  targetID,
		Participants:  participants,
	})
}

func (a *app) runPlayGift(s *playSession, args string) error {
	parts := strings.Fields(args)
	if len(parts) < 2 {
		return errors.New("/gift requires: /gift <npc> <item>")
	}
	s.refreshView(a)
	target, err := s.resolveSceneActor(parts[0])
	if err != nil {
		return err
	}
	itemID, itemLabel, err := s.resolvePlayerInventoryItem(strings.Join(parts[1:], " "))
	if err != nil {
		return err
	}
	if err := a.transferInventoryItem(s.playerNodeID, target.ID, itemID, 1); err != nil {
		return err
	}
	s.refreshView(a)
	s.currentTargetID = target.ID
	fmt.Printf("[system] 你将 %s 交给了 %s。\n", itemLabel, s.actorDisplayName(target.ID))
	return a.invokePlayInteraction(s, playInteractionSpec{
		Mode:          sdk.InteractionModeGiftResponse,
		AudienceScope: sdk.InteractionAudiencePrivate,
		EventType:     sdk.InteractionEventGift,
		ItemID:        itemID,
		Input:         fmt.Sprintf("玩家向你赠送了物品 %s。请基于权威状态和场景事实回应。", itemID),
		TargetNodeID:  target.ID,
		Participants:  []string{s.playerNodeID, target.ID},
	})
}

func (a *app) runPlayShowItem(s *playSession, args string) error {
	parts := strings.Fields(args)
	if len(parts) < 2 {
		return errors.New("/show_item requires: /show_item <npc> <item>")
	}
	s.refreshView(a)
	target, err := s.resolveSceneActor(parts[0])
	if err != nil {
		return err
	}
	itemID, itemLabel, err := s.resolvePlayerInventoryItem(strings.Join(parts[1:], " "))
	if err != nil {
		return err
	}
	s.currentTargetID = target.ID
	return a.invokePlayInteraction(s, playInteractionSpec{
		Mode:          sdk.InteractionModeDirectDialogue,
		AudienceScope: sdk.InteractionAudiencePrivate,
		EventType:     sdk.InteractionEventShowItem,
		ItemID:        itemID,
		Input:         fmt.Sprintf("玩家向你展示了物品 %s。请基于权威状态和场景事实回应。", itemLabel),
		TargetNodeID:  target.ID,
		Participants:  []string{s.playerNodeID, target.ID},
	})
}

func (a *app) runPlayTrade(s *playSession, args string) error {
	targetArg := strings.TrimSpace(args)
	if targetArg == "" {
		targetArg = strings.TrimSpace(s.currentTargetID)
	}
	if targetArg == "" {
		return errors.New("/trade requires a target npc or an active talk target")
	}
	s.refreshView(a)
	target, err := s.resolveSceneActor(targetArg)
	if err != nil {
		return err
	}
	s.currentTargetID = target.ID
	return a.invokePlayInteraction(s, playInteractionSpec{
		Mode:          sdk.InteractionModeTradeDialogue,
		AudienceScope: sdk.InteractionAudiencePrivate,
		EventType:     sdk.InteractionEventTradeRequest,
		Input:         "玩家想和你谈交易或议价。请基于当前权威状态、库存、金钱和关系回应。",
		TargetNodeID:  target.ID,
		Participants:  []string{s.playerNodeID, target.ID},
	})
}

func (a *app) runPlayThreaten(s *playSession, args string) error {
	targetArg := strings.TrimSpace(args)
	if targetArg == "" {
		targetArg = strings.TrimSpace(s.currentTargetID)
	}
	if targetArg == "" {
		return errors.New("/threaten requires a target npc or an active talk target")
	}
	s.refreshView(a)
	target, err := s.resolveSceneActor(targetArg)
	if err != nil {
		return err
	}
	s.currentTargetID = target.ID
	return a.invokePlayInteraction(s, playInteractionSpec{
		Mode:          sdk.InteractionModeDirectDialogue,
		AudienceScope: sdk.InteractionAudiencePrivate,
		EventType:     sdk.InteractionEventThreaten,
		Input:         "玩家正在以威胁性的方式逼迫你回应。请基于场景、关系和即时风险判断回应。",
		TargetNodeID:  target.ID,
		Participants:  []string{s.playerNodeID, target.ID},
	})
}

func (a *app) invokePlayInteraction(s *playSession, spec playInteractionSpec) error {
	s.refreshView(a)
	if strings.TrimSpace(spec.TargetNodeID) == "" {
		return errors.New("interaction target is required")
	}
	s.turnIndex++
	ctx := &sdk.InvokeContext{
		IncludeRelatedNodes: a.cfg.PlayIncludeRelated,
		DynamicInterfaces: []sdk.DynamicInterface{
			sdk.NewAuthorityDataRequest("play_authority", 8),
		},
	}
	if mode := strings.TrimSpace(a.cfg.PlayPipelineMode); mode != "" {
		ctx.PipelineMode = mode
	}
	resp, err := s.client.ExecuteInteraction(&sdk.InteractionExecuteRequest{
		WorldID:            s.worldID,
		ActorNodeID:        s.playerNodeID,
		TargetNodeID:       spec.TargetNodeID,
		SceneNodeID:        s.currentSceneID,
		SessionID:          s.sessionID,
		TaskType:           "npc_dialogue",
		Message:            spec.Input,
		ParticipantNodeIDs: uniqueParticipantIDs(spec.Participants, s.playerNodeID, spec.TargetNodeID),
		Mode:               spec.Mode,
		AudienceScope:      spec.AudienceScope,
		TurnIndex:          s.turnIndex,
		Event: &sdk.InteractionEvent{
			Type:   spec.EventType,
			ItemID: spec.ItemID,
			Args:   spec.EventArgs,
		},
		Context: ctx,
	})
	if err != nil {
		return err
	}
	resp, err = a.resolvePlayResponse(resp)
	if err != nil {
		return err
	}
	label := s.actorDisplayName(spec.TargetNodeID)
	if strings.TrimSpace(resp.Reply) == "" {
		fmt.Printf("[%s] （无文本回复）\n", label)
	} else {
		fmt.Printf("[%s] %s\n", label, strings.TrimSpace(resp.Reply))
	}
	if len(resp.ActionCalls) > 0 {
		fmt.Printf("[system] 引擎产生了 %d 个 action_calls，当前 play v1 仅展示，不在本地直接落地。\n", len(resp.ActionCalls))
	}
	return nil
}

func (a *app) resolvePlayResponse(resp *sdk.InvokeResponse) (*sdk.InvokeResponse, error) {
	current := resp
	for current != nil && current.HasPendingDataRequest() {
		if !a.cfg.PlayAutoWorker {
			return current, errors.New("play response paused for authority data; enable --auto-worker")
		}
		processed, ok, err := a.processOnePendingTaskDetailed()
		if err != nil {
			return nil, err
		}
		if !ok || processed == nil || processed.Callback == nil || processed.Callback.Resumed == nil {
			return current, errors.New("authority callback did not resume play response")
		}
		current = processed.Callback.Resumed
	}
	return current, nil
}

func (a *app) printPlayExecutionResult(s *playSession, result *playerintent.ExecutionResult) {
	if result == nil || len(result.Outcomes) == 0 {
		return
	}
	for _, outcome := range result.Outcomes {
		message := strings.TrimSpace(outcome.Summary)
		if message == "" {
			message = fmt.Sprintf("%s applied", strings.TrimSpace(outcome.Type))
		}
		fmt.Printf("[system] %s\n", message)
	}
	if strings.TrimSpace(s.currentSceneID) != "" {
		if actor, ok := s.view.Actor(s.playerNodeID); ok && actor != nil && strings.TrimSpace(actor.LocationID) != strings.TrimSpace(s.currentSceneID) {
			s.currentSceneID = actor.LocationID
			fmt.Println(s.renderSceneSummary())
		}
	}
}

func uniqueParticipantIDs(explicit []string, defaults ...string) []string {
	return engine.CanonicalParticipantNodeIDs(explicit, defaults)
}

func (a *app) transferInventoryItem(fromActorID, toActorID, itemID string, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	a.authorityMu.Lock()
	defer a.authorityMu.Unlock()
	state := a.authority
	if state == nil {
		return errors.New("authority state not initialized")
	}
	from := state.Actors[strings.TrimSpace(fromActorID)]
	to := state.Actors[strings.TrimSpace(toActorID)]
	if from == nil || to == nil {
		return errors.New("source or target actor not found in authority state")
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

func (s *playSession) refreshView(a *app) {
	if view := a.authorityView(); view != nil {
		s.view = view
		if locationID, ok := view.ActorLocation(s.playerNodeID); ok {
			s.currentSceneID = locationID
		}
	}
}

func (s *playSession) resolveSceneActor(arg string) (*workerstate.ActorState, error) {
	trimmed := strings.TrimSpace(arg)
	if trimmed == "" {
		return nil, errors.New("target actor name or id is required")
	}
	id, ok := s.view.FindActorIDByName(trimmed)
	if !ok {
		return nil, fmt.Errorf("actor %q not found", trimmed)
	}
	actor, ok := s.view.Actor(id)
	if !ok || actor == nil {
		return nil, fmt.Errorf("actor %q not found", trimmed)
	}
	if strings.TrimSpace(actor.LocationID) != strings.TrimSpace(s.currentSceneID) {
		return nil, fmt.Errorf("actor %s is not in current scene %s", actor.ID, s.currentSceneID)
	}
	return actor, nil
}

func (s *playSession) resolvePlayerInventoryItem(arg string) (string, string, error) {
	trimmed := strings.TrimSpace(arg)
	if trimmed == "" {
		return "", "", errors.New("item name or id is required")
	}
	itemID, ok := s.view.FindItemIDByName(trimmed)
	if !ok {
		itemID = trimmed
	}
	entry, ok := s.view.ActorInventoryEntry(s.playerNodeID, itemID)
	if !ok || entry == nil || entry.Quantity <= 0 {
		return "", "", fmt.Errorf("player does not have item %q", trimmed)
	}
	label := entry.ItemID
	if item, ok := s.view.State().Items[itemID]; ok && item != nil && strings.TrimSpace(item.Name) != "" {
		label = item.Name
	}
	return itemID, label, nil
}

func (s *playSession) actorDisplayName(actorID string) string {
	actor, ok := s.view.Actor(actorID)
	if !ok || actor == nil {
		return actorID
	}
	if strings.TrimSpace(actor.Name) != "" {
		return actor.Name
	}
	return actor.ID
}

func (s *playSession) renderSceneSummary() string {
	scene, ok := s.view.Scene(s.currentSceneID)
	if !ok || scene == nil {
		return fmt.Sprintf("当前场景: %s", s.currentSceneID)
	}
	lines := []string{fmt.Sprintf("当前场景: %s (%s)", fallback(scene.Name, scene.ID), scene.ID)}
	if desc := strings.TrimSpace(scene.Description); desc != "" {
		lines = append(lines, desc)
	}
	if strings.TrimSpace(s.currentTargetID) != "" {
		lines = append(lines, fmt.Sprintf("当前目标: %s", s.actorDisplayName(s.currentTargetID)))
	}
	if summary := renderSceneFlags(scene.Flags); summary != "" {
		lines = append(lines, "场景状态: "+summary)
	}
	if prompt := s.renderScenePrompt(); prompt != "" {
		lines = append(lines, prompt)
	}
	lines = append(lines, s.renderOccupants())
	return strings.Join(lines, "\n")
}

func (s *playSession) renderOccupants() string {
	actors := s.view.ActorsAtScene(s.currentSceneID)
	if len(actors) == 0 {
		return "同场角色: 无"
	}
	parts := make([]string, 0, len(actors))
	for _, actor := range actors {
		if actor == nil {
			continue
		}
		label := fallback(actor.Name, actor.ID)
		if actor.ID == s.playerNodeID {
			label += " [you]"
		}
		if actor.ID == s.currentTargetID {
			label += " [target]"
		}
		parts = append(parts, fmt.Sprintf("- %s (%s)", label, actor.ID))
	}
	sort.Strings(parts)
	return "同场角色:\n" + strings.Join(parts, "\n")
}

func (s *playSession) renderPlayerState() string {
	actor, ok := s.view.Actor(s.playerNodeID)
	if !ok || actor == nil {
		return fmt.Sprintf("玩家: %s", s.playerNodeID)
	}
	lines := []string{
		fmt.Sprintf("玩家: %s (%s)", fallback(actor.Name, actor.ID), actor.ID),
		fmt.Sprintf("HP: %d/%d", actor.HP, actor.MaxHP),
		fmt.Sprintf("金钱: %d", actor.Money),
		fmt.Sprintf("位置: %s", fallbackSceneName(s.view, actor.LocationID)),
	}
	if len(actor.Inventory) > 0 {
		items := make([]string, 0, len(actor.Inventory))
		for _, entry := range actor.Inventory {
			label := entry.ItemID
			if item, ok := s.view.State().Items[entry.ItemID]; ok && item != nil && strings.TrimSpace(item.Name) != "" {
				label = item.Name
			}
			items = append(items, fmt.Sprintf("%s x%d", label, entry.Quantity))
		}
		lines = append(lines, "背包: "+strings.Join(items, ", "))
	} else {
		lines = append(lines, "背包: 空")
	}
	return strings.Join(lines, "\n")
}

func (s *playSession) renderTargetStatus() string {
	if strings.TrimSpace(s.currentTargetID) == "" {
		return "当前没有对话目标。"
	}
	return fmt.Sprintf("当前对话目标: %s", s.actorDisplayName(s.currentTargetID))
}

func (s *playSession) renderRoomState() string {
	participants := s.sceneParticipantIDs()
	parts := make([]string, 0, len(participants))
	for _, id := range participants {
		label := s.actorDisplayName(id)
		if id == s.playerNodeID {
			label += " [you]"
		}
		if id == s.currentTargetID {
			label += " [target]"
		}
		parts = append(parts, fmt.Sprintf("- %s (%s)", label, id))
	}
	sort.Strings(parts)
	if len(parts) == 0 {
		return "房间参与者: 无"
	}
	return "房间参与者:\n" + strings.Join(parts, "\n")
}

func (s *playSession) renderScenePrompt() string {
	participants := s.sceneParticipantIDs()
	npcs := make([]string, 0, len(participants))
	for _, id := range participants {
		if id == s.playerNodeID {
			continue
		}
		npcs = append(npcs, s.actorDisplayName(id))
	}
	if len(npcs) == 0 {
		return "提示: 当前场景没有可互动 NPC。"
	}
	sort.Strings(npcs)
	if strings.TrimSpace(s.currentTargetID) != "" {
		return fmt.Sprintf("提示: 直接输入文本可与 %s 对话，也可用 /+say 发起房间公开发言。", s.actorDisplayName(s.currentTargetID))
	}
	return fmt.Sprintf("提示: 可用 /+talk <npc> 选择目标。当前可互动对象: %s。", strings.Join(npcs, ", "))
}

func renderSceneFlags(flags map[string]any) string {
	if len(flags) == 0 {
		return ""
	}
	parts := make([]string, 0, len(flags))
	for key, value := range flags {
		parts = append(parts, fmt.Sprintf("%s=%v", key, value))
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

func (s *playSession) resolveDirectDialogueTarget() (string, error) {
	if strings.TrimSpace(s.currentTargetID) != "" {
		actor, ok := s.view.Actor(s.currentTargetID)
		if ok && actor != nil && actor.ID != s.playerNodeID && strings.TrimSpace(actor.LocationID) == strings.TrimSpace(s.currentSceneID) {
			return s.currentTargetID, nil
		}
	}
	actors := s.view.ActorsAtScene(s.currentSceneID)
	npcs := make([]*workerstate.ActorState, 0, len(actors))
	for _, actor := range actors {
		if actor == nil || actor.ID == s.playerNodeID {
			continue
		}
		npcs = append(npcs, actor)
	}
	if len(npcs) == 0 {
		return "", errors.New("no NPC is available for direct dialogue in the current scene")
	}
	if len(npcs) == 1 {
		return npcs[0].ID, nil
	}
	labels := make([]string, 0, len(npcs))
	for _, actor := range npcs {
		labels = append(labels, s.actorDisplayName(actor.ID))
	}
	sort.Strings(labels)
	return "", fmt.Errorf("multiple dialogue targets are available (%s); use /+talk <npc>", strings.Join(labels, ", "))
}

func (s *playSession) resolveGroupChatTarget(preferred string) (string, []string, error) {
	participants := s.sceneParticipantIDs()
	if len(participants) == 0 {
		return "", nil, errors.New("current scene has no participants")
	}
	if preferred != "" {
		target, err := s.resolveSceneActor(preferred)
		if err != nil {
			return "", nil, err
		}
		if target.ID == s.playerNodeID {
			return "", nil, errors.New("group chat target cannot be the player")
		}
		return target.ID, participants, nil
	}
	if strings.TrimSpace(s.currentTargetID) != "" {
		for _, id := range participants {
			if id == s.currentTargetID {
				return s.currentTargetID, participants, nil
			}
		}
	}
	for _, id := range participants {
		if id != s.playerNodeID {
			return id, participants, nil
		}
	}
	return "", nil, errors.New("no NPC participant available in current scene")
}

func (s *playSession) sceneParticipantIDs() []string {
	actors := s.view.ActorsAtScene(s.currentSceneID)
	ids := make([]string, 0, len(actors))
	for _, actor := range actors {
		if actor != nil {
			ids = append(ids, actor.ID)
		}
	}
	return uniqueParticipantIDs(ids)
}

func playHelpText() string {
	return strings.Join([]string{
		"/+help                        查看帮助",
		"/+look                        查看当前场景与同场角色",
		"/+who                         列出当前场景角色",
		"/+state                       查看玩家权威状态摘要",
		"/+talk <npc>                  选择当前对话目标",
		"/+target                      查看当前对话目标",
		"/+room                        查看当前房间/场景参与者",
		"/+clear_target                清除当前对话目标",
		"/+say <message>               向当前房间公开发言，由一个主响应 NPC 回应",
		"/+ask <npc> <message>         在群聊语境下点名某个 NPC 回应",
		"/+act <text>                  让玩家节点先解释自然语言，再走权威校验、执行和 NPC/群聊响应",
		"/+gift <npc> <item>           向 NPC 送礼，并先在游戏侧权威状态落地",
		"/+show_item <npc> <item>      向 NPC 展示你当前持有的物品",
		"/+trade [npc]                 发起交易/议价对话",
		"/+threaten [npc]              发起威胁式对话",
		"/+exit                        退出 play 模式",
		"兼容旧写法 /cmd               当前仍支持旧的 /help、/talk 这类输入",
		"直接输入文本                    向当前目标发送自然语言对话",
	}, "\n")
}

func fallback(primary, secondary string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return secondary
}

func fallbackSceneName(view *workerstate.StateView, sceneID string) string {
	scene, ok := view.Scene(sceneID)
	if !ok || scene == nil {
		return sceneID
	}
	return fallback(scene.Name, scene.ID)
}
