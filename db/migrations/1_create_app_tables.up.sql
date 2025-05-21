CREATE TABLE app_user (
    id UUID PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    meta_seq INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE league (
    id UUID PRIMARY KEY,
    league_name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    meta_seq INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE league_user_association (
    league_id UUID NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    meta_seq INTEGER NOT NULL DEFAULT 0,

    PRIMARY KEY (league_id, user_id),
    FOREIGN KEY (league_id) REFERENCES league(id),
    FOREIGN KEY (user_id) REFERENCES app_user(id)
);

CREATE TYPE Sport AS ENUM ('football');

CREATE TABLE season (
    id UUID PRIMARY KEY,
    league_id UUID NOT NULL REFERENCES league(id),
    year_start INTEGER NOT NULL,
    year_end INTEGER NOT NULL,
    sport Sport NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    meta_seq INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE game_pick (
    id UUID PRIMARY KEY,
    season_id UUID NOT NULL REFERENCES season(id),
    selected_team_name TEXT NOT NULL,
    opponent_team_name TEXT NOT NULL,
    spread_selection INTEGER NOT NULL,
    spread_result INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    meta_seq INTEGER NOT NULL DEFAULT 0
);
