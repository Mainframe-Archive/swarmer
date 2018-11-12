FROM alpine:3.7

LABEL "org.mfhq.domain"="swarm"

RUN mkdir /app && mkdir /app/bin

# Set the working directory to /app
WORKDIR /app

# Install dependencies
RUN apk update && \
    apk upgrade && \
    apk add jq git alpine-sdk go linux-headers bash

WORKDIR /app/go-ethereum

# Copy script for starting swarm into the container
COPY start.sh /app

CMD ./start.sh .
