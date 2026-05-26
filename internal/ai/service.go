package ai

import (
	"context"
	"encoding/json"
	"fmt"
	stdhtml "html"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/tphakala/birdnet-go/internal/ai/llm"
	"github.com/tphakala/birdnet-go/internal/classifier"
	"github.com/tphakala/birdnet-go/internal/conf"
	"github.com/tphakala/birdnet-go/internal/datastore/v2/entities"
	"github.com/tphakala/birdnet-go/internal/datastore/v2/repository"
	"github.com/tphakala/birdnet-go/internal/ebird"
	"github.com/tphakala/birdnet-go/internal/logger"
	"github.com/tphakala/birdnet-go/internal/secrets"
	xhtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	defaultCacheHours       = 4
	defaultReportDays       = 1
	llmRequestTimeout       = 30 * time.Second
	maxTopSpeciesRows       = 12
	maxNotableSpeciesRows   = 8
	highConfidenceThreshold = 0.9

	cacheFileName = "daily_report.json"
)

var (
	imageURLPattern = regexp.MustCompile(`https?://[^\s)\]>"]+\.(?:png|jpg|jpeg|gif|webp|svg)`) //nolint:gochecknoglobals
	anyURLPattern   = regexp.MustCompile(`https?://[^\s)\]>"]+`)                                //nolint:gochecknoglobals
	imgTagPattern   = regexp.MustCompile(`(?is)<img\b[^>]*>`)                                   //nolint:gochecknoglobals
	allowedHTMLTags = map[string]struct{}{                                                      //nolint:gochecknoglobals
		"h3":     {},
		"p":      {},
		"ul":     {},
		"ol":     {},
		"li":     {},
		"table":  {},
		"thead":  {},
		"tbody":  {},
		"tr":     {},
		"th":     {},
		"td":     {},
		"strong": {},
		"em":     {},
		"br":     {},
	}
)

type ReportService struct {
	settings    *conf.Settings
	detection   repository.DetectionRepository
	weather     repository.WeatherRepository
	labelRepo   repository.LabelRepository
	ebirdClient *ebird.Client
	cacheDir    string
	log         logger.Logger
}

func (s *ReportService) settingsSnapshot() *conf.Settings {
	return conf.CurrentOrFallback(s.settings)
}

type ReportPayload struct {
	Report      string `json:"report"`
	GeneratedAt string `json:"generatedAt"`
	Cached      bool   `json:"cached"`
}

type reportCacheFile struct {
	Report            string `json:"report"`
	GeneratedAt       int64  `json:"generatedAt"`
	Provider          string `json:"provider"`
	Model             string `json:"model"`
	BaseURL           string `json:"baseUrl"`
	SystemPrompt      string `json:"systemPrompt"`
	ReportDays        int    `json:"reportDays"`
	CacheHours        int    `json:"cacheHours"`
	AIEnabled         bool   `json:"aiEnabled"`
	WeatherEnabled    bool   `json:"weatherEnabled"`
	EBirdEnabled      bool   `json:"ebirdEnabled"`
	GeneratedUnixHour int64  `json:"generatedUnixHour"`
}

type reportStats struct {
	TotalDetections      int
	UniqueSpecies        int
	HighConfidenceCount  int
	HighConfidencePct    float64
	Confidence95To100    int
	Confidence90To94     int
	Confidence80To89     int
	ConfidenceBelow80    int
	PeakHour             int
	PeakHourCount        int
	QuietHour            int
	QuietHourCount       int
	TopSpecies           []speciesRow
	NotableSpecies       []speciesRow
	Weather              weatherSummary
	EBirdContextIncluded bool
	EBirdRecentBySpecies map[string]bool
}

type speciesRow struct {
	CommonName       string
	ScientificName   string
	Detections       int
	AvgConfidencePct float64
	PeakHour         int
	FirstSeenUnix    int64
	LastSeenUnix     int64
	EBirdURL         string
	WikipediaURL     string
	AllAboutBirdsURL string
}

type weatherSummary struct {
	Available bool
	TempMin   float64
	TempMax   float64
	TempAvg   float64
	Pressure  float64
	WindAvg   float64
	WindMax   float64
	Humidity  float64
	Condition string
}

func NewReportService(
	settings *conf.Settings,
	detectionRepo repository.DetectionRepository,
	weatherRepo repository.WeatherRepository,
	labelRepo repository.LabelRepository,
	ebirdClient *ebird.Client,
) *ReportService {
	l := logger.Global().Module("ai")
	cacheDir := filepath.Join("data", "cache", "ai")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		l.Error("failed to create AI cache directory", logger.Error(err))
	}

	return &ReportService{
		settings:    settings,
		detection:   detectionRepo,
		weather:     weatherRepo,
		labelRepo:   labelRepo,
		ebirdClient: ebirdClient,
		cacheDir:    cacheDir,
		log:         l,
	}
}

func (s *ReportService) GetDailyReport(ctx context.Context, bypassCache bool) (*ReportPayload, error) {
	settings := s.settingsSnapshot()
	if !settings.AI.Enabled {
		return nil, fmt.Errorf("AI analysis is disabled in settings")
	}
	apiKey, err := s.resolveProviderAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve AI provider API key: %w", err)
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("AI provider API key is not configured")
	}

	if !bypassCache {
		if cached, ok := s.loadValidCache(); ok {
			return &ReportPayload{
				Report:      cached.Report,
				GeneratedAt: time.Unix(cached.GeneratedAt, 0).Format(time.RFC3339),
				Cached:      true,
			}, nil
		}
	}

	report, generatedAt, shouldCache, err := s.generateReport(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	if !bypassCache && shouldCache {
		s.saveCache(report, generatedAt)
	}

	return &ReportPayload{
		Report:      report,
		GeneratedAt: generatedAt.Format(time.RFC3339),
		Cached:      false,
	}, nil
}

func (s *ReportService) loadValidCache() (*reportCacheFile, bool) {
	settings := s.settingsSnapshot()
	cachePath := filepath.Join(s.cacheDir, cacheFileName)
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, false
	}

	var c reportCacheFile
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, false
	}

	cacheHours := settings.AI.CacheHours
	if cacheHours < 1 {
		cacheHours = defaultCacheHours
	}

	if c.Provider != effectiveProvider(settings.AI.Provider) ||
		c.Model != effectiveModel(settings.AI) ||
		c.BaseURL != effectiveBaseURL(settings.AI) ||
		c.SystemPrompt != settings.AI.SystemPrompt ||
		c.ReportDays != effectiveReportDays(settings.AI.ReportDays) ||
		c.CacheHours != cacheHours ||
		c.AIEnabled != settings.AI.Enabled ||
		c.WeatherEnabled != (settings.Realtime.Weather.Provider != "none") ||
		c.EBirdEnabled != settings.Realtime.EBird.Enabled {
		return nil, false
	}

	if time.Since(time.Unix(c.GeneratedAt, 0)) > time.Duration(cacheHours)*time.Hour {
		return nil, false
	}

	return &c, true
}

func (s *ReportService) saveCache(report string, generatedAt time.Time) {
	settings := s.settingsSnapshot()
	cacheHours := settings.AI.CacheHours
	if cacheHours < 1 {
		cacheHours = defaultCacheHours
	}

	payload := reportCacheFile{
		Report:            report,
		GeneratedAt:       generatedAt.Unix(),
		Provider:          effectiveProvider(settings.AI.Provider),
		Model:             effectiveModel(settings.AI),
		BaseURL:           effectiveBaseURL(settings.AI),
		SystemPrompt:      settings.AI.SystemPrompt,
		ReportDays:        effectiveReportDays(settings.AI.ReportDays),
		CacheHours:        cacheHours,
		AIEnabled:         settings.AI.Enabled,
		WeatherEnabled:    settings.Realtime.Weather.Provider != "none",
		EBirdEnabled:      settings.Realtime.EBird.Enabled,
		GeneratedUnixHour: generatedAt.Unix() / 3600,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		s.log.Warn("failed to marshal AI report cache", logger.Error(err))
		return
	}

	cachePath := filepath.Join(s.cacheDir, cacheFileName)
	if err := os.WriteFile(cachePath, data, 0o644); err != nil {
		s.log.Warn("failed to write AI report cache", logger.Error(err))
	}
}

func (s *ReportService) generateReport(ctx context.Context, apiKey string) (string, time.Time, bool, error) {
	settings := s.settingsSnapshot()
	now := time.Now()
	lookback := time.Duration(effectiveReportDays(settings.AI.ReportDays)) * 24 * time.Hour
	start := now.Add(-lookback).Unix()
	end := now.Unix()

	dets, _, err := s.detection.GetByDateRange(ctx, start, end, 10000, 0)
	if err != nil {
		return "", time.Time{}, false, fmt.Errorf("failed to load detections: %w", err)
	}

	if len(dets) == 0 {
		return fmt.Sprintf("<h1>AI Analysis</h1>\n\n<p>No detections were found in the last %d day(s).</p>", effectiveReportDays(settings.AI.ReportDays)), now, true, nil
	}

	stats, err := s.computeStats(ctx, dets)
	if err != nil {
		return "", time.Time{}, false, err
	}

	narrative, providerID, model, narrativeErr := s.generateNarrative(ctx, stats, apiKey)
	shouldCache := true
	if narrativeErr != nil {
		s.log.Warn("AI narrative generation failed, using fallback", logger.String("provider", providerID), logger.String("model", model), logger.Error(narrativeErr))
		narrative = "## Narrative unavailable\n\nAI narrative generation is currently unavailable. Backend-generated report tables are still shown below."
		shouldCache = false
	}
	narrative = sanitizeNarrative(narrative)

	sections := []string{
		"<h1>AI Analysis</h1>",
		narrative,
		s.renderDailyTotals(stats),
		s.renderTopSpecies(stats),
		s.renderNotableSpecies(stats),
		s.renderWeatherSummary(stats),
		s.renderConfidenceBreakdown(stats),
	}

	return strings.Join(sections, "\n\n"), now, shouldCache, nil
}

func (s *ReportService) computeStats(ctx context.Context, dets []*entities.Detection) (*reportStats, error) {
	stats := &reportStats{}
	hourly := make([]int, 24)
	byScientific := make(map[string][]*entities.Detection)
	labelIDs := make(map[uint]struct{})
	labelsByID := make(map[uint]string)

	for _, d := range dets {
		stats.TotalDetections++
		labelIDs[d.LabelID] = struct{}{}
		if d.Confidence >= highConfidenceThreshold {
			stats.HighConfidenceCount++
		}

		pct := d.Confidence * 100
		switch {
		case pct >= 95:
			stats.Confidence95To100++
		case pct >= 90:
			stats.Confidence90To94++
		case pct >= 80:
			stats.Confidence80To89++
		default:
			stats.ConfidenceBelow80++
		}

		h := time.Unix(d.DetectedAt, 0).Hour()
		hourly[h]++

		sci := fmt.Sprintf("label-%d", d.LabelID)

		if label, err := s.labelRepo.GetByID(ctx, d.LabelID); err == nil && label != nil && strings.TrimSpace(label.ScientificName) != "" {
			sci = label.ScientificName
			labelsByID[d.LabelID] = sci
		}
		byScientific[sci] = append(byScientific[sci], d)
	}

	stats.UniqueSpecies = len(labelIDs)
	if stats.TotalDetections > 0 {
		stats.HighConfidencePct = float64(stats.HighConfidenceCount) / float64(stats.TotalDetections) * 100
	}

	stats.PeakHour, stats.PeakHourCount = peakHour(hourly)
	stats.QuietHour, stats.QuietHourCount = quietHour(hourly)

	var taxonomyCodeMap map[string]string

	if _, scientificIndex, err := classifier.LoadTaxonomyData(""); err == nil {
		taxonomyCodeMap = make(map[string]string, len(scientificIndex))
		for sciName, code := range scientificIndex {
			taxonomyCodeMap[strings.ToLower(sciName)] = code
		}
	} else {
		taxonomyCodeMap = make(map[string]string)
	}

	if s.ebirdClient != nil {
		if taxonomy, err := s.ebirdClient.GetTaxonomy(ctx, ""); err == nil {
			for _, entry := range taxonomy {
				taxonomyCodeMap[strings.ToLower(entry.ScientificName)] = entry.SpeciesCode
			}
		}
	}

	settings := s.settingsSnapshot()
	stats.TopSpecies = buildSpeciesRows(byScientific, settings.BirdNET.Labels, taxonomyCodeMap, settings.AI.UTMParameters, maxTopSpeciesRows)
	stats.NotableSpecies = buildNotableSpeciesRows(byScientific, settings.BirdNET.Labels, taxonomyCodeMap, settings.AI.UTMParameters, maxNotableSpeciesRows)
	stats.Weather = s.fetchWeatherSummary(ctx)
	stats.EBirdContextIncluded = settings.Realtime.EBird.Enabled && s.ebirdClient != nil

	if stats.EBirdContextIncluded {
		recentMap, err := s.fetchEBirdRecentMap(ctx)
		if err == nil {
			stats.EBirdRecentBySpecies = recentMap
			stats.TopSpecies = applyEBirdContext(stats.TopSpecies, recentMap)
			stats.NotableSpecies = applyEBirdContext(stats.NotableSpecies, recentMap)
		}
	}

	return stats, nil
}

func (s *ReportService) fetchEBirdRecentMap(ctx context.Context) (map[string]bool, error) {
	settings := s.settingsSnapshot()
	if s.ebirdClient == nil || !settings.Realtime.EBird.Enabled {
		return map[string]bool{}, nil
	}
	lat := settings.BirdNET.Latitude
	lng := settings.BirdNET.Longitude
	if lat == 0 && lng == 0 {
		return map[string]bool{}, nil
	}
	obs, err := s.ebirdClient.GetRecentObservations(ctx, lat, lng, 14)
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(obs))
	for _, o := range obs {
		name := strings.TrimSpace(strings.ToLower(o.ScientificName))
		if name == "" {
			continue
		}
		out[name] = true
	}
	return out, nil
}

func applyEBirdContext(rows []speciesRow, recentMap map[string]bool) []speciesRow {
	if len(rows) == 0 || len(recentMap) == 0 {
		return rows
	}
	updated := make([]speciesRow, 0, len(rows))
	for _, row := range rows {
		normalized := strings.TrimSpace(strings.ToLower(row.ScientificName))
		if !recentMap[normalized] {
			updated = append(updated, row)
			continue
		}
		updated = append(updated, row)
	}
	return updated
}

func (s *ReportService) fetchWeatherSummary(ctx context.Context) weatherSummary {
	if s.weather == nil {
		return weatherSummary{Available: false}
	}

	date := time.Now().Format(time.DateOnly)
	items, err := s.weather.GetHourlyWeather(ctx, date)
	if err != nil || len(items) == 0 {
		return weatherSummary{Available: false}
	}

	ws := weatherSummary{Available: true, TempMin: items[0].TempMin, TempMax: items[0].TempMax}
	var tempSum, windSum, humiditySum, pressureSum float64
	for i, w := range items {
		tempSum += w.Temperature
		windSum += w.WindSpeed
		humiditySum += float64(w.Humidity)
		pressureSum += float64(w.Pressure)
		if i == 0 || w.TempMin < ws.TempMin {
			ws.TempMin = w.TempMin
		}
		if i == 0 || w.TempMax > ws.TempMax {
			ws.TempMax = w.TempMax
		}
		if w.WindGust > ws.WindMax {
			ws.WindMax = w.WindGust
		}
		if ws.Condition == "" && w.WeatherDesc != "" {
			ws.Condition = w.WeatherDesc
		}
	}
	ws.TempAvg = tempSum / float64(len(items))
	ws.Pressure = pressureSum / float64(len(items))
	ws.WindAvg = windSum / float64(len(items))
	ws.Humidity = humiditySum / float64(len(items))

	return ws
}

func (s *ReportService) generateNarrative(ctx context.Context, stats *reportStats, apiKey string) (string, string, string, error) {
	settings := s.settingsSnapshot()
	requestCtx, cancel := context.WithTimeout(ctx, llmRequestTimeout)
	defer cancel()
	providerID := effectiveProvider(settings.AI.Provider)
	model := effectiveModel(settings.AI)
	provider, err := llm.NewProvider(settings.AI, apiKey, s.log)
	if err != nil {
		return "", providerID, model, fmt.Errorf("failed to create AI provider %q: %w", providerID, err)
	}

	facts := map[string]any{
		"totalDetections":      stats.TotalDetections,
		"uniqueSpecies":        stats.UniqueSpecies,
		"highConfidencePct":    fmt.Sprintf("%.1f", stats.HighConfidencePct),
		"reportWindowDays":     effectiveReportDays(settings.AI.ReportDays),
		"peakHour":             stats.PeakHour,
		"peakHourCount":        stats.PeakHourCount,
		"quietHour":            stats.QuietHour,
		"quietHourCount":       stats.QuietHourCount,
		"weatherAvailable":     stats.Weather.Available,
		"ebirdContextIncluded": stats.EBirdContextIncluded,
	}
	if stats.Weather.Available {
		facts["weather"] = map[string]any{
			"tempMinC":      fmt.Sprintf("%.1f", stats.Weather.TempMin),
			"tempMaxC":      fmt.Sprintf("%.1f", stats.Weather.TempMax),
			"tempAvgC":      fmt.Sprintf("%.1f", stats.Weather.TempAvg),
			"pressureHpa":   fmt.Sprintf("%.1f", stats.Weather.Pressure),
			"humidityPct":   fmt.Sprintf("%.1f", stats.Weather.Humidity),
			"windAvgMS":     fmt.Sprintf("%.1f", stats.Weather.WindAvg),
			"windGustMaxMS": fmt.Sprintf("%.1f", stats.Weather.WindMax),
			"condition":     orUnknown(stats.Weather.Condition),
		}
	}

	factsJSON, _ := json.MarshalIndent(facts, "", "  ")

	// Safety note appended after facts — does not override output format instructions.
	const factsSafetyNote = "Do not invent metrics; only use the provided Facts and say \"unavailable\" when data is missing. Do not include <img> tags, image URLs, or arbitrary external links."

	var prompt string
	if strings.TrimSpace(settings.AI.SystemPrompt) != "" {
		// User has a custom system prompt — honour it exactly. Only append the
		// facts safety note so we never contradict the user's own formatting rules.
		prompt = settings.AI.SystemPrompt + "\n\n" + factsSafetyNote + "\n\nFacts:\n" + string(factsJSON)
	} else {
		// No custom prompt configured — use a safe neutral default.
		prompt = "You are generating narrative commentary for a BirdNET-Go daily report. " +
			"Write a clear, detailed summary in well-structured HTML using <h3>, <p>, <ul>, <li>, and <table> tags. " +
			"Do not include <html> or <body> tags. " +
			factsSafetyNote + "\n\nFacts:\n" + string(factsJSON)
	}

	s.log.Info("AI prompt payload",
		logger.String("provider", providerID),
		logger.String("model", model),
		logger.Int("prompt_length", len(prompt)),
		logger.String("prompt", prompt),
	)
	result, err := provider.Generate(requestCtx, llm.GenerateRequest{SystemPrompt: "", Prompt: prompt, Model: model})
	if err != nil {
		return "", providerID, model, err
	}

	return strings.TrimSpace(result.Text), providerID, model, nil
}

func (s *ReportService) resolveProviderAPIKey() (string, error) {
	settings := s.settingsSnapshot()
	apiKey, source, err := secrets.ResolveWithSource(settings.AI.APIKeyFile, settings.AI.APIKey)
	if err != nil {
		return "", err
	}
	if source == secrets.SecretSourceEnvOrText && !secrets.IsEnvReference(settings.AI.APIKey) && strings.TrimSpace(settings.AI.APIKey) != "" {
		s.log.Warn("plaintext secret in use; migrate to env var or secret file",
			logger.String("field", "ai.apiKey"),
			logger.String("source", "plaintext"),
		)
	}
	return apiKey, nil
}

func sanitizeNarrative(input string) string {
	// Remove any markdown code block wrappers the LLM might have added
	out := strings.ReplaceAll(input, "```html", "")
	out = strings.ReplaceAll(out, "```", "")

	out = imgTagPattern.ReplaceAllString(out, "")
	out = imageURLPattern.ReplaceAllString(out, "[image-url-removed]")
	out = anyURLPattern.ReplaceAllStringFunc(out, func(v string) string {
		lower := strings.ToLower(v)
		if strings.Contains(lower, "/api/v2/media/") {
			return v
		}
		return "[url-removed]"
	})
	out = sanitizeNarrativeHTML(out)
	return strings.TrimSpace(out)
}

// sanitizeNarrativeHTML keeps only a small allowlist of safe structural tags
// and strips all attributes from allowed tags.
func sanitizeNarrativeHTML(input string) string {
	nodes, err := xhtml.ParseFragment(strings.NewReader(input), &xhtml.Node{Type: xhtml.ElementNode, Data: "div", DataAtom: atom.Div})
	if err != nil {
		return regexp.MustCompile(`(?is)<[^>]+>`).ReplaceAllString(input, "")
	}

	var b strings.Builder
	for _, n := range nodes {
		renderSanitizedNode(&b, n)
	}
	return b.String()
}

func renderSanitizedNode(b *strings.Builder, n *xhtml.Node) {
	switch n.Type {
	case xhtml.TextNode:
		b.WriteString(n.Data)
	case xhtml.ElementNode:
		tag := strings.ToLower(n.Data)
		_, allowed := allowedHTMLTags[tag]
		if allowed {
			b.WriteString("<")
			b.WriteString(tag)
			b.WriteString(">")
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			renderSanitizedNode(b, child)
		}
		if allowed && tag != "br" {
			b.WriteString("</")
			b.WriteString(tag)
			b.WriteString(">")
		}
	case xhtml.ErrorNode, xhtml.DocumentNode, xhtml.CommentNode, xhtml.DoctypeNode, xhtml.RawNode:
		// Do nothing for other node types
	}
}

func (s *ReportService) renderDailyTotals(stats *reportStats) string {
	return strings.Join([]string{
		"<h2>Daily Totals</h2>",
		`<table><thead><tr><th>Metric</th><th>Value</th></tr></thead><tbody>`,
		fmt.Sprintf("<tr><td>Total detections</td><td>%d</td></tr>", stats.TotalDetections),
		fmt.Sprintf("<tr><td>Unique species</td><td>%d</td></tr>", stats.UniqueSpecies),
		fmt.Sprintf("<tr><td>High-confidence detections</td><td>%d (%.1f%%)</td></tr>", stats.HighConfidenceCount, stats.HighConfidencePct),
		fmt.Sprintf("<tr><td>Peak hour</td><td>%s (%d detections)</td></tr>", hourWindow(stats.PeakHour), stats.PeakHourCount),
		fmt.Sprintf("<tr><td>Quietest hour</td><td>%s (%d detections)</td></tr>", hourWindow(stats.QuietHour), stats.QuietHourCount),
		"</tbody></table>",
	}, "\n")
}

func (s *ReportService) renderTopSpecies(stats *reportStats) string {
	rows := []string{"<h2>Top Detections</h2>", `<table><thead><tr><th>Thumbnail</th><th>Common</th><th>Scientific</th><th>Detections</th><th>Avg confidence</th><th>Peak</th><th>Links</th></tr></thead><tbody>`}
	for _, row := range stats.TopSpecies {
		thumb := imageHTML(row.ScientificName, row.CommonName)
		links := speciesLinksHTML(row)
		rows = append(rows, fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td><td>%d</td><td>%.1f%%</td><td>%s</td><td>%s</td></tr>", thumb, esc(row.CommonName), esc(row.ScientificName), row.Detections, row.AvgConfidencePct, hourWindow(row.PeakHour), links))
	}
	rows = append(rows, "</tbody></table>")
	return strings.Join(rows, "\n")
}

func (s *ReportService) renderNotableSpecies(stats *reportStats) string {
	rows := []string{"<h2>Rare / Notable Detections</h2>", `<table><thead><tr><th>Thumbnail</th><th>Species</th><th>Detections</th><th>Avg confidence</th><th>First seen</th><th>Last seen</th><th>Links</th></tr></thead><tbody>`}
	for _, row := range stats.NotableSpecies {
		links := speciesLinksHTML(row)
		rows = append(rows, fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%d</td><td>%.1f%%</td><td>%s</td><td>%s</td><td>%s</td></tr>", imageHTML(row.ScientificName, row.CommonName), esc(row.CommonName), row.Detections, row.AvgConfidencePct, time.Unix(row.FirstSeenUnix, 0).Format("01-02-2006 15:04:05"), time.Unix(row.LastSeenUnix, 0).Format("01-02-2006 15:04:05"), links))
	}
	rows = append(rows, "</tbody></table>")
	return strings.Join(rows, "\n")
}

func (s *ReportService) renderWeatherSummary(stats *reportStats) string {
	if !stats.Weather.Available {
		return "<h2>Weather Summary</h2>\n\n<p>Weather data unavailable for this report window.</p>"
	}
	return strings.Join([]string{
		"<h2>Weather Summary</h2>",
		`<table><thead><tr><th>Metric</th><th>Value</th></tr></thead><tbody>`,
		fmt.Sprintf("<tr><td>Temperature</td><td>min %.1f°C / max %.1f°C / avg %.1f°C</td></tr>", stats.Weather.TempMin, stats.Weather.TempMax, stats.Weather.TempAvg),
		fmt.Sprintf("<tr><td>Wind</td><td>avg %.1f m/s / max %.1f m/s</td></tr>", stats.Weather.WindAvg, stats.Weather.WindMax),
		fmt.Sprintf("<tr><td>Humidity</td><td>%.1f%%</td></tr>", stats.Weather.Humidity),
		fmt.Sprintf("<tr><td>Conditions</td><td>%s</td></tr>", esc(orUnknown(stats.Weather.Condition))),
		"</tbody></table>",
	}, "\n")
}

func (s *ReportService) renderConfidenceBreakdown(stats *reportStats) string {
	return strings.Join([]string{
		"<h2>Confidence Breakdown</h2>",
		`<table><thead><tr><th>Bucket</th><th>Count</th></tr></thead><tbody>`,
		fmt.Sprintf("<tr><td>95–100%%</td><td>%d</td></tr>", stats.Confidence95To100),
		fmt.Sprintf("<tr><td>90–94%%</td><td>%d</td></tr>", stats.Confidence90To94),
		fmt.Sprintf("<tr><td>80–89%%</td><td>%d</td></tr>", stats.Confidence80To89),
		fmt.Sprintf("<tr><td>&lt;80%%</td><td>%d</td></tr>", stats.ConfidenceBelow80),
		"</tbody></table>",
	}, "\n")
}

func getSpeciesInfo(labels []string, scientificName string) (commonName, ebirdCode string) {
	for _, l := range labels {
		sci, rest, found := strings.Cut(l, "_")
		if found && sci == scientificName {
			if idx := strings.LastIndex(rest, "_"); idx > 0 {
				return rest[:idx], rest[idx+1:]
			}
			return rest, ""
		}
	}
	return scientificName, ""
}

func buildSpeciesRows(grouped map[string][]*entities.Detection, labels []string, taxonomyCodeMap map[string]string, utmParams string, limit int) []speciesRow {
	rows := make([]speciesRow, 0, len(grouped))
	for sci, items := range grouped {
		if len(items) == 0 {
			continue
		}
		var confSum float64
		first := items[0].DetectedAt
		last := items[0].DetectedAt
		hourCounts := make([]int, 24)
		for _, d := range items {
			confSum += d.Confidence
			if d.DetectedAt < first {
				first = d.DetectedAt
			}
			if d.DetectedAt > last {
				last = d.DetectedAt
			}
			hourCounts[time.Unix(d.DetectedAt, 0).Hour()]++
		}
		pHour, _ := peakHour(hourCounts)
		commonName, ebirdCode := getSpeciesInfo(labels, sci)
		if ebirdCode == "" && taxonomyCodeMap != nil {
			ebirdCode = taxonomyCodeMap[strings.ToLower(sci)]
		}
		ebirdURL := ""
		if ebirdCode != "" {
			ebirdURL = fmt.Sprintf("https://ebird.org/species/%s", url.QueryEscape(ebirdCode))
			ebirdURL = appendUTMParameters(ebirdURL, utmParams)
		}

		wikiURL := fmt.Sprintf("https://wikipedia.org/wiki/%s", url.PathEscape(strings.ReplaceAll(sci, " ", "_")))
		wikiURL = appendUTMParameters(wikiURL, utmParams)
		aabURL := ""
		if commonName != "" && commonName != sci {
			cleanedCommon := strings.ReplaceAll(commonName, "'", "")
			aabURL = fmt.Sprintf("https://allaboutbirds.org/guide/%s", url.PathEscape(strings.ReplaceAll(cleanedCommon, " ", "_")))
			aabURL = appendUTMParameters(aabURL, utmParams)
		}

		rows = append(rows, speciesRow{
			CommonName:       commonName,
			ScientificName:   sci,
			Detections:       len(items),
			AvgConfidencePct: confSum / float64(len(items)) * 100,
			PeakHour:         pHour,
			FirstSeenUnix:    first,
			LastSeenUnix:     last,
			EBirdURL:         ebirdURL,
			WikipediaURL:     wikiURL,
			AllAboutBirdsURL: aabURL,
		})
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].Detections > rows[j].Detections })
	if len(rows) > limit {
		rows = rows[:limit]
	}
	return rows
}

func peakHour(hourly []int) (int, int) {
	peakH, peakC := 0, -1
	for h, c := range hourly {
		if c > peakC {
			peakH, peakC = h, c
		}
	}
	if peakC < 0 {
		return 0, 0
	}
	return peakH, peakC
}

func quietHour(hourly []int) (int, int) {
	quietH, quietC := 0, int(^uint(0)>>1)
	for h, c := range hourly {
		if c < quietC {
			quietH, quietC = h, c
		}
	}
	if quietC == int(^uint(0)>>1) {
		return 0, 0
	}
	return quietH, quietC
}

func hourWindow(hour int) string {
	end := (hour + 1) % 24
	return fmt.Sprintf("%02d:00–%02d:00", hour, end)
}

func imageHTML(scientificName, commonName string) string {
	escSci := url.QueryEscape(scientificName)
	escCom := url.QueryEscape(commonName)
	alt := esc(orUnknown(commonName))
	return fmt.Sprintf(`<img src="/api/v2/media/species-image?name=%s&common=%s" alt="%s" loading="lazy" width="64" height="64" style="border-radius:4px;">`, escSci, escCom, alt)
}

func ebirdLinkHTML(link string) string {
	if strings.TrimSpace(link) == "" {
		return "—"
	}
	return fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener noreferrer" title="Open eBird species page">eBird</a>`, esc(link))
}

func speciesLinksHTML(row speciesRow) string {
	var links []string
	if strings.TrimSpace(row.EBirdURL) != "" {
		links = append(links, fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener noreferrer" title="Open eBird species page">eBird</a>`, esc(row.EBirdURL)))
	}
	if strings.TrimSpace(row.WikipediaURL) != "" {
		links = append(links, fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener noreferrer" title="Open Wikipedia page">Wikipedia</a>`, esc(row.WikipediaURL)))
	}
	if strings.TrimSpace(row.AllAboutBirdsURL) != "" {
		links = append(links, fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener noreferrer" title="Open All About Birds page">All About Birds</a>`, esc(row.AllAboutBirdsURL)))
	}
	if len(links) == 0 {
		return "—"
	}
	return strings.Join(links, "<br>")
}

func appendUTMParameters(rawURL, utm string) string {
	utm = strings.TrimSpace(utm)
	if utm == "" {
		return rawURL
	}
	utm = strings.TrimLeft(utm, "?&")
	if strings.Contains(rawURL, "?") {
		return rawURL + "&" + utm
	}
	return rawURL + "?" + utm
}

func buildNotableSpeciesRows(grouped map[string][]*entities.Detection, labels []string, taxonomyCodeMap map[string]string, utmParams string, limit int) []speciesRow {
	rows := buildSpeciesRows(grouped, labels, taxonomyCodeMap, utmParams, 1000)
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Detections == rows[j].Detections {
			return rows[i].AvgConfidencePct > rows[j].AvgConfidencePct
		}
		return rows[i].Detections < rows[j].Detections
	})
	if len(rows) > limit {
		rows = rows[:limit]
	}
	return rows
}

func esc(v string) string {
	return stdhtml.EscapeString(v)
}

func effectiveProvider(provider string) string {
	provider = strings.TrimSpace(strings.ToLower(provider))
	if provider == "" {
		return llm.ProviderGemini
	}
	return provider
}

func effectiveModel(settings conf.AISettings) string {
	model := strings.TrimSpace(settings.Model)
	if model != "" {
		return model
	}
	switch effectiveProvider(settings.Provider) {
	case llm.ProviderOpenAI:
		return llm.DefaultOpenAIModel
	case llm.ProviderOpenRouter:
		return llm.DefaultOpenRouterModel
	case llm.ProviderOllama:
		return llm.DefaultOllamaModel
	case llm.ProviderAnthropic:
		return llm.DefaultAnthropicModel
	default:
		return llm.DefaultGeminiModel
	}
}

func effectiveBaseURL(settings conf.AISettings) string {
	baseURL := strings.TrimSpace(settings.BaseURL)
	if baseURL != "" {
		return strings.TrimRight(baseURL, "/")
	}
	switch effectiveProvider(settings.Provider) {
	case llm.ProviderOpenAI:
		return llm.DefaultOpenAIBaseURL
	case llm.ProviderOpenRouter:
		return llm.DefaultOpenRouterBaseURL
	case llm.ProviderOllama:
		return llm.DefaultOllamaBaseURL
	default:
		return ""
	}
}

func effectiveReportDays(days int) int {
	if days < 1 {
		return defaultReportDays
	}
	if days > 31 {
		return 31
	}
	return days
}

func orUnknown(v string) string {
	if strings.TrimSpace(v) == "" {
		return "unknown"
	}
	return v
}
