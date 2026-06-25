package youtube

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// httpClient is a shared client with a 30-second timeout.
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

const maxResponseSize = 2 * 1024 * 1024 // 2MB cho HTML page

// oEmbedResponse is the JSON structure returned by the YouTube oEmbed API.
type oEmbedResponse struct {
	Title        string `json:"title"`
	AuthorName   string `json:"author_name"`
	AuthorURL    string `json:"author_url"`
	ThumbnailURL string `json:"thumbnail_url"`
	ThumbnailW   int    `json:"thumbnail_width"`
	ThumbnailH   int    `json:"thumbnail_height"`
	ProviderName string `json:"provider_name"`
}

// videoInfo holds all video information after aggregation.
type videoInfo struct {
	Title       string
	Channel     string
	ChannelURL  string
	Description string
	Views       string
	Duration    string
	PublishDate string
	Thumbnail   string
	VideoURL    string
	VideoID     string
}

// GetVideoInfo retrieves detailed information for a YouTube video from its URL.
// Uses the oEmbed API + HTML meta parsing — no API key required.
func GetVideoInfo(videoURL string) (string, error) {
	// Step 1: Parse and validate URL, extract Video ID
	videoID, err := extractVideoID(videoURL)
	if err != nil {
		return "", err
	}

	canonicalURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)

	info := &videoInfo{
		VideoID:  videoID,
		VideoURL: canonicalURL,
	}

	// Step 2: Call oEmbed API to get basic info
	if err := fetchOEmbed(canonicalURL, info); err != nil {
		// oEmbed failure is not critical, continue with HTML fallback
		info.Title = "(unavailable from oEmbed)"
	}

	// Step 3: Fetch HTML page to get additional metadata
	fetchHTMLMeta(canonicalURL, info)

	// Step 4: Format result
	return formatResult(info), nil
}

// extractVideoID parses a YouTube URL and returns the Video ID.
// Supports the following formats:
//   - https://www.youtube.com/watch?v=VIDEO_ID
//   - https://youtu.be/VIDEO_ID
//   - https://www.youtube.com/shorts/VIDEO_ID
//   - https://www.youtube.com/embed/VIDEO_ID
//   - https://m.youtube.com/watch?v=VIDEO_ID
func extractVideoID(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("only http/https is supported, got: '%s'", u.Scheme)
	}

	host := strings.ToLower(u.Hostname())

	// Format: youtu.be/VIDEO_ID
	if host == "youtu.be" {
		id := strings.TrimPrefix(u.Path, "/")
		if id == "" {
			return "", fmt.Errorf("video ID not found in URL: %s", rawURL)
		}
		return strings.Split(id, "/")[0], nil
	}

	// Only allow YouTube domains
	if host != "www.youtube.com" && host != "youtube.com" && host != "m.youtube.com" {
		return "", fmt.Errorf("URL is not a YouTube domain: %s", host)
	}

	// Format: /watch?v=VIDEO_ID
	if u.Path == "/watch" {
		id := u.Query().Get("v")
		if id == "" {
			return "", fmt.Errorf("URL is missing 'v' parameter: %s", rawURL)
		}
		return id, nil
	}

	// Format: /shorts/VIDEO_ID or /embed/VIDEO_ID
	pathParts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(pathParts) >= 2 && (pathParts[0] == "shorts" || pathParts[0] == "embed" || pathParts[0] == "v") {
		return pathParts[1], nil
	}

	return "", fmt.Errorf("could not identify Video ID from URL: %s", rawURL)
}

// fetchOEmbed calls the YouTube oEmbed API.
func fetchOEmbed(videoURL string, info *videoInfo) error {
	oembedURL := fmt.Sprintf("https://www.youtube.com/oembed?url=%s&format=json",
		url.QueryEscape(videoURL))

	resp, err := httpClient.Get(oembedURL)
	if err != nil {
		return fmt.Errorf("oEmbed request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("oEmbed returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return fmt.Errorf("cannot read oEmbed response: %w", err)
	}

	var oembed oEmbedResponse
	if err := json.Unmarshal(body, &oembed); err != nil {
		return fmt.Errorf("cannot parse oEmbed JSON: %w", err)
	}

	info.Title = oembed.Title
	info.Channel = oembed.AuthorName
	info.ChannelURL = oembed.AuthorURL
	info.Thumbnail = oembed.ThumbnailURL

	return nil
}

// fetchHTMLMeta fetches the video's HTML page and parses meta tags.
func fetchHTMLMeta(videoURL string, info *videoInfo) {
	req, err := http.NewRequest(http.MethodGet, videoURL, nil)
	if err != nil {
		return
	}

	// Simulate a browser so YouTube returns full HTML
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return
	}

	html := string(body)

	// Parse OpenGraph & Schema.org meta tags
	if info.Description == "" {
		info.Description = extractMeta(html, "og:description")
		if info.Description == "" {
			info.Description = extractMeta(html, "description")
		}
	}

	if info.Title == "" || info.Title == "(unavailable from oEmbed)" {
		if t := extractMeta(html, "og:title"); t != "" {
			info.Title = t
		}
	}

	if info.PublishDate == "" {
		info.PublishDate = extractMeta(html, "datePublished")
		if info.PublishDate == "" {
			info.PublishDate = extractMetaItemProp(html, "datePublished")
		}
		if info.PublishDate == "" {
			info.PublishDate = extractMetaItemProp(html, "uploadDate")
		}
	}

	if info.Duration == "" {
		dur := extractMeta(html, "duration")
		if dur == "" {
			dur = extractMetaItemProp(html, "duration")
		}
		if dur != "" {
			info.Duration = formatISO8601Duration(dur)
		}
	}

	if info.Views == "" {
		info.Views = extractMetaItemProp(html, "interactionCount")
	}

	if info.Channel == "" {
		// Fallback: try getting from link tag
		info.Channel = extractLinkItemprop(html, "name")
	}

	if info.Thumbnail == "" {
		info.Thumbnail = extractMeta(html, "og:image")
	}
}

// extractMeta extracts the content value from a <meta property="name" content="value">
// or <meta name="name" content="value"> tag.
func extractMeta(html, name string) string {
	// Pattern for property="name" or name="name"
	patterns := []string{
		fmt.Sprintf(`<meta\s+(?:[^>]*?\s)?property="%s"\s+content="([^"]*)"`, regexp.QuoteMeta(name)),
		fmt.Sprintf(`<meta\s+(?:[^>]*?\s)?content="([^"]*)"\s+(?:[^>]*?\s)?property="%s"`, regexp.QuoteMeta(name)),
		fmt.Sprintf(`<meta\s+(?:[^>]*?\s)?name="%s"\s+content="([^"]*)"`, regexp.QuoteMeta(name)),
		fmt.Sprintf(`<meta\s+(?:[^>]*?\s)?content="([^"]*)"\s+(?:[^>]*?\s)?name="%s"`, regexp.QuoteMeta(name)),
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) >= 2 {
			return decodeHTMLEntities(matches[1])
		}
	}
	return ""
}

// extractMetaItemProp extracts the value from a <meta itemprop="name" content="value"> tag.
func extractMetaItemProp(html, name string) string {
	patterns := []string{
		fmt.Sprintf(`<meta\s+itemprop="%s"\s+content="([^"]*)"`, regexp.QuoteMeta(name)),
		fmt.Sprintf(`<meta\s+content="([^"]*)"\s+itemprop="%s"`, regexp.QuoteMeta(name)),
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) >= 2 {
			return decodeHTMLEntities(matches[1])
		}
	}
	return ""
}

// extractLinkItemprop extracts content from <span itemprop="name">value</span> or similar.
func extractLinkItemprop(html, name string) string {
	pattern := fmt.Sprintf(`<link\s+itemprop="%s"\s+content="([^"]*)"`, regexp.QuoteMeta(name))
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(html)
	if len(matches) >= 2 {
		return decodeHTMLEntities(matches[1])
	}
	return ""
}

// formatISO8601Duration converts an ISO 8601 duration (PT4M13S) to a readable format (4:13).
func formatISO8601Duration(iso string) string {
	iso = strings.TrimSpace(iso)
	if !strings.HasPrefix(iso, "PT") {
		return iso
	}

	re := regexp.MustCompile(`PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?`)
	matches := re.FindStringSubmatch(iso)
	if len(matches) < 4 {
		return iso
	}

	hours := matches[1]
	minutes := matches[2]
	seconds := matches[3]

	if minutes == "" {
		minutes = "0"
	}
	if seconds == "" {
		seconds = "00"
	} else if len(seconds) == 1 {
		seconds = "0" + seconds
	}

	if hours != "" && hours != "0" {
		if len(minutes) == 1 {
			minutes = "0" + minutes
		}
		return fmt.Sprintf("%s:%s:%s", hours, minutes, seconds)
	}

	return fmt.Sprintf("%s:%s", minutes, seconds)
}

// decodeHTMLEntities decodes common HTML entities.
func decodeHTMLEntities(s string) string {
	replacer := strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", "\"",
		"&#39;", "'",
		"&apos;", "'",
		"&#x27;", "'",
		"&#x2F;", "/",
	)
	return replacer.Replace(s)
}

// formatResult formats video info into a structured string.
func formatResult(info *videoInfo) string {
	var sb strings.Builder

	sb.WriteString("VIDEO YOUTUBE INFO\n")
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	writeField(&sb, "Title", info.Title)
	writeField(&sb, "Channel", info.Channel)
	if info.ChannelURL != "" {
		writeField(&sb, "Channel URL", info.ChannelURL)
	}
	writeField(&sb, "Published", info.PublishDate)
	writeField(&sb, "Duration", info.Duration)
	writeField(&sb, "Views", formatViewCount(info.Views))

	sb.WriteString(strings.Repeat("─", 60) + "\n")

	if info.Description != "" {
		sb.WriteString("Description:\n")
		// Limit description to 500 characters to keep output concise
		desc := info.Description
		if len(desc) > 500 {
			desc = desc[:500] + "..."
		}
		sb.WriteString(desc + "\n")
	}

	sb.WriteString(strings.Repeat("─", 60) + "\n")
	writeField(&sb, "Video ID", info.VideoID)
	writeField(&sb, "Thumbnail", info.Thumbnail)
	writeField(&sb, "URL", info.VideoURL)

	return sb.String()
}

func writeField(sb *strings.Builder, label, value string) {
	if value == "" {
		value = "(no data)"
	}
	sb.WriteString(fmt.Sprintf("%-13s: %s\n", label, value))
}

// formatViewCount formats a view count number for readability.
func formatViewCount(views string) string {
	if views == "" {
		return ""
	}

	// Add thousands separators
	n := len(views)
	if n <= 3 {
		return views
	}

	var result strings.Builder
	remainder := n % 3
	if remainder > 0 {
		result.WriteString(views[:remainder])
	}
	for i := remainder; i < n; i += 3 {
		if result.Len() > 0 {
			result.WriteString(".")
		}
		result.WriteString(views[i : i+3])
	}

	return result.String() + " views"
}
