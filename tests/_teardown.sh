#!/usr/bin/env bash
pid=$(ps aux | grep [m]oto_server | awk '{print $2}')
kill ${pid}
