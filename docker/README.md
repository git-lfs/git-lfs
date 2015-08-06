# README #

## TL;DR version ##
1. Run the dockers

        ./docker/run_dockers.bsh
        
2. **Enjoy** all your new package files in

        ./repos/
        
### Slightly longer version ###

1. Generate GPG keys for everything (See GPG Signing)
2. `export REPO_HOSTNAME=repo.something.com`
3. Generate all of the dependency RPMS

        ./docker/run_dockers.bsh ./docker/git-lfs-full-build_*.dockerfile`
        
4. Generate git-lfs/repo packages and sign all packages (including the previous
dependencies)

        ./docker/run_dockers.bsh
        
5. Host the `/repo` on the `REPO_HOSTNAME` server
6. Test the repos and git-lfs in a client environment

        ./docker/test_dockers.bsh

##Using the Dockers##

All docker commands need to either be run as root **or** as a user with docker 
permissions. Adding your user name to the docker group (or setting up boot2docker 
environment) is probably the easiest.

### Running Dockers ###

In order to run the dockers, the docker has to be built, and then run with a
lot of arguments to get the mount points right, etc... A convenient script is 
supplied to make this all easy. Simply run

    ./docker/run_docker.bsh
    
All the `git-lfs_*` images are built automatically, and then run.

To only run certain docker images, supply them as arguments, e.g.

    ./docker/run_docker.bsh ./docker/git-lfs_debian_7.dockerfile
    ./docker/run_docker.bsh ./docker/git-lfs_debian_*.dockerfile
    ./docker/run_docker.bsh ./docker/git-lfs_*[6-8].dockerfile
    
And only those images will be run.

### Building Dockers (Optional) ###

`run_dockers.bsh` calls `build_dockers.bsh`, but you can still call the
script manually to get the building out of the way once while you go make a cup
of tea/coffee.

In order to use the docker **images**, they have to be built so that they are
ready to be used. For OSes like Debian, this is a fairly quick process. 
However CentOS takes considerably longer time, since it has to build go, ruby,
or git from source, depending on the version. Fortunately, you can build the 
docker images JUST once, and you won't have to build it again (until the 
`DOCKER_LFS_BUILD_VERSION` changes.) The build script uses a downloaded release
from github of git-lfs to bootstrap the CentOS image and build/install all the 
necessary software.

This means all the compiling, yum/apt-get/custom dependency compiling is done 
once and saved. (This is done in CentOS by using the already existing 
`./rpm/rpm_build.bsh` script to bootstrap the image and saving the image.)

The script that takes care of ALL of these details for you is

    ./docker/build_dockers.bsh

###Development in Dockers###

Sometimes you don't want to just build git-lfs and destroy the container, you
want to get in there, run a lot of command, debug, develop, etc... To do this, 
the best command to run is bash, and then you have an interactive shell to use

    ./docker/run_docker.bsh {image name(s)}.dockerfile -- bash

After listing the image(s) you want to run, add a double dash (--) and then any 
command (and arguments) you want executed in the docker. Remember, the command
you are executing has to be in the docker image.

## Docker images ##

There are currently three types of docker images:

1. Building images: `git-lfs_{OS NAME}_{OS VERSION}.dockerfile` - These build
git-lfs and save the package/repository in the `/repo` direrctory. This image
also signs all rpms/debs if gpg signing is setup
2. Environment building images: `git-lfs-full-build_{OS_NAME}_{OS_VERSION}.dockerfile` -
These build or install the environment (dependencies) for building git-lfs. These 
are mostly important for CentOS because without these, many dependencies have 
to be built by a developer. These containers should create packages for these 
dependencies and place them in `/repo`
3. Testing images: `git-lfs-test_{OS_NAME}_{OS_VERSION}.dockerfile` - These images
should install the repo and download the git-lfs packages and dependencies to test
that everything is working, including the GPG signatures. Unlike the first two types,
testing images are not guaranteed to work without GPG signatures. They should 
also run the test and integration scripts after installing git-lfs to verify
everything is working in a **non-developer** setup. (With the exception that go
is needed to build the tests...)

This default behavior for both `./docker/run_dockers.bsh` and 
`./docker/build_dockers.bsh` is to run/build all of the _building images_. These
containers will use the currently checked-out version of git-lfs and copied it 
into the docker, and run `git clean -xdf` to remove any non-tracked files, 
(but non-committed changes are kept). git-lfs is built, and a packages/repo is 
created for each container.

These are all a developer would need to test the different OSes. And create the
git-lfs rpm or deb packages in the `/repo` directory. 

In order to distribute git-lfs **and** build dependencies, the dependencies that 
that were built for you by `build_docker.bsh` need to be saved too. Most of these
are downloaded by yum/apt-get and do not need to be saved, but a few are not.
In order to save the necessary dependencies, call `./docker/run_dockers.bsh` on 
`git-lfs-full-build_*.dockerfile` before `git-lfs_*.dockerfile` and the rpms 
will be extracted from the bootstrapped images and saved in the `./repo` directory.
(This _can_ be done in one command. Order matters)

    ./docker/run_dockers.bsh ./docker/git-lfs-full-build_*.dockerfile ./docker/git-lfs_*.dockerfile

This is most important for CentOS 6 where git 1.8.2 or newer is not available, 
only git 1.7.1 is available, so every user either has to build git from source, 
or use the rpms generated by the `git-lfs-full-build_centos_6` image. Calling
the environment building images only needs to be done once, they should remain in
the `./repo` directory afterwards.

### Run Docker Environment Variables ###

There are a few environment variables you can set to easily adjust the behavior
of the `run_docker.bsh` script.

`export` before calling `run_docker.bsh`

`REPO_HOSTNAME` - Override the hostname for all the repos generated/tested (see below)

`DOCKER_AUTOBUILD` - Default 1. `run_docker.bsh` always calls `build_docker.bsh`
to ensure your docker image is up-to-date. Sometimes you may not want this. If 
set to 0, it will not build docker images before running.

`AUTO_REMOVE` - Default 1. Docker containers are automatically deleted on 
exit. If set to 0, the docker containers will not be automatically deleted upon 
exit. This can be useful for a post mortem analysis (using other docker commands
not covered here). Just make sure you clean up the docker containers manually.

### Build Docker Environment Variables ###

`export` before calling `run_docker.bsh`/`build_docker.bsh`. 

`DOCKER_LFS_BUILD_VERSION` - The version of LFS used to bootstrap the (CentOS)
environment. This does not need to be bumped every version. This can be a tag 
or a sha.

##Deploying/Building Repositories##

When `./docker/run_dockers.bsh` is done building git-lfs and generating packages,
it automatically creates a repository for distribution too. Each distro gets a
repo generated in `./repos/{DISTRO_NAME}/{VERSION #}`. Just drop the repo
directory onto a webserver and you have a fully functioning Linux repo. (See
Testing the Repositories below for more detail)

The two major packages included are:
`git-lfs-....*` - the git-lfs package
`git-lfs-repo-release....*` - A package to install the repo.

When building, all **untracked** files are removed during RPM generation (except
any stray directories containing a .git folder will not be cleared. This 
shouldn't be the case, unless you are temporarily storing another git repo in 
the git repo. This is a safety mechanism in git, so just keep in mind if you
are producing packages.)

### Setting the website URL ###

The git-lfs-repo-release must contain the URL where the repo is to be hosted.
The current default value is `git-lfs.github.com` but this can be overridden
using the `REPO_HOSTNAME` env var, e.g.

    export REPO_HOSTNAME=www.notgithub.uk.co 
    ./docker/run_dockers.bsh
    
Now all the `git-lfs-repo-release....*` files will point to that URL instead

_Hint_: `REPO_HOSTNAME` can also be `www.notgithub.uk.co:2213/not_root_dir`

### Testing the Repositories ###

To test that all the OSes can download the packages, install, and run the tests
again, run

    ./test_dockers.bsh
    
(which is basically just `./docker/run_dockers.bsh ./docker/git-lfs-test_*`)

Remember to set `REPO_HOSTNAME` if you changed it for `./docker/build_docker.bsh`
This can also be used to run a local test (on `localhost:{Port Number}`, for
example)

An easy way to test the repositories locally, is to run them on a simple webserver such as

    cd ./repos
    python -m SimpleHTTPServer {Port number}

or

    cd ./repos
    ruby -run -ehttpd . -p{Port Number}

## GPG signing ###

For private repo testing, GPG signing can be skipped. apt-get and yum can 
install .deb/.rpm directly without gpg keys and everything will work (with
certain flags). This section is for distribution in a repo. Most if not all 
this functionality is automatically disabled when there is no signing key 
(`./docker/git-lfs_*.key`).

In order to sign packages, you need to generate and place GPG keys in the right
place. The general procedure for this is

    gpg --gen-key

    1. 4 - RSA
    2. 4096 bits
    3. Some length of time or 0 for infinite
    4. y for yes
    5. Signer name (Will become part of the key and uid)
    6. Email address (Will become part of the key and uid)
    7. Comment (Will become part of the key)
    8. O for Okay
    9. Enter a secure password, make sure you will not forget it
    10. Generate Entropy!
    
    gpg --export-secret-key '<key ID>!' >
    
e.g. `gpg --export-secret-key '547CF247!' > ./docker/git-lfs_centos_7.key`
    
*NOTE*: the **!** is important in this command

Keep in mind, .key files must NEVER be accidentally committed to the repo.

_What if you don't have gpg handy?_ Just enter one of the dockers (-- bash) and
generate them in there, and save them in the /src dir to get them out of the docker.

### GPG Agent ###

To prevent MANY passphrase entries at random times, a gpg-agent docker is used to
cache your signing key. This is done automatically for you, whenever you call
`./docker/run_dockers.bsh` on a building image (`git-lfs_*.dockerfile`). It can
be manually preloaded by calling `./docker/gpg-agent_preload.bsh`. It will ask 
you for your passphrase, once for each unique key out of all the dockers. So if
you use the same key for every docker, it will only prompt once. If you have 5
different keys, you'll have prompts, with only the the key ID to tell you which
is which.

The gpg agent TTL is set to 1 year. If this is not acceptable for you, set the 
`GPG_MAX_CACHE` and `GPG_DEFAULT_CACHE` environment variables (in seconds) before
starting the gpg-agent daemon.

`./docker/gpg-agent_start.bsh` starts the gpg-agent daemon. It is called 
automatically by `./docker/gpg-agent_preload.bsh`

`./docker/gpg-agent_stop.bsh` stops the gpg-agent daemon. It is called 
automatically by `./docker/gpg-agent_preload.bsh`

`./docker/gpg-agent_preload.bsh` is called automatically by 
`./docker/run_dockers.bsh` when running any of the signing dockers. 

`./docker/gpg-agent_preload.bsh -r` - Stops and restarts the gpg agent daemon.
This is useful for reloading keys when you update them in your host.

### GPG capabilities by Distro ###

Every distro has its own GPG signing capability. This is why every signing 
docker (`git-lfs_*.dockerfile`) can have an associated key (`git-lfs_*.key`)

Debian **will** work with 4096 bit RSA signing subkeys like [1] suggests, but will
also work with 4096 bit RSA signing keys.

CentOS will **not** work with subkeys[3]. CentOS 6 and 7 will work with 4096 bit 
RSA signing keys

CentOS 5 will **not** work with v4 signatures. The rpms will be so unrecognizable 
that it can't even be installed with --nogpgcheck. It should work with RSA on v3.
However, it does not. It builds v3 correctly, but for some reason the GPG check
fails for RSA. CentOS 5 will **not** work with 2048 bit DSA keys... I suspect 
2048 is too big for it to fathom. CentOS 5 **will** work with 1024 bit DSA keys.

You can make a 4096 RSA key for Debian and CentOS 6/7 (4 for step 1 above, and
4096 for step 2) and a 1024 DSA key for CentOS 5 (3 for step 1 above, and 1024
for step 2). And only have two keys... Or optionally a 4096 RSA subkey for Debain
[1]. Or a key for each distro. Dealers choice. You should have least two since 
1024 bit isn't that great and you are only using it for CentOS 5 because nothing
else works.

[1] https://www.digitalocean.com/community/tutorials/how-to-use-reprepro-for-a-secure-package-repository-on-ubuntu-14-04

[2] https://iuscommunity.org/pages/CreatingAGPGKeyandSigningRPMs.html#exporting-the-public-gpg-key

[3] http://www.redhat.com/archives/rpm-list/2006-November/msg00105.html


## Adding additional OSes ##

To add another operating system, simply follow the already existing pattern, 
and all the scripts will pick them up. A new dockerfile should be named to

    ./docker/git-lfs_{OS NAME}_{OS VERSION #}.dockerfile
    
where **{OS NAME}** and **{OS VERSION #}** should not contain underscores (\_).
Any files that needs to be added to the docker image must be in the `./docker`
directory. This is the docker context root that all of the dockers are built in.

The docker image should run a script that builds using the files in /src (but
don't modify them...) and write its repo files to the /repo directory inside 
the docker container. Writing to /repo in the docker will cause the files to 
end up in

    ./repos/{OS NAME}/{OS VERSION #}/
    
Unlike standard Dockerfiles, these support two extra features. The first one is
the command `SOURCE`. Similar to `FROM`, only instead of inheriting the image,
it just includes all the commands from another Dockerfile (Minus `FROM` and 
`MAINTAINER` commands) instead. This is useful to make multiple images that work
off of each other without having to know the container image names, and without 
manually making multiple Dockerfiles have the exact same commands.

The second feature is a variable substitution in the form of `[{ENV_VAR_NAME}]`
These will be replaced with values from calling environment or blanked out if
the environment variable is not defined.

## Docker Cheat sheet ##

Install https://docs.docker.com/installation/

* list running dockers

    docker ps
    
* list stopped dockers too

    docker ps -a
    
* Remove all stopped dockers

    docker rm $(docker ps --filter=status=exited -q)
    
* List docker images

    docker images

* Remove unused docker images

    docker rmi $(docker images -a --filter=dangling=true -q)
    
* Run another command (like bash) in a running docker

    docker exec -i {docker name} {command}

* Stopping a docker (signal 15 to the main pid)

    docker stop {docker name}

* Killing a docker (signal 9 to the main pid)

    docker kill {docker name}

# Troubleshooting #

1. I started one of the script, and am trying to stop it with Ctrl+C. It is
ignoring many Ctrl+C's

    This happens a lot when calling programs like apt-get, yum, etc... From the
    host, you can still use ps, pgrep, kill, pkill, etc... commands to kill the
    PIDs in a docker. You can also use `docker ps` to find the container
    name/id and then used `docker stop` (signal 15) or `docker kill`
    (signal 9) to stop the docker. You can also use 'docker exec' to start another
    bash or kill command inside that container
    
2. How do I re-enter a docker after it failed/succeeded?

    Dockers are immediately deleted upon exit. The best way to work in a docker
    is to run bash (See Development in Dockers). This will let you to run the 
    main build command and then continue.
    
3. That answer's not good enough. How do I resume a docker?

    Well, first you have to set the environment variable `AUTO_REMOVE=0` 
    before running the image you want to resume. This will keep the docker 
    around after stopping. (Be careful! They multiply like rabbits.) Then
    
        docker commit {container name/id} {new_name}
    
    Then you can `docker run` that new image.
