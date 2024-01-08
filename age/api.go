package age

import (
	"age-downloader/download"
	"context"
	"fmt"
	"log"
	"path"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/levigross/grequests"
)

// search anime by name
// TODO paginate
func (client *AnimeClient) SearchAnime(anime string, page int) AnimeSearchResponse {
	resp, err := grequests.Get(fmt.Sprintf("%s/search?query=%s&page=%d", client.Domain, anime, page), nil)

	if err != nil {
		log.Fatalln("Unable to make request: ", err)
	}

	doc, err := goquery.NewDocumentFromReader(resp.RawResponse.Body)

	if err != nil {
		log.Fatalln("Unable to parse html: ", err)
	}

	// get summary of serach keyword
	searchKeywordEle := doc.Find(".search_keywords").First()

	fmt.Println(searchKeywordEle.Text())

	searchReponse := AnimeSearchResponse{
		Page:    page,
		Results: []AnimeSearchItem{},
	}

	// get search results of current page
	doc.Find(".card.cata_video_item.py-4").Each(func(_ int, node *goquery.Selection) {
		titleNode := node.Find(".card-title a").First()
		introNode := node.Find(".video_detail_info.desc").First()
		searchItem := AnimeSearchItem{
			Name:  titleNode.Text(),
			Url:   titleNode.AttrOr("href", ""),
			Intro: introNode.Text(),
		}
		searchReponse.Results = append(searchReponse.Results, searchItem)
	})

	return searchReponse
}

// fetch episodes by anime search result
// TODO tab options
func (client *AnimeClient) FetchAnimeEpisodes(item AnimeSearchItem) []AnimeEpisode {
	resp, err := grequests.Get(item.Url, nil)

	if err != nil {
		log.Fatalln("Unable to make request: ", err)
	}

	doc, err := goquery.NewDocumentFromReader(resp.RawResponse.Body)

	if err != nil {
		log.Fatalln("Unable to parse html: ", err)
	}

	animeEpisodes := []AnimeEpisode{
		{
			Name: "All",
			Url:  "",
		},
	}

	doc.Find(".video_detail_episode li").Each(func(_ int, node *goquery.Selection) {
		titleNode := node.Find("a").First()
		animeEpisode := AnimeEpisode{
			Name: titleNode.Text(),
			Url:  titleNode.AttrOr("href", ""),
		}
		animeEpisodes = append(animeEpisodes, animeEpisode)
	})

	return animeEpisodes
}

// download anime
func (client *AnimeClient) DownloadAnime(item AnimeEpisode, dir string) {
	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true),
	}
	options = append(chromedp.DefaultExecAllocatorOptions[:], options...)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// run task list
	var iframeUrl string
	var res string

	err := chromedp.Run(ctx,
		chromedp.Navigate(item.Url),
		chromedp.WaitReady(`iframe`, chromedp.ByQuery),
		chromedp.EvaluateAsDevTools(`document.getElementsByTagName("iframe")[0].src;`, &iframeUrl),
	)

	if err != nil {
		log.Fatal("Fetch iframe link failed: ", err)
	}

	err = chromedp.Run(ctx,
		chromedp.Navigate(iframeUrl),
		chromedp.WaitReady(`video`, chromedp.ByQuery),
		chromedp.EvaluateAsDevTools(`document.getElementsByTagName("video")[0].src;`, &res),
	)

	if err != nil {
		log.Fatal("Fetch video link failed: ", err)
	}

	// fmt.Printf("fetch video link: %v\n", res)

	dest := path.Join(dir, fmt.Sprintf("%s.mp4", item.Name))
	download.Run(res, dest)
}
