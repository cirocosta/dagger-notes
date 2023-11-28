# sample pipeline

the goal here is to try figuring out ways in which we can build pipelines
leveraging as much as possible the underlying tech (graphs all the way).

in this first iteration, I'm setting up the pipeline "from beginning to
finish", thinking about it in a serial manner: "first run these that can be ran
concurrently, then those that should wait for the first ones, then we finalize":


```

checks   ----> builds          ------------> publishing
  |            |                              |
  |            |                              |
  +- vet       +- image -> imagescan          '- relocating image -> push tag
  +- src scan  '- binary
  '- test      
```

currently implemented:

- checks
  - vet
  - test

- builds
  - image
    - image scan

see
[runPipelines](https://github.com/cirocosta/dagger-notes/blob/0e0a26f2e852c17bc987385d15ce8417cec1f4fe/sample/ci/pipelines.go#L10-L22).


## output

currently, looks like this (pretty neat):

```
┣─╮
│ ▽ init
│ █ [0.11s] connect
│ ┣ [0.10s] starting engine
│ ┣ [0.01s] starting session
│ ┃ OK!
│ ┻
█ [21.95s] go run ./ci
┣─╮
│ ▼ sample
│ ┣─╮
│ │ ▼ checks
│ │ ┣─╮
│ │ │ ▽ host.directory .
│ │ │ █ [0.04s] upload . from xps (client id: ilewf4ser1bdt9nrs00ob8pjf) (exclude: ./ci)
│ │ │ ┣ [0.00s] transferring .:
│ │ │ █ [0.04s] upload . from xps (client id: ilewf4ser1bdt9nrs00ob8pjf) (exclude: ./ci)
│ │ │ █ [0.00s] blob://sha256:8c88742647411e4ebb8b2bb91c2850a9849c158b4df0ba0e1f4833ff5e6f05cb
│ │ │ ┣─╮ blob://sha256:8c88742647411e4ebb8b2bb91c2850a9849c158b4df0ba0e1f4833ff5e6f05cb
│ │ │ ┻ │
│ │ ┣─╮ │
│ │ │ ▽ │ from golang:1.21
│ │ │ █ │ [0.14s] resolve image config for docker.io/library/golang:1.21
│ │ │ █ │ [0.02s] pull docker.io/library/golang:1.21
│ │ │ ┣ │ [0.02s] resolve docker.io/library/golang:1.21@sha256:81cd210ae58a6529d832af2892db822b30d84f817a671b8e1c15cff0b271a3db
│ │ │ ┻ │
│ │ █◀──┤ [0.02s] copy / /app
│ │ █   │ [7.32s] exec go vet ./...
│ │ █   │ [7.37s] exec go test -v ./... -run Succeed
│ │ ┃   │ ?       sample  [no test files]
│ │ ┃   │ ?       sample/e2e      [no test files]
│ │ ┃   │ === RUN   TestAdd_Succeed
│ │ ┃   │ --- PASS: TestAdd_Succeed (0.00s)
│ │ ┃   │ PASS
│ │ ┃   │ ok      sample/pkg      0.001s
│ │ ┻   │
│ ┣─╮   │
│ │ ▼   │ builds
│ │ ┣─╮ │
│ │ │ ▽ │ docker build
│ │ │ █ │ [0.06s] [builder 1/4] FROM docker.io/library/golang:alpine@sha256:110b07af87238fbdc5f1df52b00927cf58ce3de358eeeb1854f10a8b5e5e1411
│ │ │ ┣ │ [0.06s] resolve docker.io/library/golang:alpine@sha256:110b07af87238fbdc5f1df52b00927cf58ce3de358eeeb1854f10a8b5e5e1411
│ │ │ █ │ [0.06s] [stage-1 1/3] FROM gcr.io/distroless/static:latest@sha256:6706c73aae2afaa8201d63cc3dda48753c09bcd6c300762251065c0f7e602b25
│ │ │ ┣ │ [0.05s] resolve gcr.io/distroless/static:latest@sha256:6706c73aae2afaa8201d63cc3dda48753c09bcd6c300762251065c0f7e602b25
│ │ │ ┣─┼─╮ [stage-1 1/3] FROM gcr.io/distroless/static:latest@sha256:6706c73aae2afaa8201d63cc3dda48753c09bcd6c300762251065c0f7e602b25
│ │ │ █◀╯ │ [0.05s] copy / /
│ │ │ █   │ CACHED [builder 2/4] WORKDIR /app
│ │ │ █   │ [0.04s] [builder 3/4] COPY . .
│ │ │ ┣─╮ │ [builder 3/4] COPY . .
│ │ │ ┻ │ │
│ │ █◀──╯ │ [5.23s] [builder 4/4] RUN set -x && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -trimpath -tags osusergo,netgo,static_build -o sample .
│ │ ┃     │ + CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -trimpath -tags osusergo,netgo,static_build -o sample .
│ │ █◀────╯ CACHED [stage-1 2/3] COPY --chown=nonroot:nonroot --from=builder /app/sample .
│ │ ┣─╮
│ │ │ ▽ from ghcr.io/aquasecurity/trivy:canary
│ │ │ █ [0.10s] resolve image config for ghcr.io/aquasecurity/trivy:canary
│ │ │ █ [0.03s] pull ghcr.io/aquasecurity/trivy:canary
│ │ │ ┣ [0.03s] resolve ghcr.io/aquasecurity/trivy:canary@sha256:37f6369895ce2c624baea8a95abe9e93f426bd4484947a681cc47cbc918c959d
│ │ │ ┣─╮ pull ghcr.io/aquasecurity/trivy:canary
│ │ │ ┻ │
│ │ █◀──╯ [4.89s] exec trivy image ttl.sh/hello-dagger-8841381@sha256:549e7049c3165ff13b2ba3f86c5354411a1a496d908bf9c872b86a1dad5fce25
│ │ ┃     2023-11-12T22:50:46.469Z        INFO    Need to update DB
│ │ ┃     2023-11-12T22:50:46.469Z        INFO    DB Repository: ghcr.io/aquasecurity/trivy-db
...
│ │ ┃     2023-11-12T22:50:49.697Z        INFO    Number of language-specific files: 0
│ │ ┃
│ │ ┃     ttl.sh/hello-dagger-8841381@sha256:549e7049c3165ff13b2ba3f86c5354411a1a496d908bf9c872b86a1dad5fce25 (debian 11.8)
│ │ ┃     =================================================================================================================
│ │ ┃     Total: 0 (UNKNOWN: 0, LOW: 0, MEDIUM: 0, HIGH: 0, CRITICAL: 0)
┻ ┻ ┻
• Engine: cb35846c5782 (version devel ())
⧗ 22.07s ✔ 46 ∅ 2
```
