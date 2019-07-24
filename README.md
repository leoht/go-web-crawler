# Go web crawler

Very simple crawler that uses goroutines to parse web pages and get the list of pages with their static assets (stylesheets, scripts and images).

To install dependencies:
```
$ brew install dep
$ dep ensure
```

To crawl a domain (example): 
```
$ go run *.go "leohetsch.com"
```

The program also outputs a json of the sitemap
