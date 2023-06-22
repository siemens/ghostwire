Please note that the address family icon is centered vertically with respect to
the font's baseline, as MAC, IPv4 and IPv6 addresses don't contain any glyphs
with descenders.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { nifHw } from "models/gw/mock";

<ComponentCard>
  <NifAddressList nif={nifHw} />
</ComponentCard>;
```
