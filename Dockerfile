FROM jrottenberg/ffmpeg
 COPY bin/xkcd /xkcd
 COPY token.txt /token.txt
 ADD events.json /events.json
 ENTRYPOINT ["/xkcd"]
