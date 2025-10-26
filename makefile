.PHONY: tidy
tidy:
	cd pkg && go mod tidy
	cd services/log_service && go mod tidy
	cd services/user_service && go mod tidy
	cd services/wallet_service && go mod tidy

.PHONY: init
init:
	cd pkg && go mod init pkg
	cd services/log_service && go mod init log_service
	cd services/user_service && go mod init user_service
	cd services/wallet_service && go mod init wallet_service
	go work init
	go work use ./pkg
	go work use ./services/log_service
	go work use ./services/user_service
	go work use ./services/wallet_service