# Bolt card service installation using Docker

### install Docker engine and Docker compose

- [Docker engine download &
   install](https://docs.docker.com/engine/install/)

### Set up the boltcard server
- Run `./docker_init.sh` to set up the initial data
- Put the `tls.cert` file and `admin.macaroon` files in the project root directory

### https setup

- set up the domain A record to point to the server
- set up the server hosting firewall to allow open access to https (port 443) only

### database setup

- copy the `.env.example` file to `.env` and change the database password

### service bring-up and running
```
$ sudo groupadd docker
$ sudo usermod -aG docker ${USER}
(log out & in again)
$ docker volume create caddy_data
// add -d option for detached mode
$ docker compose up
```

### run boltcard server with own reverse proxy
If you already have reverse proxy in your enviroment which controls/terminates TLS connections, Boltcard server wont be ready to use, because of existence of own reverse proxy (Caddy). Caddy wont be abble to obtain TLS certificate for your domain name. Run different docker-compose, that will start Boltcard server without Caddy and your reverse proxy will handle TLS.

```
// add -d option for detached mode
$ docker-compose up -f docker-compose-own-reverse-proxy.yml
```


### stop docker
```
$ docker compose down
```
To delete the database and reset the docker volume, run `docker compose down --volumes`
*NOTE:  caddy_data volume won't be removed even if you run `docker compose down --volumes` because it's an external volume. **Make sure to wipe your programmed cards before wiping the database***

### check container logs

- [Docker Logs](https://docs.docker.com/engine/reference/commandline/logs/)

```
$ docker logs [OPTIONS] CONTAINER
```

Run `$ docker ps` to list containers and get container names/ids

#### running internal API commands
-  `docker exec boltcard_main curl 'localhost:9001/createboltcard?card_name=card_5&enable=false&tx_max=1000&day_max=10000&uid_privacy=true&allow_neg_bal=true'`
-  `docker exec boltcard_main curl 'localhost:9001/updateboltcard?card_name=card_5&enable=true&tx_max=100&day_max=1000'`
-  `docker exec boltcard_main curl 'localhost:9001/wipeboltcard?card_name=card_5'`
