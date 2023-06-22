/*
Package ieappicon implements a Gostwire decorator for adorning containers
belonging to Industrial Edge app composer projects with their app icons.

This decorator retrieves the app icons from the local IE edge core's app engine
data base. It caches project icons and only looks for potential project updates
whenever a new container with an associated composer project gets started. This
cache additionally caches negative lookups.
*/
package ieappicon
