# README #

## TL;DR version ##
1. Run the dockers

        ./docker/run_dockers.bsh
        
2. **Enjoy** all your new package files in

        ./repos/

##Using the Dockers##

All these commands need to either be run as root OR as a user in the docker group
Adding your username to the docker group is probably the easiest

### Running Dockers ###

In order to run the dockers, the docker has to be built, and then run with a
lot of arguments to get the mount points right, etc... A convenient script is 
supplied to make this all easy. Simply run

    ./docker/run_docker.bsh
    
All the git-lfs_* images are built automatically, and then run. The default 
command is the build script. The currently checkout version of git-lfs is 
copied into the docker, and `git clean -xdf` is run to remove any 
non-tracked files, but non-committed changes are kept. Then git-lfs is built, 
and a packages is created (deb or rpm)

To only run certain docker images, supply them as arguments, e.g.

    ./docker/run_docker.bsh git-lfs_debian_7
    ./docker/run_docker.bsh git-lfs_debian_*
    ./docker/run_docker.bsh git-lfs_*[6-8]
    
And only those images will be run.

#### Environment Variables ####

There are a few environment variables you can set to easily adjust the behavior
of the `run_docker.bsh` script.

REPO_HOSTNAME - Override the hostname for all the repos generated/tested (see below)

BUILD_LOCAL - Set to 1 (default) to use the currently checked out version of
the git-lfs to build against. If it's not 1 the released archived is downloaded
and built against. Currently only works for RPMs. DEB always builds the currently
checkout version. Build local only affect the version of the code used in generating
the rpms, not the scripts running to generate the rpms (e.g. `./rpm/build_rpms.bsh`)

DOCKER_AUTOBUILD - Default 1. `run_docker.bsh` always calls `build_docker.bsh` to
ensure your docker image is up-to-date. Sometimes you may not want this. If set
this to 0, it will not build docker images before running

AUTO_REMOVE - Default 1. Docker containers are automatically deleted on 
exit. If set to 0, the docker containers will not be automatically deleted upon 
exit. This can be useful for a post mortem analysis (using other docker commands
not covered here). Just make sure you clean up the docker containers manually.

#### Build Environment Variables ####

These can be before calling `run_docker.bsh`, they are actuallys just being 
passed to `build_docker.bsh`. 

LFS_VERSION - The version of LFS used to bootstrap the environment. This does 
not need to be bumped every version (but can be). Currently set to track git-lfs.go 
version.



###Development with Dockers###

Sometimes you don't want to just build git-lfs and destroy the container, you
want to get in there, run a lot of command, DEVELOP! To do this, the best
command to run is bash, and then you have an interactive shell to use. To do this

    ./docker/run_docker.bsh {image(s)} -- bash

After listing the image(s) you want to run, add a double dash (--) and then any 
command (and arguments) you want executed in the docker. Remember, the command
you are executing has to be part of the docker image.

### Building Dockers (Optional) ###

`run_dockers.bsh` calls `build_dockers.bsh`, but you can still call the
script manually to get it all out of the way once while you go make some 
tea/coffee.

In order to use the docker **images**, they have to be built so that they are
ready to be used. For OSes like Debian, this is a fairly quick process. 
However CentOS takes considerably longer time, since it has to build go, ruby,
or git from source, depending on the distro. Fortunately, you can build the 
docker images JUST once, and you won't have to build it again (until the version
changed.) The build script uses a downloaded release from github of git-lfs to
bootstrap the CentOS image and build/install all the necessary software. Currently
the only way to change what version the image is built off of is by changing the
URL in the Dockerfile for git-lfs_centos_*. 

This means all the compiling, yum/apt-get/custom dependency compiling is done 
once and saved. (This is done in CentOS by using the already existing 
`./rpm/rpm_build.bsh` script to bootstrap the image and saving the image.)

The script that takes care of ALL of these details for you is

    ./docker/build_dockers.bsh
    
All the git-lfs_* images will be built automatically. These are all a
developer would need to test the different OSes. And create the git-lfs rpm or
deb packages.

However, in order to distribute git-lfs or build dependencies, the packages 
that were installed for you by `build_docker.bsh` need to be saved too.
This is currently only a CentOS problem. In order to generate THOSE rpms, 
the git-lfs-full-build_* will use NON-bootstrapped images to build every package
and git-lfs and generate rpms. This takes as long as building the image in the
first place, but you don't get the benefit of a saved state. The 
`./rpm/rpm_build.bsh` script will build all of its dependencies when you
run the dockers, making the rpms available. These images can be built by running

    ./docker/docker_build.bsh ./docker/git-lfs-full-build_*

This is most important for CentOS 6 where git 1.8.2 or newer is not available, 
only git 1.7.1 is available, so every user either has to build git from source, 
or use the rpms generated by the `git-lfs-full-build_centos_6` image. This 
will only needs to be done once.

(To manually build a docker image, run 
`docker build -t $(basename ${DockerDir}) -f ${DockerDir}/Dockerfile ./docker`


##Deploying/Building Repositories##

When `./docker/run_dockers.bsh` is done building git-lfs and the rpms/deb,
it actually creates a repository for distribution too. Each distro gets a repo
generated in `./repos/{DISTRO_NAME}/{VERSION #}`. Just drop the repo
directory onto a webserver and you have a fully functioning Linux repo. (See
Testing the Repositories below for more detail)

The two major packages included are:
git-lfs-....* - the git-lfs package
git-lfs-repo-release....* - A package to install the repo.

When using `BUILD_LOCAL=1`, all UNTRACKED files are removed during RPM 
generation (except any stray directories containing a .git folder will not be
cleared. This shouldn't be the case, unless you are temporarily storing another
git repo in the git repo. This is a safety mechanism in git, so just keep in mind
if you are producing packages.)

### Setting the website URL ###

The git-lfs-repo-release must contain the URL where the repo is to be hosted.
The current default value is 'git-lfs.github.com' but this can be overridden
using the REPO_HOSTNAME env var, e.g.

    REPO_HOSTNAME=www.notgithub.uk.co ./docker/run_dockers.bsh
    
Now all the git-lfs-repo-release....* files will point to that URL instead

### GPG signing ###

For private repo testing, GPG signing can be skipped. apt-get and yum can 
install .deb/.rpm directly without gpg keys and everything will work. This 
section is for distribution in a repo. Most if not all this functionality is 
automatically disabled when there is no signing key present
(`./docker/signing.key`).

In order to sign packages, you need to generate and place GPG keys in the right place

1. gpg --gen-key

    1. 1 - RSA and RSA
    2. 4096 bits
    3. Some length of time or 0 for infinite
    4. y for yes
    5. Signer name (Will become part of the key and uid)
    6. Email address (Will become part of the key and uid)
    7. Comment (Will become part of the key)
    8. O for Okay
    9. Enter a very secure password, make sure you will not forget it
    10. Generate Entropy!
    
2. gpg -a --export > ./docker/public.key

3. gpg -a --export-secret-keys > ./docker/signing.key

Keep in mind, signing.key must NEVER be accidentally committed to the repo. 

To prevent MANY passphrase entries at random times, the gpg-agent is used to
cache your signing key. This is done by running gpg-agent in the host, and passing
the connection to each docker image. This will be done for you automatically by
calling the `./docker/preload_key.bsh` script. This can be called manually
before any other command just to get the pass phrase entry out of the way before
you start running everything.

GPG agent TTL is set to 5 hours. This should be plenty to build everything. If this is
not good for you, set the GPG_MAX_CACHE and GPG_DEFAULT_CACHE environment variables
(in seconds)

### GPG capabilities by Distro ###

Debian WILL work with 4096 bit RSA signing subkeys like [1] suggests, but will
also work with 4096 bit RSA signing keys.

CentOS will NOT work with subkeys[3]. CentOS 6 and 7 will work with 4096 bit RSA 
signing keys

CentOS 5 will NOT work with v4 signatures. The rpms will be so unrecognizable 
that it can't even be installed with --nogpgcheck. It should work with RSA on v3.
However, I could not get it to. It builds v3 correctly, but for some reason
the GPG check fails for RSA. However `yum install --nogpgcheck` does work! CentOS 5
will NOT work with 2048 bit DSA keys... I suspect 2048 is too big for it to
fathom. CentOS 5 WILL work with 1024 bit DSA keys. Either sign it with the key
instructions above and install it with the `yum install --nogpgcheck` OR
create a NEW DSA key just for CentOS 5 and sign with that. 

[1] https://www.digitalocean.com/community/tutorials/how-to-use-reprepro-for-a-secure-package-repository-on-ubuntu-14-04

[2] https://iuscommunity.org/pages/CreatingAGPGKeyandSigningRPMs.html#exporting-the-public-gpg-key

[3] http://www.redhat.com/archives/rpm-list/2006-November/msg00105.html

### Testing the Repositories ###

To test that all the OSes can download the rpm/debs, install, and run the tests
again, run

    ./test_dockers.bsh
    
(which is basically just `./docker/run_dockers.bsh ./docker/git-lfs-test_*`)

Remember to set REPO_HOSTNAME if you changed it for `./docker/build_docker.bsh`
This can also be used to run a local test (on `localhost:{Port Number}`, for
example)

An easy way to test the repositories locally, is to run them on a simple webserver such as

    cd ./repos
    python -m SimpleHTTPServer {Port number}

or

    cd ./repos
    ruby -run -ehttpd . -p{Port Number}


## Adding addition OSes ##

To add another operating system, simply follow the already existing pattern, 
and all the scripts will pick them up. A new Dockerfile should go in a directory
named

    ./docker/git-lfs_{OS NAME}_{OS VERSION #}
    
where **{OS NAME}** and **{OS VERSION #}** should not contain underscores (`_`).
Any files that needs to be added to the docker image can be dropped in the 
`./docker` directory, since that is the context root they are built against 
(not the directory containing the Dockerfile like most dockers)

The docker image should run a script that write it's repo files to the /repo
directory inside the docker container. Writing to /repo in the docker will cause the
files to end up in

    ./repos/{OS NAME}/{OS VERSION #}/
    
Unlike standard Dockerfiles, these support two extra features. The first one is
the command `SOURCE`. Similar to `FROM`, only instead of inheriting the image,
it includes all the commands from another Dockerfile (Minus other `FROM` and 
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


# Troubleshooting #

1. I started one of the script, and am trying to stop it with Ctrl+C. It is
ignoring many Ctrl+C's

    This happens a lot when calling programs like apt-get, yum, etc... From the
    host, you can still use ps, pgrep, kill, pkill, etc... commands to kill the
    PIDs in a docker. You can also use `docker ps` to find the container
    name/id and then used `docker stop` (signal 15) or `docker kill`
    (signal 9) to stop the docker
    
2. How do I re-enter a docker after it failed/succeeded?

    Dockers are immediately deleted upon exit. The best way to work in a docker
    is to run bash. This will let you to run the main build command and then
    continue.
    
3. That answer's not good enough. How do I resume a docker?

    Well, first you have to set the environment variable `AUTO_REMOVE=0` 
    before running the image you want to resume. This will keep the docker 
    around after stopping. (Be careful! They multiply like rabbits.) Then
    
        docker commit {container name/id} {new_name}
    
    Then you can `docker run` that new image.

4. Everything in the ./repos directory is owned by root

    That is currently a side effect of the dockers being run as root.
