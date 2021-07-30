#! /bin/bash

if [ ! -f /var/run/prom-website-exporter.pid ]; then
    echo "Monitor not running"
    /home/ec2-user/prom-website-exporter/run.sh
fi