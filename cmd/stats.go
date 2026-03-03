package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chuy/git-why/internal/git"
	"github.com/spf13/cobra"
)

type StatsColors struct {
	Header string
	File   string
	Author string
	Date   string
	Reset  string
}

var statsColors = StatsColors{
	Header: "\033[1m",
	File:   "\033[35m",
	Author: "\033[33m",
	Date:   "\033[32m",
	Reset:  "\033[0m",
}

var statsPlainColors = StatsColors{
	Header: "",
	File:   "",
	Author: "",
	Date:   "",
	Reset:  "",
}

var statsCmd = &cobra.Command{
	Use:   "stats <file>",
	Short: "Show authorship statistics for a file",
	Long: `Display macro view of who has touched a file most,
what parts change most frequently, and change frequency over time.

Shows:
- Top contributors by lines owned
- Change frequency heatmap
- Monthly activity chart`,
	Args: cobra.ExactArgs(1),
	RunE: runStats,
}

func init() {
	rootCmd.AddCommand(statsCmd)
}

func runStats(cmd *cobra.Command, args []string) error {
	file := args[0]

	repo := git.NewRepo()
	if !repo.IsGitRepo() {
		return fmt.Errorf("not a git repository")
	}

	stats, err := repo.Stats(file)
	if err != nil {
		return err
	}

	if jsonOutput {
		return outputStatsJSON(file, stats)
	}

	printStatsHeader(file, stats)
	printTopAuthors(stats)
	printMonthlyActivity(stats)

	return nil
}

func getStatsColors() StatsColors {
	if plainOutput {
		return statsPlainColors
	}
	return statsColors
}

func printStatsHeader(file string, stats *git.FileStats) {
	c := getStatsColors()

	fmt.Printf("\n%s%sStatistics for %s%s\n", c.Header, c.File, file, c.Reset)
	fmt.Printf("Total lines: %d | Total commits: %d\n\n", stats.TotalLines, stats.TotalCommits)
}

func printTopAuthors(stats *git.FileStats) {
	c := getStatsColors()

	fmt.Printf("%sTop Contributors:%s\n", c.Header, c.Reset)

	maxShow := 5
	if len(stats.Authors) < maxShow {
		maxShow = len(stats.Authors)
	}

	maxAuthorLen := 20
	for i := 0; i < maxShow; i++ {
		if len(stats.Authors[i].Author) > maxAuthorLen {
			maxAuthorLen = len(stats.Authors[i].Author)
		}
	}
	authorFmt := fmt.Sprintf("  %%s@%%-%ds%%s %%s %%5.1f%%%% (%%d lines)\n", maxAuthorLen)

	for i := 0; i < maxShow; i++ {
		author := stats.Authors[i]
		bar := statsProgressBar(author.Percentage, 20)
		fmt.Printf(authorFmt,
			c.Author, author.Author, c.Reset,
			bar,
			author.Percentage, author.LinesCount,
		)
	}

	if len(stats.Authors) > maxShow {
		fmt.Printf("  ... and %d more contributors\n", len(stats.Authors)-maxShow)
	}
	fmt.Println()
}

func printMonthlyActivity(stats *git.FileStats) {
	if len(stats.LinesByMonth) == 0 {
		return
	}

	c := getStatsColors()

	fmt.Printf("%sMonthly Activity (last 12 months):%s\n", c.Header, c.Reset)

	maxChanges := 0
	for _, m := range stats.LinesByMonth {
		if m.Additions+m.Deletions > maxChanges {
			maxChanges = m.Additions + m.Deletions
		}
	}

	for _, m := range stats.LinesByMonth {
		total := m.Additions + m.Deletions
		barWidth := 0
		if maxChanges > 0 {
			barWidth = (total * 20) / maxChanges
		}

		addBar := strings.Repeat("+", statsMin(m.Additions*20/maxChanges, barWidth))
		delBar := strings.Repeat("-", statsMin(m.Deletions*20/maxChanges, 20-barWidth))

		green := "\033[32m"
		red := "\033[31m"
		reset := "\033[0m"
		if plainOutput {
			green = ""
			red = ""
			reset = ""
		}

		barDisplay := fmt.Sprintf("%s%s%s%s%s%s", green, addBar, reset, red, delBar, reset)
		barLen := len(addBar) + len(delBar)
		padding := strings.Repeat(" ", 20-barLen)

		fmt.Printf("  %s %s%s %5d changes\n",
			m.Month.Format("Jan 2006"),
			barDisplay, padding,
			total,
		)
	}
	fmt.Println()
}

// render hotLines removed

func statsProgressBar(percentage float64, width int) string {
	filled := int(percentage * float64(width) / 100)
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return bar
}

func statsMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func outputStatsJSON(file string, stats *git.FileStats) error {
	output := struct {
		File         string            `json:"file"`
		TotalLines   int               `json:"totalLines"`
		TotalCommits int               `json:"totalCommits"`
		Authors      []git.AuthorStats `json:"authors"`
		MonthlyData  []git.MonthStats  `json:"monthlyData,omitempty"`
	}{
		File:         file,
		TotalLines:   stats.TotalLines,
		TotalCommits: stats.TotalCommits,
		Authors:      stats.Authors,
		MonthlyData:  stats.LinesByMonth,
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonData))
	return nil
}
