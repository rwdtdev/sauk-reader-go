#!/bin/bash
#Проверим идет считывание из сеокета
#если нет, то запустим
line=$(ps aux | grep socketRead.sh | grep -v grep)
if [ -z "$line" ]
then
    cd /var/www/bin && sudo bash socketRead.sh &
else
    echo "Soket reader already running."
fi
