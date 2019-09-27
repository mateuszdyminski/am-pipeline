#!/bin/sh

sed -i "s/API_SERVER/$API_SERVER/g" /web/statics/app/services/users-srv.js

/bin/parent caddy --conf=/etc/Caddyfile --log stdout --agree=false