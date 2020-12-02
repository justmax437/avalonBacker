package main

import (
	"github.com/google/uuid"
	"github.com/justmax437/avalonBacker/api"
	"log"
)

type VoteStorage struct {
	missionVotes map[uuid.UUID]int8
	playersVoted map[uint64]bool
}

func NewVoteStorage() *VoteStorage {
	return &VoteStorage{
		make(map[uuid.UUID]int8, 0),
		make(map[uint64]bool),
	}
}

func (v *VoteStorage) AddPositiveMissionVote(id uuid.UUID, player *api.Player) {
	if _, alreadyVoted := v.playersVoted[player.Id]; alreadyVoted {
		log.Println("repeated vote attempt by", player)
		return
	}
	v.missionVotes[id]++
	v.playersVoted[player.Id] = true
}

func (v *VoteStorage) AddNegativeMissionVote(id uuid.UUID, player *api.Player) {
	if _, alreadyVoted := v.playersVoted[player.Id]; alreadyVoted {
		log.Println("repeated vote attempt by", player)
		return
	}
	v.missionVotes[id]--
	v.playersVoted[player.Id] = true
}

func (v *VoteStorage) GetMissionVotesCountForGame(id uuid.UUID) int8 {
	return v.missionVotes[id]
}
