#!/usr/bin/php -q

//Проверка существования файлов в папке files
//отправление данных в файле на сервер
//удаление переданного файла

<?php 
//Получим дату и время
date_default_timezone_set('UTC');
$date = date(DATE_ATOM); //Текущее время (оно же время передачи)

//Найдем файлы в папке files
$dir = '/var/www/html/files';
foreach (glob($dir . '/*.rfid') as $fileName) {
    //Получим время создания файла из наименования
    //$dateRead = basename($fileName);
    $dateRead = explode('_', basename($fileName));
    $dateRead = date(DATE_ATOM, $dateRead[0]);


    //Откроем файл и передадим его содержимое на сервер
    $fileName = $dir . '/' . basename($fileName);

    echo 'file:' . $fileName . ' size:' . filesize($fileName) . PHP_EOL;
    $file = file_get_contents($fileName);
    if (filesize($fileName) > 0) {
        echo 'body:' . $file . PHP_EOL;

        //Передадим данные файла на сервер трека
        file_put_contents('send_file_trek.txt', file_get_contents('http://tester.spotrek.ru/projects/3/import.php?action=add&date=' . $date . '&str=' . $dateRead . '***' . $file));
        file_put_contents('send_file.txt', file_get_contents('https://infopuma.ru/piscanner/?controller=1&scanner=1&data=' . $file . '&date_read=' . $dateRead . ''));

        $result = file_get_contents($dir . '/../' . 'send_file.txt');
        echo 'result:' . $result . ' size:' . strlen($result) . PHP_EOL;
        unlink($dir . '/../' . 'send_file.txt');

        //Удалим файл если есть результат
        if ((strlen($result) == 76) || ($result == '200')) {
	    unlink($fileName);
	    echo '****deleted****' . PHP_EOL;
	}
    }
}




