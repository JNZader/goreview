package report

import (
	"encoding/json"
	"io"

	"github.com/JNZader/goreview/goreview/internal/review"
)

// JSONReporter generates JSON reports.
type JSONReporter struct {
	Indent bool
}

func (r *JSONReporter) Format() string { return "json" }

func (r *JSONReporter) Generate(result *review.Result) (string, error) {
	var data []byte
	var err error

	if r.Indent {
		data, err = json.MarshalIndent(result, "", "  ")
	} else {
		data, err = json.Marshal(result)
	}

	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *JSONReporter) Write(result *review.Result, w io.Writer) error {
	encoder := json.NewEncoder(w)
	if r.Indent {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(result)
}
