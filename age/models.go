package age

type AnimeClient struct {
	Domain string
}

type AnimeSearchItem struct {
	Name  string
	Intro string
	Url   string
}

type AnimeEpisode struct {
	Name string
	Url  string
}

type AnimeSearchResponse struct {
	Results []AnimeSearchItem
	Page    int
}
