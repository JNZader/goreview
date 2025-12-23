package report

import (
	"encoding/json"
	"io"

	"github.com/JNZader/goreview/goreview/internal/providers"
	"github.com/JNZader/goreview/goreview/internal/review"
)

// SARIFReporter generates SARIF 2.1.0 reports.
type SARIFReporter struct{}

func (r *SARIFReporter) Format() string { return "sarif" }

// SARIF types
type sarifReport struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name    string      `json:"name"`
	Version string      `json:"version"`
	Rules   []sarifRule `json:"rules,omitempty"`
}

type sarifRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description struct {
		Text string `json:"text"`
	} `json:"shortDescription"`
}

type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifMessage    `json:"message"`
	Locations []sarifLocation `json:"locations,omitempty"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	PhysicalLocation struct {
		ArtifactLocation struct {
			URI string `json:"uri"`
		} `json:"artifactLocation"`
		Region *sarifRegion `json:"region,omitempty"`
	} `json:"physicalLocation"`
}

type sarifRegion struct {
	StartLine   int `json:"startLine"`
	EndLine     int `json:"endLine,omitempty"`
	StartColumn int `json:"startColumn,omitempty"`
	EndColumn   int `json:"endColumn,omitempty"`
}

func (r *SARIFReporter) Generate(result *review.Result) (string, error) {
	report := r.buildReport(result)
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *SARIFReporter) Write(result *review.Result, w io.Writer) error {
	report := r.buildReport(result)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func (r *SARIFReporter) buildReport(result *review.Result) *sarifReport {
	report := &sarifReport{
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarifRun{{
			Tool: sarifTool{
				Driver: sarifDriver{
					Name:    "goreview",
					Version: "1.0.0",
				},
			},
			Results: []sarifResult{},
		}},
	}

	for _, file := range result.Files {
		if file.Response == nil {
			continue
		}

		for _, issue := range file.Response.Issues {
			res := sarifResult{
				RuleID:  string(issue.Type),
				Level:   r.mapLevel(issue.Severity),
				Message: sarifMessage{Text: issue.Message},
			}

			if issue.Location != nil {
				loc := sarifLocation{}
				loc.PhysicalLocation.ArtifactLocation.URI = file.File
				if issue.Location.StartLine > 0 {
					loc.PhysicalLocation.Region = &sarifRegion{
						StartLine: issue.Location.StartLine,
						EndLine:   issue.Location.EndLine,
					}
				}
				res.Locations = append(res.Locations, loc)
			}

			report.Runs[0].Results = append(report.Runs[0].Results, res)
		}
	}

	return report
}

func (r *SARIFReporter) mapLevel(severity providers.Severity) string {
	switch severity {
	case providers.SeverityCritical, providers.SeverityError:
		return "error"
	case providers.SeverityWarning:
		return "warning"
	default:
		return "note"
	}
}
