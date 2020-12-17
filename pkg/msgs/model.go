/*
 * Copyright (c)  The One True Way 2020. Use as described in the license. The authors accept no libility for the use of this software.  It is offered "As IS"  Have fun with it
 */

package msgs

const ENVELOPE_VERSION_1 = 1
const CLOUD_ID = "cloud"
const NB_MSG_PREFIX="natssync-nb"
const SB_MSG_PREFIX="natssync-sb"
const ECHOLET_SUFFIX="echolet"
const ECHO_SUBJECT_BASE="echo"
type MessageEnvelope struct {
	EnvelopeVersion int
	RecipientID     string
	SenderID        string
	Message         string
	Signature       string
	MsgKey          string
}
