#!/bin/bash

isHomebrewInstalled() {
  if [ ! -x "$(command -v brew)" ]; then
    echo "Installing homebrew..."
    ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
    echo "Homebrew Installtion Complete"
  else
    echo "Homebrew is already installed"
  fi
}

isGoInstalled() {
  if [ ! -x "$(command -v go)" ]; then
    echo "Installing Go..."
    installGo
  else
    echo "Go is already installed"
  fi
}

installGo() {
  VERSION="1.13.5"
  PLATFORM="darwin-amd64"
  PACKAGE_NAME="go$VERSION.$PLATFORM.tar.gz"
  [ -z "$GOROOT" ] && GOROOT="$HOME/.go"
  [ -z "$GOPATH" ] && GOPATH="$HOME/go"

  if [ -n "`$SHELL -c 'echo $ZSH_VERSION'`" ]; then
    shell_profile="zshrc"
  elif [ -n "`$SHELL -c 'echo $BASH_VERSION'`" ]; then
    shell_profile="bashrc"
  fi

  if [ -d "$GOROOT" ]; then
    echo "The Go install directory ($GOROOT) already exists. Exiting."
    exit 1
  fi

  echo "Downloading $PACKAGE_NAME ..."

  if hash wget 2>/dev/null; then
    wget https://storage.googleapis.com/golang/$PACKAGE_NAME -O /tmp/go.tar.gz
  else
    curl -o /tmp/go.tar.gz https://storage.googleapis.com/golang/$PACKAGE_NAME
  fi

  if [ $? -ne 0 ]; then
    echo "Download go package failed! Exiting."
    exit 1
  fi

  echo "Extracting File..."
  mkdir -p "$GOROOT"
  tar -C "$GOROOT" --strip-components=1 -xzf /tmp/go.tar.gz
  touch "$HOME/.${shell_profile}"
  {
    echo '# GoLang'
    echo "export GOROOT=${GOROOT}"
    echo 'export PATH=$GOROOT/bin:$PATH'
    echo "export GOPATH=$GOPATH"
    echo 'export PATH=$GOPATH/bin:$PATH'
    } >> "$HOME/.${shell_profile}"

  mkdir -p $GOPATH/{src,pkg,bin}
  echo -e "\nGo $VERSION was installed into $GOROOT.\nMake sure to relogin into your shell or run:"
  echo -e "\n\tsource $HOME/.${shell_profile}\n\nto update your environment variables."
  rm -f /tmp/go.tar.gz
}

isDockerInstalled() {
  if [ ! -x "$(command -v docker)" ]; then
    echo "Installing Docker..."
    brew cask install docker
    echo "Docker Installation Complete"
  else
    echo "Docker is already installed"
  fi
}

isDockerRunning() {
  rep=$(curl -s --unix-socket /var/run/docker.sock http://ping > /dev/null)
  status=$?

  if [ "$status" == "7" ]; then
    echo "Docker is not running, open docker app..."
    open /Applications/Docker.app
  else
    echo "Docker is up and running"
  fi
}

isKindInstalled() {
  if [ ! -x "$(command -v kind)" ]; then
    echo "Installing Kind..."
    brew install kind
    echo "Kind Installtion Complete"
    echo "Creating cluster"
    kind create cluster
    echo "Cluster was Created"
  else
    echo "Kind is already installed"
  fi
}

isHomebrewInstalled
isGoInstalled
isDockerInstalled
isDockerRunning
isKindInstalled
