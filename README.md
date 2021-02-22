# spr
`spr` can preview swagger.

## Features
- preview swagger
- live reload

## Instllation

```sh
$ git clone https://github.com/skanehira/spr
$ cd spr && go install ./...
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
https://github.com/swagger-api/swagger-ui
