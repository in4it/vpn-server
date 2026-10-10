package network

import (
	"strings"
	"testing"
)

const procNetRouteHeader = "Iface\tDestination\tGateway \tFlags\tRefCnt\tUse\tMetric\tMask\t\tMTU\tWindow\tIRTT\n"

func TestParseDefaultGwInterface(t *testing.T) {
	tests := []struct {
		name    string
		routes  string
		want    string
		wantErr bool
	}{
		{
			name: "default route is not the first route",
			routes: procNetRouteHeader +
				"docker0\t000011AC\t00000000\t0001\t0\t0\t0\t0000FFFF\t0\t0\t0\n" +
				"ens4\t0100000A\t00000000\t0005\t0\t0\t100\tFFFFFFFF\t0\t0\t0\n" +
				"ens5\t00000000\t0100000A\t0003\t0\t0\t100\t00000000\t0\t0\t0\n",
			want: "ens5",
		},
		{
			name: "default route first",
			routes: procNetRouteHeader +
				"eth0\t00000000\t0101A8C0\t0003\t0\t0\t0\t00000000\t0\t0\t0\n" +
				"eth0\t0001A8C0\t00000000\t0001\t0\t0\t0\t00FFFFFF\t0\t0\t0\n",
			want: "eth0",
		},
		{
			name: "lowest metric wins",
			routes: procNetRouteHeader +
				"wlan0\t00000000\t0101A8C0\t0003\t0\t0\t600\t00000000\t0\t0\t0\n" +
				"eth0\t00000000\t0101A8C0\t0003\t0\t0\t100\t00000000\t0\t0\t0\n",
			want: "eth0",
		},
		{
			name: "no default route",
			routes: procNetRouteHeader +
				"ens4\t0100000A\t00000000\t0001\t0\t0\t100\tFFFFFFFF\t0\t0\t0\n",
			wantErr: true,
		},
		{
			name:    "only header",
			routes:  procNetRouteHeader,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDefaultGwInterface(strings.NewReader(tt.routes))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got interface %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
