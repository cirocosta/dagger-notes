# engine

welp, it turns out it's not that easy to just `go run ./cmd/engine` as there's
a bunch of extra setup that's expected to have occurred before in order to have
a functioning worker.

from `internal/mage/util/engine.go`'s `devEngineContainer`, we can see that we need:

- apk packages
  - for buildkit
    - git
    - openssh
    - pigz
    - xz
  - for cni
    - iptables
    - ip6tables
    - dnsmasq

- binaries
  - runc
    - `/usr/local/bin/runc`
    - built from `opencontainers/runc`, currently v1.1.9
  - buildctl
    - `/usr/local/bin/buildctl`
    - built from `moby/buildkit` at a specific commit due to dependencies'
      incompatibilities
  - dagger-shim
    - `/usr/local/bin/dagger-shim`
    - built from our src, `./cmd/shim`
  - engine
    - `/usr/local/bin/dagger-engine`
    - built from our src, `./cmd/engine`
  - dagger    
    - `/usr/local/bin/dagger`
    - built from our src, `./cmd/dagger`
  - qemu
    - rootfs from `tonistiigi/binfmt@sha256:e06789462ac7e2e096b53bfd9e607412426850227afeb1d0f5dfa48a731e0ba5` available under `/usr/local/bin`
		- ```
      Filetree
      ├── buildkit-qemu-aarch64
      ├── buildkit-qemu-arm
      ├── buildkit-qemu-i386
      ├── buildkit-qemu-mips64
      ├── buildkit-qemu-mips64el
      ├── buildkit-qemu-ppc64le
      ├── buildkit-qemu-riscv64
      └── buildkit-qemu-s390x
      ```
  - cni plugins
    - built from scratch (of github.com/containernetworking/plugins) to bump go
      runtime as well as fix dependencies that might have cves
    - these plugins go under `/opt/cni/bin/`
      - builds our own `./cmd/dnsname` binary and puts under
        `/opt/cni/bin/dnsname`
      - builds `bridge`, `loopback`, `firewall`, and `host-local` plugins (all
        going under `/opt/cni/bin/<plugin_name>`

- sdks
  - go sdk engine tarball
    - `/usr/local/share/dagger/go-module-sdk-image.tar`
    - essentially, alpine container with the `codegen` binary (`./cmd/codegen`)
  - python sdk engine directory
    - `/usr/local/share/dagger/python-sdk/runtime`
    - essentially, the `sdk/python` directory including `runtime/` and a few
      other directories

- files & directories
  - dagger state dir: `/var/lib/dagger`
  - dagger engine config file: `/etc/dagger/engine.toml`
  - dagger entrypoint: `/usr/local/bin/dagger-entrypoint.sh`
    - pretty much just ensuring we can run under `cgroup v2`

as of right now, that's all produced making use of dagger itself (not using
modules, *yet*).

with the image built, it can be ran as


```bash
docker run run -d \
  -v dagger-engine.dev:/var/lib/dagger \       # state
  --name dagger-engine.dev \
  --privileged \
  localhost/dagger-engine.dev:latest \         # << tag'ed from bin/engine.tar
    --debug`
```

note: a good way of getting a fresh build and having a session targetting that
engine specifically is to `./hack/dev bash` to run the mage target that does
the building and running of the engine in a container, then setting env vars
that get propagated to the `bash` session such that it targets exactly that
engine (as well as using the binary we've just built for creating sessions
etc).

note2.: in order to get a debugging session, it's useful to have a `dlv`
container and then run it as part of the same pid namespace as dagger-engine,
e.g.:

```
docker run -it --privileged --rm --pid=container:dagger-engine.dev a attach 1
```
