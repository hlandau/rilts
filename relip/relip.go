package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/hlandau/xlog"
	"github.com/libgit2/git2go"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var log, Log = xlog.New("relip")

var (
	repoArg            = kingpin.Arg("repository", "repository path").Default(".").ExistingDir()
	licencePathFlag    = kingpin.Flag("licence-path", "path to directory containing licence files").ExistingDir()
	allowedLicenceFlag = kingpin.Flag("licence", "identifier for an allowed licence (may be specified multiple times)").Short('L').Strings()
	branchFlag         = kingpin.Flag("branch", "refspec to check").Short('B').Default("HEAD").String()
)

var reSHA256 = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)

func main() {
	kingpin.Parse()

	allowedLicences := map[string]struct{}{}

	if *licencePathFlag == "" {
		gopath := os.Getenv("GOPATH")
		if gopath != "" {
			p := filepath.Join(gopath, "src/github.com/hlandau/rilts/licences")
			if _, err := os.Stat(p); err == nil {
				*licencePathFlag = p
			}
		}
	}

	if *licencePathFlag == "" {
		log.Fatalf("must specify licence path")
	}

	// Load allowed licences.
	for _, L := range *allowedLicenceFlag {
		if reSHA256.MatchString(L) {
			allowedLicences[L] = struct{}{}
		} else {
			p := filepath.Join(*licencePathFlag, "COPYING."+strings.ToUpper(L))
			b, err := ioutil.ReadFile(p)
			log.Fatale(err, "could not find licence: ", L)
			h := sha256.New()
			h.Write(b)
			sum := h.Sum(nil)
			Ls := hex.EncodeToString(sum)
			allowedLicences[Ls] = struct{}{}
		}
	}

	var err error
	*repoArg, err = filepath.Abs(*repoArg)
	log.Fatale(err, "cannot absolutize repository path")

	_, err = os.Stat(filepath.Join(*repoArg, ".git"))
	if err == nil {
		*repoArg = filepath.Join(*repoArg, ".git")
	}

	// Open repository.
	repo, err := git.OpenRepository(*repoArg)
	log.Fatale(err, "couldn't open repository: ", *repoArg)

	obj, err := repo.RevparseSingle(*branchFlag) //repo.Head() //ref, err := repo.LookupReference(*branchFlag)
	log.Fatale(err, "couldn't find ref - are there any commits?: ", *branchFlag)

	c, err := obj.AsCommit()
	log.Fatale(err, "refspec does not specify a commit")

	var errors []error

	// Check commits.
	err = checkCommit(c, allowedLicences, &errors)
	log.Fatale(err)

	if len(errors) > 0 {
		log.Errorf("There were %v errors:", len(errors))
		for _, e := range errors {
			log.Errorf("  %v", e)
		}

		os.Exit(1)
	}
}

func checkCommit(c *git.Commit, allowedLicences map[string]struct{}, errors *[]error) error {
	retroPersons := map[string]struct{}{}

	for {
		err := checkCommitLocal(c, retroPersons, allowedLicences, errors)
		if err != nil {
			return err
		}

		if c.ParentCount() == 0 {
			break
		}

		c = c.Parent(0)
	}

	return nil
}

func checkCommitLocal(c *git.Commit, retroPersons, allowedLicences map[string]struct{}, errors *[]error) error {
	msg := c.Message()
	stanzas, err := extractStanzas(msg)
	if err != nil {
		return err
	}

	type rrp struct {
		Person string
		People []string
	}

	var requiredRetroPeople []rrp
	requiredRetroPeopleM := map[string]struct{}{}
	dRetroPeople := map[string]struct{}{}

	hasAllowedHash := false
	for _, s := range stanzas {
		for _, ks := range knownStanzas {
			m := ks.Regexp.FindStringSubmatch(s)
			if m == nil {
				continue
			}

			switch ks.Type {
			case STCurrent:
				hashIdx := matchByName(ks.Regexp, "hash")
				//personIdx := matchByName(ks.Regexp, "person")

				hash := strings.ToLower(m[hashIdx])
				if len(hash) != 64 {
					log.Warnf("invalid hash: %s", hash)
					continue
				}

				_, ok := allowedLicences[hash]
				if ok {
					hasAllowedHash = true
				}

			case STRetroactiveCompleteness:
				personIdx := matchByName(ks.Regexp, "person")
				personsIdx := matchByName(ks.Regexp, "persons")

				persons, err := parsePersons(m[personsIdx], m[personIdx])
				if err != nil {
					log.Warnf("couldn't parse persons list: %v", err)
					continue
				}
				requiredRetroPeople = append(requiredRetroPeople, rrp{m[personIdx], persons})
				if _, ok := requiredRetroPeopleM[m[personIdx]]; ok {
					log.Warnf("repeated retroactive licencing completeness declaration for person %s", m[personIdx])
					continue
				}
				requiredRetroPeopleM[m[personIdx]] = struct{}{}

			case STRetroactiveEntity:
				personIdx := matchByName(ks.Regexp, "person")
				hashIdx := matchByName(ks.Regexp, "hash")
				if personIdx < 0 {
					continue
				}

				hash := strings.ToLower(m[hashIdx])
				if len(hash) != 64 {
					log.Warnf("invalid hash: %s", hash)
					continue
				}

				_, ok := allowedLicences[hash]
				if !ok {
					continue
				}

				dRetroPeople[m[personIdx]] = struct{}{}

			default:
				panic("unreachable")
			}
		}
	}

L:
	for _, r := range requiredRetroPeople {
		for _, p := range r.People {
			_, ok := dRetroPeople[p]
			if !ok {
				continue L
			}
		}

		retroPersons[r.Person] = struct{}{}
	}

	if !hasAllowedHash {
		if checkSig(c.Author(), retroPersons) {
			hasAllowedHash = true
		}
	}

	if !hasAllowedHash {
		if !isTrivial(c) {
			*errors = append(*errors, fmt.Errorf("licence unprovable: %v (%v <%v>)", c.Id().String(), c.Author().Name, c.Author().Email))
		}
	}

	return nil
}

func isTrivial(c *git.Commit) bool {
	tr, err := c.Tree()
	if c.ParentCount() == 0 {
		return false
	}

	pc := c.Parent(0)
	log.Fatale(err)

	ptr, err := pc.Tree()
	log.Fatale(err)

	do := &git.DiffOptions{
		Flags:            git.DiffNormal,
		IgnoreSubmodules: git.SubmoduleIgnoreNone,
		Pathspec:         nil,
		NotifyCallback:   nil,
		MaxSize:          0,
		OldPrefix:        "",
		NewPrefix:        "",
	}

	diff, err := c.Owner().DiffTreeToTree(tr, ptr, do)
	log.Fatale(err, "diff")

	numAdded := 0

	defer diff.Free()
	err = diff.ForEach(func(delta git.DiffDelta, x float64) (git.DiffForEachHunkCallback, error) {
		return func(hunk git.DiffHunk) (git.DiffForEachLineCallback, error) {
			return func(line git.DiffLine) error {
				switch line.Origin {
				case git.DiffLineAddition:
					numAdded += line.NumLines
				default:
				}
				return nil
			}, nil
		}, nil
	}, git.DiffDetailLines)
	log.Fatale(err, "diff iterate")

	return numAdded <= 3
}

func checkSig(sig *git.Signature, allowed map[string]struct{}) bool {
	_, ok := allowed[sig.Name]
	if ok {
		return true
	}

	_, ok = allowed[sig.Name+" <"+sig.Email+">"]
	if ok {
		return true
	}

	return false
}

func matchByName(r *regexp.Regexp, n string) int {
	ns := r.SubexpNames()
	for i, x := range ns {
		if x == n {
			return i
		}
	}
	return -1
}

type StanzaType int

const (
	STCurrent StanzaType = iota
	STRetroactiveEntity
	STRetroactiveCompleteness
)

type KnownStanza struct {
	Regexp *regexp.Regexp
	Type   StanzaType
}

var knownStanzas = []*KnownStanza{
	// Standard Single-Commit Licencing Declaration
	&KnownStanza{
		Regexp: regexp.MustCompile(`^(I|We)(, (?P<person>.+),)? hereby licen[cs]e these changes under the licen[cs]e with SHA256 hash (?P<hash>[a-fA-F0-9]{64}).$`),
		Type:   STCurrent,
	},

	// Standard Retroactive Licencing Completeness Declaration
	&KnownStanza{
		Regexp: regexp.MustCompile(`^As regards this commit, and all commits upon which this commit depends, (?P<person>.+) hereby declares that no entity other than (?P<persons>.+) has a copyright interest in any such commit \(and the changes therein\) authored by their person.$`),
		Type:   STRetroactiveCompleteness,
	},

	// Standard Retroactive Licencing Entity Declaration
	&KnownStanza{
		Regexp: regexp.MustCompile(`^To the extent that (I|we), (?P<person>.+), have a copyright interest in the changes in this commit, and the changes in all commits upon which this commit depends, including changes occluded by subsequent changes, (I|we) hereby licence those changes under the copyright licence with SHA256 hash (?P<hash>[0-9a-fA-F]{64}).$`),
		Type:   STRetroactiveEntity,
	},

	// Standard Retroactive Licencing Entity Declaration v1 (Deprecated)
	&KnownStanza{
		Regexp: regexp.MustCompile(`^To the extent that I have a copyright interest in the files in this repository, and the sequence of changes leading to those files, and all intermediate states resulting from a partial application of those changes, including changes occluded by subsequent changes, I hereby licence those files and changes present and past under the copyright licence with SHA256 hash (?P<hash>[a-fA-F0-9]{64}).$`),
		Type:   STRetroactiveEntity,
	},
}

var reStanza = regexp.MustCompile(`^©[:!] *([^\r\n]*) *(\r\n|\n)?$`)

func extractStanzas(msg string) ([]string, error) {
	var stanzas []string

	var s string
	rr := strings.NewReader(msg)
	r := bufio.NewReader(rr)
	for {
		L, err := r.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		m := reStanza.FindStringSubmatch(L)
		if m == nil {
			if len(s) > 0 {
				stanzas = append(stanzas, s)
				s = ""
			}
			continue
		}

		if len(s) > 0 {
			s += " "
		}
		s += strings.Trim(m[1], " \t\r\n")
	}

	if len(s) > 0 {
		stanzas = append(stanzas, s)
	}

	return stanzas, nil
}

var rePersonList = regexp.MustCompile(`(,? and |, )`)

// John Smith <jsmith@example.com>
// John Smith <jsmith@example.com>, John Smith <jsmith@example.com>
// John Smith <jsmith@example.com> and John Smith <jsmith@example.com>
// John Smith <jsmith@example.com>, John Smith <jsmith@example.com> and John Smith <jsmith@example.com>
func parsePersons(persons, theirPerson string) ([]string, error) {
	L := rePersonList.Split(persons, -1)
	for idx := range L {
		if L[idx] == "their person" && theirPerson != "" {
			L[idx] = theirPerson
		}
	}
	return L, nil
}
