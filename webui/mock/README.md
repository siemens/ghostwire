# Generate Mock Data

This folder contains the parts which allow developers to create a "mock"
container and virtual networking system configuration, and then to take a
snapshot of mock data from it.

- `docker-compose.yaml`: creates/starts a set of containers, some of them with
  shared network namespaces. Additionally creates containers connected via
  MACVLAN interfaces to a host's Ethernet interface.

- `make up`: looks for an Ethernet network interface and then uses that as the
  "parent" with a Docker MACVLAN network driver. Starts a bunch of mock
  containers in different network and network namespace setups.

- `make kindup`: creates a new KinD Kubernetes-in-Docker test cluster consisting
  of one control plane node as well as one worker node.

- `make mockdata`: retrieves the discovery JSON data, cleans it, and writes the
  discovery data to `src/models/gw/mock`.

- `make down`: removes the mocked container and virtual communication system
  configuration.

- `make kinddown`: removes the KinD test cluster.
