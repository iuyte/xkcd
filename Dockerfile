FROM jrottenberg/ffmpeg:latest
 COPY bin/xkcd /xkcd
 COPY events.json /events.json
 ENTRYPOINT ["/xkcd"]
