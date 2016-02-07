#!/bin/sh
addgroup --system botbrother
adduser --system --no-create-home --ingroup botbrother --disabled-password --disabled-login botbrother

if [ ! -d /var/log/botbrother ]; then
  mkdir -p /var/log/botbrother
  chown botbrother:botbrother /var/log/botbrother
  chmod 750 /var/log/botbrother
fi
