package main

import (
	"testing"

	"github.com/peterneutron/powerkit-go/pkg/powerkit"
)

func TestResolveDumpOptions(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		args    []string
		want    powerkit.FetchOptions
		wantErr bool
	}{
		{
			name:   "all",
			source: "all",
			want:   powerkit.FetchOptions{QueryIOKit: true, QuerySMC: true},
		},
		{
			name:   "all fallback",
			source: "all",
			args:   []string{"fallback"},
			want: powerkit.FetchOptions{
				QueryIOKit:             true,
				QuerySMC:               true,
				ForceTelemetryFallback: true,
			},
		},
		{
			name:   "all fallback flag",
			source: "all",
			args:   []string{"--fallback"},
			want: powerkit.FetchOptions{
				QueryIOKit:             true,
				QuerySMC:               true,
				ForceTelemetryFallback: true,
			},
		},
		{
			name:   "smc",
			source: "smc",
			want:   powerkit.FetchOptions{QuerySMC: true},
		},
		{
			name:   "iokit",
			source: "iokit",
			want:   powerkit.FetchOptions{QueryIOKit: true},
		},
		{
			name:    "smc rejects args",
			source:  "smc",
			args:    []string{"fallback"},
			wantErr: true,
		},
		{
			name:    "iokit rejects args",
			source:  "iokit",
			args:    []string{"fallback"},
			wantErr: true,
		},
		{
			name:    "all rejects too many args",
			source:  "all",
			args:    []string{"fallback", "extra"},
			wantErr: true,
		},
		{
			name:    "all rejects unknown arg",
			source:  "all",
			args:    []string{"telemetry"},
			wantErr: true,
		},
		{
			name:    "unknown source",
			source:  "battery",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveDumpOptions(tt.source, tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("resolveDumpOptions returned nil error")
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveDumpOptions returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("resolveDumpOptions = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseMagsafeStateArg(t *testing.T) {
	tests := []struct {
		arg  string
		want powerkit.MagsafeLEDState
	}{
		{arg: colorSystem, want: powerkit.LEDSystem},
		{arg: colorOff, want: powerkit.LEDOff},
		{arg: colorAmber, want: powerkit.LEDAmber},
		{arg: colorGreen, want: powerkit.LEDGreen},
		{arg: colorErrOnce, want: powerkit.LEDErrorOnce},
		{arg: colorErrPermSlow, want: powerkit.LEDErrorPermSlow},
		{arg: colorErrPermFast, want: powerkit.LEDErrorPermFast},
		{arg: colorErrPermOff, want: powerkit.LEDErrorPermOff},
	}

	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			got, ok := parseMagsafeStateArg(tt.arg)
			if !ok {
				t.Fatalf("parseMagsafeStateArg(%q) rejected valid value", tt.arg)
			}
			if got != tt.want {
				t.Fatalf("parseMagsafeStateArg(%q) = %v, want %v", tt.arg, got, tt.want)
			}
		})
	}

	if got, ok := parseMagsafeStateArg("blue"); ok {
		t.Fatalf("parseMagsafeStateArg accepted invalid value as %v", got)
	}
}

func TestParseAssertionType(t *testing.T) {
	tests := []struct {
		arg  string
		want powerkit.AssertionType
	}{
		{arg: assertionTypeSystem, want: powerkit.AssertionTypePreventSystemSleep},
		{arg: assertionTypeDisplay, want: powerkit.AssertionTypePreventDisplaySleep},
	}

	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			got, ok := parseAssertionType(tt.arg)
			if !ok {
				t.Fatalf("parseAssertionType(%q) rejected valid value", tt.arg)
			}
			if got != tt.want {
				t.Fatalf("parseAssertionType(%q) = %v, want %v", tt.arg, got, tt.want)
			}
		})
	}

	if got, ok := parseAssertionType("idle"); ok {
		t.Fatalf("parseAssertionType accepted invalid value as %v", got)
	}
}

func TestParseLowPowerSetArg(t *testing.T) {
	tests := []struct {
		arg  string
		want bool
	}{
		{arg: actionOn, want: true},
		{arg: actionOff, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			got, ok := parseLowPowerSetArg(tt.arg)
			if !ok {
				t.Fatalf("parseLowPowerSetArg(%q) rejected valid value", tt.arg)
			}
			if got != tt.want {
				t.Fatalf("parseLowPowerSetArg(%q) = %v, want %v", tt.arg, got, tt.want)
			}
		})
	}

	if got, ok := parseLowPowerSetArg("enabled"); ok {
		t.Fatalf("parseLowPowerSetArg accepted invalid value as %v", got)
	}
}

func TestMagsafeStateToString(t *testing.T) {
	tests := []struct {
		state powerkit.MagsafeLEDState
		want  string
	}{
		{state: powerkit.LEDSystem, want: "System"},
		{state: powerkit.LEDOff, want: "Off"},
		{state: powerkit.LEDAmber, want: "Amber"},
		{state: powerkit.LEDGreen, want: "Green"},
		{state: powerkit.LEDErrorOnce, want: "Error (Once)"},
		{state: powerkit.LEDErrorPermSlow, want: "Error (Perm Slow)"},
		{state: powerkit.LEDErrorPermFast, want: "Error (Perm Fast)"},
		{state: powerkit.LEDErrorPermOff, want: "Error (Perm Off)"},
		{state: powerkit.MagsafeLEDState(0xff), want: "Unknown (0xff)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := MagsafeStateToString(tt.state); got != tt.want {
				t.Fatalf("MagsafeStateToString(%v) = %q, want %q", tt.state, got, tt.want)
			}
		})
	}
}

func TestEventTypeToString(t *testing.T) {
	tests := []struct {
		eventType powerkit.EventType
		want      string
	}{
		{eventType: powerkit.EventTypeBatteryUpdate, want: "Battery Update"},
		{eventType: powerkit.EventTypeSystemWillSleep, want: "System Will Sleep"},
		{eventType: powerkit.EventTypeSystemDidWake, want: "System Did Wake"},
		{eventType: powerkit.EventType(99), want: "Unknown Event"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := EventTypeToString(tt.eventType); got != tt.want {
				t.Fatalf("EventTypeToString(%v) = %q, want %q", tt.eventType, got, tt.want)
			}
		})
	}
}
