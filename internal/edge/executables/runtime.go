//go:generate goderive .

package executables

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/caos/orbiter/internal/core/helpers"
	"github.com/pkg/errors"
)

var executables map[string][]byte

var populate = func() {}

func Populate() {
	populate()
}

func PreBuilt(mainDir string) ([]byte, error) {
	executable, ok := executables[mainDir]
	if !ok {
		return nil, errors.Errorf("%s was notnot prebuilt", mainDir)
	}
	return executable, nil
}

func PreBuild(builds <-chan BuiltTuple) (err error) {
	sp := selfPath()
	tmpFile := filepath.Join(sp, "prebuilt.tmp")
	outFile := filepath.Join(sp, "prebuilt.go")
	prebuilt, err := os.Create(tmpFile)
	if err != nil {
		return errors.Wrapf(err, "creating %s failed", tmpFile)
	}
	defer func() {
		if err != nil {
			os.Remove(tmpFile)
		}
		err = os.Rename(tmpFile, outFile)
	}()

	if _, err = prebuilt.WriteString(`package executables

func init() {
	populate = func(){
		executables = map[string][]byte{`); err != nil {
		return err
	}

	for pt := range deriveFmapPack(pack, builds) {
		bin, packed, packErr := pt()
		err = helpers.Concat(err, packErr)
		if err != nil {
			continue
		}

		if _, err = prebuilt.WriteString(fmt.Sprintf(`
		"%s": unpack("%s"),`, filepath.Base(bin.MainDir), *packed)); err != nil {
			continue
		}
	}

	if err != nil {
		os.Remove(prebuilt.Name())
		return err
	}

	_, err = prebuilt.WriteString(`
		}
	}
}
`)
	return err
}

func packedTupleFunc(bin Bin) func(*string, error) packedTuple {
	return func(packed *string, err error) packedTuple {
		return deriveTuplePacked(bin, packed, err)
	}
}

type packedTuple func() (Bin, *string, error)

func pack(built BuiltTuple) packedTuple {

	bin, err := built()
	packedTuple := packedTupleFunc(bin)

	defer func() {
		os.Remove(bin.OutDir)
	}()

	if err != nil {
		return packedTuple(nil, err)
	}

	executable, err := os.Open(bin.OutDir)
	if err != nil {
		return packedTuple(nil, err)
	}

	var gzipBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&gzipBuffer)
	_, err = io.Copy(gzipWriter, executable)
	if err != nil {
		return packedTuple(nil, errors.Wrap(err, "gzipping failed"))
	}

	if err := gzipWriter.Close(); err != nil {
		return packedTuple(nil, errors.Wrap(err, "closing gzip writer failed"))
	}

	packed := base64.StdEncoding.EncodeToString(gzipBuffer.Bytes())

	return packedTuple(&packed, nil)
}

func unpack(executable string) []byte {
	gzipNodeAgent, err := base64.StdEncoding.DecodeString(executable)
	if err != nil {
		panic(errors.Wrap(err, "decoding node agent from base64 failed"))
	}
	gzipReader, err := gzip.NewReader(bytes.NewReader(gzipNodeAgent))
	if err != nil {
		panic(errors.Wrap(err, "ungzipping node agent failed"))
	}
	unpacked, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		panic(errors.Wrap(err, "reading unpacked node agent failed"))
	}
	return unpacked
}
