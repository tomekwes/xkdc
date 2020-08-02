package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type comicNumber int

const (
	BaseURL        string        = "https://xkcd.com"
	DefaultTimeout time.Duration = 30 * time.Second
	LatestComic    comicNumber   = 0
)

type ComicResponse struct {
	Num        int    `json:"num"`
	Month      string `json:"month"`
	Day        string `json:"day"`
	Year       string `json:"year"`
	Title      string `json:"title"`
	Alt        string `json:"alt"`
	Img        string `json:"img"`
	Safe_title string `json:"safe_title"`
	Link       string `json:"link"`
	News       string `json:"news"`
	Transcript string `json:"transcript"`
}

type Comic struct {
	Title       string `json:"title"`
	Number      int    `json:"number"`
	Date        string `json:"date"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

type XKCDClient struct {
	client  *http.Client
	baseURL string
}

func (cr ComicResponse) formatedDate() string {
	return fmt.Sprintf("%s-%s-%s", cr.Year, cr.Month, cr.Day)
}

func (cr ComicResponse) Comic() Comic {

	return Comic{
		Title:       cr.Title,
		Number:      cr.Num,
		Date:        cr.formatedDate(),
		Description: cr.Alt,
		Image:       cr.Img,
	}
}

func (c Comic) PrettyString() string {
	return fmt.Sprintf("Title: %s\nNumber: %d\nDescription: %s\nDate: %s",
		c.Title, c.Number, c.Description, c.Date)
}

func (c Comic) JSON() string {
	cJSON, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(cJSON)
}

func (x *XKCDClient) setTimeout(d time.Duration) {

	x.client.Timeout = d
}

func (x *XKCDClient) buildURL(n comicNumber) string {
	var fullURL string
	if n == LatestComic {

		fullURL = fmt.Sprintf("%s/info.0.json", x.baseURL)
	} else {

		fullURL = fmt.Sprintf("%s/%d/info.0.json", x.baseURL, int(n))
	}
	return fullURL
}

func (x *XKCDClient) Fetch(number comicNumber, save bool) (Comic, error) {
	resp, err := x.client.Get(x.buildURL(number))

	if err != nil {
		return Comic{}, err
	}
	defer resp.Body.Close()

	var cr ComicResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return Comic{}, err
	}

	if save {
		err := x.Save(cr.Img, ".")
		if err != nil {
			return Comic{}, err
		}
	}
	return cr.Comic(), nil
}

func (x *XKCDClient) Save(url, fpath string) error {

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	absPath, _ := filepath.Abs(fpath)
	filePath := fmt.Sprintf("%s/%s", absPath, path.Base(url))

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func main() {

	comicNo := flag.Int("n", int(LatestComic), "Comic number to download")
	save := flag.Bool("s", false, "Save imgeg on hard drive")
	outputType := flag.String("o", "", "Print metadata to stdout\nAvaible: text, JSON")
	flag.Parse()

	var clientInstance XKCDClient

	clientInstance.baseURL = BaseURL
	clientInstance.client = &http.Client{Timeout: DefaultTimeout}

	c, err := clientInstance.Fetch(comicNumber(*comicNo), *save)

	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	switch strings.ToLower(*outputType) {
	case "text":
		fmt.Println(c.PrettyString())
	case "json":
		fmt.Println(c.JSON())

	}
}
