package swap_test

import (
	"io/ioutil"
	"testing"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/core/operator/nodeagent/edge/dep/swap"
)

func TestEnsure(t *testing.T) {

	before := `#
# /etc/fstab
# Created by anaconda on Wed Sep  4 04:47:43 2019
#
# Accessible filesystems, by reference, are maintained under '/dev/disk'
# See man pages fstab(5), findfs(8), mount(8) and/or blkid(8) for more info
#
/dev/mapper/centos-root /                       xfs     defaults        0 0
UUID=2be0c56f-32f4-4155-af7a-1db773af5ff0 /boot                   xfs     defaults        0 0
/dev/mapper/centos-home /home                   xfs     defaults        0 0
/dev/mapper/centos-swap swap                    swap    defaults        0 0
`

	after := `#
# /etc/fstab
# Created by anaconda on Wed Sep  4 04:47:43 2019
#
# Accessible filesystems, by reference, are maintained under '/dev/disk'
# See man pages fstab(5), findfs(8), mount(8) and/or blkid(8) for more info
#
/dev/mapper/centos-root /                       xfs     defaults        0 0
UUID=2be0c56f-32f4-4155-af7a-1db773af5ff0 /boot                   xfs     defaults        0 0
/dev/mapper/centos-home /home                   xfs     defaults        0 0
#/dev/mapper/centos-swap swap                    swap    defaults        0 0
`

	testFile, err := testFile(before)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = swap.New(testFile).Ensure(common.Package{Version: "enabled"}, common.Package{Version: "disabled"}); err != nil {
		t.Fatal(err)
	}

	changed, err := ioutil.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(changed) != after {
		t.Fatalf("expected:\n%s\n\ninstead:\n%s", after, string(changed))
	}
}

func testFile(content string) (string, error) {
	tmpFile, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()
	_, err = tmpFile.WriteString(content)
	return tmpFile.Name(), err
}
