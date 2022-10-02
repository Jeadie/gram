#!/bin/bash


## Get GOOS and GOARCH
case $(uname -m) in
    i386)   goarch="386" ;;
    i686)   goarch="386" ;;
    x86_64) goarch="amd64" ;;
    *) echo "Unknown GOARCH ${uname -m}"; exit 1 ;;
esac

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    goos="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    goos="darwin"
elif [[ "$OSTYPE" == "cygwin" ]]; then
    goos="windows"
elif [[ "$OSTYPE" == "msys" ]]; then
    goos="windows"
elif [[ "$OSTYPE" == "win32" ]]; then
    goos="windows"
elif [[ "$OSTYPE" == "android"* ]]; then
    goos="android"
elif [[ "$OSTYPE" == "freebsd"* ]]; then
    goos="freebsd"
elif [[ "$OSTYPE" == "solaris"* ]]; then
    goos="solaris"
elif [[ "$OSTYPE" == "netbsd" ]]; then
    goos="netbsd"
elif [[ "$OSTYPE" == "FreeBSD" ]]; then
    goos="freebsd"
elif [[ "$OSTYPE" == "openbsd"* ]]; then
    goos="openbsd"
elif [[ "$OSTYPE" == "darwin9" ]]; then
    goos="ios"
else 
    echo "Could not determine GOOS for OSTYPE=$OSTYPE"
    exit 1 
fi


# Download and extract binary
tar_gz_url=$(curl -s https://api.github.com/repos/jeadie/gram/releases/latest | jq -r ".assets[] | select(.name | test(\"gram-${goos}-${goarch}.tar.gz\")) | .browser_download_url" | grep -v "md5")
wget -q $tar_gz_url -O gram.tar.gz > /dev/null
tar -xf gram.tar.gz 
rm gram.tar.gz

# Set gram command 
chmod +x gram

# Set gram binary to path and alias
alias gram="$(pwd)/gram"
echo "export PATH=$PATH:$(pwd)/gram" > ~/.bashrc
echo "export PATH=$PATH:$(pwd)/gram" > ~/.zshrc


