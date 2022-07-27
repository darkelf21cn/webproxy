.PHONY: build push

IMG_NAME		= darkelf21cn/webproxy

build:
	@docker build -t $(IMG_NAME) .

push:
	@docker push $(IMG_NAME)
