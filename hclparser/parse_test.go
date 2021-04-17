package hclparser

import (
	"os"
	"path"
	"strings"
	"testing"
)

func TestGetTfResourcesByPath(t *testing.T) {
	rp, ok := os.LookupEnv("RUNSH_PATH")
	if ok {
		rpc := strings.Split(rp, ":")
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
					path: path.Join(rpc[0], "/generator/output/42"),
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
	} else {
		t.Skip("RUNSH_PATH is not defined")
	}
}
