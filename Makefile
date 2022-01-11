UNAME:=$(shell uname|sed 's/.*/\u&/')
OS:=$(shell echo $(GOOS)| sed 's/.*/\u&/')
PKG=$(shell basename $$(pwd))

ifeq  ($(FILENAME),)
FNAME:=main
else
FNAME:=$(FILENAME)
endif

main: pkg/*.go clean
ifeq ($(GOOS),)
	@printf "OS not specified, defaulting to: \e[33m$(UNAME)\e[39m\n"
else
	@printf "OS specified: \e[33m$$(echo $$GOOS | sed 's/.*/\u&/' )\e[39m\n"
endif
	@echo "Building..."
	@cd pkg; export GOARCH=amd64; \
	export CGO_ENABLED=0;\
	export GitCommit=`git rev-parse HEAD | cut -c -7`;\
	export BuildTime=`date -u +%Y%m%d.%H%M%S`;\
	export Authors=`git log --format='%aN' | sort -u | sed "s@root@@"  | tr '\n' ';' | sed "s@;;@;@g" | sed "s@;@; @g" | sed "s@\(.*\); @\1@" | sed "s@[[:blank:]]@SpAcE@g"`;\
	export GitTag=$$(TAG=`git tag --contains $$(git rev-parse HEAD) | sort -R | tr '\n' ' '`; if [ "$$(printf "$$TAG")" ]; then printf "$$TAG"; else printf "undefined"; fi);\
	go build -ldflags "-X main.BuildNo=$$BITBUCKET_BUILD_NUMBER -X main.GitCommit=$$GitCommit -X main.Tag=$$GitTag -X main.BuildTime=$$BuildTime -X main.Authors=$$Authors" -o "$(FNAME)" ./...
	@mv pkg/main build/main
	@printf "\e[32mSuccess!\e[39m\n"
clean:  
	@printf "Cleaning up \e[32mmain\e[39m...\n"
	sudo rm -rf ./data
	rm -f main $(FNAME)

install: clean main
	mv $(FNAME) "$$GOPATH/bin/$(PKG)"

vet: 
	@echo "Running go vet..."
	@go vet || (printf "\e[31mGo vet failed, exit code $$?\e[39m\n"; exit 1)
	@printf "\e[32mGo vet success!\e[39m\n"

test: clean vet main
ifeq ($(UNAME),$(OS))
	@echo "Running $(FNAME)..."
	@./$(FNAME) --version || (printf "\e[31mBuild failed, exit code  $$?\e[39m\n"; exit 1)
	@printf "\e[32mSuccess!\e[39m\n"
else ifeq ($(GOOS),)
	@echo "Running main..."
	@./$(FNAME) --version || (printf "\e[31mBuild failed, exit code  $$?\e[39m\n"; exit 1)
	@printf "\e[32mSuccess!\e[39m\n"
else
	@printf "Looks like you built a binary for \e[33m$$(echo $$GOOS | sed 's/.*/\u&/' )\e[39m but you're running \e[35m$$(uname -s)\e[39m\n"
	@printf "\e[31mRefusing to run the executable.\e[39m\n"
endif

run: main
	@echo "Running $(FNAME)..."
	./$(FNAME) || true
