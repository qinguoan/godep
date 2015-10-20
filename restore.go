package main

import (
	"log"
)

var cmdRestore = &Command{
	Usage: "restore [-v]",
	Short: "check out listed dependency versions in GOPATH",
	Long: `
Restore checks out the Godeps-specified version of each package in GOPATH.

If -v is given, verbose output is enabled.
`,
	Run: runRestore,
}

func init() {
	cmdRestore.Flag.BoolVar(&verbose, "v", false, "enable verbose output")
}

func runRestore(cmd *Command, args []string) {
	g, err := loadDefaultGodepsFile()
	if err != nil {
		log.Fatalln(err)
	}
	hadError := map[Dependency]bool{}
	for _, dep := range g.Deps {

		err := download(dep)
		if err != nil {
			hadError[dep] = true
		}
	}
	for _, dep := range g.Deps {
		err := restore(dep)
		if err != nil && hadError[dep] {
			log.Printf("download or restore: %v err: %v", dep.ImportPath, err)
			hadError[dep] = true
		} else {
			log.Printf("restore: %v OK", dep.ImportPath)
		}
	}
}

// download downloads the given dependency.
func download(dep Dependency) error {
	// make sure pkg exists somewhere in GOPATH

	args := []string{"get", "-d"}
	if verbose {
		args = append(args, "-v")
	}

	return runIn(".", "go", append(args, dep.ImportPath)...)
}

// restore checks out the given revision.
func restore(dep Dependency) error {
	ps, err := LoadPackages(dep.ImportPath)
	if err != nil {
		return err
	}
	pkg := ps[0]

	dep.vcs, err = VCSForImportPath(dep.ImportPath)
	if err != nil {
		dep.vcs, _, err = VCSFromDir(pkg.Dir, pkg.Root)
		if err != nil {
			return err
		}
	}

	if !dep.vcs.exists(pkg.Dir, dep.Rev) {
		dep.vcs.vcs.Download(pkg.Dir)
	}
	return dep.vcs.RevSync(pkg.Dir, dep.Rev)
}
