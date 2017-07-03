# README #

## TL;DR version ##
1. Run the dockers

        ./docker/run_dockers.bsh
        
2. **Enjoy** all your new package files in

        ./repos/
        
### Slightly longer version ###

1. Generate GPG keys for everything (See GPG Signing)
2. `export REPO_HOSTNAME=repo.something.com`
4. Generate git-lfs/repo packages and sign all packages

        ./docker/run_dockers.bsh
        
5. Host the `/repo` on the `REPO_HOSTNAME` server
6. Test the repos and git-lfs in a client environment

        ./docker/test_dockers.bsh

## Using the Dockers ##

All docker commands need to either be run as root **or** as a user with docker 
permissions. Adding your user name to the docker group (or setting up boot2docker 
environment) is probably the easiest.

For Mac and Windows users, the git-lfs repo needs to be in your Users directory 
or else boot2docker magic won't work. Alternatively, you could add addition
mount points like 
[this](http://stackoverflow.com/questions/26639968/boot2docker-startup-script-to-mount-local-shared-folder-with-host)

### Running Dockers ###

In order to run the dockers, the docker has to be run with a
lot of arguments to get the mount points right, etc... A convenient script is 
supplied to make this all easy. Simply run

    ./docker/run_docker.bsh
    
All the images are pulled automatically, and then run.

To only run certain docker images, supply them as arguments, e.g.

    ./docker/run_docker.bsh debian_7
    ./docker/run_docker.bsh centos_7 debian_8
    ./docker/run_docker.bsh centos_{6,7}

And only those images will be run.

### Development in Dockers ###

Sometimes you don't want to just build git-lfs and destroy the container, you
want to get in there, run a lot of command, debug, develop, etc... To do this, 
the best command to run is bash, and then you have an interactive shell to use

    ./docker/run_docker.bsh {image name(s)} -- bash

After listing the image(s) you want to run, add a double dash (--) and then any 
command (and arguments) you want executed in the docker. Remember, the command
you are executing has to be in the docker image.

## Docker images ##

There are currently three type of docker images:

1. Building images: `{OS NAME}_{OS VERSION}` - These build
git-lfs and save the package/repository in the `/repo` direrctory. This image
also signs all rpms/debs if gpg signing is setup
2. Environment building images: `{OS_NAME}_{OS_VERSION}_env` -
These build or install the environment (dependencies) for building git-lfs. These 
are mostly important for CentOS because without these, many dependencies have 
to be built by a developer. These containers should create packages for these 
dependencies and place them in `/repo`
3. Testing images: `{OS_NAME}_{OS_VERSION}_test` - These images
should install the repo and download the git-lfs packages and dependencies to test
that everything is working, including the GPG signatures. Unlike the first two types,
testing images are not guaranteed to work without GPG signatures. They should 
also run the test and integration scripts after installing git-lfs to verify
everything is working in a **non-developer** setup. (With the exception that go
is needed to build the tests...)

This default behavior for `./docker/run_dockers.bsh`
is to run all of the _building images_. These
containers will use the currently checked-out version of git-lfs and copy it 
into the docker, and run `git clean -xdf` to remove any non-tracked files, 
(but non-committed changes are kept). git-lfs is built, and a packages/repo is 
created for each container.

These are all a developer would need to test the different OSes. And create the
git-lfs rpm or deb packages in the `/repo` directory. 

In order to distribute git-lfs **and** build dependencies, the dependencies that 
that were built to create the docker images need to be saved too. Most of these
are downloaded by yum/apt-get and do not need to be saved, but a few are not.
In order to save the necessary dependencies, call `./docker/run_dockers.bsh` on 
`{OS_NAME}_{OS_VERSION}_env` and the rpms 
will be extracted from the images and saved in the `./repo` directory.
(This _can_ be done in one command)

    ./docker/run_dockers.bsh centos_6_env centos_6

This isn't all that important anymore, unless you want ruby2 and the gems used to
make the man pages for CentOS 6 where ruby2 is not natively available. Calling
the environment building images only needs to be done once, they should remain in
the `./repo` directory afterwards.

### Run Docker Environment Variables ###

There are a few environment variables you can set to easily adjust the behavior
of the `run_docker.bsh` script.

`export` before calling `run_docker.bsh`

`REPO_HOSTNAME` - Override the hostname for all the repos generated/tested (see below)

`DOCKER_AUTOPULL` - Default 1. `run_docker.bsh` always pulls the latest version of
the lfs dockers. If set to 0, it will not check to see if a new pull is needed,
and you will always run off of your currently cached images docker images.

`AUTO_REMOVE` - Default 1. Docker containers are automatically deleted on 
exit. If set to 0, the docker containers will not be automatically deleted upon 
exit. This can be useful for a post mortem analysis (using other docker commands
not covered here). Just make sure you clean up the docker containers manually.

`DOCKER_OTHER_OPTIONS` - Any additional arguments you may want to pass to the
docker run command. This can be particularly useful when having to help docker
with dns, etc... For example `DOCKER_OTHER_OPTIONS="--dns 8.8.8.8"`

If for some reason on Windows, you need to add a -v mount, folder names need to
start with `//driveleter/dir...` instead of `/driveleter/dir...` to fool MINGW32

## Deploying/Building Repositories ##

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
    
    gpg --export-secret-key '<key ID>!' > filename.key
    
e.g. `gpg --export-secret-key '547CF247!' > ./docker/git-lfs_centos_7.key`
    
*NOTE*: the **!** is important in this command

Keep in mind, .key files must NEVER be accidentally committed to the repo.

_What if you don't have gpg handy?_ Just enter one of the dockers (-- bash) and
generate them in there, and save them in the /src dir to get them out of the docker.
Or `docker run -it --rm -v $(pwd):/key OS_NAME:OS_VERSION bash`, and generate in
that docker and save to the `/key` directory

### GPG Agent ###

To prevent MANY passphrase entries at random times, a gpg-agent docker is used to
cache your signing key. This is done automatically for you, whenever you call
`./docker/run_dockers.bsh` on a building image (`git-lfs_*.dockerfile`). It can
be manually preloaded by calling `./docker/gpg-agent_preload.bsh`. It will ask 
you for your passphrase, once for each unique key out of all the dockers. So if
you use the same key for every docker, it will only prompt once. If you have 5
different keys, you'll have prompts, with only the key ID to tell you which
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

You can make a 4096 RSA key for Debian and CentOS 6/7 (4 for step 1 above, and
4096 for step 2). And only have two keys... Or optionally a 4096 RSA subkey for Debain
[1]. Or a key for each distro. Dealers choice.

[1] https://www.digitalocean.com/community/tutorials/how-to-use-reprepro-for-a-secure-package-repository-on-ubuntu-14-04

[2] https://iuscommunity.org/pages/CreatingAGPGKeyandSigningRPMs.html#exporting-the-public-gpg-key

[3] http://www.redhat.com/archives/rpm-list/2006-November/msg00105.html


## Adding additional OSes ##

To add another operating system,  it needs to be added to the lfs_dockers 
repo and uploaded to docker hub. Then all that is left is to add it to the 
IMAGES list in `run_dockers.bsh` and `test_dockers.bsh`

Follow the already existing pattern `{OS NAME}_{OS VERSION #}` where 
**{OS NAME}** and **{OS VERSION #}** should not contain underscores (\_).

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
