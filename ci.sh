#!/bin/bash



branch_name=$(git symbolic-ref --short -q HEAD)

if [[ $branch_name == "" ]]; then
  source ./scripts/cihelper.sh
  branch_name=$CODEBUILD_GIT_BRANCH
fi

echo "Building for branch:" $branch_name
ipfs_built="not yet"
tipfs_built="not yet"

build_ipfs(){
  if [[ $ipfs_built == "yes" ]]; then
    return
  fi
  echo "building IPFS node"
  docker build -t tezoscommons/tezos-ipfs:ipfs-s3-$branch_name ./docker/ipfs-s3
  docker push tezoscommons/tezos-ipfs:ipfs-s3-$branch_name
  ipfs_built="yes"
}

build_tipfs(){
  if [[ $tipfs_built == "yes" ]]; then
    return
  fi
  docker build -t tezoscommons/tezos-ipfs:$branch_name ./
  docker push tezoscommons/tezos-ipfs:$branch_name
  tipfs_built="yes"
}

# build ipfs on any branch containing ipfs
if [[ $branch_name == *"ipfs"* ]]; then
  build_ipfs
fi

# build app on any release branch
if [[ $branch_name == "v"* ]]; then
  build_tipfs
fi

# build on a few special branches
if [[ $branch_name == "master" || $branch_name == "testing" || $branch_name == "mvp" ]]; then
  build_tipfs
fi