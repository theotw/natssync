/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

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

	newConfig.CertDir = path + "/../../testfiles"
	newConfig.CacheMgr = "mem"
	newConfig.Keystore = "file"

	pkg.Config = newConfig

}
