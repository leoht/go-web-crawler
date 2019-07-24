package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Please specify root URL to crawl")
		os.Exit(1)
	}
	rootURL := args[0]

	c := NewCrawler(rootURL)
	sitemap, err := c.CrawlWebsite()

	if err != nil {
		log.Fatal(err)
	}

	jsonBytes, _ := json.Marshal(sitemap)
	_ = ioutil.WriteFile("sitemap.json", jsonBytes, 0644)

	return
}
