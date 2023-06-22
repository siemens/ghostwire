<img src="public/ghostwire192.png" style="height: 8ex; float: right;">

This UI style guide introduces and show-cases the UI components of our new
Ghostwire web user interface. At the same time, it also serves as some weird
kind of testbed to see if and how our UI components (hopefully correctly)
render.

> **Note:** the component examples use the Ghostwire light theme. In particular,
> the Ghostwire light theme uses "Petrol" as its primary color, and orange as
> the secondary color. This applies only to the components, but not to this
> Style Guide itself outside any Ghostwire components.

### Ghostwire

In case you're new to Ghostwire, here's a short explanation of what it does:

> Ghostwire discovers the virtual communication inside Linux hosts, such as
> virtual (and real) network stacks, virtual bridges (Ethernet switches),
> virtual Ethernet wires, overlay Ethernet networks, et cetera. Additionally,
> Ghostwire also discovers how containers (or uncontainerized, stand-alone
> processes) use these virtual network stacks. On top of this, Ghostwire then
> scans any containers found for their DNS client configurations. Wait,
> there's even more: netstat (and socket status) on steroids ... which
> transport ports are listening or connected for which container, process, et
> cetera?

The job of the UI then is to present this mess to unsuspecting users in the
hope they might bring sense to the virtual communication spaghetti.

### Wait ... All These UI Components are ... Functions?!

Ghostwire's UI components are so-called [React](https://reactjs.org/) "[function
components](https://reactjs.org/docs/components-and-props.html#function-and-class-components)":
they are described as (Typescript) *functions* ... as opposed to classes (and
objects). If you think we're nuts, rest asured we thought that too.

But then we quickly noticed how sleek developing new components works with a
functional pattern. In fact, no more class boilerplate. And even function
components can be stateful, thanks to React's
[hook](https://reactjs.org/docs/hooks-intro.html) architecture.

### Basic Ghostwire Model Terminology

Some quick hints as to the terminology used in Ghostwire's discovery model
should help you in finding your way around the UI code. We had to come up with
our own terminology here, as up to this very date there unfortunately is no
terminology already established. This might be due to the fact that tools
usually either focus only on namespaces, or processes, or containers, but
never transgress these domains all the time and in the same tool.

- **containee**: "something" contained in (or confined by) a network namespace.
  There are so-called "primitive" containees as well as container groups. Typed
  as `Containee`.

  - **primitive containee** (`PrimitiveContainee`, as opposed to groups of
    containers):
    - **container** is the highest level type of containee, that is a user-space
      artefact managed by a container engine. Ghostwire learns the container's
      name as well as the top-most process inside the container (the so-called
      "ealdorman" process). Type is `Container`.
    - **busybox** is a process (or subtree of processes) inside a namespace, but
      not managed by any container engine. (And seriously, even if `systemd`
      thinks otherwise, it won't ever be a real container engine, no way.) Type
      is `Busybox`.
    - **sandbox** is a bind-mount keeping a network namespace alive. This type
      of containee only gets reported by the Ghostwire discovery engine when
      there is neither a container or busybox containee "inside" a particular
      network namespace, but only the bind mount. Type is `Sandbox`.

  - **pod** is a "strong" group of containers, **sharing the same network
    namespace**. The term "pod" here describes the abstract concept, as opposed
    to specific incarnations, such as "Kuberentes pods", et cetera. Type is
    `Pod`.

- **ealdorman** process: the most senior process inside a (network) namespace,
  both in its position within the process hiearchy, as well as in its age
  relative to when the Linux kernel was booted. The ealdorman is the most senior
  of possibly several *leader* processes in a namespace. In its simplest case
  with only a single leader process, this leader process simultaneously is also
  the ealdorman. The reason for ealdorman processes in the discovery model is
  purely cosmetic: it helps simplifying things in case there are additional
  "visiting" processes attached to network namespaces, which originally do not
  belong to the container process(es).
  
  > And yes, I've obviously read Bernard Cornwell's "Uhtred" 📚 series.

- **leader** process: the topmost process within the process hierarchy still
  inside a specific (network) namespace. It is possible to have multiple leader
  processes inside the same namespace, such as when sharing the same network
  namespace between multiple containers (for instance, as it is the case with
  Kubernetes pods).

### No Redux

The UI doesn't use [Redux](https://redux.js.org/) – as it is, the Redux
architecture cannot handle recursive state information (and isn't designed for
it on purpose). However Ghostwire's discovery information model makes heavy use
of (indirect) recursive object references for good reason, such as between
network namespaces and containees, between namespaces and network interfaces,
between related network interfaces, et cetera. This allows fast navigation on
the discovery information. Using Redux would required a complete remodelling
with questionable outcome, making navigation on the information model awkward
and extremely cumbersome.

Instead, the Ghostwire UI uses [jōtai](https://github.com/pmndrs/jotai) for its
state management – which is wholy sufficient for these purposes.

### Built-In Discovery Mock Data

This UI style guide comes with built-in discovery mock data that can be found
in `models/gw/mock/`. This mock data has been taken from a real system (please
also see below); not every aspect of the mock data is fully controlled, but it
always contains a well-known subset of containers and virtual network
configuration.

Most of the time in this style guide you'll see this import in order to use
the (frozen and thus reliable) discovery mock data:

```tsx static
import { discovery } from 'models/gw/mock'
```

### Rebuilding Discovery Mock Data

Under normal circumstances, you should never have need to rebuild the discovery
mock data. But then, discovery mock data can be (re)generated using the
docker-compose project in `(../)mock/`; change into the `mock/` directory and
then run the following commands – expecting the `kind` command to be installed
in your `$(HOME)/go/bin` in order to create Kubernetes clusters inside Docker
containers:

```bash
make up kindup
make mockdata
make kinddown down
```

If successfull, this automatically updates
`src/models/gw/mock/mockdata.json`.

### Note

You can always run the UI Style Guide server using `yarn stylegui`.

Finally, you can create a static version using `yarn styleguide:build`, that you
will find afterwards in `styleguide/`.
