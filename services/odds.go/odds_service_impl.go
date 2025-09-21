package odds

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type oddsServiceImpl struct {
	apiKey string
	client *http.Client
}

func mapOddsResponseToGames(allGames []OddsGameResponse) ([]Game, error) {
	var games []Game

	for _, gameResp := range allGames {
		// Find DraftKings bookmaker
		var dkBookmaker *Bookmaker
		for _, bm := range gameResp.Bookmakers {
			if bm.Key == "draftkings" {
				dkBookmaker = &bm
				break
			}
		}

		if dkBookmaker == nil {
			continue
		}

		// Find spreads market
		var spreadsMarket *Market
		for _, market := range dkBookmaker.Markets {
			if market.Key == "spreads" {
				spreadsMarket = &market
				break
			}
		}

		if spreadsMarket == nil {
			continue
		}

		var home SpreadOutcome
		var away SpreadOutcome

		// Match each outcome to home/away
		for _, outcome := range spreadsMarket.SpreadOutcome {
			if outcome.Name == gameResp.HomeTeam {
				home = SpreadOutcome{
					Name:  outcome.Name,
					Point: outcome.Point,
				}
			} else if outcome.Name == gameResp.AwayTeam {
				away = SpreadOutcome{
					Name:  outcome.Name,
					Point: outcome.Point,
				}
			}
		}

		// Only include game if both home and away spreads are found
		if home.Name == "" || away.Name == "" {
			continue
		}

		game := Game{
			ID:           gameResp.ID,
			SportKey:     gameResp.SportKey,
			CommenceTime: gameResp.CommenceTime,
			Completed:    gameResp.Completed,
			Scores:       gameResp.Scores,
			Spreads: Spread{
				HomeTeam: home,
				AwayTeam: away,
			},
		}

		games = append(games, game)
	}

	return games, nil
}

func (s *oddsServiceImpl) GetNFLGames(from, to string) ([]Game, error) {
	url := fmt.Sprintf(
		"https://api.the-odds-api.com/v4/sports/americanfootball_nfl/odds?regions=us&markets=spreads&oddsFormat=american&commenceTimeFrom=%s&commenceTimeTo=%s&apiKey=%s",
		from, to, s.apiKey,
	)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch odds: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 response: %s", resp.Status)
	}

	var allGames []OddsGameResponse
	if err := json.NewDecoder(resp.Body).Decode(&allGames); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Filter for Espnbet only
	for i := range allGames {
		for _, bookmaker := range allGames[i].Bookmakers {
			if bookmaker.Key == "espnbet" {
				allGames[i].Bookmakers = []Bookmaker{bookmaker}
				break
			}
		}
	}

	return mapOddsResponseToGames(allGames)
}

func (s *oddsServiceImpl) GetGameByID(gameID string) (*Game, error) {
	from := time.Now().UTC().AddDate(0, 0, -1).Format(time.RFC3339)
	to := time.Now().UTC().AddDate(0, 0, 7).Format(time.RFC3339)

	games, err := s.GetNFLGames(from, to)
	if err != nil {
		return nil, err
	}

	for _, g := range games {
		if g.ID == gameID {
			return &g, nil
		}
	}
	return nil, fmt.Errorf("game not found")
}

func (s *oddsServiceImpl) GetCompletedGames(from, to string) ([]Game, error) {
	url := fmt.Sprintf(
		"https://api.the-odds-api.com/v4/sports/americanfootball_nfl/scores?daysFrom=1&apiKey=%s",
		s.apiKey,
	)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch completed games: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 response: %s", resp.Status)
	}

	var allGames []OddsGameResponse
	if err := json.NewDecoder(resp.Body).Decode(&allGames); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var completedGames []Game
	for _, gameResp := range allGames {
		if !gameResp.Completed {
			continue
		}

		game := Game{
			ID:           gameResp.ID,
			SportKey:     gameResp.SportKey,
			CommenceTime: gameResp.CommenceTime,
			Completed:    gameResp.Completed,
			Scores:       gameResp.Scores,
		}

		completedGames = append(completedGames, game)
	}

	return completedGames, nil
}
