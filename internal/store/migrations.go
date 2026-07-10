package store

import (
	"fmt"
	"sync/atomic"

	"gorm.io/gorm"
)

var migrationsEnabled atomic.Bool

func init() {
	migrationsEnabled.Store(true)
}

func ConfigureMigrationsEnabled(enabled bool) {
	migrationsEnabled.Store(enabled)
}

type MigrationStep struct {
	Name string
	Run  func(db *gorm.DB) error
}

type MigrationRunner struct {
	steps []MigrationStep
}

func NewMigrationRunner() *MigrationRunner {
	return NewMigrationRunnerWithSteps(defaultMigrationSteps())
}

func NewMigrationRunnerWithSteps(steps []MigrationStep) *MigrationRunner {
	cloned := make([]MigrationStep, len(steps))
	copy(cloned, steps)
	return &MigrationRunner{steps: cloned}
}

func (r *MigrationRunner) Run(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("migration db is nil")
	}
	for _, step := range r.steps {
		if step.Run == nil {
			continue
		}
		if err := step.Run(db); err != nil {
			return fmt.Errorf("migration %s: %w", step.Name, err)
		}
	}
	return nil
}

func RunMigrations(db *gorm.DB) error {
	if !migrationsEnabled.Load() {
		return nil
	}
	return NewMigrationRunner().Run(db)
}

func defaultMigrationSteps() []MigrationStep {
	return []MigrationStep{
		{
			Name: "schema_auto_migrate",
			Run: func(db *gorm.DB) error {
				return db.AutoMigrate(
					&NodeModel{},
					&ComponentModel{},
					&RelationModel{},
					&MemoryModel{},
					&TimelineModel{},
					&InferenceLogModel{},
					&IdempotencyKeyModel{},
					&WorldSnapshotModel{},
					&WorldPolicyModel{},
					&WorldSettingsModel{},
					&PropagationChainModel{},
					&AsyncCallbackRecordModel{},
					&PausedExecutionModel{},
					&RuntimeTaskModel{},
				)
			},
		},
		{
			Name: "legacy_inference_logs_to_logs",
			Run:  migrateInferenceLogsToLogs,
		},
	}
}
