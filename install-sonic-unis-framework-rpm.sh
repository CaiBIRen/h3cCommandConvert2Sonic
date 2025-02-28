#!/bin/bash

# default
BRANCH="develop"
Version="develop"
Tag=`date | sed 's/[[:space:]]//g' | sed 's/[:]//g'`

while getopts "v:t:b:" arg
do
        case $arg in
             v)
                Version=$OPTARG
                echo "v's arg:$OPTARG" ;;
             t)
                Tag=$OPTARG
                echo "t's arg:$OPTARG" ;;
		
             b)
                BRANCH=$OPTARG
                echo "b's arg:$OPTARG";;
             ?)
                #当有不认识的选项的时候arg为?
                echo "unkonw argument" exit 1 ;;
         esac
done


SOURCE_DIR="sonic-unis-framework"
path=`dirname $0`
echo $path

export AGENT_NAME="sonic-unis-framework"
export AGENT_VERSION=$Version
export AGENT_TAG=$Tag
export RPM_BUILD_ROOT="\${RPM_BUILD_ROOT}"

# #cd $SOURCE_DIR
# git checkout $BRANCH
# git pull

#cd 
#cd ..

echo `dirname $0`

# exchange spec info
echo `export | grep AGENT`

rm sonic-unis-framework -f
#rm $path/$SOURCE_DIR/Makefile
envsubst < ./spec/template.spec > ./sonic-unis-framework.spec
#envsubst < $path/$SOURCE_DIR/template.mk > $path/$SOURCE_DIR/Makefile

# tar gz
TARNAME=$AGENT_NAME-$AGENT_VERSION.tar.gz
echo $TARNAME

cd ../
tar czf $TARNAME ./$SOURCE_DIR/
cd -

echo "begin move tar to rmpbuild sources"
mkdir -p ~/rpmbuild
mkdir -p ~/rpmbuild/SOURCES
mv ../$TARNAME ~/rpmbuild/SOURCES/

# rpmbuild
# echo "begin rpmbuild -ba sonic-unis-framework.spec"
# rpmbuild -ba sonic-unis-framework.spec

# clean

