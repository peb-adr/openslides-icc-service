build-dev:
	docker build . --target development --tag openslides-icc-dev

run-tests:
	docker build . --target testing --tag openslides-icc-test
	docker run openslides-icc-test
