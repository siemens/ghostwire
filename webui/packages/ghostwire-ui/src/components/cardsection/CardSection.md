Unless specified otherwise, a card section shows its caption as well its
children.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { Box, Card } from "@mui/material";

<ComponentCard>
  <Card>
    <Box p={2} width="20em" style={{textTransform: 'uppercase'}}>
      <CardSection caption="dixit">
        lorem ipsum dolor sit amet, consetetur sadipsicing eliter, sed superbus
        esse semper latinum labere.
      </CardSection>
    </Box>
  </Card>
</ComponentCard>;
```

However, card section can be made collapsible/expandable by specifying their
intial `collapsible=` state `expand` or `collapse` (but not `never`). Users
then can collapse or expand a section as to their hearts' desire.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { Box, Card } from "@mui/material";

<ComponentCard>
  <Card>
    <Box p={2} width="20em" style={{textTransform: 'uppercase'}}>
      <CardSection caption="dixit" collapsible="expand">
        lorem ipsum dolor sit amet, consetetur sadipsicing eliter, sed superbus
        esse semper latinum labere.
      </CardSection>
    </Box>
  </Card>
</ComponentCard>;
```
