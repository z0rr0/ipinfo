#!/usr/bin/env bash
# Example srcript for for automatic update
# https://dev.maxmind.com/geoip/geoipupdate/
#
# Please check variables before usage:
# CONTAINER - docker contrainer name
# DB_TO - target database file
# U - user

GEOCMD="$(command -v geoipupdate)"
DCR="$(command -v docker)"
CPCMD="$(command -v cp)"
CONTAINER="ipinfo_1"
DB_FROM="/usr/share/GeoIP/GeoLite2-City.mmdb"
DB_TO="/storage/GeoLite2-City.mmdb"
U="user"

/bin/date --rfc-3339=seconds

if [ -z "$GEOCMD" ]; then
  echo "not found geoipupdate command: $GEOCMD"
  exit 1
fi

if [ -z "$DCR" ]; then
  echo "not found docker command: $DCR"
  exit 1
fi

if ! ${GEOCMD} -v
then
  echo "failed checked/updated"
  exit 2
fi

if ! ${CPCMD} -vf ${DB_FROM} ${DB_TO}
then
  echo "failed copy"
  exit 3
fi

/bin/chown ${U}:${U} ${DB_TO}
/bin/chmod 640 ${DB_TO}

if ! ${DCR} restart ${CONTAINER}
then
  echo "failed docker restart"
  exit 4
fi

echo "successful done"
