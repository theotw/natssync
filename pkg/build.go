/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package pkg

import (
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
)

const buildDateFileLocation = "BUILD_DATE"
var buildDate string

func loadBuildDate() {
	date, err := ioutil.ReadFile(buildDateFileLocation)
	if err != nil {
		log.Errorf("Unable to load build date information: %s", err)
		buildDate = "unknown"
		return
	}
	buildDate = strings.TrimSpace(string(date))
	log.Debugf("Loaded build date %s from file %s", buildDate, buildDateFileLocation)
}

func GetBuildDate() string {
	if buildDate == "" {
		loadBuildDate()
	}
	return buildDate
}
