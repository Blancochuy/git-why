package git

import (
	"bufio"
	"fmt"
	"sort"
	"strings"
	"time"
)

type AuthorStats struct {
	Author     string
	LinesCount int
	Percentage float64
}

type FileStats struct {
	TotalLines   int
	TotalCommits int
	Authors      []AuthorStats
	LinesByMonth []MonthStats
	HotLines     []HotLine
}

type MonthStats struct {
	Month     time.Time
	Changes   int
	Additions int
	Deletions int
}

type HotLine struct {
	LineNum int
	Count   int
	Authors []string
}

func (r *Repo) Stats(file string) (*FileStats, error) {
	blameOutput, err := r.runGit("blame", "--line-porcelain", file)
	if err != nil {
		return nil, err
	}

	authorCounts := make(map[string]int)
	lineAuthors := make(map[int]string)
	lineNum := 0

	scanner := bufio.NewScanner(strings.NewReader(blameOutput))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "author ") {
			author := strings.TrimPrefix(line, "author ")
			lineNum++
			authorCounts[author]++
			lineAuthors[lineNum] = author
		}
	}

	totalLines := lineNum

	var authors []AuthorStats
	for author, count := range authorCounts {
		authors = append(authors, AuthorStats{
			Author:     author,
			LinesCount: count,
			Percentage: float64(count) / float64(totalLines) * 100,
		})
	}

	sort.Slice(authors, func(i, j int) bool {
		return authors[i].LinesCount > authors[j].LinesCount
	})

	commits, err := r.runGit("log", "--oneline", "--follow", "--", file)
	if err != nil {
		return nil, err
	}

	commitCount := len(strings.Split(strings.TrimSpace(commits), "\n"))
	if commits == "" {
		commitCount = 0
	}

	months, err := r.getMonthlyStats(file)
	if err != nil {
		months = nil
	}

	hotLines := r.findHotLines(lineAuthors, 5)

	return &FileStats{
		TotalLines:   totalLines,
		TotalCommits: commitCount,
		Authors:      authors,
		LinesByMonth: months,
		HotLines:     hotLines,
	}, nil
}

func (r *Repo) getMonthlyStats(file string) ([]MonthStats, error) {
	output, err := r.runGit("log", "--format=%aI", "--numstat", "--", file)
	if err != nil {
		return nil, err
	}

	monthlyData := make(map[string]*MonthStats)

	scanner := bufio.NewScanner(strings.NewReader(output))
	var currentDate time.Time

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		if !strings.Contains(line, "\t") {
			if t, err := time.Parse(time.RFC3339, strings.TrimSpace(line)); err == nil {
				currentDate = t
			}
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) >= 2 {
			monthKey := currentDate.Format("2006-01")
			if _, ok := monthlyData[monthKey]; !ok {
				monthlyData[monthKey] = &MonthStats{
					Month: time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, time.UTC),
				}
			}

			var add, del int
			fmt.Sscanf(parts[0], "%d", &add)
			fmt.Sscanf(parts[1], "%d", &del)

			monthlyData[monthKey].Additions += add
			monthlyData[monthKey].Deletions += del
			monthlyData[monthKey].Changes++
		}
	}

	var months []MonthStats
	for _, ms := range monthlyData {
		months = append(months, *ms)
	}

	sort.Slice(months, func(i, j int) bool {
		return months[i].Month.Before(months[j].Month)
	})

	if len(months) > 12 {
		months = months[len(months)-12:]
	}

	return months, nil
}

func (r *Repo) findHotLines(lineAuthors map[int]string, topN int) []HotLine {
	lineCounts := make(map[int]int)
	lineAuthorList := make(map[int][]string)

	for line, author := range lineAuthors {
		lineCounts[line]++
		if !contains(lineAuthorList[line], author) {
			lineAuthorList[line] = append(lineAuthorList[line], author)
		}
	}

	type lineStat struct {
		line  int
		count int
	}

	var stats []lineStat
	for line, count := range lineCounts {
		stats = append(stats, lineStat{line: line, count: count})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].count > stats[j].count
	})

	var hotLines []HotLine
	for i := 0; i < topN && i < len(stats); i++ {
		hotLines = append(hotLines, HotLine{
			LineNum: stats[i].line,
			Count:   stats[i].count,
			Authors: lineAuthorList[stats[i].line],
		})
	}

	return hotLines
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
