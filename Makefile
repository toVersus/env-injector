VERSION=0.1.0
APPNAME=toversus/env-injector

release: clean build push clean

build:
	docker build -t ${APPNAME}:${VERSION} .

push:
	docker push ${APPNAME}:${VERSION}

clean:
	docker rm -f ${APPNAME}:${VERSION} 2> /dev/null || true

.PHONY: release clean build push
