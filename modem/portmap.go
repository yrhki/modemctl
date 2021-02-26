package modem

import "net"

type PortMapProtocol int

const (
	PortMapProtocolUDP PortMapProtocol = iota
	PortMapProtocolTCP
	PortMapProtocolTCPUDP
)

type PortMap struct {
	Enabled bool
	id uint64
	RemoteHost, LocalHost net.IP
	LocalPort int
	Protocol PortMapProtocol
	ExternalPortRange PortRange
	Description string
}
