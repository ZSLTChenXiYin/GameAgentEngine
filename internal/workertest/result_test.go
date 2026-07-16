package workertest

import "testing"

func TestCollectorAddAndCopy(t *testing.T) {
	var c Collector
	c.Add("node", "create", "http", "passed", "node_id=n1")
	checks := c.Checks()
	if len(checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(checks))
	}
	checks[0].Area = "mutated"
	if c.Checks()[0].Area != "node" {
		t.Fatal("expected Checks to return a copy")
	}
}
