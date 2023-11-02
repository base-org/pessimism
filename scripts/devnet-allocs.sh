## (1) Fetch monorepo binary at specific version used by Pessimism
VERSION=$(cat go.mod | grep ethereum-optimism/optimism | awk '{print $2}' | sed 's/\/v//g')
REPO_NAME=optimism-$(echo ${VERSION} | sed 's/v//g')

echo "Downloading ${REPO_NAME} ..."
wget https://github.com/ethereum-optimism/optimism/archive/refs/tags/${VERSION}.zip

## (2) Unzip and enter the monorepo
echo "Unzipping..."
unzip ${VERSION}.zip
rm -rf ${VERSION}.zip

## (3) Get version string without first 'v'
VERSION=$(echo ${VERSION} | sed 's/v//g')
echo "Version: ${VERSION}"
cd optimism-${VERSION}

## (4) Install monorepo dependencies
{
    ## (4.a) Generate devnet allocations and persist them all into .devnet folder
    echo "Initializing monorepo..." &&
    git submodule update --init --recursive &&
    make devnet-allocs &&
    mv .devnet ../.devnet &&
    mv packages/contracts-bedrock/deploy-config/devnetL1.json ../.devnet/devnetL1.json

    STATUS = $?
} ; {
    ## (4.b) Force cleanup of monorepo 
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
}
