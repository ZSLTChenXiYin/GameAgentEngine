package store

import (
	"fmt"
	"sync"
	"testing"
)

func initWorldSettingsStoreTestDB(t *testing.T) string {
	t.Helper()
	if err := Init("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())); err != nil {
		t.Fatalf("init db: %v", err)
	}
	world := &NodeModel{UUID: NewUUID(), Name: "Store Settings World", NodeType: "world"}
	if err := CreateNode(world); err != nil {
		t.Fatalf("create world: %v", err)
	}
	if err := DB.Model(world).Update("world_id", world.ID).Error; err != nil {
		t.Fatalf("set world_id: %v", err)
	}
	return world.UUID
}

func TestGetOrCreateWorldSettingsConcurrent(t *testing.T) {
	worldID := initWorldSettingsStoreTestDB(t)
	const workers = 8

	var wg sync.WaitGroup
	errCh := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := GetOrCreateWorldSettings(worldID); err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent get or create failed: %v", err)
		}
	}

	var count int64
	if err := DB.Model(&WorldSettingsModel{}).Where("world_uuid = ?", worldID).Count(&count).Error; err != nil {
		t.Fatalf("count world settings: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 world settings row, got %d", count)
	}
}

func TestUpsertWorldSettingsWithMaskAfterExistingRecord(t *testing.T) {
	worldID := initWorldSettingsStoreTestDB(t)
	if _, err := GetOrCreateWorldSettings(worldID); err != nil {
		t.Fatalf("seed world settings: %v", err)
	}

	settings, err := UpsertWorldSettingsWithMask(worldID, &WorldSettingsModel{PipelineMode: "polling"}, &WorldSettingsUpdateMask{PipelineMode: true})
	if err != nil {
		t.Fatalf("upsert world settings: %v", err)
	}
	if settings.PipelineMode != "polling" {
		t.Fatalf("expected pipeline mode polling, got %q", settings.PipelineMode)
	}

	var count int64
	if err := DB.Model(&WorldSettingsModel{}).Where("world_uuid = ?", worldID).Count(&count).Error; err != nil {
		t.Fatalf("count world settings: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 world settings row after upsert, got %d", count)
	}
}
