package models

type Series struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	CurrentEpisode int    `json:"current_episode"`
	TotalEpisodes  int    `json:"total_episodes"`
	ImageURL       string `json:"image_url"`
}

type SeriesWithRating struct {
	Series
	Rating int `json:"rating"`
}
