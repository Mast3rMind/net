package main

import (
	"fmt"
	"github.com/izqui/tuntap/tuntap"
	"io"
	"net"
	"os/exec"
	"reflect"
)

const (
	name  = "tun0"
	link  = "fe80::1"
	mask  = "beef::/112"
	MY_ID = "beef::52"
)

type PacketsCallback chan *tuntap.IPPacket

func main() {

	cb := make(PacketsCallback)
	go setupInterface(cb)

	for {

		select {
		case packet := <-cb:

			source := net.IP(packet.Header.SourceAddr())
			dest := net.IP(packet.Header.DestAddr())
			me := net.ParseIP(MY_ID)

			if dest.IsMulticast() {

				fmt.Println("Multicast packet received")

			} else if reflect.DeepEqual(dest, me) {
				//Packet is for me?
				fmt.Println("Packet for me received from", source)
			} else {

				fmt.Println("Packet that is not for me received", source, dest)
			}
		}
	}

}

func setupInterface(cb PacketsCallback) {

	tund, err := tuntap.Open(name, tuntap.DevTun, false)
	panicOnError(err)
	fmt.Println("Created interface", name)

	cmd := exec.Command("ifconfig", name, "inet6", link, "up")
	out, err := cmd.Output()
	panicOnError(err)
	fmt.Println("Assign address", link, "to interface", name)

	cmd = exec.Command("route", "-n", "add", "-inet6", mask, link)
	out, err = cmd.Output()
	panicOnError(err)
	fmt.Println(string(out))

	for {

		packet, err := tund.ReadPacket()

		if err == nil && packet != nil {

			cb <- packet
		}
	}
}

func panicOnError(err error) {
	if err != nil && err != io.EOF {
		panic(err)
	}
}
