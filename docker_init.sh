#!/bin/bash
echo Enter the domain name excluding the protocol
read domainname

echo Enter your LND node gRPC domain
read lnd_host

echo LND node gRPC port
read lnd_port
sed -i "1s/.*/https:\/\/$domainname/" Caddyfile_docker
sed -i "s/[(]'HOST_DOMAIN'[^)]*[)]/(\'HOST_DOMAIN\', \'$domainname\')/" sql/settings.sql
echo writing the domain name to $domainname ...

PASSWORD=$(date +%s|sha256sum|base64|head -c 32)
if [[ ! -e .env ]]; then
    cp .env.example .env
fi
sed -i "s/^DB_PASSWORD=.*$/DB_PASSWORD=$PASSWORD/g" .env
decrypt_key=$(hexdump -vn16 -e'4/4 "%08x" 1 "\n"' /dev/random)
echo $decrypt_key

sed -i "s/[(]'LOG_LEVEL'[^)]*[)]/(\'LOG_LEVEL\', \'DEBUG\')/" sql/settings.sql
sed -i "s/[(]'AES_DECRYPT_KEY'[^)]*[)]/(\'AES_DECRYPT_KEY\', \'$decrypt_key\')/" sql/settings.sql
sed -i "s/[(]'MIN_WITHDRAW_SATS'[^)]*[)]/(\'MIN_WITHDRAW_SATS\', \'1\')/" sql/settings.sql
sed -i "s/[(]'MAX_WITHDRAW_SATS'[^)]*[)]/(\'MAX_WITHDRAW_SATS\', \'1000000\')/" sql/settings.sql
sed -i "s/[(]'LN_HOST'[^)]*[)]/(\'LN_HOST\', \'$lnd_host\')/" sql/settings.sql
sed -i "s/[(]'LN_PORT'[^)]*[)]/(\'LN_PORT\', \'$lnd_port\')/" sql/settings.sql
sed -i "s/[(]'LN_TLS_FILE'[^)]*[)]/(\'LN_TLS_FILE\', \'\/boltcard\/tls.cert\')/" sql/settings.sql
sed -i "s/[(]'LN_MACAROON_FILE'[^)]*[)]/(\'LN_MACAROON_FILE\', \'\/boltcard\/admin.macaroon\')/" sql/settings.sql
sed -i "s/[(]'FEE_LIMIT_SAT'[^)]*[)]/(\'FEE_LIMIT_SAT\', \'10\')/" sql/settings.sql
sed -i "s/[(]'FEE_LIMIT_PERCENT'[^)]*[)]/(\'FEE_LIMIT_PERCENT\', \'0.5\')/" sql/settings.sql
sed -i "s/[(]'FUNCTION_LNURLW'[^)]*[)]/(\'FUNCTION_LNURLW\', \'ENABLE\')/" sql/settings.sql
sed -i "s/[(]'FUNCTION_LNURLP'[^)]*[)]/(\'FUNCTION_LNURLP\', \'DISABLE\')/" sql/settings.sql
sed -i "s/[(]'FUNCTION_EMAIL'[^)]*[)]/(\'FUNCTION_EMAIL\', \'DISABLE\')/" sql/settings.sql
sed -i "s/[(]'LN_INVOICE_EXPIRY_SEC'[^)]*[)]/(\'LN_INVOICE_EXPIRY_SEC\', \'3600\')/" sql/settings.sql
