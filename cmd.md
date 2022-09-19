```shell
go run main.go dkg --config dkg/id-10001-input.yaml &
go run main.go dkg --config dkg/id-10002-input.yaml &
go run main.go dkg --config dkg/id-10003-input.yaml &
go run main.go dkg --config dkg/id-10004-input.yaml &
```

```shell
go run main.go signer --config signer/id-10001-input.yaml &
go run main.go signer --config signer/id-10002-input.yaml &
#go run main.go signer --config signer/id-10003-input.yaml &
go run main.go signer --config signer/id-10004-input.yaml &
```

```shell
go run main.go reshare --config reshare/id-10001-input.yaml &
go run main.go reshare --config reshare/id-10002-input.yaml &
go run main.go reshare --config reshare/id-10003-input.yaml &
```
