package main

import (
	"github.com/golang-migrate/migrate/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var steps int

var downMigrationCmd = &cobra.Command{
	Use:   "down",
	Short: "Run down migration for a specified number of steps",
	Run: func(cmd *cobra.Command, args []string) {
		if steps <= 0 {
			log.Fatal().Msg("Please specify a valid number of steps using the --steps flag")
		} else {
			DownWithSteps(steps)
		}
	},
}

func DownWithSteps(steps int) {
	log.Info().Int("steps", steps).Msg("Rolling back migration...")
	logLatestMigrationVersion()

	// Perform the rollback
	if err := m.Steps(-steps); err != nil {
		if err == migrate.ErrNoChange {
			log.Info().Msg("No migrations to roll back")
			return
		} else {
			log.Fatal().Caller().Err(err).Msg("Failed to roll back migration steps")
		}
	}

	logLatestMigrationVersion()
	log.Info().Int("steps", steps).Msg("Migration steps rolled back successfully")
}
