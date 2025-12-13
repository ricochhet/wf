package config

type Flags struct {
	Dotfile   string `json:"dotfile"`
	DryRun    bool   `json:"dryRun"`
	LogTime   bool   `json:"logTime"`
	MongoURI  string `json:"mongoUri"`
	Database  string `json:"database"`
	DBData    string `json:"dbdata"`
	WFData    string `json:"wfdata"`
	Index     int    `json:"index"`
	Mode      string `json:"mode"`
	Global    string `json:"string"`
	Debug     bool   `json:"debug"`
	QuickEdit bool   `json:"quickEdit"` // noop on non-Windows systems.
	// Internal
	Args []string `json:"args"`
}
