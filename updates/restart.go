package updates

import (
	"context"
	"sync"
	"time"

	"github.com/tevino/abool"

	"github.com/safing/portbase/log"
	"github.com/safing/portbase/modules"
)

const (
	// RestartExitCode will instruct portmaster-start to restart the process immediately, potentially with a new version.
	RestartExitCode = 23
)

var (
	restartTask      *modules.Task
	restartPending   = abool.New()
	restartTriggered = abool.New()

	restartTime     time.Time
	restartTimeLock sync.Mutex
)

// IsRestarting returns whether a restart has been triggered.
func IsRestarting() bool {
	return restartTriggered.IsSet()
}

// RestartIsPending returns whether a restart is pending.
func RestartIsPending() (pending bool, restartAt time.Time) {
	if restartPending.IsNotSet() {
		return false, time.Time{}
	}

	restartTimeLock.Lock()
	defer restartTimeLock.Unlock()

	return true, restartTime
}

// DelayedRestart triggers a restart of the application by shutting down the
// module system gracefully and returning with RestartExitCode. The restart
// may be further delayed by up to 10 minutes by the internal task scheduling
// system. This only works if the process is managed by portmaster-start.
func DelayedRestart(delay time.Duration) {
	// Check if restart is already pending.
	if !restartPending.SetToIf(false, true) {
		return
	}

	// Schedule the restart task.
	log.Warningf("updates: restart triggered, will execute in %s", delay)
	restartAt := time.Now().Add(delay)
	restartTask.Schedule(restartAt)

	// Set restartTime.
	restartTimeLock.Lock()
	defer restartTimeLock.Unlock()
	restartTime = restartAt
}

// AbortRestart aborts a (delayed) restart.
func AbortRestart() {
	if restartPending.SetToIf(true, false) {
		log.Warningf("updates: restart aborted")

		// Cancel schedule.
		restartTask.Schedule(time.Time{})
	}
}

// TriggerRestartIfPending triggers an automatic restart, if one is pending.
// This can be used to prepone a scheduled restart if the conditions are preferable.
func TriggerRestartIfPending() {
	if restartPending.IsSet() {
		restartTask.StartASAP()
	}
}

// RestartNow immediately executes a restart.
// This only works if the process is managed by portmaster-start.
func RestartNow() {
	restartPending.Set()
	restartTask.StartASAP()
}

func automaticRestart(_ context.Context, _ *modules.Task) error {
	// Check if the restart is still scheduled.
	if restartPending.IsNotSet() {
		return nil
	}

	// Trigger restart.
	if restartTriggered.SetToIf(false, true) {
		log.Warning("updates: initiating (automatic) restart")

		// Set restart exit code.
		modules.SetExitStatusCode(RestartExitCode)
		// Do not use a worker, as this would block itself here.
		go modules.Shutdown() //nolint:errcheck
	}

	return nil
}
