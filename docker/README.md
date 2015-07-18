# README #

Document
DOCKER_CMD

## TL;DR version ##
1. Build the docker images

        ./docker/build_dockers.bsh

2. Run the dockers

        ./docker/run_dockers.bsh

3. **Enjoy** all your new package files in

        ./docker/repos/

##Using the Dockers##

### Building Dockers ###

In order to use the docker **images**, they have to be built them so that they
are ready to be used. For OSes like Debian, this is a fairly quick process. 
However CentOS takes considerably longer time, since it has to build go, ruby,
or git from source, depending on the version. Fortunately, you can build the 
docker image JUST once, and you won't have to build it again (unless something
significant changes, which should be fairly uncommon). This means all the 
compiling, yum/apt-get is done once and saved. (This is done in CentOS by
running the ```./rpm/rpm_build.bsh``` script and saving the image.)

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
