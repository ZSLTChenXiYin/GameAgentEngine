package store

import "gorm.io/gorm"

// CreateWorldSnapshot persists a world snapshot metadata record.
func CreateWorldSnapshot(m *WorldSnapshotModel) error {
	if m.UUID == "" {
		m.UUID = NewUUID()
	}
	return Write(func(db *gorm.DB) error {
		return db.Create(m).Error
	})
}

// GetWorldSnapshotBySnapshotWorld returns snapshot metadata for a copied world.
func GetWorldSnapshotBySnapshotWorld(snapshotWorldUUID string) (*WorldSnapshotModel, error) {
	var m WorldSnapshotModel
	err := DB.Where("snapshot_world_uuid = ?", snapshotWorldUUID).First(&m).Error
	return &m, err
}

// GetWorldSnapshotBySnapshotWorldTx returns snapshot metadata for a copied world within a transaction.
func GetWorldSnapshotBySnapshotWorldTx(tx *gorm.DB, snapshotWorldUUID string) (*WorldSnapshotModel, error) {
	var m WorldSnapshotModel
	err := tx.Where("snapshot_world_uuid = ?", snapshotWorldUUID).First(&m).Error
	return &m, err
}

// ListWorldSnapshotsBySourceWorld returns snapshot metadata rows for a source world.
func ListWorldSnapshotsBySourceWorld(sourceWorldUUID, reason string) ([]WorldSnapshotModel, error) {
	var list []WorldSnapshotModel
	q := DB.Where("source_world_uuid = ?", sourceWorldUUID)
	if reason != "" {
		q = q.Where("reason = ?", reason)
	}
	err := q.Order("created_at DESC").Order("id DESC").Find(&list).Error
	return list, err
}

// ListWorldSnapshotsBySourceWorldTx returns snapshot metadata rows for a source world within a transaction.
func ListWorldSnapshotsBySourceWorldTx(tx *gorm.DB, sourceWorldUUID, reason string) ([]WorldSnapshotModel, error) {
	var list []WorldSnapshotModel
	q := tx.Where("source_world_uuid = ?", sourceWorldUUID)
	if reason != "" {
		q = q.Where("reason = ?", reason)
	}
	err := q.Order("created_at DESC").Order("id DESC").Find(&list).Error
	return list, err
}
