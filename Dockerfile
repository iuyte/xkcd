FROM jrottenberg/ffmpeg:latest
 ARG DISCORD_TOKEN
 ENV DISCORD_TOKEN $(DISCORD_TOKEN)
 COPY bin/xkcd /xkcd
 COPY events.json /events.json
 ENTRYPOINT ["/xkcd", "$DISCORD_TOKEN"]
