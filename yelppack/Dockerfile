FROM docker-dev.yelpcorp.com/lucid_yelp
MAINTAINER Keith Mosher <kmosher@yelp.com>

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -yq \
    go \
    git \
    build-essential \
    ruby1.9.1 rubygems1.9.1 \
    libopenssl-ruby1.9.1 \
    ruby1.9.1-dev \
    --no-install-recommends

ENV GOPATH /go
ENV RUBYOPT="-KU -E utf-8:utf-8"
RUN gem install json -v 1.8.3 --no-rdoc --no-ri
RUN gem install fpm --no-rdoc --no-ri -v 1.4.0
RUN ln -s /var/lib/gems/1.9.1/bin/fpm /usr/local/bin/fpm

RUN git clone https://github.com/hashicorp/terraform.git /go/src/github.com/hashicorp/terraform && \
    cd /go/src/github.com/hashicorp/terraform && \
    git checkout 56926a18f126c02877c5c25438f5a7c1805bbc87
