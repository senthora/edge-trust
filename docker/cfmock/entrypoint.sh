#!/bin/sh

mockcfctl set

exec nginx -g 'daemon off;'
