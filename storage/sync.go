package storage

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

const (
	metaKindGame      = "game"
	metaKindSaveState = "savestate"
)

// SyncFromS3 rebuilds BoltDB from S3 metadata on startup.
// It also writes metadata to S3 for any existing BoltDB entry that lacks it (migration).
// No-op when S3 is not configured.
func (s *Storage) SyncFromS3() error {
	if s.config.S3 == nil {
		return nil
	}

	log.Info("syncing metadata from S3...")

	if err := s.syncGames(); err != nil {
		return err
	}
	if err := s.syncSaveStates(); err != nil {
		return err
	}

	log.Info("S3 sync complete")
	return nil
}

func (s *Storage) syncGames() error {
	// S3 → BoltDB: import games found on S3
	metas, err := s.fileStore.ListMeta(metaKindGame)
	if err != nil {
		return err
	}
	s3Ids := make(map[string]bool, len(metas))
	for _, data := range metas {
		var g Game
		if err := json.Unmarshal(data, &g); err != nil {
			log.Warnf("sync: invalid game metadata JSON, skipping: %s", err)
			continue
		}
		if _, err := s.fileStore.Head(g.Id, ""); err != nil {
			log.Warnf("sync: game %s (%s) has no ROM file in S3, skipping", g.Id, g.Name)
			continue
		}
		if err := s.store.Upsert(g.Id, g); err != nil {
			log.Warnf("sync: failed to upsert game %s: %s", g.Id, err)
			continue
		}
		s3Ids[g.Id] = true
	}

	// BoltDB → S3: write metadata for games not yet on S3 (migration)
	var all []Game
	if err := s.store.Find(&all, nil); err != nil {
		return err
	}
	for _, g := range all {
		if s3Ids[g.Id] {
			continue
		}
		data, err := json.Marshal(g)
		if err != nil {
			continue
		}
		if err := s.fileStore.SaveMeta(metaKindGame, g.Id, data); err != nil {
			log.Warnf("sync: failed to write game meta %s to S3: %s", g.Id, err)
		}
	}

	migrated := 0
	for _, g := range all {
		if !s3Ids[g.Id] {
			migrated++
		}
	}
	log.Infof("sync: %d game(s) from S3, %d migrated to S3", len(metas), migrated)
	return nil
}

func (s *Storage) syncSaveStates() error {
	// S3 → BoltDB
	metas, err := s.fileStore.ListMeta(metaKindSaveState)
	if err != nil {
		return err
	}
	s3Ids := make(map[string]bool, len(metas))
	for _, data := range metas {
		var ss SaveState
		if err := json.Unmarshal(data, &ss); err != nil {
			log.Warnf("sync: invalid savestate metadata JSON, skipping: %s", err)
			continue
		}
		if _, err := s.fileStore.Head(ss.Id, FileExtensionSaveState); err != nil {
			log.Warnf("sync: savestate %s has no .sav file in S3, skipping", ss.Id)
			continue
		}
		if err := s.store.Upsert(ss.Id, ss); err != nil {
			log.Warnf("sync: failed to upsert savestate %s: %s", ss.Id, err)
			continue
		}
		s3Ids[ss.Id] = true
	}

	// BoltDB → S3 (migration)
	var all []SaveState
	if err := s.store.Find(&all, nil); err != nil {
		return err
	}
	for _, ss := range all {
		if s3Ids[ss.Id] {
			continue
		}
		data, err := json.Marshal(ss)
		if err != nil {
			continue
		}
		if err := s.fileStore.SaveMeta(metaKindSaveState, ss.Id, data); err != nil {
			log.Warnf("sync: failed to write savestate meta %s to S3: %s", ss.Id, err)
		}
	}

	migrated := 0
	for _, ss := range all {
		if !s3Ids[ss.Id] {
			migrated++
		}
	}
	log.Infof("sync: %d savestate(s) from S3, %d migrated to S3", len(metas), migrated)
	return nil
}
