package main

import (
	"fmt"
	"machine"
	"time"
)

type OutputState [4]bool

func main() {
	outputs := []machine.Pin{machine.A0, machine.A1, machine.A2, machine.A3}
	for _, out := range outputs {
		out.Configure(machine.PinConfig{
			Mode: machine.PinOutput,
		})
		out.High()
	}

	led := machine.LED
	led.Configure(machine.PinConfig{
		Mode: machine.PinOutput,
	})
	led.High()

	uart := machine.UART1
	uart.Configure(machine.UARTConfig{
		BaudRate: 9600,
	})

	okC := make(chan OutputState, 1)

	go func() {
		uart.Write([]byte("ready\n"))
		state := OutputState{true, true, true, true}
		for {
			b, err := uart.ReadByte()
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			if b < 'A' || b > 'A'+0b1111 {
				uart.Write([]byte("error: invalid character\n"))
				continue
			}
			bits := b - 'A'
			state[0] = bits&0b1000 == 0b1000
			state[1] = bits&0b0100 == 0b0100
			state[2] = bits&0b0010 == 0b0010
			state[3] = bits&0b0001 == 0b0001
			select {
			case okC <- state:
				uart.Write([]byte("ok\n"))
			default:
				uart.Write([]byte("error: chan full\n"))
			}
		}
	}()

	var started bool
	state := OutputState{true, true, true, true}
	lastGood := time.Time{}
	fmt.Fprintln(uart, "# ready")
	for {
		time.Sleep(100 * time.Millisecond)
		select {
		case state = <-okC:
			started = true
			fmt.Fprintf(uart, "# set state %v\n", state)
			led.Low()
			lastGood = time.Now()
			for i := range state {
				outputs[i].Set(state[i])
			}
			continue
		default:
			if time.Since(lastGood) < 5*time.Second {
				continue
			}
			if started {
				fmt.Fprintf(uart, "# timeout after %v; sleep 30\n", time.Since(lastGood))
				led.High()
				for _, out := range outputs {
					out.High()
				}
				time.Sleep(30 * time.Second)
			}
		}
	}
}
