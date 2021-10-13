// Package schema defines database and JSON schema as structs, as well as functions for creating and using these structs
package schema

type Leaderboard struct {
	Thing
	Users []LeaderboardEntry `json:"users" binding:"required"`
}

// // The following is an example of a LeaderboardEntry reply
// "name": "Greenitthe",
// "totalGolems": 47,
// "coins": 123456,
// "callsLastHour": 212,
// "achievementCount": 7,
// "creationTimestamp": 0000000000

type LeaderboardEntry struct {
	Rank int `json:"rank" binding:"required"`
	PublicUserInfo
}

var Leaderboards = map[string]Leaderboard {
	"coin-leaders": {Thing:Thing{HasSymbol:HasSymbol{Symbol:"coin-leaders"}, Name:"Coin Leaders", Description:"Top 10 Users by Coins"}, Users:make([]LeaderboardEntry, 0)},
	"golem-leaders": {Thing:Thing{HasSymbol:HasSymbol{Symbol:"golem-leaders"}, Name:"Golem Leaders", Description:"Top 10 Users by Golem Count"}, Users:make([]LeaderboardEntry, 0)},
	"achievement-leaders": {Thing:Thing{HasSymbol:HasSymbol{Symbol:"achievement-leaders"}, Name:"Achievement Leaders", Description:"Top 10 Users by Achievements Completed"}, Users:make([]LeaderboardEntry, 0)},
}

type LeaderboardDescriptionResponse struct {
	Symbol string `json:"symbol" binding:"required"`
	Name string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

func GetLeaderboardDescriptionResponses(boards []Leaderboard) ([]LeaderboardDescriptionResponse) {
	response := make([]LeaderboardDescriptionResponse, 0)
	for _, board := range boards {
		temp := LeaderboardDescriptionResponse{
			Symbol: board.Symbol,
			Name: board.Name,
			Description: board.Description,
		}
		response = append(response, temp)
	}
	return response
}