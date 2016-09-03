# Repository Integrated Licence Tracking Specification (RILTS)

The Repository Integrated Licence Tracking Specification (RILTS) is a
specification for the recording of open-source software licence information
inside version control repositories.

It is common for licences to be declared by placing a `LICENSE` or `COPYING`
file inside a repository. However, this purports to assign legal significance
to the mere adjacency of such a file to copyright-protected source code in a
repository, or adjacency and the use of such a particular filename. This is
rather ambiguous, and is risky, as it creates the possibility of a contributor
claiming that they were unaware of or did not notice the presence of this file.
It seems unclear that a user contributing to a repository can be held, legally,
to be licencing their contributions under a certain licence simply due to the
presence of such a file.

[relip](relip) is a proof tool which can be used to prove the licences of
repositories which use the RILTS specification. Currently, it supports Git
repositories only.

[**Read the RILTS specification.**](RILTS.md)
