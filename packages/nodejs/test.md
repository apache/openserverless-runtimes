export NUV_ROOT=/workspaces/all/olaris
cd /workspaces/all/nuvolaris-runtimes-ng

nuv ide nodejs

## start
nuv ide nodejs clean DIR=$PWD/runtime/nodejs/test/withreqs
ls -l $PWD/runtime/nodejs/test/withreqs.zip

# not found
nuv ide nodejs environment DIR=$PWD/runtime/nodejs/test/withreqs
# built
nuv ide nodejs environment DIR=$PWD/runtime/nodejs/test/withreqs
# not rebuilt

echo "" >>$PWD/runtime/nodejs/test/withreqs/package.json
nuv ide nodejs environment DIR=$PWD/runtime/nodejs/test/withreqs
nuv ide nodejs environment DIR=$PWD/runtime/nodejs/test/withreqs
# modified and rebuilt

nuv ide nodejs compile DIR=$PWD/runtime/nodejs/test/withreqs

unzip -l $PWD/runtime/nodejs/test/withreqs.zip
## end

## start
nuv ide nodejs clean DIR=$PWD/runtime/nodejs/test/multifile
ls -l $PWD/runtime/nodejs/test/multifile.zip
# not found
nuv ide nodejs environment DIR=$PWD/runtime/nodejs/test/multifile
# do nothing

# modified and rebuilt
nuv ide nodejs compile DIR=$PWD/runtime/nodejs/test/multifile
unzip -l $PWD/runtime/nodejs/test/multifile.zip
## end



