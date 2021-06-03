
#执行
cd path_go_project
sudo sh ./shells/home_fds.sh
sudo sh ./shells/stop_fds.sh


#结束
launchctl stop nginx
sudo nginx -s stop
launchctl stop fds

#检查是否存在进程
ps aux | grep nginx 
ps aux | grep fds 


FastDFS v6.07


