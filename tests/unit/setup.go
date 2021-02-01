/*
Unit test setup side effects
 */

package unit

import (
	"os"

	"github.com/theotw/natssync/pkg"
)

func init() {
	newConfig := pkg.NewConfiguration()
	path, _ := os.Getwd()

	newConfig.CertDir = path+"/../../testfiles"
	newConfig.CacheMgr = "mem"
	newConfig.Keystore = "file"

	pkg.Config = newConfig
}
