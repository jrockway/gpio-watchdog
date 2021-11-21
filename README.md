# gpio-watchdog

This is a small program for a Feather M0 that accepts input on the UART to describe which GPIOs to
turn on, and resets to a default state if the input stops arriving. I use this from a Raspberry Pi
to ensure that GPIOs go to a known-good state if the Raspberry Pi program crashes.

Depending on your application, you may not want to resume execution after pausing for 30 seconds,
but that is what this particular program does.

To build: `tinygo build -target=feather-m0 -o out.uf2`

UART1 is used, with TX on pin D10, and RX on pin D11. I don't know why tinygo doesn't expose the
UART that is labelled on the Feather M0's silkscreen (UART0?), but it doesn't and I didn't care.

## Input

Set the GPIOs and poke the watchdog timer by writing a byte 'A'-'P'. The GPIO state is represented
as four bits, a 0 bit is low, a 1 bit is high. The bit mask is added to 'A' (0x41). A represents
0b0000, or all low. B represents 0b0001, or output 1 low, output 2 low, output 3 low, output 4 high.

If no input is received for 5 seconds, all inputs are set to high, and no data is read for 30
seconds. The process then continues as normal. (If no input is received after the 30 seconds, the
GPIOs stay high until input arrives and represents a different state.)

## Output

If a byte is successfully processed and represents a valid state, "ok\n" is printed. If an error
occurs, a description is printed prefixed by "error: ", like "error: invalid character\n". Debugging
information is output preceeded by a "#". All output is newline-separated.

## Example session

```
< # ready
> A
< ok
< # set state [false false false false]
> A
< ok
< # set state [false false false false]
<5 seconds pass>
< # timeout after 5.001 seconds; sleep 30
> A
> error: chan full
<30 seconds pass>
> A
< ok
< # set state [false false false false]
```

## Why UART?

It seemed impossible to get tinygo to be an I2C target (instead of controller) or a SPI peripheral.
UART was my third choice.

## Why busy-wait for input?

We can't:

```go
timeout := time.After(5*time.Second)
select {
case st := <-okC
    ...
case <-timeout:
    ...
}
```

Because of https://github.com/tinygo-org/tinygo/issues/1037
