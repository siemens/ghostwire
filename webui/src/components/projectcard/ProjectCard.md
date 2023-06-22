Docker composer projects group sets of containers. This grouping is indicated
using `Project` components. The avatar icon depends on the flavor of the
project.

```tsx
import { ComponentCard } from 'styleguidist/ComponentCard';
import { NetnsPlainCard } from 'components/netnsplaincard';
import { ContainerState, ContainerFlavors } from 'models/gw';
import { discovery } from 'models/gw/mock';
const containees = Object.values(discovery.networkNamespaces)
  .map(netns => netns.containers)
  .flat()
  .filter(cntr => !cntr.name.startsWith('chrom') && !cntr.name.startsWith('firefox'))
  .slice(0, 4)
const ieproject = {
  name: 'myieapp',
  flavor: ContainerFlavors.IEAPP,
}
const projcontainer = { 
  ...containees[3], 
  name: 'composerapp_service_1', 
  state: ContainerState.Running, 
  engineType: 'docker', 
  flavor: 'docker', 
  labels: { 'com.docker.compose.project': 'fooproject' } 
}
ieproject.containers = [
  { 
    ...containees[3], 
    name: 'myieapp_foobar_1', 
    state: ContainerState.Running, 
    engineType: 'docker', 
    flavor: ContainerFlavors.IEAPP, 
    labels: { 'com.docker.compose.project': 'ieappproject' } 
  },
  projcontainer,
]

;<>
  <ComponentCard>
    <ProjectCard project={ieproject}>
      {ieproject.containers.map((cntr, idx) => 
        <NetnsPlainCard key={idx} netns={cntr.netns} />
      )}
    </ProjectCard>
  </ComponentCard>
</>
```
