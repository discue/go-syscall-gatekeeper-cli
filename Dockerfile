FROM debian:stable-slim AS build

ARG APP_DEV_DEPENDENCIES="build-essential ca-certificates curl golang libseccomp-dev pkg-config"

# Install packages
RUN apt-get -qq update; \
    apt-get install -qqy --no-install-recommends \
    ${APP_DEV_DEPENDENCIES} \
    && rm -rf /var/lib/apt/lists/*

# Download latest release
COPY docker/download.sh /download.sh
RUN chmod +x /download.sh \
  && /download.sh \
  && chmod +x /cuandari 

FROM debian:stable-slim 

ARG APP_TARGET_NAME=cuandari
ARG APP_TARGET_FOLDER=/home/${APP_TARGET_NAME}
ARG APP_RUNTIME_DEPENDENCIES="libseccomp2"
ARG USER=${APP_TARGET_NAME}
ARG USER_GROUP=sandbox
ARG USER_ID=327

# Install packages
RUN apt-get -qq update; \
    apt-get install -qqy --no-install-recommends \
    ${APP_RUNTIME_DEPENDENCIES} \
    && rm -rf /var/lib/apt/lists/*

# Create a custom user
RUN groupadd -r ${USER_GROUP} && useradd --no-log-init -r -u ${USER_ID} -g ${USER_GROUP} ${USER} 

# Copy necessary files from the build image
# > Only the go folder will remain and the go folder will only contain the binary we built in earlier stage
COPY --from=build --chown=${USER}:${USER_GROUP} /${APP_TARGET_NAME} ${APP_TARGET_FOLDER}/${APP_TARGET_NAME}

# Make process manager executable
RUN chmod 550 ${APP_TARGET_FOLDER}/${APP_TARGET_NAME} 

# Set user and working directory
USER ${USER}:${USER_GROUP}
WORKDIR ${APP_TARGET_FOLDER}

# Run the binary with a process manager that forwards all signals
ENTRYPOINT ["./cuandari", "run", "--allow-file-system-read", "--allow-network-local-sockets", "--"]
CMD ["ls", "-la"]