package relay

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kevin-env/kevin/internal/container"
)

// UDPPoolSizeEnvVar overrides the number of UDP relay ports the relay
// container publishes for SOCKS5 UDP ASSOCIATE sessions - a host/local-dev
// setting, not a kevin.cue field, read directly here and by
// internal/plugins/builtin/kubernetes so both sides agree on the pool size without
// either learning it from the other.
const UDPPoolSizeEnvVar = "KEVIN_RELAY_UDP_POOL_SIZE"

// UDPHostPortsEnvVar names the environment variable that confines the host
// ports of the UDP pool to a "low-high" range. Unset, the engine picks any
// free port.
const UDPHostPortsEnvVar = "KEVIN_RELAY_UDP_HOST_PORTS"

// defaultUDPPoolSize is the number of UDP relay ports published when
// UDPPoolSizeEnvVar is unset.
const defaultUDPPoolSize = 16

// udpRelayPortBase is the first container-side port in the relay's fixed
// UDP ASSOCIATE pool. The container publishes udpPoolSize() consecutive
// ports starting here, so a SOCKS5 ASSOCIATE session can bind one before
// whatever needs it (a container, a kind pod) exists.
const udpRelayPortBase = 40000

// UDPPoolSize reads UDPPoolSizeEnvVar, falling back to defaultUDPPoolSize
// when unset - shared by this package and internal/plugins/builtin/kubernetes so both
// sides of the pool (the relay container's published ports and the
// cluster relay forwarder's) agree on its size without either learning it
// from the other.
func UDPPoolSize() (int, error) {
	v := os.Getenv(UDPPoolSizeEnvVar)
	if v == "" {
		return defaultUDPPoolSize, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 0, ErrInvalidUDPPoolSize
	}
	return n, nil
}

// UDPPublishSpecs returns the loopback publish spec for each of
// containerPorts: Docker's choice of free port when UDPHostPortsEnvVar is
// unset, otherwise a free port inside that range. It returns
// [ErrInvalidUDPHostPorts] when the variable is not a valid range, and
// nothing when containerPorts is empty, so a pool size of 0 ignores it.
func UDPPublishSpecs(containerPorts []int) ([]string, error) {
	if len(containerPorts) == 0 {
		return nil, nil
	}
	hostPorts := os.Getenv(UDPHostPortsEnvVar)
	if hostPorts != "" {
		lo, hi, ok := strings.Cut(hostPorts, "-")
		l, errLo := strconv.Atoi(lo)
		h, errHi := strconv.Atoi(hi)
		if !ok || errLo != nil || errHi != nil || l < 1 || h > 65535 || l > h {
			return nil, ErrInvalidUDPHostPorts
		}
	}
	specs := make([]string, len(containerPorts))
	for i, p := range containerPorts {
		specs[i] = fmt.Sprintf("127.0.0.1:%s:%d/udp", hostPorts, p)
	}
	return specs, nil
}

// udpRelayPorts lists n container-side UDP ports, starting at
// udpRelayPortBase.
func udpRelayPorts(n int) []int {
	ports := make([]int, n)
	for i := range ports {
		ports[i] = udpRelayPortBase + i
	}
	return ports
}

// udpRelayPortsArg appends --udp-relay-ports naming the n-port range
// starting at udpRelayPortBase to args, or leaves args untouched when n is
// zero - kevin-relay's own flag then defaults to no UDP ASSOCIATE capacity.
func udpRelayPortsArg(n int, args []string) []string {
	if n == 0 {
		return args
	}
	return append(args, "--udp-relay-ports", fmt.Sprintf("%d-%d", udpRelayPortBase, udpRelayPortBase+n-1))
}

// UDPAddrsFromInfo builds a container-port -> host-address map from every
// UDP port info publishes - whatever the relay's UDP pool actually
// published when it was created, regardless of the caller's own current
// UDPPoolSizeEnvVar (a container reused across a pool-size change keeps
// its original pool; see docs/site/content/docs/concepts/relay.md).
func UDPAddrsFromInfo(info container.Container) map[int]string {
	var addrs map[int]string
	for port, addr := range info.Ports {
		name, ok := strings.CutSuffix(port, "/udp")
		if !ok {
			continue
		}
		containerPort, err := strconv.Atoi(name)
		if err != nil {
			continue
		}
		if addrs == nil {
			addrs = make(map[int]string, len(info.Ports))
		}
		addrs[containerPort] = addr
	}
	return addrs
}
