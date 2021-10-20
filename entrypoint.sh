#!/bin/sh

if [[ -z $CONSUL_ADDRESS ]]; then
        echo "exec meshdemo..."
        exec /opt/consuldemo/bin/meshdemo
else
        echo "exec consuldemo..."
        exec /opt/consuldemo/bin/consuldemo
fi