#!/bin/bash

## Fetch monorepo binary at specific version used by Pessimism
VERSION=$(cat go.mod | grep ethereum-optimism/optimism | awk '{print $2}' | sed 's/\/v//g')
VERSION=$(echo ${VERSION} | sed 's/v//g')

REPO_NAME=optimism-$(echo ${VERSION} | sed 's/v//g')

echo "Downloading ${REPO_NAME} ..."
git clone --branch v${VERSION} https://github.com/ethereum-optimism/optimism.git ${REPO_NAME}

VERSION=$(echo ${VERSION} | sed 's/v//g')
echo "Version: ${VERSION}"
cd ${REPO_NAME}

echo "Initializing monorepo..."
make install-geth &&
git submodule update --init --recursive &&
make devnet-allocs &&
cp -R .devnet ../. &&
mv packages/contracts-bedrock/deploy-config/devnetL1.json ../.devnet/devnetL1.json

STATUS=$?

## Force cleanup of monorepo 
echo "${STATUS} Cleaning up ${REPO_NAME} repo ..." &&
cd ../ &&
rm -rf ${REPO_NAME}

if [ $? -eq 0 ] ; then
    echo "Successfully cleaned up ${REPO_NAME} repo"
    exit ${STATUS}
else
    echo "Failed to clean up ${REPO_NAME} repo"
    exit ${STATUS}

fi
