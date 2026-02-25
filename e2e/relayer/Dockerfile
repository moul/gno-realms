FROM node:24.11.0-alpine3.22

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"

WORKDIR /usr/src/app

RUN apk add --no-cache \
    curl \
    bash \
    gnome-keyring \
    dbus \
    libcap \
    sudo

# Copy the relayer source from local checkout
COPY . .
RUN rm -rf node_modules

RUN --mount=type=cache,target=/root/.npm \
    --mount=type=cache,id=pnpm,target=/pnpm/store \
    npm install -g pnpm typescript && \
    pnpm install && \
    pnpm build

RUN cp -r docker/relayer/bin/* /bin/

ENV PATH="/bin:/usr/src/app/scripts:$PATH"

ENTRYPOINT ["/bin/with_keyring"]
CMD ["ibc-v2-ts-relayer", "relay"]
