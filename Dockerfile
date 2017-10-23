FROM jrottenberg/ffmpeg:latest
 COPY bin/xkcd /xkcd
 ADD events.json /events.json
 ENTRYPOINT ["/xkcd"]
