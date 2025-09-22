CREATE TABLE app_user (
    id UUID PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE league (
    id UUID PRIMARY KEY,
    league_name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE user_league_association (
    user_id UUID NOT NULL,
    league_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (league_id, user_id),
    FOREIGN KEY (league_id) REFERENCES league(id),
    FOREIGN KEY (user_id) REFERENCES app_user(id)
);

CREATE TYPE Sport AS ENUM ('football');

CREATE TABLE sport_season (
    id UUID PRIMARY KEY,
    sport Sport NOT NULL,
    year_start INTEGER NOT NULL,
    year_end INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE sport_season_week (
    id UUID PRIMARY KEY,
    sport_season_id UUID NOT NULL REFERENCES sport_season(id) ON DELETE CASCADE,
    week_number INTEGER NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE league_season (
    id UUID PRIMARY KEY,
    league_id UUID NOT NULL REFERENCES league(id) ON DELETE CASCADE,
    sport_season_id UUID NOT NULL REFERENCES sport_season(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE game_pick (
    id UUID PRIMARY KEY,
    league_season_id UUID NOT NULL REFERENCES league_season(id),
    sport_season_week_id UUID NOT NULL REFERENCES sport_season_week(id),
    user_id UUID NOT NULL REFERENCES app_user(id),
    selected_team_name TEXT NOT NULL,
    opponent_team_name TEXT NOT NULL,
    spread_line INTEGER NOT NULL,  -- renamed from spread_selection
    spread_result INTEGER NOT NULL,
    points_assigned INTEGER NOT NULL,
    points_awarded INTEGER DEFAULT 0,
    odds_game_id TEXT NOT NULL,
    is_finalized BOOLEAN NOT NULL DEFAULT FALSE,
    outcome TEXT CHECK (outcome IN ('win','loss','push')),
    margin_against_spread INTEGER,
    covered BOOLEAN,
    finalized_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_user_odds_game_week UNIQUE (user_id, odds_game_id, sport_season_week_id)
);

CREATE INDEX IF NOT EXISTS idx_game_pick_finalized ON game_pick (is_finalized);
CREATE INDEX IF NOT EXISTS idx_game_pick_odds_game_id ON game_pick (odds_game_id);
