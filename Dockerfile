# Build stage: this all gets discarded later
FROM golang:1.10-alpine3.7 AS build

# start by installing git, which is fundamental to installing go code
RUN apk add --no-cache git openssh-client
# now install glide, which both tendermint and ndev-developed software
# use for dependency management
RUN go get github.com/Masterminds/glide

# accept whatever key the server provides
ENV KH="/root/.ssh/known_hosts"
ENV GITHUB="github.com"
RUN mkdir -p $(dirname ${KH})
RUN touch ${KH} && ssh-keygen -R ${GITHUB}
RUN ssh-keyscan -t rsa ${GITHUB} >> ${KH}

# Copy the context into the container and build the ndau application
ENV NDAU=$GOPATH/src/github.com/oneiro-ndev/ndaunode
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


# Terminal build stage: should only contain the static executable
FROM alpine:3.7

# only copy the binary artifacts we produced earlier
COPY --from=build /bin/ndaunode /bin/

# ndaunode listens here for its requests
# tendermint connects to this port
EXPOSE 46658/TCP

ENTRYPOINT ["/bin/ndaunode"]
