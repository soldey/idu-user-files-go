# IDU User files API on GO
## Installation

```
go mod download
```
\+ .env file

## Boot
### win
```
set APP_ENV=development& go run main.go
```
### unix
```
APP_ENV=development go run main.go
```

## Stress tests
### win + grafana/k6
```
cat script.js | docker run --rm -i grafana/k6 run -
```