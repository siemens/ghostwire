/*
Package appicons provides the “--prefer-showing-app-icons” CLI flag for
defaulting the Gostwire web UI to showing Industrial Edge app icons (and
titles), when available.

The flag is automatically added to the root command as a persistent flag. The
command then gets the flag value by calling [PreferShowing]:

	prefer := appicons.PreferShowing(cmd)

Please note that this flag only sets the default in the web UI: an explicit
user setting stored in the browser always takes precedence.
*/
package appicons
