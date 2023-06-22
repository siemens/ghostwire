# API (Woes)

> Are you looking for Gostwire's [REST API](rest-api) instead?

## Pitfalls

### `net.IP`

Go's [`net.IP`](https://pkg.go.dev/net#IP) handles both IPv4 and IPv6. Luckily,
most of the time, users of the `net.IP` package won't ever notice. However,
there are some ugly pitfalls in the design of `net.IP` to be aware of when
working more closely with IPv4 and IPv6.

Basically, `net.IP` is nothing more than a glorified `[]byte` of either length 4
or 16; there are even the corresponding length constants `net.IPv4len` and
`net.IPv6len` defined...

> [!ATTENTION] Never assume that `len(net.IP) == net.IPv6len` means that this is
> an IPv6 address. Go usually stores IPv4 addresses as IPv4-mapped IPv6
> addresses.

`net.IP` almost always stores IPv4 addresses as IPv6 addresses (unless you trick
it): it stores the IPv4 addresses as so-called "[IPv4-mapped IPv6 addresses]()".
See also: [RFC 4291
2.5.5.2](https://datatracker.ietf.org/doc/html/rfc4291#section-2.5.5.2) and
[RFC 6890](https://datatracker.ietf.org/doc/html/rfc6890).

For instance, `net.ParseIP` and others return IPv4 addresses as IPv4-mapped IPv6
addresses.

In turn, checking for an IPv4 address using `len(ipaddr) == net.IPv4len` will
fail miserably. Go basically forces programmers to check for IPv4 addresses by
trying to convert a given `net.IP` using the `To4` method. If this is either a
4-byte long IPv4 address or an IPv4-mapped IPv6 address, then `To4` will return
a 4-byte long `net.IP`, otherwise `nil`.

```go
myip := net.ParseIP("192.168.6.66")
println(len(myip)) // prints "16"
if ipv4 := myip.To4(); ipv4 != nil {
    println(len(ipv4)) // prints "4"
}

```

Comparing an IPv4 address (such as one gotten from `To4`) with `net.IPv4zero`
will fail when your IPv4 address has `net.IPv4len`: the length of `net.IPv4zero`
is 16!

> [!NOTE] The Gostwire API returns IPv4 addresses **always** as `net.IP`s with a
> length of `IPv4len` – as does the Go netlink Gostwire relies on. No if's or
> but's.

## Gostwire's API Design

### Network Interfaces

In Gostwire, `NetworkInterface` is actually a "network interface" interface and
is implemented by all backing types of network interfaces.

The attributes, or properties, common to all network interfaces (`NifAttrs`) are
then accessed through the `Nif()*NifAttrs` accessor[^oops], including the
specific `Type` of network interface.

For some types, Gostwire defines additional attributes, such as for bridges in
form of `BridgeAttrs`. If `nif` is a `NetworkInterface` interface, then all
attributes of a bridge are accessible via `nif.(Bridge).Bridge()` returning a
`*BridgeAttr` pointer.

```go
// Get all bridge "ports" of the bridge network interface in nif
// using a type assertion assignment
var nif NetworkInterface
if bridge, ok := nif.(network.Bridge); ok {
    bridge := bridge.Bridge()
    for _, port := range bridge.Ports {
        // ...
    }
}
```

---

#### References

[^oops]: [Object Oriented Inheritance in
  Go](https://hackthology.com/object-oriented-inheritance-in-go.html) by _Tim
  Henderson_, please take special note of the section entitled "Limitations of
  Embedding as Inheritance".
