package app

// CLIConfig holds the global configuration for CLI commands.
// It maps directly to Viper configuration keys.
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
	// ProgressNone disables progress reporting.
	ProgressNone ProgressMode = "none"
	// ProgressHuman enables human-readable progress reporting (default).
	ProgressHuman ProgressMode = "human"
	// ProgressJSON enables JSON-formatted progress reporting for machine consumption.
	ProgressJSON ProgressMode = "json"
)

// ProgressOutput represents a structured progress event.
// It is used for JSON output generation during long-running operations.
type ProgressOutput struct {
	Timestamp string  `json:"timestamp"`
	Stage     string  `json:"stage"`
	Message   string  `json:"message"`
	Current   int64   `json:"current,omitempty"`
	Total     int64   `json:"total,omitempty"`
	Percent   float64 `json:"percent,omitempty"`
}
