#!/bin/bash

VERSION=$(cat go.mod | grep ethereum-optimism/optimism | awk '{print $2}' | sed 's/\/v//g')

REPO_NAME=optimism-$(echo ${VERSION})

echo "Downloading ${REPO_NAME} ..."

## commit hash 
if [ ${#$(echo $VERSION)} -gt 6] ; then
    git clone --branch ${VERSION} https://github.com/ethereum-optimism/optimism.git ${REPO_NAME}

## version tag
else 
    git clone --branch v${VERSION} https://github.com/ethereum-optimism/optimism.git ${REPO_NAME}
fi


cd ${REPO_NAME}

echo "Initializing monorepo..."
make install-geth &&
git submodule update --init --recursive &&
make devnet-allocs &&
cp -R .devnet ../. &&
mv packages/contracts-bedrock/deploy-config/devnetL1.json ../.devnet/devnetL1.json

STATUS=$?

## Force cleanup of monorepo 
echo "${STATUS} Cleaning up ${REPO_NAME} repo ..."
cd ../ &&
rm -rf ${REPO_NAME}

if [ $? -eq 0 ] ; then
    echo "Successfully cleaned up ${REPO_NAME} repo"
    exit ${STATUS}
else
    echo "Failed to clean up ${REPO_NAME} repo"
    exit ${STATUS}
fi
