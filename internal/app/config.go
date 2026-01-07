package app

type CLIConfig struct {
	RootPath     string
	Include      []string
	Exclude      []string
	Output       string
	MaxSize      int64
	EnforceLimit bool

	// LLM configuration
	SendGemini    bool
	GeminiModel   string
	GeminiOutput  string
	GeminiTimeout int

	// Template configuration
	Template   string
	Task       string
	Rules      string
	CustomVars map[string]string

	// Scanner overrides
	Workers        int
	IncludeHidden  bool
	IncludeIgnored bool

	ProgressMode ProgressMode
}

// ProgressMode defines how progress is reported during CLI operations
type ProgressMode string

const (
	ProgressNone  ProgressMode = "none"
	ProgressHuman ProgressMode = "human"
	ProgressJSON  ProgressMode = "json"
)

// ProgressOutput represents a progress event for output
type ProgressOutput struct {
	Timestamp string  `json:"timestamp"`
	Stage     string  `json:"stage"`
	Message   string  `json:"message"`
	Current   int64   `json:"current,omitempty"`
	Total     int64   `json:"total,omitempty"`
	Percent   float64 `json:"percent,omitempty"`
}
