package api

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/labstack/echo/v4"
)

const (
	defaultVisitorLimit = 100
	maxVisitorLimit     = 500
	visitorLogPath      = "logs/visitor.log"
)

type visitorLogEntry struct {
	Time            string `json:"time"`
	Level           string `json:"level"`
	Message         string `json:"msg"`
	Module          string `json:"module"`
	Method          string `json:"method"`
	Path            string `json:"path"`
	Query           string `json:"query"`
	Status          int    `json:"status"`
	IP              string `json:"ip"`
	RealIP          string `json:"real_ip"`
	Host            string `json:"host"`
	Referer         string `json:"referer"`
	UserAgent       string `json:"user_agent"`
	CFConnectingIP  string `json:"cf_connecting_ip"`
	CFCountry       string `json:"cf_country"`
	CFRay           string `json:"cf_ray"`
	XForwardedFor   string `json:"x_forwarded_for"`
	XForwardedProto string `json:"x_forwarded_proto"`
	Tunneled        bool   `json:"tunneled"`
	Authenticated   bool   `json:"authenticated"`
	LatencyMS       int64  `json:"latency_ms"`
}

type visitorCount struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

type visitorLogResponse struct {
	Entries         []visitorLogEntry `json:"entries"`
	TotalReturned   int               `json:"totalReturned"`
	UniqueIPs       int               `json:"uniqueIps"`
	UniqueCountries int               `json:"uniqueCountries"`
	TopIPs          []visitorCount    `json:"topIps"`
	TopCountries    []visitorCount    `json:"topCountries"`
	TopReferers     []visitorCount    `json:"topReferers"`
	TopPaths        []visitorCount    `json:"topPaths"`
	LogPath         string            `json:"logPath"`
}

func (c *Controller) GetVisitorLog(ctx echo.Context) error {
	limit := parseVisitorLimit(ctx.QueryParam("limit"))

	entries, err := readRecentVisitorEntries(visitorLogPath, limit)
	if err != nil {
		if os.IsNotExist(err) {
			return ctx.JSON(http.StatusOK, visitorLogResponse{
				Entries: []visitorLogEntry{},
				LogPath: visitorLogPath,
			})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Could not read visitor log",
		})
	}

	return ctx.JSON(http.StatusOK, buildVisitorLogResponse(entries, visitorLogPath))
}

func parseVisitorLimit(raw string) int {
	if raw == "" {
		return defaultVisitorLimit
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		return defaultVisitorLimit
	}
	if limit > maxVisitorLimit {
		return maxVisitorLimit
	}
	return limit
}

func readRecentVisitorEntries(path string, limit int) ([]visitorLogEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lines := make([]string, 0, limit)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(lines) < limit {
			lines = append(lines, line)
			continue
		}
		copy(lines, lines[1:])
		lines[len(lines)-1] = line
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	entries := make([]visitorLogEntry, 0, len(lines))
	for _, line := range lines {
		var entry visitorLogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if entry.Message != "page visit" {
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func buildVisitorLogResponse(entries []visitorLogEntry, logPath string) visitorLogResponse {
	ipCounts := make(map[string]int)
	countryCounts := make(map[string]int)
	refererCounts := make(map[string]int)
	pathCounts := make(map[string]int)

	for _, entry := range entries {
		if entry.IP != "" {
			ipCounts[entry.IP]++
		}
		if entry.CFCountry != "" {
			countryCounts[entry.CFCountry]++
		}
		if entry.Referer != "" {
			refererCounts[entry.Referer]++
		}
		if entry.Path != "" {
			pathCounts[entry.Path]++
		}
	}

	return visitorLogResponse{
		Entries:         entries,
		TotalReturned:   len(entries),
		UniqueIPs:       len(ipCounts),
		UniqueCountries: len(countryCounts),
		TopIPs:          topVisitorCounts(ipCounts, 10),
		TopCountries:    topVisitorCounts(countryCounts, 10),
		TopReferers:     topVisitorCounts(refererCounts, 10),
		TopPaths:        topVisitorCounts(pathCounts, 10),
		LogPath:         logPath,
	}
}

func topVisitorCounts(counts map[string]int, limit int) []visitorCount {
	items := make([]visitorCount, 0, len(counts))
	for key, count := range counts {
		items = append(items, visitorCount{Key: key, Count: count})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Key < items[j].Key
		}
		return items[i].Count > items[j].Count
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}
