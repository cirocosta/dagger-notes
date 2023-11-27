# dagger modules

or, "how does `dagger call` magic works?"

ps.: i'm not part of the dagger team and am just exploring - info here might
be wildly innacurate/wrong.


## making the case for dagger modules, from a high level

before we jump into how it works internally, let's find out why you'd want to
use dagger in the first place. grab a coffee.


### `my-app` and the "root Makefile"

Consider that we have a repository `my-app`, and through the software
development lifecycle (sdlc), we want to ensure that we're able to perform:

- linting
- testing
- building
- publishing

well, naturally one could go ahead and write some bash scripts under `./hack`
and then make targets of a `Makefile` such that you:

- `make lint`
- `make test`
- `make build`
- `make publish`

from the perspective of a newcomer into the project, that sounds pretty solid -
quick glance over that makefile and you know you're covered when it comes to
attaining those parts of the sdlc. but, are you really? let's dig into what
those targets would look like (remember, you're a newcomer looking at the
`my-app` repository for the first time).


## `make lint`

`my-app` being a Go application, we chose to run
[`golangci-lint`](https://github.com/golangci/golangci-lint) as the linter, as
well as ... say, [`gitleaks`](https://github.com/gitleaks/gitleaks) to check if
we're accidentally leaking credentials throughout our codebase.

looking at the implementation behind `make lint` we then spot that it's calling
`golangci-lint run` and `gitleaks detect` (i.e., it's encapsulating the detail
of which commands to use and arguments to supply), but, what about the binaries
themselves? if you get them installed in your machine, can you even tell if
you're running the proper version of them? maybe `apt` is giving you a super
outdated version.

```
make lint

  --> exec golangci-lint run
        xxx: golangci-lint not found? "yo, go install this way"

  --> exec gitleaks detect .
        xxx: gitleaks not found? "yo, go install it that way"
```

"Go being Go", there's actually a nice way of dealing with this - instead of
execing `golangci-lint` directly, create a `tools` module, declare
`golangci-lint` and `gitleaks` as dependencies and boom, you can `go run
-mod=tools/go.mod github.com/golangci/golangci-lint/cmd/...` and no worries
about having those binaries ready at PATH, and also no worries with regards to
versioning, life's good.

so, ok, `Makefile`s are good enough for this, and Go got our back for this
target:

```
GCI     := go run -mod tools/go.mod github.com/golang...
GLEAKS  := go run -mod tools/go.mod github.com/gitleaks...

make lint
  $(GCI) run ..
  $(GLEAKS) detect ...
```

all the tooling we need is appropriately versioned, we build on execution if
not already built, all cached, all good!

we can even go a step ahead and split the target in two, make them dependencies
of `make lint` and then let `make` take care of concurrently running them:

```
lint: lint-gcilint lint-gitleaks
  echo "yay"
```

soo ... sounds like we got all we need, right? well, not really - we're not
mentioning that Go had to already been installed before (and yes, the version
of Go matters - we didn't have Go modules a while ago, remember?), and let's
not forget `make` itself.

to summarize, a tiny `make lint` already assumes a bunch:

- you've got `make` installed
- you've got the right version of `go` installed 
- the tools being invoked have been made Go dependencies of a distinct tools
  package (which you point `dependabot` at to ensure it's always up to date)
- you hope those tools are always `go install`able with ease
- you hope the scripts behind `make lint` won't fall into bash/scripting traps
  (ala `sed` in linux vs macos and many others)

uf, ok, let's make it worse.


## make publish

ok, `make lint` was easy once we rely on the Go toolchain to help us out. But,
is that always the case? not really.

for instance, it's not uncomon for a `publish` target to take a container image
and then relocate from an ephemeral registry / tarball to an "official" remote
registry, and a tool like [`skopeo`](https://github.com/containers/skopeo) is
great for that.

despite `skopeo` being written in Go, it's not just a static binary that you
`GOBIN=/bin go install`  and call the day - not only we need to have some extra
dependencies available in the system at build time, its use of CGO ends up
forcing us to have a few libraries available at run time too. 

```
make publish

  skopeo copy oci-archive://image.tar docker://cirocosta/my-app
  |
  '--> uhh, ok, `go install` won't make it, how do I install this?

       (and btw, not only on my linux machine, how do I make sure that
        it's also good to go on my peers' macos instances?)
```

this is an example of where linux containers shine - I've got a non-trivial
tool whose installation and use require more than just a binary and I'd love to
not pollute my environment with those, aka it'd be great to sandbox it all with
a lightweight primitive (wink, containers).

so, we tailor a Dockerfile for coming up with a container image for `skopeo`,
then adopt that strategy of running `skopeo` from within a container based on
an image that has the right libraries and binaries in place:


```

 make publish

  docker run -v $(pwd):/...

          skopeo-container----------------
          |
          |   skopeo copy \
img.tar --+---> oci-archive://./image.tar \        ""relocate src file to dst""
 creds? --+---> docker://cirocosta/my-app
          |
          |

```

with some bash we could isolate this whole thing in a function that takes care
of spinning up a container (good ol `docker run` with some volume mounts and
env vars), but .. wait a second - there's a new dependency on the block:
`docker` itself.

so, ok, newcomer to this project have

- `make` installed
- right version of `go` setup
- docker up and running
- hopes and dreams


### hmm, what if .. we containerised it all

well, if we're going for containers for `skopeo` (which is nice, allows us to
not care about `go` and whatever other dependencies we might need), why not
adopt the same for the rest of the sdlc?

```
lint


  docker run -v $(pwd):/... <gci-lint image>
  
            gci-lint------
            |
    src ----+--> golangci-lint run .        ""lint src dir""
            |
            |


  docker run -v $(pwd):/... <gitleaks image>

            gitleaks------
            |
    src ----+--> gitleaks detect .          ""check src dir""
            |
            |
```

well, for the newcomer, a consumer of this repo, it's easier now as they don't
have to deal with figuring out the versions/installation of anything aside from
`docker` itself.

otoh, the maintainers still have to come up with the Dockerfiles for building
the container images that will provide that tooling - this can either be a
single "golden ci image" that contains all the tools included, or one for each
tool. regardless, we get the benefits of:

- everybody building the necessary tools the same way, no matter where (locally
  / ci)
- makes visible the dependencies associated with the tools for both runtime and
  buildtime (ideally we don't end up with golden ci images as even with
  multistage builds, those could make hard to see exactly what's necessary for
  each tool at runtime)

```
golangci-lint/gitleaks/others...

  Dockerfile

      FROM alpine:git AS source-provider
        RUN git clone golangci-lint ...
      
      FROM golang:alpine AS builder
        COPY --from=source-provider /src ...
        RUN go build ...

      FROM static AS runtime
        COPY --from=builder ...
        ENTRYPOINT [ /bin/golangci-lint ...]

skopeo

  Dockerfile
      FROM alpine:git AS source-provider
        RUN git clone skopeo ...
      
      FROM golang:alpine AS builder
        RUN apk add --update <buildtime_deps>
        RUN go build ...

      FROM alpine AS runtime
        RUN apk add --update <runtime_deps>
        COPY --from=builder ...
        ENTRYPOINT [ /bin/skopo ...]
```

in terms of getting this organized in our `ci` directory, that could look like
something as such

```
/ci
  /tooling
    /golang-ci-lint
      Dockerfile
      golang-ci-lint.sh    -.
    /gitleaks               |
      Dockerfile            |
      gitleaks.sh          -+--   these `.sh` wrapping `docker run`s
    /skopeo                 |
      Dockerfile            |
      skopeo.sh            -'
```

where, going one more step ahead, we could minimize the surface even further by
moving the knowledge of how to run the tools inside those container images by
wrapping the execution of `docker run` with some bash:


```bash
# e.g., `skopeo/skopeo.sh`

main () {
        test $# -eq 0 && show_usage_help
        local command=$1
        shift

        case $command in
                relocate)
                      relocate $1 $2
                      return
                      ;;
...

relocate () {
        local src=$1
        local dst=$2

        docker build ...
        
        docker run --rm 
          -v $(pwd)/image.tar:... \
          <skopeo_img> \
          skopeo copy oci-archive://$src docker://$dst ...
```

such that now `skopeo.sh image.tar cirocosta/my-app` runs a container with
skopeo and its runtime dependencies, doing what we wanted - neat

```
skopeo.sh relocate foo bar
  
  --> builds the base skopeo container image if it's not been built already
  --> runs a container with our image.tar, relocating with `skopeo` as we wanted,
      which works as `skopeo` does indeed find the dynamic libs it needs, and
      has the `image.tar` available due to the mount
```


but we're back to ... a bunch of bash scripts .. in a repo.. having to be very
careful about the tailoring of those `docker run` commands.


### back to dagger

ok so we've arrived at a few conclusions here:

i. we'd like to sandbox the buildinging of that tooling
  - versioning "inputs" to the build
    - compilers, dependencies, etc
      - ideally, these inputs are very explicit in the build process

ii. we'd like to sandbox the execution of that tooling
  - with clearly defined inputs and outputs
    - inputs sometimes coming from the filesystem
    - outputs sometimes landing on the filesystem

iii. we'd like to not use bash to tie these together
  - cross-platform gotchas and lack of compile-time checks
  - tough to deal with concurrency
  - nobody's favorite language, usually full of trickery


dagger is certainly appealing from the perspective of how one can express these
concepts with plain Go (or other languages) - for instance, I can write a
function that promises to the caller that they'll have a container that embeds
`golangci-lint` for whoever executes commands in it.


```go
// BaseContainer provides a container with `golang-lint` built and available at
// PATH.
//
func (m *GolangCILint) BaseContainer() *dagger.Container {
        buildArgs := []dagger.BuildArg{
                dagger.BuildArg{
                        Name:  "GOLANG_IMAGE",
                        Value: "golang:alpine",
                },
                dagger.BuildArg{
                        Name:  "GOLANGCI_LINT_SRC_URL",
                        Value: "https://github.com/golangci/golangci-lint",
                },
                dagger.BuildArg{
                        Name:  "GOLANGCI_LINT_SRC_REV",
                        Value: "v1.55.2",
                },
        }

        return m.buildContext().
                DockerBuild(dagger.DirectoryDockerBuildOpts{
                        BuildArgs: buildArgs,
                })
}

// buildContext provisions a directory that has all the context necessary for
// building the container image that provides `golangci-lint` (i.e., a 
// directory with a Dockerfile)
//
func (m *GolangCILint) buildContext() *dagger.Directory {
        return m.dag.
                Directory().
                WithNewFile("Dockerfile", dockerfileContent)
}
```

such that we could leverage that "promise" of a container and then execute
`golangci-lint` right there.

```go
// "resolve" the base container """promise""" by either building the container
// image or leveraging what's already cached, then create another "promise" of
// a container that will execute the `golangci-lint run` command, then force
// a resolution with `.Stdout()` ultimately grabbing what golangci-lint wrote
// to the std output.
//
        gcilint.BaseContainer().
                WithMountedDirectory(sourceCodeMountPath, src).
                WithWorkdir(sourceCodeMountPath).
                WithExec([]string{
                        "golangci-lint", "run",
                }).
								Stdout(ctx)  //...
```


later on we could also make that available as a function of our `GolangCILint`
type by having a method that returns you the promise of having `golangci-lint
run` on source code you provide


```go
// Lint provides a container for running golangci-lint on the source directory
// provided.
func (m *GolangCILint) Lint(src *dagger.Directory) *dagger.Container {
        return m.BaseContainer().
                WithMountedDirectory(sourceCodeMountPath, src).
                WithWorkdir(sourceCodeMountPath).
                WithExec([]string{
                        "golangci-lint", "run",
                })
}
```

which allows us to go back to that very first intention we had of having a
`make lint` where now we can rely on these pieces of Go code and have dagger
taking care of both the sandboxed builds, but also executions, e.g.:

```
  // """
  // i'm promising you a directory with two files:
  //   - `golangci-lint.json`, containing the results found from the linter
  //   - `gitleaks.json`, containing the results found from gitleaks
  // """
  reportsDir := dag.Directory().
    WithFile("golangci-lint.json", gcilint.LintReport(src)).
    WithFile("gitleaks.json", gitleaks.DetectionReport(src))

  // """
  // now, give me those reports
  // """
  //    --> dagger then works through that chain and concurrently goes from
  //        building the container images (if necessary) to then running the
  //        executables, grabbing the files as they're ready, putting in a
  //        directory, then exporting that directory to our hosts' 
  //        `./out/reports`.
  //
  //            -- powerful
  _, err := reportsDir.Export(ctx, "./out/reports")
```

### hmm ok, what about dagger modules

TBC

