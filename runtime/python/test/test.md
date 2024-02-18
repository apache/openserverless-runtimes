export NUV_ROOT=/workspaces/all/olaris
cd /workspaces/all/nuvolaris-runtimes-ng

nuv ide python

## start
nuv ide python clean DIR=$PWD/runtime/python/test/withreqs
ls -l $PWD/runtime/python/test/withreqs.zip

# not found
nuv ide python environment DIR=$PWD/runtime/python/test/withreqs
# built
nuv ide python environment DIR=$PWD/runtime/python/test/withreqs
# not rebuilt

echo "" >>$PWD/runtime/python/test/withreqs/requirements.txt
nuv ide python environment DIR=$PWD/runtime/python/test/withreqs
nuv ide python environment DIR=$PWD/runtime/python/test/withreqs
# modified and rebuilt

nuv ide python compile DIR=$PWD/runtime/python/test/withreqs

unzip -l $PWD/runtime/python/test/withreqs.zip
## end

## start
nuv ide python clean DIR=$PWD/runtime/python/test/multifile
ls -l $PWD/runtime/python/test/multifile.zip

# not found
nuv ide python environment DIR=$PWD/runtime/python/test/multifile
# built
nuv ide python environment DIR=$PWD/runtime/python/test/multifile
# not rebuilt

# modified and rebuilt
nuv ide python compile DIR=$PWD/runtime/python/test/multifile
unzip -l $PWD/runtime/python/test/multifile.zip
## end

## start
nuv ide python clean DIR=$PWD/runtime/python/test/multifile
ls -l $PWD/runtime/python/test/multifile.zip

# not found
nuv ide python environment DIR=$PWD/runtime/python/test/multifile
# built
nuv ide python environment DIR=$PWD/runtime/python/test/withreqs
# not rebuilt

echo "" >>$PWD/runtime/python/test/withreqs/requirements.txt
nuv ide python environment DIR=$PWD/runtime/python/test/withreqs
nuv ide python environment DIR=$PWD/runtime/python/test/withreqs
# modified and rebuilt

nuv ide python compile DIR=$PWD/runtime/python/test/withreqs

unzip -l $PWD/runtime/python/test/withreqs.zip
## end





