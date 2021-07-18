#! /bin/bash

/usr/local/sbin/daemonize -a \
    -c /home/ec2-user/prom-website-exporter \
    -e /var/log/monitor.log \
    -o /var/log/monitor.log \
    -p /var/run/monitor.pid \
    -l /var/run/monitor.pid \
    /home/ec2-user/prom-website-exporter/monitoring

