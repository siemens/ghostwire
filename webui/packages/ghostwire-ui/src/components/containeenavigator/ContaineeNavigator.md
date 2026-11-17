Thanks to our mock data, this containee navigator shows quite some stuff,
including Kubernetes pods, neatly grouped into namespaces.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { discovery } from "models/gw/mock";

;<ComponentCard>
    <ContaineeNavigator allnetns={discovery.networkNamespaces} filterEmpty />
</ComponentCard>
```
