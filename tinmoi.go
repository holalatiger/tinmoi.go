package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/mmcdole/gofeed"
)

var rssSources = map[string]string{
	"TechCrunch":     "http://feeds.feedburner.com/TechCrunch/",
	"Wired":          "https://www.wired.com/feed/rss",
	"The Verge":      "https://www.theverge.com/rss/index.xml",
	"Ars Technica":   "http://feeds.arstechnica.com/arstechnica/index/",
	"Mashable":       "http://feeds.mashable.com/Mashable",
	"Hacker News":    "https://news.ycombinator.com/rss",
	"Product Hunt":   "https://www.producthunt.com/feed",
	"Engadget":       "https://www.engadget.com/rss.xml",
	"VentureBeat":    "https://venturebeat.com/feed/",
	"Gizmodo":        "https://gizmodo.com/rss",
	"Themeisle News": "https://themeisle.com/blog/rss-feeds-list/#news",
}

type Article struct {
	Source  string `json:"source"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Summary string `json:"summary"`
}

type Report struct {
	Timestamp      time.Time `json:"timestamp"`
	SuccessSources []string  `json:"successSources"`
	FailSources    []string  `json:"failSources"`
	Articles       []Article `json:"articles"`
}

type ManualArticle struct {
	Article
	ManualSummary string `json:"manual_summary,omitempty"`
	ManualOpinion string `json:"manual_opinion,omitempty"`
}

type ManualReport struct {
	Timestamp      time.Time       `json:"timestamp"`
	SuccessSources []string        `json:"successSources"`
	FailSources    []string        `json:"failSources"`
	Articles       []ManualArticle `json:"articles"`
}

var (
	lastReport Report
	mutex      sync.Mutex
)

const manualReportFile = "report_manually_update.json"

func trimSummary(text string, length int) string {
	if len(text) > length {
		return text[:length] + "..."
	}
	return text
}

func printProgressBar(current, total int, eta time.Duration) {
	barWidth := 40
	ratio := float64(current) / float64(total)
	filled := int(ratio * float64(barWidth))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	orange := color.New(color.FgHiYellow).Add(color.BgRed).SprintFunc()
	etaStr := fmt.Sprintf("%02d phút %02d giây", int(eta.Minutes()), int(eta.Seconds())%60)
	fmt.Printf("\r%s ETA: %s %d/%d (%.1f%%) ", orange("Hệ thống đang lấy tin, tiến độ: ["+bar+"]"), etaStr, current, total, ratio*100)
}

func crawlRSSFeeds() (successSources, failSources []string, articles []Article) {
	fp := gofeed.NewParser()
	var mu sync.Mutex
	var wg sync.WaitGroup

	totalSources := len(rssSources)
	startTime := time.Now()

	currentSource := 0

	for source, url := range rssSources {
		wg.Add(1)
		go func(source, url string) {
			defer wg.Done()
			feed, err := fp.ParseURL(url)
			mu.Lock()
			defer mu.Unlock()
			currentSource++
			elapsed := time.Since(startTime)
			avgPerSource := elapsed / time.Duration(currentSource)
			remaining := avgPerSource * time.Duration(totalSources-currentSource)
			printProgressBar(currentSource, totalSources, remaining)

			if err != nil {
				failSources = append(failSources, source)
				return
			}
			successSources = append(successSources, source)
			count := 0
			for _, item := range feed.Items {
				if count >= 5 {
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
	fmt.Print("\n")
	return
}

func saveReportToFile(filename string, report interface{}) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

func loadReportFromFile(filename string, report interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, report)
}

func printResults(report Report) {
	blue := color.New(color.FgHiBlue).SprintFunc()
	purple := color.New(color.FgMagenta).SprintFunc()
	black := color.New(color.FgBlack).SprintFunc()
	boldBlue := color.New(color.FgHiBlue).Add(color.Bold).SprintFunc()

	fmt.Printf("0. Thống kê quá trình crawl tin:\n")
	fmt.Printf("Số nguồn crawl thành công: %d\n", len(report.SuccessSources))
	fmt.Printf("Số nguồn crawl thất bại: %d\n", len(report.FailSources))
	fmt.Printf("Danh sách nguồn crawl thành công: %s\n", strings.Join(report.SuccessSources, ", "))
	fmt.Printf("Danh sách nguồn crawl thất bại: %s\n\n", strings.Join(report.FailSources, ", "))

	fmt.Printf("1. Tin mới nhất:\n")
	for i, article := range report.Articles {
		fmt.Printf("%d. Nguồn tin: %s\n", i+1, blue(article.Source))
		fmt.Printf("   Tiêu đề: %s\n", black(article.Title))
		fmt.Printf("   URL: %s\n", purple(article.URL))
		fmt.Printf("   Tóm tắt ngắn gọn: %s\n", black(trimSummary(article.Summary, 300)))
		fmt.Printf("   Nhận định:\n")
		fmt.Printf("   - Đây là bài viết mới gần đây từ nguồn uy tín trong lĩnh vực công nghệ và khởi nghiệp.\n")
		fmt.Printf("   - Phần tóm tắt cung cấp thông tin cơ bản và hữu ích để cập nhật xu hướng.\n")
		fmt.Printf("   - Cần đọc kỹ bài đầy đủ để hiểu chi tiết và tác động.\n\n")
	}

	fmt.Printf("Đã crawl %s, vui lòng vào mục thống kê để xem chi tiết.\n\n", boldBlue(fmt.Sprintf("%d *** đầu báo ***", len(rssSources))))
}

func readMultilineInput(prompt string) string {
	fmt.Println(prompt + " (Nhập -- rồi Enter để kết thúc):")
	reader := bufio.NewReader(os.Stdin)
	var lines []string
	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		if line == "--" {
			break
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func manualUpdateReport() {
	mutex.Lock()
	if lastReport.Timestamp.IsZero() {
		fmt.Println("Chưa có báo cáo crawl tin nào. Vui lòng crawl tin trước.")
		mutex.Unlock()
		return
	}

	manualReport := ManualReport{
		Timestamp:      lastReport.Timestamp,
		SuccessSources: lastReport.SuccessSources,
		FailSources:    lastReport.FailSources,
		Articles:       make([]ManualArticle, len(lastReport.Articles)),
	}

	for i, art := range lastReport.Articles {
		manualReport.Articles[i] = ManualArticle{Article: art}
	}
	mutex.Unlock()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n=== Danh sách các tin đã crawl ===")
		for i, art := range manualReport.Articles {
			fmt.Printf("%d: %s\n", i+1, art.Title)
		}
		fmt.Println("Nhập số thứ tự tin muốn cập nhật (0 để thoát):")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "0" {
			break
		}
		idx, err := strconv.Atoi(input)
		if err != nil || idx < 1 || idx > len(manualReport.Articles) {
			fmt.Println("Số thứ tự không hợp lệ, vui lòng thử lại.")
			continue
		}
		idx-- // zero-based

		art := &manualReport.Articles[idx]

		fmt.Printf("\nTiêu đề tin: %s\n", art.Title)
		fmt.Print("Nhập Tóm tắt ngắn gọn (tiếng Việt): ")
		summary, _ := reader.ReadString('\n')
		summary = strings.TrimSpace(summary)
		opinion := readMultilineInput("Nhập Nhận định (tiếng Việt)")
		art.ManualSummary = summary
		art.ManualOpinion = opinion

		fmt.Printf("Đã cập nhật tin thứ %d thành công.\n", idx+1)
	}

	err := saveReportToFile(manualReportFile, manualReport)
	if err != nil {
		fmt.Println("Lỗi khi lưu báo cáo thủ công:", err)
	} else {
		fmt.Println("Đã lưu báo cáo cập nhật thủ công vào file:", manualReportFile)
	}
}

func extractManualReport() {
	var manualReport ManualReport
	err := loadReportFromFile(manualReportFile, &manualReport)
	if err != nil {
		fmt.Println("Lỗi hoặc chưa có báo cáo cập nhật thủ công:", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n=== Báo cáo tổng đã cập nhật thủ công ===")
		for i, art := range manualReport.Articles {
			fmt.Printf("%d: %s\n", i+1, art.Title)
		}
		fmt.Println("Nhập số thứ tự muốn xem chi tiết (0 để thoát):")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "0" {
			break
		}
		idx, err := strconv.Atoi(input)
		if err != nil || idx < 1 || idx > len(manualReport.Articles) {
			fmt.Println("Số thứ tự không hợp lệ, vui lòng thử lại.")
			continue
		}
		idx-- // zero-based

		art := manualReport.Articles[idx]
		blue := color.New(color.FgHiBlue).SprintFunc()
		purple := color.New(color.FgMagenta).SprintFunc()
		black := color.New(color.FgBlack).SprintFunc()

		fmt.Printf("\nNguồn tin: %s\n", blue(art.Source))
		fmt.Printf("Tiêu đề: %s\n", black(art.Title))
		fmt.Printf("URL: %s\n", purple(art.URL))

		if art.ManualSummary != "" {
			fmt.Printf("Tóm tắt ngắn gọn (thủ công): %s\n", black(trimSummary(art.ManualSummary, 300)))
		} else {
			fmt.Printf("Tóm tắt ngắn gọn: %s\n", black(trimSummary(art.Summary, 300)))
		}

		if art.ManualOpinion != "" {
			fmt.Printf("Nhận định (thủ công):\n%s\n\n", art.ManualOpinion)
		} else {
			fmt.Printf("Nhận định: Chưa có nhận định thủ công.\n\n")
		}
	}
}

func menu() {
	fmt.Println("Chọn mục:")
	fmt.Println("1 - Bắt đầu crawl tin")
	fmt.Println("2 - Trích xuất báo cáo tin đã crawl")
	fmt.Println("3 - Cập nhật thủ công báo cáo")
	fmt.Println("4 - Trích xuất báo cáo tổng (báo cáo thủ công mới nhất)")
	fmt.Println("0 - Thoát")
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		menu()
		fmt.Print("Nhập lựa chọn: ")
		choiceStr, _ := reader.ReadString('\n')
		choiceStr = strings.TrimSpace(choiceStr)

		switch choiceStr {
		case "1":
			fmt.Println("Đang bắt đầu crawl tin...")
			successSources, failSources, articles := crawlRSSFeeds()
			mutex.Lock()
			lastReport = Report{
				Timestamp:      time.Now(),
				SuccessSources: successSources,
				FailSources:    failSources,
				Articles:       articles,
			}
			mutex.Unlock()
			printResults(lastReport)
		case "2":
			mutex.Lock()
			if lastReport.Timestamp.IsZero() {
				fmt.Println("Chưa có báo cáo crawl tin nào. Vui lòng crawl tin trước.")
				mutex.Unlock()
				continue
			}
			printResults(lastReport)
			mutex.Unlock()
		case "3":
			manualUpdateReport()
		case "4":
			extractManualReport()
		case "0":
			fmt.Println("Thoát chương trình.")
			return
		default:
			fmt.Println("Lựa chọn không hợp lệ, vui lòng thử lại.")
		}
		fmt.Println()
	}
}
