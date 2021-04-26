# swagger-preview
`swagger-preview` can preview swagger.

## Features
- preview swagger
- live reload

## Instllation
This is only can install with Go 1.16~ 

```sh
$ go install github.com/skanehira/swagger-preview/cmd/spr@latest
```

## Usage

```
$ spr swagger.yaml
2021/02/02 21:51:46 start server: 9999
2021/02/02 21:51:46 watching swagger.yaml

$ PORT=8080 spr api/swagger.yaml
2021/02/02 21:51:46 start server: 8080
2021/02/02 21:51:46 watching api/swagger.yaml
```

## Author
skanehira

## Thanks
- https://github.com/swagger-api/swagger-ui
