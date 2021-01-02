resolve-gateway-api:
	ytt -f services/gateway-api | ko resolve -f -

deploy-gateway-api:
	kapp deploy -a test-api -c -f <(ytt -f services/gateway-api | ko resolve -f -)

proto:
	protoc proto/moraceae.proto --go_out=paths=source_relative:.
