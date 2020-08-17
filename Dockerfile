FROM golang:1.14-stretch AS build
ARG ARCH=amd64
ARG GOARM

ADD . $GOPATH/src/github.com/iketheadore/skycoin_node_alert_bot/

ENV GOARCH="$ARCH" \
		GOARM="$GOARM" \
		CGO_ENABLED="0" \
		GOOS="linux"

RUN cd $GOPATH/src/github.com/iketheadore/skycoin_node_alert_bot && \
		go install ./...

RUN apt-get update && \
    apt-get install -y ca-certificates

RUN /bin/bash -c 'mkdir -p /tmp/files/{usr/bin,/etc/ssl}'
RUN cp -r /go/bin/* /tmp/files/usr/bin/
RUN cp -r /etc/ssl/certs /tmp/files/etc/ssl/certs


FROM busybox

COPY --from=build /tmp/files /

ENTRYPOINT ["skycoin_node_alert_bot"]
