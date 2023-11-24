#!/bin/bash
for ((;;))
#в бесконечном цыкле
do
	#создадим уникальное имя файла
	suffix=$((1000 + $RANDOM % 8999))
	current_date_time=`date +"%s"` 
	file_name="/var/www/html/files/"$current_date_time"_"$suffix".rfid"

	#запишем пакет сокета в файл
	nc -l -p 8090 > $file_name

	#вызовем процедуру передачи файла
	#/usr/bin/php /var/www/html/toServer.php &
done

