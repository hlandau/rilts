# Repository Integrated Licence Tracking Specification (RILTS)

The Repository Integrated Licence Tracking Specification (RILTS) is a specification
for the recording of open-source software licence information inside version control
repositories.

It is common for licences to be declared by placing a `LICENSE` or `COPYING`
file inside a repository. However, this purports to assign legal significance
to the mere adjacency of such a file to copyright-protected source code in a
repository, or adjacency and the use of such a particular filename. This is
rather ambiguous, and is risky, as it creates the possibility of a contributor
claiming that they were unaware of or did not notice the presence of this file.
It seems unclear that a user contributing to a repository can be held, legally,
to be licencing their contributions under a certain licence simply due to the
presence of such a file.

An affirmative declaration is highly desirable for the purpose of proving that
a contribution was licenced under a specific licence. There is precedent for this
in the form of the Linux Kernel project's “Developer Certificate of Origin”, in
which contributions must include a `Signed-off-by:` line.

However, this attempts to assign special meaning to the term `Signed-off-by` as
referring to a specific declaration hosted elsewhere. This is rather ambiguous,
and does not specify a specific licence.

A version control system tracks the history of software development as a series
of contributions, and in this regard it makes a great deal of sense to track
contribution licencing in the version control system, via commit messages, than
as files in the repository. Because proof of licencing — in the form of an
explicit declaration — is tracked by commit, the tree of commits leading to a
certain repository state can be proven back to the initial commit by an
automated tool, so long as these declarations are in a machine-readable form.

RILTS provides a specification for standard licencing language in commits. The
precise wording specified is intended to be a legally unambiguous declaration,
but because precise wording is used, these legal incantations can be identified
automatically, allowing the licence status of a repository tip to be proven
automatically. This provides much higher degrees of licence assurance than
conventional methods, such as `COPYING` files or ambiguous `Signed-off-by`
lines. Because this proof is automated, it can be re-run automatically, and
thus made the subject of unit tests. This provides high and continuous
assurance of licencing purity.

A distinction is sometimes made between the author of a contribution and its
committer. The author is the entity which must issue a declaration.

## Specification

A commit is compliant with RILTS if the commit message includes a RILTS stanza.

A RILTS stanza is a contiguous series of lines each beginning with `©!` or
`©:`.

This prefix identifies the text as containing licencing information, allowing
it to be extracted for the commit for further verification.

A RILTS incantation is the text derived by taking the stanza, stripping the
prefix, stripping leading and trailing whitespace from each line, then joining
the lines with a single space separating each of them.

Multiple RILTS stanzas should be separated by one or more blank lines.
The preferred location for a RILTS stanza is at the bottom of the commit text.

### Person Specification

The keyword `PERSON` in the standard incantations below should be replaced with
a person specification. This is a person's full name, optionally followed by
their e. mail address in `<>`, i.e. the standard Git style:

    John Smith <jsmith@example.com>     (preferred)
    John Smith
    Foocorp Inc. <info@example.com>     (preferred)
    Foocorp Inc.

`PERSON` may differ from both the author and committer if the copyright holder
is an organization, and the author is submitting the contribution on their
behalf.

### Person List Specification

The keyword `PERSON-LIST` must follow one of the following formats:

    PERSON
    PERSON and PERSON
    1*(PERSON, )PERSON[,] and PERSON

### Standard Incantations

#### Standard Single-Commit Licencing Declaration

The author should be explicitly specified:

    ©! [I/We], PERSON,
    ©! hereby licence these changes under the licence with SHA256 hash
    ©! xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.

As an older, deprecated format, and only when the committer and author are the
same:

    ©! I hereby licence these changes under the licence with SHA256 hash
    ©! xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.

#### Standard Retroactive Licencing Declaration v2

This declaration should be used when adopting RILTS in an existing repository.

**Standard Retroactive Licencing Completeness Declaration**. Firstly, for every
historical author of a commit leading up to the commit in question, the
following stanza should be specified:

    ©! As regards this commit, and all commits upon which this commit depends,
    ©! PERSON hereby declares that no entity other than PERSON-LIST has a copyright
    ©! interest in any such commit (and the changes therein) authored by their person.

`PERSON-LIST` in this context may contain the special value `their person`, which is
equivalent to the `PERSON` specified.

**Standard Retroactive Licencing Entity Declaration**. Secondly, for every
person listed in any `PERSON-LIST` in any such stanza, the following must be
specified as a RILTS stanza:

    ©! To the extent that [I/we], PERSON, have a copyright interest in the changes
    ©! in this commit, and the changes in all commits upon which this commit depends,
    ©! including changes occluded by subsequent changes, [I/we] hereby licence those
    ©! changes under the copyright licence with SHA256 hash
    ©! xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.

#### Standard Retroactive Licencing Declaration v1 (Deprecated)

This should not be used in new commits.

    ©! To the extent that I have a copyright interest in the files
    ©! in this repository, and the sequence of changes leading to those files,
    ©! and all intermediate states resulting from a partial application of those
    ©! changes, including changes occluded by subsequent changes, I hereby licence
    ©! those files and changes present and past under the copyright licence with
    ©! SHA256 hash
    ©! xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.

### Proof Process

To prove the licencing of a specific commit X, perform the following process:

  1. Let C := X. Let R be a list of zero or more `PERSON`, initially empty.
     Let L be a list of SHA256 licence hashes, initially empty.
  2. Examine the commit message for C. Find a RILTS stanza and check it against
     a database of valid phraseology. From this, determine whether the stanza
     expresses a single-commit declaration, a retroactive licencing
     completeness declaration, or no (or an unrecognized or retroactive
     licencing entity) declaration. If there is a retroactive licencing
     completeness declaration, check that there are retroactive licencing
     entity declarations for all persons listed in the retroactive licencing
     completeness declaration. Otherwise, the retroactive licencing
     completeness declaration is ignored. Note the SHA256 hash indicated.
  3. If the commit has no recognized declaration:
     1. Check whether the author of the commit is in R.
     2. If such a person is not found, stop with error.
  4. If the commit has a retroactive licencing completeness declaration, add
     every person in the person list in the declaration to R.
  5. Add every SHA256 hash in every licence declaration in the commit to L.
  6. Let C be the parent commit of C. If C is the initial commit, stop with
     success and output L.
  7. Goto 2.

After running this process, L should be verified as containing only licence
hashes authorized by the project.

Note that this proof process does not work when retroactive licencing is used
and one desires to prove the licencing of a commit chain terminating in a
commit prior to an applicable retroactive licencing declaration. The process
could be adapted for this but this would be more complicated. However, the
process above can be trivially modified to output a list of affected commits,
and then run on a later commit which includes the retroactive declaration in
its heritage. The commit for which proof is desired can then be checked against
the list of commit IDs.

### Well-Known Licence Hashes

Licences are identified by SHA256 hash. A selection of preferred licence texts
embodying common licences are stored in the [RILTS
repository](https://github.com/hlandau/rilts/tree/master/licences). Although
there are multiple possible text files which can embody the same licence text,
this would need to needless proliferation of different hashes. Therefore, it is
strongly preferred to use the licence files in this directory if there exists
one matching the licence you wish to use.

To derive the SHA256 sum for such a file, run `sha256sum` on it.

**WARNING:** On Windows, Git automatically changes line endings to CRLF. The
licence files stored in this repository use UNIX-style LF line endings and
should be hashed as such. A `.gitattributes` file is used in this repository to
disable this behaviour, so you shouldn't have any issues with this, but you
should keep this in mind if storing such files in other repositories where
their hash-invariance is important.
