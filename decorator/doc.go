/*
Package decorator defines the plugin interface for Gostwire decorators that
post-process the discovered Gostwire information model and add usefully
information. Gostwire decorators differ from lxkns decorators in that Gostwire
decorators get specific detail information about the discovered network
namespaces (such as network interfaces and their relationships, network
addresses, and much more) in addition to the containers discovered.

Please note that not all sub-packages are Gostwire decorators; for instance, the
“ieappicon” decorator is a plain lxkns decorator instead.

The plugin architecture allows to add more decorators easily, either to the
Gostwire code or in applications leveraging the Gostwire discovery engine.
*/
package decorator
