package models

type Winner struct {
	ID               int    `json:"id"`
	Player           string `json:"player"`
	Nationality      string `json:"nationality"`
	Club             string `json:"club"`
	Year             int    `json:"year"`
	Votes            int    `json:"votes"`
	Position         string `json:"position"`
	GoalsThatSeason  int    `json:"goals_that_season"`
}
