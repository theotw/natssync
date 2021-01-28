/*
Unit test setup side effects
 */

package unit

import (
	"os"

	"github.com/theotw/natssync/pkg"
)

func init() {
	path, _ := os.Getwd()
	pkg.Config.CertDir = path+"/../../testfiles"
	pkg.Config.CacheMgr = "mem"
	pkg.Config.Keystore = "file"
}
