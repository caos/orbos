package main

import (
	"bytes"
	"io"
	"os"
)

func initORBITERTest(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) (err error) {

	print, err := orbctl()
	if err != nil {
		return err
	}

	print.Args = append(print.Args, "file", "print", "orbiter-e2e.yml")
	print.Stderr = os.Stderr

	overwrite, err := orbctl()
	if err != nil {
		return err
	}

	overwrite.Args = append(overwrite.Args, "file", "overwrite", "orbiter.yml", "--stdin")
	overwrite.Stderr = os.Stderr

	r, w := io.Pipe()
	print.Stdout = w
	overwrite.Stdin = r

	var overwriteStdout bytes.Buffer
	overwrite.Stdout = &overwriteStdout

	if err := print.Start(); err != nil {
		return err
	}
	if err := overwrite.Start(); err != nil {
		return err
	}
	if err := print.Wait(); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	if err := overwrite.Wait(); err != nil {
		return err
	}
	_, err = io.Copy(os.Stdout, &overwriteStdout)
	return err
}
