In most cases, there's no need for a `DynVarsProvider` unless you want to
override the dynamic variables passed in to your application by the server.
Point in case is this stylebook where we enable capture links; the following
example is to show a working principle.

```tsx
import { DynVarsProvider, useDynVars } from "components/dynvars";

const FooBar = ({ text }) => {
  const dynvars = useDynVars();
  return (
    <div>
      {text}: {dynvars.foobar || <em>nothing</em>}
    </div>
  );
};

<>
  <FooBar text="default dynvars" />
  <DynVarsProvider value={{ foobar: "hooray!" }}>
    <FooBar text="locally overriden dynvars" />
  </DynVarsProvider>
</>;
```
