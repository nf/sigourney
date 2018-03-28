FROM debian:sid
ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends golang git mercurial ca-certificates build-essential curl pkg-config portaudio19-dev wget unzip cmake openjdk-7-jdk && apt-get clean
RUN wget http://downloads.sourceforge.net/project/portmedia/portmidi/217/portmidi-src-217.zip && unzip portmidi-src-217.zip && rm portmidi-src-217.zip
ENV JAVA_HOME /usr/lib/jvm/java-7-openjdk-amd64
RUN cd portmidi && cmake .
RUN sed -i 's/pm_java\/pm_java/pm_java/' portmidi/pm_java/CMakeFiles/pmdefaults_target.dir/build.make # see http://stackoverflow.com/questions/14127821/install-portmidi-library
RUN cd portmidi && make install
ENV LD_LIBRARY_PATH /usr/local/lib
ADD . /gopath/src/github.com/nf/sigourney
ENV GOPATH /gopath
RUN go get github.com/nf/sigourney
WORKDIR /gopath/src/github.com/nf/sigourney
EXPOSE 8080
CMD /gopath/bin/sigourney -listen="0.0.0.0:8080" -browser=false