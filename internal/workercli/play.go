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

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/workerstate"
	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

type playSession struct {
	client          *sdk.Client
	view            *workerstate.StateView
	state           *workerstate.WorldState
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

func (a *app) runPlay() error {
	statePath := strings.TrimSpace(a.cfg.StateFile)
	if statePath == "" {
		return errors.New("play mode requires --state-file")
	}
	state, err := workerstate.LoadFile(statePath)
	if err != nil {
		return err
	}
	view := workerstate.NewStateView(state)
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
		state:          state,
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
	fmt.Println("输入 /help 查看命令；直接输入文本会发给当前对话目标。")
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
	if len(view.Actors()) == 1 {
		return view.Actors()[0].ID, nil
	}
	return "", errors.New("play mode requires --player-node-id when state file has multiple actors and none is marked kind=player")
}

func parsePlayCommand(line string) playCommand {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return playCommand{}
	}
	parts := strings.Fields(strings.TrimPrefix(trimmed, "/"))
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
		fmt.Println(s.renderSceneSummary())
		return false, nil
	case "who":
		fmt.Println(s.renderOccupants())
		return false, nil
	case "state":
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
	default:
		return false, fmt.Errorf("unknown command %q", cmd.Name)
	}
}

func (a *app) setPlayTarget(s *playSession, arg string) error {
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
	if strings.TrimSpace(s.currentTargetID) == "" {
		return errors.New("no active talk target; use /talk <npc>")
	}
	s.turnIndex++
	targetNodeID := strings.TrimSpace(s.currentTargetID)
	interaction := &sdk.InteractionContext{
		Mode:               "direct_dialogue",
		SpeakerNodeID:      s.playerNodeID,
		TargetNodeID:       targetNodeID,
		SceneNodeID:        s.currentSceneID,
		ParticipantNodeIDs: []string{s.playerNodeID, targetNodeID},
		AudienceScope:      "private",
		TurnIndex:          s.turnIndex,
		Event:              &sdk.InteractionEvent{Type: "speech"},
	}
	ctx := &sdk.InvokeContext{
		IncludeRelatedNodes: a.cfg.PlayIncludeRelated,
		Interaction:         interaction,
		DynamicInterfaces: []sdk.DynamicInterface{
			sdk.NewDynamicDataRequest(
				"play_authority",
				"game_client_request_data",
				sdk.WithDescription("Query authoritative game-side state such as HP, inventory, money, quest, scene, occupancy, and immediate room state during play mode."),
				sdk.WithQueryTypes("player_state", "player_inventory", "player_wallet", "player_location", "npc_location", "scene_state", "room_state", "task_state", "item_presence"),
				sdk.WithMaxQueries(8),
			),
		},
	}
	if mode := strings.TrimSpace(a.cfg.PlayPipelineMode); mode != "" {
		ctx.PipelineMode = mode
	}
	resp, err := s.client.Invoke(&sdk.InvokeRequest{
		WorldID:   s.worldID,
		NodeID:    targetNodeID,
		TaskType:  "npc_dialogue",
		SessionID: s.sessionID,
		Context:   ctx,
		Messages:  []sdk.ChatMessage{{Role: "user", Content: input}},
	})
	if err != nil {
		return err
	}
	if strings.TrimSpace(resp.Reply) == "" {
		fmt.Printf("[%s] （无文本回复）\n", targetNodeID)
		return nil
	}
	label := s.actorDisplayName(targetNodeID)
	fmt.Printf("[%s] %s\n", label, strings.TrimSpace(resp.Reply))
	if len(resp.ActionCalls) > 0 {
		fmt.Printf("[system] 引擎产生了 %d 个 action_calls，当前 play v1 仅展示，不在本地直接落地。\n", len(resp.ActionCalls))
	}
	return nil
}

func (s *playSession) resolveSceneActor(arg string) (*workerstate.ActorState, error) {
	trimmed := strings.TrimSpace(arg)
	if trimmed == "" {
		return nil, errors.New("/talk requires a target actor name or id")
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
			items = append(items, fmt.Sprintf("%s x%d", entry.ItemID, entry.Quantity))
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

func playHelpText() string {
	return strings.Join([]string{
		"/help                 查看帮助",
		"/look                 查看当前场景与同场角色",
		"/who                  列出当前场景角色",
		"/state                查看玩家权威状态摘要",
		"/talk <npc>           选择当前对话目标",
		"/target               查看当前对话目标",
		"/clear_target         清除当前对话目标",
		"/exit                 退出 play 模式",
		"直接输入文本            向当前目标发送自然语言对话",
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
