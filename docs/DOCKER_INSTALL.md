# Bolt card service installation using Docker

### install Docker engine and Docker compose

[Docker engine download & install](https://docs.docker.com/engine/install/)
[Docker compose download & install](https://docs.docker.com/compose/install/)

### Set up the boltcard server
edit `.env` to set up the database connection
edit `settings.sql` to set up [bolt card system settings](SETTINGS.md)
edit `Caddyfile` to set the boltcard domain name

### https setup

set up the domain A record to point to the server

set up the server hosting firewall to allow open access to https (port 443) only


### service bring-up and running
```
$ docker volumes create caddy_data
// add -d for detached mode
$ docker-compose up -d
```

### stop docker
```
$ docker-compose down
```
To delete the database and reset the docker volume, run `docker-compose down --volumes`
*NOTE:  caddy_data volume won't be removed even if you run `docker-compose down --volumes` because it's an external volume.*  


#### running create bolt card command
-  `docker exec boltcard_main createboltcard/createboltcard`  to see options
-  `docker exec boltcard_main createboltcard/createboltcard -enable -allow_neg_bal -tx_max=1000 -day_max=10000 -name=card_1`  for example
-  this will give you a one-time link in text and QR code form