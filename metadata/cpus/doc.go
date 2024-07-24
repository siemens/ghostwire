/*
Package cpus provides a metadata plugin returning information about the CPUs in
the system that are currently “online”.

The information about the CPUs being online is taken from
“/sys/devices/system/cpu/online” (a list of CPU ranges in textual format,
separated by commas).

The (JSON) metadata is structured (inside “metadata”) as follows:

  - “cpus”: slice of two-elements slice ranges.
*/
package cpus
