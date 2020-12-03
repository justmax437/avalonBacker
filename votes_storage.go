package main

import (
	"github.com/google/uuid"
	"github.com/justmax437/avalonBacker/api"
	"log"
)

type VoteStorage struct {
	missionVotes         map[uuid.UUID]int8
	teamVotes            map[uuid.UUID]int8
	playersVotedMissions map[uuid.UUID]map[uint64]bool
	playersVotedTeams    map[uuid.UUID]map[uint64]bool
}

func NewVoteStorage() *VoteStorage {
	return &VoteStorage{
		make(map[uuid.UUID]int8, 0),
		make(map[uuid.UUID]int8, 0),
		make(map[uuid.UUID]map[uint64]bool),
		make(map[uuid.UUID]map[uint64]bool),
	}
}

func (v *VoteStorage) AddPositiveMissionVote(gameId uuid.UUID, player *api.Player) {
	if _, alreadyVoted := v.playersVotedMissions[gameId][player.Id]; alreadyVoted {
		log.Println(gameId, "repeated vote attempt by", player)
		return
	}
	v.missionVotes[gameId]++
	v.playersVotedMissions[gameId][player.Id] = true
}

func (v *VoteStorage) AddNegativeMissionVote(gameId uuid.UUID, player *api.Player) {
	if _, alreadyVoted := v.playersVotedMissions[gameId][player.Id]; alreadyVoted {
		log.Println(gameId, "repeated vote attempt by", player)
		return
	}
	v.playersVotedMissions[gameId][player.Id] = true
}

func (v *VoteStorage) GetMissionVotesCountForGame(id uuid.UUID) int8 {
	return v.missionVotes[id]
}

func (v *VoteStorage) NumberOfPlayersVotedForMission(id uuid.UUID) int {
	return len(v.playersVotedMissions[id])
}

func (v *VoteStorage) AddPositiveTeamVote(gameId uuid.UUID, player *api.Player) {
	if _, alreadyVoted := v.playersVotedMissions[gameId][player.Id]; alreadyVoted {
		log.Println(gameId, "repeated vote attempt by", player)
		return
	}

	v.teamVotes[gameId]++
	v.playersVotedTeams[gameId][player.Id] = true
}

func (v *VoteStorage) AddNegativeTeamVote(gameId uuid.UUID, player *api.Player) {
	if _, alreadyVoted := v.playersVotedTeams[gameId][player.Id]; alreadyVoted {
		log.Println(gameId, "repeated vote attempt by", player)
		return
	}
	v.teamVotes[gameId]--
	v.playersVotedTeams[gameId][player.Id] = true
}

func (v *VoteStorage) GetTeamVotesCountForGame(id uuid.UUID) int8 {
	return v.teamVotes[id]
}

func (v *VoteStorage) NumberOfPlayersVotedForTeam(id uuid.UUID) int {
	return len(v.playersVotedTeams[id])
}

func (v *VoteStorage) ResetVotes(gameId uuid.UUID) {
	v.missionVotes[gameId] = 0
	v.missionVotes[gameId] = 0
	v.playersVotedMissions[gameId] = make(map[uint64]bool)
	v.playersVotedTeams[gameId] = make(map[uint64]bool)
}
