# Example Downstream Module

A minimal example of a downstream Go module composing its own Gostwire service
from the upstream [`cmd/gostwire/app`](../../cmd/gostwire/app) package, adding:

- an additional decorator labelling all containers (`decorator/hello`),
- an additional HTTP endpoint `/api/v1/hello` (`endpoint`).

This module is built and smoke-tested by CI using the upstream `Dockerfile`;
please see [Downstream Distributions](../../docs/downstream.md) for how to
build it.
