#! /bin/bash

if [ ! -f /var/run/monitor.pid ]; then
    echo "Monitor not running"
    /home/ec2-user/prom-website-exporter/monitoring/run.sh
fi