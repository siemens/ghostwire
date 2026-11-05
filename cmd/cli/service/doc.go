/*
Package service provides the “--http”, “--shutdown”, and “--allowed-origin”
CLI flags for running the Gostwire service.

The flags are automatically added to the root command as persistent flags. The
command then gets the flag values by calling [Address], [ShutdownTimeout], and
[AllowedOrigins]:

	addr := service.Address(cmd)
*/
package service
