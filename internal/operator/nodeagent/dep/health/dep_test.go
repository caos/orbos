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
ExecStart=/usr/local/bin/health --listen 0.0.0.0:6701/ambassador/v0/check_ready "--protocol" "https" "--ip" "10.172.0.4" "--port" "30443" "--path" "/ambassador/v0/check_ready" "--status" "200" "--proxy=true"
Restart=always
MemoryMax=20M
MemoryLimit=20M
RestartSec=10

[Install]
WantedBy=multi-user.target
`),
		},
		want:  "0.0.0.0:6701/ambassador/v0/check_ready ",
		want1: `--listen 0.0.0.0:6701/ambassador/v0/check_ready "--protocol" "https" "--ip" "10.172.0.4" "--port" "30443" "--path" "/ambassador/v0/check_ready" "--status" "200" "--proxy=true"`,
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
