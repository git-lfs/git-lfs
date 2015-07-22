# README #

## TL;DR version ##
1. Run the dockers

        ./docker/run_dockers.bsh
        
2. ???

3. **Enjoy** all your new package files in

        ./docker/repos/

##Using the Dockers##

### Building Dockers (Optional) ###

```run_dockers.bsh``` now call build_dockers ```build_dockers.bsh```, but you
can still call the script manually to get it all out of the way once while you
go make that tea/coffee.

In order to use the docker **images**, they have to be built so that they are
ready to be used. For OSes like Debian, this is a fairly quick process. 
However CentOS takes considerably longer time, since it has to build go, ruby,
or git from source, depending on the version. Fortunately, you can build the 
docker images JUST once, and you won't have to build it again (unless something
significant changes, which should be fairly uncommon). This means all the 
compiling, yum/apt-get/custom dependency compiling is done once and saved. 
(This is done in CentOS by using the already existing ```./rpm/rpm_build.bsh```
script to bootstrap the image and saving the image.)

There is a script to take care of ALL of these details for you. Simply run

    ./docker/build_dockers.bsh
    
And all the git-lfs_* images will be built automatically. These are all a
developer would need to test the different OSes. 

If you were more interested in creating all the special packages for CentOS 
(For example, you may want to hand out the git rpm for CentOS 6, since only 
git 1.7.1 is available, or else everyone will have to build git from source 
just to use git-lfs on CentOS 6), then the git-lfs-full-build_* dockers are
useful. Instead of building everyone necessary into the image, these contain
bare CentOS image, so the ```./rpm/rpm_build.bsh``` script will build all of
its dependencies when you run the dockers, making the rpms available. These
images can be build by running

    ./docker/docker_build.bsh ./docker/git-lfs-full-build_*
    
In fact, any subset of images can be built only, by passing the list of of 
directories as argument. (Each directory contains a Dockerfile, this is used to
tell the .bsh scripts which you want to work on AND names them in 
```docker images``` based on the directory name.)

(To manually build a docker, run ```docker build -t $(basename ${DockerDir}) 
-f ${DockerDir}/Dockerfile ./docker```

### Running Dockers ###

After the docker images are build, a lot of arguments need to me added to get
the mount points right, etc... again, a convenient script is supplied to make
this all easy. By running

    ./docker/run_docker.bsh
    
All the git-lfs_* images build their packages and tranfer them to the 
```./docker/repos``` directory.


Copies your current source
Cleans the copies, so all untracked files are deleted, but uncommited changes are kept

###Development with Dockers###

##Deploying/Building Repositories##

When using ```BUILD_LOCAL=1```, all UNTRACKED files are removed during RPM 
generation, except any stray directories containing a .git folder will not be
cleared. This shouldn't be the case, unless you are temporarily storing another
git repo in the git repo. This is a safty mechanism in git, so just keep in mind
if you are producing packages.

### Setting the website URL ###

### GPG signing ###

For private repo testing, GPG signing can be skipped. apt-get and yum can 
install .deb/.rpm directly without gpg keys and everything will work. This 
section is for distribution in a repo. Most if not all this functionality is 
automatically disabled when there is no signing key present.

Or order to sign packages, you need to place the keys in the right place

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

3. gpg --export-secret-keys > ./docker/signing.key

Keep in mind, signing.key must NEVER be accidentally commited to the repo. 

Signing in CentOS 5 is... problemsome. While it can not handle v4 signatures,
it should work with RSA on v3. However, it doesn't. So... The only way around it
is to either sign it with the key instructinos above aand install it with the
```yum install --nogpgcheck``` OR create a NEW DSA key just for CentOS 5. Even 
then, the default size of 2048 did not work. A DSA key of 1024 bits does work. 

To prevent MANY passphrase entries at random times, the gpg-agent is used to
cache your signing key. This is done by running gpg-agent in the host, and passing
the connection to each docker image. This will be done for you automatically by
calling the ```./docker/preload_key.bsh``` script. This can be called manually
before any other command just to get the pass phrase entry out of the way before
you start running everything.

GPG agent ttl set to 5 hours, should be plenty to build everything. If this is
not good for you, set the GPG_MAX_CACHE and GPG_DEFAULT_CACHE environment variables
(in seconds)

[1] https://www.digitalocean.com/community/tutorials/how-to-use-reprepro-for-a-secure-package-repository-on-ubuntu-14-04
[2] https://iuscommunity.org/pages/CreatingAGPGKeyandSigningRPMs.html#exporting-the-public-gpg-key
[3] http://www.redhat.com/archives/rpm-list/2006-November/msg00105.html
- Rpms do NOT SUPPORT subkeys. So don't try

### Testing the Repositories ###

To test that all the OSes can download the rpm/debs, install, and run the tests
again, run

    ./test_dockers.bsh
    
(which is basically just ```./docker/run_dockers.bsh ./docker/git-lfs-test_*```)

REPO_HOSTNAME can be used for BOTH ```run_dockers.bsh``` and ```test_dockers.bsh``` 
to run a local test (on ```localhost:{Port Number}```, for example)

An easy way to test the repositories, is to run host them on a webserver such as

    cd ./docker/repos
    python -m SimpleHTTPServer {Port number}

or

    cd ./docker/repos
    ruby -run -ehttpd . -p{Port Number}


## Adding addition OSes ##

To add another operating system, simply follow the already existing pattern, 
and all the script will them them up. A new Dockerfile should go in a directory
named

    ./docker/git-lfs_{OS NAME}_{OS VERSION}
    
where **{OS NAME}** and **{OS VERSION}** should not contain underscores (_).
Any files that need to be added to the docker image can be dropped in the 
```./docker``` directory, since that is the root they are built against (not 
the directory containing the Dockerfile)

The docker image should write it's repo files to /repo inside the docker, and
they will end up in

    ./docker/repos/{OS NAME}/{OS VERSION}/

## Docker Cheat sheet ##

http://docs.docker.com/ Install -> Docker Engine -> Installation on ...

* list running dockers

    docker ps
    
* list stopped dockers

    docker ps -a
    
* Remove all stopped dockers

    docker rm $(docker ps --filter=status=exited -q)

1. How much space are all these Dockers taking up?

    No idea. sudo du /var/lib/docker


# Troubleshooting #

1. I started one of the script, and am trying to stop it with Ctrl+C. It is
ignoring many Ctrl+C's

    This happens a lot when calling programs like apt-get, yum, etc... From the
    host, you can still use ps, pgrep, kill, pkill, etc... commands to kill the
    PIDs in a docker.
    
2. How do I re-enter a docker after it failed/suceeded?

    Dockers are immediately deleted upon exit. The best way to work in a docker
    is to run bash. This will let you to run the main build command and then
    continue.
    
3. That answer's not good enough. How do I resume a docker?

    Well, first you have to remove the ```--rm``` flag. This will keep the 
    docker around after stopping. Be careful! They multiply like rabbits. Then
    
        ```docker commit {container name/id} {new_name}```
    
    Then you can ```docker run``` that new image.
    