# Example Downstream Module

A minimal example of a downstream Go module composing its own Gostwire service
from the upstream [`cmd/gostwire/app`](../../cmd/gostwire/app) package, adding:

- an additional decorator labelling all containers (`decorator/hello`),
- an additional HTTP endpoint `/api/v1/hello`, as well as an override of the
  upstream `/json` endpoint (`endpoint`),
- a web app with its own `package.json`, MUI themes, and branding, using the
  `ghostwire-ui` library via `portal:` (`webui`).

CI builds the module with the upstream `Dockerfile` and smoke-tests the
resulting image, and also builds the web app; please see [Downstream
Distributions](../../docs/downstream.md) for how to build them.
