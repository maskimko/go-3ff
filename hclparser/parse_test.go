package hclparser

import (
	"testing"
)

func TestGetTfResourcesByPath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "Smoke test",
			args: args{
				path: "/home/maskimko/Work/Wix/Terraform/rg/generator/output/42",
			},
			wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTfResourcesByPath(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTfResourcesByPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil || len(got) == 0 {
				t.Errorf("GetTfResourcesByPath() got = %v, want not empty slice of terraform resource names", got)
			}
		})
	}
}
