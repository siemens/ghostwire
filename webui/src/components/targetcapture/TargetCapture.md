Small-sized and a medium-sized target capture buttons:

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { initialNetns } from "models/gw/mock";
import { firstContainee } from "models/gw";

// skip lo...
const somenif = Object.values(initialNetns.nifs).sort((nif) => nif.name)[1];

<>
  <ComponentCard>
    <TargetCapture nif={somenif} size="small" demo target={firstContainee} />
  </ComponentCard>
  <ComponentCard>
    <TargetCapture nif={somenif} size="medium" demo target={firstContainee} />
  </ComponentCard>
</>;
```
