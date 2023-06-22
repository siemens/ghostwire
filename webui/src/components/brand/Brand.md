```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { DynVarsProvider, useDynVars } from "components/dynvars";

<>
  <ComponentCard>
    Default brand: "<Brand />".
  </ComponentCard>
  <DynVarsProvider value={{ brand: "Edgeshark" }}>
    <ComponentCard>
      Rebranded: "<Brand />".
    </ComponentCard>
  </DynVarsProvider>
</>;
```
