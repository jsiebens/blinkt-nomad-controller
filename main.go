package main

import (
	"github.com/jsiebens/nomad-blinkt/pkg/blinkt"
	"github.com/jsiebens/nomad-blinkt/pkg/metrics"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	client, _ := metrics.NewClient(metrics.DefaultConfig())

	brightness := 0.5
	bl := blinkt.NewBlinkt(brightness)

	bl.Setup()

	Delay(500)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-signalChan:
			bl.Cleanup()
			return
		default:
			m, err := client.Metrics()

			if err == nil {

				var x = 0
				for _, g := range m.Gauges {
					if g.Name == "nomad.client.allocations.running" {
						x = int(g.Value)
						break
					}
				}

				for i := 0; i < 8; i++ {
					if (x - 1) >= i {
						bl.SetPixelBrightness(i, brightness)
						bl.SetPixelHex(i, "00FF00")
					} else {
						bl.SetPixelBrightness(i, 0)
						bl.SetPixelHex(i, "000000")
					}
				}

				bl.Show()
			} else {
				bl.FlashAll(2, "FF0000")
			}

			Delay(1000)
		}
	}
}

// Delay maps to time.Sleep, for ms milliseconds
func Delay(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}
