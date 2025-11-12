module github.com/mickamy/go-keyset/examples

go 1.24.0

replace github.com/mickamy/go-keyset => ..

replace github.com/mickamy/go-keyset/kgorm => ../kgorm

replace github.com/mickamy/go-keyset/ksql => ../ksql

require (
	github.com/mickamy/go-keyset v0.0.0
	github.com/mickamy/go-keyset/kgorm v0.0.0
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.31.1
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.6 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/text v0.31.0 // indirect
)
