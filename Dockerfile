FROM jrottenberg/ffmpeg:latest
 COPY bin/xkcd /xkcd
 COPY token.txt /token.txt
 ADD events.json /events.json
 ENTRYPOINT ["/xkcd"]
