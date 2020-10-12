module github.com/freddy33/qsm-go/backend

require (
	github.com/c2h5oh/datasize v0.0.0-20200825124411-48ed595a09d2
	github.com/freddy33/qsm-go/m3util v0.0.0-latest
	github.com/freddy33/qsm-go/model v0.0.0-20200626140801-6f9a15bc7381
	github.com/freddy33/urlquery v1.2.4
	github.com/golang/protobuf v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/joho/godotenv v1.3.0
	github.com/lib/pq v1.7.1
	github.com/rs/cors v1.7.0
	github.com/stretchr/testify v1.3.0
)

replace github.com/freddy33/qsm-go/m3util => ../m3util

replace github.com/freddy33/qsm-go/model => ../model

go 1.14
