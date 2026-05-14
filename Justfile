set shell := ["bash", "-eu", "-o", "pipefail", "-c"]

import 'just/go.just'
import 'just/docker.just'

# List all available recipes
default:
    @just --list
