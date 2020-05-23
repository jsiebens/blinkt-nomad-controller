// MIT License

// Copyright (c) 2017 Alex Ellis
// Copyright (c) 2017 Isaac "Ike" Arias

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE

package blinkt

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ikester/gpio"
)

// DAT is the Data pin for Blinkt
const DAT int = 23

// CLK is the clock pin for Blinkt
const CLK int = 24

const redIndex int = 0
const greenIndex int = 1
const blueIndex int = 2
const brightnessIndex int = 3

// default raw brightness. Not to be used user-side
const defaultBrightnessInt int = 10

//upper and lower bounds for user specified brightness
const minBrightness float64 = 0.0
const maxBrightness float64 = 1.0

func convertBrightnessToInt(brightness float64) int {
	if !inRangeFloat(minBrightness, brightness, maxBrightness) {
		log.Fatalf("Supplied brightness was %#v - value should be between: %#v and %#v", brightness, minBrightness, maxBrightness)
	}
	return int(brightness * 31.0)
}

func inRangeFloat(minVal float64, testVal float64, maxVal float64) bool {
	return (testVal >= minVal) && (testVal <= maxVal)
}

// Hex2RGB converts a hexadecimal color string (e.g. #FFCC66) to RGB
func Hex2RGB(color string) (int, int, int) {
	red, _ := strconv.ParseInt(color[:2], 16, 32)
	green, _ := strconv.ParseInt(color[2:4], 16, 32)
	blue, _ := strconv.ParseInt(color[4:6], 16, 32)
	return int(red), int(green), int(blue)
}

// Blinkt holds the pixel array and related functions
type Blinkt struct {
	pixels          [8][4]int
	ShowAnimOnStart bool
	CaptureExit     bool
	ShowAnimOnExit  bool
	ClearOnExit     bool
}

// NewBlinkt creates a Blinkt to interact with. You must call "Setup()" after initial config.
func NewBlinkt(brightness ...float64) Blinkt {
	//brightness is optional so set the default
	brightnessInt := defaultBrightnessInt
	//override the default if the user has supplied a brightness value
	if len(brightness) > 0 {
		brightnessInt = convertBrightnessToInt(brightness[0])
	}
	bl := Blinkt{
		pixels:          initPixels(brightnessInt),
		ShowAnimOnStart: true,
		ShowAnimOnExit:  true,
		ClearOnExit:     true,
	}
	return bl
}

// Clear sets all the pixels to off, you still have to call Show.
func (bl *Blinkt) Clear() {
	r := 0
	g := 0
	b := 0
	bl.SetAll(r, g, b)
}

// Show updates the LEDs with the values from SetPixel/Clear.
func (bl *Blinkt) Show() *Blinkt {

	for i := 0; i < 4; i++ {
		bl.writeByte(0)
	}

	for p := range bl.pixels {
		brightness := bl.pixels[p][brightnessIndex]
		r := bl.pixels[p][redIndex]
		g := bl.pixels[p][greenIndex]
		b := bl.pixels[p][blueIndex]

		// 0b11100000 (224)
		bitwise := 224
		bl.writeByte(bitwise | brightness)
		bl.writeByte(b)
		bl.writeByte(g)
		bl.writeByte(r)
	}

	for i := 0; i < 4; i++ {
		bl.writeByte(0)
	}

	// Extra 4 bits for a total of 36
	gpio.DigitalWrite(DAT, 0)
	for i := 0; i < 4; i++ {
		gpio.DigitalWrite(CLK, 1)
		gpio.DigitalWrite(CLK, 0)
	}

	return bl
}

// SetAll sets all pixels to specified r, g, b colour. Show must be called to update the LEDs.
func (bl *Blinkt) SetAll(r int, g int, b int) *Blinkt {
	for p := range bl.pixels {
		bl.SetPixel(p, r, g, b)
	}
	return bl
}

// SetPixel sets an individual pixel to specified r, g, b colour. Show must be called to update the LEDs.
func (bl *Blinkt) SetPixel(p int, r int, g int, b int) *Blinkt {
	bl.pixels[p][redIndex] = r
	bl.pixels[p][greenIndex] = g
	bl.pixels[p][blueIndex] = b
	return bl
}

// SetPixelHex sets an individual pixel to specified Hex colour. Show must be called to update the LEDs.
func (bl *Blinkt) SetPixelHex(p int, color string) *Blinkt {
	r, g, b := Hex2RGB(color)
	return bl.SetPixel(p, r, g, b)
}

// SetBrightness sets the brightness of all pixels. Brightness supplied should be between: 0.0 to 1.0
func (bl *Blinkt) SetBrightness(brightness float64) *Blinkt {
	brightnessInt := convertBrightnessToInt(brightness)
	for p := range bl.pixels {
		bl.pixels[p][brightnessIndex] = brightnessInt
	}
	return bl
}

// SetPixelBrightness sets the brightness of pixel p. Brightness supplied should be between: 0.0 to 1.0
func (bl *Blinkt) SetPixelBrightness(p int, brightness float64) *Blinkt {
	brightnessInt := convertBrightnessToInt(brightness)
	bl.pixels[p][brightnessIndex] = brightnessInt
	return bl
}

// ShowInitialAnim displays a "start" light animation
func (bl *Blinkt) ShowInitialAnim() {
	red := 0
	green := 255
	blue := 0

	bl.Clear()
	bl.SetPixel(3, red, green, blue)
	bl.SetPixel(4, red, green, blue)
	for i := 1; i <= 10; i++ {
		bl.SetBrightness(float64(i) * 0.05)
		bl.Show()
		time.Sleep(70 * time.Millisecond)
	}

	for pixel := 2; pixel >= 0; pixel-- {
		mirrorPixel := 7 - pixel
		bl.SetPixel(pixel, red, green, blue)
		bl.SetPixel(mirrorPixel, red, green, blue)
		bl.Show()
		time.Sleep(80 * time.Millisecond)
	}

	bl.Clear()
	bl.Show()
}

// ShowFinalAnim displays a "start" light animation
func (bl *Blinkt) ShowFinalAnim() {
	red := 255
	green := 0
	blue := 0

	bl.SetAll(red, green, blue)
	bl.Show()
	time.Sleep(80 * time.Millisecond)

	for pixel := 0; pixel < 3; pixel++ {
		mirrorPixel := 7 - pixel
		bl.SetPixel(pixel, 0, 0, 0)
		bl.SetPixel(mirrorPixel, 0, 0, 0)
		bl.Show()
		time.Sleep(80 * time.Millisecond)
	}

	for i := 10; i > 0; i-- {
		bl.SetBrightness(float64(i) * 0.05)
		bl.Show()
		time.Sleep(70 * time.Millisecond)
	}

	bl.Clear()
	bl.Show()
}

// FlashPixel will flash a pixel on and off specified times and color
func (bl *Blinkt) FlashPixel(pixel, times int, color string) {
	red, green, blue := Hex2RGB(color)
	bl.SetPixel(pixel, red, green, blue)
	for i := 0; i < times; i++ {
		bl.SetPixelBrightness(pixel, 1.0)
		bl.Show()
		time.Sleep(30 * time.Millisecond)
		bl.SetPixelBrightness(pixel, 0)
		bl.Show()
		time.Sleep(30 * time.Millisecond)
	}
}

// FlashAll will flash all pixels on and off specified times and color
func (bl *Blinkt) FlashAll(times int, color string) {
	red, green, blue := Hex2RGB(color)
	bl.SetAll(red, green, blue)
	for i := 0; i < times; i++ {
		bl.SetBrightness(1.0)
		bl.Show()
		time.Sleep(30 * time.Millisecond)
		bl.SetBrightness(0)
		bl.Show()
		time.Sleep(30 * time.Millisecond)
	}
}

func initPixels(brightness int) [8][4]int {
	var pixels [8][4]int
	for i := range pixels {
		pixels[i][0] = 0
		pixels[i][1] = 0
		pixels[i][2] = 0
		pixels[i][3] = brightness
	}
	return pixels
}

// Setup gets Blinkt ready to display
func (bl *Blinkt) Setup() {
	gpio.Setup()
	gpio.PinMode(DAT, gpio.OUTPUT)
	gpio.PinMode(CLK, gpio.OUTPUT)
	if bl.ShowAnimOnStart {
		bl.ShowInitialAnim()
	}
	if bl.CaptureExit {
		bl.SetupExit()
	}
}

// SetupExit captures Interrupt and SIGTERM signals to handle program exit
func (bl *Blinkt) SetupExit() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		for range signalChan {
			log.Println("Blinkt: processing exit")
			bl.Cleanup()
		}
	}()
}

// Cleanup does final cleanup of resources
func (bl *Blinkt) Cleanup() {
	if bl.ShowAnimOnExit {
		bl.ShowFinalAnim()
	}
	if bl.ClearOnExit {
		bl.Clear()
		bl.Show()
	}
	gpio.Cleanup()
}

func (bl *Blinkt) writeByte(val int) {
	for i := 0; i < 8; i++ {
		// 0b10000000 = 128
		gpio.DigitalWrite(DAT, val&128>>7)
		gpio.DigitalWrite(CLK, 1)
		gpio.DigitalWrite(CLK, 0)
		val = val << 1
	}
}
