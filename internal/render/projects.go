package render

import (
	"fmt"
	"os"
	"strings"

	"github.com/jadmadi/thermal/internal/thermal"
)

const (
	projectWidth = 34
	toolsWidth   = 12
	daysWidth    = 4
	lastWidth    = 10
	rankWidth    = 3
	defaultTop   = 20
)

// RenderProjects prints a project leaderboard ranked by tokens. Rows past the
// top slice are summarised instead of printed, and --json carries them all.
func RenderProjects(rep thermal.ProjectReport, noColor bool) string {
	colors := !noColor && IsTerminal() && os.Getenv("NO_COLOR") == ""

	highlight := func(s string) string { return ColorCode(colors, "1;38;5;255", s) }
	dim := func(s string) string { return ColorCode(colors, "38;5;239", s) }
	gold := func(s string) string { return ColorCode(colors, "1;33", s) }

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s %s %s\n\n",
		highlight("Thermal"),
		dim("·"),
		highlight("projects"),
	))

	if len(rep.Rows) == 0 {
		sb.WriteString(fmt.Sprintf("  %s\n", dim("No project activity in the selected window.")))
		return sb.String()
	}

	headers := []string{"#", "Project", "Tools", "Tokens", "Cost", "Days", "Last"}
	alignRight := []bool{false, false, false, true, true, true, false}
	widths := []int{rankWidth, projectWidth, toolsWidth, numberWidth, numberWidth, daysWidth, lastWidth}

	sb.WriteString("  ")
	for i, h := range headers {
		cell := thermal.PadRight(h, widths[i])
		if alignRight[i] {
			cell = thermal.PadLeft(h, widths[i])
		}
		sb.WriteString(dim(cell))
		if i < len(headers)-1 {
			sb.WriteString("  ")
		}
	}
	sb.WriteString("\n")

	rule := 2 * (len(widths) - 1)
	for _, w := range widths {
		rule += w
	}
	sb.WriteString("  " + dim(strings.Repeat("─", rule)) + "\n")

	shown := len(rep.Rows)
	if shown > defaultTop {
		shown = defaultTop
	}
	for i := 0; i < shown; i++ {
		row := rep.Rows[i]
		printProjectRow(&sb, i+1, row, widths, nil)
	}
	if rest := len(rep.Rows) - shown; rest > 0 {
		sb.WriteString(fmt.Sprintf("  %s\n", dim(fmt.Sprintf("… and %d more (use --json for the full list)", rest))))
	}

	sb.WriteString("  " + dim(strings.Repeat("─", rule)) + "\n")
	totals := rep.Totals
	sb.WriteString("  ")
	cells := []string{
		thermal.PadRight("", rankWidth),
		thermal.PadRight("Total", projectWidth),
		thermal.PadRight(strings.Join(totals.Tools, ","), toolsWidth),
		thermal.PadLeft(thermal.CompactNumber(totals.Tokens), numberWidth),
		thermal.PadLeft(formatCostOrDash(totals.Cost), numberWidth),
		thermal.PadLeft(fmt.Sprintf("%d", totals.ActiveDays), daysWidth),
		thermal.PadRight(totals.LastDay, lastWidth),
	}
	for i, c := range cells {
		sb.WriteString(gold(c))
		if i < len(cells)-1 {
			sb.WriteString("  ")
		}
	}
	sb.WriteString("\n")

	if totals.EstimatedCost > 0 {
		sb.WriteString(fmt.Sprintf("\n  %s\n", dim(fmt.Sprintf("~ %s of the total is estimated from pricing data.",
			formatCost(totals.EstimatedCost)))))
	}
	if len(totals.MissingPricing) > 0 {
		models := totals.MissingPricing
		if len(models) > 4 {
			models = append(append([]string{}, models[:4]...), fmt.Sprintf("+%d more", len(models)-4))
		}
		sb.WriteString(fmt.Sprintf("  %s\n", dim("No pricing for: "+strings.Join(models, ", "))))
	}

	sb.WriteString("\n")
	return sb.String()
}

func printProjectRow(sb *strings.Builder, rank int, row thermal.ProjectRow, widths []int, style func(string) string) {
	cells := []string{
		thermal.PadLeft(fmt.Sprintf("%d.", rank), widths[0]),
		thermal.PadRight(projectLabel(row.Project, widths[1]), widths[1]),
		thermal.PadRight(truncate(strings.Join(row.Tools, ","), widths[2]), widths[2]),
		thermal.PadLeft(thermal.CompactNumber(row.Tokens), widths[3]),
		thermal.PadLeft(formatCostOrDash(row.Cost), widths[4]),
		thermal.PadLeft(fmt.Sprintf("%d", row.ActiveDays), widths[5]),
		thermal.PadRight(row.LastDay, widths[6]),
	}
	sb.WriteString("  ")
	for i, c := range cells {
		if style != nil {
			c = style(c)
		}
		sb.WriteString(c)
		if i < len(cells)-1 {
			sb.WriteString("  ")
		}
	}
	sb.WriteString("\n")
}

func formatCostOrDash(v float64) string {
	if v <= 0 {
		return "—"
	}
	return formatCost(v)
}

// projectLabel shortens a project path for the table. The home directory
// becomes ~, and long paths keep their tail segments because the last parts
// identify the repository.
func projectLabel(path string, width int) string {
	label := thermal.FormatPath(path)
	if runeLen(label) <= width {
		return label
	}
	parts := strings.Split(strings.Trim(label, "/"), "/")
	for i := 1; i < len(parts); i++ {
		candidate := "…/" + strings.Join(parts[i:], "/")
		if runeLen(candidate) <= width {
			return candidate
		}
	}
	return truncate(label, width)
}
