# sample


```

test      --.
vet         +--> image  -->  image scan   ----+--> publish
binary    --'      |                          |
                   '------>  e2e test --------'

```

i.e. we have a dependency graph as such:

```
publish:
  - e2e_test
  - image_scan
  - binary

image_scan:
  - image

e2e_test:
  - image

binary:
  - test
  - vet

# image builds a container image out of source code, making it available as a
# file (tarball)
#
# image depends on both `test` and `vet` finishing (e.g., say, `image` is an
# expensive to run stage), so, even though it doesn't consume an output from
# those, we'd like to have a synchronization barrier here somehow.
#
image:
  - test
  - vet
```
