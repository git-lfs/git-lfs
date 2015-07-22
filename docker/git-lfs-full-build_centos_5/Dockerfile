FROM centos:5
MAINTAINER Andy Neff <andyneff@users.noreply.github.com>

#Docker RUN example, pass in the git-lfs checkout copy you are working with
LABEL RUN="docker run -v git-lfs-repo-dir:/src" -v repo_dir:/repo"


RUN yum install -y createrepo rsync gnupg gnupg2 expect

#Add the simple build repo script
ADD rpm_sign.exp signing.key centos_script.bsh /tmp/

CMD /tmp/centos_script.bsh