FROM debian:sid
ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update
RUN apt-get install -y --no-install-recommends libportmidi-dev golang git mercurial ca-certificates build-essential curl  pkg-config portaudio19-dev
ADD . /gopath/src/github.com/nf/sigourney
ENV GOPATH /gopath
RUN go get -x -d github.com/rakyll/portmidi
RUN cd /gopath/src/github.com/rakyll/portmidi; curl https://github.com/proppy/portmidi/commit/91ca6c43fe5607b3150a720b733c580835c56716.patch | patch -p1
RUN go get -x github.com/nf/sigourney
EXPOSE 8080
CMD /gopath/bin/sigourney