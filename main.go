package main

import (
	"flag"
	"fmt"
	"github.com/jsiebens/nomad-blinkt/pkg/blinkt"
	"github.com/jsiebens/nomad-blinkt/pkg/metrics"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var r = flag.String("resource", "allocations", "resource to monitor")
var max = flag.Int("max", 8, "maximum allowed allocations (used when -resource=allocations")

func main() {
	flag.Parse()

	client, _ := metrics.NewClient(metrics.DefaultConfig())

	brightness := 0.5
	bl := blinkt.NewBlinkt(brightness)

	bl.Setup()

	Delay(100)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)

	var step = 0
	for {
		select {
		case <-signalChan:
			bl.Cleanup()
			return
		default:

			a, err := client.PercentageOfAllocatedResource(*r, *max)

			if err == nil {

				x := int(a * 8)

				fmt.Printf("%s: %f -> %d\n", *r, a, x)

				for i := 0; i < 8; i++ {
					if (x - 1) >= i {
						bl.SetPixelBrightness(i, brightness)
						bl.SetPixelHex(i, getColor(i))
					} else {
						bl.SetPixelBrightness(i, 0)
						bl.SetPixelHex(i, "000000")
					}
				}

				bl.Show()

				step++
			} else {
				bl.FlashAll(2, "FF0000")
			}

			Delay(5000)
		}
	}
}

func getColor(i int) string {
	switch i {
	case 6:
		return "FFA500"
	case 7:
		return "FF0000"
	default:
		return "00FF00"
	}
}

// Delay maps to time.Sleep, for ms milliseconds
func Delay(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}
