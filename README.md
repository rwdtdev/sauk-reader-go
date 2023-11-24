# Установка считывателя в ТЧ-10

Для реализации процесса прослеживаемости комплектности вьезжаемого поезда с составными активами.

## На чем написан

PHP 7.3


BASH-скрипты

## Установка на контроллер

1. Установить Apache вэб-сервер.


Выполнить команды последовательно:
```sh
sudo apt update
```
```sh
sudo apt upgrade
```
```sh
sudo apt install apache2
```
```sh
sudo systemctl enable apache2
```
2. Установить PHP.


Выполнить команды последовательно:
```sh
sudo apt install php
```

```sh
sudo apt install libapache2-mod-php
```

3. Скопировать содержимое папки `html` в папку `/var/www/html`


4. Настроить сеть, изменив содержимое файла `/etc/network/interfaces`

```sh
auto lo
iface lo inet loopback
auto eth0
iface eth0 inet static
address 192.168.1.39
netmask 255.255.255.0
auto eth1
iface eth1 inet static
address 192.168.88.2
netmask 255.255.255.0
gateway 192.168.88.1
```

5. Добавить задания в крон суперпользователя, командой: `sudo su && crontab -e`

```sh
*/1 * * * * cd /var/www/html/ && /usr/bin/php telemDebian.php > /dev/null 2>&1
*/1 * * * * cd /var/www/bin/ && sudo bash starter.sh > /dev/null 2>&1
*/1 * * * * cd /var/www/html/ && sudo /usr/bin/php toServer.php > log.txt
```
