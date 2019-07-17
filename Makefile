NAME		:= azzzak/mystem
VER			:= `cat VERSION`
GIT_VER	:= v${VER}
IMAGE		:= ${NAME}:${VER}
LATEST	:= ${NAME}:latest

image:
	@docker build --rm --build-arg APP_VER=${GIT_VER} -t ${IMAGE} -t ${LATEST} .

push:
	@docker push ${NAME}

tag:
	@git tag -a ${GIT_VER} -m "Version ${VER}"
	@git push origin ${GIT_VER}