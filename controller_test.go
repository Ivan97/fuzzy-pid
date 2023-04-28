package fuzzy_pid

import (
	"fmt"
	"testing"
	"time"
)

func TestController(t *testing.T) {

	pid := NewFuzzyPid(1200, 650, 0.3, 1.0, 0.6, 0.01, 0.02, 0.01).SetDefaultConfig()
	target := 500.0
	actual := 0.0

	for i := 0; i < 20; i++ {
		signal := pid.Realize(target, actual)
		actual += signal
		fmt.Printf("===============signal: %f actual: %f======================\n", signal, actual)
		time.Sleep(500 * time.Millisecond)
	}

}
