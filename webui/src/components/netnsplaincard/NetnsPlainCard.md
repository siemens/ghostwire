```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { initialNetns } from "models/gw/mock";
import { firstContainee } from "models/gw";

const [navigatedTo, setNavigatedTo] = React.useState("");

<>
  <p>NAVIGATED TO: ⟨{navigatedTo}⟩</p>
  <ComponentCard>
    <NetnsPlainCard
      netns={initialNetns}
      onNavigation={(navnif) =>
        setNavigatedTo(`${navnif.name} [${firstContainee(navnif.netns).name}]`)
      }
    />
  </ComponentCard>
</>;
```
