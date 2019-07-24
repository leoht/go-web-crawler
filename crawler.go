package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type PageAssetType string

const (
	Stylesheet PageAssetType = "stylesheet"
	Image      PageAssetType = "image"
	Script     PageAssetType = "script"
)

type PageAsset struct {
	Type PageAssetType `json:"type"`
	URL  string        `json:"url"`
}

type Page struct {
	URL    string      `json:"url"`
	Assets []PageAsset `json:"assets"`
}
type Sitemap struct {
	Pages []Page `json:"pages"`
}

func NewSitemap() *Sitemap {
	return &Sitemap{
		Pages: make([]Page, 0),
	}
}

func NewPageWithAssets(url string, assets []PageAsset) Page {
	return Page{
		URL:    strings.TrimRight(url, "/"),
		Assets: assets,
	}
}

func (s *Sitemap) AddPage(page Page) {
	s.Pages = append(s.Pages, page)
}

func (s *Sitemap) HasPage(url string) bool {
	for _, page := range s.Pages {
		if page.URL == url {
			return true
		}
	}
	return false
}

type Crawler struct {
	RootDomain string
}

func NewCrawler(rootDomain string) Crawler {
	return Crawler{
		RootDomain: rootDomain,
	}
}

// Start the crawling process
// Launches goroutines for each page to crawl and get pages through a channel
// Returns when there's no more new pages to crawl
func (c Crawler) CrawlWebsite() (*Sitemap, error) {
	linksChan := make(chan string)
	pageChan := make(chan Page)
	pageOkChan := make(chan bool)
	rootURL := "http://" + c.RootDomain
	sitemap := NewSitemap()

	go crawlPage(rootURL, linksChan, pageChan, pageOkChan, sitemap)

	for pageOkCount := 0; pageOkCount <= len(sitemap.Pages); {
		select {
		case url := <-linksChan:
			if strings.HasPrefix(url, "/") {
				url = rootURL + url
			}
			if c.shouldFollowUrl(url) && !sitemap.HasPage(url) {
				go crawlPage(url, linksChan, pageChan, pageOkChan, sitemap)
			}
		case page := <-pageChan:
			if !sitemap.HasPage(page.URL) {
				sitemap.AddPage(page)
			}
		case <-pageOkChan:
			pageOkCount++
		}
	}

	return sitemap, nil
}

func (c Crawler) shouldFollowUrl(urlString string) bool {
	url, err := url.Parse(urlString)
	if err != nil {
		return false
	}

	return url.Host == c.RootDomain && strings.HasPrefix(url.Scheme, "http")
}

func crawlPage(url string, linksChan chan string, pageChan chan Page, pageOkChan chan bool, sitemap *Sitemap) {
	resp, err := http.Get(url)

	if err != nil {
		fmt.Println(err)
		return
	}

	body := resp.Body
	defer body.Close()

	assets, newLinksFound := tokenizeAndCrawlElements(body, linksChan, sitemap)

	if newLinksFound == 0 {
		pageOkChan <- true
	}

	pageChan <- NewPageWithAssets(url, assets)
}

// Tokenization and parsing of HTML elements
// for assets and links
func tokenizeAndCrawlElements(body io.Reader, linksChan chan string, sitemap *Sitemap) (assets []PageAsset, newLinksFound int) {
	tokenizer := html.NewTokenizer(body)
	assets = make([]PageAsset, 0)

	for {
		nextType := tokenizer.Next()

		switch {
		case nextType == html.StartTagToken:
			token := tokenizer.Token()
			if isTag(token, "a") {
				href := parseLinkHref(token, linksChan)
				if href != "" && !sitemap.HasPage(href) {
					newLinksFound++
				}
			}
			if isStylesheetTag(token) {
				assets = append(assets, parseAssetFromToken(token, "href", Stylesheet))
			}
			if isScriptTag(token) {
				assets = append(assets, parseAssetFromToken(token, "src", Script))
			}
			if isTag(token, "img") {
				assets = append(assets, parseAssetFromToken(token, "src", Image))
			}
			continue
		case nextType == html.ErrorToken:
			return assets, newLinksFound
		}
	}
}

func parseLinkHref(token html.Token, linksChan chan string) (href string) {
	if url := getHtmlAttribute(token, "href"); url != "" {
		linksChan <- url
		return url
	}
	return ""
}

func parseAssetFromToken(token html.Token, attrName string, assetType PageAssetType) PageAsset {
	if urlValue := getHtmlAttribute(token, attrName); urlValue != "" {
		return PageAsset{
			Type: assetType,
			URL:  urlValue,
		}
	}
	return PageAsset{}
}

func retrievePageBody(uri string) (io.Reader, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func isTag(token html.Token, tag string) bool {
	return token.Data == tag
}

func isStylesheetTag(token html.Token) bool {
	return isTag(token, "link") && getHtmlAttribute(token, "rel") == "stylesheet"
}

func isScriptTag(token html.Token) bool {
	return isTag(token, "script") && getHtmlAttribute(token, "src") != ""
}

func getHtmlAttribute(t html.Token, name string) (value string) {
	for _, a := range t.Attr {
		if a.Key == name {
			value = a.Val
		}
	}
	return value
}
