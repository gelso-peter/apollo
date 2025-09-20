package cmd

import (
	"apollo/db"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
)

var (
	sport     string
	yearStart int
	yearEnd   int
	numWeeks  int
)

func init() {
	CreateSeasonCmd.Flags().StringVarP(&sport, "sport", "", "", "Sport name (e.g. football)")
	CreateSeasonCmd.Flags().IntVarP(&yearStart, "year-start", "", 2025, "Start year")
	CreateSeasonCmd.Flags().IntVarP(&yearEnd, "year-end", "", 2026, "End year")
	CreateSeasonCmd.Flags().IntVarP(&numWeeks, "weeks", "", 23, "Number of weeks")

	// Mark sport and year-start as required
	CreateSeasonCmd.MarkFlagRequired("sport")
	CreateSeasonCmd.MarkFlagRequired("year-start")

	rootCmd.AddCommand(CreateSeasonCmd)
}

var CreateSeasonCmd = &cobra.Command{
	Use:   "create-season",
	Short: "Create a sport season (and weeks if football)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		db.ConnectDB()
		defer db.CloseDB()

		// Step 1: Create the sport season
		seasonIDStr, err := CreateSportSeason(ctx, db.DB, sport, yearStart, yearEnd)
		if err != nil {
			return err
		}
		fmt.Printf("Season created with ID: %s\n", seasonIDStr)

		// Step 2: If football, create weeks
		if sport == "football" {
			seasonID, _ := uuid.Parse(seasonIDStr)
			err = CreateSeasonWeeks(ctx, db.DB, seasonID, yearStart, numWeeks)
			if err != nil {
				return fmt.Errorf("failed to create football season weeks: %w", err)
			}
			fmt.Println("Football season weeks created successfully!")
		}

		return nil
	},
}

func CreateSportSeason(ctx context.Context, db *pgxpool.Pool, sport string, yearStart int, yearEnd int) (string, error) {
	id := uuid.New()

	_, err := db.Exec(ctx, `
		INSERT INTO sport_season (
			id, sport, year_start, year_end, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, now(), now()
		)
	`, id, sport, yearStart, yearEnd)

	if err != nil {
		return "", fmt.Errorf("failed to create season: %w", err)
	}

	return id.String(), nil
}

func CreateSeasonWeeks(ctx context.Context, db *pgxpool.Pool, seasonID uuid.UUID, yearStart int, numberOfWeeks int) error {
	startDate := time.Date(2025, time.September, 2, 0, 0, 0, 0, time.UTC)
	for startDate.Weekday() != time.Thursday {
		startDate = startDate.AddDate(0, 0, 1) // add 1 day
	}

	batch := &pgx.Batch{}

	for i := 0; i < numberOfWeeks; i++ {
		weekID := uuid.New()
		weekStart := startDate.AddDate(0, 0, i*7)
		weekEnd := weekStart.AddDate(0, 0, 6)

		batch.Queue(`
			INSERT INTO sport_season_week (
				id, sport_season_id, week_number, start_date, end_date,
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, now(), now())
		`, weekID, seasonID, i+1, weekStart, weekEnd)
	}

	br := db.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < numberOfWeeks; i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("failed to insert week %d: %w", i+1, err)
		}
	}

	return nil
}
