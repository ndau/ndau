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

# Copy the context into the container and build the ndau application
ENV NDAU=$GOPATH/src/github.com/oneiro-ndev/ndau
COPY ./pkg ${NDAU}/pkg
COPY ./cmd ${NDAU}/cmd
COPY ./glide.* ${NDAU}/
WORKDIR ${NDAU}

# note the concatenated command here. This is a special case:
# we want to be absolutely sure that "glide install" doesn't use a cached version
# when we build the app, so we &&-concatenate the commands.
RUN glide install && \
    CGO_ENABLED=0 GOOS=linux go install -a -ldflags '-extldflags "-static"' ./cmd/ndaunode
# we need to copy the produced executable to a known path because the
# $GOPATH environment variable doesn't persist into the run container
RUN cp ${GOPATH}/bin/ndaunode /bin/

# Remove the secret just in case its directory gets copied below.
RUN rm -vf /root/.ssh/id_rsa

# Terminal build stage: should only contain the static executable
FROM alpine:3.7

# only copy the binary artifacts we produced earlier
COPY --from=build /bin/ndaunode /bin/

# ndaunode listens here for its requests
# tendermint connects to this port
EXPOSE 26658/TCP

ENTRYPOINT ["/bin/ndaunode"]
