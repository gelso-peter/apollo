package cron

import (
	"apollo/db"
	"apollo/services/odds.go"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GameFinalizer struct {
	gamePickRepo *db.GamePickRepository
	oddsService  odds.OddsService
	ctx          context.Context
}

func NewGameFinalizer(dbPool *pgxpool.Pool, oddsService odds.OddsService) *GameFinalizer {
	return &GameFinalizer{
		gamePickRepo: db.NewGamePickRepository(dbPool),
		oddsService:  oddsService,
		ctx:          context.Background(),
	}
}

// FinalizeGames processes unfinalized picks and updates them if the games are completed
func (gf *GameFinalizer) FinalizeGames() error {
	log.Println("Starting game finalization process...")

	// Get all unfinalized picks
	unfinalizedPicks, err := gf.gamePickRepo.GetUnfinalizedPicks(gf.ctx)
	if err != nil {
		return fmt.Errorf("failed to get unfinalized picks: %w", err)
	}

	if len(unfinalizedPicks) == 0 {
		log.Println("No unfinalized picks found")
		return nil
	}

	log.Printf("Found %d unfinalized picks to process", len(unfinalizedPicks))

	// Get completed games from odds API
	from := time.Now().UTC().AddDate(0, 0, -7).Format(time.RFC3339) // Look back 7 days
	to := time.Now().UTC().Format(time.RFC3339)
	completedGames, err := gf.oddsService.GetCompletedGames(from, to)
	if err != nil {
		return fmt.Errorf("failed to get completed games: %w", err)
	}

	log.Printf("Found %d completed games from odds API", len(completedGames))

	// Create a map of completed games by ID for quick lookup
	completedGameMap := make(map[string]odds.Game)
	for _, game := range completedGames {
		completedGameMap[game.ID] = game
	}

	// Process each unfinalized pick
	processedCount := 0
	for _, pick := range unfinalizedPicks {
		if pick.OddsGameID == nil {
			continue // Skip picks without odds_game_id
		}

		completedGame, exists := completedGameMap[*pick.OddsGameID]
		if !exists {
			continue // Game not completed yet
		}

		// Calculate the spread result
		spreadResult, pointsAwarded, err := gf.calculateSpreadResult(pick, completedGame)
		if err != nil {
			log.Printf("Error calculating spread result for pick %s: %v", pick.ID, err)
			continue
		}

		// Update the pick in the database
		err = gf.gamePickRepo.UpdateGamePickResult(gf.ctx, pick.ID, spreadResult, pointsAwarded)
		if err != nil {
			log.Printf("Error updating pick %s: %v", pick.ID, err)
			continue
		}

		log.Printf("Finalized pick %s: spread_result=%d, points=%d", pick.ID, spreadResult, pointsAwarded)
		processedCount++
	}

	log.Printf("Game finalization completed. Processed %d picks", processedCount)
	return nil
}

// calculateSpreadResult determines if the user's pick beat the spread and calculates points
func (gf *GameFinalizer) calculateSpreadResult(pick db.GamePick, completedGame odds.Game) (int32, int32, error) {
	if len(completedGame.Scores) < 2 {
		return 0, 0, fmt.Errorf("incomplete score data for game %s", completedGame.ID)
	}

	// Find the scores for the selected team and opponent
	var selectedTeamScore, opponentTeamScore int32
	var selectedTeamFound, opponentTeamFound bool

	for _, score := range completedGame.Scores {
		if score.Name == pick.SelectedTeamName {
			selectedTeamScore = score.Score
			selectedTeamFound = true
		} else if score.Name == pick.OpponentTeamName {
			opponentTeamScore = score.Score
			opponentTeamFound = true
		}
	}

	if !selectedTeamFound || !opponentTeamFound {
		return 0, 0, fmt.Errorf("could not find scores for teams %s vs %s", pick.SelectedTeamName, pick.OpponentTeamName)
	}

	// Calculate the actual point differential
	actualDifferential := selectedTeamScore - opponentTeamScore

	// Check if the pick beat the spread
	// If user picked -3 (favored by 3), they need to win by more than 3
	// If user picked +3 (underdog by 3), they need to lose by less than 3 or win
	beatSpread := false

	if pick.SpreadSelection < 0 {
		// User picked the favorite
		beatSpread = actualDifferential > -pick.SpreadSelection
	} else {
		// User picked the underdog
		beatSpread = actualDifferential > -pick.SpreadSelection
	}

	// Calculate points and spread result
	var spreadResult int32 = 0 // 0 for loss
	var pointsAwarded int32 = 0

	if beatSpread {
		spreadResult = 1                    // 1 for win
		pointsAwarded = pick.PointsAssigned // Award the points they assigned
	}

	return spreadResult, pointsAwarded, nil
}

// RunPeriodically runs the game finalization process on specific days at 6:00 AM
// Runs on: Friday, Sunday, Monday, and Tuesday mornings at 6:00 AM
func (gf *GameFinalizer) RunPeriodically(stop <-chan struct{}) {
	// Define target days (Friday=5, Sunday=0, Monday=1, Tuesday=2)
	targetDays := map[time.Weekday]bool{
		time.Friday:  true,
		time.Sunday:  true,
		time.Monday:  true,
		time.Tuesday: true,
	}

	// Check if today is a target day and if it's past 6 AM, run immediately
	now := time.Now()
	if targetDays[now.Weekday()] && now.Hour() >= 6 {
		log.Println("Running initial game finalization (today is a target day and it's past 6 AM)")
		if err := gf.FinalizeGames(); err != nil {
			log.Printf("Error during initial game finalization: %v", err)
		}
	}

	// Create a ticker that checks every hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	log.Println("Game finalization cron job started - runs on Fri/Sun/Mon/Tue at 6:00 AM")

	for {
		select {
		case <-ticker.C:
			now := time.Now()

			// Check if today is a target day and it's 6 AM
			if targetDays[now.Weekday()] && now.Hour() == 6 && now.Minute() < 60 {
				log.Printf("Running scheduled game finalization - %s at %02d:%02d",
					now.Weekday().String(), now.Hour(), now.Minute())

				if err := gf.FinalizeGames(); err != nil {
					log.Printf("Error during scheduled game finalization: %v", err)
				}
			}
		case <-stop:
			log.Println("Stopping game finalization service")
			return
		}
	}
}
