/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package msgs

import (
	"errors"
	"fmt"
	"strings"

	"github.com/theotw/natssync/pkg"
	"github.com/theotw/natssync/pkg/bridgemodel"
)

const ENVELOPE_VERSION_1 = 1 //EBC AES
const ENVELOPE_VERSION_2 = 2 // CBC AES
const ENVELOPE_VERSION_3 = 3 // CBC AES, update version
const ECHOLET_SUFFIX = "echolet"
const ECHO_SUBJECT_BASE = "echo"
const NATSSYNC_MESSAGE_PREFIX = "natssyncmsg"

type MessageEnvelope struct {
	EnvelopeVersion int
	RecipientID     string
	SenderID        string
	Message         string
	Signature       string
	MsgKey          string
	KeyID           string
}

func MakeReplySubject(replyToLocationID string) string {
	replySubject := fmt.Sprintf("%s.%s.%s", NATSSYNC_MESSAGE_PREFIX, replyToLocationID, bridgemodel.GenerateUUID())
	return replySubject
}
func MakeNBReplySubject() string {
	replySubject := fmt.Sprintf("%s.%s.%s", NATSSYNC_MESSAGE_PREFIX, pkg.CLOUD_ID, bridgemodel.GenerateUUID())
	return replySubject
}

func MakeEchoSubject(clientID string) string {
	subject := fmt.Sprintf("%s.%s.%s", NATSSYNC_MESSAGE_PREFIX, clientID, ECHO_SUBJECT_BASE)
	return subject
}
func MakeMessageSubject(locationID string, params string) string {
	if len(params) == 0 {
		return fmt.Sprintf("%s.%s", NATSSYNC_MESSAGE_PREFIX, locationID)
	}
	return fmt.Sprintf("%s.%s.%s", NATSSYNC_MESSAGE_PREFIX, locationID, params)
}

type ParsedSubject struct {
	OriginalSubject string
	LocationID      string
	AppData         []string //dotted strings parts after the location ID
}

func ParseSubject(subject string) (*ParsedSubject, error) {
	parts := strings.Split(subject, ".")
	if len(parts) < 2 || (parts[0] != NATSSYNC_MESSAGE_PREFIX) {
		return nil, errors.New("invalid.message.subject")
	}
	ret := new(ParsedSubject)
	ret.LocationID = parts[1]
	ret.AppData = parts[2:]
	ret.OriginalSubject = subject

	return ret, nil
}
