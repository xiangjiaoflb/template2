#!/bin/bash

#获取当前目录
shpath=`pwd`
runsh=run.sh #创建的运行脚本
killsh=kill.sh #kill的脚本
serverport=8000 #程序使用的端口
programname=template #程序名

file=/etc/cron.d/${programname}

echo "#!/bin/bash

#判断程序是否在运行
pid=\`ps -A|grep ${programname}\`
if [ -z \"\$pid\" ]; then
    ulimit -n 65535 
    #到程序所在的目录
    cd ${shpath}
    ./${programname} -port ${serverport} > ${programname}.log 2>&1 &
    echo \$! > ./${programname}.pid

    #启动时间
    date >> ${programname}start.log
fi" > $runsh

chmod +x $runsh

echo "#!/bin/bash

pid=\$(cat ./${programname}.pid)
kill -3 \${pid}
" > ${killsh}

chmod +x ${killsh}


#写定时任务
  echo "*/1 * * * * root ${shpath}/${runsh}
*/1 * * * * root sleep 5; ${shpath}/${runsh}
*/1 * * * * root sleep 10; ${shpath}/${runsh}
*/1 * * * * root sleep 15; ${shpath}/${runsh}
*/1 * * * * root sleep 20; ${shpath}/${runsh}
*/1 * * * * root sleep 25; ${shpath}/${runsh}
*/1 * * * * root sleep 30; ${shpath}/${runsh}
*/1 * * * * root sleep 35; ${shpath}/${runsh}
*/1 * * * * root sleep 40; ${shpath}/${runsh}
*/1 * * * * root sleep 45; ${shpath}/${runsh}
*/1 * * * * root sleep 50; ${shpath}/${runsh}
*/1 * * * * root sleep 55; ${shpath}/${runsh}" > $file