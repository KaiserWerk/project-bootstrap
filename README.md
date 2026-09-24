# pb — Project Bootstrap CLI (MVP)

This repository contains a minimal scaffold for the `pb` CLI (Project Bootstrap).

Build:

```bash
cd cmd/pb
go build -o ../../pb
```

Run examples:

```bash
# show usage
./pb

# search (stub)
./pb search auth

# create project (stub)
./pb create-project web-go demo

# add module (stub)
./pb add-module frontend-auth
```

Configuration:

Create a global config at `~/.pb/pb-config.yaml` (see `pb-config.example.yaml` in this repo). Example:

```yaml
sources:
  - github.com/xyz/lib
  - gitlab.com/abc/otherlib
```

Next steps:

- Implement registry clone and search
- Implement `create-project` templating and variable prompts
- Implement `add-module` installation flow and Go adapter

Lockfile (`pb.lock`):

- After installing a module, `pb` now records a `pb.lock` entry for the module including:
  - `commit`: the source repository commit hash (when available)
  - `files`: a map of installed file paths -> SHA256 hashes

This helps detect drift and enables deterministic installs and future verification tooling.
