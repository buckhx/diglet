FROM alpine

MAINTAINER buckhx

ENV DIG_ROOT=/opt/diglet DIG_TILES=/opt/diglet/var/tiles
RUN mkdir -p ${DIG_ROOT}/bin ${DIG_TILES} && ln -s ${DIG_ROOT}/bin/diglet /usr/local/bin/diglet
ADD dist/diglet ${DIG_ROOT}/bin/

CMD diglet tms --port 8080 $DIG_TILES
EXPOSE "8080"
