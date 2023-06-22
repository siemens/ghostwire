.PHONY: help up down mockdata kindup kinddown

help:
	@echo "targets:"
	@echo "    help        shows this help."
	@echo "    up          creates a set of test containers with serveral types of networks"
	@echo "    kindup      creates a KinD test cluster"
	@echo "    mockdata    writes cleaned-up discovery data to src/models/gw/mock/mockdata.json"
	@echo "    down        removes test containers and their networks"
	@echo "    kinddown    removes the KinD test cluster"

MACMASTER = $(shell ls -l /sys/class/net/ | grep -v virtual | awk '{print $$9}' | grep -E '^(en|eth)')

up:
	@echo "designated MACVLAN master: ${MACMASTER}"
	MACMASTER=${MACMASTER} docker-compose -p mock -f docker-compose.yaml up --detach --scale sharednet=3
	docker-compose -p mock -f docker-compose.yaml stop -t 1 foostopped || true
	docker-compose -p mock -f docker-compose.yaml pause foopaused || true
	docker-compose -p mock -f docker-compose.yaml restart -t 1 foorestarted || true
	./vxlan.sh

down:
	docker-compose -p mock -f docker-compose.yaml down -t 1
	ip link show vxtestlan123 >/dev/null 2>&1 && sudo ip link del vxtestlan123

kindup:
	${HOME}/go/bin/kind create cluster --name kindlymock --config kindlymock2.yaml
	kubectl run --pod-running-timeout=2m -i --tty busybox --image=busybox -- sh

kinddown:
	${HOME}/go/bin/kind delete cluster --name kindlymock
	docker network rm kind

mockdata:
	npx ts-node -O '{"module": "commonjs"}' getmockdata.ts "/tmp/$$$$-mockdata.json" && mv -f "/tmp/$$$$-mockdata.json" ../src/models/gw/mock/mockdata.json
	echo "done moving to src/models/gw/mock/mockdata.json"
