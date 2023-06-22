This example shows the routes of the initial network namespace, that is, of
the host IP stack; please note that we don't show routing entries from the
so-called "local table" (number 255) on purpose in order to avoid clutter.

Just to show off in this example, we also sort the routing entries by address
family first, and then by destination address.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { initialNetns } from "models/gw/mock";
import { RouteTableLocal, orderRoutes } from "models/gw";

// As the <Route> component filters out local table routes, we would end up
// with lots of empty <ComponentCard>s; thus, we filter them out on the example
// level.
<>
  {initialNetns.routes
    .filter((route) => route.table !== RouteTableLocal)
    .sort(orderRoutes)
    .map((route, idx) => (
      <ComponentCard key={idx}>
        <Route route={route} />
      </ComponentCard>
    ))}
</>;
```
