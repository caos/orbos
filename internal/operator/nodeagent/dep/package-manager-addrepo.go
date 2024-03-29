package dep

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func (p *PackageManager) rembasedAdd(repo *Repository) error {

	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()
	outBuf := new(bytes.Buffer)
	defer errBuf.Reset()

	cmd := exec.Command("yum-config-manager", "--add-repo", repo.Repository)
	cmd.Stderr = errBuf
	cmd.Stdout = outBuf
	err := cmd.Run()

	out := outBuf.String()
	if p.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		fmt.Println(out)
	}

	if err != nil && !strings.Contains(out, fmt.Sprintf("Cannot add repo from %s as is a duplicate of an existing repo", repo.Repository)) {
		return fmt.Errorf("adding yum repository %s failed with stderr %s: %w", repo.Repository, out, err)
	}

	return nil
}

func (p *PackageManager) debbasedAdd(repo *Repository) error {

	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()

	resp, err := http.Get(repo.KeyURL)
	if err != nil {
		return fmt.Errorf("getting key from url %s failed: %w", repo.KeyURL, err)
	}
	defer resp.Body.Close()
	cmd := exec.Command("apt-key", "add", "-")
	cmd.Stdin = resp.Body
	cmd.Stderr = errBuf

	if p.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("adding key failed with stderr %s: %w", errBuf.String(), err)
	}
	errBuf.Reset()
	p.monitor.WithFields(map[string]interface{}{
		"url": repo.KeyURL,
	}).Debug("Added repository key from url")

	if repo.KeyFingerprint != "" {
		buf := new(bytes.Buffer)
		defer buf.Reset()

		cmd := exec.Command("apt-key", "fingerprint", repo.KeyFingerprint)
		cmd.Stdout = buf
		cmd.Stderr = errBuf
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("verifying fingerprint %s failed with stderr %s: %w", repo.KeyFingerprint, errBuf.String(), err)
		}

		if p.monitor.IsVerbose() {
			fmt.Println(strings.Join(cmd.Args, " "))
		}

		errBuf.Reset()
		p.monitor.WithFields(map[string]interface{}{
			"url":         repo.KeyURL,
			"fingerprint": repo.KeyFingerprint,
		}).Debug("Checked fingerprint")
		found := false
		for {
			line, err := buf.ReadString('\n')
			if p.monitor.IsVerbose() {
				fmt.Println(line)
			}

			if strings.HasPrefix(line, "uid") {
				p.monitor.WithFields(map[string]interface{}{
					"uid": strings.TrimSpace(strings.TrimPrefix(line, "uid")),
				}).Debug("Added and verified repository key")
				found = true
				break
			}
			if line == "\n" || err != nil {
				break
			}
		}
		if !found {
			return fmt.Errorf("no key with fingerprint %s found", repo.KeyFingerprint)
		}
	}

	cmd = exec.Command("add-apt-repository", "-y", repo.Repository)
	cmd.Stderr = errBuf

	if p.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("adding repository %s failed with stderr %s: %w", repo.Repository, errBuf.String(), err)
	}
	errBuf.Reset()
	p.monitor.WithFields(map[string]interface{}{
		"repository": repo.Repository,
	}).Debug("Added repository")

	cmd = exec.Command("apt-get", strings.Fields("--assume-yes --allow-downgrades update")...)
	cmd.Stderr = errBuf
	if p.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("updating indices failed with stderr %s: %w", errBuf.String(), err)
	}
	errBuf.Reset()
	p.monitor.Debug("Updated index")
	return nil
}
