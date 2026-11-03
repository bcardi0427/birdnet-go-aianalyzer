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
	"github.com/tphakala/birdnet-go/internal/errors"
	"github.com/tphakala/birdnet-go/internal/logger"
	"github.com/tphakala/birdnet-go/internal/secrets"
	"github.com/tphakala/birdnet-go/internal/suncalc"
	xhtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	defaultCacheHours       = 4
	defaultReportDays       = 1
	llmRequestTimeout       = 120 * time.Second
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

// ReportService orchestrates the generation, caching, and retrieval of
// AI-powered daily bird activity reports.
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

// ReportPayload represents the returned data structure of a generated AI report,
// including the raw report text, timestamp, and cache status.
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
	TemperatureUnit   string `json:"temperatureUnit"` // bust cache when user switches °C↔°F
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
	HourlyCounts         []int
	AllSpecies           []speciesRow
	TopSpecies           []speciesRow
	NotableSpecies       []speciesRow
	Weather              weatherSummary
	PriorWeather         weatherSummary
	DailyWeather         []dailyWeatherRow
	EBirdContextIncluded bool
	EBirdRecentBySpecies map[string]bool
	CivilDawn            string
	Sunrise              string
	Sunset               string
	CivilDusk            string
	PriorComparison      priorComparisonStats
}

type dailyWeatherRow struct {
	Date      string  `json:"date"`
	TempMin   float64 `json:"tempMin"`
	TempMax   float64 `json:"tempMax"`
	TempAvg   float64 `json:"tempAvg"`
	Pressure  float64 `json:"pressure"`
	WindAvg   float64 `json:"windAvg"`
	WindMax   float64 `json:"windMax"`
	Humidity  float64 `json:"humidity"`
	Condition string  `json:"condition"`
}

type speciesDelta struct {
	CommonName     string `json:"commonName"`
	ScientificName string `json:"scientificName"`
	CurrentCount   int    `json:"currentCount"`
	PriorCount     int    `json:"priorCount"`
	Delta          int    `json:"delta"`
	ChangePct      string `json:"changePct,omitempty"`
}

type priorComparisonStats struct {
	Available          bool
	WindowDays         int
	PriorDetections    int
	DetectionDelta     int
	DetectionChangePct string
	PriorUniqueSpecies int
	SpeciesDelta       int
	PriorCounts        map[string]int
	NewArrivals        []speciesDelta
	DepartedSpecies    []speciesDelta
	TopSurges          []speciesDelta
	TopDeclines        []speciesDelta
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

// NewReportService creates and initializes a new ReportService instance,
// ensuring the cache directories are created.
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

// GetDailyReport retrieves the daily AI bird detection report, either generating
// a new one using the configured provider or returning a cached version if valid.
func (s *ReportService) GetDailyReport(ctx context.Context, bypassCache bool) (*ReportPayload, error) {
	settings := s.settingsSnapshot()
	if !settings.AI.Enabled {
		return nil, errors.Newf("AI analysis is disabled in settings").
			Component("ai").
			Category(errors.CategoryConfiguration).
			Build()
	}
	apiKey, err := s.resolveProviderAPIKey()
	if err != nil {
		return nil, errors.Newf("failed to resolve AI provider API key: %w", err).
			Component("ai").
			Category(errors.CategoryConfiguration).
			Build()
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.Newf("AI provider API key is not configured").
			Component("ai").
			Category(errors.CategoryConfiguration).
			Build()
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
		c.EBirdEnabled != settings.Realtime.EBird.Enabled ||
		c.TemperatureUnit != settings.Realtime.Dashboard.TemperatureUnit {
		return nil, false
	}

	if time.Since(time.Unix(c.GeneratedAt, 0)) > 5*time.Minute {
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
		TemperatureUnit:   settings.Realtime.Dashboard.TemperatureUnit,
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
	s.saveReportHistory(data, generatedAt)
}

func (s *ReportService) saveReportHistory(data []byte, generatedAt time.Time) {
	histDir := filepath.Join(s.cacheDir, "history")
	if err := os.MkdirAll(histDir, 0o755); err != nil {
		s.log.Warn("failed to create AI report history directory", logger.Error(err))
		return
	}
	fileName := fmt.Sprintf("report_%s.json", generatedAt.Format("2006-01-02_15-04-05"))
	filePath := filepath.Join(histDir, fileName)
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		s.log.Warn("failed to write AI report history", logger.Error(err))
		return
	}

	entries, err := os.ReadDir(histDir)
	if err != nil || len(entries) <= 50 {
		return
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	for i := range len(entries) - 50 {
		_ = os.Remove(filepath.Join(histDir, entries[i].Name()))
	}
}

func (s *ReportService) generateReport(ctx context.Context, apiKey string) (string, time.Time, bool, error) {
	settings := s.settingsSnapshot()
	now := time.Now()
	lookback := time.Duration(effectiveReportDays(settings.AI.ReportDays)) * 24 * time.Hour
	startTime := now.Add(-lookback)
	endTime := now
	start := startTime.Unix()
	end := endTime.Unix()

	dets, _, err := s.detection.GetByDateRange(ctx, start, end, 10000, 0)
	if err != nil {
		return "", time.Time{}, false, errors.Newf("failed to load detections: %w", err).
			Component("ai").
			Category(errors.CategoryDatabase).
			Build()
	}

	if len(dets) == 0 {
		return fmt.Sprintf("<h1>AI Analysis</h1>\n\n<p>No detections were found in the last %d day(s).</p>", effectiveReportDays(settings.AI.ReportDays)), now, true, nil
	}

	prevStartTime := now.Add(-2 * lookback)
	prevEndTime := startTime
	prevStart := prevStartTime.Unix()
	prevDets, _, prevErr := s.detection.GetByDateRange(ctx, prevStart, start, 10000, 0)
	if prevErr != nil {
		s.log.Warn("failed to load prior-period detections for comparison", logger.Error(prevErr))
	}

	stats, err := s.computeStats(ctx, dets, prevDets, startTime, endTime, prevStartTime, prevEndTime)
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

	rawSections := []string{
		"<h1>AI Analysis</h1>",
		narrative,
		s.renderDailyTotals(stats),
		s.renderPeriodComparison(stats),
		s.renderTopSpecies(stats),
		s.renderNotableSpecies(stats),
		s.renderWeatherSummary(stats),
		s.renderConfidenceBreakdown(stats),
	}
	var sections []string
	for _, sec := range rawSections {
		if strings.TrimSpace(sec) != "" {
			sections = append(sections, sec)
		}
	}

	return strings.Join(sections, "\n\n"), now, shouldCache, nil
}

func (s *ReportService) computeStats(
	ctx context.Context,
	dets, prevDets []*entities.Detection,
	startTime, endTime, prevStartTime, prevEndTime time.Time,
) (*reportStats, error) {
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
	allSpecies := buildSpeciesRows(byScientific, settings.BirdNET.Labels, taxonomyCodeMap, settings.AI.UTMParameters, 1000)
	stats.AllSpecies = allSpecies
	if len(allSpecies) > maxTopSpeciesRows {
		stats.TopSpecies = make([]speciesRow, maxTopSpeciesRows)
		copy(stats.TopSpecies, allSpecies[:maxTopSpeciesRows])
	} else {
		stats.TopSpecies = make([]speciesRow, len(allSpecies))
		copy(stats.TopSpecies, allSpecies)
	}
	stats.NotableSpecies = buildNotableSpeciesRows(byScientific, settings.BirdNET.Labels, taxonomyCodeMap, settings.AI.UTMParameters, maxNotableSpeciesRows)
	stats.HourlyCounts = hourly
	stats.Weather, stats.DailyWeather = s.fetchWindowWeatherSummary(ctx, startTime, endTime)
	stats.PriorWeather, _ = s.fetchWindowWeatherSummary(ctx, prevStartTime, prevEndTime)
	stats.EBirdContextIncluded = settings.Realtime.EBird.Enabled && s.ebirdClient != nil

	if stats.EBirdContextIncluded {
		recentMap, err := s.fetchEBirdRecentMap(ctx)
		if err == nil {
			stats.EBirdRecentBySpecies = recentMap
			stats.TopSpecies = applyEBirdContext(stats.TopSpecies, recentMap)
			stats.NotableSpecies = applyEBirdContext(stats.NotableSpecies, recentMap)
		}
	}

	if settings.BirdNET.Latitude != 0 || settings.BirdNET.Longitude != 0 {
		sc := suncalc.NewSunCalc(settings.BirdNET.Latitude, settings.BirdNET.Longitude)
		if st, err := sc.GetSunEventTimes(time.Now()); err == nil {
			stats.CivilDawn = st.CivilDawn.Format("15:04")
			stats.Sunrise = st.Sunrise.Format("15:04")
			stats.Sunset = st.Sunset.Format("15:04")
			stats.CivilDusk = st.CivilDusk.Format("15:04")
		}
	}

	stats.PriorComparison = s.computePriorComparison(ctx, byScientific, prevDets, labelsByID, settings.BirdNET.Labels, effectiveReportDays(settings.AI.ReportDays))

	return stats, nil
}

func (s *ReportService) computePriorComparison(
	ctx context.Context,
	currentByScientific map[string][]*entities.Detection,
	prevDets []*entities.Detection,
	labelsByID map[uint]string,
	labels []string,
	windowDays int,
) priorComparisonStats {
	if len(prevDets) == 0 {
		return priorComparisonStats{
			Available:   false,
			WindowDays:  windowDays,
			PriorCounts: make(map[string]int),
		}
	}

	prevByScientific := make(map[string]int)
	for _, d := range prevDets {
		sci, ok := labelsByID[d.LabelID]
		if !ok {
			sci = fmt.Sprintf("label-%d", d.LabelID)
			if label, err := s.labelRepo.GetByID(ctx, d.LabelID); err == nil && label != nil && strings.TrimSpace(label.ScientificName) != "" {
				sci = label.ScientificName
				labelsByID[d.LabelID] = sci
			}
		}
		prevByScientific[sci]++
	}

	currentTotal := 0
	for _, items := range currentByScientific {
		currentTotal += len(items)
	}
	priorTotal := len(prevDets)
	detDelta := currentTotal - priorTotal
	detPct := "N/A"
	if priorTotal > 0 {
		detPct = fmt.Sprintf("%+.1f%%", float64(detDelta)/float64(priorTotal)*100)
	}

	allSciMap := make(map[string]struct{}, len(currentByScientific)+len(prevByScientific))
	for sci := range currentByScientific {
		allSciMap[sci] = struct{}{}
	}
	for sci := range prevByScientific {
		allSciMap[sci] = struct{}{}
	}

	var newArrivals []speciesDelta
	var departedSpecies []speciesDelta
	var topSurges []speciesDelta
	var topDeclines []speciesDelta

	for sci := range allSciMap {
		currentCount := len(currentByScientific[sci])
		priorCount := prevByScientific[sci]
		delta := currentCount - priorCount
		commonName, _ := getSpeciesInfo(labels, sci)

		sd := speciesDelta{
			CommonName:     orUnknown(commonName),
			ScientificName: sci,
			CurrentCount:   currentCount,
			PriorCount:     priorCount,
			Delta:          delta,
		}

		switch {
		case priorCount == 0 && currentCount > 0:
			newArrivals = append(newArrivals, sd)
		case priorCount > 0 && currentCount == 0:
			departedSpecies = append(departedSpecies, sd)
		case priorCount > 0 && currentCount > priorCount:
			sd.ChangePct = fmt.Sprintf("+%.1f%%", float64(delta)/float64(priorCount)*100)
			topSurges = append(topSurges, sd)
		case currentCount > 0 && currentCount < priorCount:
			sd.ChangePct = fmt.Sprintf("%.1f%%", float64(delta)/float64(priorCount)*100)
			topDeclines = append(topDeclines, sd)
		}
	}

	sort.Slice(newArrivals, func(i, j int) bool {
		return newArrivals[i].CurrentCount > newArrivals[j].CurrentCount
	})
	if len(newArrivals) > 15 {
		newArrivals = newArrivals[:15]
	}

	sort.Slice(departedSpecies, func(i, j int) bool {
		return departedSpecies[i].PriorCount > departedSpecies[j].PriorCount
	})
	if len(departedSpecies) > 15 {
		departedSpecies = departedSpecies[:15]
	}

	sort.Slice(topSurges, func(i, j int) bool {
		return topSurges[i].Delta > topSurges[j].Delta
	})
	if len(topSurges) > 10 {
		topSurges = topSurges[:10]
	}

	sort.Slice(topDeclines, func(i, j int) bool {
		return topDeclines[i].Delta < topDeclines[j].Delta
	})
	if len(topDeclines) > 10 {
		topDeclines = topDeclines[:10]
	}

	return priorComparisonStats{
		Available:          true,
		WindowDays:         windowDays,
		PriorDetections:    priorTotal,
		DetectionDelta:     detDelta,
		DetectionChangePct: detPct,
		PriorUniqueSpecies: len(prevByScientific),
		SpeciesDelta:       len(currentByScientific) - len(prevByScientific),
		PriorCounts:        prevByScientific,
		NewArrivals:        newArrivals,
		DepartedSpecies:    departedSpecies,
		TopSurges:          topSurges,
		TopDeclines:        topDeclines,
	}
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

func (s *ReportService) fetchWindowWeatherSummary(ctx context.Context, startTime, endTime time.Time) (weatherSummary, []dailyWeatherRow) {
	if s.weather == nil {
		return weatherSummary{Available: false}, nil
	}

	if startTime.IsZero() || endTime.IsZero() {
		return s.fetchWeatherSummaryForDate(ctx, time.Now().Format(time.DateOnly))
	}

	loc := startTime.Location()
	startDate := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, loc)
	endDate := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 0, 0, 0, 0, loc)

	var allItems []entities.HourlyWeather
	var dailyRows []dailyWeatherRow

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format(time.DateOnly)
		items, err := s.weather.GetHourlyWeather(ctx, dateStr)
		if err != nil || len(items) == 0 {
			continue
		}

		var dayItems []entities.HourlyWeather
		for i := range items {
			item := &items[i]
			if !item.Time.IsZero() {
				itemUnix := item.Time.Unix()
				if itemUnix >= startTime.Unix() && itemUnix <= endTime.Unix() {
					dayItems = append(dayItems, *item)
					allItems = append(allItems, *item)
				}
			} else {
				dayItems = append(dayItems, *item)
				allItems = append(allItems, *item)
			}
		}

		if len(dayItems) > 0 {
			dailyRows = append(dailyRows, computeDailyWeatherRow(dateStr, dayItems))
		}
	}

	if len(allItems) == 0 {
		return weatherSummary{Available: false}, nil
	}

	return computeOverallWeatherSummary(allItems), dailyRows
}

func (s *ReportService) fetchWeatherSummaryForDate(ctx context.Context, dateStr string) (weatherSummary, []dailyWeatherRow) {
	if s.weather == nil {
		return weatherSummary{Available: false}, nil
	}
	items, err := s.weather.GetHourlyWeather(ctx, dateStr)
	if err != nil || len(items) == 0 {
		return weatherSummary{Available: false}, nil
	}
	row := computeDailyWeatherRow(dateStr, items)
	ws := computeOverallWeatherSummary(items)
	return ws, []dailyWeatherRow{row}
}

func computeDailyWeatherRow(date string, items []entities.HourlyWeather) dailyWeatherRow {
	row := dailyWeatherRow{
		Date:    date,
		TempMin: items[0].TempMin,
		TempMax: items[0].TempMax,
	}
	var tempSum, windSum, humiditySum, pressureSum float64
	for i := range items {
		w := &items[i]
		tempSum += w.Temperature
		windSum += w.WindSpeed
		humiditySum += float64(w.Humidity)
		pressureSum += float64(w.Pressure)
		if i == 0 || w.TempMin < row.TempMin {
			row.TempMin = w.TempMin
		}
		if i == 0 || w.TempMax > row.TempMax {
			row.TempMax = w.TempMax
		}
		if w.WindGust > row.WindMax {
			row.WindMax = w.WindGust
		}
		if row.Condition == "" && w.WeatherDesc != "" {
			row.Condition = w.WeatherDesc
		}
	}
	row.TempAvg = tempSum / float64(len(items))
	row.Pressure = pressureSum / float64(len(items))
	row.WindAvg = windSum / float64(len(items))
	row.Humidity = humiditySum / float64(len(items))
	return row
}

func computeOverallWeatherSummary(items []entities.HourlyWeather) weatherSummary {
	row := computeDailyWeatherRow("", items)
	return weatherSummary{
		Available: true,
		TempMin:   row.TempMin,
		TempMax:   row.TempMax,
		TempAvg:   row.TempAvg,
		Pressure:  row.Pressure,
		WindAvg:   row.WindAvg,
		WindMax:   row.WindMax,
		Humidity:  row.Humidity,
		Condition: row.Condition,
	}
}

func (s *ReportService) buildFacts(stats *reportStats) map[string]any {
	settings := s.settingsSnapshot()
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

	facts["confidenceBreakdown"] = map[string]int{
		"95To100Pct": stats.Confidence95To100,
		"90To94Pct":  stats.Confidence90To94,
		"80To89Pct":  stats.Confidence80To89,
		"below80Pct": stats.ConfidenceBelow80,
	}

	if len(stats.HourlyCounts) == 24 {
		hourlyList := make([]map[string]any, 24)
		for h := range 24 {
			hourlyList[h] = map[string]any{
				"hour":       hourWindow(h),
				"detections": stats.HourlyCounts[h],
			}
		}
		facts["hourlyActivity"] = hourlyList
	}

	if stats.Sunrise != "" || stats.Sunset != "" {
		facts["solarTimes"] = map[string]string{
			"civilDawn": stats.CivilDawn,
			"sunrise":   stats.Sunrise,
			"sunset":    stats.Sunset,
			"civilDusk": stats.CivilDusk,
		}
	}

	if settings.BirdNET.Latitude != 0 || settings.BirdNET.Longitude != 0 {
		facts["location"] = map[string]float64{
			"latitude":  settings.BirdNET.Latitude,
			"longitude": settings.BirdNET.Longitude,
		}
	}

	speciesSource := stats.AllSpecies
	if len(speciesSource) == 0 {
		speciesSource = stats.TopSpecies
	}
	limit := len(speciesSource)
	if limit > 50 {
		limit = 50
	}
	speciesList := make([]map[string]any, 0, limit)
	for i := range limit {
		row := speciesSource[i]
		share := 0.0
		if stats.TotalDetections > 0 {
			share = float64(row.Detections) / float64(stats.TotalDetections) * 100
		}
		item := map[string]any{
			"commonName":       orUnknown(row.CommonName),
			"scientificName":   row.ScientificName,
			"detections":       row.Detections,
			"sharePct":         fmt.Sprintf("%.1f%%", share),
			"avgConfidencePct": fmt.Sprintf("%.1f%%", row.AvgConfidencePct),
			"peakHour":         hourWindow(row.PeakHour),
		}
		if row.FirstSeenUnix > 0 {
			item["firstSeen"] = time.Unix(row.FirstSeenUnix, 0).Format(time.DateTime)
		}
		if row.LastSeenUnix > 0 {
			item["lastSeen"] = time.Unix(row.LastSeenUnix, 0).Format(time.DateTime)
		}
		if stats.EBirdContextIncluded && stats.EBirdRecentBySpecies != nil {
			item["recentInArea"] = stats.EBirdRecentBySpecies[strings.ToLower(row.ScientificName)]
		}
		if stats.PriorComparison.Available && stats.PriorComparison.PriorCounts != nil {
			priorCount := stats.PriorComparison.PriorCounts[row.ScientificName]
			item["priorDetections"] = priorCount
			item["delta"] = row.Detections - priorCount
			switch {
			case priorCount == 0:
				item["trend"] = "new_arrival"
			case row.Detections > priorCount:
				item["trend"] = "increased"
			case row.Detections < priorCount:
				item["trend"] = "decreased"
			default:
				item["trend"] = "stable"
			}
		}
		speciesList = append(speciesList, item)
	}
	facts["speciesBreakdown"] = speciesList

	isFahrenheit := settings.Realtime.Dashboard.TemperatureUnit == conf.TemperatureUnitFahrenheit

	if stats.PriorComparison.Available {
		priorCompMap := map[string]any{
			"precedingWindowDays":  stats.PriorComparison.WindowDays,
			"priorTotalDetections": stats.PriorComparison.PriorDetections,
			"detectionDelta":       stats.PriorComparison.DetectionDelta,
			"detectionChangePct":   stats.PriorComparison.DetectionChangePct,
			"priorUniqueSpecies":   stats.PriorComparison.PriorUniqueSpecies,
			"speciesDelta":         stats.PriorComparison.SpeciesDelta,
			"newArrivals":          stats.PriorComparison.NewArrivals,
			"departedSpecies":      stats.PriorComparison.DepartedSpecies,
			"topSurges":            stats.PriorComparison.TopSurges,
			"topDeclines":          stats.PriorComparison.TopDeclines,
		}

		if stats.Weather.Available && stats.PriorWeather.Available {
			if isFahrenheit {
				currTempAvg := stats.Weather.TempAvg*1.8 + 32
				priorTempAvg := stats.PriorWeather.TempAvg*1.8 + 32
				currWindAvg := stats.Weather.WindAvg * 2.23694
				priorWindAvg := stats.PriorWeather.WindAvg * 2.23694

				priorCompMap["weatherComparison"] = map[string]any{
					"priorTempMinF":    fmt.Sprintf("%.1f", stats.PriorWeather.TempMin*1.8+32),
					"priorTempMaxF":    fmt.Sprintf("%.1f", stats.PriorWeather.TempMax*1.8+32),
					"priorTempAvgF":    fmt.Sprintf("%.1f", priorTempAvg),
					"tempAvgDeltaF":    fmt.Sprintf("%+.1f°F", currTempAvg-priorTempAvg),
					"priorWindAvgMph":  fmt.Sprintf("%.1f", priorWindAvg),
					"priorWindGustMph": fmt.Sprintf("%.1f", stats.PriorWeather.WindMax*2.23694),
					"windAvgDeltaMph":  fmt.Sprintf("%+.1f mph", currWindAvg-priorWindAvg),
					"priorPressureHpa": fmt.Sprintf("%.1f", stats.PriorWeather.Pressure),
					"pressureDeltaHpa": fmt.Sprintf("%+.1f hPa", stats.Weather.Pressure-stats.PriorWeather.Pressure),
					"priorHumidityPct": fmt.Sprintf("%.1f", stats.PriorWeather.Humidity),
					"humidityDeltaPct": fmt.Sprintf("%+.1f%%", stats.Weather.Humidity-stats.PriorWeather.Humidity),
					"priorCondition":   orUnknown(stats.PriorWeather.Condition),
				}
			} else {
				priorCompMap["weatherComparison"] = map[string]any{
					"priorTempMinC":    fmt.Sprintf("%.1f", stats.PriorWeather.TempMin),
					"priorTempMaxC":    fmt.Sprintf("%.1f", stats.PriorWeather.TempMax),
					"priorTempAvgC":    fmt.Sprintf("%.1f", stats.PriorWeather.TempAvg),
					"tempAvgDeltaC":    fmt.Sprintf("%+.1f°C", stats.Weather.TempAvg-stats.PriorWeather.TempAvg),
					"priorWindAvgMS":   fmt.Sprintf("%.1f", stats.PriorWeather.WindAvg),
					"priorWindGustMS":  fmt.Sprintf("%.1f", stats.PriorWeather.WindMax),
					"windAvgDeltaMS":   fmt.Sprintf("%+.1f m/s", stats.Weather.WindAvg-stats.PriorWeather.WindAvg),
					"priorPressureHpa": fmt.Sprintf("%.1f", stats.PriorWeather.Pressure),
					"pressureDeltaHpa": fmt.Sprintf("%+.1f hPa", stats.Weather.Pressure-stats.PriorWeather.Pressure),
					"priorHumidityPct": fmt.Sprintf("%.1f", stats.PriorWeather.Humidity),
					"humidityDeltaPct": fmt.Sprintf("%+.1f%%", stats.Weather.Humidity-stats.PriorWeather.Humidity),
					"priorCondition":   orUnknown(stats.PriorWeather.Condition),
				}
			}
		}

		facts["priorPeriodComparison"] = priorCompMap
	}

	if stats.Weather.Available {
		if isFahrenheit {
			facts["weather"] = map[string]any{
				"tempMinF":       fmt.Sprintf("%.1f", stats.Weather.TempMin*1.8+32),
				"tempMaxF":       fmt.Sprintf("%.1f", stats.Weather.TempMax*1.8+32),
				"tempAvgF":       fmt.Sprintf("%.1f", stats.Weather.TempAvg*1.8+32),
				"pressureHpa":    fmt.Sprintf("%.1f", stats.Weather.Pressure),
				"humidityPct":    fmt.Sprintf("%.1f", stats.Weather.Humidity),
				"windAvgMph":     fmt.Sprintf("%.1f", stats.Weather.WindAvg*2.23694),
				"windGustMaxMph": fmt.Sprintf("%.1f", stats.Weather.WindMax*2.23694),
				"condition":      orUnknown(stats.Weather.Condition),
			}
			facts["temperatureUnitPreference"] = "Fahrenheit (°F)"
			facts["windSpeedUnitPreference"] = "miles per hour (mph)"
		} else {
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
			facts["temperatureUnitPreference"] = "Celsius (°C)"
			facts["windSpeedUnitPreference"] = "meters per second (m/s)"
		}

		if len(stats.DailyWeather) > 0 {
			dailyList := make([]map[string]any, 0, len(stats.DailyWeather))
			for _, dw := range stats.DailyWeather {
				if isFahrenheit {
					dailyList = append(dailyList, map[string]any{
						"date":           dw.Date,
						"tempMinF":       fmt.Sprintf("%.1f", dw.TempMin*1.8+32),
						"tempMaxF":       fmt.Sprintf("%.1f", dw.TempMax*1.8+32),
						"tempAvgF":       fmt.Sprintf("%.1f", dw.TempAvg*1.8+32),
						"windAvgMph":     fmt.Sprintf("%.1f", dw.WindAvg*2.23694),
						"windGustMaxMph": fmt.Sprintf("%.1f", dw.WindMax*2.23694),
						"pressureHpa":    fmt.Sprintf("%.1f", dw.Pressure),
						"humidityPct":    fmt.Sprintf("%.1f", dw.Humidity),
						"condition":      orUnknown(dw.Condition),
					})
				} else {
					dailyList = append(dailyList, map[string]any{
						"date":          dw.Date,
						"tempMinC":      fmt.Sprintf("%.1f", dw.TempMin),
						"tempMaxC":      fmt.Sprintf("%.1f", dw.TempMax),
						"tempAvgC":      fmt.Sprintf("%.1f", dw.TempAvg),
						"windAvgMS":     fmt.Sprintf("%.1f", dw.WindAvg),
						"windGustMaxMS": fmt.Sprintf("%.1f", dw.WindMax),
						"pressureHpa":   fmt.Sprintf("%.1f", dw.Pressure),
						"humidityPct":   fmt.Sprintf("%.1f", dw.Humidity),
						"condition":     orUnknown(dw.Condition),
					})
				}
			}
			facts["dailyWeather"] = dailyList
		}
	}

	return facts
}

func (s *ReportService) buildPrompt(stats *reportStats) string {
	settings := s.settingsSnapshot()
	facts := s.buildFacts(stats)
	factsJSON, _ := json.MarshalIndent(facts, "", "  ")

	var comparisonInstruction string
	if stats.PriorComparison.Available {
		comparisonInstruction = "\n\nMANDATORY SECTION REQUIREMENT:\n" +
			"In addition to your other sections, you MUST include a dedicated <h3>Period-over-Period Avian Comparison</h3> section in your narrative.\n" +
			"In this section, thoroughly analyze the changes between the current window and the preceding baseline window:\n" +
			"1. Avian Population & Species Shifts: Discuss the overall volume shift (Total Detections delta) and species richness change between the two periods.\n" +
			"2. Species Turnover & Migration: Detail specific new arrival species (migrant arrivals), species that departed, and species with notable surges or declines.\n" +
			"3. Habitat & Ecological Context: Analyze what these species shifts indicate about local habitat usage, seasonal foraging/breeding behaviors, and migration progression for this time period.\n" +
			"4. Environmental Context: Note how any observed weather or temperature changes between the periods (from weatherComparison and dailyWeather) may have influenced these avian movements."
	}

	// Safety note appended after facts — does not override output format instructions.
	const factsSafetyNote = "Do not invent metrics; only use the provided Facts and say \"unavailable\" when data is missing. Do not include <img> tags, image URLs, or arbitrary external links. When discussing weather, temperature, or wind speed in the narrative, always format them using the user's preferred units as specified in temperatureUnitPreference and windSpeedUnitPreference in the Facts. The provided Facts include the species-level breakdown table (with individual counts, share percentage, and average confidence), prior-period comparison data (preceding window detections, new arrivals, departed species, surges, and weather comparison deltas), multi-day daily weather history, 24-hour activity distribution, confidence buckets, and solar schedule."

	if strings.TrimSpace(settings.AI.SystemPrompt) != "" {
		return settings.AI.SystemPrompt + comparisonInstruction + "\n\n" + factsSafetyNote + "\n\nFacts:\n" + string(factsJSON)
	}

	return "You are generating narrative commentary for a BirdNET-Go daily report. " +
		"Write a clear, detailed summary in well-structured HTML using <h3>, <p>, <ul>, <li>, and <table> tags. " +
		"Do not include <html> or <body> tags. " +
		comparisonInstruction + "\n\n" +
		factsSafetyNote + "\n\nFacts:\n" + string(factsJSON)
}

func (s *ReportService) generateNarrative(ctx context.Context, stats *reportStats, apiKey string) (narrative, providerID, model string, err error) {
	settings := s.settingsSnapshot()
	requestCtx, cancel := context.WithTimeout(ctx, llmRequestTimeout)
	defer cancel()
	providerID = effectiveProvider(settings.AI.Provider)
	model = effectiveModel(settings.AI)
	provider, provErr := llm.NewProvider(settings.AI, apiKey, s.log)
	if provErr != nil {
		return "", providerID, model, errors.Newf("failed to create AI provider %q: %w", providerID, provErr).
			Component("ai").
			Category(errors.CategoryConfiguration).
			Build()
	}

	prompt := s.buildPrompt(stats)

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

func (s *ReportService) renderPeriodComparison(stats *reportStats) string {
	if !stats.PriorComparison.Available {
		return ""
	}

	settings := s.settingsSnapshot()
	isFahrenheit := settings.Realtime.Dashboard.TemperatureUnit == conf.TemperatureUnitFahrenheit

	var b strings.Builder
	fmt.Fprintf(&b, "<h2>Period Comparison (vs. Preceding %d Days)</h2>\n", stats.PriorComparison.WindowDays)
	b.WriteString("<table><thead><tr><th>Metric</th><th>Current Period</th><th>Prior Period</th><th>Change</th></tr></thead><tbody>\n")

	fmt.Fprintf(&b, "<tr><td>Total detections</td><td>%d</td><td>%d</td><td>%+d (%s)</td></tr>\n",
		stats.TotalDetections, stats.PriorComparison.PriorDetections, stats.PriorComparison.DetectionDelta, stats.PriorComparison.DetectionChangePct)

	speciesDeltaSign := fmt.Sprintf("%+d", stats.PriorComparison.SpeciesDelta)
	fmt.Fprintf(&b, "<tr><td>Unique species</td><td>%d</td><td>%d</td><td>%s</td></tr>\n",
		stats.UniqueSpecies, stats.PriorComparison.PriorUniqueSpecies, speciesDeltaSign)

	fmt.Fprintf(&b, "<tr><td>New arrival species</td><td>%d</td><td>—</td><td>+%d new</td></tr>\n",
		len(stats.PriorComparison.NewArrivals), len(stats.PriorComparison.NewArrivals))
	fmt.Fprintf(&b, "<tr><td>Departed species</td><td>—</td><td>%d</td><td>-%d departed</td></tr>\n",
		len(stats.PriorComparison.DepartedSpecies), len(stats.PriorComparison.DepartedSpecies))

	if stats.Weather.Available && stats.PriorWeather.Available {
		if isFahrenheit {
			currAvg := stats.Weather.TempAvg*1.8 + 32
			priorAvg := stats.PriorWeather.TempAvg*1.8 + 32
			currMin := stats.Weather.TempMin*1.8 + 32
			currMax := stats.Weather.TempMax*1.8 + 32
			priorMin := stats.PriorWeather.TempMin*1.8 + 32
			priorMax := stats.PriorWeather.TempMax*1.8 + 32
			currWind := stats.Weather.WindAvg * 2.23694
			priorWind := stats.PriorWeather.WindAvg * 2.23694

			fmt.Fprintf(&b, "<tr><td>Average temperature</td><td>%.1f°F (%.1f–%.1f°F)</td><td>%.1f°F (%.1f–%.1f°F)</td><td>%+.1f°F</td></tr>\n",
				currAvg, currMin, currMax, priorAvg, priorMin, priorMax, currAvg-priorAvg)
			fmt.Fprintf(&b, "<tr><td>Average wind speed</td><td>%.1f mph</td><td>%.1f mph</td><td>%+.1f mph</td></tr>\n",
				currWind, priorWind, currWind-priorWind)
		} else {
			fmt.Fprintf(&b, "<tr><td>Average temperature</td><td>%.1f°C (%.1f–%.1f°C)</td><td>%.1f°C (%.1f–%.1f°C)</td><td>%+.1f°C</td></tr>\n",
				stats.Weather.TempAvg, stats.Weather.TempMin, stats.Weather.TempMax,
				stats.PriorWeather.TempAvg, stats.PriorWeather.TempMin, stats.PriorWeather.TempMax,
				stats.Weather.TempAvg-stats.PriorWeather.TempAvg)
			fmt.Fprintf(&b, "<tr><td>Average wind speed</td><td>%.1f m/s</td><td>%.1f m/s</td><td>%+.1f m/s</td></tr>\n",
				stats.Weather.WindAvg, stats.PriorWeather.WindAvg, stats.Weather.WindAvg-stats.PriorWeather.WindAvg)
		}
		fmt.Fprintf(&b, "<tr><td>Barometric pressure</td><td>%.1f hPa</td><td>%.1f hPa</td><td>%+.1f hPa</td></tr>\n",
			stats.Weather.Pressure, stats.PriorWeather.Pressure, stats.Weather.Pressure-stats.PriorWeather.Pressure)
		fmt.Fprintf(&b, "<tr><td>Average humidity</td><td>%.1f%%</td><td>%.1f%%</td><td>%+.1f%%</td></tr>\n",
			stats.Weather.Humidity, stats.PriorWeather.Humidity, stats.Weather.Humidity-stats.PriorWeather.Humidity)
	}
	b.WriteString("</tbody></table>\n")

	hasTurnover := len(stats.PriorComparison.NewArrivals) > 0 || len(stats.PriorComparison.TopSurges) > 0 ||
		len(stats.PriorComparison.TopDeclines) > 0 || len(stats.PriorComparison.DepartedSpecies) > 0

	if hasTurnover {
		b.WriteString("\n<h3>Notable Species Shifts & Migration Movements</h3>\n")
		b.WriteString("<table><thead><tr><th>Movement</th><th>Species</th><th>Current Count</th><th>Prior Count</th><th>Change</th></tr></thead><tbody>\n")

		for _, item := range stats.PriorComparison.NewArrivals {
			fmt.Fprintf(&b, "<tr><td><strong>New Arrival</strong></td><td>%s</td><td>%d</td><td>0</td><td>+100%% (new)</td></tr>\n",
				esc(item.CommonName), item.CurrentCount)
		}
		for _, item := range stats.PriorComparison.TopSurges {
			fmt.Fprintf(&b, "<tr><td>Surge</td><td>%s</td><td>%d</td><td>%d</td><td>%+d (%s)</td></tr>\n",
				esc(item.CommonName), item.CurrentCount, item.PriorCount, item.Delta, item.ChangePct)
		}
		for _, item := range stats.PriorComparison.TopDeclines {
			fmt.Fprintf(&b, "<tr><td>Decline</td><td>%s</td><td>%d</td><td>%d</td><td>%d (%s)</td></tr>\n",
				esc(item.CommonName), item.CurrentCount, item.PriorCount, item.Delta, item.ChangePct)
		}
		for _, item := range stats.PriorComparison.DepartedSpecies {
			fmt.Fprintf(&b, "<tr><td>Departed</td><td>%s</td><td>0</td><td>%d</td><td>-%d (absent)</td></tr>\n",
				esc(item.CommonName), item.PriorCount, item.PriorCount)
		}
		b.WriteString("</tbody></table>")
	}

	return b.String()
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
	settings := s.settingsSnapshot()
	isFahrenheit := settings.Realtime.Dashboard.TemperatureUnit == conf.TemperatureUnitFahrenheit

	var tempStr, windStr string
	if isFahrenheit {
		tempMinF := stats.Weather.TempMin*1.8 + 32
		tempMaxF := stats.Weather.TempMax*1.8 + 32
		tempAvgF := stats.Weather.TempAvg*1.8 + 32
		windAvgMph := stats.Weather.WindAvg * 2.23694
		windMaxMph := stats.Weather.WindMax * 2.23694
		tempStr = fmt.Sprintf("min %.1f°F / max %.1f°F / avg %.1f°F", tempMinF, tempMaxF, tempAvgF)
		windStr = fmt.Sprintf("avg %.1f mph / max %.1f mph", windAvgMph, windMaxMph)
	} else {
		tempStr = fmt.Sprintf("min %.1f°C / max %.1f°C / avg %.1f°C", stats.Weather.TempMin, stats.Weather.TempMax, stats.Weather.TempAvg)
		windStr = fmt.Sprintf("avg %.1f m/s / max %.1f m/s", stats.Weather.WindAvg, stats.Weather.WindMax)
	}

	rows := []string{
		"<h2>Weather Summary</h2>",
		`<table><thead><tr><th>Metric</th><th>Value</th></tr></thead><tbody>`,
		fmt.Sprintf("<tr><td>Temperature</td><td>%s</td></tr>", tempStr),
		fmt.Sprintf("<tr><td>Wind</td><td>%s</td></tr>", windStr),
		fmt.Sprintf("<tr><td>Humidity</td><td>%.1f%%</td></tr>", stats.Weather.Humidity),
		fmt.Sprintf("<tr><td>Conditions</td><td>%s</td></tr>", esc(orUnknown(stats.Weather.Condition))),
		"</tbody></table>",
	}

	if len(stats.DailyWeather) > 1 {
		rows = append(rows, "\n<h3>Daily Weather History</h3>", `<table><thead><tr><th>Date</th><th>Temp Range</th><th>Avg Temp</th><th>Wind</th><th>Humidity</th><th>Conditions</th></tr></thead><tbody>`)
		for _, dw := range stats.DailyWeather {
			var dTempRange, dAvgTemp, dWind string
			if isFahrenheit {
				dTempRange = fmt.Sprintf("%.1f–%.1f°F", dw.TempMin*1.8+32, dw.TempMax*1.8+32)
				dAvgTemp = fmt.Sprintf("%.1f°F", dw.TempAvg*1.8+32)
				dWind = fmt.Sprintf("%.1f mph (gust %.1f)", dw.WindAvg*2.23694, dw.WindMax*2.23694)
			} else {
				dTempRange = fmt.Sprintf("%.1f–%.1f°C", dw.TempMin, dw.TempMax)
				dAvgTemp = fmt.Sprintf("%.1f°C", dw.TempAvg)
				dWind = fmt.Sprintf("%.1f m/s (gust %.1f)", dw.WindAvg, dw.WindMax)
			}
			rows = append(rows, fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%.1f%%</td><td>%s</td></tr>",
				esc(dw.Date), dTempRange, dAvgTemp, dWind, dw.Humidity, esc(orUnknown(dw.Condition))))
		}
		rows = append(rows, "</tbody></table>")
	}

	return strings.Join(rows, "\n")
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
