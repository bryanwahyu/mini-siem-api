//go:build linux
// +build linux

package logsource

import "fmt"

type JournaldTailer struct{ Units []string }

func (t *JournaldTailer) Start() error { return fmt.Errorf("journald tailer not implemented in this minimal build") }
func (t *JournaldTailer) Stop() error  { return nil }
