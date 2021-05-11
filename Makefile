# We make very simple use of Makefiles here.  Make is really just
# a program that allows you to evaluate a collection of rules to
# determine what needs doing.  In most cases you would want to
# make use of more of the power Makefiles can offer, but we just
# use the simplest subset here to keep things simple.
#
# If you want to learn more about Makefiles here are some 
# resources:
#
#     - https://makefiletutorial.com/
#     - https://www.oreilly.com/library/view/managing-projects-with/0596006101/
#     - https://www.gnu.org/software/make/manual/make.pdf
#

all: gen test lint vet build

build: spanlisten spanfetch restserver

spanlisten:
	@cd cmd/spanlisten && go build -o ../../bin/spanlisten

spanfetch:
	@cd cmd/spanfetch && go build -o ../../bin/spanfetch

restserver:
	@cd cmd/restserver && go build -o ../../bin/restserver

lint:
	@revive ./...

vet:
	@go vet ./...

test:
	@go test ./...

gen:
	@buf generate

count:
	@gocloc --not-match-d pkg/apipb .

init:
	@go get -u google.golang.org/protobuf/cmd/protoc-gen-go \
		github.com/bufbuild/buf/cmd/buf \
		github.com/mgechev/revive \
		github.com/hhatto/gocloc/cmd/gocloc