This example renders some `NifNavigator` components, including a MACVLAN
master interface as well as a few VETH peer interfaces. When clicking on a
(navigatable) network interface badge, a pop-up menu appears listing the
related network interfaces.

You'll most probably see several `eth0` network interfaces, which actually are
VETH interfaces ... rest asured, they belong to different containers, this
example just doesn't show the containers. When you click on such a network
interface badge, you'll see the related VETH peer interface and the container
it belongs to (usually `init (1)`).

When selecting one of the related network interfaces, the "NAVIGATED TO:"
information will update accordingly to show that the callback triggered
correctly and also got correct interface navigation information.

```tsx
import { ComponentCard } from "styleguidist/ComponentCard";
import { discovery, nifHw } from "models/gw/mock";
import { firstContainee, orderNifByName } from "models/gw";

// Pick up some network interfaces which we suppose have related network
// interfaces...
const nifs = [nifHw].concat(
  Object.values(discovery.networkNamespaces)
    .map((netns) => netns.nifs.filter((nif) => nif.kind === "veth"))
    .flat()
    .sort(orderNifByName)
    .slice(0, 3)
);

const [navigatedTo, setNavigatedTo] = React.useState("");

<>
  <p>NAVIGATED TO: {navigatedTo}</p>
  {nifs.map((nif, idx) => (
    <ComponentCard key={idx}>
      <NifNavigator
        nif={nif}
        onNavigation={(navnif) =>
          setNavigatedTo(`${navnif.name} [containee: ${firstContainee(navnif.netns).name}]`)
        }
      />
    </ComponentCard>
  ))}
</>;
```
