package api

import (
	"testing"

	"github.com/edgedelta/updater/core"
	"github.com/google/go-cmp/cmp"
)

func TestConstructURLWithParams(t *testing.T) {
	tests := []struct {
		base    string
		params  *core.ParamConf
		ctxVars map[string]string
		wantURL string
	}{
		{
			base:    "https://example.org/some/test-endpoint",
			params:  nil,
			wantURL: "https://example.org/some/test-endpoint",
		},
		{
			base: "https://example.org/some/test-endpoint",
			params: &core.ParamConf{
				QueryParams: map[string]string{
					"name":          "JohnDoe",
					"123":           "param 2!",
					"query__string": "Go programming language",
				},
			},
			wantURL: "https://example.org/some/test-endpoint?123=param+2%21&name=JohnDoe&query__string=Go+programming+language",
		},
		{
			base: "https://example.org/some/test-endpoint",
			params: &core.ParamConf{
				QueryParams: map[string]string{
					"name":          "JohnDoe",
					"123":           "param 2!",
					"query__string": "Go programming language",
					"var-1":         `{{ index .Vars "var_1" }}`,
					"var-2":         `{{ index .Vars "var 2" }}`,
				},
			},
			ctxVars: map[string]string{
				"var_1": "Contextual var - 1",
				"var 2": "  ctx var, 2",
			},
			wantURL: "https://example.org/some/test-endpoint?123=param+2%21&name=JohnDoe&query__string=Go+programming+language&var-1=Contextual+var+-+1&var-2=++ctx+var%2C+2",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.base, func(t *testing.T) {
			got, err := constructURLWithParams(tc.base, tc.params, tc.ctxVars)
			if err != nil {
				t.Fatalf("constructURLWithParams failed, err: %v", err)
			}
			if diff := cmp.Diff(tc.wantURL, got); diff != "" {
				t.Errorf("Constructed URL mismatch, want %s, got %s instead", tc.wantURL, got)
			}
		})
	}
}
