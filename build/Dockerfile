FROM ubuntu:xenial
MAINTAINER Yelp <darwin@yelp.dog>

ENV GO_VERSION=1.11
ENV TF_VERSION=0.12.1

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install --no-install-recommends -yq \
    build-essential \
    git \
    rpm \
    ruby \
    ruby-dev \
    wget

RUN wget --no-check-certificate https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz && tar xzf go${GO_VERSION}.linux-amd64.tar.gz && mv go /usr/local
ENV PATH /usr/bin:/bin:/usr/sbin:/sbin:/usr/local/bin:/usr/local/sbin:/usr/local/go/bin:/go/bin
ENV GOPATH /go
RUN mkdir /go
ENV RUBYOPT="-KU -E utf-8:utf-8"
RUN gem install json --no-rdoc --no-ri -v 2.2.0
RUN gem install ffi --no-rdoc --no-ri -v 1.10.0
RUN gem install fpm --no-rdoc --no-ri -v 1.11.0

RUN git clone https://github.com/hashicorp/terraform.git /go/src/github.com/hashicorp/terraform && \
    cd /go/src/github.com/hashicorp/terraform && \
    git checkout v${TF_VERSION}
