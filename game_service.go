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
	votes    *VoteStorage
}

func NewGameService(s GameSessionStorage, votes *VoteStorage) *simpleGameService {
	if s == nil {
		log.Fatal("GameSessionStorage not provided")
	}
	gs := new(simpleGameService)
	gs.sessions = s
	gs.votes = votes
	return gs
}

func (g *simpleGameService) CreateSession(_ context.Context, config *api.GameConfig) (*api.GameSession, error) {
	if !checkNumberOfPlayersValid(len(config.GoodTeam.Members), len(config.EvilTeam.Members)) {
		return nil, errors.New("provided teams are not balanced by the game rules")
	}

	newGame := new(GameInstance)
	newGame.GameConfig = *config
	newGame.GameId = &api.UUID{Value: uuid.New().String()}
	newGame.State = api.GameSession_GAME_CREATED
	newGame.MissionTeam = api.MissionTeam{}
	newGame.Mission = api.PendingMission{
		MissionNumber:       0, // 0 means no mission
		TeamPickingAttempts: 0,
	}
	newGame.LastMissionResult = nil

	allPLayers := make([]*api.Player, 0, len(config.EvilTeam.Members)+len(config.GoodTeam.Members))
	allPLayers = append(
		append(allPLayers, config.EvilTeam.Members...),
		config.GoodTeam.Members...)
	shufflePlayers(allPLayers)
	newGame.Leader = allPLayers[0]
	newGame.CurrentLeaderIndex = 0
	newGame.AllPlayers = allPLayers

	g.votes.ResetVotes(apiIDToUUID(newGame.GetGameId()))
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
		return &types.Empty{}, g.sessions.CloseSession(gameId)
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
	return game.GetEvilTeam(), nil
}

func (g *simpleGameService) GetVirtuousTeam(_ context.Context, session *api.GameSession) (*api.VirtuousTeam, error) {
	game, err := g.sessions.GetSession(apiIDToUUID(session.GameId))
	if err != nil {
		return nil, errors.New("failed to read session data: " + err.Error())
	}
	return game.GetGoodTeam(), nil
}

func (g *simpleGameService) PushGameState(_ context.Context, session *api.GameSession) (*api.GameSession, error) {
	//Explicitly ignore everything except game id received from clients
	//Game state date from outside cannot be trusted
	game, err := g.sessions.GetSession(apiIDToUUID(session.GameId))
	if err != nil {
		return nil, errors.New("failed to read session data: " + err.Error())
	}

	switch game.GetState() {
	// At this stage we have teams that are balanced and ready to play
	// Everything is ready for first mission
	//TODO Move all state handling to separate entity (GameStateHandler interface)
	case api.GameSession_GAME_CREATED:
		game.State = api.GameSession_MISSION_TEAM_PICKING
		game.Mission = api.PendingMission{
			MissionNumber:       1,
			TeamPickingAttempts: 0,
		}

		if err := g.sessions.StoreSession(game); err != nil {
			return nil, errors.New("failed to store session data: " + err.Error())
		}

		return &game.GameSession, nil
	case api.GameSession_MISSION_TEAM_PICKING:
		//This state is where the client picks team for mission by calling AssignMissionTeam
		if len(game.MissionTeam.Members) == 0 {
			return nil, errors.New("mission team was not assigned, call AssignMissionTeam first")
		} else {
			game.State = api.GameSession_MISSION_TEAM_VOTING
			if err := g.sessions.StoreSession(game); err != nil {
				return nil, errors.New("failed to store session data: " + err.Error())
			}
		}
		return &game.GameSession, nil
	case api.GameSession_MISSION_TEAM_VOTING:
		if game.TotalPlayersCount() > int(g.votes.NumberOfPlayersVotedForTeam(apiIDToUUID(session.GetGameId()))) {
			log.Println(game.GameId, "not all players voted")
			return nil, errors.New("not all players voted")
		}

		if g.votes.GetTeamVotesCountForGame(apiIDToUUID(session.GameId)) <= 0 {
			// Mission Failed
			game.State = api.GameSession_MISSION_TEAM_PICKING
			game.Mission.TeamPickingAttempts++
			if game.Mission.TeamPickingAttempts == 6 {
				game.State = api.GameSession_EVIL_TEAM_WON
				game.EndgameReason = "Прошло 5 неудачных голосований за состав команды"
				return &game.GameSession, nil
			}

			if game.TotalPlayersCount() == game.CurrentLeaderIndex+1 {
				game.CurrentLeaderIndex = 0
			} else {
				game.CurrentLeaderIndex++
			}

			game.Leader = game.AllPlayers[game.CurrentLeaderIndex]

			if err := g.sessions.StoreSession(game); err != nil {
				return nil, errors.New("failed to store session data: " + err.Error())
			}

			return &game.GameSession, nil
		}

		//GameInstance.MissionTeam is already set in AssignMissionTeam call, so we just proceed to next state
		game.State = api.GameSession_MISSION_SUCCESS_VOTING

		if err := g.sessions.StoreSession(game); err != nil {
			return nil, errors.New("failed to store session data: " + err.Error())
		}

		return &game.GameSession, nil
	case api.GameSession_MISSION_SUCCESS_VOTING:
		// TODO check if everyone voted when votes storage are done

		if len(game.MissionTeam.Members) > g.votes.NumberOfPlayersVotedForMission(apiIDToUUID(session.GetGameId())) {
			log.Println(game.GameId, "not all players in mission team voted")
			return nil, errors.New("not all players in mission team voted")
		}

		failVotesRequired := 1
		if game.Mission.MissionNumber == 4 &&
			game.TotalPlayersCount() >= 7 {
			failVotesRequired = 2
		}

		failVotes := len(game.MissionTeam.Members) - int(g.votes.GetMissionVotesCountForGame(apiIDToUUID(game.GetGameId())))
		game.State = api.GameSession_MISSION_ENDED
		game.LastMissionResult = &api.MissionResult{
			Failed:        failVotes >= failVotesRequired,
			PositiveVotes: int32(game.TotalPlayersCount() - failVotes),
			NegativeVotes: int32(failVotes),
		}

		g.votes.ResetVotes(apiIDToUUID(game.GetGameId()))
		if err := g.sessions.StoreSession(game); err != nil {
			return nil, errors.New("failed to store session data: " + err.Error())
		}

		return &game.GameSession, nil
	case api.GameSession_MISSION_ENDED:
		return nil, nil
	default:
		return nil, errors.New("unknown game state encountered")
	}

	//Should never come to this
	return nil, errors.New("unknown error")
}

func (g *simpleGameService) GetPendingMission(_ context.Context, session *api.GameSession) (*api.PendingMission, error) {
	game, err := g.sessions.GetSession(apiIDToUUID(session.GameId))
	if err != nil {
		return nil, errors.New("failed to read session data: " + err.Error())
	}

	if game.Mission.GetMissionNumber() == 0 {
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

	if game.GetState() != api.GameSession_MISSION_TEAM_PICKING {
		return nil, errors.New("mission teams assignment only allowed in MISSION_TEAM_PICKING state")
	}

	game.MissionTeam = *assignReq.Team
	if err = g.sessions.StoreSession(game); err != nil {
		return nil, err
	}

	return &types.Empty{}, nil
}

func (g *simpleGameService) GetMissionTeam(_ context.Context, session *api.GameSession) (*api.MissionTeam, error) {
	game, err := g.sessions.GetSession(apiIDToUUID(session.GameId))
	if err != nil {
		return nil, errors.New("failed to read session data: " + err.Error())
	}
	if game.GetState() > api.GameSession_MISSION_TEAM_PICKING && game.State <= api.GameSession_MISSION_ENDED {
		return nil, errors.New("no active mission")
	}
	return &game.MissionTeam, nil
}

func (g *simpleGameService) VoteForMissionTeam(_ context.Context, ctx *api.VoteContext) (*types.Empty, error) {
	game, err := g.sessions.GetSession(apiIDToUUID(ctx.Session.GetGameId()))
	if err != nil {
		return nil, errors.New("failed to read session data: " + err.Error())
	}

	if len(game.MissionTeam.Members) == 0 {
		return nil, errors.New("mission team was not assigned, call AssignMissionTeam first")
	}

	if game.GetState() != api.GameSession_MISSION_TEAM_VOTING {
		return nil, errors.New("mission team votes are only allowed in MISSION_TEAM_VOTING state")
	}

	if ctx.GetVote() == api.VoteContext_NEGATIVE {
		g.votes.AddNegativeTeamVote(apiIDToUUID(ctx.Session.GetGameId()), ctx.Voter)
	}

	if ctx.GetVote() == api.VoteContext_POSITIVE {
		g.votes.AddPositiveTeamVote(apiIDToUUID(ctx.Session.GetGameId()), ctx.Voter)
	}

	return &types.Empty{}, nil
}

func (g *simpleGameService) VoteForMissionSuccess(_ context.Context, ctx *api.VoteContext) (*types.Empty, error) {
	game, err := g.sessions.GetSession(apiIDToUUID(ctx.Session.GetGameId()))
	if err != nil {
		return nil, errors.New("failed to read session data: " + err.Error())
	}

	if game.GetState() != api.GameSession_MISSION_SUCCESS_VOTING {
		return nil, errors.New("mission team votes are only allowed in MISSION_SUCCESS_VOTING state")
	}

	switch ctx.Vote {
	case api.VoteContext_NEGATIVE:
		g.votes.AddNegativeMissionVote(apiIDToUUID(game.GetGameId()), ctx.Voter)
	case api.VoteContext_POSITIVE:
		g.votes.AddPositiveMissionVote(apiIDToUUID(game.GetGameId()), ctx.Voter)
	}

	return &types.Empty{}, nil
}

func (g *simpleGameService) AssassinateAllegedMerlin(_ context.Context, ctx *api.AssassinationContext) (*api.AssassinationOutcome, error) {
	game, err := g.sessions.GetSession(apiIDToUUID(ctx.Session.GetGameId()))
	if err != nil {
		return nil, errors.New("failed to read session data: " + err.Error())
	}

	if game.GetState() != api.GameSession_POST_MISSIONS_ACTIONS {
		return nil, errors.New("assassinations are only available during POST_MISSIONS_ACTIONS state")
	}

	if game.GoodTeam.Merlin.Id == ctx.Target.Id {
		//Evils successfully found merlin
		game.State = api.GameSession_EVIL_TEAM_WON
		game.EndgameReason = "Мерлин был убит ассасином"

		if err := g.sessions.StoreSession(game); err != nil {
			return nil, errors.New("failed to store session data: " + err.Error())
		}

		return &api.AssassinationOutcome{
			Session:         &game.GameSession,
			MerlinWasKilled: true,
		}, nil
	} else {
		//Merlin stays alive
		game.State = api.GameSession_VIRTUOUS_TEAM_WON
		game.EndgameReason = "Все миссии завершены и Ассасину не удалось убить Мерлина"

		if err := g.sessions.StoreSession(game); err != nil {
			return nil, errors.New("failed to store session data: " + err.Error())
		}

		return &api.AssassinationOutcome{
			Session:         &game.GameSession,
			MerlinWasKilled: false,
		}, nil
	}
}
func apiIDToUUID(id *api.UUID) uuid.UUID {
	return uuid.MustParse(id.GetValue())
}
