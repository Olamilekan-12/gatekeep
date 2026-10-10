package config

import (
	"testing"
	"time"
)

// allKeys is every variable Load reads. Each subtest clears them first so the
// result never depends on the developer's or CI runner's ambient environment.
var allKeys = []string{
	"GATEKEEP_ADDR",
	"GATEKEEP_SHUTDOWN_TIMEOUT",
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    Config
		wantErr bool
	}{
		{
			name: "defaults when nothing is set",
			want: Config{Addr: ":8080", ShutdownTimeout: 10 * time.Second},
		},
		{
			name: "reads values from the environment",
			env: map[string]string{
				"GATEKEEP_ADDR":             ":9090",
				"GATEKEEP_SHUTDOWN_TIMEOUT": "30s",
			},
			want: Config{Addr: ":9090", ShutdownTimeout: 30 * time.Second},
		},
		{
			name: "treats an empty value as unset",
			env:  map[string]string{"GATEKEEP_ADDR": ""},
			want: Config{Addr: ":8080", ShutdownTimeout: 10 * time.Second},
		},
		{
			name:    "rejects a malformed duration",
			env:     map[string]string{"GATEKEEP_SHUTDOWN_TIMEOUT": "10sec"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range allKeys {
				t.Setenv(k, "")
			}
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			got, err := Load()

			if tt.wantErr {
				if err == nil {
					t.Fatalf("Load() = %+v, want an error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() returned an unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Load() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
