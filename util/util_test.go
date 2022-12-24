package util

import (
	"reflect"
	"testing"
)

func TestGetPathDataFromCSV(t *testing.T) {
	type args struct {
		csvPath string
	}
	tests := []struct {
		name string
		args args
		want [][3]float64
	}{
		{"pathImport", args{"../test_data/path.csv"}, [][3]float64{{4.656281015533591271e+05, 1.059697990362517303e+05, 1000.0000000000000000000000}, {4.697615370872759959e+05, 1.059752773382947780e+05, 1000.0000000000000000000000}, {4.747092632187826675e+05, 1.048383400561068265e+05, 1000.0000000000000000000000}, {4.761507699336261721e+05, 1.021576807585244824e+05, 1000.0000000000000000000000}, {4.751697623862638138e+05, 9.804030687632088666e+04, 1000.0000000000000000000000}, {4.547814635303586256e+05, 9.384979078477300936e+04, 1000.0000000000000000000000}, {4.502560147261911770e+05, 9.080751501439491403e+04, 1000.0000000000000000000000}, {4.495445221333308727e+05, 9.042598606925050262e+04, 1000.0000000000000000000000}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPathDataFromCSV(tt.args.csvPath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPathDataFromCSV() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPathLength(t *testing.T) {
	type args struct {
		path [][3]float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{"pathLength", args{GetPathDataFromCSV("../test_data/path.csv")}, 43561.22251758873},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPathLength(tt.args.path); got != tt.want {
				t.Errorf("GetPathLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckPathExists(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"existingPath", args{"../test_data/path.csv"}, "../test_data/path.csv"},
		{"existingPath", args{"../test_data/alts.csv"}, "../test_data/alts.csv"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckPathExists(tt.args.path); got != tt.want {
				t.Errorf("CheckPathExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestCheckSliceLen(t *testing.T) {
// 	type args struct {
// 		slice          []T
// 		requiredLength int
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want []T
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := CheckSliceLen(tt.args.slice, tt.args.requiredLength); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("CheckSliceLen() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
