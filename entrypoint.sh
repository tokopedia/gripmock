#!/bin/sh

if [ ! -d /etc/nginx/ssl ]
then
  mkdir /etc/nginx/ssl
  CWD=`pwd`
  cd /etc/nginx/ssl
  openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes \
              -subj "/C=US/ST=NRW/L=Earth/O=CompanyName/OU=IT/CN=localhost/emailAddress=email@localhost"
  cd $CWD
fi

mkdir -p /run/nginx/ && nginx

gripmock $1
