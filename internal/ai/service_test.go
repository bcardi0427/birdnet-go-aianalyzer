package ai

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tphakala/birdnet-go/internal/conf"
	"github.com/tphakala/birdnet-go/internal/datastore/v2/entities"
)

func TestBuildFacts_IncludesSpeciesBreakdownAndHourlyDistribution(t *testing.T) {
	settings := &conf.Settings{
		AI: conf.AISettings{
			ReportDays: 3,
		},
		BirdNET: conf.BirdNETConfig{
			Latitude:  35.5,
			Longitude: -78.5,
		},
		Realtime: conf.RealtimeSettings{
			Dashboard: conf.Dashboard{
				TemperatureUnit: conf.TemperatureUnitFahrenheit,
			},
		},
	}

	service := &ReportService{
		settings: settings,
	}

	hourly := make([]int, 24)
	hourly[13] = 47
	hourly[14] = 20

	allSpecies := []speciesRow{
		{
			CommonName:       "Clay-coloured Sparrow",
			ScientificName:   "Spizella pallida",
			Detections:       145,
			AvgConfidencePct: 88.7,
			PeakHour:         13,
			FirstSeenUnix:    1779600000,
			LastSeenUnix:     1779620000,
		},
		{
			CommonName:       "Osprey",
			ScientificName:   "Pandion haliaetus",
			Detections:       32,
			AvgConfidencePct: 90.1,
			PeakHour:         10,
			FirstSeenUnix:    1779601000,
			LastSeenUnix:     1779619000,
		},
	}

	stats := &reportStats{
		TotalDetections:      177,
		UniqueSpecies:        2,
		HighConfidencePct:    85.5,
		Confidence95To100:    50,
		Confidence90To94:     40,
		Confidence80To89:     80,
		ConfidenceBelow80:    7,
		PeakHour:             13,
		PeakHourCount:        47,
		QuietHour:            0,
		QuietHourCount:       0,
		HourlyCounts:         hourly,
		AllSpecies:           allSpecies,
		TopSpecies:           allSpecies,
		CivilDawn:            "05:30",
		Sunrise:              "06:00",
		Sunset:               "20:00",
		CivilDusk:            "20:30",
		EBirdContextIncluded: true,
		EBirdRecentBySpecies: map[string]bool{
			"spizella pallida": true,
		},
		Weather: weatherSummary{
			Available: true,
			TempMin:   15.0,
			TempMax:   25.0,
			TempAvg:   20.0,
			Pressure:  1013.2,
			Humidity:  75.0,
			WindAvg:   3.0,
			WindMax:   6.0,
			Condition: "partly cloudy",
		},
	}

	facts := service.buildFacts(stats)
	require.NotNil(t, facts)

	// Summary checks
	assert.Equal(t, 177, facts["totalDetections"])
	assert.Equal(t, 2, facts["uniqueSpecies"])
	assert.Equal(t, "85.5", facts["highConfidencePct"])
	assert.Equal(t, 3, facts["reportWindowDays"])

	// Confidence breakdown
	confBreakdown, ok := facts["confidenceBreakdown"].(map[string]int)
	require.True(t, ok)
	assert.Equal(t, 50, confBreakdown["95To100Pct"])
	assert.Equal(t, 40, confBreakdown["90To94Pct"])
	assert.Equal(t, 80, confBreakdown["80To89Pct"])
	assert.Equal(t, 7, confBreakdown["below80Pct"])

	// Hourly activity
	hourlyList, ok := facts["hourlyActivity"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, hourlyList, 24)
	assert.Equal(t, "13:00–14:00", hourlyList[13]["hour"])
	assert.Equal(t, 47, hourlyList[13]["detections"])

	// Solar times
	solarTimes, ok := facts["solarTimes"].(map[string]string)
	require.True(t, ok)
	assert.Equal(t, "06:00", solarTimes["sunrise"])
	assert.Equal(t, "20:00", solarTimes["sunset"])

	// Location
	location, ok := facts["location"].(map[string]float64)
	require.True(t, ok)
	assert.InDelta(t, 35.5, location["latitude"], 0.0001)
	assert.InDelta(t, -78.5, location["longitude"], 0.0001)

	// Species breakdown
	speciesBreakdown, ok := facts["speciesBreakdown"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, speciesBreakdown, 2)

	first := speciesBreakdown[0]
	assert.Equal(t, "Clay-coloured Sparrow", first["commonName"])
	assert.Equal(t, "Spizella pallida", first["scientificName"])
	assert.Equal(t, 145, first["detections"])
	assert.Equal(t, "81.9%", first["sharePct"])
	assert.Equal(t, "88.7%", first["avgConfidencePct"])
	assert.Equal(t, "13:00–14:00", first["peakHour"])
	assert.Equal(t, true, first["recentInArea"])
	assert.NotEmpty(t, first["firstSeen"])
	assert.NotEmpty(t, first["lastSeen"])

	// Weather preference
	assert.Equal(t, "Fahrenheit (°F)", facts["temperatureUnitPreference"])
	weatherMap, ok := facts["weather"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, weatherMap, "tempAvgF")
	assert.Equal(t, "partly cloudy", weatherMap["condition"])
}

func TestBuildFacts_HandlesEmptySpeciesAndInactiveFeatures(t *testing.T) {
	settings := &conf.Settings{
		AI: conf.AISettings{
			ReportDays: 1,
		},
	}

	service := &ReportService{
		settings: settings,
	}

	stats := &reportStats{
		TotalDetections: 0,
		UniqueSpecies:   0,
	}

	facts := service.buildFacts(stats)
	require.NotNil(t, facts)
	assert.Equal(t, 0, facts["totalDetections"])
	assert.Empty(t, facts["speciesBreakdown"])
	assert.Nil(t, facts["hourlyActivity"])
	assert.Nil(t, facts["solarTimes"])
	assert.Nil(t, facts["location"])
	assert.Nil(t, facts["priorPeriodComparison"])
}

func TestBuildFacts_IncludesPriorPeriodComparison(t *testing.T) {
	const crowSci = "Corvus brachyrhynchos"
	const sparrowSci = "Spizella pallida"

	settings := &conf.Settings{
		AI: conf.AISettings{
			ReportDays: 3,
		},
	}
	service := &ReportService{settings: settings}

	stats := &reportStats{
		TotalDetections: 100,
		UniqueSpecies:   2,
		AllSpecies: []speciesRow{
			{CommonName: "American Crow", ScientificName: crowSci, Detections: 60},
			{CommonName: "Clay-coloured Sparrow", ScientificName: sparrowSci, Detections: 40},
		},
		PriorComparison: priorComparisonStats{
			Available:          true,
			WindowDays:         3,
			PriorDetections:    80,
			DetectionDelta:     20,
			DetectionChangePct: "+25.0%",
			PriorUniqueSpecies: 2,
			SpeciesDelta:       0,
			PriorCounts: map[string]int{
				crowSci:    40,
				sparrowSci: 0,
			},
			NewArrivals: []speciesDelta{
				{CommonName: "Clay-coloured Sparrow", ScientificName: sparrowSci, CurrentCount: 40, PriorCount: 0, Delta: 40},
			},
			TopSurges: []speciesDelta{
				{CommonName: "American Crow", ScientificName: crowSci, CurrentCount: 60, PriorCount: 40, Delta: 20, ChangePct: "+50.0%"},
			},
		},
	}

	facts := service.buildFacts(stats)
	require.NotNil(t, facts)

	priorComp, ok := facts["priorPeriodComparison"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, 3, priorComp["precedingWindowDays"])
	assert.Equal(t, 80, priorComp["priorTotalDetections"])
	assert.Equal(t, 20, priorComp["detectionDelta"])
	assert.Equal(t, "+25.0%", priorComp["detectionChangePct"])

	speciesBreakdown, ok := facts["speciesBreakdown"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, speciesBreakdown, 2)

	crow := speciesBreakdown[0]
	assert.Equal(t, 40, crow["priorDetections"])
	assert.Equal(t, 20, crow["delta"])
	assert.Equal(t, "increased", crow["trend"])

	sparrow := speciesBreakdown[1]
	assert.Equal(t, 0, sparrow["priorDetections"])
	assert.Equal(t, 40, sparrow["delta"])
	assert.Equal(t, "new_arrival", sparrow["trend"])
}

func TestComputePriorComparison(t *testing.T) {
	const crowSci = "Corvus brachyrhynchos"
	const sparrowSci = "Spizella pallida"
	const nuthatchSci = "Sitta carolinensis"

	service := &ReportService{}

	currentBySci := map[string][]*entities.Detection{
		crowSci: {
			{LabelID: 1, Confidence: 0.95},
			{LabelID: 1, Confidence: 0.90},
			{LabelID: 1, Confidence: 0.85},
		},
		sparrowSci: {
			{LabelID: 2, Confidence: 0.92},
			{LabelID: 2, Confidence: 0.88},
		},
	}

	prevDets := []*entities.Detection{
		{LabelID: 1, Confidence: 0.91}, // Corvus brachyrhynchos was 1
		{LabelID: 3, Confidence: 0.89}, // Sitta carolinensis was 1 (departed)
	}

	labelsByID := map[uint]string{
		1: crowSci,
		2: sparrowSci,
		3: nuthatchSci,
	}

	labels := []string{
		"Corvus brachyrhynchos_American Crow_c1",
		"Spizella pallida_Clay-coloured Sparrow_c2",
		"Sitta carolinensis_White-breasted Nuthatch_c3",
	}

	comparison := service.computePriorComparison(t.Context(), currentBySci, prevDets, labelsByID, labels, 3)

	assert.True(t, comparison.Available)
	assert.Equal(t, 3, comparison.WindowDays)
	assert.Equal(t, 2, comparison.PriorDetections)
	assert.Equal(t, 3, comparison.DetectionDelta) // 5 current - 2 prior
	assert.Equal(t, 2, comparison.PriorUniqueSpecies)
	assert.Equal(t, 0, comparison.SpeciesDelta) // 2 current - 2 prior

	// New arrival: Spizella pallida (2 current, 0 prior)
	require.Len(t, comparison.NewArrivals, 1)
	assert.Equal(t, "Clay-coloured Sparrow", comparison.NewArrivals[0].CommonName)
	assert.Equal(t, sparrowSci, comparison.NewArrivals[0].ScientificName)
	assert.Equal(t, 2, comparison.NewArrivals[0].CurrentCount)
	assert.Equal(t, 0, comparison.NewArrivals[0].PriorCount)

	// Departed: Sitta carolinensis (0 current, 1 prior)
	require.Len(t, comparison.DepartedSpecies, 1)
	assert.Equal(t, "White-breasted Nuthatch", comparison.DepartedSpecies[0].CommonName)
	assert.Equal(t, nuthatchSci, comparison.DepartedSpecies[0].ScientificName)

	// Top surge: Corvus brachyrhynchos (1 prior -> 3 current)
	require.Len(t, comparison.TopSurges, 1)
	assert.Equal(t, "American Crow", comparison.TopSurges[0].CommonName)
	assert.Equal(t, 2, comparison.TopSurges[0].Delta)
}

func TestSaveReportHistory_PrunesBeyondFifty(t *testing.T) {
	tempDir := t.TempDir()

	service := &ReportService{
		cacheDir: tempDir,
	}

	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range 55 {
		ts := baseTime.Add(time.Duration(i) * time.Hour)
		service.saveReportHistory([]byte(`{"report":"test"}`), ts)
	}

	histDir := filepath.Join(tempDir, "history")
	entries, err := os.ReadDir(histDir)
	require.NoError(t, err)
	assert.Len(t, entries, 50)
}
