// bluetooth is a gokrazy helper that loads Bluetooth kernel modules on boot.
//
// Example:
//   Include the bluetooth package in your gokr-packer command:
//   % gokr-packer -update=yes \
//     github.com/gokrazy/breakglass \
//     github.com/gokrazy/bluetooth
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

func logic() error {
	flag.Parse()

	// modprobe the hci_uart driver for Raspberry Pi (3B+, others)
	for _, mod := range []string{
		"kernel/crypto/ecc.ko",
		"kernel/crypto/ecdh_generic.ko",
		"kernel/net/bluetooth/bluetooth.ko",
		"kernel/drivers/bluetooth/btbcm.ko",
		"kernel/drivers/bluetooth/hci_uart.ko",
	} {
		if err := loadModule(mod); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	dev := "hci0"
	target, err := checkBluetoothInterface(dev)
	if err != nil {
		log.Printf("Bluetooth interface %v not found.", target)
	} else {
		fmt.Printf("Bluetooth device %v: %v\n", dev, target)
	}

	// gokrazy should not supervise this process even when manually started.
	os.Exit(125)
	return nil
}

func checkBluetoothInterface(device string) (string, error) {
	target, err := os.Readlink("/sys/class/bluetooth/hci0")
	if err != nil {
		return "", fmt.Errorf("Bluetooth interface %v not found", device)
	}
	return target, nil
}

func loadModule(mod string) error {
	f, err := os.Open(filepath.Join("/lib/modules", release, mod))
	if err != nil {
		return err
	}
	defer f.Close()

	if err := unix.FinitModule(int(f.Fd()), "", 0); err != nil {
		if err != unix.EEXIST &&
			err != unix.EBUSY &&
			err != unix.ENODEV &&
			err != unix.ENOENT {
			return fmt.Errorf("FinitModule(%v): %v", mod, err)
		}
	}
	modname := strings.TrimSuffix(filepath.Base(mod), ".ko")
	log.Printf("modprobe %v", modname)
	return nil
}

var release = func() string {
	var uts unix.Utsname
	if err := unix.Uname(&uts); err != nil {
		fmt.Fprintf(os.Stderr, "minitrd: %v\n", err)
		os.Exit(1)
	}
	return string(uts.Release[:bytes.IndexByte(uts.Release[:], 0)])
}()

func main() {
	if err := logic(); err != nil {
		log.Fatal(err)
	}
}
