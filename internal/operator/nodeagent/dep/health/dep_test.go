package health

import "testing"

func Test_extractArguments(t *testing.T) {
	type args struct {
		content []byte
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{{
		name: "It should extract arguments",
		args: args{
			content: []byte(`
[Unit]
Description=Healthchecks Proxy
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/health --http 0.0.0.0:6702/healthz "200@https://34.78.197.0:6666/healthz"
Restart=always
MemoryMax=20M
MemoryLimit=20M
RestartSec=10

[Install]
WantedBy=multi-user.target
`),
		},
		want:  "0.0.0.0:6702/healthz",
		want1: `"200@https://34.78.197.0:6666/healthz"`,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := extractArguments(tt.args.content)
			if got != tt.want {
				t.Errorf("extractArguments() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("extractArguments() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
