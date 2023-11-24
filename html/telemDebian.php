#!/usr/bin/php -q

//Телематика контроллера
//Данные отравляются один раз в минуту не дожидаясь подтверждения

<?php 
date_default_timezone_set('UTC');
$date = date(DATE_ATOM);
file_put_contents('output_trek.txt', file_get_contents('http://tester.spotrek.ru/projects/3/import.php?action=add&date=' . $date . '&str=raspberry_online' . '***' . $date));
file_put_contents('output_puma.txt', file_get_contents('https://infopuma.ru/piscanner/?controller=1&status=1&date=' . $date . ''));


