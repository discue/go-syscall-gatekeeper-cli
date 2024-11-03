FROM ubuntu AS build

ARG APP_BUILD_FOLDER=/app
ARG APP_MODULE_NAME=go-syscall-gatekeeper
ARG APP_TARGET_FOLDER=/home/${APP_MODULE_NAME}
ARG APP_TARGET_NAME=app
ARG USER=${APP_MODULE_NAME}
ARG USER_GROUP=gatekeeper
ARG USER_ID=1327
ARG TINI_VERSION=v0.19.0

ENV GOPATH=${APP_BUILD_FOLDER}/go

RUN mkdir -p ${APP_BUILD_FOLDER}

# Install packages
RUN apt-get -qq update; \
    apt-get install -qqy --no-install-recommends \
    build-essential ca-certificates golang libseccomp-dev pkg-config strace \
    && rm -rf /var/lib/apt/lists/*

# Install process manager
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini ${APP_BUILD_FOLDER}/tini

# Copy whole workspace
COPY . ${APP_BUILD_FOLDER}/app

# ENV GATEKEEPER_SYSCALLS_ALLOW_LIST=open

# Build the application and remove the source code
RUN cd ${APP_BUILD_FOLDER}/app \ 
    && go mod vendor \
    && ls -l \
    && go generate ./... \
    && cat app/buildtime-config/syscalls.go \
    && go install \
    && mv ${GOPATH}/bin/${APP_MODULE_NAME} ${GOPATH}/bin/${APP_TARGET_NAME} \
    && rm -rf ${APP_BUILD_FOLDER}/app

FROM ubuntu AS final

ARG APP_BUILD_FOLDER=/app
ARG APP_MODULE_NAME=gatekeeper
ARG APP_TARGET_FOLDER=/home/${APP_MODULE_NAME}
ARG APP_TARGET_NAME=app
ARG PORT=8080
ARG USER=${APP_MODULE_NAME}
ARG USER_GROUP=gatekeeper
ARG USER_ID=1327

ENV PORT=${PORT}
EXPOSE ${PORT}

# Install packages
RUN apt-get -qq update; \
    apt-get install -qqy --no-install-recommends \
    curl libseccomp-dev ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js 22 using curl and the NodeSource repository
RUN curl -fsSL https://deb.nodesource.com/setup_23.x | bash - && \
    apt-get install -y nodejs

# Create a custom user
RUN groupadd -r ${USER_GROUP} && useradd --no-log-init -r -u ${USER_ID} -g ${USER_GROUP} ${USER} 

# Copy necessary files from the build image
# > Only tiny and go folder will remain and go folder will only contain the binary we built in earlier stage
COPY --from=build --chown=${USER}:${USER_GROUP} ${APP_BUILD_FOLDER} ${APP_TARGET_FOLDER}
COPY server.js ${APP_TARGET_FOLDER}/server.js

# Make process manager executable
RUN chmod 740 ${APP_TARGET_FOLDER}/tini \
    && chmod 740 ${APP_TARGET_FOLDER}/go/bin/app 

# Set user and working directory
USER ${USER}:${USER_GROUP}
WORKDIR ${APP_TARGET_FOLDER}

# ENTRYPOINT ["./go/bin/app"]
ENTRYPOINT ["./tini", "--"]
CMD ["./go/bin/app", "/bin/node", "./server.js"]