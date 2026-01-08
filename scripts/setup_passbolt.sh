#!/bin/bash

docker-compose -f docker-compose-ce.yaml up -d
docker compose -f docker-compose-ce.yaml \
  exec passbolt su -m -c "/usr/share/php/passbolt/bin/cake \
    passbolt register_user \
      -u admin@example.com \
      -f Admin \
      -l User \
      -r admin" -s /bin/sh www-data
