### Operational Status

The operational status of a network interface is rendered inside the badge using
an icon. As an additional visual clue, the status icon is colored according to
state. Please note that the "unknown" status is rendered the same as the "up"
state: "unknown" means that the corresponding network interface driver doesn't
report any status, yet the interface the interface is not considered to be down.

```tsx
import { Box } from "@mui/material";
import { ComponentCard } from "styleguidist/ComponentCard";
import { OperationalState } from "models/gw";
import { nifHw } from "models/gw/mock";
import { DynVarsProvider } from "components/dynvars";

let nifs = [
  OperationalState.Unknown,
  OperationalState.Dormant,
  OperationalState.Down,
  OperationalState.LowerLayerDown,
  OperationalState.Up,
].map((opstat) => ({
  ...nifHw,
  operstate: opstat,
  isPhysical: false,
  isPromiscuous: false,
  macvlans: undefined,
}));

<DynVarsProvider value={{}}>
  {nifs.map((nif, idx) => (
    <div key={idx}>
      <p>{nif.operstate}:</p>
      <ComponentCard>
        <NifBadge nif={nif} />
      </ComponentCard>
    </div>
  ))}
</DynVarsProvider>;
```

### Physical Network Interface

"Physical" network interfaces are designed by a "NIC" icon preceeding the badge.

```tsx
import { Box } from "@mui/material";
import { ComponentCard } from "styleguidist/ComponentCard";
import { OperationalState } from "models/gw";
import { nifHw } from "models/gw/mock";
import { DynVarsProvider } from "components/dynvars";

let nif = {
  ...nifHw,
  isPhysical: true,
  isPromiscuous: false,
  macvlans: undefined,
};

<DynVarsProvider value={{}}>
  <ComponentCard>
    <NifBadge nif={nif} />
  </ComponentCard>
</DynVarsProvider>;
```

### Promiscuous Mode

A network interface put into promiscuous mode displays a lughole (eavesdropper)
icon.

```tsx
import { Box } from "@mui/material";
import { ComponentCard } from "styleguidist/ComponentCard";
import { OperationalState } from "models/gw";
import { nifHw } from "models/gw/mock";
import { DynVarsProvider } from "components/dynvars";

let nif = {
  ...nifHw,
  isPhysical: true,
  isPromiscuous: true,
  macvlans: undefined,
};

<DynVarsProvider value={{}}>
  <ComponentCard>
    <NifBadge nif={nif} />
  </ComponentCard>
</DynVarsProvider>;
```

### Type of Network Interface

The type of network interface is indicated by an icon inside badge.

```tsx
import { Box } from "@mui/material";
import { ComponentCard } from "styleguidist/ComponentCard";
import { OperationalState } from "models/gw";
import { nifHw } from "models/gw/mock";
import { DynVarsProvider } from "components/dynvars";

let nifs = [
  ["", "hardware interface"],
  ["lo", "loopback"],
  ["bridge", "virtual bridge"],
  ["veth", "VETH"],
  ["macvlan", "MACVLAN"],
].map((meta) => ({
  ...nifHw,
  kind: meta[0],
  description: meta[1],
  isPhysical: false,
  isPromiscuous: false,
  macvlans: undefined,
}));
nifs.push({
  ...nifs[0],
  macvlans: [nifs[0]],
  description: "(hardware network interface) MACVLAN master",
});

<DynVarsProvider value={{}}>
  {nifs.map((nif, idx) => (
    <div key={idx}>
      <p>{nif.description}:</p>
      <ComponentCard>
        <NifBadge nif={nif} />
      </ComponentCard>
    </div>
  ))}
</DynVarsProvider>;
```

### Etc.

In this example, we first render a set of `NifBadge`s, which double as
navigation targets (anchors). We then add a `NifBadge` "button" linking to the
"enp3s0" network interface badge above. When clicking on this button, the
network interface badge for "enp3s0" should (smoothly) scroll into view, if
necessary. For best effect, scroll the set of network interfaces out of view so
only the final hyperlink button is visible and then click on it.

```tsx
import { Box } from "@mui/material";
import { ComponentCard } from "styleguidist/ComponentCard";
import { initialNetns, nifHw, nifMacvlan } from "models/gw/mock";
import { DynVarsProvider } from "components/dynvars";

const nifs = [...initialNetns.nifs, nifMacvlan];

<DynVarsProvider value={{}}>
  {nifs.map((nif, idx) => (
    <ComponentCard key={idx}>
      <NifBadge nif={nif} />
    </ComponentCard>
  ))}
  <Box style={{ marginTop: "30ex" }}>
    <NifBadge nif={nifHw} button />
  </Box>
</DynVarsProvider>;
```

### Stretchy Badge

A network interface badge optionally can stretch in order to take up the
available horizontal space; this requires the `stretch` property, but also
suitable element width setting, as otherwise the `NifBadge` component won't
stretch.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { nifHw } from "models/gw/mock";
import { DynVarsProvider } from "components/dynvars";

<DynVarsProvider value={{}}>
  <ComponentCard>
    <NifBadge style={{ width: "15em" }} nif={nifHw} stretch />
  </ComponentCard>
  <ComponentCard>
    <NifBadge style={{ width: "15em" }} nif={nifHw} />
  </ComponentCard>
</DynVarsProvider>;
```

### No Tooltip

In some contexts where an outer component already provides a tooltip, you
might not want to have an additional tooltip on the network interface badge,
as this quickly gets confusing. Use `notooltip` to switch off the tooltip on a
network interface badge.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { nifHw } from "models/gw/mock";
import { DynVarsProvider } from "components/dynvars";

<DynVarsProvider value={{}}>
  <ComponentCard>
    <NifBadge nif={nifHw} notooltip />
  </ComponentCard>
</DynVarsProvider>;
```

### Capture Links

Additional network interface capture buttons need to be requested through the
`capture` property and then are enabled only through the dynamic variable
`enableMonolith`, passed in by the server serving the application (see also
[DynVarsProvider](#dynvars)). Depending on the alignment of the badge, the
capture button will be shown either to the right or left of the badge.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { nifHw } from "models/gw/mock";

<>
  <ComponentCard>
    <NifBadge nif={nifHw} capture notooltip />
  </ComponentCard>
  <ComponentCard>
    <NifBadge nif={nifHw} capture alignRight notooltip />
  </ComponentCard>
</>;
```

### Plain Badge Versus Button Badge

A plain network interface badge side-by-side to a button badge. The button
variant has rounded edges and reacts to hovering the mouse over it or clicking
it, not least with a ripple animation.

```tsx
import { nifHw } from "models/gw/mock";
import { DynVarsProvider } from "components/dynvars";

<DynVarsProvider value={{}}>
  <NifBadge nif={nifHw} /> ⟵ versus ⟶ <NifBadge nif={nifHw} button />
</DynVarsProvider>;
```
