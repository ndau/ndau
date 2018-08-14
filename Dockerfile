# Build stage: this all gets discarded later
FROM golang:1.10-alpine3.7 AS build

# start by installing fundamental utilities
# openssh-client is required for git
# git is required for glide
# bash is required for check-github-fingerprint.sh (array syntax)
RUN apk add --no-cache git openssh-client bash

# Note that in this file we're using a large number of independent RUN
# instructions instead of concatenating everything into a compressed
# &&-sequence. We expect to build this on Docker versions post-17.05,
# and take advantage of multi-stage builds, which means that the size
# penalties of this approach do not apply: everything in the build container
# is discarded. We do get to keep the caching and clarity benefits, though.

# Normally there is a very clear rule:
# DO NOT DO THIS.
# If we copy a keyfile into a container, its contents can be
# recovered later, even if we delete the file from the container.
#
# Multi-stage builds change that rule. We still want to be careful
# about how we handle the keyfile, but the general rule is now that
# non-terminal stages vanish irrecoverably from the final container.
# This means that it's not only easier to get compact containers, but
# it's no longer a huge security issue to pass in a keyfile to a non-
# terminal container.
ARG SSH_KEY_FILE="github_chaos_deploy"
COPY ${SSH_KEY_FILE} /root/.ssh/id_rsa
COPY ./bin/check-github-fingerprint.sh /root/
RUN /root/check-github-fingerprint.sh
RUN chmod 0600 /root/.ssh/*

# now install glide, which both tendermint and ndev-developed software
# use for dependency management
RUN go get github.com/Masterminds/glide

# set up working directory
ENV NDAU=$GOPATH/src/github.com/oneiro-ndev/ndau
WORKDIR ${NDAU}

# Copy the glide stuff into the container and fetch dependencies
# Doing this will cache the deps so long as docker detects glide.lock
# updates appropriately, but you'll still want to run production
# builds with --no-cache
COPY ./glide.* ${NDAU}/
RUN git config --global url.git@github.com:.insteadof https://github.com/ && \
    glide install

# Copy the source into the container and build
COPY ./pkg ${NDAU}/pkg
COPY ./cmd ${NDAU}/cmd
RUN CGO_ENABLED=0 GOOS=linux go install -a -ldflags '-extldflags "-static"' ./cmd/ndaunode

# we need to copy the produced executable to a known path because the
# $GOPATH environment variable doesn't persist into the run container
RUN cp ${GOPATH}/bin/ndaunode /bin/

# Remove the secret just in case its directory gets copied below.
RUN rm -vf /root/.ssh/id_rsa

# Terminal build stage: should only contain the static executable
FROM alpine:3.7
# Add the ca certs so that we can do outbound connections to https sites (honeycomb)
RUN apk add --no-cache  ca-certificates

# only copy the binary artifacts we produced earlier
COPY --from=build /bin/ndaunode /bin/

# ndaunode listens here for its requests
# tendermint connects to this port
EXPOSE 26658/TCP

ENTRYPOINT ["/bin/ndaunode"]
