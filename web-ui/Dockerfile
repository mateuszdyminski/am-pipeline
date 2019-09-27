# Arguments
ARG CADDY_VERSION

FROM abiosoft/caddy:${CADDY_VERSION}

RUN mkdir -p /web

ENV API_SERVER=""

# Copy from repo
RUN mkdir -p /web/statics
COPY statics /web/statics/
COPY run.sh /web/run.sh
COPY Caddyfile /etc/Caddyfile

EXPOSE 2015

ENTRYPOINT [ "/web/run.sh" ]
