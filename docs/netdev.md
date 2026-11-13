# netdev Topologies

In Linux kernel lingo, "_netdev_" refers to "network devices" representing
network interfaces. Please note that the term "NIC" does not always corresponds
one-to-one with a netdev: a NIC may carry multiple netdevs, such as in case of
SR-IOV NICs, but also other NIC types.

When we here talk about "netdev topologies", we're actually referring to two
different topologies:
- **between** netdevs, such as between a MACVLAN vNIC and its so-called "master"
  NIC.
- **inside** netdevs in form of "hardware" RX and TX queues and how the Linux
  kernel serves them.

## Between netdevs

A prominent feature of Ghostwire is its discovery of the virtual networking
topology inside Linux hosts, such as which virtual Ethernet (VETH) interfaces
belong to each other and which ends are plugged into virtual bridges. Ghostwire
gathers this topological information mostly from the RTNETLINK netlink API. The
only exception are the SR-IOV PF-VF relationships which Ghostwire needs to
gather iinstea from the _bus_ details of the particular SR-IOV PF and VF
netdevs/NICs. 

For some nice pitfalls in how RTNETLINK returns inter-NIC topology information,
please refer to the [rtnetlink](rtnetlink) section.

## Inside netdevs

Now we're digging deeper, as we want to discover "inter-NIC" topologies, in
particular:

- netdev hardware aspects:
  - the available RX and TX "hardware" queues of a particular NIC,
  - the assigned IRQ(s), especially how they relate to individual RX and TX
    queue(s) where available on more complex NIC hardware,
- associated Linux kernel tasks serving a particular netdev:
  - where applicable, the IRQ handler kernel tasks serving the netdev IRQ(s) –
    usually only for Linux PREEMPT_RT,
  - where enabled (preferably by [`tuned`](https://tuned-project.org/) instead
    of some rogue kernel drivers), the dedicated NAPI kernel tasks and their
    relationships to the queues.

As it turns out, the Linux kernel has two, three APIs to provide such
information, albeit with differing quality and caveats attached:

- the `sysfs`, which lacks certain details but is "always" available;
- the "netdev family" NETLINK API/protocol, which has all necessary details (as
  long as drivers provide the necessary information, which they mostly don't)
  ... yet only works for NICs that are operational UP.

### sysfs Details about netdevs

There are two areas inside the `sysfs` that are of interest to us:
- `/sys/class/net/$IFNAME`.
- and `/sys/kernel/irq/$IRQNO`.

Combined, these areas tell us which RX and TX queues a netdev/NIC has on offer
and which IRQs it is using.

#### netdev Queues and IRQs

<br>

> [!IMPORTANT] Please note that `/sys/class/net` is network namespace-aware, but
> in a whacky and most probably unexpected way.

Any `sysfs` mount point freezes its `net` view based on the caller's current
network namespace at that time in the past. When later some process reads from
within that mount point, the reader's current network namespace doesn't matter
anymore. Instead, it just matters which mounted sysfs is read from, and that
dasdardly depends on the reader's current mount namespace – unless the reader
cunningly reads from `/proc/$PID/root/sys/class/net` to assume the viewpoint of
another process with PID `$PID`.

With these dirty details out of our ways, this is how `sysfs` exposes queue and
IRQ information of netdevs/NICs:

- `/sys/class/net/$IFNAME/`: `$IFNAME` is the netdev's/NIC's **current** name.
  - `queues/`
    - `rx-$NO/` and `tx-$NO/`: identifies an individual RX _or_ TX queue and its
      number. Please note that there is a shaky area in the kernel where it is
      assumed that a queue's number and a queue's ID are identical. We're here
      only interested in the _directory name_, but not in any files and
      directories _inside_ it, as this name tells us the queue types and IDs.
  - `device/`
    - `irq`: pseudo file containing a single IRQ number in base 10 textual
      representation together with a trailing newline, such as `42\n`.
    - `msi_irqs/`: contains individual pseudo-files with names reflecting the
      particular IRQ number (as usual, in base 10 textual representation), such
      as `123` and `666`.
      - `$IRQNO`: pseudo-file reporting the type of IRQ, such as `msix`; as
        usual, there's a trailing newline.

#### IRQ Details

The IRQ details tell us more about how particular IRQs relate to netdev/NIC
queues. Please note that it is not unusual to have a single IRQ assigned to a
pair of one RX and one TX queue, instead of one for RX, another for TX.

- `/sys/kernel/irq/$IRQNO`: pseudo directory with further properties of this
  particular IRQ.
  - `actions`: comma-separated list of associated actions, as registered by
    device drivers. For netdevs, action names follow the structure
    `$IFNAME-$QUEUETYPE-$QUEUEID`, with `-` separating the three individual
    elements:
    - `$IFNAME` is the **current** name of the corresponding netdev; this part
      of the action changes as a netdev gets renamed.
    - `$QUEUETYPE` indicates whether this IRQ applies to an RX queue, a TX
      queue, or to both a pair of RX and TX queues. Formats vary, but most
      commonly seem are (ignoring casing) `rx`, `tx`, and `rxtx` as well as
      `txrx`.
    - `$QUEUEID` finally specifies the queue(s) number/ID, in base 10 number
      format and corresponding with the queues listed in
      `/sys/class/net/$IFNAME/queues.

### procfs

Usually only on Linux [PREEMPT_RT
kernels](https://realtime-linux.org/getting-started-with-preempt_rt-guide/), the
IRQ(s) of a NIC are handled by so-called "IRQ kernel _tasks_" instead of IRQ
kernel _handlers_. The crucial difference here is that IRQ kernel _tasks_ are
subject to process/task scheduling and thus can be individually prioritized
(including their scheduling policy).

#### IRQ Handler Tasks

IRQ handler tasks are "kernel threads" that are children of the kernel process
`kthreadd` (PID2) and a name matching `irq/\d+-.*`.

### netdev NETLINK API

The "netdev" family NETLINK API/protocol is provided by Ghostwire's
`netdev/nlnetdev` package. This package currently implements only a subset of
the netdev family to the extend needed to discover the intra-NIC queue-IRQ-NAPI
topology.
