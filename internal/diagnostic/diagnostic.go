package diagnostic

type Severity string

const (
	Info Severity = "info"
	Warning Severity = "warning"
	Error Severity = "error"
)

type Item struct {
	Severity Severity `json:"severity" yaml:"severity"`
	Code string `json:"code" yaml:"code"`
	Message string `json:"message" yaml:"message"`
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}
