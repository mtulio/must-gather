FROM registry.svc.ci.openshift.org/openshift/release:golang-1.12 AS builder
WORKDIR /go/src/github.com/openshift/must-gather
COPY . .
ENV GO_PACKAGE github.com/openshift/must-gather
RUN curl -Ls https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64 -o /usr/bin/jq \
    && strip /usr/bin/jq \
    && chmod u+x /usr/bin/jq

FROM registry.svc.ci.openshift.org/openshift/origin-v4.0:cli
COPY --from=builder /go/src/github.com/openshift/must-gather/collection-scripts/* /usr/bin/
COPY --from=builder /usr/bin/jq /usr/bin/jq

