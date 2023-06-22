For network interfaces with related interfaces, (only) these related network
interfaces are rendered. If the related network interface is in a different
network namespace, then additional information about the related network
namespace is rendered in form of the entities (containers, stand-alone
processes) "inside" the related namespace.

```tsx
import { ComponentCard } from 'styleguidist/ComponentCard';
import { nifVeth1, nifVeth2, nifMacvlan } from 'models/gw/mock';

<>{[nifVeth1, nifVeth2, nifMacvlan].map((nif, idx) =>
    <ComponentCard key={idx}>
        <RelatedNif nif={nif} />
    </ComponentCard>
)}</>
```

**Nothing gets rendered** for network interfaces without any related interfaces,
such as hardware interfaces without attached macvlan interfaces, the loopback
interface, et cetera.

```tsx
import { ComponentCard } from 'styleguidist/ComponentCard';
import { nifHw } from 'models/gw/mock';

<>
    <ComponentCard>
        <RelatedNif nif={nifHw} />
    </ComponentCard>
</>
```
