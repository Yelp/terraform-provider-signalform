FROM ubuntu:trusty

MAINTAINER team-metrics <team-metrics@yelp.com>

ENV GO_VERSION=1.9
ENV TF_VERSION=0.11.7

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -yq \
    wget \
    git \
    build-essential \
    ruby1.9.1 rubygems1.9.1 \
    libopenssl-ruby1.9.1 \
    ruby1.9.1-dev \
    rpm \
    --no-install-recommends

RUN wget --no-check-certificate https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz && tar xzf go${GO_VERSION}.linux-amd64.tar.gz && mv go /usr/local
ENV PATH /usr/bin:/bin:/usr/sbin:/sbin:/usr/local/bin:/usr/local/sbin:/usr/local/go/bin:/go/bin
ENV GOPATH /go
RUN mkdir /go
ENV RUBYOPT="-KU -E utf-8:utf-8"
RUN gem install json -v 1.8.3 --no-rdoc --no-ri
RUN gem install ffi --no-rdoc --no-ri -v 1.9.18
RUN gem install fpm --no-rdoc --no-ri -v 1.4.0

RUN git clone https://github.com/hashicorp/terraform.git /go/src/github.com/hashicorp/terraform && \
    cd /go/src/github.com/hashicorp/terraform && \
    git checkout v${TF_VERSION}
