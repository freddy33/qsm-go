module github.com/freddy33/qsm-go/client

require (
	github.com/freddy33/qsm-go/m3util v0.0.0-latest
	github.com/freddy33/qsm-go/model v0.0.0-latest
	github.com/golang/protobuf v1.4.2
	github.com/joho/godotenv v1.3.0
	github.com/stretchr/testify v1.3.0
)

replace github.com/freddy33/qsm-go/m3util => ../m3util

replace github.com/freddy33/qsm-go/model => ../model

go 1.13
