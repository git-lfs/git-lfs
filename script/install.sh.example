prefix="/usr/local"

if [ "$PREFIX" != "" ] ; then
  prefix=$PREFIX
elif [ "$BOXEN_HOME" != "" ] ; then
  prefix=$BOXEN_HOME
fi

mkdir -p $prefix/bin

rm -rf $prefix/bin/git-lfs*
for g in git*; do
  cp $g "$prefix/bin/$g"
done

git lfs init
