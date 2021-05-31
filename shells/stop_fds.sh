#!/bin/bash

#结束
launchctl stop nginx
sudo nginx -s stop
launchctl stop fds

# sleep 1s 表示延迟一秒  
# sleep 1m 表示延迟一分钟  
# sleep 1h 表示延迟一小时  
# sleep 1d 表示延迟一天   

sleep 5s

#检查是否存在进程
ps aux | grep nginx
ps aux | grep fds
