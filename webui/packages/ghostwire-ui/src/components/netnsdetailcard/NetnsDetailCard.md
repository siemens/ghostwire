A single network namespace details card ... with lots of details, it seems.

```tsx
import { Provider } from 'jotai';
import { ComponentCard } from "styleguidist/ComponentCard";
import { initialNetns, crowdedNetns } from "models/gw/mock";

<Provider>
  <ComponentCard>
    <NetnsDetailCard netns={initialNetns} />
  </ComponentCard>
  <ComponentCard maxwidth="30em">
    <NetnsDetailCard netns={crowdedNetns} />
  </ComponentCard>
</Provider>
```
