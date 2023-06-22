# Ghostwire SPA Helper Servers

The Ghostwire server server is only required when working on and testing that a
production build has been correctly built to work with client-side URL routing.
Additionally, the rewriting proxy server is only needed when working on and
testing things related in the app to the `X-Forwarded-Uri` header. Most of the
time you thus won't need the servers in this directory.

## Fake Ghostwire Server

Make sure you have a Ghostwire service running on localhost port 5000, while in
the root repository directory for the Ghostwire UI run:

```bash
python3 -m fakegw.server
```

This will start a "fake" Ghostwire server listening on `[::]:5555`. It serves
the current static production build of the Ghostwire UI from the `build/`
directory (so don't forget to do a `yarn build` at any time so that there is
something to serve).

The JSON discovery information REST API endpoint is forwarded to the real
Ghostwire discovery service at `localhost:5000/json`.

## Fake Ingress Rewriting Proxy Server

To emulate and test serving the Ghostwire UI SPA correctly behind rewriting
proxy servers, also start:

```bash
python3 -m fakegw.ingress
```

This starts a minimalist rewriting reverse proxy on `[::]:5556`. It only serves
the path `/edgeshark/...` by forwarding it to the fake Ghostwire server on port
`:5555`, rewriting the URL paths by removing `/edgeshark/`. Most importantly, it
inserts an `X-Forwarded-Uri` HTTP header field which specifies the original
request URL as seen by this fake ingress proxy server.

Everything else will return an HTTP status 404.
