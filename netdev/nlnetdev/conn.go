// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nlnetdev

import (
	"fmt"

	"github.com/mdlayher/genetlink"
	"github.com/mdlayher/netlink"
)

// See: https://elixir.bootlin.com/linux/v6.9.5/source/include/uapi/linux/netdev.h#L9
const (
	NetdevFamilyName    = "netdev"
	NetdevFamilyVersion = 1
)

// See: https://elixir.bootlin.com/linux/v6.8/source/include/uapi/linux/netdev.h#L135
const (
	_ = iota
	NETDEV_CMD_DEV_GET
	NETDEV_CMD_DEV_ADD_NTF
	NETDEV_CMD_DEV_DEL_NTF
	NETDEV_CMD_DEV_CHANGE_NTF
	NETDEV_CMD_PAGE_POOL_GET
	NETDEV_CMD_PAGE_POOL_ADD_NTF
	NETDEV_CMD_PAGE_POOL_DEL_NTF
	NETDEV_CMD_PAGE_POOL_CHANGE_NTF
	NETDEV_CMD_PAGE_POOL_STATS_GET
	NETDEV_CMD_QUEUE_GET // only since 6.8+
	NETDEV_CMD_NAPI_GET  // only since 6.8+
)

// Conn represents a (generic) NETLINK connection of the "netdev" family to the
// kernel, for diagnosing network interfaces regarding their available RX/TX
// queues and associated NAPIs.
type Conn struct {
	conn   *genetlink.Conn
	family genetlink.Family
}

// Assuming that the family information stays constant during the runtime of any
// application making use of this module, we fetch the netdev family information
// only once.
func init() {
	var conn *genetlink.Conn
	var err error

	defer func() {
		familyErr = err
		if conn != nil {
			_ = conn.Close()
		}
	}()

	conn, err = genetlink.Dial(nil)
	if err != nil {
		return
	}
	netdevFamily, err = conn.GetFamily(NetdevFamilyName)
	if err != nil {
		err = fmt.Errorf("generic NETLINK family %q not available, reason: %w",
			NetdevFamilyName, err)
		return
	}
	if netdevFamily.Version != NetdevFamilyVersion {
		err = fmt.Errorf("expected family %q version %d, got %d",
			NetdevFamilyName, NetdevFamilyVersion, netdevFamily.Version)
		return
	}
}

var (
	netdevFamily genetlink.Family
	familyErr    error
)

// Dial a NETLINK connection for diagnosing network interfaces about their
// queues and NAPIs. In case you need to diagnose netdevs in a different network
// namespace, lock the calling go routine to an OS-level thread, switch that
// thread to the target network namespace, dial, and afterwards restore the
// original network namespace attachment and unlock. The NETLINK connection
// stays connected to the network namespace active when calling Dial.
func Dial(config *netlink.Config) (*Conn, error) {
	if familyErr != nil {
		return nil, familyErr
	}
	c, err := genetlink.Dial(config)
	if err != nil {
		return nil, err
	}
	return &Conn{
		conn:   c,
		family: netdevFamily,
	}, nil
}

// Close the NETLINK connection.
func (c *Conn) Close() error {
	return c.conn.Close()
}
