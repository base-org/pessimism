echo "Downloading Optimism monorepo..."
git clone https://github.com/ethereum-optimism/optimism.git
cd optimism
git checkout develop

echo "Initializing monorepo..."
make install-geth
make submodules
make devnet-allocs
mv .devnet ../.devnet
mv packages/contracts-bedrock/deploy-config/devnetL1.json ../.devnet/devnetL1.json

echo "Cleaning up..."
cd ../
rm -rf optimism
