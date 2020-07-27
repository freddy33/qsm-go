module github.com/freddy33/qsm-go/client

require (
	github.com/freddy33/qsm-go/utils v0.0.0-20200626135554-7ccf7c6a8f91
	github.com/freddy33/qsm-go/model v0.0.0-latest
	github.com/stretchr/testify v1.3.0
	github.com/golang/protobuf v1.4.2
)

replace github.com/freddy33/qsm-go/utils => ../utils
replace github.com/freddy33/qsm-go/model => ../model

go 1.13
