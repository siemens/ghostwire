# Testing

Only for the first time, you need to install the required `nerdctl` and CNI
plugin binaries into your development system or CI VM:

```bash
make install-tools
```

To run the full test suite, once as root and once as your ordinary user, do:

```bash
make testfull
```

Please note that some tests may take some time, especially when creating a
Kubernetes test cluster in Docker.
