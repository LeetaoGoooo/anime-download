package download

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/melbahja/got"
	"gitlab.com/poldi1405/go-ansi"
	"gitlab.com/poldi1405/go-indicators/progress"
	"golang.org/x/term"
)

func color(content ...interface{}) string {
	return ansi.Blue(fmt.Sprint(content...))
}

func getWidth() int {

	if width, _, err := term.GetSize(0); err == nil && width > 0 {
		return width
	}

	return 80
}

func getProgessStyle(style *string, left *string, right *string) {
	if runtime.GOOS == "windows" {
		*style, *left, *right = "double-", "[", "]"
	} else {
		*style, *left, *right = "block", "▕", "▏"
	}
}

func Run(url string, dest string) {
	var (
		g     *got.Got           = got.New()
		p     *progress.Progress = new(progress.Progress)
		style string
		left  string
		right string
	)

	getProgessStyle(&style, &left, &right)
	p.SetStyle(style)

	// Progress func.
	g.ProgressFunc = func(d *got.Download) {

		// 55 is just an estimation of the text showed with the progress.
		// it's working fine with $COLUMNS >= 47
		p.Width = getWidth() - 55

		perc, err := progress.GetPercentage(float64(d.Size()), float64(d.TotalSize()))
		if err != nil {
			perc = 100
		}

		var bar string
		if getWidth() <= 46 {
			bar = ""
		} else {
			bar = right + color(p.GetBar(perc, 100)) + left
		}

		fmt.Printf(
			" %6.2f%% %s %s/%s @ %s/s%s\r",
			perc,
			bar,
			humanize.Bytes(d.Size()),
			humanize.Bytes(d.TotalSize()),
			humanize.Bytes(d.Speed()),
			ansi.ClearRight(),
		)
	}

	info, err := os.Stdin.Stat()

	if err != nil {
		log.Fatalf("err: %v \n", err)
	}

	// Piped stdin
	if info.Mode()&os.ModeNamedPipe > 0 || info.Size() > 0 {

		if err := multiDownload(g, bufio.NewScanner(os.Stdin), dest); err != nil {
			log.Fatalf("err: %v \n", err)
		}
	}

	if err = download(g, url, dest); err != nil {
		log.Fatalf("err: %v \n", err)
	}

	fmt.Print(ansi.ClearLine())
	fmt.Printf("✔ %s\n", url)

}

func download(g *got.Got, url string, dest string) (err error) {

	if url, err = getURL(url); err != nil {
		return err
	}

	return g.Do(&got.Download{
		URL:      url,
		Dest:     dest,
		Interval: 150,
	})
}

func multiDownload(g *got.Got, scanner *bufio.Scanner, dir string) error {

	for scanner.Scan() {

		url := strings.TrimSpace(scanner.Text())

		if url == "" {
			continue
		}

		if err := download(g, url, dir); err != nil {
			return err
		}

		fmt.Print(ansi.ClearLine())
		fmt.Printf("✔ %s\n", url)
	}

	return nil
}

func getURL(URL string) (string, error) {

	u, err := url.Parse(URL)

	if err != nil {
		return "", err
	}

	// Fallback to https by default.
	if u.Scheme == "" {
		u.Scheme = "https"
	}

	return u.String(), nil
}
