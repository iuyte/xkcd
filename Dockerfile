FROM jrottenberg/ffmpeg:latest
 ARG TOKEN
 RUN echo $TOKEN > /token.txt
 RUN echo ";" >> /token.txt
 COPY bin/xkcd /xkcd
 COPY events.json /events.json
 ENTRYPOINT ["/xkcd"]
