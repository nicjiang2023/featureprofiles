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

package zr_fec_uncorrectable_frames_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/openconfig/featureprofiles/internal/fptest"
	"github.com/openconfig/featureprofiles/internal/samplestream"
	"github.com/openconfig/ondatra"
	"github.com/openconfig/ondatra/gnmi"
	"github.com/openconfig/ondatra/gnmi/oc"
)

const (
	sampleInterval       = 10 * time.Second
	targetOutputPowerdBm = -10
	targetFrequencyHz    = 193100000
	intUpdateTime        = 2 * time.Minute
)

func TestMain(m *testing.M) {
	fptest.RunTests(m)
}

func validateFecUncorrectableBlocks(t *testing.T, dut *ondatra.DUTDevice, stream *samplestream.SampleStream[uint64]) {
	fecStream := stream.Next(t)
	if fecStream == nil {
		t.Fatalf("Fec Uncorrectable Blocks was not streamed in the most recent subscription interval")
	}
	fec, ok := fecStream.Val()
	if !ok {
		t.Fatalf("Error capturing streaming Fec value")
	}
	if reflect.TypeOf(fec).Kind() != reflect.Int64 {
		t.Fatalf("fec value is not type int64")
	}
	if fec != 0 {
		t.Fatalf("Got FecUncorrectableBlocks got %d, want 0", fec)
	}
}

func TestZrUncorrectableFrames(t *testing.T) {
	dut := ondatra.DUT(t, "dut")

	for _, port := range []string{"port1", "port2"} {
		t.Run(fmt.Sprintf("Port:%s", port), func(t *testing.T) {
			dp := dut.Port(t, "port1")
			gnmi.Await(t, dut, gnmi.OC().Interface(dp.Name()).OperStatus().State(), intUpdateTime, oc.Interface_OperStatus_UP)

			// Derive transceiver names from ports.
			tr := gnmi.Get(t, dut, gnmi.OC().Interface(dp.Name()).Transceiver().State())
			component := gnmi.OC().Component(tr)

			outputPower := gnmi.Get(t, dut, component.OpticalChannel().TargetOutputPower().State())
			if outputPower != targetOutputPowerdBm {
				t.Fatalf("Output power does not match target output power, got: %v want :%v", outputPower, targetOutputPowerdBm)
			}

			frequency := gnmi.Get(t, dut, component.OpticalChannel().Frequency().State())
			if frequency != targetFrequencyHz {
				t.Fatalf("Frequency does not match target frequency, got: %v want :%v", frequency, targetFrequencyHz)
			}

			streamFec := samplestream.New(t, dut, gnmi.OC().TerminalDevice().Channel(0).Otn().FecUncorrectableBlocks().State(), sampleInterval)
			defer streamFec.Close()
			validateFecUncorrectableBlocks(t, dut, streamFec)

			// Toggle interface enabled
			gnmi.Update(t, dut, gnmi.OC().Interface(dp.Name()).Enabled().Config(), bool(false))
			gnmi.Update(t, dut, gnmi.OC().Interface(dp.Name()).Enabled().Config(), bool(true))

			validateFecUncorrectableBlocks(t, dut, streamFec)
		})
	}
}
