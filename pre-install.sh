#!/bin/sh
addgroup --system discodns
adduser --system --no-create-home --ingroup discodns --disabled-password --disabled-login discodns

if [ ! -d /var/log/discodns ]; then
  mkdir -p /var/log/discodns
  chown discodns:discodns /var/log/discodns
  chmod 750 /var/log/discodns
fi
