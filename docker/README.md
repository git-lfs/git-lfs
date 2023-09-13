# Git LFS Linux Docker Builds

In order to build Linux distribution packages of Git LFS in both the
Debian and RPM package formats and across multiple architectures, the
GitHub Actions workflows for this repository run package build scripts
in a set of Docker containers.  The containers are based on Docker images
created from the Dockerfiles in the `git-lfs/build-dockers`
[repository](https://github.com/git-lfs/build-dockers).

Each Docker image contains either the `debian_script.bsh` script or
`centos_script.bsh` script from the `git-lfs/build-dockers` repository,
as appropriate for the expected output package format.  The relevant script
is executed by default when the image is loaded in a container and builds
Git LFS distribution packages for one or more architectures.  The packages
are written into the `/repo` directory, which is typically a mounted volume
in the Docker container.

## Building Linux Packages

The `docker/run_dockers.bsh` script in this repository provides a
convenient way to run some or all of the Git LFS Docker builds, assuming
the Docker images already exist.  (The `build_dockers.bsh` script in
the `git-lfs/build-dockers` repository may be used to create the Docker
images.)

When run without arguments, the `run_dockers.bsh` script will run
builds for all of the OS versions listed by the `script/distro-tool` utility
when it is passed the `--image-names` option.  This list should match the
available Dockerfiles in the `git-lfs/build-dockers` repository and
therefore also the corresponding Docker images created by that repository's
`build_dockers.bsh` script.

The `run_dockers.bsh` script may also be used to run builds for
only a subset of the available OS versions, e.g.:
```
$ docker/run_dockers.bsh debian_12
$ docker/run_dockers.bsh debian_12 rocky_9
```

The resultant packages of Git LFS will be generated in directories named
`./repos/{OS NAME}/{OS VERSION}`.  Debian packages are written into
those directories, while RPM packages are stored in further levels of
subdirectories, e.g., `./repos/rocky/9/RPMS/x86_64`.

The Docker containers created from each image are removed after use,
unless the `AUTO_REMOVE` environment variable is set to a value other
than `1`.

The Docker images will be removed as well if the `--prune` command-line
argument is supplied.

By default, packages are built for the `amd64` architecture.  Cross-platform
builds may be performed using the `--arch={ARCH}` command-line option if the
requested architecture is supported by the build script in the Docker image
for the given OS and version.  At present, only the `debian_script.bsh`
build script in the Debian images supports cross-platform builds.

If the current user has write permission on the `/var/run/docker.sock`
file descriptor or belongs to the `docker` group, the `run_dockers.bsh`
script will run the `docker` command directly; otherwise it will attempt
to run the command using `sudo`.

### Environment Variables

There are several environment variables which adjust the behavior
of the `run_dockers.bsh` script:

- `AUTO_REMOVE` - Default `1`.  Docker containers are automatically deleted
upon exit unless this variable is set to a value other than `1`.  Retaining
a container may be useful for debugging purposes.  Note that the container
must be removed manually in this case.

- `DOCKER_OTHER_OPTIONS` - Any additional arguments to be passed to the
Docker `run` command.

### Docker Image Development

When developing or debugging the build scripts included in the Docker
images, it may be valuable to start an interactive shell in a container
instead of letting the build script run by default.  To start the Bash shell,
for example, use:

```
$ docker/run_dockers.bsh {OS NAME}_{OS VERSION} -- bash
```

Any command available in the Docker image may be executed instead of Bash
by using the `--` separator, followed by the command to be run.

## Adding Docker Images

To add another Docker image, a new Dockerfile needs to be committed to the
`git-lfs/build-dockers` repository, and then a corresponding new entry should
be added to the `script/lib/distro.rb` file in this repository with an
`image` key whose value matches the OS name and version used in the
name of the new Dockerfile.
