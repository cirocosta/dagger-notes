# Notes from trying Dagger out

Here's a collection of notes from a weekend of playing with
[Dagger](https://dagger.io/).

> [!WARNING]  
> This is literally my first time trying dagger out. Please don't rely on what
> you're seeing here for anything serious as I might be incredibly wrong.
>
> No, this is not a tutorial, just raw notes.


## Setting up the environment

Before anything else, fwiw I'll be trying everything out on a bare metal
machine as such:

```console
$ cat /etc/lsb-release

DISTRIB_ID=Ubuntu
DISTRIB_RELEASE=22.04
DISTRIB_CODENAME=jammy
DISTRIB_DESCRIPTION="Ubuntu 22.04.3 LTS"

$ uname -a
Linux xps 6.2.0-35-generic #35~22.04.1-Ubuntu  ...
```

With `docker` already installed/up and running:

```console
$ docker --version
Docker version 24.0.7, build afdd53b
```


## Building dagger

### Cloning then looking around

I know it's usually more useful to just go grab the binaries and get started
that way, but, I really want to be able to tinker with this, so, let's try a
dev build and see how far we get.

First, clone the repo then get there, then let's see how we can get this built.

```console
$ git clone https://github.com/dagger/dagger
$ cd dagger
```

Looking around (I know, I know, RTFM, ciro!), `./install.sh` looked really
promising but it seems like it's more for having a clean install *without*
getting the binaries built locally - the script looks pretty solid with a nice
simple `execute()` function that gets invoked as the entrypoint

```

"""
hmm nice, downloads, runs some checksums, checks if
we can put the binary where we want, installs completions,
neat neat
"""


      execute() {
      ...
          log_debug "downloading files into ${tmpdir}"
>         http_download "${tmpdir}/${tarball}" "${tarball_url}"
>         http_download "${tmpdir}/${checksum}" "${checksum_url}"
          hash_sha256_verify "${tmpdir}/${tarball}" "${tmpdir}/${checksum}"
          srcdir="${tmpdir}"
          (cd "${tmpdir}" && untar "${tarball}")
          test ! -d "${bin_dir}" && install -d "${bin_dir}"
>         install "${srcdir}/${binexe}" "${bin_dir}"
          log_debug "display shell completion instructions"
>         install_shell_completion
          log_info "installed ${bin_dir}/${binexe}"
          rm -rf "${tmpdir}"
      }

      execute
```

easy to read - gj! (but, not what I'm looking for). `./hack` looks like is what
I'm looking for


```
"""
good ol' hack dir with scripts, sounds like where we want to go
to get things built from scratch
"""


    $ tree -L 1 -C ./hack

    ./hack
>   ├── dev
    ├── make
    ├── README.md
    └── with-dev
```

but let's read the manual.

Looking at the README.md, it indeed points out (at the end) that I should read
the manual for what I'm trying to do (`./CONTRIBUTING.md`).


### Going through CONTRIBUTING.md

"how to run a development engine" - there we go! indeed, `./hack/dev` gives us
what we want.

first thing that catches my attention is that ... well, they're not using
Makefiles, but instead, `mage`, which yeah, makes sense given the approach of,
e.g., writing a pipeline using Go - why not the build scripts too?

```
"""
interesting, we'll use mage to run some build scripting
here, neat - we should take a look at what's going on
there
"""


      #!/usr/bin/env bash

      set -e -u -x

      DAGGER_SRC_ROOT="$(cd $(dirname "${BASH_SOURCE[0]}")/.. && pwd)"
      MAGEDIR="$DAGGER_SRC_ROOT/internal/mage"

      pushd $MAGEDIR
>     go run main.go -w $DAGGER_SRC_ROOT engine:dev
      popd


"""
rest seems unrelated to what I'm trying to achieve
as I'll be running just `./hack/dev` with no further positional
arguments so those exports and eval don't really matter
"""

```

looking at the mage file for `engine:dev`, it looks a little magical that this
will "just work"? 


```
""" 
wait a sec, so we'll be sort of "self-compiling" here? if so, I imagine we're
relying on some form of pre-built stable release of dagger itself to make this
work?
"""


      func (t Engine) Dev(ctx context.Context) error {
>             c, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
              ...

>             c = c.Pipeline("engine").Pipeline("dev")
              ...
>             _, err = c.Container().Export(ctx, tarPath, dagger.ContainerExportOpts{
>                     PlatformVariants: platformVariants,
>                     ForcedCompression: dagger.Gzip,
>             })
              ...
```

while running `./hack/dev` we can see that something is going on:

1. somehow and engine showed up in my list of docker containers

```
$ docker ps -a
CONTAINER ID   IMAGE                              COMMAND                  CREATED          STATUS          PORTS     NAMES
3e11b4997a9d   registry.dagger.io/engine:v0.9.3   "dagger-entrypoint.s…"   12 seconds ago   Up 10 seconds             dagger-engine-7b45c2238c1141a1
```

this makes me think that we're using a stable build (0.9.3) to use as the
bootstrapping of dev builds so that we can do dogfooding from very early on in
the dev cycle of the team (great!)


2. but, the build seems to have had a problem - opportunity to learn more

i'm not sure exactly why, but got a problem with the resolution of the alpine
image:

```
      14: blob://sha256:b62097f540319097cb1349c4a8578bcfaf7d0b6641094d86c72a51494fb044e4 DONE
      14: > in engine > dev > host.directory .
      14: blob://sha256:b62097f540319097cb1349c4a8578bcfaf7d0b6641094d86c72a51494fb044e4 DONE

      21: resolve image config for docker.io/library/golang:1.21.3-alpine3.18
      21: > in engine > dev > from golang:1.21.3-alpine3.18
      21: resolve image config for docker.io/library/golang:1.21.3-alpine3.18 ERROR: 
          failed to do request: 
>               Head "https://registry-1.docker.io/v2/library/golang/manifests/1.21.3-alpine3.18": EOF
      Error: input:1: pipeline.pipeline.container.from failed to do request: 
>               Head "https://registry-1.docker.io/v2/library/golang/manifests/1.21.3-alpine3.18": EOF

      exit status 1
```


sounds more like a connectivity issue (given the EOF on a simple HEAD request),
and yeah, we can confirm the tag is sane:

```

"""
yeah, as expected yes, there is a corresponding entry in the
index for the tag & platform we're aiming at
"""


        $ docker buildx imagetools inspect golang:1.21.3-alpine3.18

        Name:      docker.io/library/golang:1.21.3-alpine3.18
        MediaType: application/vnd.docker.distribution.manifest.list.v2+json
        Digest:    sha256:96a8a701943e7eabd81ebd0963540ad660e29c3b2dc7fb9d7e06af34409e9ba6

>       Manifests:
>         Name:      docker.io/library/golang:1.21.3-alpine3.18@sha256:4f95f6bd37a96bb17ff610ed3bb424fc7d2926e08da4ed2276ed5f279d377852
>         MediaType: application/vnd.docker.distribution.manifest.v2+json
>         Platform:  linux/amd64

        ...
```

for sanity checking, let me try pulling *anything* else that's not from dockerhub (maybe I'm just having bad luck):

```
"""
ok, gcr is not dockerhub, if it's a connectivity issue from daemon to anywhere,
then this should fail
"""

        $ docker pull gcr.io/distroless/static-debian12
        ...
        672354a91bfa: Pull complete
        Digest: sha256:0c3d36f317d6335831765546ece49b60ad35933250dc14f43f0fd1402450532e
>       Status: Downloaded newer image for gcr.io/distroless/static-debian12:latest
>       gcr.io/distroless/static-debian12:latest

        $ docker pull docker.io/library/golang:1.21.3-alpine3.18
>       Error response from daemon: Head
        "https://registry-1.docker.io/v2/library/golang/manifests/1.21.3-alpine3.18":
        Get
        "https://auth.docker.io/token?account=cirocosta&scope=repository%3Alibrary%2Fgolang%3Apull&service=registry.docker.io":
        EOF

```

uhhhh, so, I proceed to go more "defaults only" with my docker daemon config
(modifying `/etc/docker/daemon.json` to not have a bumped up
`max-concurrent-downloads`), then restart the engine (`systemctl restart
docker.service`), then yes, I get to move further but later find another EOF:

```
"""
welp, dockerhub seems to indeed be having a hard time, but, guess what?
thankfully I can keep running `./hack/dev` and because of how we can rely on
idempotency and cache-reuse that dagger is providing us here, it's frustrating
to deal with the flakiness of the registry not serving me well but hey, the
caching makes this sufferable!
"""


        151: [18.2s] retrying in 4s
        151: sha256:1b0dfc2f3a464bee155c6e863e879c5b6024610a49654d225ea46127e48ed7a7 0B / 15.8KiB
>       151: [27.3s] error: failed to copy: httpReadSeeker: failed open: 
>           failed to do request: 
>               Get "https://registry-1.docker.io/v2/tonistiigi/xx/blobs/sha256:1b0df...": EOF
>       151: pull docker.io/tonistiigi/xx:1.2.1 DONE

>       149: exec xx-apk update ERROR: failed to copy: httpReadSeeker: failed open: failed to do request: 
>               Get "https://registry-1.docker.io/v2/tonistiigi/xx/blobs/sha256:1b0dfc2f...7": EOF
>       149: > in engine > dev
>       149: exec xx-apk update ERROR: failed to copy: httpReadSeeker: failed open: failed to do request: 
>               Get "https://registry-1.docker.io/v2/tonistiigi/xx/blobs/sha256:1b0dfc2...": EOF
>       Error: input:1: pipeline.pipeline.container.from.withEnvVariable.
>           withEnvVariable.withEnvVariable.withEnvVariable.withExec.withDirectory.withExec.
>           withExec.withMountedCache.withMountedCache.withMountedDirectory.withWorkdir.
>           withExec.withExec.withExec.withExec.file 
>               failed to compute cache key: failed to copy: httpReadSeeker: failed open: 
>                 failed to do request: 
>                     Get "https://registry-1.docker.io/v2/tonistiigi/xx/blobs/sha256:1b0dfc2f3a...": EOF

        exit status 1

```

a couple runs later, we're good to go! despite the annoyance of the registry
acting oddly, it turns out that this was a great way of showcasing how being
able to rely hard on the caching makes the experience "ok" as you can just
re-run and not have time being wasted on the intermediary steps.

## It's up!

ok, with the build done, I can see that now I have actually two instances of the engine up:

```
        $ docker ps -a
        ID   IMAGE                                STATUS     NAMES
>       cb3  localhost/dagger-engine.dev:latest   Up 2  min  dagger-engine.dev
>       103  registry.dagger.io/engine:v0.9.3     Up 11 min  dagger-engine-7b45c2238c1141a1
```

given the last few lines of the output on the successful run of `./hack/dev`

```
"""
`docker-container://dagger-engine.dev` seems like a way of somehow
telling the CLI that it should try connecting to a process inside the
container named `dagger-engine.dev`, thus, I'm a bit more certain that we
can get rid of that old one (`7b45c...`)
"""


        225: export file /bin/dagger to host bin/dagger
        225: > in engine > dev
>       export _EXPERIMENTAL_DAGGER_CLI_BIN=bin/dagger
>       export _EXPERIMENTAL_DAGGER_RUNNER_HOST=docker-container://dagger-engine.dev
>       225: export file /bin/dagger to host bin/dagger DONE
```

my guess here is that the older one was the one utilized for the bootstrapping
of it, and that the new one is what's running our freshly built engine with
local source, so I'll proceed with getting rid of the old one
(:crossed_fingers:), and that we can point `PATH` to `./bin` to leverage the
freshly built CLI from there.


```
        $ dagger version
>       dagger devel () (registry.dagger.io/engine) linux/amd64
```

