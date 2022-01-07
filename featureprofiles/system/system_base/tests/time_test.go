package system_base_test

import (
	"testing"
	"time"

	"github.com/openconfig/ondatra"
)

// TestCurrentDateTime verifies that the current date and time state path can
// be parsed as RFC3339 time format.
//
// telemetry_path:/system/state/current-datetime
func TestCurrentDateTime(t *testing.T) {
	t.Skip("Need working implementation to validate against")

	dut := ondatra.DUT(t, "dut1")
	now := dut.Telemetry().System().CurrentDatetime().Get(t)
	_, err := time.Parse(time.RFC3339, now)
	if err != nil {
		t.Errorf("Failed to parse current time: got %s: %s", now, err)
	}
}

// TestBootTime verifies the timestamp that the system was last restarted can
// be read and is not an unreasonable value.
//
// telemetry_path:/system/state/boot-time
func TestBootTime(t *testing.T) {
	dut := ondatra.DUT(t, "dut1")
	bt := dut.Telemetry().System().BootTime().Get(t)

	// Boot time should be after Dec 22, 2021 00:00:00 GMT in nanoseconds
	if bt < 1640131200000000000 {
		t.Errorf("Unexpected boot timestamp: got %d; check clock", bt)
	}
}

// TestTimeZone verifies the timezone-name config values can be read and set
//
// config_path:/system/config/timezone-name
// telemetry_path:/system/state/timezone-name
func TestTimeZone(t *testing.T) {
	t.Skip("Need working implementation to validate against")

	testCases := []struct {
		description string
		tz          string
	}{
		{"UTC", "Etc/UTC"},
		{"GMT", "Etc/GMT"},
		{"Short UTC", "UTC"},
		{"Short GMT", "GMT"},
		{"America/Chicago", "America/Chicago"},
		{"PST8PDT", "PST8PDT"},
		{"Europe/London", "Europe/London"},
	}

	dut := ondatra.DUT(t, "dut1")

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			config := dut.Config().System().Clock().TimezoneName()
			state := dut.Telemetry().System().Clock().TimezoneName()

			config.Replace(t, testCase.tz)

			configGot := config.Get(t)
			if configGot != testCase.tz {
				t.Errorf("Config timezone: got %s, want %s", configGot, testCase.tz)
			}

			stateGot := state.Await(t, 5*time.Second, testCase.tz)
			if stateGot.Val(t) != testCase.tz {
				t.Errorf("State domainname: got %v, want %s", stateGot, testCase.tz)
			}

			config.Delete(t)
			if qs := config.Lookup(t); qs.IsPresent() == true {
				t.Errorf("Delete timezone fail: got %v", qs)
			}
		})
	}
}
