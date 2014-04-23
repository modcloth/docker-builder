package bob

import (
	"github.com/rafecolton/bob/dclient"
	"github.com/rafecolton/bob/log"
	"github.com/rafecolton/bob/parser"
)

import (
	"github.com/deckarep/golang-set"
	"github.com/onsi/gocleanup"
)

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

/*
A Builder is the struct that actually does the work of moving files around and
executing the commands that do the docker build.
*/
type Builder struct {
	dockerClient dclient.DockerClient
	log.Log
	workdir         string
	isRegular       bool
	nextSubSequence *parser.SubSequence
}

/*
SetNextSubSequence sets the next subsequence within bob to be processed. This
function is exported because it is used explicitly in tests, but in Build(), it
is intended to be used as a helper function.
*/
func (bob *Builder) SetNextSubSequence(subSeq *parser.SubSequence) {
	bob.nextSubSequence = subSeq
}

/*
NewBuilder returns an instance of a Builder struct.  The function exists in
case we want to initialize our Builders with something.
*/
func NewBuilder(logger log.Log, shouldBeRegular bool) *Builder {
	if !shouldBeRegular {
		return &Builder{
			isRegular: false,
		}
	}

	client, err := dclient.NewDockerClient(logger, shouldBeRegular)

	if err != nil {
		return nil
	}

	return &Builder{
		dockerClient: client,
		Log:          logger,
		isRegular:    true,
	}
}

/*
Build is currently a placeholder function but will eventually be used to do the
actual work of building.
*/
func (bob *Builder) Build(commands *parser.CommandSequence) error {
	/*
		  Steps:
		  1. loop through command sequence
		  2. for each subsequence
		  	a. CleanWorkdir()
			b. set next subsequence()
			c. Setup()
			d. run build command (make sure it's in the right dir)
			e. get uuid from build command, use to modify tag commands
			f. run tag commands
			g. run push commands
	*/
	return nil
}

/*
Setup moves all of the correct files into place in the temporary directory in
order to perform the docker build.
*/
func (bob *Builder) Setup() error {
	if bob.nextSubSequence == nil {
		return errors.New("no command sub sequence set, cannot perform setup")
	}

	meta := bob.nextSubSequence.Metadata
	fileSet := mapset.NewSet()

	if len(meta.Included) == 0 {
		files, err := ioutil.ReadDir(bob.Repodir())
		if err != nil {
			return err
		}

		for _, v := range files {
			fileSet.Add(v.Name())
		}
	} else {
		for _, v := range meta.Included {
			fileSet.Add(v)
		}
	}

	// subtract any excludes from fileSet
	for _, exclude := range meta.Excluded {
		if fileSet.Contains(exclude) {
			fileSet.Remove(exclude)
		}
	}

	if fileSet.Contains("Dockerfile") {
		fileSet.Remove("Dockerfile")
	}

	// add the Dockerfile
	fileSet.Add(meta.Dockerfile)

	workdir := bob.Workdir()
	repodir := bob.Repodir()

	// copy the actual files over
	for file := range fileSet.Iter() {
		src := fmt.Sprintf("%s/%s", repodir, file)
		dest := fmt.Sprintf("%s/%s", workdir, file)

		if file == meta.Dockerfile {
			dest = fmt.Sprintf("%s/%s", workdir, "Dockerfile")
		}

		fileInfo, err := os.Stat(src)
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			err = CopyDir(src, dest)
		} else {
			err = CopyFile(src, dest)
		}
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	return nil
}

/*
Repodir is the dir from which we are using files for our docker builds.
*/
func (bob *Builder) Repodir() string {
	if !bob.isRegular {
		repoDir := "spec/fixtures/repodir"
		return fmt.Sprintf("%s/%s", os.ExpandEnv("${PWD}"), repoDir)
	}
	return fmt.Sprintf("%s", os.ExpandEnv("${PWD}"))
}

/*
Workdir returns bob's working directory, creating one first if bob.workdir is
currently set to empty string.
*/
func (bob *Builder) Workdir() string {
	if !bob.isRegular {
		specWorkdir := "spec/fixtures/workdir"
		return fmt.Sprintf("%s/%s", os.ExpandEnv("${PWD}"), specWorkdir)
	}

	tmp, err := ioutil.TempDir("", "bob")
	if err != nil {
		return ""
	}
	return tmp
}

/*
CleanWorkdir effectively does a rm -rf and mkdir -p on bob's workdir.  Intended
to be used before using the workdir (i.e. before new command groups).
*/
func (bob *Builder) CleanWorkdir() error {
	workdir := bob.Workdir()

	if !bob.isRegular {
		readme := fmt.Sprintf("%s/README.txt", workdir)
		os.RemoveAll(workdir)
		err := os.MkdirAll(workdir, 0755)
		if err != nil {
			return err
		}

		file, err := os.Create(readme)
		if err != nil {
			return err
		}
		defer file.Close()

		bytes := []byte("This directory tree is used for specs - please do not modify.\n")
		if _, err := file.Write(bytes); err != nil {
			return err
		}

		return nil
	}

	os.RemoveAll(workdir)
	err := os.MkdirAll(workdir, 0755)
	if err != nil {
		return err
	}

	gocleanup.Register(func() {
		os.RemoveAll(workdir)
	})

	return nil
}

/*
LatestImageTaggedWithUUID accepts a uuid and invokes the underlying utility
DockerClient to determine the id of the most recently created image tagged with
the provided uuid.
*/
func (bob *Builder) LatestImageTaggedWithUUID(uuid string) string {
	// eat the error and let it fail when we try to run the docker command
	id, err := bob.dockerClient.LatestImageTaggedWithUUID(uuid)
	bob.Println(err)
	return id
}