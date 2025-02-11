package interception

import (
	"github.com/safing/portmaster/firewall/interception/nfq"
	"github.com/safing/portmaster/network/packet"
)

// start starts the interception.
func start(ch chan packet.Packet) error {
	return StartNfqueueInterception(ch)
}

// stop starts the interception.
func stop() error {
	return StopNfqueueInterception()
}

// ResetVerdictOfAllConnections resets all connections so they are forced to go thought the firewall again.
func ResetVerdictOfAllConnections() error {
	return nfq.DeleteAllMarkedConnection()
}
