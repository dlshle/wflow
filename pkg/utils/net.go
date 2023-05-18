package utils

import (
	"net"

	"github.com/dlshle/gommon/errors"
)

func GetMACAddress() (addr string, err error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
			var addrs []net.Addr
			addrs, err = iface.Addrs()
			if err != nil {
				continue
			}

			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					mac := iface.HardwareAddr
					if mac != nil {
						return mac.String(), nil
					}
				}
			}
		}
	}
	err = errors.Error("MAC address not found")
	return
}
