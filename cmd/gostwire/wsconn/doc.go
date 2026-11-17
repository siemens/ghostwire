/*
Package wsconn wraps a server-side websocket connection with its own
human-readable unique ID. This helps to clearly map log debug and error
messages to their respective websocket connections, thus keeping them clearly
separated. Additionally, we also associate the capturing process (if any)
with this connection, so we can sanely manage it.

Websocket handshakes are only accepted from the same origin as the request's
host (or forwarded host), unless additional origins have been allowed using
[AllowOrigins]; the Gostwire service sets the allowed origins from its
“--allowed-origin” CLI flag. Additional HTTP endpoints registered by downstream
builds can thus use [NewWSConn] to upgrade their requests to websocket
connections with the same origin checks.
*/
package wsconn
