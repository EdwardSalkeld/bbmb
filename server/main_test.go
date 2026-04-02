package main

import (
	"flag"
	"testing"
)

func TestParseConfigDefaults(t *testing.T) {
	cfg, err := parseConfig(flag.NewFlagSet("bbmb-server", flag.ContinueOnError), nil)
	if err != nil {
		t.Fatalf("parseConfig returned error: %v", err)
	}

	if cfg.TCPAddress != ":9876" {
		t.Fatalf("TCPAddress = %q, want %q", cfg.TCPAddress, ":9876")
	}

	if cfg.MetricsAddress != ":9877" {
		t.Fatalf("MetricsAddress = %q, want %q", cfg.MetricsAddress, ":9877")
	}
}

func TestParseConfigOverridesPorts(t *testing.T) {
	args := []string{"--port", "1234", "--metrics-port", "4321"}
	cfg, err := parseConfig(flag.NewFlagSet("bbmb-server", flag.ContinueOnError), args)
	if err != nil {
		t.Fatalf("parseConfig returned error: %v", err)
	}

	if cfg.TCPAddress != ":1234" {
		t.Fatalf("TCPAddress = %q, want %q", cfg.TCPAddress, ":1234")
	}

	if cfg.MetricsAddress != ":4321" {
		t.Fatalf("MetricsAddress = %q, want %q", cfg.MetricsAddress, ":4321")
	}
}

func TestParseConfigRejectsOutOfRangePorts(t *testing.T) {
	testCases := []struct {
		name string
		args []string
	}{
		{name: "zero broker port", args: []string{"--port", "0"}},
		{name: "negative metrics port", args: []string{"--metrics-port", "-1"}},
		{name: "broker port too large", args: []string{"--port", "70000"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseConfig(flag.NewFlagSet("bbmb-server", flag.ContinueOnError), tc.args)
			if err == nil {
				t.Fatal("parseConfig returned nil error, want validation failure")
			}
		})
	}
}
