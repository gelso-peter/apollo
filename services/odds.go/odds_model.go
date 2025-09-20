package odds

import "time"

type OddsGameResponse struct {
	ID           string      `json:"id"`
	SportKey     string      `json:"sport_key"`
	HomeTeam     string      `json:"home_team"`
	AwayTeam     string      `json:"away_team"`
	CommenceTime time.Time   `json:"commence_time"`
	Bookmakers   []Bookmaker `json:"bookmakers"`
	Completed    bool        `json:"completed"`
	Scores       []TeamScore `json:"scores,omitempty"`
}

type Game struct {
	ID           string      `json:"id"`
	SportKey     string      `json:"sport_key"`
	CommenceTime time.Time   `json:"commence_time"`
	Spreads      Spread      `json:"spreads"`
	Completed    bool        `json:"completed"`
	Scores       []TeamScore `json:"scores,omitempty"`
}

type TeamScore struct {
	Name  string `json:"name"`
	Score int32  `json:"score"`
}

type Bookmaker struct {
	Key        string    `json:"key"`
	Title      string    `json:"title"`
	LastUpdate time.Time `json:"last_update"`
	Markets    []Market  `json:"markets"`
}

type Market struct {
	Key           string          `json:"key"` // e.g. "spreads"
	SpreadOutcome []SpreadOutcome `json:"outcomes"`
}

type SpreadOutcome struct {
	Name  string  `json:"name"`  // team name
	Point float64 `json:"point"` // spread (e.g. -6.5)
}

type Spread struct {
	HomeTeam SpreadOutcome `json:"home_team"`
	AwayTeam SpreadOutcome `json:"away_team"`
}
