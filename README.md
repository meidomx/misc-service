# misc-service
Misc services

## A. Services included

* [x] LDAP Server with Postgresql backend
  * [x] Add/Delete/Modify/Search/Bind/Unbind
  * [x] Initialize LDAP
  * [x] TLS/StartTLS
  * [ ] RootDSE/Modify password
* [x] Full text search service
  * [x] Insert or update/Delete/Simple search

## B. API

### B.1 LDAP: standard LDAP client

### B.2 Full text search service - `http api`
  * Insert or update: `PUT /full_text/document/:doc_id`
  * Delete: `DELETE /full_text/document/:doc_id`
  * Search: `GET /full_text/documents?q=keyword`
  * Document definition: 
```go
package full_text
type CommonDocument struct {
  Id   string `json:"doc_id" form:"doc_id"`
  Body string `json:"body" form:"body"`
}
```

## C. Dependency services

* [X] Postgresql

## D. Utility commands

### 1. Generate certificate using OpenSSL

```shell
openssl genrsa -des3 -out example.ldap.enc.key 2048
openssl rsa -in example.ldap.enc.key -out example.ldap.key
openssl req -new -key example.ldap.key -out example.ldap.csr
openssl x509 -req -days 36500 -in example.ldap.csr -signkey example.ldap.key -out example.ldap.crt
```

## E. Dependencies

* [x] github.com/BurntSushi/toml - toml configuration
* [x] github.com/jackc/pgx/v4 - postgresql client
* [x] github.com/jimlambrt/gldap - ldap client
* [x] github.com/gin-gonic/gin - http service
* [x] github.com/spf13/afero - file system
* [x] github.com/blevesearch/bleve - full text search
