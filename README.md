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


## Checking what we have so far

So, after `./hack/dev` we were left with a binary built out of our local source
code, and two containers running dagger in our docker engine, but then I got
rid of the old one, leaving just the one named `dagger.engine`:

```
        $ docker ps -a
        ID   IMAGE                                COMMAND                  NAMES
>       cb3  localhost/dagger-engine.dev:latest   "dagger-entrypoint.s…"   dagger-engine.dev
```

But, what's going on in that container? first we can check the definition of
the container itself to get some idea of how it's setup (maybe it gives us some
hints about whether it's using the host daemon or if this is some
"docker-in-docker"-a-like setup - more like "buildkit+runc in docker" i'd
guess)

```
        ---
        - Id: cb35846c5782ea62668c3a2bab42008880eb306c0b57465867906e37453041b5
          Path: dagger-entrypoint.sh
          Args: [ "--debug" ]
          HostConfig:
> 1         Binds: [ "dagger-engine.dev:/var/lib/dagger" ]
            NetworkMode: default
            PortBindings: {}
> 2         Privileged: true
> 3         PublishAllPorts: false
          Mounts:
          - Type: volume
            Name: dagger-engine.dev
            Source: "/var/lib/docker/volumes/dagger-engine.dev/_data"
            Destination: "/var/lib/dagger"
          Config:
            Env:
> 4         - _EXPERIMENTAL_DAGGER_CLOUD_TOKEN
>           - _EXPERIMENTAL_DAGGER_CLOUD_URL
>           - _EXPERIMENTAL_DAGGER_GPU_SUPPORT
>           - _EXPERIMENTAL_DAGGER_CACHE_CONFIG
>           Image: localhost/dagger-engine.dev:latest
          NetworkSettings:
> 5         Networks: {bridge: {IPAddress: 172.17.0.3}}
```

from the config, we can infer a couple things:

1. we're having a named volume mounted at `/var/lib/dagger`, it'd be
   interesting to see later on what's put in there

2. running the container in privileged mode (my guess is due to the fact that
   we're also running buildkit inside?)

3. we're not automatically publishing ports exposed via EXPOSE to the host

4. some environment variables seem to be toggling feature flags

5. plain standard bridge networking being utilized

we can also notice that the container has `dnsmasq` running alongside the
engine:

```
        $ docker exec -it dagger-engine.dev ps aux

        PID   USER     COMMAND
          1   root     /usr/local/bin/dagger-engine \
                         --config /etc/dagger/engine.toml --debug
>        33   root     /usr/sbin/dnsmasq --keep-in-foreground \
>                        --log-facility=- --log-debug -u root \
>                        --conf-file=/var/run/containers/cni/dnsname/dagger/dnsmasq.con
```

and that we *don't* see `buildkitd` in there (could it be that it's bringing it
up in a "daemonless" fashion? we'll see).

I think for now I'm good with stopping here in terms of going more in-depth
into the building process and instead focusing on the tutorials/guides to learn
more about how this all works in the background. let's go!


## quick start

from a high-level, it seems like dagger works as such:

```

definition of my pipeline using
the Go dagger sdk                       materialization of the
    '                                       executions
    '                                           '
    '                    (container)            '
    '          .-----dagger-engine.dev?---------+--------.
    '          |                                '        |
  .....        |       .............       ...........   |
  CI.GO ---GRAPHQL---> DAGGER-ENGINE ----> OCI RUNTIME   |
  .....        |       .............       ...........   |
               |            '                            |
               '------------+----------------------------'
                            '
                      computer llb,
                  tells buildkit to solve?

```

let's move on an see how much more detail we can add to this (pretty sure
there's much more to it) - onto setting up a small sample!


## sample

setting up a sample, the first surprise is that my hypothesis that we'd be able
to actually just go ahead and use our new container seems to be .. wrong - we
end up getting a fresh new one:

```
        $ go run main.go

>       Creating new Engine session... OK!
>       Establishing connection to Engine... 1: connect
        1: > in init
        1: starting engine


"""
wut? i thought we'd use the 'dagger-engine.dev' 
container, but, i guess not?
"""
```

perhaps to "fix" this we can instead have those environment variables that we
had seen before in the `hack/dev` script to force the sdk (through our client)
to connect to the right one and not try to initialize a new engine.

so, here I proceed to remove that old container

```
docker rm -f dagger-engine-7b45c2238c1141a1
```

then run with the environment variable set


```
_EXPERIMENTAL_DAGGER_RUNNER_HOST=docker-container://dagger-engine.dev go run ./main.go
```

and indeed! no more re-initialization, and we can rely on that single engine
that we had built - sweet!

something that's important to note is that in the `hack/dev` script we also saw
another environment variable set: `_EXPERIMENTAL_DAGGER_CLI_BIN`.

to figure out why that'd be needed in the first place, a good approach is to
trace all the `execve`s that are happening across the system so that we could
perhaps discover is we have a separate `dagger` cli being used that is not the
one we just built (thus, making the case for having the environment variable
pointing at our build).

using `bcc`'s execsnoopvis, we have our answer - `dagger` is somehow spinning a
different binary for ... "some reason".


```
"""
huh, so, there's some magic here! `dagger session` seems to be a hidden
command that the cli has to ... perhaps act as a bridge to something?
we'll see later
"""


        $ sudo ./execsnoop.py
        PCOMM            PID     PPID    RET ARGS
        go               703968  686918    0 /usr/local/go/bin/go run main.go
        ...
        main             704119  703968    0 /tmp/go-build2805719434/b001/exe/main
>       dagger-0.9.3     704126  704119    0 /home/cirocosta/.cache/dagger/dagger-0.9.3 \
>                                               session \
>                                                 --label dagger.io/sdk.name:go \
>                                                 --label dagger.io/sdk.version:0.9.3
        ...
```

so, yes, that environment variable apparently is very important - having it
set, we can see that indeed, we have dagger leveraging the binary we've just
built:

```
"""
yay, using our binary!
"""

        go               705596  705595    0 /usr/local/go/bin/go run ./main.go
        main             705762  705596    0 /tmp/go-build1313601040/b001/exe/main
>       dagger           705769  705762    0 /home/cirocosta/dev/cirocosta/dagger/bin/dagger \
>                                               session \
>                                                   --label dagger.io/sdk.name:go \
>                                                   --label dagger.io/sdk.version:0.9.3
```

so ok, it seems like our high-level view of the system was really missing
something - this `dagger session` seems to be something that is actually in
between the engine and our client. so, let's figure out how that works.


## the dagger session thing

When the Go SDK tries to establish the connection, it brings brings up `dagger
session` process inheriting the environment from the execution of our sample
(so that, e.g., `..._RUNNER_HOST` env var gets passed through to `dagger
session`). 

With the session process being up, it then gets back from `dagger session`
the port in which `dagger session` has a tcp socket waiting for connections to
proxy through itself to the `dagger engine` running in the container.


```

  go run main.go

     ----> brings up `dagger session`
                         |
                         |
                  listens on ephemeral port
                         |
     <-------------------'
            tells `go run` what
          that port is and a session token


    ...


    `main.go` (Go client sdk)  makes HTTP requests to `dagger session`'s HTTP
    server that is proxying those to the dagger engine indicated by 
    _EXPERIMENTAL_DAGGER_RUNNER_HOST.
    

               (uuid 'token' as basic auth)
               (looback+port from dagsess)
                          '
                          '
  go run main.go    ----HTTP---> dagger session --HTTP?--> dagger engine

                   (graphql)                  (graphql?)


```

ps.: interestingly, because our `DAGGER_RUNNER_HOST` env var indicates a
container (`docker-container://dagger-engine.dev`), when `dagger session` is
creating a buildkit client it is able to connect to a buildkit inside the
docker container using `docker exec -i ... buildctl dial-stdio`, which is
essentially providing a stdio-based proxy to connecting to the daemon.

so ... soo .... sooo .... `dagger-engine` is actually a "fork" of `buildkit`,
and `dagger session` is where all the magic of graphql etc takes place? i'm
probably too confused now hahah


