-- Remove the unique constraint
ALTER TABLE game_pick DROP CONSTRAINT IF EXISTS unique_user_odds_game_week;

-- Remove the odds_game_id column
ALTER TABLE game_pick DROP COLUMN IF EXISTS odds_game_id;