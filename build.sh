#!/bin/bash
VERSION=1.3
GOOS=linux go build

mkdir -p build_dir
mkdir -p build_dir/usr/local/bin
mkdir -p build_dir/etc/init
mkdir -p build_dir/etc/default
cp botbrother build_dir/usr/local/bin
cp etc/init/botbrother.conf build_dir/etc/init/botbrother.conf
cp etc/default/botbrother build_dir/etc/default/botbrother
chmod 755 build_dir/usr/local/bin/botbrother

fpm \
-s dir \
-t deb  \
-v ${VERSION} \
--name botbrother \
--description "BotBrother - The AWS Slackbot" \
--prefix "/" \
--before-install pre-install.sh \
--after-install post-install.sh \
--deb-user root \
--deb-group root \
-C build_dir .

rm -rf build_dir
