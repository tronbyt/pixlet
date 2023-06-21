FROM golang:latest

RUN apt update && apt upgrade && apt install unzip libwebp-dev python3-venv python3-pip -y

WORKDIR /tmp
RUN curl -fsSL https://deb.nodesource.com/setup_16.x | bash - && apt-get install -y nodejs && node -v

#uncomment below if you want to compile during build
#RUN git clone https://github.com/tidbyt/pixlet.git
#WORKDIR /tmp/pixlet
#RUN npm install && npm run build && make build

CMD ["bash"]