# misc-service
Misc services

## A. Services included

* [X] LDAP Server with Postgresql backend
  * [X] Add/Delete/Modify/Search/Bind/Unbind
  * [X] Initialize LDAP
  * [X] TLS

## B. Dependent services

* [X] Postgresql

## C. Utility commands

### 1. Generate certificate using OpenSSL

```shell
openssl genrsa -des3 -out example.ldap.enc.key 2048
openssl rsa -in example.ldap.enc.key -out example.ldap.key
openssl req -new -key example.ldap.key -out example.ldap.csr
openssl x509 -req -days 36500 -in example.ldap.csr -signkey example.ldap.key -out example.ldap.crt
```
