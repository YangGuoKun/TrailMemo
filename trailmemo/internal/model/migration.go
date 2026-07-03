package model

import (
	"errors"

	"github.com/trailmemo/internal/agent/memory"
	"github.com/trailmemo/internal/config"
	platformlogger "github.com/trailmemo/internal/platform/logger"
	"go.uber.org/zap"
)

func AutoMigrate() error {
	db := config.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

	err := db.AutoMigrate(
		&User{},
		&Route{},
		&Checkpoint{},
		&Checkin{},
		&Post{},
		&Comment{},
		&Like{},
		&Favorite{},
		&Share{},
		&memory.AgentRun{},
		&memory.AgentStep{},
		&memory.AgentArtifact{},
		&memory.AgentUserPreference{},
		&memory.AgentSession{},
	)

	if err != nil {
		return err
	}

	platformlogger.L().Info("migration_completed",
		zap.String("event", "migration_completed"),
	)
	return nil
}
