#
# Main Dockerfile 
#
# Should create images < 10MB :)
#
FROM alpine:3.5
LABEL maintainer "james.tarball@newtonsystems.co.uk"

# Add Label Badges to Dockerfile powered by microbadger
ARG VCS_REF
LABEL org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="e.g. https://github.com/microscaling/microscaling"

ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH

RUN mkdir -p $GOPATH/bin
COPY app/main $GOPATH/bin

CMD ["/go/bin/main"]
