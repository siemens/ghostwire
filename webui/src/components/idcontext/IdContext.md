This example demonstrates how to introduce separate DOM element identifier
contexts, or prefixes. The example wrapper used in this style guide
automatically establishes identifier contexts for each example individually.
However, in your own applications you need to establish different contexts
"manually" as required.

```tsx
import { Box, Grid, Card } from "@mui/material";
import { useContextualId } from "components/idcontext";

const ShowIdContext = ({ id }) => {
  const uid = useContextualId(id);
  return <div>DOM element identifier is "{uid}".</div>;
};

<Grid container direction="column" p={2}>
  <Box m={1} p={1}>
    <Card>
      <IdContext>
        <ShowIdContext id="A" />
        <ShowIdContext id="B" />
      </IdContext>
    </Card>
  </Box>
  <Box m={1} p={1}>
    <Card>
      <IdContext>
        <ShowIdContext id="A" />
      </IdContext>
    </Card>
  </Box>
</Grid>;
```
