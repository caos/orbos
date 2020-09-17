package main

import (
	"fmt"
	"github.com/caos/documentation/pkg/docu"
	"github.com/caos/orbos/internal/operator/boom/api"
	orbiterkinds "github.com/caos/orbos/internal/operator/orbiter/kinds"
	zitadelkinds "github.com/caos/orbos/internal/operator/zitadel/kinds"
	"os"
	"path/filepath"
)

func main() {
	boomInfos := api.GetDocuInfo()
	for _, boomInfo := range boomInfos {
		orbiterDoc := docu.New()

		for _, kind := range boomInfo.Kinds {
			for _, version := range kind.Versions {

				if err := orbiterDoc.Parse(kind.Path, version.Struct); err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}

				folderPath := filepath.Join("docs", "boom", "yml", version.Version)

				if err := os.RemoveAll(folderPath); err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}

				if err := orbiterDoc.GenerateMarkDown(folderPath, version.SubKinds); err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}
			}

		}
	}

	orbiterInfos := orbiterkinds.GetDocuInfo()
	for _, docInfo := range orbiterInfos {
		orbiterDoc := docu.New()

		for _, kind := range docInfo.Kinds {
			for _, version := range kind.Versions {

				if err := orbiterDoc.Parse(kind.Path, version.Struct); err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}

				folderPath := filepath.Join("docs", "orbiter", "yml", docInfo.Name, kind.Kind)

				if err := os.RemoveAll(folderPath); err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}

				if err := orbiterDoc.GenerateMarkDown(folderPath, version.SubKinds); err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}
			}

		}
	}

	zitadelInfos := zitadelkinds.GetDocuInfo()
	for _, docInfo := range zitadelInfos {
		orbiterDoc := docu.New()

		for _, kind := range docInfo.Kinds {
			for _, version := range kind.Versions {

				if err := orbiterDoc.Parse(kind.Path, version.Struct); err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}
				folderPath := filepath.Join("docs", "zitadel", "yml", docInfo.Name, kind.Kind)

				if err := os.RemoveAll(folderPath); err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}

				if err := orbiterDoc.GenerateMarkDown(folderPath, version.SubKinds); err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}
			}

		}
	}
}
