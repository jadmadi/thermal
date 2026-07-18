package thermal

type Tool string

const (
	ToolAll         Tool = "all"
	ToolAuto        Tool = "auto"
	ToolMiMoCode    Tool = "mimocode"
	ToolOpenCode    Tool = "opencode"
	ToolCodex       Tool = "codex"
	ToolDevin       Tool = "devin"
	ToolAgy         Tool = "agy"
	ToolCommandCode Tool = "command-code"
	ToolCodewhale   Tool = "codewhale"
)

type Options struct {
	Tool    string
	DBPath  string
	Weeks   int
	JSON    bool
	NoColor bool
	Verbose bool
}

type Summary struct {
	Tool             string  `json:"tool"`
	Sessions         int     `json:"sessions"`
	LifetimeTokens   int64   `json:"lifetimeTokens"`
	InputTokens      int64   `json:"inputTokens"`
	OutputTokens     int64   `json:"outputTokens"`
	ReasoningTokens  int64   `json:"reasoningTokens"`
	CacheTokens      int64   `json:"cacheTokens"`
	Cost             float64 `json:"cost"`
	LongestSessionMs int64   `json:"longestSessionMs"`
	// New analytics fields — populated by tools that have them; 0/nil otherwise.
	LinesAdded       int64             `json:"linesAdded"`
	LinesDeleted     int64             `json:"linesDeleted"`
	FilesTouched     int64             `json:"filesTouched"`
	AgentBreakdown   map[string]int    `json:"agentBreakdown,omitempty"`
	ModelBreakdown   map[string]int64  `json:"modelBreakdown,omitempty"`
}

type DailyRow struct {
	Day    string `json:"day"`
	Tokens int64  `json:"tokens"`
	Turns  int    `json:"turns"`
}

type DayActivity struct {
	Tokens int64
	Turns  int
}

type ToolResult struct {
	Tool          Tool
	Name          string
	Summary       Summary
	Daily         []DailyRow
	CurrentStreak int
	LongestStreak int
	ActiveDays    int
	TotalActivity int64
	DataPath      string
}
