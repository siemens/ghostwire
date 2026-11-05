/*
Package brand provides the “--brand” and “--brandicon” CLI flags for branding
the Gostwire web UI.

The flags are automatically added to the root command as persistent flags. The
command then gets the brand name and icon by calling [Name] and [Icon]:

	name := brand.Name(cmd)
	icon := brand.Icon(cmd)
*/
package brand
