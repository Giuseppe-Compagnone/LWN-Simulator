module github.com/Giuseppe-Compagnone/lwn-engine

go 1.26.5

require (
	github.com/Giuseppe-Compagnone/lwn-contracts v1.0.0
	github.com/dop251/goja v0.0.0-20261018113103-d42f16bb6161
	github.com/google/uuid v1.6.0
)

require (
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	golang.org/x/text v0.43.0 // indirect
)

replace github.com/Giuseppe-Compagnone/lwn-contracts => ../contracts
