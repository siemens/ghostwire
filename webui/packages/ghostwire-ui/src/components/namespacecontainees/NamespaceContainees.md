In this example there are several containers, processes, et cetera attached to
the same network namespace, hitting the display limit ingrained into the
`NamespaceContainees` component. The list is thus cut short and an ellipsis
shown to indicate that some inhabitants of the network namespace have been
left out.

Users can navigate from the list to the corresponding network namespace (card).
Clicking or tapping on a reference badge will scroll the network namespace into
view, if necessary.

```tsx
import { Provider } from 'jotai';
import { Box } from "@mui/material";
import { ComponentCard } from "styleguidist/ComponentCard";
import { NetnsDetailCard } from "components/netnsdetailcard";
import { crowdedNetns } from "models/gw/mock";

<Provider>
  <ComponentCard>
    <NamespaceContainees netns={crowdedNetns} />
  </ComponentCard>
  <Box style={{marginTop: '30ex'}}>
    <NetnsDetailCard netns={crowdedNetns} />
  </Box>
</Provider>;
```
