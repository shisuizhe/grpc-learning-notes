#!/bin/bash

set -e

yum install -y gcc-c++ pcre pcre-devel openssl openssl-devel

cd /opt
wget http://nginx.org/download/nginx-1.19.2.tar.gz
tar -xzf nginx-1.19.2.tar.gz
cd nginx-1.19.2
./configure --prefix=/opt/nginx --with-http_ssl_module --with-http_v2_module --with-http_stub_status_module
make && make install

ln -s /opt/nginx/sbin/nginx /usr/local/bin/nginx

cd ..
rm -rf nginx-1.19.2
rm -rf nginx-1.19.2.tar.gz
