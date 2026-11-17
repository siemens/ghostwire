/*
Package webui provides the Ghostwire web UI build as a file system, for
serving it using the WithUI option of the Gostwire service app package.

When built with the “webui” build tag, [FS] returns the web UI build embedded
from the “build” directory. Otherwise, [FS] returns the “webui/build” directory
relative to the current working directory, such as for development builds.
*/
package webui
