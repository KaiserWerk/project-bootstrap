# pb — Project Bootstrap CLI

A CLI tool to finally stop you from needing to find specific pieces of code in your older projects. Set up once and be done with it.
pb is designed to help you quickly bootstrap new projects and download reusable code modules across your projects.
Code is stored in any git repository so there's no need for a centralized storage solution.

I created `pb` to finally stop fidgeting with gists, locally saved snippets and code copied from other repositories.
With `pb` I have central sources for all my reusable code, which I call `modules` in this project.
`templates` are project blueprints that can be used to quickly set up new projects.

Initially, this went more in the direction of a package manager, but since modules contain code placed in a project with
the intent to modify them, this thought was discarded. It's more of a code snippet and project bootstrapping tool.

You are not restricted to using modules and templates from a single source repository. 
You can configure multiple source repositories in your global configuration file.

Also, you're not language-dependent with this. I love go, but you can use it with any programming language. 
Maybe you can come up with even better use cases for your own projects. Let me know!

## Installation

Get the binary for your OS/Architecture from the [Releases](https://github.com/KaiserWerk/project-bootstrap/releases) page.
I wholeheartedly recommend placing the binary in a directory included in your system's PATH.

## Usage:

![creeating a project and adding a module](docs/images/pb-create-project.gif)

```bash
# show commands and parameters
pb

# create a global configuration file. Afterwards, set up at least one source repository.
pb create-config

# create project in the current directory
pb create-project <template name> <project name>
# e.g.:
pb create-project web-go myproject

# add module in the current project
pb add-module <module name>
# e.g.:
pb add-module frontend-auth

# search for a module or template
pb search auth # This will list all found modules which contain the word auth in their name.

# Commands for creating resources

# Create a module (a collection of code files to be placed in projects, which may be modified afterwards)
# This will create a folder structure and a module.yaml file.
pb create-module <module name>
# e.g.:
pb create-module my-cool-module

# Create a project template
# This will create a folder structure and a template.yaml file.
pb create-template <template name>
# e.g.:
pb create-template my-cool-template

# (re)build the local sources cache from the configured source repositories
pb cache
```

## Setting up a source repository

1. create a git repository in an empty directory.
2. In the repository, create a folder named `tool-registry` and cd into it.

3. Create a registry file with the command `pb create-registry`.
This creates a `registry.yaml` file in the current directory (should be `<repository>/tool-registry/registry.yaml`).

4. Create the folders `modules` and `templates` inside the `tool-registry` directory so they reside alongside the `registry.yaml` file.

5. Place your created versioned modules and templates inside the `modules` and `templates` folders, respectively.

6. Commit and push your changes to the remote repository. This repository can now be used as a source in your global configuration file.