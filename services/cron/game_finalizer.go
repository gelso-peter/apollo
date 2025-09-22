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

		// Calculate the detailed spread result
		outcome, marginAgainstSpread, covered, spreadResult, pointsAwarded, err := gf.calculateDetailedSpreadResult(pick, completedGame)
		if err != nil {
			log.Printf("Error calculating spread result for pick %s: %v", pick.ID, err)
			continue
		}

		// Update the pick in the database with detailed results
		err = gf.gamePickRepo.UpdateGamePickResultDetailed(gf.ctx, pick.ID, outcome, marginAgainstSpread, covered, spreadResult, pointsAwarded)
		if err != nil {
			log.Printf("Error updating pick %s: %v", pick.ID, err)
			continue
		}

		log.Printf("Finalized pick %s: outcome=%s, margin=%d, covered=%v, spread_result=%d, points=%d",
			pick.ID, outcome, marginAgainstSpread, covered, spreadResult, pointsAwarded)
		processedCount++
	}

	log.Printf("Game finalization completed. Processed %d picks", processedCount)
	return nil
}

// calculateDetailedSpreadResult determines the detailed outcome of a spread bet
func (gf *GameFinalizer) calculateDetailedSpreadResult(pick db.GamePick, completedGame odds.Game) (string, int32, *bool, int32, int32, error) {
	if len(completedGame.Scores) < 2 {
		return "", 0, nil, 0, 0, fmt.Errorf("incomplete score data for game %s", completedGame.ID)
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
		return "", 0, nil, 0, 0, fmt.Errorf("could not find scores for teams %s vs %s", pick.SelectedTeamName, pick.OpponentTeamName)
	}

	// Calculate the actual point differential (selected team score - opponent team score)
	actualDifferential := selectedTeamScore - opponentTeamScore

	// Convert spread_line to proper decimal (assuming it's stored as integer * 10, e.g., 25 = +2.5)
	spreadValue := float64(pick.SpreadLine) / 10.0

	// Calculate margin against spread: how much better/worse the team did vs the spread
	// Positive margin = covered the spread, Negative = didn't cover
	marginAgainstSpread := float64(actualDifferential) - spreadValue
	marginAgainstSpreadInt := int32(marginAgainstSpread * 10) // Store as int with 1 decimal precision

	log.Printf("Pick %s: %s (%d) vs %s (%d), spread_line: %d (%.1f), actual_diff: %d, margin: %.1f (%d)",
		pick.ID, pick.SelectedTeamName, selectedTeamScore, pick.OpponentTeamName, opponentTeamScore,
		pick.SpreadLine, spreadValue, actualDifferential, marginAgainstSpread, marginAgainstSpreadInt)

	// Determine outcome and coverage based on margin against spread
	var outcome string
	var covered *bool
	var spreadResult int32 = 0 // 0 for loss/push
	var pointsAwarded int32 = 0

	if marginAgainstSpread > 0 {
		// Win: covered the spread - award the full points the user assigned
		outcome = "win"
		trueBool := true
		covered = &trueBool
		spreadResult = 1
		pointsAwarded = pick.PointsAssigned // Award the points user originally assigned
	} else if marginAgainstSpread < 0 {
		// Loss: didn't cover the spread - award no points
		outcome = "loss"
		falseBool := false
		covered = &falseBool
		spreadResult = 0
		pointsAwarded = 0 // No points awarded for losses
	} else {
		// Push: exactly hit the spread (rare with .5 spreads) - award no points
		outcome = "push"
		covered = nil // NULL for pushes
		spreadResult = 0
		pointsAwarded = 0 // No points awarded for pushes
	}

	return outcome, marginAgainstSpreadInt, covered, spreadResult, pointsAwarded, nil
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
