package main

import (
	"context"
	"errors"
	"github.com/gogo/protobuf/types"
	"github.com/google/uuid"
	"github.com/justmax437/avalonBacker/api"
	"log"
)

type simpleGameService struct {
	sessions GameSessionStorage
}

func NewGameService(s GameSessionStorage) *simpleGameService {
	if s == nil {
		log.Fatal("GameSessionStorage not provided")
	}
	gs := new(simpleGameService)
	gs.sessions = s
	return gs
}

func (g *simpleGameService) CreateSession(_ context.Context, config *api.GameConfig) (*api.GameSession, error) {
	if !checkNumberOfPlayersValid(len(config.GoodTeam.Members), len(config.EvilTeam.Members)) {
		return nil, errors.New("provided teams are not balanced by the game rules")
	}

	newGame := new(GameInstance)
	newGame.GameConfig = *config
	newGame.GameId.Value = uuid.New().String()
	newGame.State = api.GameSession_GAME_CREATED
	newGame.MissionTeam.Members = nil
	newGame.LastMissionResult = nil

	allPLayers := make([]*api.Player, len(config.EvilTeam.Members)+len(config.GoodTeam.Members))
	newGame.Leader = pickRandomLeader(
		append(
			append(allPLayers, config.EvilTeam.Members...),
			config.GoodTeam.Members...),
	)

	err := g.sessions.StoreSession(newGame)
	if err == nil {
		return &newGame.GameSession, nil
	} else {
		return nil, errors.New("failed to store game session: " + err.Error())
	}

}

func (g *simpleGameService) TerminateSession(_ context.Context, session *api.GameSession) (*types.Empty, error) {
	gameId, err := uuid.Parse(session.GameId.Value)
	if err != nil {
		return nil, errors.New("failed to parse session UUID: " + err.Error())
	}

	if exist, err := g.sessions.CheckExistence(gameId); err == nil && exist {
		return nil, g.sessions.CloseSession(gameId)
	} else {
		return nil, errors.New("failed to terminate session: " + err.Error())
	}
}

func (g *simpleGameService) GetSession(_ context.Context, gameId *api.UUID) (*api.GameSession, error) {
	gi, err := g.sessions.GetSession(apiIDToUUID(gameId))
	if err != nil {
		return nil, errors.New("failed to read session data: " + err.Error())
	}
	return &gi.GameSession, nil
}

func (g *simpleGameService) GetEvilTeam(_ context.Context, session *api.GameSession) (*api.EvilTeam, error) {
	game, err := g.sessions.GetSession(apiIDToUUID(session.GameId))
	if err != nil {
		return nil, errors.New("failed to read session data: " + err.Error())
	}
	return game.EvilTeam, nil
}

func (g *simpleGameService) GetVirtuousTeam(_ context.Context, session *api.GameSession) (*api.VirtuousTeam, error) {
	game, err := g.sessions.GetSession(apiIDToUUID(session.GameId))
	if err != nil {
		return nil, errors.New("failed to read session data: " + err.Error())
	}
	return game.GoodTeam, nil
}

func (g *simpleGameService) PushGameState(_ context.Context, session *api.GameSession) (*api.GameSession, error) {
	panic("implement me")
}

func (g *simpleGameService) GetPendingMission(_ context.Context, session *api.GameSession) (*api.PendingMission, error) {
	game, err := g.sessions.GetSession(apiIDToUUID(session.GameId))
	if err != nil {
		return nil, errors.New("failed to read session data: " + err.Error())
	}

	if game.Mission.MissionNumber == 0 {
		return nil, errors.New("no mission in progress")
	} else {
		return &game.Mission, nil
	}
}

func (g *simpleGameService) AssignMissionTeam(_ context.Context, assignReq *api.AssignTeamContext) (*types.Empty, error) {
	game, err := g.sessions.GetSession(apiIDToUUID(assignReq.Session.GameId))
	if err != nil {
		return nil, errors.New("failed to read session data: " + err.Error())
	}

	if game.State != api.GameSession_MISSION_TEAM_PICKING {
		return nil, errors.New("mission teams assignment only allowed in MISSION_TEAM_PICKING state")
	}

	game.MissionTeam = *assignReq.Team
	if err = g.sessions.StoreSession(game); err != nil {
		return nil, err
	}

	return nil, nil
}

func (g *simpleGameService) GetMissionTeam(_ context.Context, session *api.GameSession) (*api.MissionTeam, error) {
	game, err := g.sessions.GetSession(apiIDToUUID(session.GameId))
	if err != nil {
		return nil, errors.New("failed to read session data: " + err.Error())
	}
	if game.State > api.GameSession_MISSION_TEAM_PICKING && game.State <= api.GameSession_MISSION_ENDED {
		return nil, errors.New("no active mission")
	}
	return &game.MissionTeam, nil
}

func (g *simpleGameService) VoteForMissionTeam(_ context.Context, context *api.VoteContext) (*types.Empty, error) {
	panic("implement me")
}

func (g *simpleGameService) VoteForMissionSuccess(_ context.Context, context *api.VoteContext) (*types.Empty, error) {
	panic("implement me")
}

func apiIDToUUID(id *api.UUID) uuid.UUID {
	return uuid.MustParse(id.Value)
}
