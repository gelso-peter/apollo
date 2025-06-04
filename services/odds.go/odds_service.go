package odds

import (
	"net/http"
	"sync"
)

type OddsService interface {
	GetNFLGames(from, to string) ([]Game, error)
	GetGameByID(gameID string) (*Game, error)
}

var (
	once     sync.Once
	instance OddsService
)

// InitOddsService initializes the singleton with your API key.
// Call this once in your main/server init.
func InitOddsService(apiKey string) {
	once.Do(func() {
		instance = &oddsServiceImpl{
			apiKey: apiKey,
			client: &http.Client{},
		}
	})
}

// GetOddsService returns the singleton instance. Panics if InitOddsService hasn't been called.
func GetOddsService() OddsService {
	if instance == nil {
		panic("odds service not initialized — call InitOddsService(apiKey) first")
	}
	return instance
}
