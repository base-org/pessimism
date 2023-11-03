#!/bin/bash

## (1) Fetch monorepo binary at specific version used by Pessimism
VERSION=$(cat go.mod | grep ethereum-optimism/optimism | awk '{print $2}' | sed 's/\/v//g')
REPO_NAME=optimism-$(echo ${VERSION} | sed 's/v//g')

echo "Downloading ${REPO_NAME} ..."
git clone https://github.com/ethereum-optimism/optimism.git ${REPO_NAME}

## (2) Unzip and enter the monorepo
echo "Unzipping..."
unzip ${VERSION}.zip
rm -rf ${VERSION}.zip

## (3) Get version string without first 'v'
VERSION=$(echo ${VERSION} | sed 's/v//g')
echo "Version: ${VERSION}"
cd ${REPO_NAME}
git checkout ${VERSION}

## (4) Install monorepo dependencies
echo "Initializing monorepo..."
make install-geth &&
git submodule update --init --recursive &&
make devnet-allocs &&
mv .devnet ../.devnet &&
mv packages/contracts-bedrock/deploy-config/devnetL1.json ../.devnet/devnetL1.json

STATUS=$?

## (5) Force cleanup of monorepo 
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
