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

-- Shared, immutable canonical seasons (e.g., NFL 2025)
CREATE TABLE sport_season (
    id UUID PRIMARY KEY,
    sport Sport NOT NULL,
    year_start INTEGER NOT NULL,
    year_end INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Shared 17-week structure for each sport_season
CREATE TABLE sport_season_week (
    id UUID PRIMARY KEY,
    sport_season_id UUID NOT NULL REFERENCES sport_season(id) ON DELETE CASCADE,
    week_number INTEGER NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- League-specific seasons (e.g., “League A NFL 2025 Season”)
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
    spread_selection INTEGER NOT NULL,  -- e.g. -3 for favored team
    spread_result INTEGER NOT NULL,     -- e.g. +7 if team beat the spread
    points_assigned INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
