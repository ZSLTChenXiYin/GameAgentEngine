package service

import (
	"fmt"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

func initStateComponentTestDB(t *testing.T) {
	t.Helper()
	if err := store.Init("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())); err != nil {
		t.Fatalf("init db: %v", err)
	}
}

func createStateComponentWorld(t *testing.T) string {
	t.Helper()
	world, err := CreateNode("", "World", "world", nil)
	if err != nil {
		t.Fatalf("create world: %v", err)
	}
	if err := store.DB.Model(world).Update("world_id", world.ID).Error; err != nil {
		t.Fatalf("set world id: %v", err)
	}
	return world.UUID
}

func TestUpsertStateComponentCreatesAndUpdatesStructuredComponents(t *testing.T) {
	initStateComponentTestDB(t)
	worldID := createStateComponentWorld(t)

	created, err := UpsertStateComponent(worldID, engine.CompWorldState, engine.WorldStateComponent{
		Summary:  "initial",
		KeyFacts: []string{"fact-a"},
	})
	if err != nil {
		t.Fatalf("create state component: %v", err)
	}
	if created == nil || created.ComponentType != string(engine.CompWorldState) {
		t.Fatalf("unexpected created component: %#v", created)
	}

	updated, err := UpsertStateComponent(worldID, engine.CompWorldState, engine.WorldStateComponent{
		Summary:  "updated",
		KeyFacts: []string{"fact-b"},
	})
	if err != nil {
		t.Fatalf("update state component: %v", err)
	}
	if updated.UUID != created.UUID {
		t.Fatalf("expected upsert to reuse component, got %s != %s", updated.UUID, created.UUID)
	}

	loaded, err := GetStateComponent(worldID, engine.CompWorldState)
	if err != nil {
		t.Fatalf("get state component: %v", err)
	}
	if loaded == nil || loaded.Data == "" {
		t.Fatalf("expected stored data, got %#v", loaded)
	}
	if loaded.ComponentType != string(engine.CompWorldState) {
		t.Fatalf("unexpected component type: %s", loaded.ComponentType)
	}
	if len(engine.ValidComponentTypes()) < 14 {
		t.Fatalf("expected new component types registered, got %#v", engine.ValidComponentTypes())
	}
}
