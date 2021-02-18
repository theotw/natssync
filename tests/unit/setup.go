/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

/*
Unit test setup side effects
*/

package unit

import (
	"io/ioutil"
	"os"

	"github.com/theotw/natssync/pkg"
)

func init() {
	newConfig := pkg.NewConfiguration()
	parentDir := os.TempDir()
	keystoreDir, _ := ioutil.TempDir(parentDir, "keystoretest")
	pkg.Config.CertDir = keystoreDir
	newConfig.CacheMgr = "mem"
	newConfig.Keystore = "file"

	pkg.Config = newConfig

}
