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

type GameSessionStorage interface {
	StoreSession(instance *GameInstance) error
	GetSession(id uuid.UUID) (*GameInstance, error)
	CloseSession(id uuid.UUID) error
	CheckExistance(id uuid.UUID) (bool, error)
	NumberOfGames() (uint, error)
}

var storage GameSessionStorage

func setSessionsStorage(st GameSessionStorage) {
	storage = st
}
