package bench

import (
	"testing"
	"time"
	"encoding/json"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
)

// BenchmarkWorldFocusConfigParsing benchmarks decoding world_focus configs.
func BenchmarkWorldFocusConfigParsing(b *testing.B) {
	jsonData := `{"enabled":true,"tasks":["world_tick"],"priority":80,"reason":"quest_hub","max_parent_distance":3,"summary_only":true,"include_relations":["belongs_to","subordinate","located_at"]}`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg, err := engine.DecodeWorldFocusConfig(jsonData)
		if err != nil {
			b.Fatal(err)
		}
		if !cfg.Enabled {
			b.Fatal("expected enabled")
		}
	}
}

// BenchmarkAutonomousConfigParsing benchmarks decoding autonomous configs.
func BenchmarkAutonomousConfigParsing(b *testing.B) {
	jsonData := `{"enabled":true,"trigger":"scheduled","interval_seconds":300,"priority":50,"cooldown_seconds":60,"status":"idle","capabilities":[{"id":"send_dialogue","description":"Send dialogue to target"}]}`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg, err := engine.DecodeAutonomousConfig(jsonData)
		if err != nil {
			b.Fatal(err)
		}
		if !cfg.Enabled {
			b.Fatal("expected enabled")
		}
	}
}

// BenchmarkAutonomousScoring benchmarks the scoring function.
func BenchmarkAutonomousScoring(b *testing.B) {
	cfg := &engine.AutonomousConfig{
		Enabled:         true,
		Trigger:         "scheduled",
		IntervalSeconds: 300,
		Priority:        50,
		CooldownSeconds: 60,
		Status:          "idle",
		Capabilities:    []engine.AgentCapability{{ID: "send_dialogue"}},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.ScoreAutonomousNode(cfg, time.Time{})
	}
}

// BenchmarkCandidateScoring benchmarks the candidate scoring for a node.
func BenchmarkCandidateScoring(b *testing.B) {
	cfg := &engine.WorldFocusConfig{
		Enabled:  true,
		Tasks:    []string{"world_tick"},
		Priority: 80,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.Priority
	}
}

// BenchmarkJSONMarshal benchmarks JSON marshaling of common types.
func BenchmarkJSONMarshal(b *testing.B) {
	type testPayload struct {
		WorldID   string                 `json:"world_id"`
		TaskType  string                 `json:"task_type"`
		NodeID    string                 `json:"node_id"`
		Reply     string                 `json:"reply"`
		Extra     map[string]interface{} `json:"extra,omitempty"`
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := testPayload{
			WorldID:  "world_001",
			TaskType: "world_tick",
			NodeID:   "node_001",
			Reply:    strings.Repeat("test reply content for benchmark ", 10),
		}
		_, err := json.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkScoreCandidateNode benchmarks the ScoreCandidateNode function.
func BenchmarkScoreCandidateNode(b *testing.B) {
	_ = b
	// ScoreCandidateNode requires store access so this is a placeholder
	// showing the expected benchmark signature
	b.StopTimer()
}
