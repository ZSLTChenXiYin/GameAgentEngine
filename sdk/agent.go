package sdk

import (
	"fmt"
	"log"
)

// Agent 提供面向开发者的高层配置构建器。
type Agent struct {
	client     *Client
	worldID    string
	worldName  string
	nodes      []nodeDef
	components []compDef
	relations  []relDef
}

// nodeDef 表示待创建节点的本地定义。
type nodeDef struct {
	name     string
	nodeType string
	parentID string
}

// compDef 表示待创建组件的本地定义。
type compDef struct {
	nodeID string
	cType  string
	data   string
}

// relDef 表示待创建关系的本地定义。
type relDef struct {
	srcID, tgtID, rType string
	weight              int
	props               string
}

// NewAgent 创建一个绑定指定服务地址和 API Key 的 Agent 构建器。
func NewAgent(baseURL, apiKey string) *Agent {
	return &Agent{
		client: NewClient(baseURL, apiKey),
	}
}

// CreateWorld 初始化或复用一个同名世界。
func (a *Agent) CreateWorld(name string) (string, error) {
	if name == "" {
		name = "default"
	}
	a.worldName = name
	worlds, err := a.client.GetWorlds()
	if err != nil {
		return "", fmt.Errorf("get worlds: %w", err)
	}
	for _, w := range worlds {
		if w.Name == name {
			a.worldID = w.ID
			return w.ID, nil
		}
	}
	// 若不存在同名世界，则创建新的 world 节点。
	id, err := a.client.CreateNode("", name, "world", "")
	if err != nil {
		return "", fmt.Errorf("create world: %w", err)
	}
	a.worldID = id
	return id, nil
}

// WorldID 返回当前 Agent 绑定的世界 ID。
func (a *Agent) WorldID() string { return a.worldID }

// AddNode 注册一个待创建节点。
func (a *Agent) AddNode(name, nodeType, parentID string) string {
	id := name + "_" + nodeType
	a.nodes = append(a.nodes, nodeDef{name: name, nodeType: nodeType, parentID: parentID})
	return id
}

// AddComponent 注册一个待创建组件。
func (a *Agent) AddComponent(nodeID, compType, data string) {
	a.components = append(a.components, compDef{nodeID: nodeID, cType: compType, data: data})
}

// AddRelation 注册一条待创建关系。
func (a *Agent) AddRelation(srcID, tgtID, relType string, weight int) {
	a.relations = append(a.relations, relDef{srcID: srcID, tgtID: tgtID, rType: relType, weight: weight})
}

// Apply 将本地登记的节点、组件和关系写入远端服务。
func (a *Agent) Apply() error {
	if a.worldID == "" {
		return fmt.Errorf("no world created; call CreateWorld first")
	}
	nodeIDs := map[string]string{}

	// 先创建节点，记录名字到真实 ID 的映射。
	for _, n := range a.nodes {
		pid := n.parentID
		if pid != "" {
			if mapped, ok := nodeIDs[pid]; ok {
				pid = mapped
			}
		}
		id, err := a.client.CreateNode(a.worldID, n.name, n.nodeType, pid)
		if err != nil {
			return fmt.Errorf("create node %s: %w", n.name, err)
		}
		key := n.name + "_" + n.nodeType
		nodeIDs[key] = id
		log.Printf("[sdk] node %s (%s) = %s", n.name, n.nodeType, id[:8])
	}

	// 再补充组件数据。
	for _, c := range a.components {
		nid := c.nodeID
		if mapped, ok := nodeIDs[c.nodeID]; ok {
			nid = mapped
		} else if c.nodeID == "world" {
			nid = a.worldID
		}
		id, err := a.client.AddComponent(nid, c.cType, c.data)
		if err != nil {
			log.Printf("[sdk] component error: %v", err)
		} else {
			log.Printf("[sdk] component %s = %s", c.cType, id[:8])
		}
	}

	// 最后建立关系边。
	for _, r := range a.relations {
		src := r.srcID
		tgt := r.tgtID
		if m, ok := nodeIDs[r.srcID]; ok {
			src = m
		}
		if m, ok := nodeIDs[r.tgtID]; ok {
			tgt = m
		}
		id, err := a.client.CreateRelation(a.worldID, src, tgt, r.rType, r.weight)
		if err != nil {
			log.Printf("[sdk] relation error: %v", err)
		} else {
			log.Printf("[sdk] relation %s = %s", r.rType, id[:8])
		}
	}
	return nil
}

// InvokeNPC 向指定 NPC 发起对话推理请求。
func (a *Agent) InvokeNPC(nodeID string, messages []ChatMessage) (*InvokeResponse, error) {
	return a.client.Invoke(&InvokeRequest{
		WorldID:  a.worldID,
		TaskType: "npc_dialogue",
		NodeID:   nodeID,
		Messages: messages,
	})
}

// InvokeWorldTick 推进一次世界时间线。
func (a *Agent) InvokeWorldTick(gameTime string) (*InvokeResponse, error) {
	resp, err := a.client.AdvanceTick(a.worldID, "scheduled", gameTime)
	if err != nil {
		return nil, err
	}
	return resp.Invoke, nil
}

// RunAutonomousNode 手动触发某个节点的自主行为周期。
func (a *Agent) RunAutonomousNode(nodeID string) (*InvokeResponse, error) {
	return a.client.RunAutonomousNode(a.worldID, nodeID)
}

// InvokeEventImpact 发起一次世界事件影响评估请求。
func (a *Agent) InvokeEventImpact(event *WorldEvent) (*InvokeResponse, error) {
	return a.client.EventImpact(a.worldID, event)
}

// Client 返回底层 HTTP SDK 客户端。
func (a *Agent) Client() *Client { return a.client }
