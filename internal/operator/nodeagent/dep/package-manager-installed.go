package dep

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

func (p *PackageManager) debbasedInstalled() error {
	return p.listAndParse(exec.Command("apt", "list", "--installed"), "Listing...", func(line string) (string, string, error) {
		parts := strings.Split(line, "/")
		if len(parts) < 2 {
			return "", "", errors.Errorf(`Splitting line "%s" by a forward slash failed`, line)
		}

		versionParts := strings.Fields(parts[1])
		if len(versionParts) < 2 {
			return "", "", errors.Errorf(`Splitting "%s" (the part after the forward slash) by empty characters failed`, parts[1])
		}

		return parts[0], versionParts[1], nil
	})
}

func (p *PackageManager) rembasedInstalled() error {
	return p.listAndParse(exec.Command("yum", "list", "installed"), "Installed Packages", func(line string) (string, string, error) {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			return "", "", errors.Errorf(`Splitting line "%s" empty characters failed`, line)
		}
		packageParts := strings.Split(parts[0], ".")
		if len(packageParts) < 1 {
			return "", "", errors.Errorf(`Splitting "%s" (the part before the first empty characters) by a dot failed`, parts[0])
		}

		return packageParts[0], parts[1], nil
	})
}

func (p *PackageManager) listAndParse(listCommand *exec.Cmd, afterLineContaining string, parse func(line string) (string, string, error)) error {

	p.installed = make(map[string]string)
	if p.monitor.IsVerbose() {
		fmt.Println(strings.Join(listCommand.Args, " "))
	}

	stdout, err := listCommand.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "getting stdout pipe for list command failed")
	}
	bufferedReader := bufio.NewReader(stdout)

	err = listCommand.Start()
	if err != nil {
		return errors.Wrap(err, "listing packages failed")
	}

	doParse := false
	for err == nil {
		var line string
		line, err = bufferedReader.ReadString('\n')
		if p.monitor.IsVerbose() {
			fmt.Println(line)
		}

		if strings.Contains(line, afterLineContaining) {
			doParse = true
			continue
		}

		if !doParse {
			continue
		}

		pkg, version, _ := parse(line)
		p.installed[pkg] = version
		p.monitor.WithFields(map[string]interface{}{
			"package": pkg,
			"version": version,
		}).Debug("Found installed package")
	}

	if err == io.EOF {
		err = nil
	}

	if waitErr := listCommand.Wait(); waitErr != nil {
		return errors.Wrap(waitErr, "waiting for list packages command failed")
	}

	return errors.Wrap(err, "reading and parsing installed packages failed")
}
