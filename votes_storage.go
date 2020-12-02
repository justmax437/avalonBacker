package main

import (
	"github.com/google/uuid"
	"github.com/justmax437/avalonBacker/api"
	"log"
)

type VoteStorage struct {
	missionVotes         map[uuid.UUID]int8
	teamVotes            map[uuid.UUID]int8
	playersVotedMissions map[uint64]bool
	playersVotedTeams    map[uint64]bool
}

func NewVoteStorage() *VoteStorage {
	return &VoteStorage{
		make(map[uuid.UUID]int8, 0),
		make(map[uuid.UUID]int8, 0),
		make(map[uint64]bool),
		make(map[uint64]bool),
	}
}

func (v *VoteStorage) AddPositiveMissionVote(id uuid.UUID, player *api.Player) {
	if _, alreadyVoted := v.playersVotedMissions[player.Id]; alreadyVoted {
		log.Println("repeated vote attempt by", player)
		return
	}
	v.missionVotes[id]++
	v.playersVotedMissions[player.Id] = true
}

func (v *VoteStorage) AddNegativeMissionVote(id uuid.UUID, player *api.Player) {
	if _, alreadyVoted := v.playersVotedMissions[player.Id]; alreadyVoted {
		log.Println("repeated vote attempt by", player)
		return
	}
	v.missionVotes[id]--
	v.playersVotedMissions[player.Id] = true
}

func (v *VoteStorage) GetMissionVotesCountForGame(id uuid.UUID) int8 {
	return v.missionVotes[id]
}

func (v *VoteStorage) AddPositiveTeamVote(id uuid.UUID, player *api.Player) {
	if _, alreadyVoted := v.playersVotedMissions[player.Id]; alreadyVoted {
		log.Println("repeated vote attempt by", player)
		return
	}
	v.teamVotes[id]++
	v.playersVotedTeams[player.Id] = true
}

func (v *VoteStorage) AddNegativeTeamVote(id uuid.UUID, player *api.Player) {
	if _, alreadyVoted := v.playersVotedTeams[player.Id]; alreadyVoted {
		log.Println("repeated vote attempt by", player)
		return
	}
	v.teamVotes[id]--
	v.playersVotedTeams[player.Id] = true
}

func (v *VoteStorage) GetTeamVotesCountForGame(id uuid.UUID) int8 {
	return v.teamVotes[id]
}

func (v *VoteStorage) ResetVotes() {
	v.missionVotes = make(map[uuid.UUID]int8, 0)
	v.missionVotes = make(map[uuid.UUID]int8, 0)
	v.playersVotedMissions = make(map[uint64]bool)
	v.playersVotedTeams = make(map[uint64]bool)
}
