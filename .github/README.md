# Half-Life TV Manager <img align="right" src="./HLTV-Manager.png" alt="HLTV Launcher" width="210" height="200"/>

Сервис запускается в docker контейнере.

Сервис запускает hltv сервера в контейнерах.

Сервис позволяет скачивать демки, также автоматически контролирует и удаляет старые демки.

## Описание

Half-Life TV Manager - Позволяет запускать неограниченное количество hltv серверов. Предоставляет сайт для скачивания демок hltv.

## Характеристики

- Сервис устанавливается и запускается с помощью docker.
- Все настраивается через yaml конфигурации. (Временно)
- Поддержка запуска несколько HLTV серверов.
- Сайт для скачивания демок.
- Автоматические удаление демок.
- Оффлайн демки. (Временно)

## Установка

<details>
  <summary>Ubuntu</summary>

- Скачиваем docker-compose 

    `sudo apt update && sudo apt upgrade`

    `sudo apt install docker-compose`

- Скачиваем контейнер HLTV

    `sudo docker pull ghcr.io/wesstorn/hltv-files:v1.3`

- Загружаем Hltv-Manager и заходим в него

    `git clone --branch self-hosted https://github.com/WessTorn/HLTV-Manager.git`

    `cd HLTV-Manager`

    Настройка docker-compose конфига

    `nano .env`

    Настраиваем наши HLTV

    `nano hltv-runners.yaml`

- Запускаем сервис

    `sudo docker-compose up -d`

- Docker команды

    `sudo docker-compose up -d` - Запустить в фоне

    `sudo docker-compose up` - Запустить в текущей сессии (показывает логи)

    `sudo docker-compose down` - Остановить сервис

    `sudo docker-compose logs` - Посмотреть логи
</details>


## В будущем

- Конфигурация, настройка, запуск HLTV через сайт.
- Live терминалы HLTV
- Поддержка hltv с прямыми трансляциями.
- Amxx api часть для удаленной работы с hltv сервером.
