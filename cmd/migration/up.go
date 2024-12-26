package main

import (
	"github.com/go-errors/errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/spf13/cobra"

	"github.com/rs/zerolog/log"
)

var upMigrationCmd = &cobra.Command{
	Use:   "up",
	Short: "Run up migration",
	Run: func(cmd *cobra.Command, args []string) {
		Up()
	},
}

var getLatestVersionCmd = &cobra.Command{
	Use:   "latest",
	Short: "Get the latest migration version",
	Run: func(cmd *cobra.Command, args []string) {
		logLatestMigrationVersion()
	},
}

func Up() {
	log.Info().Msg("Starting migration up process...")

	logLatestMigrationVersion()

	err := m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		log.Info().Msg("No migration has been applied")
		return
	} else if err != nil {
		log.Fatal().Caller().Err(err).Send()
	}

	logLatestMigrationVersion()
}
