FROM jrottenberg/ffmpeg:latest
 ARG TOKEN
 ARG YOUTUBE_TOKEN
 RUN echo $TOKEN > /token.txt
 RUN echo ";" >> /token.txt
 RUN echo $YOUTUBE_TOKEN > /youtube.txt
 RUN echo ";" >> /youtube.txt
 COPY bin/xkcd /xkcd
 COPY events.json /events.json
 ENTRYPOINT ["/xkcd"]
