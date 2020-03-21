#!/bin/bash

installHomebrew() {
  if [ -x "$(command -v brew)" ]; then
    echo "Homebrew is already installed"
  else
    echo "Installing homebrew..."
    ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
    echo "Homebrew Installtion Complete"
  fi
}

installCurl() {
  if [ -x "$(command -v curl)" ]; then
    echo "Curl is already installed"
  else
    echo "Installing Curl..."
    brew install curl
    echo "Homebrew Installtion Complete"
  fi
}

installGVM() {
  # shellcheck source=src/lib.sh
  source "$HOME/.gvm/scripts/gvm"
  gvm version &> /dev/null
  gvmVersionStatus=$?
  if [ "$gvmVersionStatus" == "0" ]; then
    echo "gvm is already installed"
  else
    echo "Installing gvm..."
    $SHELL < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
    # shellcheck source=src/lib.sh
    source "$HOME/.gvm/scripts/gvm"
    if [ -n "$($SHELL -c 'echo $ZSH_VERSION')" ]; then
      shell_profile="zshrc"
    elif [ -n "$($SHELL -c 'echo $BASH_VERSION')" ]; then
      shell_profile="bashrc"
    fi
    touch "$HOME/.${shell_profile}"
    {
    echo '# GVM'
    echo "export GVM_ROOT=$HOME/.gvm"
    echo '. $GVM_ROOT/scripts/gvm-default'
    } >> "$HOME/.${shell_profile}"
    echo "gvm Installation Complete"
  fi
}

# todo change to version comparison
installGoWithGVM() {
  gvm use go1.13.5 &> /dev/null
  gvmUseStatus=$?
  if [ "$gvmUseStatus" == "0" ]; then
    echo "Use go version go1.13.5"
  else
    echo "Installing go1.13.5"
    gvm install go1.13.5
    gvm use go1.13.5
    echo "go1.13.5 Installation Complete, now using go1.13.5"
  fi
}

installDocker() {
  if [ -x "$(command -v docker)" ]; then
    echo "Docker is already installed"
  else
    echo "Installing Docker..."
    brew cask install docker
    echo "Docker Installation Complete"
  fi
}

runDocker() {
  curl -s --unix-socket /var/run/docker.sock http://ping &> /dev/null
  status=$?
  if [ "$status" == "7" ]; then
    echo "Docker is not running, open docker app..."
    open /Applications/Docker.app
  else
    echo "Docker is already up and running"
  fi
}

installKubectl() {
  if [ -x "$(command -v kubectl)" ]; then
    echo "Kubectl is already installed"
  else
    brew install kubectl
  fi
}

installKind() {
  if [ -x "$(command -v kind)" ]; then
    echo "Kind is already installed"
  else
    echo "Installing Kind..."
    brew install kind
    echo "Kind Installtion Complete"
    echo "Creating cluster..."
    kind create cluster
    echo "Cluster was Created"
  fi
}

runCluster() {
  kubectl cluster-info &> /dev/null
  clusterStatus=$?
  if [ ! "$clusterStatus" == "0" ]; then
    echo "Creating cluster..."
    kind create cluster
    echo "Cluster was Created"
  else
    echo "Cluster is already up and running"
  fi
}

main() {
  installHomebrew
  installCurl
  installGVM
  installGoWithGVM
  installDocker
  runDocker
  installKubectl
  installKind
  runCluster
}

main
