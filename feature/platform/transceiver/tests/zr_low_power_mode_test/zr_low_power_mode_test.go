// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zr_low_power_mode_test

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/openconfig/featureprofiles/internal/fptest"
	"github.com/openconfig/featureprofiles/internal/samplestream"
	"github.com/openconfig/ondatra"
	"github.com/openconfig/ondatra/gnmi"
)

const (
	samplingInterval              = 10 * time.Second
	targetOutputPowerdBm          = -10
	targetOutputPowerTolerancedBm = 1
	targetFrequencyHz             = 193100000
	targetFrequencyToleranceHz    = 100000
)

func TestMain(m *testing.M) {
	fptest.RunTests(m)
}

// validateStreamOutput validates that the OC path is streamed in the most recent subscription interval.
func validateStreamOutput(t *testing.T, dut *ondatra.DUTDevice, streams map[string]*samplestream.SampleStream[string]) {
	for key, stream := range streams {
		output := stream.Next(t)
		if output == nil {
			t.Fatalf("OC path for %s not streamed in the most recent subscription interval", key)
		}
		value, ok := output.Val()
		if !ok {
			t.Fatalf("Error capturing streaming value for %s", key)
		}
		if reflect.TypeOf(value).Kind() != reflect.String {
			t.Fatalf("Return value is not type string for key :%s", key)
		}
		if value == "" {
			t.Fatalf("OC path empty for %s", key)
		}
		t.Logf("Value for OC path %s: %s", key, value)
	}
}

// validateOutputPower validates that the output power is streamed in the most recent subscription interval.
func validateOutputPower(t *testing.T, dut *ondatra.DUTDevice, streams map[string]*samplestream.SampleStream[float64]) {
	for key, stream := range streams {
		outputStream := stream.Next(t)
		if outputStream == nil {
			t.Fatalf("OC path for %s not streamed in the most recent subscription interval", key)
		}
		outputPower, ok := outputStream.Val()
		if !ok {
			t.Fatalf("Error capturing streaming value for %s", key)
		}
		// Check output power value is of correct type
		if reflect.TypeOf(outputPower).Kind() != reflect.Float64 {
			t.Fatalf("Return value is not type float64 for key :%s", key)
		}
		t.Logf("Output power for %s: %f", key, outputPower)
	}
}

func TestLowPowerMode(t *testing.T) {
	dut := ondatra.DUT(t, "dut")

	for _, port := range []string{"port1", "port2"} {
		t.Run(fmt.Sprintf("Port:%s", port), func(t *testing.T) {
			dp := dut.Port(t, port)

			gnmi.Update(t, dut, gnmi.OC().Interface(dp.Name()).Enabled().Config(), bool(false))

			// Derive transceiver names from ports.
			tr := gnmi.Get(t, dut, gnmi.OC().Interface(dp.Name()).Transceiver().State())

			// Stream all inventory information.
			streamSerialNo := samplestream.New(t, dut, gnmi.OC().Component(tr).SerialNo().State(), samplingInterval)
			defer streamSerialNo.Close()
			streamPartNo := samplestream.New(t, dut, gnmi.OC().Component(tr).PartNo().State(), samplingInterval)
			defer streamPartNo.Close()
			streamType := samplestream.New(t, dut, gnmi.OC().Component(tr).Type().State(), samplingInterval)
			defer streamType.Close()
			streamDescription := samplestream.New(t, dut, gnmi.OC().Component(tr).Description().State(), samplingInterval)
			defer streamDescription.Close()
			streamMfgName := samplestream.New(t, dut, gnmi.OC().Component(tr).MfgName().State(), samplingInterval)
			defer streamMfgName.Close()
			streamMfgDate := samplestream.New(t, dut, gnmi.OC().Component(tr).MfgDate().State(), samplingInterval)
			defer streamMfgDate.Close()
			streamHwVersion := samplestream.New(t, dut, gnmi.OC().Component(tr).HardwareVersion().State(), samplingInterval)
			defer streamHwVersion.Close()
			streamFirmwareVersion := samplestream.New(t, dut, gnmi.OC().Component(tr).FirmwareVersion().State(), samplingInterval)
			defer streamFirmwareVersion.Close()

			allStream := map[string]*samplestream.SampleStream[string]{
				"serialNo":        streamSerialNo,
				"partNo":          streamPartNo,
				"description":     streamDescription,
				"mfgName":         streamMfgName,
				"mfgDate":         streamMfgDate,
				"hwVersion":       streamHwVersion,
				"firmwareVersion": streamFirmwareVersion,
			}
			validateStreamOutput(t, dut, allStream)

			opInst := samplestream.New(t, dut, gnmi.OC().Component(tr).OpticalChannel().OutputPower().Instant().State(), samplingInterval)
			defer opInst.Close()
			if opInstN := opInst.Next(t); opInstN != nil {
				if _, ok := opInstN.Val(); ok {
					t.Fatalf("streaming /components/component/optical-channel/state/output-power/instant is not expected to be reported")
				}
			}

			opAvg := samplestream.New(t, dut, gnmi.OC().Component(tr).OpticalChannel().OutputPower().Avg().State(), samplingInterval)
			defer opAvg.Close()
			if opAvgN := opAvg.Next(t); opAvgN != nil {
				if _, ok := opAvgN.Val(); ok {
					t.Fatalf("streaming /components/component/optical-channel/state/output-power/avg is not expected to be reported")
				}
			}

			opMin := samplestream.New(t, dut, gnmi.OC().Component(tr).OpticalChannel().OutputPower().Min().State(), samplingInterval)
			defer opMin.Close()
			if opMinN := opMin.Next(t); opMinN != nil {
				if _, ok := opMinN.Val(); ok {
					t.Fatalf("streaming /components/component/optical-channel/state/output-power/min is not expected to be reported")
				}
			}

			opMax := samplestream.New(t, dut, gnmi.OC().Component(tr).OpticalChannel().OutputPower().Max().State(), samplingInterval)
			defer opMax.Close()
			if opMaxN := opMax.Next(t); opMaxN != nil {
				if _, ok := opMaxN.Val(); ok {
					t.Fatalf("streaming /components/component/optical-channel/state/output-power/max is not expected to be reported")
				}
			}

			gnmi.Update(t, dut, gnmi.OC().Interface(dp.Name()).Enabled().Config(), bool(true))

			powerStreamMap := map[string]*samplestream.SampleStream[float64]{
				"inst": opInst,
				"avg":  opAvg,
				"min":  opMin,
				"max":  opMax,
			}

			validateOutputPower(t, dut, powerStreamMap)

			// Derive transceiver names from ports.
			component := gnmi.OC().Component(tr)

			outputPower := gnmi.Get(t, dut, component.OpticalChannel().TargetOutputPower().State())
			if math.Abs(float64(outputPower)-float64(targetOutputPowerdBm)) > targetOutputPowerTolerancedBm {
				t.Fatalf("Output power is not within expected tolerance, got: %v want: %v tolerance: %v", outputPower, targetOutputPowerdBm, targetOutputPowerTolerancedBm)
			}

			frequency := gnmi.Get(t, dut, component.OpticalChannel().Frequency().State())
			if math.Abs(float64(frequency)-float64(targetFrequencyHz)) > targetFrequencyToleranceHz {
				t.Fatalf("Frequency is not within expected tolerance, got: %v want: %v tolerance: %v", frequency, targetFrequencyHz, targetFrequencyToleranceHz)
			}
		})
	}
}
