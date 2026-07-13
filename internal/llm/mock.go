package llm

import (
	"fmt"
	"strings"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
)

// mockProvider 是离线开发阶段使用的本地模拟 Provider。
type mockProvider struct{}

// NewMockProvider 创建一个返回固定结果的模拟 Provider。
func NewMockProvider() engine.LLMProvider {
	return &mockProvider{}
}

// Chat 根据 prompt 特征返回预设的 JSON 响应。
func (p *mockProvider) Chat(req *engine.LLMChatRequest) (*engine.LLMResult, error) {
	time.Sleep(200 * time.Millisecond)

	systemPrompt := ""
	var messages []engine.ChatMessage
	if req != nil {
		systemPrompt = req.SystemPrompt
		messages = req.Messages
	}

	var reply string
	if strings.Contains(systemPrompt, "自主行为决策器") {
		reply = `{"reply":"自主行为周期完成，当前没有必要行动。","action_calls":[],"memory_updates":[]}`
	} else if strings.Contains(systemPrompt, "world_tick") || strings.Contains(systemPrompt, "推演") {
		reply = `{"reply":"世界平稳运行，无明显重大事件。","world_change_plan":{"impact_level":"minor","summary":"世界状态稳定。","world_events":[{"event_type":"conflict","scope":"border","description":"边境小规模摩擦","confidence":0.7}],"proposed_actions":[{"api_name":"adjust_relation","args":{"delta":-5}}]},"memory_updates":[{"node_id":"world","content":"世界处于相对和平状态","level":"world"}]}`
	} else if strings.Contains(systemPrompt, "event_impact") || strings.Contains(systemPrompt, "重要事件") {
		reply = `{"reply":"该事件对世界格局产生了影响。","world_change_plan":{"impact_level":"major","summary":"事件改变了势力平衡。","world_events":[{"event_type":"diplomatic_shift","scope":"region","description":"势力关系变化","confidence":0.85}],"proposed_actions":[{"api_name":"update_relation","args":{"value":-20}}]},"memory_updates":[{"node_id":"world","content":"发生了重大外交事件","level":"shared"}]}`
	} else {
		lastMsg := ""
		if len(messages) > 0 {
			lastMsg = messages[len(messages)-1].Content
		}
		reply = fmt.Sprintf(`{"reply":"（作为角色回应：%s）","action_calls":[],"memory_updates":[{"node_id":"npc","content":"玩家说：%s","level":"short_term"}]}`, lastMsg, truncate(lastMsg, 50))
	}

	return &engine.LLMResult{Content: reply, Model: "mock", Tokens: 50}, nil
}

// ModelName 返回模拟 Provider 的模型名称。
func (p *mockProvider) ModelName() string { return "mock" }

// init 在编译期校验 mockProvider 实现了 LLMProvider 接口。
func init() { var _ engine.LLMProvider = (*mockProvider)(nil) }

// truncate 按最大字符数截断字符串，便于构造短文本回显。
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
