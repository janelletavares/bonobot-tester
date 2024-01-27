FROM ubuntu:focal

#USER root

ENV DEBIAN_FRONTEND noninteractive

ENV ZOOM_DISPLAY_NAME=""
ARG ZOOM_MEETING_ID=""
ARG ZOOM_MEETING_PASSCODE=""
ARG ZOOM_SESSION_LENGTH_SECONDS=30

RUN apt-get update
# The latest version of selenium and geckodriver did not work with the latest version of python
RUN apt-get install python3.8 firefox python3-pip wget -y

RUN pip3 install selenium # 4.17.2

RUN wget https://github.com/mozilla/geckodriver/releases/download/v0.34.0/geckodriver-v0.34.0-linux64.tar.gz -O /tmp/geckodriver.tar.gz
RUN tar -C /opt -xzf /tmp/geckodriver.tar.gz
RUN chmod 755 /opt/geckodriver
RUN ln -fs /opt/geckodriver /usr/bin/geckodriver
RUN ln -fs /opt/geckodriver /usr/local/bin/geckodriver

RUN mkdir /app
COPY participant.py /app

CMD ["python3", "/app/participant.py"]
