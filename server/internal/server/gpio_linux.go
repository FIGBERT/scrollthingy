package server

import (
	"strconv"
	"time"

	lksdk "github.com/livekit/server-sdk-go/v2"
	gpio "github.com/warthog618/go-gpiocdev"
)

const GPIO_CHIP = "gpiochip0"
const GPIO_DELAY = time.Millisecond

type ports struct {
	direction, step, power *gpio.Line
}

func (s *Server) setupGPIO() (error, *int) {
	dir := 23
	stp := 24
	pwr := 25

	direction, err := gpio.RequestLine(GPIO_CHIP, dir, gpio.AsOutput(0))
	if err != nil {
		return err, &dir
	}
	step, err := gpio.RequestLine(GPIO_CHIP, stp, gpio.AsOutput(0))
	if err != nil {
		return err, &stp
	}
	power, err := gpio.RequestLine(GPIO_CHIP, pwr, gpio.AsOutput(0))
	if err != nil {
		return err, &stp
	}

	s.ports = &ports{direction, step, power}
	return nil, nil
}

func (s *Server) CleanupGPIO() {
	s.ports.direction.Reconfigure(gpio.AsInput)
	s.ports.direction.Close()
	s.ports.step.Reconfigure(gpio.AsInput)
	s.ports.step.Close()
	s.ports.power.Reconfigure(gpio.AsInput)
	s.ports.power.Close()
}

func (s *Server) handleScroll(reader *lksdk.TextStreamReader, participant string) {
	delta, err := strconv.Atoi(reader.ReadAll())
	if err != nil || participant != s.state.current {
		return
	}

	if delta < 0 {
		s.ports.direction.SetValue(0)
		delta *= -1
	} else {
		s.ports.direction.SetValue(1)
	}

	for _ = range delta {
		s.ports.step.SetValue(1)
		<-time.After(GPIO_DELAY)
		s.ports.step.SetValue(0)
	}
}

func (s *Server) killPowerToMotor() {
	s.ports.power.SetValue(1)
}

func (s *Server) restorePowerToMotor() {
	s.ports.power.SetValue(0)
}
