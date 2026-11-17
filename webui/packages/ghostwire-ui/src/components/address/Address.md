Various addresses, IPv4, IPv6, and even some MAC-layer address; please note that
we've included an all-zero MAC address, which should not be rendered at all
(such all-zero MACs pop up in the context of loopback interfaces):

```tsx
import { Box, Typography } from "@mui/material";
import { ComponentCard } from "styleguidist/ComponentCard";
import { mockAddresses } from "models/gw/mock";
import { orderAddresses } from "models/gw";

<>
  {mockAddresses.sort(orderAddresses).map((addr, idx) => (
    <Box py={1} key={idx}>
      <Typography variant="body1" color="textSecondary">
        {addr.address}/{addr.prefixlen}
      </Typography>
      <ComponentCard paragraph>
        <Address address={addr} />
      </ComponentCard>
    </Box>
  ))}
</>;
```

In some contexts the address family icon will clutter and distration, so it can
be switched off using the `nofamilyicon` property. Hopefully, IKEA customer
relationship management won't object...

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { AddressFamily } from "models/gw";
import { nifLo } from "models/gw/mock";
const loaddr = nifLo.addresses.find(
  (addr) => addr.family === AddressFamily.IPv4
);

<ComponentCard>
  <Address address={loaddr} nofamilyicon />
</ComponentCard>;
```

The tooltip can optionally be disabled; this is used when displaying network
addresses as part of routes, where inner tooltips would be really vexating and
annoying.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { AddressFamily } from "models/gw";
import { nifLo } from "models/gw/mock";
const loaddr = nifLo.addresses.find(
  (addr) => addr.family === AddressFamily.IPv4
);

<ComponentCard>
  <Address address={loaddr} notooltip />
</ComponentCard>;
```
