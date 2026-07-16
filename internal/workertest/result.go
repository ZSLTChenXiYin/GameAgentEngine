package workertest

type CheckResult struct {
	Area      string `json:"area"`
	Operation string `json:"operation"`
	Channel   string `json:"channel,omitempty"`
	Status    string `json:"status"`
	Evidence  string `json:"evidence"`
}

type Collector struct {
	checks []CheckResult
}

func (c *Collector) Add(area, operation, channel, status, evidence string) {
	c.checks = append(c.checks, CheckResult{
		Area:      area,
		Operation: operation,
		Channel:   channel,
		Status:    status,
		Evidence:  evidence,
	})
}

func (c *Collector) Checks() []CheckResult {
	if len(c.checks) == 0 {
		return nil
	}
	out := make([]CheckResult, len(c.checks))
	copy(out, c.checks)
	return out
}
