```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { initialNetns } from "models/gw/mock";

<ComponentCard>
  {initialNetns.containers.map((cntr) => (
    <ContaineeDetails key={cntr.name} containee={cntr} />
  ))}
</ComponentCard>;
```

This component also supports filtering by IP address family in order to avoid
severe shocks with IPv6-adverse users.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { initialNetns } from "models/gw/mock";
import { AddressFamily } from "models/gw";

<ComponentCard>
  {initialNetns.containers.map((cntr) => (
    <ContaineeDetails key={cntr.name} containee={cntr} families={[AddressFamily.IPv4]} />
  ))}
</ComponentCard>;
```
