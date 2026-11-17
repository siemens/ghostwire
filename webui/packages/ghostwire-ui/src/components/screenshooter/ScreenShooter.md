```tsx
import { useRef } from 'react'
import { Button, Box, Card } from "@mui/material";
import { SnackbarProvider } from 'notistack'
import {
  ScreenShooter,
  useScreenShooterModal,
} from "components/screenshooter";

const Component = () => {

  const setModal = useScreenShooterModal();

  return (
      <Button 
        variant="outlined" 
        color="primary" 
        onClick={() => setModal(true)}
      >
        Capture
      </Button>
  );
};

const Ex = () => {

  const someref = useRef();
  const setModal = useScreenShooterModal();

  return (
    <>
      <Box p={2} ref={someref}>
        <Card>
          <Box p={2}>FOO<b>BAR</b>!</Box>
        </Card>
      </Box>
      <Button 
        variant="outlined" 
        color="primary" 
        onClick={() => setModal(someref.current)}
      >
        Capture
      </Button>
    </>
  );
};

<SnackbarProvider maxSnack={3}>
  <ScreenShooter>
    <Ex/>
  </ScreenShooter>
</SnackbarProvider>
```
