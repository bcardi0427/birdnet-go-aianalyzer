package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tphakala/birdnet-go/internal/conf"
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
}
