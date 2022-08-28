# Bolt card service installation

## hardware & o/s

1 GHz processor, 2 GB RAM, 10GB storage minimum  
Ubuntu 20.04 LTS server

## With docker & docker-compose
### 1. Download the boltcard repository

`$ git clone https://github.com/boltcard/boltcard`

### 2. Get a macaroon and tls.cert from the lightning node

Create a macaroon with limited permissions to the lightning node  
[lncli download & install](https://github.com/lightningnetwork/lnd/blob/master/docs/INSTALL.md)
```
$ lncli \                                                    
--rpcserver=lightning-node.io:10009 \
--macaroonpath=admin.macaroon \
--tlscertpath="tls.cert" \
bakemacaroon uri:/routerrpc.Router/SendPaymentV2 > SendPaymentV2.macaroon.hex

$ xxd -r -p SendPaymentV2.macaroon.hex SendPaymentV2.macaroon
```
Copy tls.cert and SendPaymentV2.macaroon to your boltcard directory

### 3. Configure and run

Edit the .env file to your preference and run

```
docker-compose up -d
```

This will spin up a *postgresql* container, and the *boltcard service* container available at port **9000**. For publishing with a domain name and https, you can use a reverse proxy like nginx, traefik or caddy.

You can monitor with ```docker logs container_name```.

## Without docker

### login

create and use a user named `ubuntu`

### install Go

[Go download & install](https://go.dev/doc/install)  
`$ go version` >= 1.18.3

### install Postgres

[Postgres download & install](https://www.postgresql.org/download/linux/ubuntu/)  
`$ psql --version` >= 12.11

### install Caddy

[Caddy download & install](https://caddyserver.com/docs/install)  
`$ caddy version` >= 2.5.2

### download the boltcard repository

`$ git clone https://github.com/boltcard/boltcard`

### get a macaroon and tls.cert from the lightning node

create a macaroon with limited permissions to the lightning node  
[lncli download & install](https://github.com/lightningnetwork/lnd/blob/master/docs/INSTALL.md)
```
$ lncli \                                                    
--rpcserver=lightning-node.io:10009 \
--macaroonpath=admin.macaroon \
--tlscertpath="tls.cert" \
bakemacaroon uri:/routerrpc.Router/SendPaymentV2 > SendPaymentV2.macaroon.hex

$ xxd -r -p SendPaymentV2.macaroon.hex SendPaymentV2.macaroon
```

### setup the boltcard server
edit `boltcard.service` in the section named `boltcard service settings`  
edit `Caddyfile` to set the boltcard domain name  

### database creation
edit `create_db.sql` to set the cardapp password  
`$ sudo -u postgres createuser -s ubuntu`  
`$ ./s_create_db`  

### boltcard service install
`$ sudo cp boltcard.service /etc/systemd/system/boltcard.service`  
`$ ./s_build`  
`$ sudo systemctl enable boltcard`  
`$ sudo systemctl status boltcard`

### https setup
set up the domain A record to point to the server  
set up the server hosting firewall to allow open access to https (port 443) only  

### caddy setup for https
`$ sudo cp Caddyfile /etc/caddy`  
`$ sudo systemctl stop caddy`  
`$ sudo systemctl start caddy`  
`$ sudo systemctl status caddy`  
you should see 'certificate obtained successfully' in the service log

### service bring-up and testing
#### service log
the service log should be monitored on a separate console while tests are run  
`$ journalctl -u boltcard.service -f`
#### local http
`$ curl http://127.0.0.1:9000/ln?1`  
this should respond with 'bad request' and show up in the service log  
#### remote https
navigate to the service URL from a browser, for example `https://card.yourdomain.com/ln?2`  
this should respond with 'bad request' and show up in the service log  
#### bolt card
[create a bolt card](CARD_ANDROID.md) with the URI pointing to this server  
use a PoS setup to read the bolt card, e.g. [Breez wallet](https://breez.technology/)   
monitor the service log to ensure decryption, authentication, payment rules and lightning payment work as expected  
#### production use
ensure that LOG_LEVEL is set to PRODUCTION  
ensure that all secrets are minimally available  
ensure that you have good operational security practices  
monitor the system for unusual activity  

# Further information and support

[bolt card FAQ](FAQ.md)  
[bolt card telegram group](https://t.me/bolt_card)
