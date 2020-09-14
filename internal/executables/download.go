package executables

import (
	"fmt"
	"net/http"
)

type Downloadable struct {
	Key    string
	URL    string
	Client *http.Client
}

func Download(downloadable ...Downloadable) <-chan PackableTuple {
	return deriveFmapDownload(download, toDownloadableChan(downloadable))
}

func download(dl Downloadable) (pt PackableTuple) {
	defer func() {
		if _, err := pt(); err != nil {
			fmt.Printf("Downloading %s from %s failed\n", dl.Key, dl.URL)
			return
		}
		fmt.Printf("Successfully downloaded %s from %s\n", dl.Key, dl.URL)
	}()

	resp, err := dl.Client.Get(dl.URL)
	if err != nil {
		resp.Body.Close()
		return deriveTuplePackable(nil, err)
	}

	return deriveTuplePackable(&packable{
		key:  dl.Key,
		data: resp.Body,
	}, nil)
}

func toDownloadableChan(downloadable []Downloadable) <-chan Downloadable {
	dlChan := make(chan Downloadable, 0)
	go func() {
		for _, bin := range downloadable {
			dlChan <- bin
		}
		close(dlChan)
	}()
	return dlChan
}
