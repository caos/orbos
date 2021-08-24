package nginx_test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/caos/orbos/internal/operator/nodeagent/dep/nginx"

	"github.com/caos/orbos/internal/operator/common"
)

var systemdPackageConfigEntries = map[string]string{
	"Systemd[Unit]SomeConfig":     "hodor",
	"Systemd[Service]LimitNOFILE": "8192",
	"Systemd[Install]SomeConfig":  "hodor",
}

var currentManaged = fmt.Sprintf(`
[Unit]
%s
SomeConfig=hodor
Description=nginx - high performance web server

[Service]
%s
LimitNOFILE=8192
Type=forking

[Install]
%s
SomeConfig=hodor
Bla=blubb
`, nginx.LineAddedComment, nginx.LineAddedComment, nginx.LineAddedComment)

var currentUnmanaged = `
[Unit]
Description=nginx - high performance web server

[Service]
Type=forking

[Install]
Bla=blubb
`

var currentBackwardsCompatibilityComment = fmt.Sprintf(`
[Unit]
Description=nginx - high performance web server

[Service]
LimitNOFILE=8192 %s
Type=forking

[Install]
Bla=blubb
`, nginx.CleanupLine)

const currentBackwardsCompatibilityManual = `
[Unit]
Description=nginx - high performance web server

[Service]
LimitNOFILE=8192
Type=forking

[Install]
Bla=blubb
`

func Test_CurrentSystemdEntries(t *testing.T) {

	type args struct {
		currentFile string
		entries     map[string]string
	}
	tests := []struct {
		name string
		args args
	}{{
		name: "Managed systemd entries are listed",
		args: args{
			currentFile: currentManaged,
			entries:     systemdPackageConfigEntries,
		},
	}, {
		name: "If nothing is managed, no entries should appear in the current state",
		args: args{
			currentFile: currentUnmanaged,
		},
	}, {
		name: "For backwards compatibility, lines with syntactically wrong comments should be removed, so reconciling should be triggered",
		args: args{
			currentFile: currentBackwardsCompatibilityComment,
			entries: map[string]string{
				"ensuresystemdconf": "yes",
			},
		},
	}, {
		name: "If entry is added manually, no entries should appear in the current state",
		args: args{
			currentFile: currentUnmanaged,
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := &common.Package{
				Config: map[string]string{
					"nginx.conf": "cfg here",
				},
			}
			nginx.CurrentSystemdEntries(strings.NewReader(tt.args.currentFile), pkg)

			expectedLen := len(tt.args.entries) + 1
			actualLen := len(pkg.Config)
			if actualLen != expectedLen {
				t.Errorf("pkg.Config has length %d instead of %d", actualLen, expectedLen)
			}

			for expectedKey, expectedValue := range tt.args.entries {
				actualValue, ok := pkg.Config[expectedKey]
				if !ok {
					t.Errorf("expected key %s but only got keys %s", expectedKey, keys(pkg.Config))
				}

				if actualValue != expectedValue {
					t.Errorf("expected key %s to have value %s but got %s", expectedKey, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestUpdateNginxService(t *testing.T) {
	type args struct {
		before string
		cfg    map[string]string
		after  string
	}

	type test struct {
		name string
		args args
	}
	tests := []test{{
		name: "Desired entries are placed in the correct sections and be preceded by the comment line",
		args: args{
			before: currentUnmanaged,
			cfg:    systemdPackageConfigEntries,
			after:  currentManaged,
		},
	}, {
		name: "Ensuring is idempotent",
		args: args{
			before: currentManaged,
			cfg:    systemdPackageConfigEntries,
			after:  currentManaged,
		},
	}, {
		name: "Syntactically wrong lines with comments are removed",
		args: args{
			before: currentBackwardsCompatibilityComment,
			cfg:    systemdPackageConfigEntries,
			after:  currentManaged,
		},
	}, {
		name: "Keys are not duplicated",
		args: args{
			before: currentBackwardsCompatibilityManual,
			cfg:    systemdPackageConfigEntries,
			after:  currentManaged,
		},
	}, {
		name: "Only keys with systemd convention syntax are used",
		args: args{
			before: currentBackwardsCompatibilityManual,
			cfg: func() map[string]string {
				newMap := map[string]string{"random": "key"}
				for k, v := range systemdPackageConfigEntries {
					newMap[k] = v
				}
				return newMap
			}(),
			after: currentManaged,
		},
	}}

	testEnsureCase := func(tt test, t *testing.T) {
		file, err := ioutil.TempFile("", "orbos-test-nginx-service-*.service")
		if err != nil {
			t.Error(err)
			return
		}
		defer file.Close()

		if _, err = file.WriteString(tt.args.before); err != nil {
			t.Error(err)
			return
		}

		if err := nginx.UpdateSystemdUnitFile(file.Name(), tt.args.cfg); err != nil {
			t.Errorf("UpdateSystemdUnitFile() error = %v", err)
		}

		actualBytes, err := ioutil.ReadFile(file.Name())
		if err != nil {
			t.Fatal(err)
		}

		actual := string(actualBytes)
		if actual != tt.args.after {
			t.Errorf("UpdateSystemdUnitFile() manipulated file has content \n\n%s\n\ninstead of\n\n%s", actual, tt.args.after)
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testEnsureCase(tt, t)
		})
	}
}

func keys(m map[string]string) string {
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	return strings.Join(ks, ", ")
}
