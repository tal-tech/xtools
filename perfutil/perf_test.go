package perfutil

import (
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	config := NewperfConfig()
	config.TimeDebug = "true"
	InitPerfWithConfig(config)
	m.Run()
	time.Sleep(time.Second)
}

func TestCountI(t *testing.T) {
	CountI("fortest.Info")
}

func TestCountW(t *testing.T) {
	CountW("fortest.Warn")
}

func TestAutoElapsed(t *testing.T) {
	defer AutoElapsed("Cost.Test", time.Now())
	time.Sleep(20 * time.Millisecond)
}

func TestAutoElapsedDebug(t *testing.T) {
	defer AutoElapsedDebug("Cost.Debug")()
	time.Sleep(200 * time.Millisecond)
}
