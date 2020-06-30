package gce

import (
	"archive/tar"
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/caos/orbos/internal/executables"
)

func ensureGcloud(ctx *context) error {
	buf := new(bytes.Buffer)
	cmd := exec.Command(gcloudBin(), "version")
	cmd.Stdout = buf
	if err := cmd.Run(); err != nil {
		return refresh(ctx)
	}

	if !strings.Contains(buf.String(), "Google Cloud SDK 293.0.0") {
		return refresh(ctx)
	}
	return nil
}

var sdkDirCache string

func sdkDir() string {
	if sdkDirCache != "" {
		return sdkDirCache
	}
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	sdkDirCache = filepath.Join(home, ".orb", "gcloud")
	return sdkDirCache
}

var gcloudBinCache string

func gcloudBin() string {
	if gcloudBinCache != "" {
		return gcloudBinCache
	}
	gcloudBinCache = filepath.Join(sdkDir(), "google-cloud-sdk", "bin", "gcloud")
	return gcloudBinCache
}

func gcloudSession(jsonkey string, do func(binary string) error) error {
	gcloud := gcloudBin()
	listBuf := new(bytes.Buffer)
	defer resetBuffer(listBuf)
	cmd := exec.Command(gcloud, "config", "configurations", "list")
	cmd.Stdout = listBuf

	if err := cmd.Run(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(listBuf)
	reactivate := ""
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if fields[1] == "True" {
			reactivate = fields[0]
			break
		}
	}

	file, err := ioutil.TempFile("", "orbiter-gce-key")
	defer os.Remove(file.Name())
	if err != nil {
		return err
	}

	_, err = file.WriteString(jsonkey)
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	cmd = exec.Command(gcloudBin(), "auth", "activate-service-account", "--key-file", file.Name())
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := do(gcloud); err != nil {
		return err
	}
	if reactivate != "" {
		cmd := exec.Command(gcloud, "config", "configurations", "activate", reactivate)
		return cmd.Run()
	}
	return nil
}

func refresh(ctx *context) (err error) {

	ctx.monitor.Debug("Refreshing gcloud")
	defer func() {
		if err != nil {
			ctx.monitor.Info("gcloud refreshed")
		}
	}()

	dir := sdkDir()

	if err := os.RemoveAll(dir); err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	client, err := storage.NewClient(ctx.ctx, *ctx.auth)
	if err != nil {
		return err
	}
	gcloud, err := client.Bucket("cloud-sdk-release").Object("google-cloud-sdk-293.0.0-linux-x86_64.tar.gz").NewReader(ctx.ctx)
	if err != nil {
		return err
	}

	packable := make(chan executables.PackableTuple)
	go func() {
		packable <- executables.NewPackableTuple("gcloud", executables.ReaderNoopCloser(gcloud))
		close(packable)
	}()

	return executables.ExtractTarGcz(packable, func(header *tar.Header, reader *tar.Reader, err error) (b bool, err2 error) {

		if err == io.EOF {
			return true, nil
		}

		name := filepath.Join(dir, header.Name)
		if strings.Contains(name, "..") {
			return false, nil
		}

		switch header.Typeflag {
		case tar.TypeDir:
			return false, os.MkdirAll(name, 0700)
		case tar.TypeReg:
			file, err := os.Create(name)
			defer func(f *os.File) {
				f.Close()
			}(file)
			if err != nil {
				return false, err
			}

			if err := file.Chmod(header.FileInfo().Mode()); err != nil {
				return false, err
			}

			if _, err := io.Copy(file, reader); err != nil {
				return false, err
			}
			file.Close()
			return false, nil
		default:
			return false, nil
		}
	})
}
