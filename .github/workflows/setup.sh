#!/bin/bash
set -e

install_java() {
  echo "Installing SDKMAN..."
  curl -s "https://get.sdkman.io" | bash
  source "$HOME/.sdkman/bin/sdkman-init.sh"
  echo "sdkman_auto_answer=true" >> ~/.sdkman/etc/config

  echo "Installing Java versions..."
  sdk install java 11.0.24-zulu || true
  sdk install java 17.0.12-zulu || true

  sdk default java 11.0.24-zulu
  sdk use java 11.0.24-zulu

  echo "JAVA11_HOME=$JAVA_HOME_11_X64" >> $GITHUB_ENV
  echo "JAVA17_HOME=$JAVA_HOME_17_X64" >> $GITHUB_ENV
  echo "JAVA_HOME=$JAVA_HOME_11_X64" >> $GITHUB_ENV
  echo "PATH=$PATH" >> $GITHUB_ENV
}

install_ccm() {
  echo "Creating virtual environment..."
  VENV_DIR="$HOME/venv"
  python3 -m venv $VENV_DIR
  source $VENV_DIR/bin/activate
  pip install --upgrade pip setuptools

  echo "Installing CCM..."
  pip install "git+https://github.com/riptano/ccm.git@${CCM_VERSION}" || true
}

install_java
install_ccm
