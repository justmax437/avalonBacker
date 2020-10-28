package main

import (
	"errors"
	"github.com/OrlovEvgeny/go-mcache"
	"github.com/google/uuid"
	"time"
)

var ErrSessionNotFound = errors.New("no session with specified UUID")

type memoryStorage struct {
	stor *mcache.CacheDriver
	ttl  time.Duration
}

func NewMemoryStorage(ttl time.Duration) GameSessionStorage {
	return &memoryStorage{mcache.New(), ttl}
}

func (i *memoryStorage) StoreSession(session *GameInstance) error {
	gameId, err := uuid.Parse(session.GameId.Value)
	if err != nil {
		return err
	}
	return i.stor.Set(gameId.String(), session, i.ttl)
}

func (i *memoryStorage) GetSession(id uuid.UUID) (*GameInstance, error) {
	data, found := i.stor.Get(id.String())
	if !found {
		return nil, ErrSessionNotFound
	}
	return data.(*GameInstance), nil
}

func (i *memoryStorage) CloseSession(id uuid.UUID) error {
	i.stor.Remove(id.String())
	return nil
}

func (i *memoryStorage) CheckExistance(id uuid.UUID) (bool, error) {
	_, found := i.stor.Get(id.String())
	return found, nil
}

func (i *memoryStorage) NumberOfGames() (uint, error) {
	return uint(i.stor.Len()), nil
}
