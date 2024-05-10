// Copyright 2019 Richard Hartmann
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package modbus

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/RichiH/modbus_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
)

func TestRegisterMetrics(t *testing.T) {
	t.Run("does not fail", func(t *testing.T) {
		reg := prometheus.NewRegistry()
		moduleName := "my_module"
		metrics := []metric{}

		if err := registerMetrics(reg, moduleName, metrics); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("registers metrics with same name and same label keys", func(t *testing.T) {
		reg := prometheus.NewRegistry()
		moduleName := "my_module"
		metrics := []metric{
			{
				Name: "my_metric",
				Help: "my_help",
				Labels: map[string]string{
					"labelKey1": "labelValueA",
					"labelKey2": "labelValueA",
				},
				Value:      1,
				MetricType: config.MetricTypeCounter,
			},
			{
				Name: "my_metric",
				Help: "my_help",
				Labels: map[string]string{
					"labelKey1": "labelValueB",
					"labelKey2": "labelValueB",
				},
				Value:      2,
				MetricType: config.MetricTypeCounter,
			},
		}

		if err := registerMetrics(reg, moduleName, metrics); err != nil {
			t.Fatal(err)
		}

		metricFamilies, err := reg.Gather()
		if err != nil {
			t.Fatal(err)
		}

		if len(metricFamilies) != 1 {
			t.Fatalf("expected %v but got %v", 1, len(metricFamilies))
		}

		for _, m := range metricFamilies {
			metrics := m.Metric
			if len(metrics) != 2 {
				t.Fatalf("expected %v metrics but got %v", 2, len(metrics))
			}
		}
	})
}

func TestParseModbusData(t *testing.T) {
	offsetZero := 0
	offsetOne := 1

	tests := []struct {
		name          string
		input         func() []byte
		metricDef     func() *config.MetricDef
		expectedValue float64
	}{
		{
			name: "bool, no bit",
			input: func() []byte {
				return []byte{uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:  config.ModbusBool,
					BitOffset: &offsetZero,
				}
			},
			expectedValue: 0,
		},
		{
			name: "bool, first bit",
			input: func() []byte {
				return []byte{uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:  config.ModbusBool,
					BitOffset: &offsetZero,
				}
			},
			expectedValue: 1,
		},
		{
			name: "bool, second bit",
			input: func() []byte {
				return []byte{uint8(2)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:  config.ModbusBool,
					BitOffset: &offsetOne,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int16, default endian (big endian)",
			input: func() []byte {
				return []byte{uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType: config.ModbusInt16,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int16, Big endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusInt16,
					Endianness: config.EndiannessBigEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int16, Little endian",
			input: func() []byte {
				return []byte{uint8(1), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusInt16,
					Endianness: config.EndiannessLittleEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint16, default endian (big endian)",
			input: func() []byte {
				return []byte{uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType: config.ModbusUInt16,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint16, Big endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusUInt16,
					Endianness: config.EndiannessBigEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint16, Little endian",
			input: func() []byte {
				return []byte{uint8(1), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusUInt16,
					Endianness: config.EndiannessLittleEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int32, default endian (big endian)",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType: config.ModbusInt32,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int32, Big endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusInt32,
					Endianness: config.EndiannessBigEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int32, Little endian",
			input: func() []byte {
				return []byte{uint8(1), uint8(0), uint8(0), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusInt32,
					Endianness: config.EndiannessLittleEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int32, Mixed endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(1), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusInt32,
					Endianness: config.EndiannessMixedEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int32, Yolo endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(1), uint8(0), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusInt32,
					Endianness: config.EndiannessYolo,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint32, default endian (big endian)",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType: config.ModbusUInt32,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint32, Big endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusUInt32,
					Endianness: config.EndiannessBigEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint32, Little endian",
			input: func() []byte {
				return []byte{uint8(1), uint8(0), uint8(0), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusUInt32,
					Endianness: config.EndiannessLittleEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint32, Mixed endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(1), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusUInt32,
					Endianness: config.EndiannessMixedEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint32, Yolo endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(1), uint8(0), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusUInt32,
					Endianness: config.EndiannessYolo,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int64, default endian (big endian)",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType: config.ModbusInt64,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int64, Big endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusInt64,
					Endianness: config.EndiannessBigEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int64, Little endian",
			input: func() []byte {
				return []byte{uint8(1), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusInt64,
					Endianness: config.EndiannessLittleEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int64, Mixed endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusInt64,
					Endianness: config.EndiannessMixedEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int64, Yolo endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(1), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusInt64,
					Endianness: config.EndiannessYolo,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint64, default endian (big endian)",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType: config.ModbusUInt64,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint64, Big endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusUInt64,
					Endianness: config.EndiannessBigEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint64, Little endian",
			input: func() []byte {
				return []byte{uint8(1), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusUInt64,
					Endianness: config.EndiannessLittleEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint64, Mixed endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusUInt64,
					Endianness: config.EndiannessMixedEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint64, Yolo endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(1), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusUInt64,
					Endianness: config.EndiannessYolo,
				}
			},
			expectedValue: 1,
		},
	}

	for _, loopTest := range tests {

		test := loopTest

		t.Run(test.name, func(t *testing.T) {
			f, err := parseModbusData(*test.metricDef(), test.input())
			if err != nil {
				t.Fatal(err)
			}

			if f != test.expectedValue {
				t.Fatalf("expected metric value to be %v but got %v", test.expectedValue, f)
			}
		})
	}
}

func TestParseModbusDataInsufficientRegisters(t *testing.T) {
	d := config.MetricDef{
		DataType: config.ModbusInt16,
	}

	_, err := parseModbusData(d, []byte{})

	if err == nil {
		t.Fatal("expected error but got nil")
	}

	switch err.(type) {
	case *InsufficientRegistersError:
	default:
		t.Fatal("expected InsufficientRegistersError")
	}
}

func TestParseModbusDataFloat32(t *testing.T) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, math.Float32bits(32))

	def := config.MetricDef{
		DataType: config.ModbusFloat32,
	}

	floatValue, err := parseModbusData(def, data)
	if err != nil {
		t.Fatal(err)
	}

	if floatValue != 32 {
		t.Fatalf("expected 32 but got %v", floatValue)
	}
}

// TestRegisterMetricTwoMetricsSameName makes sure registerMetrics reuses a
// registered metric in case there is a second one with the same name instead of
// reregistering which would cause an exception.
func TestRegisterMetricTwoMetricsSameName(t *testing.T) {
	reg := prometheus.NewRegistry()
	a := metric{"my_metric", "", map[string]string{}, 1, config.MetricTypeCounter}
	b := metric{"my_metric", "", map[string]string{}, 1, config.MetricTypeCounter}

	err := registerMetrics(reg, "my_module", []metric{a, b})
	if err != nil {
		t.Fatalf("expected no error but got: %v", err)
	}
}

// TestRegisterMetricsRecoverNegativeCounter makes sure the function properly
// recovers from a prometheus client library panic on negative counter changes.
func TestRegisterMetricsRecoverNegativeCounter(t *testing.T) {
	reg := prometheus.NewRegistry()
	a := metric{"my_metric", "", map[string]string{"key1": "value1", "key2": "value2"}, -1, config.MetricTypeCounter}

	err := registerMetrics(reg, "my_module", []metric{a})
	if err == nil {
		t.Fatal("expected an error but got nil")
	}
}


func TestScaleValue(t *testing.T) {
	tests := []struct {
		name   string
		f      *float64
		bias   *float64
		d      float64
		result float64
	}{
		{
			name:   "No scaling or bias",
			f:      nil,
			bias:   nil,
			d:      10.0,
			result: 10.0,
		},
		{
			name:   "Only scaling",
			f:      floatPtr(2.0),
			bias:   nil,
			d:      5.0,
			result: 10.0,
		},
		{
			name:   "Only bias",
			f:      nil,
			bias:   floatPtr(3.0),
			d:      7.0,
			result: 4.0,
		},
		{
			name:   "Scaling and bias",
			f:      floatPtr(2.0),
			bias:   floatPtr(3.0),
			d:      2.0,
			result: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scaleValue(tt.f, tt.bias, tt.d)
			if got != tt.result {
				t.Errorf("got %v, want %v", got, tt.result)
			}
		})
	}
}

func TestApplyTransformations(t *testing.T) {
	tests := []struct {
		name       string
		factor     *float64
		bias       *float64
		expression *string
		d          float64
		want       float64
		wantErr    bool
	}{
		{
			name:    "No transformation",
			factor:  nil,
			bias:    nil,
			d:       10.0,
			want:    10.0,
			wantErr: false,
		},
		{
			name:    "Scaling",
			factor:  floatPtr(2.0),
			bias:    nil,
			d:       10.0,
			want:    20.0,
			wantErr: false,
		},
		{
			name:    "Bias",
			factor:  nil,
			bias:    floatPtr(5.0),
			d:       10.0,
			want:    5.0,
			wantErr: false,
		},
		{
			name:       "Expression",
			factor:     nil,
			bias:       nil,
			expression: stringPtr("x**2 + 10"),
			d:          5.0,
			want:       35.0,
			wantErr:    false,
		},
		{
			name:       "Scaling with Expression",
			factor:     floatPtr(2.0),
			bias:       floatPtr(5.0),
			expression: stringPtr("x*2 + 10"),
			d:          5.0,
			want:       20,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyTransformations(tt.factor, tt.bias, tt.expression, tt.d)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyTransformations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("applyTransformations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}