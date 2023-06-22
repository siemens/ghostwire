### Badge Containee Types and States

These are examples of `ContaineeBadge`s for bind-mounted namespaces, stand-alone
processes attached to a namespace, and containers – in various states and also
with different flavors:

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { List, ListItem } from "@mui/material";
import { ContainerState, ContainerFlavors } from "models/gw";
import { discovery } from "models/gw/mock";
const containees = Object.values(discovery.networkNamespaces)
  .map((netns) => netns.containers)
  .flat()
  .filter(
    (cntr) => !cntr.name.startsWith("chrom") && !cntr.name.startsWith("firefox")
  )
  .slice(0, 4);
const morecontainees = [
  ...containees,
  {
    ...containees[3],
    state: ContainerState.Running,
    engineType: "docker",
    state: ContainerState.Paused,
  },
  {
    ...containees[3],
    state: ContainerState.Running,
    engineType: "docker",
    state: ContainerState.Exited,
  },
  {
    ...containees[3],
    name: "edge-iot-core",
    state: ContainerState.Running,
    engineType: "docker",
    flavor: ContainerFlavors.IERUNTIME,
  },
  {
    ...containees[3],
    name: "ieapp",
    state: ContainerState.Running,
    engineType: "docker",
    flavor: ContainerFlavors.IEAPP,
  },
];

<>
  {morecontainees.map((containee) => (
    <ComponentCard key={`${containee.name}-${containee.state}`}>
      <ContaineeBadge containee={containee} />
    </ComponentCard>
  ))}
</>;
```

Please notice how the tooltips reflect the container status as well as type of
container. The initial process with PID 1 also gets its own destinctive tooltip
description.

### Containee Badge Button

Additionally, a `ContaineeBadge` might act as a button referencing/linking to
the namespace UI element the containee is attached to. When acting as links,
`containeeedBadges` appear as flat UI buttons (including ripple effect, et
cetera), yet keep their status indication.

Use the `button` property (as well as an `onClick` callback handler) to enable
this aspect on a `ContaineeBadge`.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { crowdedNetns } from "models/gw/mock";
import { isContainer } from "models/gw/containee";
const containee = crowdedNetns.containers.find((containee) =>
  isContainer(containee)
);

<ComponentCard>
  <ContaineeBadge containee={containee} button onClick={() => null} />
</ComponentCard>;
```

### Capture Me If You Can

A containee badge optionally can show an additional capture button. The capture
button always appears to the right of the containee badge; it does not obey text
direction, as the badge text direction is ltr due to using ASCII latin glyphs
(even if intermediate APIs or transport protocols might use Unicode).

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { crowdedNetns } from "models/gw/mock";
import { isContainer } from "models/gw/containee";
const containee = crowdedNetns.containers.find((containee) =>
  isContainer(containee)
);

<ComponentCard>
  <ContaineeBadge containee={containee} capture />
</ComponentCard>;
```

### No Tooltip

And now without any tooltip, using `notooltip`, so no tooltip appears when
hovering with the mouse (or a pen) over the containee badge.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { crowdedNetns } from "models/gw/mock";
import { isContainer } from "models/gw/containee";
const containee = crowdedNetns.containers.find((containee) =>
  isContainer(containee)
);

<ComponentCard>
  <ContaineeBadge containee={containee} notooltip />
</ComponentCard>;
```
