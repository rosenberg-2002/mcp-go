package youtube

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// transcriptList is the XML structure containing transcript segments.
type transcriptList struct {
	XMLName xml.Name         `xml:"transcript"`
	Texts   []transcriptText `xml:"text"`
}

// transcriptText is a text segment in the transcript with start time and duration.
type transcriptText struct {
	Start string `xml:"start,attr"`
	Dur   string `xml:"dur,attr"`
	Text  string `xml:",chardata"`
}

// captionTrack holds information about a subtitle track from ytInitialPlayerResponse.
type captionTrack struct {
	BaseURL  string
	Name     string
	LangCode string
	Kind     string // "asr" = auto-generated
}

// GetTranscript retrieves the transcript (subtitles) of a YouTube video.
// lang: language code (e.g. "en", "ja"). Leave empty to get the first available track.
// Returns transcript content as text with timestamps.
func GetTranscript(videoURL string, lang string) (string, error) {
	// Step 1: Extract Video ID
	videoID, err := extractVideoID(videoURL)
	if err != nil {
		return "", err
	}

	canonicalURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)

	// Step 2: Fetch HTML page to get caption track URLs
	html, err := fetchVideoHTML(canonicalURL)
	if err != nil {
		return "", fmt.Errorf("không thể tải trang video: %w", err)
	}

	// Step 3: Parse ytInitialPlayerResponse to get caption track list
	tracks, err := extractCaptionTracks(html)
	if err != nil {
		return "", err
	}

	if len(tracks) == 0 {
		return "", fmt.Errorf("this video has no subtitles/transcript")
	}

	// Step 4: Select appropriate track
	track := selectTrack(tracks, lang)

	// Step 5: Fetch and parse transcript XML
	transcript, err := fetchTranscriptXML(track.BaseURL)
	if err != nil {
		return "", fmt.Errorf("failed to load transcript: %w", err)
	}

	// Step 6: Format result
	return formatTranscript(videoID, track, transcript), nil
}

// GetAvailableLanguages returns a list of available subtitle languages for a video.
func GetAvailableLanguages(videoURL string) (string, error) {
	videoID, err := extractVideoID(videoURL)
	if err != nil {
		return "", err
	}

	canonicalURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)

	html, err := fetchVideoHTML(canonicalURL)
	if err != nil {
		return "", fmt.Errorf("failed to load video page: %w", err)
	}

	tracks, err := extractCaptionTracks(html)
	if err != nil {
		return "", err
	}

	if len(tracks) == 0 {
		return "This video has no subtitles/transcript.", nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("AVAILABLE SUBTITLE LANGUAGES (Video: %s)\n", videoID))
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	for i, t := range tracks {
		kind := "Manual"
		if t.Kind == "asr" {
			kind = "Auto-generated (ASR)"
		}
		sb.WriteString(fmt.Sprintf("  %d. [%s] %s — %s\n", i+1, t.LangCode, t.Name, kind))
	}

	sb.WriteString(strings.Repeat("─", 60) + "\n")
	sb.WriteString(fmt.Sprintf("Total: %d language(s)\n", len(tracks)))
	sb.WriteString("Tip: Use the get_youtube_transcript tool with the lang parameter to select a language.")

	return sb.String(), nil
}

// fetchVideoHTML fetches the HTML page of a YouTube video.
func fetchVideoHTML(videoURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, videoURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("YouTube returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024)) // 5MB max cho HTML
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// extractCaptionTracks parses ytInitialPlayerResponse from HTML to get caption tracks.
func extractCaptionTracks(html string) ([]captionTrack, error) {
	// Find ytInitialPlayerResponse in HTML
	// YouTube embeds it as: var ytInitialPlayerResponse = {...};
	re := regexp.MustCompile(`ytInitialPlayerResponse\s*=\s*(\{.+?\})\s*;`)
	matches := re.FindStringSubmatch(html)

	if len(matches) < 2 {
		// Try another pattern: ytInitialPlayerResponse may be in a different script tag
		re2 := regexp.MustCompile(`ytInitialPlayerResponse\s*=\s*(\{.+?\})\s*;\s*var`)
		matches = re2.FindStringSubmatch(html)
	}

	if len(matches) < 2 {
		// Fallback: find captionTracks directly in HTML
		return extractCaptionTracksDirectly(html)
	}

	return parseCaptionTracksFromJSON(matches[1])
}

// extractCaptionTracksDirectly finds captionTracks directly in HTML
// by regex-ing baseUrl entries in playerCaptionsTracklistRenderer.
func extractCaptionTracksDirectly(html string) ([]captionTrack, error) {
	// Find all baseUrls of caption tracks
	re := regexp.MustCompile(`"captionTracks"\s*:\s*\[(.*?)\]`)
	matches := re.FindStringSubmatch(html)
	if len(matches) < 2 {
		return nil, nil // No captions, not an error
	}

	captionJSON := matches[1]
	return parseCaptionEntries(captionJSON)
}

// parseCaptionTracksFromJSON parses caption tracks from ytInitialPlayerResponse JSON.
func parseCaptionTracksFromJSON(jsonStr string) ([]captionTrack, error) {
	// Find captionTracks array in JSON
	re := regexp.MustCompile(`"captionTracks"\s*:\s*\[(.*?)\]`)
	matches := re.FindStringSubmatch(jsonStr)
	if len(matches) < 2 {
		return nil, nil
	}

	return parseCaptionEntries(matches[1])
}

// parseCaptionEntries parses each entry in the captionTracks JSON array.
func parseCaptionEntries(captionJSON string) ([]captionTrack, error) {
	var tracks []captionTrack

	// Find each track object: {"baseUrl":"...","name":{"simpleText":"..."},"languageCode":"...","kind":"..."}
	// Regex cho baseUrl
	reBase := regexp.MustCompile(`"baseUrl"\s*:\s*"(.*?)"`)
	reLang := regexp.MustCompile(`"languageCode"\s*:\s*"(.*?)"`)
	reName := regexp.MustCompile(`"simpleText"\s*:\s*"(.*?)"`)
	reKind := regexp.MustCompile(`"kind"\s*:\s*"(.*?)"`)

	// Split by object boundaries
	// Each track starts with {"baseUrl"
	trackSections := splitCaptionTracks(captionJSON)

	for _, section := range trackSections {
		track := captionTrack{}

		if m := reBase.FindStringSubmatch(section); len(m) >= 2 {
			track.BaseURL = unescapeJSON(m[1])
		} else {
			continue // Skip entries without baseUrl
		}

		if m := reLang.FindStringSubmatch(section); len(m) >= 2 {
			track.LangCode = m[1]
		}

		if m := reName.FindStringSubmatch(section); len(m) >= 2 {
			track.Name = unescapeJSON(m[1])
		}

		if m := reKind.FindStringSubmatch(section); len(m) >= 2 {
			track.Kind = m[1]
		}

		tracks = append(tracks, track)
	}

	return tracks, nil
}

// splitCaptionTracks splits the captionTracks JSON into sections for each track.
func splitCaptionTracks(s string) []string {
	var sections []string
	depth := 0
	start := -1

	for i, c := range s {
		switch c {
		case '{':
			if depth == 0 {
				start = i
			}
			depth++
		case '}':
			depth--
			if depth == 0 && start >= 0 {
				sections = append(sections, s[start:i+1])
				start = -1
			}
		}
	}

	return sections
}

// unescapeJSON unescapes special characters in a JSON string.
func unescapeJSON(s string) string {
	replacer := strings.NewReplacer(
		`\u0026`, "&",
		`\\u0026`, "&",
		`\/`, "/",
		`\\/`, "/",
		`\"`, "\"",
		`\\n`, "\n",
	)
	return replacer.Replace(s)
}

// selectTrack selects the best matching caption track.
// Priority: requested language → manual subtitles → auto-generated → first available.
func selectTrack(tracks []captionTrack, lang string) captionTrack {
	lang = strings.TrimSpace(strings.ToLower(lang))

	// Nếu có chỉ định ngôn ngữ, tìm exact match
	if lang != "" {
		// Ưu tiên manual trước
		for _, t := range tracks {
			if strings.ToLower(t.LangCode) == lang && t.Kind != "asr" {
				return t
			}
		}
		// Rồi tới ASR
		for _, t := range tracks {
			if strings.ToLower(t.LangCode) == lang {
				return t
			}
		}
	}

	// Không chỉ định hoặc không tìm thấy → ưu tiên manual tracks
	for _, t := range tracks {
		if t.Kind != "asr" {
			return t
		}
	}

	// Cuối cùng trả về track đầu tiên (ASR)
	return tracks[0]
}

// fetchTranscriptXML fetches and parses transcript XML from YouTube.
func fetchTranscriptXML(baseURL string) (*transcriptList, error) {
	req, err := http.NewRequest(http.MethodGet, baseURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("transcript endpoint returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return nil, err
	}

	var transcript transcriptList
	if err := xml.Unmarshal(body, &transcript); err != nil {
		return nil, fmt.Errorf("cannot parse transcript XML: %w", err)
	}

	return &transcript, nil
}

// formatTranscript formats the transcript result into structured text.
func formatTranscript(videoID string, track captionTrack, transcript *transcriptList) string {
	var sb strings.Builder

	kind := "Manual"
	if track.Kind == "asr" {
		kind = "Auto-generated (ASR)"
	}

	sb.WriteString("YOUTUBE VIDEO TRANSCRIPT\n")
	sb.WriteString(strings.Repeat("─", 60) + "\n")
	sb.WriteString(fmt.Sprintf("Video ID   : %s\n", videoID))
	sb.WriteString(fmt.Sprintf("Language   : %s (%s)\n", track.Name, track.LangCode))
	sb.WriteString(fmt.Sprintf("Type       : %s\n", kind))
	sb.WriteString(fmt.Sprintf("Segments   : %d\n", len(transcript.Texts)))
	sb.WriteString(strings.Repeat("─", 60) + "\n\n")

	for _, text := range transcript.Texts {
		// Decode HTML entities in transcript text
		content := decodeHTMLEntities(text.Text)
		// Replace newlines in transcript
		content = strings.ReplaceAll(content, "\n", " ")
		content = strings.TrimSpace(content)

		if content == "" {
			continue
		}

		// Format timestamp
		timestamp := formatTimestamp(text.Start)
		sb.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, content))
	}

	return sb.String()
}

// formatTimestamp converts seconds (as a float string) to MM:SS or HH:MM:SS.
func formatTimestamp(startSeconds string) string {
	// Parse float seconds
	var totalSeconds float64
	fmt.Sscanf(startSeconds, "%f", &totalSeconds)

	totalSec := int(totalSeconds)
	hours := totalSec / 3600
	minutes := (totalSec % 3600) / 60
	seconds := totalSec % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}
