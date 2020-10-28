package main

import (
	"github.com/justmax437/avalonBacker/api"
	"math/rand"
)

func checkNumberOfPlayersValid(goodPlayers, evilPlayers int) bool {
	return map[int]map[int]bool{
		3: {2: true},
		4: {2: true, 3: true},
		5: {3: true},
		6: {3: true, 4: true},
	}[goodPlayers][evilPlayers]
}

func pickRandomLeader(players []*api.Player) *api.Player {
	return players[rand.Intn(len(players))]
}
