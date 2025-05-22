$env:GOOS='linux'
$env:GOARCH='arm64'
go build
scp .\mall opi:mall/build/mall