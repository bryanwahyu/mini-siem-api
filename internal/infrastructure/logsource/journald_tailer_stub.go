//go:build !linux
// +build !linux

package logsource

type JournaldTailer struct{ Units []string }

func (t *JournaldTailer) Start() error { return nil }
func (t *JournaldTailer) Stop() error  { return nil }

