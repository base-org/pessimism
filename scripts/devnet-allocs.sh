## Download and enter the OP monorepo
echo "Downloading Optimism monorepo..."
git clone https://github.com/ethereum-optimism/optimism.git
cd optimism
git checkout develop

## Generate devnet allocations and persist them all into .devnet folder
echo "Initializing monorepo..."
make install-geth
git submodule update --init --recursive
make devnet-allocs
mv .devnet ../.devnet
mv packages/contracts-bedrock/deploy-config/devnetL1.json ../.devnet/devnetL1.json

## Clean up
echo "Cleaning up..."
cd ../
rm -rf optimism
