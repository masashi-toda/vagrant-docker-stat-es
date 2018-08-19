nohup vmstat -n -S M 1 | awk -vhostname="$(hostname)" '{ print hostname, strftime("%Y-%m-%d'T'%H:%M:%S+09:00"), $0 } { system(":") }' >> stat.log &
