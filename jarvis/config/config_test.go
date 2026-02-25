package config

import (
	"sync"
	"testing"
)

func TestCachedNetworkRace(t *testing.T) {
	NetworkString = "base"
	if err := SetNetwork(NetworkString); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup

	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10000; j++ {
				_ = Network()
			}
		}()
	}

	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			names := []string{"base", "matic", "linea", "polygon-zkevm", "monad"}
			for j := 0; j < 10000; j++ {
				_ = SetNetwork(names[j%len(names)])
			}
		}()
	}

	wg.Wait()
}
