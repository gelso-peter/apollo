-- Add odds_game_id column to track external odds API game IDs
ALTER TABLE game_pick ADD COLUMN odds_game_id TEXT;

-- Add unique constraint to prevent duplicate picks on same odds game
ALTER TABLE game_pick ADD CONSTRAINT unique_user_odds_game_week
    UNIQUE (user_id, odds_game_id, sport_season_week_id);