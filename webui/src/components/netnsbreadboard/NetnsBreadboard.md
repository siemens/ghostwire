This example has filtering enabled:

- filter out (hide) `lo` loopback network interfaces.
- filter out (hide) network namespaces with only a single, lonely loopback
  network interface.

Please note that this example filters out containers in order to reduce the
number of network namespaces rendered, except for a few white-listed
containers only and namespaces with only a loopback network interface. The
latter namespaces are then filtered out in the final rendering using the
`filterEmpty` property of the `NetnsBreadboard` component, as can be seen from
the number of network namespaces passed to the component versus the finally
rendered number of network namespaces.

A nice side effect of this test example is that it shows that the wire
rendering should correctly work with incomplete wiring, where some network
peer/master/... network interfaces in a related set of network interfaces
haven't been rendered.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { discovery } from "models/gw/mock";
import { orderNetnsByContainees } from "models/gw";

const netnses = Object.values(discovery.networkNamespaces)
  .sort(orderNetnsByContainees)
  .filter((netns) => {
    const nifcount = Object.values(netns.nifs).length;
    const whitelisted = !!["init", "ghostwire"].find((prefix) =>
      netns.containers[0].name.startsWith(prefix)
    );
    return nifcount === 1 || whitelisted;
  });

<>
  <p>passing {netnses.length} network namespaces before filtering.</p>
  <ComponentCard>
    <NetnsBreadboard netns={netnses} filterLo={true} filterEmpty={true} />
  </ComponentCard>
</>;
```
