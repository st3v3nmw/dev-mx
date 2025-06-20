package engine

import "encoding/json"

type ValidationResult struct {
	Compliant  bool     `json:"compliant"`
	Violations []string `json:"violations"`
}

func (r ValidationResult) String() string {
	out, _ := json.MarshalIndent(r, "", "  ")
	return string(out)
}
