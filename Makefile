build:
	docker build -f infrastructure/container/Dockerfile . -t service

push-image:
	docker build -f infrastructure/container/Dockerfile . -t service
	docker tag service eu.gcr.io/idyllic-silicon-343409/service
	docker push eu.gcr.io/idyllic-silicon-343409/service