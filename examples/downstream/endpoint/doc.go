/*
Package endpoint provides an example additional HTTP endpoint
“/api/v1/hello” that runs a Gostwire discovery and returns the number of
containers found. Additionally, it overrides the upstream “/json” endpoint,
adding a response header before calling the upstream handler.
*/
package endpoint
