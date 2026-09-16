package thermal

import "time"

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
	ToolZCode       Tool = "zcode"
	ToolGrok        Tool = "grok"
	ToolMuse        Tool = "muse"
	ToolClaude      Tool = "claude"
	ToolDroid       Tool = "droid"
)

type Options struct {
	Tool        string
	DBPath      string
	Weeks       int
	JSON        bool
	NoColor     bool
	Verbose     bool
	Report      string // "", "daily", "weekly", "monthly", "projects"
	Since       string // YYYY-MM-DD or YYYYMMDD
	Until       string
	Last        int
	Order       string // "asc" or "desc", default "desc"
	Breakdown   bool
	StartOfWeek string // sunday..saturday, default sunday
	Offline     bool   // never fetch pricing, use cache only
	NoEstimate  bool   // report stored cost only, skip pricing
	Sort        string // tokens, cost, days, recent; default tokens
	Top         int    // project rows to print, 0 means all
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
	LinesAdded     int64            `json:"linesAdded"`
	LinesDeleted   int64            `json:"linesDeleted"`
	FilesTouched   int64            `json:"filesTouched"`
	AgentBreakdown map[string]int   `json:"agentBreakdown,omitempty"`
	ModelBreakdown map[string]int64 `json:"modelBreakdown,omitempty"`
}

// ModelTokens holds per-model token counts for a single day. Loaders fill it
// only when the source records a model per message or session. Cache reads and
// writes stay separate because their prices differ. Unclassified holds tokens
// whose type the source does not break down, such as a session-level total.
type ModelTokens struct {
	Input        int64 `json:"input,omitempty"`
	Output       int64 `json:"output,omitempty"`
	Reasoning    int64 `json:"reasoning,omitempty"`
	CacheRead    int64 `json:"cacheRead,omitempty"`
	CacheWrite   int64 `json:"cacheWrite,omitempty"`
	Unclassified int64 `json:"unclassified,omitempty"`
}

// Cache returns the combined cache read and write count.
func (m ModelTokens) Cache() int64 {
	return m.CacheRead + m.CacheWrite
}

// Total returns the sum of every token type.
func (m ModelTokens) Total() int64 {
	return m.Input + m.Output + m.Reasoning + m.CacheRead + m.CacheWrite + m.Unclassified
}

// Add returns the element-wise sum of two token counts.
func (m ModelTokens) Add(o ModelTokens) ModelTokens {
	return ModelTokens{
		Input:        m.Input + o.Input,
		Output:       m.Output + o.Output,
		Reasoning:    m.Reasoning + o.Reasoning,
		CacheRead:    m.CacheRead + o.CacheRead,
		CacheWrite:   m.CacheWrite + o.CacheWrite,
		Unclassified: m.Unclassified + o.Unclassified,
	}
}

// DailyRow is one calendar day of activity for one tool. Tokens is the total
// across all token types; the type fields are additive and stay zero for
// activity-only tools. Cost holds cost recorded by the source, never an
// estimate.
type DailyRow struct {
	Day       string                 `json:"day"`
	Tokens    int64                  `json:"tokens"`
	Turns     int                    `json:"turns"`
	Input     int64                  `json:"input,omitempty"`
	Output    int64                  `json:"output,omitempty"`
	Reasoning int64                  `json:"reasoning,omitempty"`
	Cache     int64                  `json:"cache,omitempty"`
	Cost      float64                `json:"cost,omitempty"`
	Models    map[string]ModelTokens `json:"models,omitempty"`
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

// Grain is the bucket size for a period report.
type Grain string

const (
	GrainDay   Grain = "daily"
	GrainWeek  Grain = "weekly"
	GrainMonth Grain = "monthly"
)

// AggregateOptions controls period bucketing and row filtering. Since and
// Until accept YYYY-MM-DD or YYYYMMDD. Last counts whole periods back from
// Now, matching ccusage semantics: --last 1 is today, this week, or this
// month. Last cannot be combined with Since or Until. Now defaults to
// time.Now when zero and exists for deterministic tests.
type AggregateOptions struct {
	StartOfWeek time.Weekday
	Since       string
	Until       string
	Last        int
	Order       string // "asc" or "desc"; anything else means desc
	Now         time.Time
}

// Pricer estimates cost for a day that carries no stored cost. Implementations
// return the estimated USD cost and the names of models they have no price
// for. A nil Pricer disables estimation.
type Pricer interface {
	PriceDay(day DailyRow) (cost float64, missing []string)
}

// PeriodRow is one bucket of a daily, weekly, or monthly report. Cost is the
// total of stored and estimated cost. EstimatedCost is the part derived from
// pricing data and MissingPricing lists models that had no price. Models is
// the folded per-model token map for the bucket.
type PeriodRow struct {
	Period         string                 `json:"period"`
	Models         map[string]ModelTokens `json:"models,omitempty"`
	Input          int64                  `json:"inputTokens"`
	Output         int64                  `json:"outputTokens"`
	Reasoning      int64                  `json:"reasoningTokens"`
	Cache          int64                  `json:"cacheTokens"`
	Tokens         int64                  `json:"totalTokens"`
	Turns          int                    `json:"turns"`
	ActiveDays     int                    `json:"activeDays"`
	StoredCost     float64                `json:"storedCost"`
	EstimatedCost  float64                `json:"estimatedCost,omitempty"`
	Cost           float64                `json:"cost"`
	MissingPricing []string               `json:"missingPricing,omitempty"`
}

// Report is the payload behind thermal daily, weekly, and monthly. Rows are
// sorted per the requested order and Totals sums every row.
type Report struct {
	Type   string      `json:"type"`
	Tool   string      `json:"tool,omitempty"`
	Rows   []PeriodRow `json:"data"`
	Totals PeriodRow   `json:"totals"`
}

// ProjectDay is one day of usage attributed to one project directory. Loaders
// emit these for tools that record where a session ran. Tokens is the total
// recorded by the source; the type fields are disjoint and add up to it. Tool
// is filled by the caller with the tool's display name.
type ProjectDay struct {
	Project    string                 `json:"project"`
	Day        string                 `json:"day"`
	Tool       string                 `json:"tool,omitempty"`
	Tokens     int64                  `json:"tokens"`
	Input      int64                  `json:"input,omitempty"`
	Output     int64                  `json:"output,omitempty"`
	Reasoning  int64                  `json:"reasoning,omitempty"`
	CacheRead  int64                  `json:"cacheRead,omitempty"`
	CacheWrite int64                  `json:"cacheWrite,omitempty"`
	Cost       float64                `json:"cost,omitempty"`
	Turns      int                    `json:"turns,omitempty"`
	Models     map[string]ModelTokens `json:"models,omitempty"`
}

// ProjectRow aggregates usage for one project across tools and time.
type ProjectRow struct {
	Project        string                 `json:"project"`
	Tools          []string               `json:"tools,omitempty"`
	ToolTokens     map[string]int64       `json:"toolTokens,omitempty"`
	Input          int64                  `json:"inputTokens"`
	Output         int64                  `json:"outputTokens"`
	Reasoning      int64                  `json:"reasoningTokens"`
	Cache          int64                  `json:"cacheTokens"`
	Tokens         int64                  `json:"totalTokens"`
	Turns          int                    `json:"turns"`
	ActiveDays     int                    `json:"activeDays"`
	FirstDay       string                 `json:"firstDay,omitempty"`
	LastDay        string                 `json:"lastDay,omitempty"`
	StoredCost     float64                `json:"storedCost"`
	EstimatedCost  float64                `json:"estimatedCost,omitempty"`
	Cost           float64                `json:"cost"`
	MissingPricing []string               `json:"missingPricing,omitempty"`
	Models         map[string]ModelTokens `json:"models,omitempty"`
}

// ProjectReport is the payload behind thermal projects.
type ProjectReport struct {
	Type   string       `json:"type"`
	Rows   []ProjectRow `json:"data"`
	Totals ProjectRow   `json:"totals"`
}

// ProjectOptions filters and sorts a project report. Last counts calendar days
// back from Now, matching the report commands. Sort picks the ranking key:
// tokens (default), cost, days, or recent. Order flips the ranking: anything
// other than "asc" means largest or newest first.
type ProjectOptions struct {
	Since string
	Until string
	Last  int
	Sort  string
	Order string
	Now   time.Time
}

// ModelRow aggregates one model across tools and time. Cost is estimated from
// pricing data because recorded cost attaches to a session or day, never to a
// single model.
type ModelRow struct {
	Model          string   `json:"model"`
	Tools          []string `json:"tools,omitempty"`
	Input          int64    `json:"inputTokens"`
	Output         int64    `json:"outputTokens"`
	Reasoning      int64    `json:"reasoningTokens"`
	CacheRead      int64    `json:"cacheReadTokens"`
	CacheWrite     int64    `json:"cacheWriteTokens"`
	Tokens         int64    `json:"totalTokens"`
	Days           int      `json:"activeDays"`
	FirstDay       string   `json:"firstDay,omitempty"`
	LastDay        string   `json:"lastDay,omitempty"`
	Cost           float64  `json:"estimatedCost"`
	MissingPricing []string `json:"missingPricing,omitempty"`
}

// ModelReport is the payload behind thermal models.
type ModelReport struct {
	Type   string     `json:"type"`
	Rows   []ModelRow `json:"data"`
	Totals ModelRow   `json:"totals"`
}

// ModelOptions filters and sorts a model report. Sort picks the ranking key:
// tokens (default) or cost. Order flips the ranking.
type ModelOptions struct {
	Since string
	Until string
	Last  int
	Sort  string
	Order string
	Now   time.Time
}
