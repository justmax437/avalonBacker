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

func (g *simpleGameService) CreateSession(ctx context.Context, config *api.GameConfig) (*api.GameSession, error) {
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

func (g *simpleGameService) TerminateSession(ctx context.Context, session *api.GameSession) (*types.Empty, error) {
	gameId, err := uuid.Parse(session.GameId.Value)
	if err != nil {
		return nil, errors.New("failed to parse session UUID: " + err.Error())
	}

	if exist, err := g.sessions.CheckExistance(gameId); err == nil && exist {
		return nil, g.sessions.CloseSession(gameId)
	} else {
		return nil, errors.New("failed to terminate session: " + err.Error())
	}
}

func (g *simpleGameService) GetSession(ctx context.Context, gameId *api.UUID) (*api.GameSession, error) {
	gi, err := g.sessions.GetSession(apiIDToUUID(gameId))
	if err == ErrSessionNotFound {
		return nil, ErrSessionNotFound
	} else if err != nil {
		return nil, errors.New("failed to read session data: " + err.Error())
	}
	return &gi.GameSession, nil
}

func (g *simpleGameService) GetEvilTeam(ctx context.Context, session *api.GameSession) (*api.EvilTeam, error) {
	panic("implement me")
}

func (g *simpleGameService) GetVirtuousTeam(ctx context.Context, session *api.GameSession) (*api.VirtuousTeam, error) {
	panic("implement me")
}

func (g *simpleGameService) PushGameState(ctx context.Context, session *api.GameSession) (*api.GameSession, error) {
	panic("implement me")
}

func (g *simpleGameService) GetPendingMission(ctx context.Context, session *api.GameSession) (*api.PendingMission, error) {
	panic("implement me")
}

func (g *simpleGameService) AssignMissionTeam(ctx context.Context, session *api.GameSession) (*types.Empty, error) {
	panic("implement me")
}

func (g *simpleGameService) GetMissionTeam(ctx context.Context, session *api.GameSession) (*api.MissionTeam, error) {
	panic("implement me")
}

func (g *simpleGameService) VoteForMissionTeam(ctx context.Context, context *api.VoteContext) (*types.Empty, error) {
	panic("implement me")
}

func (g *simpleGameService) VoteForMissionSuccess(ctx context.Context, context *api.VoteContext) (*types.Empty, error) {
	panic("implement me")
}

func apiIDToUUID(id *api.UUID) uuid.UUID {
	return uuid.MustParse(id.Value)
}
