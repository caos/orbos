package executables

import (
	"archive/tar"
	"compress/gzip"
	"io"

	"github.com/go-git/go-git/v5/utils/ioutil"

	"github.com/caos/orbos/internal/helpers"
)

func ExtractTarGczFile(packables <-chan PackableTuple, file string) <-chan PackableTuple {
	return deriveFmapExtractOneFile(extractOneFileFunc(file), packables)
}

func ExtractTarGcz(packables <-chan PackableTuple, do func(header *tar.Header, tarReader *tar.Reader, err error) (bool, error)) (err error) {
	for pack := range packables {
		var p *packable
		p, err = pack()
		err = helpers.Concat(err, extractTarGz(pack, do))
		err = helpers.Concat(err, p.data.Close())
	}
	return
}

func ReaderNoopCloser(reader io.Reader) io.ReadCloser {
	return ioutil.NewReadCloser(reader, readCloser(func() error { return nil }))
}

type readCloser func() error

func (r readCloser) Close() error {
	return r()
}

func extractOneFileFunc(file string) func(pt PackableTuple) (rpt PackableTuple) {
	return func(pt PackableTuple) (rpt PackableTuple) {
		p, err := pt()
		if err != nil {
			return pt
		}

		err = extractTarGz(pt, func(header *tar.Header, tarReader *tar.Reader, readErr error) (bool, error) {
			if readErr != nil {
				return false, readErr
			}
			if header.Name == file {
				rpt = deriveTuplePackable(&packable{
					key:  p.key,
					data: ReaderNoopCloser(tarReader),
				}, nil)
				return true, nil
			}
			return false, nil
		})
		if rpt != nil {
			return rpt
		}
		return deriveTuplePackable(p, err)
	}
}

func extractTarGz(pt PackableTuple, do func(*tar.Header, *tar.Reader, error) (bool, error)) (err error) {
	dl, err := pt()
	defer func() {
		dl.data.Close()
	}()
	if err != nil {
		return err
	}

	gzipReader, err := gzip.NewReader(dl.data)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		done, err := do(header, tarReader, err)
		if done || err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

	}
}
