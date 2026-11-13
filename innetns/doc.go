/*
Package innetns provides running a function in the context of a specific network
namespace. It hides the gory details of handling more complex cases, such as
process-less and purely bind-mounted network namespaces in other mount
namespaces, et cetera.
*/
package innetns
