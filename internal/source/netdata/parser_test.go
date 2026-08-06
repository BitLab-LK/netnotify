package netdata

import "testing"

func TestParse(t *testing.T) {
	p := Parser{DefaultRecipient: "group", Provider: "gowa"}
	n, err := p.Parse(map[string]string{"alarm": "CPU", "status": "CRITICAL", "hostname": "node1"})
	if err != nil {
		t.Fatal(err)
	}
	if n.Severity != "critical" || n.Recipient != "group" {
		t.Fatalf("unexpected notification: %#v", n)
	}
}
