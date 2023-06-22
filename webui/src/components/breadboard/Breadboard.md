This example puts all the discovered network namespaces on the `BreadBoard` and
wires up the network interfaces. Hover over network interface badges or the
wires between them to enjoy some wire animation...

```tsx
import { Box } from "@mui/material";
import { ComponentCard } from "styleguidist/ComponentCard";
import { NetnsPlainCard } from "components/netnsplaincard";
import { CardTray } from 'components/cardtray';
import { useContextualId } from 'components/idcontext';
import { discovery } from "models/gw/mock";
import { orderNetnsByContainees } from "models/gw";

;<ComponentCard>
  <Breadboard netns={discovery.networkNamespaces}>
    <CardTray animate>
      {Object.values(discovery.networkNamespaces).sort(orderNetnsByContainees).map(netns =>
        <NetnsPlainCard netns={netns} key={netns.netnsid} />)}
    </CardTray>
  </Breadboard>
</ComponentCard>
```
