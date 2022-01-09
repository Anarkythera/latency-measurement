SHELL = bash

build:
	go mod tidy && go build cmd/main.go -o latencychecker

buildImage:
	docker build -t latencychecker .

test:
	mkdir -p result
	docker run -e API_KEY -v /home/carlosms/Desktop/library/interview_take_home/ably/latency-measurement/results:/tmp latencychecker -n clientA -m 5 -d 2 -w 30 &
	docker run -e API_KEY -v /home/carlosms/Desktop/library/interview_take_home/ably/latency-measurement/results:/tmp latencychecker -n clientB -m 5 -d 2 -w 30 &
	docker run -e API_KEY -v /home/carlosms/Desktop/library/interview_take_home/ably/latency-measurement/results:/tmp latencychecker -n clientC -m 5 -d 2 -w 30 &
	docker run -e API_KEY -v /home/carlosms/Desktop/library/interview_take_home/ably/latency-measurement/results:/tmp latencychecker -n clientD -m 5 -d 2 -w 30 &


.PHONY: buildImage test
