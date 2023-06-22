```tsx
import { ComponentCard } from 'styleguidist/ComponentCard';
import { discovery } from 'models/gw/mock';
import { AddressFamily } from "models/gw";

const netns = Object.values(discovery.networkNamespaces)
    .find(netns => netns.containers.length > 1)

;<ComponentCard>
  <DetailedContainees netns={netns} families={[AddressFamily.IPv4]} />
</ComponentCard>
```
