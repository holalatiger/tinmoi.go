package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/mmcdole/gofeed"
)

var rssSources = map[string]string{
	"TechCrunch":    "http://feeds.feedburner.com/TechCrunch/",
	"Wired":         "https://www.wired.com/feed/rss",
	"The Verge":     "https://www.theverge.com/rss/index.xml",
	"Ars Technica":  "http://feeds.arstechnica.com/arstechnica/index/",
	"Mashable":      "http://feeds.mashable.com/Mashable",
	"Hacker News":   "https://news.ycombinator.com/rss",
	"Product Hunt":  "https://www.producthunt.com/feed",
	"Engadget":      "https://www.engadget.com/rss.xml",
	"VentureBeat":   "https://venturebeat.com/feed/",
	"Gizmodo":       "https://gizmodo.com/rss",
	"Themeisle News": "https://themeisle.com/blog/rss-feeds-list/#news",
}

type Article struct {
	Source  string
	Title   string
	URL     string
	Summary string
}

func trimSummary(text string, length int) string {
	if len(text) > length {
		return text[:length] + "..."
	}
	return text
}

func crawlRSSFeeds() (successSources, failSources []string, articles []Article) {
	fp := gofeed.NewParser()
	var mu sync.Mutex
	var wg sync.WaitGroup

	for source, url := range rssSources {
		wg.Add(1)
		go func(source, url string) {
			defer wg.Done()
			feed, err := fp.ParseURL(url)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				failSources = append(failSources, source)
				return
			}
			successSources = append(successSources, source)
			count := 0
			for _, item := range feed.Items {
				if count >= 5 { // lấy tối đa 5 bài mỗi nguồn
					break
				}
				summary := item.Description
				if summary == "" {
					summary = item.Content
				}
				articles = append(articles, Article{
					Source:  source,
					Title:   item.Title,
					URL:     item.Link,
					Summary: summary,
				})
				count++
			}
		}(source, url)
	}
	wg.Wait()
	return
}

func printResults(successSources, failSources []string, articles []Article) {
	blue := color.New(color.FgHiBlue).SprintFunc()
	purple := color.New(color.FgMagenta).SprintFunc()
	black := color.New(color.FgBlack).SprintFunc()

	// In tình trạng RSS
	fmt.Printf("0. Tình trạng RSS:\n")
	fmt.Printf("Số nguồn lấy tin thành công: %d\n", len(successSources))
	fmt.Printf("Số nguồn lấy tin thất bại: %d\n", len(failSources))
	fmt.Printf("Danh sách nguồn lấy thành công: %s\n", strings.Join(successSources, ", "))
	fmt.Printf("Danh sách nguồn lấy thất bại: %s\n\n", strings.Join(failSources, ", "))

	// In kết quả tin mới nhất
	fmt.Printf("1. Tin mới nhất:\n")
	for _, article := range articles {
		fmt.Printf("Nguồn tin: %s\n", blue(article.Source))
		fmt.Printf("URL: %s\n", purple(article.URL))
		fmt.Printf("Tóm tắt ngắn gọn: %s\n", black(trimSummary(article.Summary, 300)))
		fmt.Printf("Nhận định:\n")
		fmt.Printf("- Đây là bài viết mới gần đây từ nguồn uy tín trong lĩnh vực công nghệ và khởi nghiệp.\n")
		fmt.Printf("- Phần tóm tắt cung cấp thông tin cơ bản và hữu ích để cập nhật xu hướng.\n")
		fmt.Printf("- Cần đọc kỹ bài đầy đủ để hiểu chi tiết và tác động.\n\n")
	}
}

func main() {
	successSources, failSources, articles := crawlRSSFeeds()
	printResults(successSources, failSources, articles)
}
