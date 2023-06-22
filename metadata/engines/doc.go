/*
Package engines implements a metadata plugin returning information about the
container engines for which containers were discovered. Container engines
without any workload won't be listed.

The (JSON) metadata is structured (inside “metadata”) as follows:

  - “container-engines”: object
  - “id”: string
  - “type”: string
  - “version”: string
*/
package engines
