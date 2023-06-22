# Gogo Land

As do several other Go-based modules, **ghostwire** also has to work around the
problems caused by Go's language design enforcing "[Composition over
Inheritance](https://en.wikipedia.org/wiki/Composition_over_inheritance)".

Let's take two use cases where embedding breaks down: a rather short, yet wide
hierarchy of "thing" properties. And we're talking here about lots of properties
(several tens) even on a single "thing", not just one, two, or three simple
properties where this might not be any problem. For example:

- **Linux network interface properties**: lots of common properties, and then
  even more for the various specific network interface types, such as bridges,
  MACVLANs, VETHs, TAPs, TUNs, et cetera.
  [@vishvananda/netlink](https://github.com/vishvananda/netlink) is one Go
  module that has to deal with this situation, Gostwire another example.

- **Linux namespace properties**: there are exactly three levels of abstraction,
  each one built upon the previous one. These are: namespaces, hierarchical
  namespaces (only PID and user), and finally user namespaces.

The source of woe is that Go features **interface polymorphism** as well as
**interface inheritance**, and then **composition**, but **no structural
inheritance**. Remember, only interfaces provide (sorry for the C++ terminology)
"virtual functions", composition doesn't.

1. Pretend that the problem doesn't exist and switch to a different project: if
   the language doesn't fit the project, then the project is to blame. Not
   really an option in our case.

2. Define a **set of interfaces**, optionally making use of interface
   inheritance (or not, it's not making the problem better or worse). Depending
   on the situation, these **interfaces may explose in terms of large numbers of
   accessors**. Then define the supporting struct types for storing the
   properties. This requires writing large amounts of accessor functions.

   [@thediveo/lxkns](https://github.com/TheDiveO/lxkns) went this route on the
   observation that there are just about 10 accessors required
   ([`namespaces.go`](https://github.com/TheDiveO/lxkns/blob/master/model/namespaces.go)),
   almost all being getters. So, "it depends".

3. Again, define a **set of interfaces**, but this time with **few accessors to
   the backing struct types**. This avoids explosion of accessors and their
   maintenance, **but don't forget to reimplement the accessor in all types**.
   This pattern also often introduces either lots of "stuttering" (to use Go
   terminology here) or alternatively **local variable redefinitions**
   (shadowing), for example:

   - [@vishvananda/netlink](https://github.com/vishvananda/netlink): defines a
     minimal `Link` interface with the `Attrs()*LinkAttrs` getter. A `Bridge`
     also implements the `Link` interface, but if you need to access
     bridge-specific attributes (properties), then you need to work with [type
     assertions](https://golang.org/ref/spec#Type_assertions) to get from the
     `Link` interface pointer to a pointer to the backing `Bridge` type. For
     instance: `bridge := bridge.(*Bridge)`.

   - Gostwire: `bridge := bridge.Bridge()` in order to access the properties of
     a bridge network interface. Please note that the problem didn't exist in
     the Ghostwire Python3 code base, because Python offers class-based
     inheritance.

   Of course, you can also go with a potentially large number of confusing
   variables in your code in order to avoid local variable redefinitions: `br :=
   bridge.(*Bridge)`. So you have nif, bridge, br, ...

Composition definitely has it uses where it thrives, just look at so many Go
modules. But it can also cause also large amounts of unnecessary coding work,
(D)RY, and even hidden runtime problems until you exactly know how to write unit
tests to unearth this class of runtime crashes.
