package main

import (
	"strconv"

	"github.com/go-errors/errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var forceMigrationCmd = &cobra.Command{
	Use:   "force [version]",
	Short: "Force migration to a specific version",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		version := args[0]
		forceVersion(version)
	},
}

func forceVersion(version string) {
	log.Info().Msg("Starting migration up process...")

	latestVer, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		log.Info().Msg("No migration has been applied")
	} else if err != nil {
		log.Fatal().Caller().Err(err).Send()
	} else {
		log.Info().Uint("latest_version", latestVer).Bool("dirty", dirty).Msg("Latest migration version")
	}

	v, err := strconv.Atoi(version)
	if err != nil {
		log.Fatal().Err(err).Msg("Invalid version number")
	}

	if err := m.Force(v); err != nil {
		log.Fatal().Err(err).Msg("Failed to force migration version")
	}

	log.Info().Int("forced_version", v).Msg("Successfully forced migration version")
}
