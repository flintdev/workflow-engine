#!/usr/bin/env bash

checkIfXcodeCommandLineToolsInstalled() {
  xcode-select -p &> /dev/null
  xcodeCommandLineToolsStatus=$?
  if [ "$xcodeCommandLineToolsStatus" == "0" ]; then
    isXcodeCommandLineToolsInstalled=true
  else
    isXcodeCommandLineToolsInstalled=false
  fi
}

checkIfMercurialInstalled() {
  brew list mercurial &> /dev/null
  mercurialStatus=$?
  if [ "$mercurialStatus" == "0" ]; then
    isMercurialInstalled=true
  else
    isMercurialInstalled=false
  fi
}

checkIsHomebrewInstalled() {
  if [ -x "$(command -v brew)" ]; then
    isHomebrewInstalled=true
  else
    isHomebrewInstalled=false
  fi
}

checkIsGVMInstalled() {
  # shellcheck source=src/lib.sh
  source "$HOME/.gvm/scripts/gvm" &> /dev/null
  gvm version &> /dev/null
  gvmVersionStatus=$?
  if [ "$gvmVersionStatus" == "0" ]; then
    isgvmInstalled=true
  else
    isgvmInstalled=false
  fi
}

checkIsDockerInstalled() {
  if [ -x "$(command -v docker)" ]; then
    isDockerInstalled=true
  else
    isDockerInstalled=false
  fi
}

checkIsDockerRunning() {
  curl -s --unix-socket /var/run/docker.sock http://ping &> /dev/null
  dockerStatus=$?
  if [ ! "$dockerStatus" == "7" ]; then
    isDockerRunning=true
  else
    isDockerRunning=false
  fi
}

checkIsKubectlInstalled() {
  if [ -x "$(command -v kubectl)" ]; then
    isKubectlInstalled=true
  else
    isKubectlInstalled=false
  fi
}

checkIsKindInstalled(){
  if [ -x "$(command -v kind)" ]; then
    isKindInstalled=true
  else
    isKindInstalled=false
  fi
}

checkIsClusterRunning() {
  kubectl cluster-info &> /dev/null
  clusterStatus=$?
  if [ ! "$clusterStatus" == "0" ]; then
    isClusterRunning=false
  else
    isClusterRunning=true
  fi
}

outputJson() {
  echo -e "{\"package\":{\"Homebrew\":\"$isHomebrewInstalled\", \"GVM\":\"$isgvmInstalled\", \
\"Mercurial\":\"$isMercurialInstalled\", \"Xcode Command Line Tools\":\"$isXcodeCommandLineToolsInstalled\", \
\"Docker\":\"$isDockerInstalled\", \"Kubectl\":\"$isKubectlInstalled\", \"Kind\":\"$isKindInstalled\"}, \
\"state\":{\"Docker\":\"$isDockerRunning\", \"Cluster\":\"$isClusterRunning\"}}"
}

main() {
  checkIfXcodeCommandLineToolsInstalled
  checkIsHomebrewInstalled
  checkIfMercurialInstalled
  checkIsGVMInstalled
  checkIsDockerInstalled
  checkIsDockerRunning
  checkIsKubectlInstalled
  checkIsKindInstalled
  checkIsClusterRunning
  outputJson
}

main


