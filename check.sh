#!/usr/bin/env bash

checkIsHomebrewInstalled() {
  if [ -x "$(command -v brew)" ]; then
    isHomebrewInstalled=true
  else
    isHomebrewInstalled=false
  fi
}

checkIsGVMInstalled() {
  # shellcheck source=src/lib.sh
  source "$HOME/.gvm/scripts/gvm"
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
\"Docker\":\"$isDockerInstalled\", \"Kubectl\":\"$isKubectlInstalled\", \"Kind\":\"$isKindInstalled\"}, \
\"state\":{\"Docker\":\"$isDockerRunning\", \"Cluster\":\"$isClusterRunning\"}}"
}

main() {
  checkIsHomebrewInstalled
  checkIsGVMInstalled
  checkIsDockerInstalled
  checkIsDockerRunning
  checkIsKubectlInstalled
  checkIsKindInstalled
  checkIsClusterRunning
  outputJson
}

main


