package main

import (
	"github.com/google/uuid"
	"github.com/justmax437/avalonBacker/api"
)

type GameInstance struct {
	api.GameSession
	api.GameConfig
	MissionTeam api.MissionTeam
	Mission     api.PendingMission
}

func (gi *GameInstance) TotalPlayersCount() int {
	return len(gi.EvilTeam.Members) + len(gi.GoodTeam.Members)
}

type GameSessionStorage interface {
	StoreSession(instance *GameInstance) error
	GetSession(id uuid.UUID) (*GameInstance, error)
	CloseSession(id uuid.UUID) error
	CheckExistence(id uuid.UUID) (bool, error)
	NumberOfGames() (uint, error)
}
