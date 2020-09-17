package dynamic

import (
	"github.com/caos/orbos/internal/docu"
	"path"
	"runtime"
)

func GetDocuInfo() (string, []*docu.Version) {
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		return path.Dir(filename), []*docu.Version{
			{
				Struct:   "Desired",
				Version:  "V2",
				SubKinds: nil,
			},
			{
				Struct:   "DesiredV1",
				Version:  "V1",
				SubKinds: nil,
			},
			{
				Struct:   "DesiredV0",
				Version:  "V0",
				SubKinds: nil,
			},
		}
	}
	return "", nil
}
