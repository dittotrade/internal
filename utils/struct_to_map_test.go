package utils

import (
	"net/url"
	"reflect"
	"testing"
)

func TestStructToURLValues(t *testing.T) {
	type Settings struct {
		SlaveID         string `json:"slave_id,omitempty"`          // slave account_id or not defined if you define a group_id
		GroupID         string `json:"group_id,omitempty"`          // group id or not defined if you define a slave_id
		MasterID        string `json:"master_id,omitempty"`         // master account_id or not defined if you want it to be applied to all your Masters
		RiskFactorValue string `json:"risk_factor_value,omitempty"` // RiskFactor value or not defined
		RiskFactorType  string `json:"risk_factor_type,omitempty"`  // RiskFactor type or not defined
		MaxOrderSize    string `json:"max_order_size,omitempty"`    // max order size in lot/volume or not defined
		CopierStatus    string `json:"copier_status,omitempty"`     // optional
	}

	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want url.Values
	}{
		{
			"test 1",
			args{
				v: Settings{
					MasterID:     "326433",
					SlaveID:      "326441",
					CopierStatus: "1",
				},
			},
			url.Values{
				"master_id":     []string{"326433"},
				"slave_id":      []string{"326441"},
				"copier_status": []string{"1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StructToURLValues(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StructToURLValues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStructToURLValuesUnderscore(t *testing.T) {
	type Settings struct {
		MasterID     string
		SlaveID      string
		CopierStatus string
		OmitThis     string
	}

	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want url.Values
	}{
		{
			"test 1",
			args{
				v: Settings{
					MasterID:     "326433",
					SlaveID:      "326441",
					CopierStatus: "1",
				},
			},
			url.Values{
				"master_id":     []string{"326433"},
				"slave_id":      []string{"326441"},
				"copier_status": []string{"1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StructToURLValues(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StructToURLValues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnderscore(t *testing.T) {
	tests := []struct{ arg, want string }{
		{arg: "ID", want: "id"},
		{arg: "Customer_ID", want: "customer_id"},
		{arg: "Customer_Id", want: "customer_id"},
		{arg: "SToken", want: "s_token"},
		{arg: "HelloWorld", want: "hello_world"},
		{arg: "Hello_World", want: "hello_world"},
		{arg: "Hello2_World", want: "hello2_world"},
		{arg: "Hello_2World", want: "hello_2_world"},
		{arg: "Hello22", want: "hello22"},
		{arg: "Fname", want: "fname"},
		{arg: "HelloB", want: "hello_b"},
		{arg: "H1a", want: "h1_a"},
		{arg: "h1a", want: "h1_a"},
		{arg: "H1A", want: "h1_a"},
	}
	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			if got := Underscore(tt.arg); tt.want != got {
				t.Errorf("Underscore(%v) = %v, want %v", tt.arg, got, tt.want)
			}
		})
	}
}
