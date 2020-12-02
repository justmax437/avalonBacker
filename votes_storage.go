package main

import "github.com/google/uuid"

type VoteStorage struct {
	missionVotes map[uuid.UUID]int8
}

func NewVoteStorage() *VoteStorage {
	return &VoteStorage{make(map[uuid.UUID]int8, 0)}

}

func (v *VoteStorage) AddPositiveMissionVote(id uuid.UUID) {
	v.missionVotes[id]++
}

func (v *VoteStorage) AddNegativeMissionVote(id uuid.UUID) {
	v.missionVotes[id]--
}

func (v *VoteStorage) GetMissionVotesCountForGame(id uuid.UUID) int8 {
	return v.missionVotes[id]
}
