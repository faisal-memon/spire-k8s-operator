# Build stage
ARG goversion
FROM golang:${goversion}-alpine as builder
RUN apk add build-base git
ADD . /spire-k8s-operator
WORKDIR /spire-k8s-operator
RUN make build

# spire-k8s-operator
FROM alpine
ENV OPERATOR=/usr/local/bin/spire-k8s-operator \
    USER_UID=1001 \
    USER_NAME=spire-k8s-operator

COPY --from=builder /spire-k8s-operator/build/bin/spire-k8s-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
