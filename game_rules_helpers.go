package main

import (
	"github.com/justmax437/avalonBacker/api"
	"math/rand"
	"time"
)

func checkNumberOfPlayersValid(goodPlayers, evilPlayers int) bool {
	return map[int]map[int]bool{
		3: {2: true},
		4: {2: true, 3: true},
		5: {3: true},
		6: {3: true, 4: true},
	}[goodPlayers][evilPlayers]
}

func shufflePlayers(players []*api.Player) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(players), func(i, j int) {
		players[i], players[j] = players[j], players[i]
	})
}
