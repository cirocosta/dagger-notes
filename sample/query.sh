#!/usr/bin/env bash

dagger query --progress=plain <<< '{ container { from(address:"hello-world") { stdout } } }'
