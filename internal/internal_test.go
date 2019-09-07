package internal_test

import (
	"testing"

	"github.com/neubot/dash/internal"
)

func TestNoLogger(t *testing.T) {
	var nl internal.NoLogger
	nl.Debug("abc")
	nl.Debugf("%s", "abc")
	nl.Info("abc")
	nl.Infof("%s", "abc")
	nl.Warn("abc")
	nl.Warnf("%s", "abc")
}
