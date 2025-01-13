package main

import (
	"kelarin/internal/config"
	dbUtil "kelarin/internal/utils/dbutil"

	"github.com/go-errors/errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	cfg *config.Config
	m   *migrate.Migrate
)

var rootCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run migration",
}

func main() {
	cfg = config.NewApp()

	db, err := dbUtil.NewPostgres(&cfg.DataBase)
	if err != nil {
		log.Fatal().Caller().Err(err).Send()
	}

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		log.Fatal().Caller().Err(err).Send()
	}

	m, err = migrate.NewWithDatabaseInstance("file://database/migrations", "postgres", driver)
	if err != nil {
		log.Fatal().Caller().Err(err).Send()
	}

	if err != nil {
		log.Fatal().Caller().Err(err).Send()
	}

	downMigrationCmd.Flags().IntVarP(&steps, "steps", "s", 0, "Specify the number of steps to roll back")

	rootCmd.AddCommand(upMigrationCmd, forceMigrationCmd, downMigrationCmd, getLatestVersionCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Caller().Err(err).Send()
	}
}

func logLatestMigrationVersion() {
	latestVer, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		log.Info().Msg("No migration has been applied")
	} else if err != nil {
		log.Fatal().Caller().Err(err).Send()
	} else {
		log.Info().Uint("latest_version", latestVer).Bool("dirty", dirty).Msg("Latest migration version")
	}
}
